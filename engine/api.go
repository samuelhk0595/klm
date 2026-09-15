package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image/png"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func respond(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}

func fail(w http.ResponseWriter, code int, text string) {
	respond(w, code, map[string]string{"error": text})
}

func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 512<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		fail(w, http.StatusBadRequest, "Invalid JSON request body or body too large.")
		return false
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		fail(w, http.StatusBadRequest, "Expected one JSON object.")
		return false
	}
	return true
}

func loopbackHost(hostport string) bool {
	host := hostport
	if strings.HasPrefix(hostport, "[") && strings.HasSuffix(hostport, "]") {
		host = hostport[1 : len(hostport)-1]
	} else if strings.Contains(hostport, ":") {
		var err error
		host, _, err = net.SplitHostPort(hostport)
		if err != nil {
			return false
		}
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", a.health)
	mux.HandleFunc("GET /api/state", a.getState)
	mux.HandleFunc("POST /api/projects", a.createProject)
	mux.HandleFunc("PATCH /api/projects/{id}", a.updateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", a.removeProject)
	mux.HandleFunc("POST /api/projects/{id}/folders", a.createFolder)
	mux.HandleFunc("GET /api/projects/{id}/paths", a.projectPaths)
	mux.HandleFunc("GET /api/projects/{id}/authoring", a.authoring)
	mux.HandleFunc("GET /api/projects/{id}/models/{harness}", a.projectModels)
	mux.HandleFunc("POST /api/projects/{id}/agents", a.saveAgent)
	mux.HandleFunc("PATCH /api/projects/{id}/agents/{item}", a.saveAgent)
	mux.HandleFunc("DELETE /api/projects/{id}/agents/{item}", a.deleteAgent)
	mux.HandleFunc("POST /api/projects/{id}/graph-drafts", a.createGraphDraft)
	mux.HandleFunc("POST /api/projects/{id}/graphs/{item}", a.saveGraph)
	mux.HandleFunc("PATCH /api/projects/{id}/graphs/{item}", a.patchGraph)
	mux.HandleFunc("DELETE /api/projects/{id}/graphs/{item}", a.deleteGraph)
	mux.HandleFunc("POST /api/projects/{id}/graphs/{item}/layout", a.saveGraphLayout)
	mux.HandleFunc("POST /api/sessions", a.createSession)
	mux.HandleFunc("GET /api/sessions/{id}/graph", a.getConversationGraph)
	mux.HandleFunc("PATCH /api/sessions/{id}/graph", a.selectConversationGraph)
	mux.HandleFunc("GET /api/graph-runs/{runID}", a.getGraphRun)
	mux.HandleFunc("POST /api/sessions/{id}/side", a.chatSessionHandler(a.sideConversation))
	mux.HandleFunc("PATCH /api/sessions/{id}/harness", a.chatSessionHandler(a.sideHarness))
	mux.HandleFunc("POST /api/sessions/{id}/consultations/{requestID}/cancel", a.chatSessionHandler(a.cancelConsultation))
	mux.HandleFunc("PATCH /api/sessions/{id}", a.chatSessionHandler(a.patchSession))
	mux.HandleFunc("GET /api/sessions/{id}/models", a.getModels)
	mux.HandleFunc("PATCH /api/sessions/{id}/settings", a.chatSessionHandler(a.updateModelSettings))
	mux.HandleFunc("GET /api/sessions/{id}/quota", a.getQuota)
	mux.HandleFunc("POST /api/sessions/{id}/messages", a.chatSessionHandler(a.message))
	mux.HandleFunc("GET /api/sessions/{id}/events", a.chatSessionHandler(a.events))
	mux.HandleFunc("POST /api/sessions/{id}/stop", a.chatSessionHandler(a.stop))
	mux.HandleFunc("POST /api/sessions/{id}/permissions/{permissionID}", a.permissionDecision)
	mux.HandleFunc("POST /api/sessions/{id}/questions/{questionID}/reply", a.answerQuestion)
	mux.HandleFunc("POST /api/dialogs/directory", a.directory)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fail(w, 404, "Endpoint not found.") })
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if !publicAPIHost(r.Host) {
			fail(w, 403, "Invalid engine host; use this computer's address on port "+apiPort+".")
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			if !publicAPIOrigin(origin, r.Host) {
				fail(w, 403, "Origin is not an allowed KLM client.")
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || kind != "application/json" {
				fail(w, 415, "Mutations require Content-Type: application/json.")
				return
			}
			a.mu.Lock()
			blocked := a.closing || a.storageErr != nil
			a.mu.Unlock()
			if blocked {
				fail(w, 503, "Engine is shutting down or state persistence has failed.")
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

// Deliberately separate from loopbackHost: the authenticated harness bridge keeps
// its original local-only host and peer validation.
func publicAPIHost(authority string) bool {
	host, port, err := net.SplitHostPort(authority)
	if err != nil || port != apiPort {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return !ip.IsUnspecified() && !ip.IsMulticast()
	}
	name, _ := os.Hostname()
	return strings.EqualFold(host, "localhost") || (name != "" && strings.EqualFold(host, name))
}

func publicAPIOrigin(origin, authority string) bool {
	if origin == "tauri://localhost" || origin == "http://tauri.localhost" || origin == "https://tauri.localhost" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil ||
		u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	host, _, err := net.SplitHostPort(authority)
	if (u.Port() == "5173" || u.Port() == "4173") &&
		(loopbackHost(u.Host) || developmentBuild && err == nil && strings.EqualFold(u.Hostname(), host)) {
		return true
	}
	return err == nil && u.Scheme == "http" && u.Port() == webPort &&
		(strings.EqualFold(u.Hostname(), host) || (loopbackHost(u.Host) && loopbackHost(authority)))
}

func (a *app) health(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.storageErr != nil {
		fail(w, 503, a.storageErr.Error())
		return
	}
	activeGraphs := 0
	for _, run := range a.state.GraphRuns {
		if graphRunActive(run.Status) {
			activeGraphs++
		}
	}
	respond(w, 200, map[string]any{"status": "ok", "version": 2, "runningSessions": len(a.runs), "activeGraphRuns": activeGraphs})
}

func (a *app) getState(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	projects := []Project{}
	sessions := []Session{}
	visible := map[string]bool{}
	for _, project := range a.state.Projects {
		if !project.Removed {
			projects = append(projects, project)
			visible[project.ID] = true
		}
	}
	for _, session := range a.state.Sessions {
		if visible[session.ProjectID] && session.GraphRunID == "" {
			sessions = append(sessions, *a.sessionViewLocked(session.ID))
		}
	}
	state := struct {
		Projects  []Project `json:"projects"`
		Sessions  []Session `json:"sessions"`
		Harnesses []Harness `json:"harnesses"`
	}{projects, sessions, a.harnesses}
	a.mu.Unlock()
	respond(w, 200, state)
}

func cleanLabel(value string, max int) (string, bool) {
	value = strings.TrimSpace(value)
	return value, value != "" && utf8.ValidString(value) && utf8.RuneCountInString(value) <= max &&
		strings.IndexFunc(value, unicode.IsControl) == -1
}

func existingDirectory(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errors.New("Folder must be an absolute existing directory.")
	}
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return "", errors.New("Folder must be an absolute existing directory.")
	}
	return path, nil
}

func validIcon(icon string) bool {
	if icon == "" {
		return true
	}
	const prefix = "data:image/png;base64,"
	if len(icon) > 128<<10 || !strings.HasPrefix(icon, prefix) {
		return false
	}
	b, err := base64.StdEncoding.Strict().DecodeString(strings.TrimPrefix(icon, prefix))
	if err != nil {
		return false
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(b))
	if err != nil || cfg.Width > 1024 || cfg.Height > 1024 {
		return false
	}
	_, err = png.Decode(bytes.NewReader(b))
	return err == nil
}

func (a *app) createProject(w http.ResponseWriter, r *http.Request) {
	var body struct{ Name, Folder, Icon string }
	if !decode(w, r, &body) {
		return
	}
	name, ok := cleanLabel(body.Name, 60)
	if !ok {
		fail(w, 400, "Project name must contain 1 to 60 characters.")
		return
	}
	folder, err := existingDirectory(body.Folder)
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	if !validIcon(body.Icon) {
		fail(w, 400, "Icon must be a PNG data URL, at most 128 KiB and 1024 by 1024 pixels.")
		return
	}
	p := Project{ID: newID(), Name: name, Folder: folder, Icon: body.Icon, Folders: []string{}}
	a.mu.Lock()
	for _, previous := range a.state.Projects {
		sameFolder := previous.Folder == folder || (runtime.GOOS == "windows" && strings.EqualFold(previous.Folder, folder))
		if previous.Removed && sameFolder {
			p.ID, p.Folders = previous.ID, previous.Folders
			break
		}
	}
	err = a.commitLocked(func(d *diskState) {
		if previous := d.project(p.ID); previous != nil {
			*previous = p
		} else {
			d.Projects = append(d.Projects, p)
		}
	})
	a.mu.Unlock()
	if err != nil {
		fail(w, 503, err.Error())
		return
	}
	respond(w, 201, p)
}

func (a *app) updateProject(w http.ResponseWriter, r *http.Request) {
	var body struct{ Name, Icon *string }
	if !decode(w, r, &body) {
		return
	}
	if body.Name == nil && body.Icon == nil {
		fail(w, 400, "Provide a project name or icon to update.")
		return
	}
	if body.Name != nil {
		name, ok := cleanLabel(*body.Name, 60)
		if !ok {
			fail(w, 400, "Project name must contain 1 to 60 characters.")
			return
		}
		body.Name = &name
	}
	if body.Icon != nil && !validIcon(*body.Icon) {
		fail(w, 400, "Icon must be a PNG data URL, at most 128 KiB and 1024 by 1024 pixels.")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	id := r.PathValue("id")
	if p := a.state.project(id); p == nil || p.Removed {
		fail(w, 404, "Project not found.")
		return
	}
	if err := a.commitLocked(func(d *diskState) {
		p := d.project(id)
		if body.Name != nil {
			p.Name = *body.Name
		}
		if body.Icon != nil {
			p.Icon = *body.Icon
		}
	}); err != nil {
		fail(w, 503, err.Error())
		return
	}
	respond(w, 200, a.state.project(id))
}

func (a *app) removeProject(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := r.PathValue("id")
	p := a.state.project(id)
	if p == nil {
		fail(w, 404, "Project not found.")
		return
	}
	for _, session := range a.state.Sessions {
		if session.ProjectID == id && a.runs[session.ID] != nil {
			fail(w, 409, "Stop running sessions before removing this project.")
			return
		}
	}
	for _, run := range a.state.GraphRuns {
		if run.ProjectID == id && graphRunActive(run.Status) {
			fail(w, 409, "Project has an active graph run.")
			return
		}
	}
	if !p.Removed {
		if err := a.commitLocked(func(d *diskState) { d.project(id).Removed = true }); err != nil {
			fail(w, 503, err.Error())
			return
		}
	}
	respond(w, 200, map[string]bool{"removed": true})
}

func (a *app) createFolder(w http.ResponseWriter, r *http.Request) {
	var body struct{ Name string }
	if !decode(w, r, &body) {
		return
	}
	name, ok := cleanLabel(body.Name, 60)
	if !ok || strings.EqualFold(name, "Ungrouped") {
		fail(w, 400, "Folder name must contain 1 to 60 characters and cannot be Ungrouped.")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	p := a.state.project(r.PathValue("id"))
	if p == nil || p.Removed {
		fail(w, 404, "Project not found.")
		return
	}
	for _, f := range p.Folders {
		if strings.EqualFold(f, name) {
			fail(w, 409, "Folder already exists.")
			return
		}
	}
	id := p.ID
	if err := a.commitLocked(func(d *diskState) {
		p := d.project(id)
		p.Folders = append(p.Folders, name)
	}); err != nil {
		fail(w, 503, err.Error())
		return
	}
	respond(w, 201, a.state.project(id))
}

func workspace(p *Project, name string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" || name == "Ungrouped" {
		return "Ungrouped", true
	}
	for _, f := range p.Folders {
		if f == name {
			return f, true
		}
	}
	return "", false
}

func (a *app) createSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProjectID string `json:"projectId"`
		Title     string `json:"title"`
		Workspace string `json:"workspace"`
		Harness   string `json:"harness"`
		Model     string `json:"model"`
	}
	if !decode(w, r, &body) {
		return
	}
	title, ok := cleanLabel(body.Title, 200)
	if !ok {
		fail(w, 400, "Session title must contain 1 to 200 characters.")
		return
	}
	model := strings.TrimSpace(body.Model)
	if model != "" {
		if _, ok := cleanLabel(model, 200); !ok || strings.HasPrefix(model, "-") {
			fail(w, 400, "Invalid model identifier.")
			return
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	p := a.state.project(body.ProjectID)
	if p == nil || p.Removed {
		fail(w, 404, "Project not found.")
		return
	}
	group, ok := workspace(p, body.Workspace)
	if !ok {
		fail(w, 400, "Workspace must be Ungrouped or an existing project folder name.")
		return
	}
	if _, ok := a.binaries[body.Harness]; !ok {
		fail(w, 400, "Harness is unknown or not installed.")
		return
	}
	s := Session{ID: newID(), ProjectID: p.ID, Title: title, Workspace: group,
		Harness: body.Harness, Model: model, Status: "idle", Events: []Event{}, CreatedAt: now(), UpdatedAt: now()}
	if err := a.commitLocked(func(d *diskState) {
		d.Sessions = append(d.Sessions, s)
		if s.Harness == "pi" {
			d.Native[s.ID] = nativeSession{Path: filepath.Join(a.dir, "sessions", s.ID+".jsonl")}
		}
	}); err != nil {
		fail(w, 503, err.Error())
		return
	}
	respond(w, 201, a.sessionViewLocked(s.ID))
}

func (a *app) patchSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title     *string `json:"title"`
		Workspace *string `json:"workspace"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Title == nil && body.Workspace == nil {
		fail(w, 400, "Provide title or workspace.")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.state.session(r.PathValue("id"))
	if s == nil {
		fail(w, 404, "Session not found.")
		return
	}
	title, group := s.Title, s.Workspace
	var ok bool
	if body.Title != nil {
		title, ok = cleanLabel(*body.Title, 200)
		if !ok {
			fail(w, 400, "Session title must contain 1 to 200 characters.")
			return
		}
	}
	if body.Workspace != nil {
		group, ok = workspace(a.state.project(s.ProjectID), *body.Workspace)
		if !ok {
			fail(w, 400, "Workspace must be Ungrouped or an existing project folder name.")
			return
		}
	}
	id := s.ID
	if err := a.commitLocked(func(d *diskState) {
		s := d.session(id)
		s.Title, s.Workspace, s.UpdatedAt = title, group, now()
	}); err != nil {
		fail(w, 503, err.Error())
		return
	}
	respond(w, 200, a.sessionViewLocked(id))
}

func (a *app) message(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text     string
		Sources  []SourceReference `json:"sources"`
		Mentions []Mention         `json:"mentions"`
	}
	if !decode(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Text) == "" || len(body.Text) > 128<<10 || !utf8.ValidString(body.Text) || strings.ContainsRune(body.Text, 0) {
		fail(w, 400, "Message must be nonempty UTF-8 text, at most 128 KiB, without NUL characters.")
		return
	}
	a.mu.Lock()
	s := a.state.session(r.PathValue("id"))
	if s == nil {
		a.mu.Unlock()
		fail(w, 404, "Session not found.")
		return
	}
	if p := a.state.project(s.ProjectID); p == nil || p.Removed {
		a.mu.Unlock()
		fail(w, 404, "Project not found.")
		return
	}
	if a.closing {
		a.mu.Unlock()
		fail(w, 503, "Engine is shutting down.")
		return
	}
	if a.runs[s.ID] != nil {
		a.mu.Unlock()
		fail(w, 409, "Session already has a running turn.")
		return
	}
	b, ok := a.binaries[s.Harness]
	if !ok {
		a.mu.Unlock()
		fail(w, 400, "Harness is not installed.")
		return
	}
	id, projectID, harness := s.ID, s.ProjectID, s.Harness
	folder := a.state.project(projectID).Folder
	a.mu.Unlock()
	// Filesystem I/O never holds the application mutex. Recheck ownership and
	// turn state after preparation, before accepting a durable user message.
	cwd, err := existingDirectory(folder)
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	payload, err := prepareMentions(r.Context(), cwd, harness, body.Text, body.Mentions)
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	a.mu.Lock()
	s = a.state.session(id)
	p := a.state.project(projectID)
	if s == nil || p == nil || p.Removed || s.ProjectID != projectID || p.Folder != folder || s.Harness != harness {
		a.mu.Unlock()
		fail(w, 409, "Conversation or project changed. Retry the message.")
		return
	}
	if a.closing || a.storageErr != nil || r.Context().Err() != nil {
		a.mu.Unlock()
		fail(w, 503, "Submission interrupted or engine unavailable. Retry the message.")
		return
	}
	if a.runs[id] != nil {
		a.mu.Unlock()
		fail(w, 409, "Session already has a running turn.")
		return
	}
	prompt, err := a.focusedPrompt(s, body.Text, body.Sources)
	if err != nil {
		a.mu.Unlock()
		fail(w, 400, err.Error())
		return
	}
	payload.Text = linkedPromptPrefix + prompt
	if err := validateSubmissionSize(payload, harness); err != nil {
		a.mu.Unlock()
		fail(w, 400, err.Error())
		return
	}
	payload.Text = prompt
	if err := a.commitLocked(func(d *diskState) {
		s := d.session(id)
		s.Status, s.UpdatedAt = "running", now()
		s.Permissions = nil
		s.Questions = nil
		e := event("user", body.Text)
		e.Data = map[string]any{}
		if len(body.Sources) > 0 {
			e.Data["sources"] = body.Sources
			s.Sources = append(s.Sources, body.Sources...)
		}
		if len(body.Mentions) > 0 {
			e.Data["mentions"] = body.Mentions
			e.Data["mentionPreparation"] = payload.Preparation
		}
		s.Events = append(s.Events, e)
	}); err != nil {
		a.mu.Unlock()
		fail(w, 503, err.Error())
		return
	}
	snapshot := *a.state.session(id)
	response := a.sessionViewLocked(id)
	native := a.state.Native[id]
	ctx, cancel := context.WithCancel(a.ctx)
	t := &turn{ctx: ctx, cancel: cancel, done: make(chan struct{})}
	a.runs[id] = t
	a.wg.Add(1)
	a.mu.Unlock()
	go a.execute(t, snapshot, native, b, cwd, payload)
	respond(w, 202, response)
}

func (a *app) stop(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	a.mu.Lock()
	if a.state.session(id) == nil {
		a.mu.Unlock()
		fail(w, 404, "Session not found.")
		return
	}
	t := a.runs[id]
	if err := a.cancelLinkedLocked(id); err != nil {
		a.mu.Unlock()
		fail(w, 503, err.Error())
		return
	}
	if t != nil {
		t.cancel()
	}
	a.mu.Unlock()
	if t != nil {
		select {
		case <-t.done:
		case <-r.Context().Done():
			return
		case <-time.After(20 * time.Second):
			fail(w, 504, "Cancellation is still in progress.")
			return
		}
	}
	a.mu.Lock()
	s := *a.sessionViewLocked(id)
	err := a.storageErr
	a.mu.Unlock()
	if err != nil {
		fail(w, 503, err.Error())
		return
	}
	respond(w, 200, s)
}

func (a *app) events(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ch := make(chan struct{}, 1)
	a.mu.Lock()
	s := a.sessionViewLocked(id)
	if s == nil {
		a.mu.Unlock()
		fail(w, 404, "Session not found.")
		return
	}
	if a.listeners[id] == nil {
		a.listeners[id] = map[chan struct{}]bool{}
	}
	a.listeners[id][ch] = true
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		delete(a.listeners[id], ch)
		if len(a.listeners[id]) == 0 {
			delete(a.listeners, id)
		}
		a.mu.Unlock()
	}()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("X-Accel-Buffering", "no")
	controller := http.NewResponseController(w)
	send := func(snapshot *Session) error {
		_ = controller.SetWriteDeadline(time.Now().Add(10 * time.Second))
		b, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		if _, err := w.Write(append(append([]byte("data: "), b...), '\n', '\n')); err != nil {
			return err
		}
		return controller.Flush()
	}
	if send(s) != nil {
		return
	}
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			a.mu.Lock()
			s = a.sessionViewLocked(id)
			a.mu.Unlock()
			if send(s) != nil {
				return
			}
		case <-heartbeat.C:
			_ = controller.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
				return
			}
			if controller.Flush() != nil {
				return
			}
		}
	}
}

// Node sessions expose their request reply routes, but are not independent chats.
func (a *app) chatSessionHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		s := a.state.session(r.PathValue("id"))
		private := s != nil && s.GraphRunID != ""
		a.mu.Unlock()
		if private {
			fail(w, 404, "Conversation not found.")
			return
		}
		handler(w, r)
	}
}

func (a *app) getConversationGraph(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.state.session(r.PathValue("id"))
	if s == nil || s.ParentID != "" || s.GraphRunID != "" {
		fail(w, 404, "Main conversation not found.")
		return
	}
	respond(w, 200, a.state.graphProjection(s.ID))
}

func (a *app) selectConversationGraph(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SelectedGraphID *string `json:"selectedGraphId"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.SelectedGraphID == nil {
		fail(w, 400, "Provide selectedGraphId; use an empty string for None.")
		return
	}
	graphID := *body.SelectedGraphID
	if graphID != "" && !validAuthoringID(graphID) {
		fail(w, 400, "Invalid graph identifier.")
		return
	}
	// Same lock ordering as authoring mutations and snapshot capture.
	a.authoringMu.Lock()
	defer a.authoringMu.Unlock()
	a.mu.Lock()
	s := a.state.session(r.PathValue("id"))
	if s == nil || s.ParentID != "" || s.GraphRunID != "" {
		a.mu.Unlock()
		fail(w, 404, "Main conversation not found.")
		return
	}
	p := a.state.project(s.ProjectID)
	if p == nil || p.Removed {
		a.mu.Unlock()
		fail(w, 404, "Project not found.")
		return
	}
	project, sessionID := *p, s.ID
	a.mu.Unlock()
	if graphID != "" {
		catalog, err := loadAuthoring(project)
		if err != nil {
			fail(w, 500, "Cannot read graph catalog: "+err.Error())
			return
		}
		found := false
		for _, graph := range catalog.Graphs {
			if graph.ID == graphID {
				if !graph.Definition.Enabled {
					fail(w, 409, "Graph is disabled.")
					return
				}
				found = true
				break
			}
		}
		if !found {
			if len(catalog.Errors) > 0 {
				fail(w, 409, "Graph is unavailable; resolve catalog errors: "+strings.Join(catalog.Errors, "; "))
				return
			}
			fail(w, 404, "Graph not found.")
			return
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	s = a.state.session(sessionID)
	p = a.state.project(project.ID)
	if a.closing || a.storageErr != nil || s == nil || p == nil || p.Removed || p.Folder != project.Folder {
		fail(w, 409, "Conversation or project changed; retry selection.")
		return
	}
	if s.SelectedGraphID != graphID {
		if err := a.commitLocked(func(d *diskState) {
			session := d.session(sessionID)
			session.SelectedGraphID, session.UpdatedAt = graphID, now()
		}); err != nil {
			fail(w, 503, err.Error())
			return
		}
	}
	respond(w, 200, a.state.graphProjection(sessionID))
}

type GraphRunSummary struct {
	ID             string          `json:"id"`
	ActivityID     string          `json:"activityId"`
	ConversationID string          `json:"conversationId"`
	GraphID        string          `json:"graphId"`
	Status         string          `json:"status"`
	Active         bool            `json:"active"`
	Revision       uint64          `json:"revision"`
	Result         *GraphRunResult `json:"result"`
	CreatedAt      string          `json:"createdAt"`
	EndedAt        string          `json:"endedAt,omitempty"`
}

func (a *app) getGraphRun(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	run := a.state.graphRun(r.PathValue("runID"))
	if run == nil {
		fail(w, 404, "Graph run not found.")
		return
	}
	respond(w, 200, GraphRunSummary{ID: run.ID, ActivityID: run.ActivityID, ConversationID: run.ConversationID, GraphID: run.GraphID, Status: run.Status, Active: graphRunActive(run.Status), Revision: a.state.GraphRevision, Result: run.Result, CreatedAt: run.CreatedAt, EndedAt: run.EndedAt})
}

func (a *app) directory(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(5*time.Minute + 10*time.Second))
	select {
	case a.picker <- struct{}{}:
		defer func() { <-a.picker }()
	default:
		fail(w, 409, "A directory picker is already open on the engine computer. Complete or cancel it, then retry.")
		return
	}
	cmd, err := pickerCommand(ctx)
	if err != nil {
		fail(w, 501, err.Error())
		return
	}
	var output cappedBuffer
	cmd.Stdout = &output
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			fail(w, 408, "Directory picker timed out or was cancelled.")
		} else {
			fail(w, 500, "Could not open the native directory picker.")
		}
		return
	}
	path := strings.TrimPrefix(output.String(), "\ufeff")
	if path == "" {
		respond(w, 200, map[string]any{"path": nil})
		return
	}
	path, err = existingDirectory(path)
	if err != nil || !utf8.ValidString(path) || output.truncated {
		fail(w, 500, "Directory picker returned an invalid path.")
		return
	}
	respond(w, 200, map[string]string{"path": path})
}
