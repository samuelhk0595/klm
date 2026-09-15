package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type SessionMetadata struct {
	SessionID string `json:"sessionId"`
	GitBranch string `json:"gitBranch,omitempty"`
}

func (a *app) getSessionMetadata(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	s := a.state.session(r.PathValue("id"))
	if s == nil {
		a.mu.Unlock()
		fail(w, 404, "Session not found.")
		return
	}
	id, cwd := s.ID, s.ExecutionCWD
	if cwd == "" {
		if project := a.state.project(s.ProjectID); project != nil {
			cwd = project.Folder
		}
	}
	a.mu.Unlock()

	metadata := SessionMetadata{SessionID: id}
	if cwd != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		metadata.GitBranch = currentGitBranch(ctx, cwd)
		cancel()
	}
	respond(w, 200, metadata)
}

func currentGitBranch(ctx context.Context, cwd string) string {
	git, ok := executable("git")
	if !ok {
		return ""
	}
	cmd := exec.CommandContext(ctx, git, "symbolic-ref", "--quiet", "--short", "HEAD")
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GCM_INTERACTIVE=Never")
	cmd.Stderr = &cappedBuffer{}
	cmd.WaitDelay = time.Second
	configureProcess(cmd)
	output, err := cmd.Output()
	if err != nil || ctx.Err() != nil || len(output) > 4096 {
		return ""
	}
	branch := strings.TrimSpace(strings.ToValidUTF8(string(output), ""))
	if branch == "HEAD" || strings.ContainsAny(branch, "\r\n") {
		return ""
	}
	return branch
}
