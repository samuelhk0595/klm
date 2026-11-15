package main

import (
	"strings"
)

func (p *adapter) trackRunnerTool(key, kind, title, status string, raw map[string]any) {
	if p.turn == nil || p.turn.runner == nil || kind != "tool" && kind != "command" {
		return
	}
	r := p.turn.runner
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeTools == nil {
		r.activeTools = map[string]string{}
	}
	if status != "running" {
		delete(r.activeTools, key)
		return
	}
	name := str(raw, "toolName")
	if name == "" {
		name = str(raw, "tool")
	}
	if name == "" {
		name = str(object(raw["item"]), "tool")
	}
	if name == "" {
		name = title
	}
	r.activeTools[key] = name
}

func (p *adapter) runnerWaiting(waiting bool) error {
	a := p.app
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.writeSQLLocked(p.turn.ctx, func(st *sqliteSessionTx) error {
		execution, err := st.runnerExecution(p.turn.runner.ExecutionID)
		if err != nil {
			return err
		}
		if execution.NativeWaiting == waiting {
			return nil
		}
		execution.NativeWaiting = waiting
		return st.putRunnerExecution(execution)
	})
}

func (a *app) runnerAttention(execution runnerExecution) (bool, int, error) {
	var jobs int
	err := a.store.readers.QueryRow(`SELECT count(*) FROM runner_jobs WHERE execution_id=? AND state IN ('starting','running','unknown')`, execution.ID).Scan(&jobs)
	if err != nil {
		return false, 0, err
	}
	a.mu.Lock()
	t := a.runs[execution.SessionID]
	if t == nil || t.runner == nil {
		a.mu.Unlock()
		return false, jobs, nil
	}
	questions := len(t.questions) + len(t.approvals)
	blockedMCP := map[string]string{}
	for _, pending := range t.questions {
		if pending.request.Elicitation != nil {
			elicitation := pending.request.Elicitation
			blockedMCP[elicitation.ConnectionID] = elicitation.CallID
		}
	}
	r := t.runner
	p := r.adapter
	a.mu.Unlock()
	if questions == 0 || jobs != 0 {
		return false, jobs, nil
	}
	if execution.Applied.Harness == "codex" && !execution.NativeWaiting {
		return false, jobs, nil
	}
	blockedNames := map[string]bool{}
	if p != nil && p.mcp != nil {
		p.mcp.mu.Lock()
		for connection, call := range p.mcp.activeCalls {
			if blockedMCP[connection] != call {
				p.mcp.mu.Unlock()
				return false, jobs, nil
			}
		}
		for name, route := range p.mcp.routes {
			if blockedMCP[route.connection.ID] != "" {
				blockedNames[name] = true
			}
		}
		p.mcp.mu.Unlock()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.uncertainMCP) > 0 {
		return false, jobs, nil
	}
	for _, name := range r.activeTools {
		if name == "ask_user" || name == "question" || name == "request_user_input" {
			continue
		}
		blocked := false
		for alias := range blockedNames {
			if name == alias || strings.HasSuffix(name, "_"+alias) {
				blocked = true
				break
			}
		}
		if !blocked {
			return false, jobs, nil
		}
	}
	return true, jobs, nil
}
