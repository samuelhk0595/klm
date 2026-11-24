package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

func (a *app) recoverRunnerExecutions(ctx context.Context) error {
	return a.reconcileRunnerExecutions(ctx, false)
}

func (a *app) reconcileRunnerExecutions(ctx context.Context, onlyOpted bool) error {
	rows, err := a.store.readers.QueryContext(ctx, `SELECT id,body FROM runner_executions
		WHERE json_extract(CAST(body AS TEXT),'$.state') IN ('running','cancelling','configuring','capture_pending','waiting_jobs')
		OR (json_extract(CAST(body AS TEXT),'$.state')='interrupted'
		AND json_extract(CAST(body AS TEXT),'$.recovery_policy')='resume_drained')`)
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
		if onlyOpted && execution.RecoveryPolicy != "resume_drained" {
			continue
		}
		a.mu.Lock()
		active := a.runs[execution.SessionID] != nil
		a.mu.Unlock()
		if active {
			continue
		}
		if until, err := time.Parse(time.RFC3339Nano, execution.RecoveryRetryAfter); err == nil && time.Now().Before(until) {
			continue
		}
		// An admission gate has not claimed a native input. All other interrupted
		// owners retain uncertainty; no lost native callback is recreated.
		state := "interrupted"
		message := "Engine stopped before this execution confirmed settlement; claimed inputs and effects were not replayed."
		if execution.State == "capture_pending" {
			state = "idle"
			message = ""
		}
		cancelled := execution.CancelRequested || execution.State == "cancelling"
		if cancelled && execution.Kind == "worker" {
			execution.CancelRequested = true
		}
		if execution.RecoveryPolicy == "resume_drained" {
			if execution.RecoveryBlocked != "mcp_result_unknown" {
				execution.RecoveryBlocked = ""
			}
			project, err := a.store.readProjectMeta(ctx, execution.ProjectID)
			if err != nil {
				return err
			}
			_, cwd, err := a.readSQLConversation(ctx, execution.SessionID)
			if err != nil {
				return err
			}
			if err := validateRunnerLaunchDirectory(project, cwd); err != nil {
				execution.RecoveryBlocked = "workspace_unavailable"
				message = err.Error()
			}
			if err := a.reconcileRunnerOwnership(ctx, &execution); err != nil {
				return err
			}
			if err := a.reconcileRunnerJobs(ctx, &execution); err != nil {
				return err
			}
		}
		if _, err := a.store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
			current, err := st.runnerExecution(execution.ID)
			if err != nil {
				return err
			}
			if current.TurnID != execution.TurnID {
				return nil
			}
			if current.CancelRequested {
				cancelled = true
				execution.CancelRequested = true
			}
			if execution.InstanceKey != "" {
				group, err := st.runnerGroup(execution.IntegrationID, execution.ProjectID, execution.InstanceKey)
				if err != nil {
					return err
				}
				if group.CancelRequested {
					cancelled = true
					execution.CancelRequested = true
				}
			}
			if execution.RecoveryPolicy == "resume_drained" && !cancelled {
				if err := a.runner.recoverDrainedTx(st, &execution); err != nil {
					return err
				}
				if execution.State == "idle" {
					return st.putRunnerExecution(execution)
				}
				message = "Automatic continuation blocked: " + execution.RecoveryBlocked + ". Original inputs and effects were not replayed."
			}
			if cancelled && (execution.InstanceKey != "" || execution.RecoveryPolicy == "resume_drained") {
				state = "cancelling"
			}
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

// This intentionally admits only a durably drained original native context.
// Abrupt owner loss, unknown jobs/effects and lost human callbacks stay explicit.
func (s *runnerService) recoverDrainedTx(st *sqliteSessionTx, execution *runnerExecution) error {
	if execution.RecoveryBlocked != "" {
		return nil
	}
	if execution.ProjectGit == nil {
		execution.RecoveryBlocked = "git_identity_capture_missing"
		return nil
	}
	if !execution.ProcessesDrained {
		execution.RecoveryBlocked = "native_process_drainage_unconfirmed"
		return nil
	}
	checks := []struct {
		query  string
		reason string
	}{
		{query: `SELECT count(*) FROM runner_jobs WHERE execution_id=? AND state IN ('starting','running','unknown')`, reason: "owned_job_outcome_unknown"},
		{query: `SELECT count(*) FROM runner_calls WHERE execution_id=? AND json_extract(body,'$.state')='admitted'`, reason: "application_call_unsettled"},
		{query: `SELECT count(*) FROM runner_calls WHERE execution_id=? AND json_extract(body,'$.state') IN ('mcp_pending','mcp_unknown')
			AND json_extract(body,'$.arguments.operation')='call'`, reason: "mcp_result_unknown"},
	}
	if err := st.runnerGroupOpen(execution.IntegrationID, execution.ProjectID, execution.InstanceKey); err != nil {
		return err
	}
	for _, check := range checks {
		var count int
		if err := st.tx.QueryRowContext(st.ctx, check.query, execution.ID).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			execution.RecoveryBlocked = check.reason
			return nil
		}
	}
	native, err := st.native(execution.SessionID)
	if err != nil {
		return err
	}
	if native.ID == "" {
		execution.RecoveryBlocked = "original_native_context_missing"
		return nil
	}
	// An unadmitted application call had no effect grant. Its lost callback can
	// be retained as evidence while the original native context continues.
	if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_calls SET body=CAST(json_set(CAST(body AS TEXT),'$.state','callback_lost') AS BLOB)
		WHERE execution_id=? AND json_extract(body,'$.state')='pending'`, execution.ID); err != nil {
		return err
	}
	inputs := []json.RawMessage{}
	rows, err := st.tx.QueryContext(st.ctx,
		`SELECT body FROM runner_inputs WHERE execution_id=? AND state='unknown' ORDER BY position LIMIT 32`, execution.ID,
	)
	if err != nil {
		return err
	}
	for rows.Next() {
		var body []byte
		if err := rows.Scan(&body); err != nil {
			rows.Close()
			return err
		}
		var input runnerInput
		if err := json.Unmarshal(body, &input); err != nil {
			rows.Close()
			return err
		}
		input.State = "unknown"
		inputs = append(inputs, runnerRawJSON(input))
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	calls := []runnerToolCall{}
	callRows, err := st.tx.QueryContext(st.ctx, `SELECT body FROM runner_calls WHERE execution_id=? AND turn_id=? ORDER BY id LIMIT 64`, execution.ID, execution.TurnID)
	if err != nil {
		return err
	}
	for callRows.Next() {
		var body []byte
		var call runnerToolCall
		if err := callRows.Scan(&body); err != nil {
			callRows.Close()
			return err
		}
		if err := json.Unmarshal(body, &call); err != nil {
			callRows.Close()
			return err
		}
		calls = append(calls, call)
	}
	if err := callRows.Close(); err != nil {
		return err
	}
	jobs := []json.RawMessage{}
	jobRows, err := st.tx.QueryContext(st.ctx, `SELECT body FROM runner_jobs WHERE execution_id=? AND turn_id=? ORDER BY id LIMIT 32`, execution.ID, execution.TurnID)
	if err != nil {
		return err
	}
	for jobRows.Next() {
		var body []byte
		if err := jobRows.Scan(&body); err != nil {
			jobRows.Close()
			return err
		}
		jobs = append(jobs, json.RawMessage(body))
	}
	if err := jobRows.Close(); err != nil {
		return err
	}
	requests := []json.RawMessage{}
	requestRows, err := st.tx.QueryContext(st.ctx, `SELECT body,reply IS NOT NULL FROM runner_requests WHERE execution_id=? AND state IN ('pending','unavailable','reply_recorded','delivery_failed') ORDER BY id LIMIT 32`, execution.ID)
	if err != nil {
		return err
	}
	for requestRows.Next() {
		var body []byte
		var answered bool
		if err := requestRows.Scan(&body, &answered); err != nil {
			requestRows.Close()
			return err
		}
		requests = append(requests, runnerRawJSON(map[string]any{"original_request": json.RawMessage(body), "answer_retained": answered, "callback_lost": true}))
	}
	if err := requestRows.Close(); err != nil {
		return err
	}
	if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_requests SET state='callback_lost' WHERE execution_id=? AND state IN ('pending','unavailable','reply_recorded','delivery_failed')`, execution.ID); err != nil {
		return err
	}
	// This is reference data, not a replay of an unknown input or a claimed
	// native callback. The immutable application role supplies recovery guidance.
	context := runnerRawJSON(map[string]any{
		"kind": "runner_recovery", "previous_turn_id": execution.TurnID,
		"original_session_id": execution.SessionID, "original_native_id": native.ID,
		"workspace": execution.Workspace, "processes_drained": true,
		"prior_inputs_with_unknown_inclusion": inputs,
		"retained_calls":                      calls, "retained_jobs": jobs, "native_requests_requiring_reestablishment": requests,
		"original_callback_consumption_confirmed": false,
	})
	if len(context) > 120<<10 {
		execution.RecoveryBlocked = "recovery_context_exceeds_input_limit"
		return nil
	}
	input := runnerInput{
		SourceKind: "application_event", SourceID: "recovery/" + execution.TurnID,
		Mode: "steer", Text: string(context),
	}
	if err := s.enqueueTx(st, *execution, input); err != nil {
		return err
	}
	execution.State, execution.Error, execution.NativeWaiting = "idle", "", false
	execution.RecoveryAttempts++
	return nil
}

func (s *runnerService) resumeUnclaimedLocked() error {
	if err := s.resumeCancellationsLocked(); err != nil {
		return err
	}
	if !s.maintenanceStarted {
		s.maintenanceStarted = true
		s.app.wg.Add(1)
		go func() {
			defer s.app.wg.Done()
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-s.app.ctx.Done():
					return
				case <-ticker.C:
					s.app.mu.Lock()
					if !s.app.closing {
						_ = s.resumeCancellationsLocked()
					}
					s.app.mu.Unlock()
					var pending bool
					_ = s.app.store.readers.QueryRowContext(s.app.ctx, `SELECT EXISTS(SELECT 1 FROM runner_executions
					WHERE json_extract(body,'$.state')='interrupted' AND json_extract(body,'$.recovery_policy')='resume_drained'
					AND json_extract(body,'$.recovery_blocked') IN ('application_call_unsettled','workspace_unavailable','owned_job_domain_unconfirmed','native_process_domain_unconfirmed','native_transport_retry'))`).Scan(&pending)
					if pending {
						_ = s.app.reconcileRunnerExecutions(s.app.ctx, true)
					}
					s.app.mu.Lock()
					if !s.app.closing {
						_ = s.resumeIdleLocked()
					}
					s.app.mu.Unlock()
				}
			}
		}()
	}
	return s.resumeIdleLocked()
}

func (s *runnerService) resumeIdleLocked() error {
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
