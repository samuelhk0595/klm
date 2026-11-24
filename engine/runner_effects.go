package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

// Admission is separate from the eventual application result. Stop seals new
// effects and waits for already-admitted effects before claiming drainage.
func (a *app) runnerAdmitEffect(w http.ResponseWriter, r *http.Request) {
	owner, err := a.runnerOwner(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var body struct {
		OperationID string `json:"operation_id"`
	}
	if !decode(w, r, &body) {
		return
	}
	var receipt runnerOperation
	a.mu.Lock()
	defer a.mu.Unlock()
	err = a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		digest := revision([]byte("effect/" + r.PathValue("callId")))
		var err error
		receipt, err = st.runnerReceipt(owner, body.OperationID, digest, nil)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		var encoded []byte
		if err := st.tx.QueryRowContext(st.ctx, `SELECT body FROM runner_calls WHERE id=?`, r.PathValue("callId")).Scan(&encoded); err != nil {
			return err
		}
		var call runnerToolCall
		if err := json.Unmarshal(encoded, &call); err != nil {
			return err
		}
		execution, err := st.runnerExecution(call.ExecutionID)
		if err != nil {
			return err
		}
		if execution.IntegrationID != owner {
			return sql.ErrNoRows
		}
		if err := st.runnerGroupOpen(execution.IntegrationID, execution.ProjectID, execution.InstanceKey); err != nil {
			return err
		}
		t := a.runs[execution.SessionID]
		if call.State != "pending" || execution.State != "running" || execution.TurnID != call.TurnID || t == nil || t.ctx.Err() != nil {
			return controlError(409, "Application effect belongs to a sealed or inactive turn.")
		}
		call.State = "admitted"
		encoded, err = runnerJSON(call)
		if err != nil {
			return err
		}
		if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_calls SET body=? WHERE id=?`, encoded, call.ID); err != nil {
			return err
		}
		if err := st.runnerEvent(execution, "effect_admitted", map[string]string{"call_id": call.ID}); err != nil {
			return err
		}
		receipt, err = st.runnerReceipt(owner, body.OperationID, digest, map[string]any{"admitted": true, "call_id": call.ID})
		return err
	})
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 202, receipt)
}

func (s *runnerService) unsettledEffects(ctx context.Context, executionID, turnID string) (int, error) {
	query := `SELECT count(*) FROM runner_calls WHERE execution_id=? AND (json_extract(body,'$.state')='admitted'
		OR (json_extract(body,'$.state') IN ('mcp_pending','mcp_unknown') AND json_extract(body,'$.arguments.operation')='call'))`
	args := []any{executionID}
	if turnID != "" {
		query += ` AND turn_id=?`
		args = append(args, turnID)
	}
	var count int
	err := s.app.store.readers.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}
