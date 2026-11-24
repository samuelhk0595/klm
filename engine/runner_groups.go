package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type runnerGroup struct {
	InstanceKey     string            `json:"instance_key"`
	ProjectID       string            `json:"project_id"`
	CancelRequested bool              `json:"cancel_requested"`
	Executions      []runnerExecution `json:"executions"`
}

func runnerGroupRecordID(key string) string { return "__runner_group/" + revision([]byte(key)) }

// A group has only a durable ownership/cancellation receipt. Participant capture
// and execution records remain the authorities for all execution data.
func (st *sqliteSessionTx) runnerGroup(owner, project, key string) (runnerGroup, error) {
	group := runnerGroup{InstanceKey: key, ProjectID: project, Executions: []runnerExecution{}}
	if key == "" || len(key) > 200 || strings.HasPrefix(key, "__") {
		return group, controlError(400, "A bounded instance key is required.")
	}
	var body []byte
	err := st.tx.QueryRowContext(st.ctx, `SELECT body FROM runner_operations WHERE integration_id=? AND id=?`, owner, runnerGroupRecordID(key)).Scan(&body)
	if err == nil {
		var receipt runnerOperation
		if err := json.Unmarshal(body, &receipt); err != nil {
			return group, err
		}
		if err := json.Unmarshal(receipt.Result, &group); err != nil {
			return group, err
		}
		if group.ProjectID != project {
			return group, controlError(409, "Instance key is already bound to another project.")
		}
		group.Executions = []runnerExecution{}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return group, err
	}
	// Preparations also bind the key before the first worker exists. Older Slice 0
	// preparations are covered without a migration or a second capture authority.
	var conflict bool
	err = st.tx.QueryRowContext(st.ctx, `SELECT EXISTS(SELECT 1 FROM runner_operations
		WHERE integration_id=? AND json_extract(CAST(body AS TEXT),'$.result.instance_key')=?
		AND json_extract(CAST(body AS TEXT),'$.result.project_id')<>?)`, owner, key, project).Scan(&conflict)
	if err != nil {
		return group, err
	}
	if conflict {
		return group, controlError(409, "Instance key is already bound to another project.")
	}
	rows, err := st.tx.QueryContext(st.ctx, `SELECT body FROM runner_executions WHERE integration_id=?
		AND json_extract(CAST(body AS TEXT),'$.instance_key')=? ORDER BY id`, owner, key)
	if err != nil {
		return group, err
	}
	defer rows.Close()
	for rows.Next() {
		var data []byte
		var execution runnerExecution
		if err := rows.Scan(&data); err != nil {
			return group, err
		}
		if err := json.Unmarshal(data, &execution); err != nil {
			return group, err
		}
		if execution.ProjectID != project {
			return group, controlError(409, "Instance key is already bound to another project.")
		}
		group.Executions = append(group.Executions, execution)
	}
	return group, rows.Err()
}

func (st *sqliteSessionTx) runnerGroupOpen(owner, project, key string) error {
	if key == "" {
		return nil
	}
	group, err := st.runnerGroup(owner, project, key)
	if err != nil {
		return err
	}
	if group.CancelRequested {
		return controlError(409, "This runner instance is durably cancelled.")
	}
	return nil
}

func (a *app) runnerGroupHTTP(w http.ResponseWriter, r *http.Request) {
	owner, err := a.runnerOwner(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	key := r.PathValue("instanceKey")
	projectID := r.URL.Query().Get("project_id")
	var body struct {
		OperationID string `json:"operation_id"`
		ProjectID   string `json:"project_id"`
	}
	if r.Method == http.MethodPost {
		if !decode(w, r, &body) {
			return
		}
		projectID = body.ProjectID
	}
	if _, err := a.store.readProjectMeta(r.Context(), projectID); err != nil {
		sqlHTTPError(w, err)
		return
	}
	var group runnerGroup
	var receipt runnerOperation
	if r.Method == http.MethodGet {
		err = a.store.readSnapshot(r.Context(), func(tx *sql.Tx) error {
			var err error
			group, err = (&sqliteSessionTx{ctx: r.Context(), tx: tx}).runnerGroup(owner, projectID, key)
			return err
		})
		if err != nil {
			sqlHTTPError(w, err)
			return
		}
		respond(w, 200, group)
		return
	}
	a.mu.Lock()
	err = a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		var err error
		group, err = st.runnerGroup(owner, projectID, key)
		if err != nil {
			return err
		}
		if r.Method == http.MethodGet {
			return nil
		}
		digest := revision(runnerRawJSON(map[string]string{"group_cancel": key, "project_id": projectID}))
		receipt, err = st.runnerReceipt(owner, body.OperationID, digest, nil)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		group.CancelRequested = true
		seal := runnerGroup{InstanceKey: key, ProjectID: projectID, CancelRequested: true, Executions: []runnerExecution{}}
		if _, err := st.runnerReceipt(owner, runnerGroupRecordID(key), revision(runnerRawJSON(seal)), seal); err != nil {
			return err
		}
		for i, execution := range group.Executions {
			execution.CancelRequested = true
			if execution.State != "cancelled" {
				execution.State = "cancelling"
			}
			if err := st.putRunnerExecution(execution); err != nil {
				return err
			}
			if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_inputs SET state='paused' WHERE execution_id=? AND state='retained'`, execution.ID); err != nil {
				return err
			}
			group.Executions[i] = execution
		}
		receipt, err = st.runnerReceipt(owner, body.OperationID, digest, seal)
		return err
	})
	if err == nil && r.Method == http.MethodPost {
		err = a.runner.resumeCancellationsLocked()
	}
	a.mu.Unlock()
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	if r.Method == http.MethodPost {
		respond(w, 202, receipt)
		return
	}
	respond(w, 200, group)
}

func (s *runnerService) resumeCancellationsLocked() error {
	rows, err := s.app.store.readers.QueryContext(s.app.ctx, `SELECT body FROM runner_executions
		WHERE json_extract(CAST(body AS TEXT),'$.cancel_requested')=1
		AND (coalesce(json_extract(CAST(body AS TEXT),'$.instance_key'),'')<>'' OR json_extract(CAST(body AS TEXT),'$.recovery_policy')='resume_drained')
		AND json_extract(CAST(body AS TEXT),'$.state')<>'cancelled'`)
	if err != nil {
		return err
	}
	items := []runnerExecution{}
	for rows.Next() {
		var body []byte
		var execution runnerExecution
		if err := rows.Scan(&body); err != nil {
			rows.Close()
			return err
		}
		if err := json.Unmarshal(body, &execution); err != nil {
			rows.Close()
			return err
		}
		items = append(items, execution)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, execution := range items {
		if s.cancelling[execution.ID] {
			continue
		}
		if s.cancelling == nil {
			s.cancelling = map[string]bool{}
		}
		s.cancelling[execution.ID] = true
		t := s.app.runs[execution.SessionID]
		if t != nil {
			t.cancel()
		}
		s.app.wg.Add(1)
		go func(execution runnerExecution, t *turn) {
			defer s.app.wg.Done()
			s.stopOwned(execution, t)
			s.app.mu.Lock()
			delete(s.cancelling, execution.ID)
			s.app.mu.Unlock()
		}(execution, t)
	}
	return nil
}

func (s *runnerService) groupOpen(ctx context.Context, execution runnerExecution) error {
	return s.app.store.readSnapshot(ctx, func(tx *sql.Tx) error {
		return (&sqliteSessionTx{ctx: ctx, tx: tx}).runnerGroupOpen(execution.IntegrationID, execution.ProjectID, execution.InstanceKey)
	})
}
