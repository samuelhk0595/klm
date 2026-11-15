package main

import (
	"context"
	"database/sql"
	"encoding/json"
)

func (a *app) recoverRunnerExecutions(ctx context.Context) error {
	rows, err := a.store.readers.QueryContext(ctx, `SELECT id,body FROM runner_executions WHERE json_extract(CAST(body AS TEXT),'$.state') IN ('running','cancelling','configuring','capture_pending','waiting_jobs')`)
	if err != nil {
		return err
	}
	type lost struct {
		id   string
		body []byte
	}
	items := []lost{}
	for rows.Next() {
		var item lost
		if err := rows.Scan(&item.id, &item.body); err != nil {
			rows.Close()
			return err
		}
		items = append(items, item)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, item := range items {
		var execution runnerExecution
		if err := json.Unmarshal(item.body, &execution); err != nil {
			return err
		}
		// An admission gate has not claimed a native input. All other interrupted
		// owners retain uncertainty; no lost native callback is recreated.
		state := "interrupted"
		message := "Engine stopped before this execution confirmed settlement; claimed inputs and effects were not replayed."
		if execution.State == "capture_pending" {
			state = "idle"
			message = ""
		}
		if _, err := a.store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
			execution.State, execution.Error, execution.NativeWaiting = state, message, false
			return st.putRunnerExecution(execution)
		}); err != nil {
			return err
		}
	}
	return a.store.write(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `UPDATE runner_calls SET body=CAST(json_set(CAST(body AS TEXT),'$.state','cancelled') AS BLOB) WHERE json_extract(CAST(body AS TEXT),'$.name')='__capture_next_turn' AND json_extract(CAST(body AS TEXT),'$.state')='pending'`)
		return err
	})
}

func (s *runnerService) resumeUnclaimedLocked() error {
	rows, err := s.app.store.readers.QueryContext(s.app.ctx, `SELECT id FROM runner_executions WHERE json_extract(CAST(body AS TEXT),'$.state')='idle'`)
	if err != nil {
		return err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.scheduleLocked(id); err != nil {
			return err
		}
	}
	return nil
}
