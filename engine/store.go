package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Project struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Folder  string   `json:"folder"`
	Icon    string   `json:"icon"`
	Folders []string `json:"folders"`
	Removed bool     `json:"removed,omitempty"`
}

type Event struct {
	ConsultationID string         `json:"consultationId,omitempty"`
	ID             string         `json:"id"`
	Type           string         `json:"type"`
	Text           string         `json:"text"`
	Title          string         `json:"title,omitempty"`
	Status         string         `json:"status,omitempty"`
	CreatedAt      string         `json:"createdAt"`
	Data           map[string]any `json:"data,omitempty"`
}

type Session struct {
	ParentID       string            `json:"parentId,omitempty"`
	Sources        []SourceReference `json:"sources,omitempty"`
	ID             string            `json:"id"`
	ProjectID      string            `json:"projectId"`
	Title          string            `json:"title"`
	Workspace      string            `json:"workspace"`
	Harness        string            `json:"harness"`
	Model          string            `json:"model,omitempty"`
	Effort         string            `json:"effort,omitempty"`
	ResolvedModel  string            `json:"resolvedModel,omitempty"`
	ResolvedEffort string            `json:"resolvedEffort,omitempty"`
	Status         string            `json:"status"`
	Events         []Event           `json:"events"`
	Permissions    []Permission      `json:"permissions,omitempty"`
	Questions      []QuestionRequest `json:"questions,omitempty"`
	Usage          *SessionUsage     `json:"usage,omitempty"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
}

type nativeSession struct {
	ID   string `json:"id,omitempty"`
	Path string `json:"path,omitempty"`
}

type diskState struct {
	Consultations []Consultation           `json:"consultations,omitempty"`
	Version       int                      `json:"version"`
	Projects      []Project                `json:"projects"`
	Sessions      []Session                `json:"sessions"`
	Native        map[string]nativeSession `json:"native"`
	Grants        []permissionGrant        `json:"permissionGrants,omitempty"`
}

func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("cannot generate secure identifier")
	}
	return hex.EncodeToString(b[:])
}

func event(kind, text string) Event {
	return Event{ID: newID(), Type: kind, Text: text, CreatedAt: now()}
}

func (d *diskState) session(id string) *Session {
	for i := range d.Sessions {
		if d.Sessions[i].ID == id {
			return &d.Sessions[i]
		}
	}
	return nil
}

func (d *diskState) project(id string) *Project {
	for i := range d.Projects {
		if d.Projects[i].ID == id {
			return &d.Projects[i]
		}
	}
	return nil
}

func loadState(dir string) (diskState, error) {
	d := diskState{Version: 1, Projects: []Project{}, Sessions: []Session{}, Native: map[string]nativeSession{}}
	b, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if errors.Is(err, os.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	d = diskState{}
	if err := json.Unmarshal(b, &d); err != nil {
		return d, fmt.Errorf("invalid state.json: %w", err)
	}
	if d.Version != 1 || d.Projects == nil || d.Sessions == nil || d.Native == nil {
		return d, errors.New("unsupported or incomplete state.json; refusing to overwrite")
	}
	ids := map[string]bool{}
	for _, p := range d.Projects {
		if p.ID == "" || ids[p.ID] || p.Folders == nil {
			return d, errors.New("invalid project in state.json")
		}
		ids[p.ID] = true
	}
	for _, s := range d.Sessions {
		if s.ID == "" || ids[s.ID] || d.project(s.ProjectID) == nil || s.Events == nil ||
			(s.Status != "idle" && s.Status != "running" && s.Status != "error") {
			return d, errors.New("invalid session in state.json")
		}
		ids[s.ID] = true
	}
	parents := map[string]bool{}
	for _, s := range d.Sessions {
		if s.ParentID != "" {
			parent := d.session(s.ParentID)
			if parent == nil || parent.ParentID != "" || parent.ProjectID != s.ProjectID || parents[s.ParentID] {
				return d, errors.New("invalid linked conversation in state.json")
			}
			parents[s.ParentID] = true
		}
	}
	for _, c := range d.Consultations {
		linked := d.linked(c.From)
		if c.ID == "" || ids[c.ID] || linked == nil || linked.ID != c.To || !validLinkedText(c.Question, 32<<10) {
			return d, errors.New("invalid consultation in state.json")
		}
		switch c.Status {
		case "queued", "answering", "completed", "failed", "cancelled", "interrupted":
		default:
			return d, errors.New("invalid consultation status in state.json")
		}
		switch c.Delivery {
		case "waiting", "pending", "delivering", "delivered", "failed", "cancelled", "interrupted":
		default:
			return d, errors.New("invalid consultation delivery in state.json")
		}
		if _, err := time.Parse(time.RFC3339Nano, c.CreatedAt); err != nil {
			return d, errors.New("invalid consultation timestamp in state.json")
		}
		ids[c.ID] = true
	}
	return d, nil
}

func saveState(dir string, d *diskState) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".state-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return replaceFile(name, filepath.Join(dir, "state.json"))
}

// Caller holds app.mu. Publish only a successfully persisted, immutable snapshot.
func (a *app) commitLocked(change func(*diskState)) error {
	if a.storageErr != nil {
		return a.storageErr
	}
	b, err := json.Marshal(a.state)
	var next diskState
	if err == nil {
		err = json.Unmarshal(b, &next)
	}
	if err == nil {
		change(&next)
		err = saveState(a.dir, &next)
	}
	if err != nil {
		a.storageErr = errors.New("state persistence failed; engine is read-only until restarted")
		for _, r := range a.runs {
			r.cancel()
		}
		return a.storageErr
	}
	a.state = next
	for _, listeners := range a.listeners {
		for ch := range listeners {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
	}
	return nil
}
