package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (a *app) runnerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/runner/capabilities", a.runnerCapabilities)
	mux.HandleFunc("POST /api/runner/connections", a.runnerConnect)
	mux.HandleFunc("POST /api/runner/captures/validate", a.runnerValidateCaptureHTTP)
	mux.HandleFunc("POST /api/runner/captures", a.runnerPrepareHTTP)
	mux.HandleFunc("POST /api/runner/groups/{instanceKey}/cancel", a.runnerGroupHTTP)
	mux.HandleFunc("GET /api/runner/groups/{instanceKey}", a.runnerGroupHTTP)
	mux.HandleFunc("GET /api/runner/connections/{integrationId}", func(w http.ResponseWriter, r *http.Request) {
		owner, err := a.runnerOwner(r)
		if err != nil {
			sqlHTTPError(w, err)
			return
		}
		if owner != r.PathValue("integrationId") {
			fail(w, 404, "Integration not found.")
			return
		}
		respond(w, 200, map[string]string{"integration_id": owner, "protocol": "1"})
	})
	mux.HandleFunc("POST /api/runner/conversations", func(w http.ResponseWriter, r *http.Request) { a.runnerCreate(w, r, "coordinator") })
	mux.HandleFunc("POST /api/runner/work", func(w http.ResponseWriter, r *http.Request) { a.runnerCreate(w, r, "worker") })
	for _, prefix := range []string{"conversations", "work"} {
		mux.HandleFunc("GET /api/runner/"+prefix+"/{runnerId}", a.runnerGet)
		mux.HandleFunc("GET /api/runner/"+prefix+"/{runnerId}/content", a.runnerContent)
	}
	mux.HandleFunc("POST /api/runner/conversations/{runnerId}/inputs", a.runnerInputHTTP)
	mux.HandleFunc("GET /api/runner/conversations/{runnerId}/inputs", a.runnerInputLookup)
	mux.HandleFunc("POST /api/runner/work/{runnerId}/inputs", a.runnerInputHTTP)
	mux.HandleFunc("GET /api/runner/work/{runnerId}/inputs", a.runnerInputLookup)
	mux.HandleFunc("DELETE /api/runner/conversations/{runnerId}/inputs/{inputId}", a.runnerRemoveInput)
	mux.HandleFunc("PATCH /api/runner/conversations/{runnerId}/configuration", a.runnerConfiguration)
	mux.HandleFunc("POST /api/runner/conversations/{runnerId}/stop", a.runnerStopHTTP)
	mux.HandleFunc("POST /api/runner/work/{runnerId}/cancel", a.runnerStopHTTP)
	mux.HandleFunc("POST /api/runner/tool-calls/{callId}/reply", a.runnerToolReply)
	mux.HandleFunc("POST /api/runner/tool-calls/{callId}/admit", a.runnerAdmitEffect)
	mux.HandleFunc("POST /api/runner/requests/{requestId}/reply", a.runnerRequestReply)
	for _, prefix := range []string{"conversations", "work"} {
		mux.HandleFunc("GET /api/runner/"+prefix+"/{runnerId}/history", a.runnerHistory)
	}
	mux.HandleFunc("GET /api/runner/events", a.runnerEvents)
	mux.HandleFunc("GET /api/runner/jobs/{jobId}/content", a.runnerJobContent)
	mux.HandleFunc("GET /api/runner/updates", a.runnerUpdates)
	mux.HandleFunc("GET /api/runner/operations/{operationId}", a.runnerOperationHTTP)
}

func (a *app) runnerOwner(r *http.Request) (string, error) {
	owner := r.Header.Get("X-KLM-Runner-ID")
	var client string
	if a.runner == nil {
		return "", controlError(503, "Runner is unavailable.")
	}
	err := a.store.readers.QueryRowContext(r.Context(), `SELECT client_id FROM runner_integrations WHERE id=?`, owner).Scan(&client)
	if err != nil || client != clientIdentity(r) {
		return "", controlError(403, "Runner integration is not bound to this client.")
	}
	return owner, nil
}

func (a *app) runnerConnect(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OperationID   string `json:"operation_id"`
		Key           string `json:"key"`
		IntegrationID string `json:"integration_id"`
	}
	if !decode(w, r, &body) {
		return
	}
	if len(body.Key) < 16 || len(body.Key) > 200 {
		fail(w, 400, "A stable integration key is required.")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	id := newID()
	encoded, _ := json.Marshal(body)
	digest := revision(encoded)
	scope := "connection/" + clientIdentity(r)
	var receipt runnerOperation
	err := a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		var err error
		receipt, err = st.runnerReceipt(scope, body.OperationID, digest, nil)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		err = st.tx.QueryRowContext(st.ctx, `SELECT id FROM runner_integrations WHERE external_key=?`, body.Key).Scan(&id)
		if err == nil {
			if body.IntegrationID != "" && body.IntegrationID != id {
				return controlError(409, "Integration identity changed.")
			}
			_, err = st.tx.ExecContext(st.ctx, `UPDATE runner_integrations SET client_id=? WHERE id=?`, clientIdentity(r), id)
			if err != nil {
				return err
			}
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		} else {
			if body.IntegrationID != "" {
				return controlError(409, "Retained integration no longer exists.")
			}
			if _, err := st.tx.ExecContext(st.ctx, `INSERT INTO runner_integrations VALUES(?,?,?)`, id, clientIdentity(r), body.Key); err != nil {
				return err
			}
		}
		receipt, err = st.runnerReceipt(scope, body.OperationID, digest, map[string]string{"integration_id": id, "protocol": "1"})
		return err
	})
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 200, json.RawMessage(receipt.Result))
}

func (a *app) runnerCreate(w http.ResponseWriter, r *http.Request, kind string) {
	owner, err := a.runnerOwner(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var body struct {
		OperationID   string        `json:"operation_id"`
		Key           string        `json:"key"`
		ProjectID     string        `json:"project_id"`
		Capture       runnerCapture `json:"capture"`
		Agent         string        `json:"agent"`
		SourceUID     string        `json:"source_uid"`
		AgentRevision string        `json:"agent_revision"`
		Prompt        string        `json:"prompt"`
		PreparationID string        `json:"preparation_id"`
	}
	if !decode(w, r, &body) {
		return
	}
	encoded, _ := json.Marshal(body)
	digest := revision(append([]byte("create/"+kind+"/"), encoded...))
	var previous []byte
	if err := a.store.readers.QueryRowContext(r.Context(), `SELECT body FROM runner_operations WHERE integration_id=? AND id=?`, owner, body.OperationID).Scan(&previous); err == nil {
		var receipt runnerOperation
		if err := json.Unmarshal(previous, &receipt); err != nil {
			sqlHTTPError(w, err)
			return
		}
		if receipt.Digest != digest {
			fail(w, 409, "Operation identity was reused with another body.")
			return
		}
		respond(w, 202, receipt)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		sqlHTTPError(w, err)
		return
	}
	if body.Key == "" || len(body.Key) > 200 {
		fail(w, 400, "A stable conversation or Assignment key is required.")
		return
	}
	project, err := a.store.readProjectMeta(r.Context(), body.ProjectID)
	if err != nil || project.Removed {
		fail(w, 404, "Project not found.")
		return
	}
	if body.PreparationID == "" {
		project.Folder, err = runnerCanonicalDirectory(project.Folder)
		if err != nil {
			sqlHTTPError(w, err)
			return
		}
	}
	var source *AgentCapture
	var prepared runnerPreparation
	if body.PreparationID != "" {
		captureSupplied := revision(runnerCaptureJSON(body.Capture)) != revision(runnerCaptureJSON(runnerCapture{}))
		if kind != "worker" || body.Agent != "" || body.SourceUID != "" || body.AgentRevision != "" || captureSupplied {
			fail(w, 400, "Prepared work uses only its retained participant capture.")
			return
		}
		prepared, err = a.readRunnerPreparation(r, owner, body.PreparationID)
		if err != nil {
			sqlHTTPError(w, err)
			return
		}
		if prepared.ProjectID != project.ID {
			fail(w, 409, "Preparation belongs to another project.")
			return
		}
		if err := a.runner.groupOpen(r.Context(), runnerExecution{IntegrationID: owner, ProjectID: project.ID, InstanceKey: prepared.InstanceKey}); err != nil {
			sqlHTTPError(w, err)
			return
		}
		if err := validateRunnerLaunchDirectory(project, prepared.Workspace.Path); err != nil {
			sqlHTTPError(w, err)
			return
		}
		project.Folder = prepared.Workspace.Path
		source = prepared.Source
		body.Capture = prepared.Capture
	} else if kind == "worker" {
		resolved, resolveErr := a.resolveNewAgent(r.Context(), NewAgentRequest{PreserveSource: true, ProjectID: body.ProjectID, Agent: body.Agent, SourceUID: body.SourceUID, Revision: body.AgentRevision, ExecutionCWD: project.Folder})
		if resolveErr != nil {
			sqlHTTPError(w, resolveErr)
			return
		}
		if resolved.Capture.Config.Harness == "" || resolved.Capture.Config.Model == "" {
			fail(w, 409, "The saved worker Agent must have an explicit harness and model.")
			return
		}
		source = &resolved.Capture
		body.Capture.Harness, body.Capture.Model, body.Capture.Effort = resolved.Config.Harness, resolved.Config.Model, resolved.Config.Effort
		body.Capture.Selection = resolved.Config.Selection
		body.Capture.Instructions = resolved.Config.Prompt + "\n\n" + body.Capture.Instructions
	}
	capture := body.Capture
	if body.PreparationID == "" {
		capture, err = a.validateRunnerCapture(r.Context(), body.Capture, project)
		if err != nil {
			sqlHTTPError(w, err)
			return
		}
	}
	execution := runnerExecution{ID: newID(), IntegrationID: owner, ProjectID: project.ID, Key: body.Key, Kind: kind, SessionID: newID(), Generation: "1", Revision: "1", State: "idle", Desired: capture, Applied: capture, Source: source}
	execution.PreparationID = body.PreparationID
	execution.InstanceKey = prepared.InstanceKey
	execution.Workspace = prepared.Workspace
	execution.RecoveryPolicy = prepared.RecoveryPolicy
	execution.ProjectGit = prepared.ProjectGit
	var receipt runnerOperation
	a.mu.Lock()
	if err := a.bindRunnerProjectGit(r.Context(), &execution); err != nil {
		a.mu.Unlock()
		sqlHTTPError(w, err)
		return
	}
	err = a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		if previous, err := st.runnerReceipt(owner, body.OperationID, digest, nil); err == nil {
			receipt = previous
			return json.Unmarshal(previous.Result, &execution)
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		var existing string
		err := st.tx.QueryRowContext(st.ctx, `SELECT id FROM runner_executions WHERE integration_id=? AND kind=? AND external_key=?`, owner, kind, body.Key).Scan(&existing)
		if err == nil {
			var err error
			execution, err = st.runnerExecution(existing)
			if err != nil {
				return err
			}
			if execution.ProjectID != body.ProjectID {
				return controlError(409, "External key is already bound to another project.")
			}
			if execution.PreparationID != body.PreparationID {
				return controlError(409, "External key is already bound to another participant capture.")
			}
		} else {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if err := st.runnerGroupOpen(owner, project.ID, execution.InstanceKey); err != nil {
				return err
			}
			session := Session{ID: execution.SessionID, Origin: runnerOrigin, IntegrationID: owner, RunnerID: execution.ID, Role: "runner_" + kind, ProjectID: project.ID, Title: "External execution", Harness: capture.Harness, Model: capture.Model, Effort: capture.Effort, YOLO: true, Status: "idle", ExecutionCWD: project.Folder, CreatedAt: now(), UpdatedAt: now(), Capabilities: ConversationCapabilities{Desired: capture.Selection}}
			if err := st.createSession(session, a.dir); err != nil {
				return err
			}
			if err := st.putRunnerExecution(execution); err != nil {
				return err
			}
			if kind == "worker" {
				if strings.TrimSpace(body.Prompt) == "" {
					return controlError(400, "Worker task is required.")
				}
				if err := a.runner.enqueueTx(st, execution, runnerInput{ID: newID(), SourceKind: "assignment", SourceID: body.Key, Mode: "steer", Text: body.Prompt}); err != nil {
					return err
				}
			}
		}
		var err2 error
		receipt, err2 = st.runnerReceipt(owner, body.OperationID, digest, execution)
		return err2
	})
	if err == nil && kind == "worker" {
		err = a.runner.scheduleLocked(execution.ID)
	}
	a.mu.Unlock()
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 202, receipt)
}

func (a *app) validateRunnerCapture(ctx context.Context, capture runnerCapture, project Project) (runnerCapture, error) {
	if capture.Model == "" || !managedLibraryHarness(capture.Harness) {
		return capture, controlError(400, "Choose an explicit supported harness and model.")
	}
	for _, tool := range capture.Manifest {
		name := str(tool, "name")
		for _, reserved := range []string{"shell_start", "job_read", "job_wait", "job_cancel", "task", "subagent", "bash", "powershell", "shell", "exec_command", "write_stdin"} {
			if name == reserved {
				return capture, controlError(400, "External manifest uses a reserved execution tool name.")
			}
		}
		if name == "" || len(name) > 100 || strings.HasPrefix(name, "__") || strings.HasPrefix(name, "project_") || strings.HasPrefix(name, "session_") || strings.HasPrefix(name, "graph_") || strings.HasPrefix(name, "linked_") || strings.HasPrefix(name, "automation_") || name == "subagent_run" {
			return capture, controlError(400, "External manifest contains an invalid or reserved tool name.")
		}
		if object(tool["inputSchema"]) == nil {
			return capture, controlError(400, "Tool inputSchema is required.")
		}
	}
	if len(capture.Manifest) > 64 || len(capture.Instructions) > 128<<10 {
		return capture, controlError(400, "External capture exceeds limits.")
	}
	config, err := a.validateResolvedAgent(ctx, ResolvedAgentConfig{Harness: capture.Harness, Model: capture.Model, Effort: capture.Effort}, project.ID, project.Folder)
	if err != nil {
		return capture, err
	}
	capture.Effort, capture.YOLO = config.Effort, true
	capture.Capabilities, err = a.captureCapabilities(ctx, capture.Selection, capture.Harness)
	if err == nil && capture.Capabilities != nil {
		for _, connection := range capture.Capabilities.Connections {
			if connection.Catalog == nil {
				return capture, controlError(409, "Refresh the selected MCP catalog before admission.")
			}
			for _, name := range connection.Selection.Tools {
				found := false
				for _, tool := range connection.Catalog.Tools {
					if tool != nil && tool.Name == name {
						found = true
						break
					}
				}
				if !found {
					return capture, controlError(409, "A selected MCP tool is unavailable in the captured catalog.")
				}
			}
		}
	}
	return capture, err
}

func (a *app) runnerValidateCaptureHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := a.runnerOwner(r); err != nil {
		sqlHTTPError(w, err)
		return
	}
	var body struct {
		ProjectID string        `json:"project_id"`
		Capture   runnerCapture `json:"capture"`
	}
	if !decode(w, r, &body) {
		return
	}
	project, err := a.store.readProjectMeta(r.Context(), body.ProjectID)
	if err != nil || project.Removed {
		fail(w, 404, "Project not found.")
		return
	}
	if _, err := a.validateRunnerCapture(r.Context(), body.Capture, project); err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 200, map[string]bool{"validated": true})
}

func (a *app) runnerOwned(r *http.Request) (runnerExecution, error) {
	owner, err := a.runnerOwner(r)
	if err != nil {
		return runnerExecution{}, err
	}
	execution, err := a.store.readRunnerExecution(r.Context(), r.PathValue("runnerId"))
	if err == nil && execution.IntegrationID != owner {
		err = sql.ErrNoRows
	}
	return execution, err
}

func (a *app) runnerGet(w http.ResponseWriter, r *http.Request) {
	execution, err := a.runnerOwned(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	inputs := []runnerInput{}
	requests := []json.RawMessage{}
	rows, err := a.store.readers.QueryContext(r.Context(), `SELECT body,state FROM runner_inputs WHERE execution_id=? ORDER BY position DESC LIMIT 32`, execution.ID)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	for rows.Next() {
		var body []byte
		var state string
		if err = rows.Scan(&body, &state); err != nil {
			break
		}
		var input runnerInput
		if err = json.Unmarshal(body, &input); err != nil {
			break
		}
		input.State = state
		inputs = append(inputs, input)
	}
	rows.Close()
	if err == nil {
		rows, err = a.store.readers.QueryContext(r.Context(), `SELECT body,state FROM runner_requests WHERE execution_id=? ORDER BY CASE WHEN state IN ('pending','reply_recorded','delivery_failed') THEN 0 ELSE 1 END,rowid DESC LIMIT 32`, execution.ID)
		if err == nil {
			for rows.Next() {
				var body []byte
				var state string
				if err = rows.Scan(&body, &state); err != nil {
					break
				}
				var value map[string]any
				if err = json.Unmarshal(body, &value); err != nil {
					break
				}
				value["state"] = state
				requests = append(requests, runnerRawJSON(value))
			}
			rows.Close()
		}
	}
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	attention, jobs, err := a.runnerAttention(execution)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var native nativeSession
	err = a.store.readSnapshot(r.Context(), func(tx *sql.Tx) error {
		var err error
		native, err = (&sqliteSessionTx{ctx: r.Context(), tx: tx}).native(execution.SessionID)
		return err
	})
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 200, map[string]any{
		"execution": execution, "inputs": inputs, "requests": requests,
		"attention_only": attention, "owned_jobs_running": jobs,
		"native_context": map[string]string{"id": native.ID},
	})
}

func runnerRawJSON(value any) json.RawMessage { body, _ := json.Marshal(value); return body }

func (a *app) runnerInputLookup(w http.ResponseWriter, r *http.Request) {
	execution, err := a.runnerOwned(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var body []byte
	var state string
	err = a.store.readers.QueryRowContext(r.Context(), `SELECT body,state FROM runner_inputs WHERE execution_id=? AND source_kind=? AND source_id=?`, execution.ID, r.URL.Query().Get("source_kind"), r.URL.Query().Get("source_id")).Scan(&body, &state)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var input runnerInput
	if err := json.Unmarshal(body, &input); err != nil {
		sqlHTTPError(w, err)
		return
	}
	input.State = state
	respond(w, 200, input)
}

func (s *runnerService) enqueueTx(st *sqliteSessionTx, execution runnerExecution, input runnerInput) error {
	if err := st.runnerGroupOpen(execution.IntegrationID, execution.ProjectID, execution.InstanceKey); err != nil {
		return err
	}
	if len(input.Text) > 128<<10 || input.SourceID == "" {
		return controlError(400, "A bounded attributed input is required.")
	}
	if input.Mode != "steer" && input.Mode != "queue" && input.Mode != "passive" {
		return controlError(400, "Input mode must be steer, queue or passive.")
	}
	if input.SourceKind != "human" && input.SourceKind != "application_event" && input.SourceKind != "tool_result" && input.SourceKind != "assignment" && input.SourceKind != "job_result" {
		return controlError(400, "Invalid input provenance.")
	}
	var previous []byte
	err := st.tx.QueryRowContext(st.ctx, `SELECT body FROM runner_inputs WHERE execution_id=? AND source_kind=? AND source_id=?`, execution.ID, input.SourceKind, input.SourceID).Scan(&previous)
	if err == nil {
		var old runnerInput
		if err := json.Unmarshal(previous, &old); err != nil {
			return err
		}
		if old.Text != input.Text || old.Mode != input.Mode || revision(runnerMentionsJSON(old.Mentions)) != revision(runnerMentionsJSON(input.Mentions)) {
			return controlError(409, "Source input already has another body.")
		}
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	input.State = "retained"
	if input.ID == "" {
		input.ID = newID()
	}
	if input.prepared != nil && len(input.Mentions) > 0 {
		if err := st.putRoot("queue_payloads", input.ID, queuedPayload{Submission: *input.prepared, Harness: execution.Applied.Harness}); err != nil {
			return err
		}
	}
	body, err := runnerJSON(input)
	if err != nil {
		return err
	}
	_, err = st.tx.ExecContext(st.ctx, `INSERT INTO runner_inputs SELECT ?,?,?,?,coalesce(max(position),0)+1,?,? FROM runner_inputs WHERE execution_id=?`, input.ID, execution.ID, input.SourceKind, input.SourceID, input.State, body, execution.ID)
	if err != nil {
		return err
	}
	if s.app.runs[execution.SessionID] != nil && (input.Mode == "steer" || input.Mode == "passive") {
		q := QueuedMessage{ID: input.ID, Text: input.Text, Mode: input.Mode, Status: "steering", RunnerSource: &input, Mentions: input.Mentions}
		if err := st.putRecord("queued_messages", sqliteOwnedKey(execution.SessionID, input.ID), execution.SessionID, input.ID, -1, q); err != nil {
			return err
		}
		text := "Input " + input.ID + " (" + input.SourceKind + ", source " + input.SourceID + "):\n" + input.Text
		if input.prepared != nil {
			text = "Input " + input.ID + " (" + input.SourceKind + ", source " + input.SourceID + "):\n" + input.prepared.piText()
		}
		prepared := submission{Text: text}
		if input.prepared != nil {
			prepared.Preparation = input.prepared.Preparation
		}
		if err := st.putRoot("queue_payloads", input.ID, queuedPayload{Submission: prepared, Harness: execution.Applied.Harness}); err != nil {
			return err
		}
	}
	return st.runnerEvent(execution, "input", input)
}

func (s *runnerService) enqueueLocked(execution runnerExecution, input runnerInput) error {
	return s.app.writeSQLLocked(s.app.ctx, func(st *sqliteSessionTx) error { return s.enqueueTx(st, execution, input) })
}

func (a *app) runnerInputHTTP(w http.ResponseWriter, r *http.Request) {
	execution, err := a.runnerOwned(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var input runnerInput
	if !decode(w, r, &input) {
		return
	}
	if input.ID != "" || input.State != "" || input.TurnID != "" {
		fail(w, 400, "Input identity and delivery state are server-owned.")
		return
	}
	if execution.Kind == "worker" && input.SourceKind == "human" {
		fail(w, 400, "Workers accept attributed application continuations; native answers use the exact request endpoint.")
		return
	}
	project, err := a.store.readProjectMeta(r.Context(), execution.ProjectID)
	if err != nil || project.Removed {
		fail(w, 404, "Project not found.")
		return
	}
	_, cwd, err := a.readSQLConversation(r.Context(), execution.SessionID)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	if err := validateRunnerLaunchDirectory(project, cwd); err != nil {
		sqlHTTPError(w, err)
		return
	}
	prepared, err := prepareMentions(r.Context(), cwd, "pi", input.Text, input.Mentions)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	input.prepared = &prepared
	encoded, _ := json.Marshal(input)
	a.mu.Lock()
	var receipt runnerOperation
	err = a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		var err error
		receipt, err = st.runnerReceipt(execution.IntegrationID, input.OperationID, revision(append([]byte(execution.ID+"/"), encoded...)), nil)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		current, err := st.runnerExecution(execution.ID)
		if err != nil {
			return err
		}
		workerUnavailable := current.CancelRequested || current.State == "cancelled" || current.State == "cancelling" || current.State == "interrupted"
		if current.Kind == "worker" && workerUnavailable {
			return controlError(409, "This worker cannot continue without reconciliation.")
		}
		if current.Kind == "worker" && current.State == "native_completed" {
			current.State = "idle"
			if err := st.putRunnerExecution(current); err != nil {
				return err
			}
		}
		execution = current
		input.ID = newID()
		if err := a.runner.enqueueTx(st, execution, input); err != nil {
			return err
		}
		receipt, err = st.runnerReceipt(execution.IntegrationID, input.OperationID, revision(append([]byte(execution.ID+"/"), encoded...)), input)
		return err
	})
	if err == nil {
		a.wakeSteeringLocked(execution.SessionID)
		err = a.runner.scheduleLocked(execution.ID)
	}
	a.mu.Unlock()
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 202, receipt)
}

func runnerMentionsJSON(mentions []Mention) []byte {
	body, _ := json.Marshal(append([]Mention{}, mentions...))
	return body
}

func (a *app) runnerEvents(w http.ResponseWriter, r *http.Request) {
	owner, err := a.runnerOwner(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	after, err := strconv.ParseUint(r.URL.Query().Get("after"), 10, 64)
	if r.URL.Query().Get("after") == "" {
		after = 0
		err = nil
	}
	if err != nil {
		fail(w, 400, "Invalid cursor.")
		return
	}
	rows, err := a.store.readers.QueryContext(r.Context(), `SELECT seq,id,execution_id,type,at,body FROM runner_events WHERE integration_id=? AND seq>? ORDER BY seq LIMIT 100`, owner, after)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	defer rows.Close()
	events := []runnerEvent{}
	next := strconv.FormatUint(after, 10)
	bytes := 0
	for rows.Next() {
		var seq uint64
		var event runnerEvent
		if err := rows.Scan(&seq, &event.ID, &event.ExecutionID, &event.Type, &event.At, &event.Data); err != nil {
			sqlHTTPError(w, err)
			return
		}
		event.Cursor = strconv.FormatUint(seq, 10)
		if len(events) > 0 && bytes+len(event.Data) > 2<<20 {
			break
		}
		bytes += len(event.Data)
		next = event.Cursor
		events = append(events, event)
	}
	respond(w, 200, map[string]any{"events": events, "next_after": next})
}

func (a *app) runnerUpdates(w http.ResponseWriter, r *http.Request) {
	if _, err := a.runnerOwner(r); err != nil {
		sqlHTTPError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}
	for {
		if _, err := w.Write([]byte("event: wake\ndata: {}\n\n")); err != nil {
			return
		}
		flusher.Flush()
		select {
		case <-a.runner.changed:
		case <-time.After(15 * time.Second):
		case <-r.Context().Done():
			return
		}
		if _, err := a.runnerOwner(r); err != nil {
			return
		}
	}
}

func (a *app) runnerOperationHTTP(w http.ResponseWriter, r *http.Request) {
	owner, err := a.runnerOwner(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var body []byte
	err = a.store.readers.QueryRowContext(r.Context(), `SELECT body FROM runner_operations WHERE integration_id=? AND id=?`, owner, r.PathValue("operationId")).Scan(&body)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 200, json.RawMessage(body))
}
