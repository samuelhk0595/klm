package main

import (
	"errors"
	"strings"
)

const (
	sessionRoleSideAgent = "side_agent"
	sessionRoleSubagent  = "subagent"
)

func subagentTitle(value string) string {
	if title, ok := cleanLabel(value, 200); ok {
		return title
	}
	return "Subagent"
}

// Subagents are persisted as hidden child sessions so their native event stream
// can use the same history, pagination, and rendering contract as a normal chat.
func (p *adapter) ensureSubagent(nativeID, title, model, effort string) (*adapter, error) {
	if nativeID == "" || len(nativeID) > 256 || strings.IndexFunc(nativeID, func(r rune) bool { return r <= ' ' || r == 127 }) != -1 {
		return nil, errors.New("Harness returned an invalid subagent identifier.")
	}
	if child := p.subagents[nativeID]; child != nil {
		return child, nil
	}
	p.app.mu.Lock()
	parent := p.app.state.session(p.id)
	if parent == nil || parent.ParentID != "" || parent.GraphRunID != "" {
		p.app.mu.Unlock()
		return nil, nil
	}
	if model == "" {
		model = parent.ResolvedModel
		if model == "" {
			model = parent.Model
		}
	}
	if effort == "" {
		effort = parent.ResolvedEffort
		if effort == "" {
			effort = parent.Effort
		}
	}
	var childSession *Session
	for i := range p.app.state.Sessions {
		candidate := &p.app.state.Sessions[i]
		if candidate.ParentID == parent.ID && candidate.Role == sessionRoleSubagent && p.app.state.Native[candidate.ID].ID == nativeID {
			childSession = candidate
			break
		}
	}
	if childSession == nil {
		created := now()
		child := Session{ID: newID(), ParentID: parent.ID, ProjectID: parent.ProjectID, Title: subagentTitle(title), Workspace: parent.Workspace,
			Role: sessionRoleSubagent, Harness: parent.Harness, Model: model, Effort: effort, Status: "idle", Events: []Event{}, CreatedAt: created, UpdatedAt: created}
		if err := p.app.commitLocked(func(d *diskState) {
			d.Sessions = append(d.Sessions, child)
			d.Native[child.ID] = nativeSession{ID: nativeID}
		}); err != nil {
			p.app.mu.Unlock()
			return nil, err
		}
		childSession = p.app.state.session(child.ID)
	}
	childID := childSession.ID
	results := map[string]string{}
	for _, e := range parent.Events {
		if e.Type == "subagent" && str(e.Data, "childSessionId") == childID {
			if execution := subagentEventExecution(e); execution != "" {
				p.keys[execution] = e.ID
				if !subagentRunning(e.Status) {
					results[execution] = e.Status
				}
			}
		}
	}
	p.app.mu.Unlock()
	child := &adapter{app: p.app, turn: p.turn, id: childID, harness: p.harness, role: sessionRoleSubagent, subagent: true, yolo: p.yolo, cwd: p.cwd,
		keys: map[string]string{}, toolNames: map[string]string{}, commands: map[string]string{}, subagentResults: results, processesDrained: true, model: model, effort: effort}
	child.stream = newStreamBatch(child)
	if p.subagents == nil {
		p.subagents = map[string]*adapter{}
	}
	p.subagents[nativeID] = child
	return child, nil
}

func subagentRunning(status string) bool {
	return status == "running" || status == "pending" || status == "started"
}

func subagentEventExecution(e Event) string {
	if execution := str(e.Data, "subagentExecutionId"); execution != "" {
		return execution
	}
	// Histories persisted before cycle correlation already contain native call
	// identifiers. Reuse their keys/results without rewriting those events.
	if messageID, partID := str(e.Data, "messageID"), str(e.Data, "partID"); messageID != "" && partID != "" {
		return "opencode/" + messageID + "/" + partID
	}
	if callID, nativeID := str(e.Data, "nativeCallId"), str(e.Data, "nativeAgentId"); callID != "" && nativeID != "" {
		return "codex/subagent/" + callID + "/" + nativeID
	}
	return ""
}

// Registration also happens during replay. Only an explicit native call/turn
// start opens a cycle; a completed cycle cannot be reopened by an old snapshot.
func (p *adapter) startSubagent(child *adapter, execution string) error {
	if child == nil || execution == "" || child.subagentExecution == execution || child.subagentResults[execution] != "" {
		return nil
	}
	if child.stream != nil {
		if err := child.stream.flush(nil); err != nil {
			return err
		}
	}
	observed := map[string]bool{}
	for _, id := range child.keys {
		observed[id] = true
	}
	child.app.mu.Lock()
	defer child.app.mu.Unlock()
	if err := child.app.commitLocked(func(d *diskState) {
		if s := d.session(child.id); s != nil {
			s.Status, s.UpdatedAt = "running", now()
			// Restored children may emit parts before task metadata identifies
			// the call. Adopt only this adapter's not-yet-correlated observations.
			for i, e := range s.Events {
				if observed[e.ID] && str(e.Data, "subagentExecutionId") == "" {
					data := map[string]any{}
					for key, value := range e.Data {
						data[key] = value
					}
					data["subagentExecutionId"] = execution
					e.Data = data
					d.setEvent(s, i, e)
				}
			}
		}
	}); err != nil {
		return err
	}
	child.subagentExecution = execution
	child.failed, child.completed = false, false
	return nil
}

// Restore observation before a reused child's frames arrive, without starting it.
func (p *adapter) restoreSubagents() error {
	p.app.mu.Lock()
	nativeIDs := []string{}
	for _, s := range p.app.state.Sessions {
		if s.ParentID == p.id && s.Role == sessionRoleSubagent {
			if nativeID := p.app.state.Native[s.ID].ID; nativeID != "" {
				nativeIDs = append(nativeIDs, nativeID)
			}
		}
	}
	p.app.mu.Unlock()
	for _, nativeID := range nativeIDs {
		if _, err := p.ensureSubagent(nativeID, "", "", ""); err != nil {
			return err
		}
	}
	return nil
}

func (p *adapter) finishSubagent(child *adapter, status string) error {
	if child == nil || child.subagentExecution == "" {
		return nil
	}
	return p.finishSubagentExecution(child, child.subagentExecution, status)
}

func (p *adapter) finishSubagentExecution(child *adapter, execution, status string) error {
	if child == nil || execution == "" || child.subagentResults[execution] != "" {
		return nil
	}
	if child.subagentExecution != execution {
		child.subagentResults[execution] = status
		return nil
	}
	if child.stream != nil {
		if err := child.stream.flush(nil); err != nil {
			return err
		}
	}
	sessionStatus := "idle"
	if status == "error" || status == "failed" || status == "interrupted" {
		sessionStatus = "error"
	}
	child.app.mu.Lock()
	defer child.app.mu.Unlock()
	err := child.app.commitLocked(func(d *diskState) {
		s := d.session(child.id)
		if s == nil {
			return
		}
		s.Status, s.UpdatedAt = sessionStatus, now()
		for i := range s.Events {
			if subagentRunning(s.Events[i].Status) && str(s.Events[i].Data, "subagentExecutionId") == execution {
				e := s.Events[i]
				e.Status = status
				d.setEvent(s, i, e)
			}
		}
		syncSubagentExecution(d, s, execution, status)
	})
	if err == nil {
		child.subagentResults[execution] = status
		child.subagentExecution = ""
	}
	return err
}

func syncSubagentParent(d *diskState, child *Session, status string) {
	syncSubagentExecution(d, child, "", status)
}

func syncSubagentExecution(d *diskState, child *Session, execution, status string) {
	if child == nil || child.Role != sessionRoleSubagent {
		return
	}
	parent := d.session(child.ParentID)
	if parent == nil {
		return
	}
	for i := range parent.Events {
		value, _ := parent.Events[i].Data["childSessionId"].(string)
		if parent.Events[i].Type == "subagent" && value == child.ID && subagentRunning(parent.Events[i].Status) && (execution == "" || str(parent.Events[i].Data, "subagentExecutionId") == execution) {
			e := parent.Events[i]
			e.Status = status
			d.setEvent(parent, i, e)
			parent.UpdatedAt = now()
		}
	}
}

func (p *adapter) closeSubagents(failed, cancelled bool) error {
	var result error
	for _, child := range p.subagents {
		result = errors.Join(result, child.stream.close())
		status := "completed"
		if cancelled {
			status = "cancelled"
		} else if failed {
			status = "error"
		}
		result = errors.Join(result, p.finishSubagent(child, status))
	}
	return result
}
