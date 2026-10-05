package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"
)

// Task configuration is durable independently of the execution transports.
// A saved declaration is distinct from ordinary chat's user-event authorization.
type TaskConfig struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Instructions      string   `json:"instructions"`
	SessionFolder     string   `json:"sessionFolder"`
	AllowedGraphIDs   []string `json:"allowedGraphIds"`
	Enabled           bool     `json:"enabled"`
	Trigger           string   `json:"trigger"`
	Days              []string `json:"days"`
	Time              string   `json:"time"`
	Questions         string   `json:"questions"`
	Harness           string   `json:"harness"`
	Model             string   `json:"model"`
	Effort            string   `json:"effort"`
	AllowParallelRuns bool     `json:"allowParallelRuns"`
	RecoverMissedRuns bool     `json:"recoverMissedRuns"`
}

type TaskDefinition struct {
	RuntimeSummary string `json:"runtimeSummary,omitempty"` // HTTP list projection only
	TaskConfig
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	Revision    uint64 `json:"revision"`
	OperationID string `json:"operationId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	DeletedAt   string `json:"deletedAt,omitempty"`
	// This revision's instructions and graph IDs are the saved authorization scope.
	AuthorizationID string `json:"authorizationId"`
}

var taskClock = regexp.MustCompile(`^(?:[01][0-9]|2[0-3]):[0-5][0-9]$`)

// Match HTML maxlength and JavaScript string.length (UTF-16 code units).
func taskTextLength(s string) int { return len(utf16.Encode([]rune(s))) }

func validateTaskConfig(c TaskConfig) error {
	if c.Name == "" || c.Name != strings.TrimSpace(c.Name) || taskTextLength(c.Name) > 80 || strings.IndexFunc(c.Name, unicode.IsControl) >= 0 {
		return errors.New("Enter a task name of at most 80 characters without control characters.")
	}
	if c.Description != strings.TrimSpace(c.Description) || taskTextLength(c.Description) > 240 {
		return errors.New("Description must be at most 240 characters.")
	}
	if c.Instructions == "" || c.Instructions != strings.TrimSpace(c.Instructions) || taskTextLength(c.Instructions) > 12000 {
		return errors.New("Enter instructions of at most 12000 characters.")
	}
	if _, ok := cleanLabel(c.SessionFolder, 80); !ok || c.SessionFolder != strings.TrimSpace(c.SessionFolder) {
		return errors.New("Select a valid session folder.")
	}
	if !slices.Contains([]string{"manual", "schedule", "webhook"}, c.Trigger) || !slices.Contains([]string{"klm", "telegram"}, c.Questions) {
		return errors.New("Select a valid trigger and question routing preference.")
	}
	if !slices.Contains([]string{"opencode", "codex", "pi"}, c.Harness) || strings.TrimSpace(c.Model) == "" || len(c.Model) > 512 || len(c.Effort) > 80 {
		return errors.New("Select a harness and model.")
	}
	seen := map[string]bool{}
	for _, day := range c.Days {
		if seen[day] || !slices.Contains([]string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}, day) {
			return errors.New("Select valid, unique weekdays.")
		}
		seen[day] = true
	}
	if !taskClock.MatchString(c.Time) || c.Trigger == "schedule" && len(c.Days) == 0 {
		return errors.New("Select weekdays and a 24-hour time (HH:mm).")
	}
	seen = map[string]bool{}
	for _, id := range c.AllowedGraphIDs {
		if !validAuthoringID(id) || seen[id] {
			return errors.New("Select valid, unique graph identifiers.")
		}
		seen[id] = true
	}
	return nil
}

func (d *diskState) task(id string) *TaskDefinition {
	for i := range d.Tasks {
		if d.Tasks[i].ID == id {
			return &d.Tasks[i]
		}
	}
	return nil
}

func validateTaskRecords(d *diskState) error {
	if d.TaskSchema != 0 && d.TaskSchema != 1 {
		return errors.New("Unsupported Tasks schema.")
	}
	ids, names, operations := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, t := range d.Tasks {
		key, op := t.ProjectID+"/"+strings.ToLower(t.Name), t.ProjectID+"/"+t.OperationID
		if d.TaskSchema != 1 || t.ID == "" || ids[t.ID] || d.project(t.ProjectID) == nil || t.Revision == 0 || t.OperationID == "" || operations[op] || t.AuthorizationID == "" || !validGraphTime(t.CreatedAt) || !validGraphTime(t.UpdatedAt) || t.DeletedAt != "" && !validGraphTime(t.DeletedAt) || t.DeletedAt == "" && names[key] {
			return fmt.Errorf("Invalid task record %q; refusing to overwrite.", t.ID)
		}
		if err := validateTaskConfig(t.TaskConfig); err != nil {
			return err
		}
		ids[t.ID], operations[op] = true, true
		if t.DeletedAt == "" {
			names[key] = true
		}
	}
	return validateTaskRuntimeRecords(d)
}

// Back up a fully replayed checkpoint, not a stale state.json without its journal.
// The schema marker fences older binaries before the first Tasks mutation.
func migrateTasks(dir string, d *diskState) error {
	if d.TaskSchema == 1 {
		return validateTaskRecords(d)
	}
	if d.TaskSchema != 0 {
		return errors.New("Unsupported Tasks schema.")
	}
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "state.pre-tasks-"+newID()+".json")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	d.TaskSchema = 1
	return saveState(dir, d)
}

func (a *app) listTasks(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	p := a.state.project(r.PathValue("id"))
	if p == nil || p.Removed {
		fail(w, 404, "Project not found.")
		return
	}
	tasks := []TaskDefinition{}
	for _, t := range a.state.Tasks {
		if t.ProjectID == p.ID && t.DeletedAt == "" {
			pending := 0
			var latest *TaskRun
			for i := range a.state.TaskRuns {
				run := &a.state.TaskRuns[i]
				if run.TaskID == t.ID {
					latest = run
					if !taskRunTerminal(run.Status) {
						pending++
					}
				}
			}
			if pending > 0 {
				t.RuntimeSummary = fmt.Sprintf("%d pending", pending)
			} else if latest != nil && (latest.Status == "failed" || latest.Status == "interrupted") {
				t.RuntimeSummary = latest.Status
			}
			tasks = append(tasks, t)
		}
	}
	respond(w, 200, tasks)
}

func (a *app) getTask(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	t := a.state.task(r.PathValue("taskId"))
	if t == nil || t.ProjectID != r.PathValue("id") {
		fail(w, 404, "Task not found.")
		return
	}
	respond(w, 200, t)
}

func (a *app) saveTask(w http.ResponseWriter, r *http.Request) {
	p, ok := a.authoringProject(w, r)
	if !ok {
		return
	}
	var body struct {
		Task        TaskConfig `json:"task"`
		Revision    uint64     `json:"revision"`
		OperationID string     `json:"operationId"`
	}
	if !decode(w, r, &body) {
		return
	}
	c := body.Task
	c.Name, c.Description, c.Instructions = strings.TrimSpace(c.Name), strings.TrimSpace(c.Description), strings.TrimSpace(c.Instructions)
	if c.Days == nil {
		c.Days = []string{}
	}
	if c.AllowedGraphIDs == nil {
		c.AllowedGraphIDs = []string{}
	}
	id := r.PathValue("taskId")
	if id == "" {
		c.Enabled = true
	}
	if err := validateTaskConfig(c); err != nil {
		fail(w, 400, err.Error())
		return
	}
	if id == "" && (strings.TrimSpace(body.OperationID) == "" || len(body.OperationID) > 128) {
		fail(w, 400, "A creation operation ID is required.")
		return
	}
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(55 * time.Second))
	if err := a.validateAgentModel(r, p, c.Harness, c.Model, c.Effort); err != nil {
		fail(w, 400, err.Error())
		return
	}
	a.authoringMu.Lock()
	defer a.authoringMu.Unlock()
	catalog, catalogErr := loadAuthoring(p)
	a.mu.Lock()
	defer a.mu.Unlock()
	currentProject := a.state.project(p.ID)
	if currentProject == nil || currentProject.Removed {
		fail(w, 404, "Project not found.")
		return
	}
	if currentProject.Folder != p.Folder {
		fail(w, 409, "Project directory changed. Reload Tasks before saving.")
		return
	}
	old := a.state.task(id)
	if id != "" && (old == nil || old.ProjectID != p.ID || old.DeletedAt != "") {
		fail(w, 404, "Task not found.")
		return
	}
	if old != nil && old.Revision != body.Revision {
		fail(w, 409, "Task changed. Reopen the editor before saving; your draft has been kept.")
		return
	}
	for _, t := range a.state.Tasks {
		if t.ProjectID != p.ID {
			continue
		}
		if id == "" && t.OperationID == body.OperationID {
			if t.DeletedAt != "" || !reflect.DeepEqual(t.TaskConfig, c) {
				fail(w, 409, "This creation request was already used. Reload Tasks before creating again.")
				return
			}
			respond(w, 200, t)
			return
		}
		if t.ID != id && t.DeletedAt == "" && strings.ToLower(t.Name) == strings.ToLower(c.Name) {
			fail(w, 409, "A task with this name already exists.")
			return
		}
	}
	if c.SessionFolder != "Tasks" && c.SessionFolder != "Ungrouped" && !(slices.Contains(currentProject.Folders, c.SessionFolder) && !slices.Contains(currentProject.ArchivedFolders, c.SessionFolder)) && !(old != nil && old.SessionFolder == c.SessionFolder) {
		fail(w, 400, "Select an active session folder.")
		return
	}
	for _, graphID := range c.AllowedGraphIDs {
		// Existing unavailable references stay visible/removable, never substituted.
		if old != nil && slices.Contains(old.AllowedGraphIDs, graphID) {
			continue
		}
		if catalogErr != nil || len(catalog.Errors) > 0 {
			fail(w, 409, "Resolve graph catalog errors before adding allowed graphs.")
			return
		}
		if !slices.ContainsFunc(catalog.Graphs, func(g GraphRecord) bool { return g.ID == graphID && g.Definition.Enabled }) {
			fail(w, 400, "Select an enabled project graph.")
			return
		}
	}
	t := TaskDefinition{TaskConfig: c, ID: newID(), ProjectID: p.ID, Revision: 1, OperationID: body.OperationID, CreatedAt: now(), UpdatedAt: now(), AuthorizationID: newID()}
	if old != nil {
		t.ID, t.OperationID, t.CreatedAt, t.Revision = old.ID, old.OperationID, old.CreatedAt, old.Revision+1
	}
	if err := a.commitTransactionLocked(func(d *diskState) error {
		if old == nil {
			d.Tasks = append(d.Tasks, t)
		} else {
			*d.task(id) = t
		}
		return nil
	}); err != nil {
		fail(w, 500, err.Error())
		return
	}
	respond(w, 200, t)
}

func (a *app) mutateTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Revision uint64 `json:"revision"`
		Enabled  *bool  `json:"enabled"`
	}
	if !decode(w, r, &body) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	p, t := a.state.project(r.PathValue("id")), a.state.task(r.PathValue("taskId"))
	if p == nil || p.Removed || t == nil || t.ProjectID != p.ID || t.DeletedAt != "" {
		fail(w, 404, "Task not found.")
		return
	}
	if t.Revision != body.Revision {
		fail(w, 409, "Task changed. Reload Tasks and retry.")
		return
	}
	if r.Method != http.MethodDelete && body.Enabled == nil {
		fail(w, 400, "An enabled value is required.")
		return
	}
	if r.Method == http.MethodDelete {
		for _, run := range a.state.TaskRuns {
			if run.TaskID == t.ID && !taskRunTerminal(run.Status) {
				fail(w, 409, "Cannot delete while runs are queued, executing or awaiting attention.")
				return
			}
		}
	}
	var result TaskDefinition
	if err := a.commitTransactionLocked(func(d *diskState) error {
		next := d.task(t.ID)
		if r.Method == http.MethodDelete {
			next.DeletedAt = now()
		} else {
			next.Enabled = *body.Enabled
		}
		next.Revision++
		next.UpdatedAt = now()
		result = *next
		return nil
	}); err != nil {
		fail(w, 500, err.Error())
		return
	}
	respond(w, 200, result)
}
