package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
)

type runnerPreparation struct {
	ProjectID      string           `json:"project_id"`
	InstanceKey    string           `json:"instance_key"`
	Workspace      *runnerWorkspace `json:"workspace"`
	Capture        runnerCapture    `json:"capture"`
	Source         *AgentCapture    `json:"source,omitempty"`
	RecoveryPolicy string           `json:"recovery_policy,omitempty"`
}

func runnerGitEnvironment() []string {
	env := []string{}
	for _, value := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(value), "GIT_") {
			env = append(env, value)
		}
	}
	return env
}

func (a *app) runnerPrepareHTTP(w http.ResponseWriter, r *http.Request) {
	owner, err := a.runnerOwner(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var body struct {
		OperationID    string          `json:"operation_id"`
		ProjectID      string          `json:"project_id"`
		InstanceKey    string          `json:"instance_key"`
		Workspace      runnerWorkspace `json:"workspace"`
		Capture        runnerCapture   `json:"capture"`
		Agent          string          `json:"agent"`
		SourceUID      string          `json:"source_uid"`
		AgentRevision  string          `json:"agent_revision"`
		RecoveryPolicy string          `json:"recovery_policy"`
	}
	if !decode(w, r, &body) {
		return
	}
	digest := revision(append([]byte("prepare/"), runnerRawJSON(body)...))
	var receipt runnerOperation
	err = a.store.readSnapshot(r.Context(), func(tx *sql.Tx) error {
		var err error
		receipt, err = (&sqliteSessionTx{ctx: r.Context(), tx: tx}).runnerReceipt(owner, body.OperationID, digest, nil)
		return err
	})
	if err == nil {
		respond(w, 200, receipt)
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		sqlHTTPError(w, err)
		return
	}
	if body.InstanceKey == "" || len(body.InstanceKey) > 200 {
		fail(w, 400, "Supply a bounded integration-owned instance key.")
		return
	}
	if body.RecoveryPolicy != "" && body.RecoveryPolicy != "resume_drained" {
		fail(w, 400, "Unsupported runner recovery policy.")
		return
	}
	err = a.store.readSnapshot(r.Context(), func(tx *sql.Tx) error {
		return (&sqliteSessionTx{ctx: r.Context(), tx: tx}).runnerGroupOpen(owner, body.ProjectID, body.InstanceKey)
	})
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	project, err := a.store.readProjectMeta(r.Context(), body.ProjectID)
	if err != nil || project.Removed {
		fail(w, 404, "Project not found.")
		return
	}
	workspace, err := inspectRunnerWorkspace(r.Context(), project, body.Workspace)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	prepared := runnerPreparation{
		ProjectID: project.ID, InstanceKey: body.InstanceKey,
		Workspace: workspace, RecoveryPolicy: body.RecoveryPolicy,
	}
	if body.Agent != "" {
		if body.SourceUID == "" || body.AgentRevision == "" {
			fail(w, 400, "Saved participant preparation requires the exact source identity and revision.")
			return
		}
		resolved, err := a.resolveNewAgent(r.Context(), NewAgentRequest{
			PreserveSource: true, ProjectID: project.ID, Agent: body.Agent,
			SourceUID: body.SourceUID, Revision: body.AgentRevision, ExecutionCWD: workspace.Path,
		})
		if err != nil {
			sqlHTTPError(w, err)
			return
		}
		if resolved.Capture.Config.Harness == "" || resolved.Capture.Config.Model == "" {
			fail(w, 409, "The saved worker Agent must have an explicit harness and model.")
			return
		}
		prepared.Source = &resolved.Capture
		body.Capture.Harness = resolved.Config.Harness
		body.Capture.Model = resolved.Config.Model
		body.Capture.Effort = resolved.Config.Effort
		body.Capture.Selection = resolved.Config.Selection
		body.Capture.Instructions = resolved.Config.Prompt + "\n\n" + body.Capture.Instructions
	}
	project.Folder = workspace.Path
	prepared.Capture, err = a.validateRunnerCapture(r.Context(), body.Capture, project)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	if prepared.Capture.ConfigurationGate {
		fail(w, 400, "Prepared participant captures are immutable and cannot use a configuration gate.")
		return
	}
	a.mu.Lock()
	err = a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		if err := st.runnerGroupOpen(owner, prepared.ProjectID, prepared.InstanceKey); err != nil {
			return err
		}
		receipt, err = st.runnerReceipt(owner, body.OperationID, digest, prepared)
		return err
	})
	a.mu.Unlock()
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 200, receipt)
}

func (a *app) readRunnerPreparation(r *http.Request, owner, operationID string) (runnerPreparation, error) {
	var prepared runnerPreparation
	var data []byte
	err := a.store.readers.QueryRowContext(r.Context(),
		`SELECT body FROM runner_operations WHERE integration_id=? AND id=?`, owner, operationID,
	).Scan(&data)
	if err != nil {
		return prepared, err
	}
	var receipt runnerOperation
	if err := json.Unmarshal(data, &receipt); err != nil {
		return prepared, err
	}
	if err := json.Unmarshal(receipt.Result, &prepared); err != nil {
		return prepared, err
	}
	if prepared.Workspace == nil || prepared.ProjectID == "" || prepared.InstanceKey == "" {
		return prepared, controlError(409, "Operation is not an immutable participant preparation.")
	}
	return prepared, nil
}
