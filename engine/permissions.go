package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
)

type Permission struct {
	ID          string         `json:"id"`
	Harness     string         `json:"harness"`
	Kind        string         `json:"kind"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Patterns    []string       `json:"patterns"`
	Details     map[string]any `json:"details,omitempty"`
	Decisions   []string       `json:"decisions"`
	AllowLabel  string         `json:"allowLabel,omitempty"`
	CreatedAt   string         `json:"createdAt"`
	Resolving   bool           `json:"resolving,omitempty"`
	SourceID    string         `json:"-"`
}

type permissionGrant struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	SessionID string `json:"sessionId,omitempty"`
	Harness   string `json:"harness"`
	Kind      string `json:"kind"`
	Key       string `json:"key"`
	CreatedAt string `json:"createdAt"`
}

type pendingApproval struct {
	request Permission
	key     string
	reply   func(bool) error
	busy    bool
}

func (p *adapter) requestPermission(req Permission, scope map[string]any, reply func(bool) error) error {
	if err := p.turn.ctx.Err(); err != nil {
		return err
	}
	if req.SourceID == "" || req.Kind == "" || reply == nil {
		return errors.New("Harness permission request is incomplete.")
	}
	key := ""
	if scope != nil {
		encoded, err := json.Marshal(scope)
		if err != nil {
			return errors.New("Could not encode the permission scope.")
		}
		sum := sha256.Sum256(encoded)
		key = hex.EncodeToString(sum[:])
	}
	if len(req.Decisions) == 0 {
		req.Decisions = []string{"once", "session", "always", "reject"}
	}
	if key == "" {
		req.Decisions = slices.DeleteFunc(slices.Clone(req.Decisions), func(choice string) bool { return choice == "session" || choice == "always" })
	}
	req.ID, req.Harness, req.CreatedAt = newID(), p.harness, now()
	if req.Patterns == nil {
		req.Patterns = []string{}
	}
	a := p.app
	a.mu.Lock()
	if a.runs[p.id] != p.turn || p.turn.ctx.Err() != nil || a.storageErr != nil {
		a.mu.Unlock()
		return errors.New("Permission request belongs to an inactive turn.")
	}
	for _, existing := range p.turn.approvals {
		if existing.request.SourceID == req.SourceID {
			a.mu.Unlock()
			return nil
		}
	}
	s := a.state.session(p.id)
	for _, grant := range a.state.Grants {
		if key != "" && slices.Contains(req.Decisions, "once") && grant.ProjectID == s.ProjectID &&
			grant.Harness == p.harness && grant.Kind == req.Kind && grant.Key == key && (grant.SessionID == "" || grant.SessionID == p.id) {
			a.mu.Unlock()
			return reply(true)
		}
	}
	err := a.commitLocked(func(d *diskState) {
		s := d.session(p.id)
		s.Permissions = append(s.Permissions, req)
		s.UpdatedAt = now()
	})
	if err == nil {
		if p.turn.approvals == nil {
			p.turn.approvals = map[string]*pendingApproval{}
		}
		p.turn.approvals[req.ID] = &pendingApproval{request: req, key: key, reply: reply}
	}
	a.mu.Unlock()
	return err
}

func (p *adapter) dismissPermission(sourceID string) error {
	a := p.app
	a.mu.Lock()
	defer a.mu.Unlock()
	for id, pending := range p.turn.approvals {
		if pending.request.SourceID != sourceID {
			continue
		}
		delete(p.turn.approvals, id)
		return a.commitLocked(func(d *diskState) {
			s := d.session(p.id)
			s.Permissions = slices.DeleteFunc(s.Permissions, func(req Permission) bool { return req.ID == id })
			s.UpdatedAt = now()
		})
	}
	return nil
}

func (a *app) permissionDecision(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Decision string `json:"decision"`
	}
	if !decode(w, r, &body) {
		return
	}
	id, requestID := r.PathValue("id"), r.PathValue("permissionID")
	a.mu.Lock()
	t := a.runs[id]
	if t == nil || t.ctx.Err() != nil || t.approvals[requestID] == nil {
		a.mu.Unlock()
		fail(w, 409, "This permission request is no longer pending.")
		return
	}
	pending := t.approvals[requestID]
	if pending.busy || !slices.Contains(pending.request.Decisions, body.Decision) ||
		!slices.Contains([]string{"once", "session", "always", "reject"}, body.Decision) {
		a.mu.Unlock()
		fail(w, 409, "This permission decision is unavailable or already being processed.")
		return
	}
	grantID := ""
	err := a.commitLocked(func(d *diskState) {
		s := d.session(id)
		if body.Decision == "session" || body.Decision == "always" {
			grantID = newID()
			grant := permissionGrant{ID: grantID, ProjectID: s.ProjectID, Harness: s.Harness, Kind: pending.request.Kind, Key: pending.key, CreatedAt: now()}
			if body.Decision == "session" {
				grant.SessionID = id
			}
			d.Grants = append(d.Grants, grant)
		}
		for i := range s.Permissions {
			if s.Permissions[i].ID == requestID {
				s.Permissions[i].Resolving = true
			}
		}
		s.UpdatedAt = now()
	})
	if err != nil {
		a.mu.Unlock()
		fail(w, 503, err.Error())
		return
	}
	pending.busy = true
	a.mu.Unlock()
	if t.ctx.Err() != nil {
		err = t.ctx.Err()
	} else {
		err = pending.reply(body.Decision != "reject")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err != nil {
		rollback := a.commitLocked(func(d *diskState) {
			d.Grants = slices.DeleteFunc(d.Grants, func(grant permissionGrant) bool { return grant.ID == grantID && grantID != "" })
			s := d.session(id)
			for i := range s.Permissions {
				if s.Permissions[i].ID == requestID {
					s.Permissions[i].Resolving = false
				}
			}
			s.UpdatedAt = now()
		})
		pending.busy = false
		if rollback != nil {
			fail(w, 503, rollback.Error())
			return
		}
		fail(w, 409, "Could not deliver the permission decision. The request may have been cancelled; refresh or stop the turn.")
		return
	}
	delete(t.approvals, requestID)
	if err := a.commitLocked(func(d *diskState) {
		s := d.session(id)
		s.Permissions = slices.DeleteFunc(s.Permissions, func(req Permission) bool { return req.ID == requestID })
		s.UpdatedAt = now()
		entry := event("status", "Permission decision: "+body.Decision)
		entry.ConsultationID = t.consultationID
		entry.Title, entry.Status = pending.request.Title, "completed"
		entry.Data = map[string]any{"harness": s.Harness, "permission": pending.request.Kind, "patterns": pending.request.Patterns, "decision": body.Decision}
		s.Events = append(s.Events, entry)
	}); err != nil {
		fail(w, 503, err.Error())
		return
	}
	respond(w, 200, a.state.session(id))
}
