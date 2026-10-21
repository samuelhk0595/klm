package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const sessionRoleGeneralAgent = "general_agent"

// A general conversation is engine-owned, never a synthetic user project.
func (d *diskState) generalSession() *Session {
	for i := range d.Sessions {
		if d.Sessions[i].Role == sessionRoleGeneralAgent {
			return &d.Sessions[i]
		}
	}
	return nil
}

// Shared by message preparation, queue execution and catalog discovery. Ordinary
// chats retain their registered project directory; engine-owned chats use theirs.
func (d *diskState) conversationDirectory(s *Session) (string, bool) {
	if s == nil {
		return "", false
	}
	if s.Role == sessionRoleGeneralAgent {
		return s.ExecutionCWD, s.ExecutionCWD != ""
	}
	if s.Role == sessionRoleSubagent && s.ProjectID == "" {
		return d.conversationDirectory(d.session(s.ParentID))
	}
	p := d.project(s.ProjectID)
	if p == nil || p.Removed {
		return "", false
	}
	return p.Folder, true
}

func conversationGuidance(s *Session) (string, error) {
	if s.Role == sessionRoleGeneralAgent {
		return loadInternalPrompt("general-agent.md")
	}
	guidance, err := sessionCollaborationPrompt()
	if err != nil {
		return "", err
	}
	topicPrompt := ""
	if s.Role == "" && s.ParentID == "" && s.GraphRunID == "" {
		if s.TopicID == "" {
			topicPrompt = "session-topic-pending.md"
		} else {
			topicPrompt = "session-topic-assigned.md"
		}
	} else if s.Role == sessionRoleSideAgent {
		topicPrompt = "session-topic-side.md"
	}
	if topicPrompt != "" {
		body, err := loadInternalPrompt(topicPrompt)
		if err != nil {
			return "", err
		}
		guidance += "\n\n" + strings.ReplaceAll(body, "{{topicId}}", s.TopicID)
	}
	return guidance, nil
}

func (a *app) generalConversation(w http.ResponseWriter, r *http.Request) {
	var body struct{}
	if !decode(w, r, &body) {
		return
	}
	// Directory creation is idempotent and happens outside the state mutex.
	cwd, err := filepath.Abs(filepath.Join(a.dir, "workspaces", "general-agent"))
	if err == nil {
		err = os.MkdirAll(cwd, 0700)
	}
	if err != nil {
		fail(w, 503, "Cannot prepare the general agent workspace.")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if s := a.state.generalSession(); s != nil {
		respond(w, 200, a.currentSessionUpdateLocked(s.ID))
		return
	}
	harness := "pi"
	for _, option := range a.harnesses {
		if option.Available {
			harness = option.ID
			break
		}
	}
	s := newTopLevelSession("", "General agent", "", harness, "", "")
	s.Role, s.ExecutionCWD = sessionRoleGeneralAgent, cwd
	if err := a.commitLocked(func(d *diskState) { a.appendTopLevelSession(d, s) }); err != nil {
		fail(w, 503, err.Error())
		return
	}
	respond(w, 201, a.currentSessionUpdateLocked(s.ID))
}
