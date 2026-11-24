package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"time"
)

type runnerTurn struct {
	uncertainMCP   map[string]bool
	mu             sync.Mutex
	activeTools    map[string]string
	adapter        *adapter // app.mu; published only after selected MCP preparation.
	ExecutionID    string
	TurnID         string
	Capture        runnerCapture
	RecoveryPolicy string
}

type runnerService struct {
	app                *app
	changed            chan struct{}
	mu                 sync.Mutex
	jobs               map[string]*runnerLiveJob
	callWake           map[string]chan struct{}
	prepared           map[string]string // app.mu; consumed at an actual complete-turn claim.
	preparedContext    map[string]string
	cancelling         map[string]bool // app.mu; one existing-owner drainage attempt per execution.
	maintenanceStarted bool
}

func (a *app) initializeRunner(ctx context.Context) error {
	if err := a.store.write(ctx, func(tx *sql.Tx) error { _, err := tx.ExecContext(ctx, runnerSchema); return err }); err != nil {
		return err
	}
	a.runner = &runnerService{app: a, changed: make(chan struct{}, 1), jobs: map[string]*runnerLiveJob{}, callWake: map[string]chan struct{}{}, prepared: map[string]string{}, preparedContext: map[string]string{}}
	context.AfterFunc(a.ctx, func() {
		a.runner.mu.Lock()
		jobs := []*runnerLiveJob{}
		for _, job := range a.runner.jobs {
			jobs = append(jobs, job)
		}
		a.runner.mu.Unlock()
		for _, job := range jobs {
			job.mu.Lock()
			job.cancelRequested = true
			job.mu.Unlock()
			_ = job.owner.Stop()
		}
	})
	if err := a.store.write(ctx, func(tx *sql.Tx) error {
		// Lost callbacks/process handles are not reconstructed by replaying input.
		_, err := tx.ExecContext(ctx, `UPDATE runner_inputs SET state='unknown' WHERE state IN ('claimed','accepted'); UPDATE runner_jobs SET state='unknown' WHERE state IN ('starting','running'); UPDATE runner_requests SET state='unavailable' WHERE state='pending';`)
		return err
	}); err != nil {
		return err
	}
	return a.recoverRunnerExecutions(ctx)
}

func (s *runnerService) wake() {
	select {
	case s.changed <- struct{}{}:
	default:
	}
}

func (s *runnerService) scheduleLocked(executionID string) error {
	a := s.app
	execution, err := a.store.readRunnerExecution(a.ctx, executionID)
	if err != nil {
		return err
	}
	if a.runs[execution.SessionID] != nil {
		return nil
	}
	if execution.CancelRequested {
		return nil
	}
	if execution.State != "idle" && execution.State != "stopped" && execution.State != "waiting_jobs" && execution.State != "blocked" {
		return nil
	}
	var actionable bool
	if err := a.store.readers.QueryRowContext(a.ctx, `SELECT EXISTS(SELECT 1 FROM runner_inputs WHERE execution_id=? AND state='retained' AND json_extract(body,'$.mode')<>'passive')`, executionID).Scan(&actionable); err != nil {
		return err
	}
	if !actionable {
		return nil
	}
	project, err := a.store.readProjectMeta(a.ctx, execution.ProjectID)
	if err != nil {
		return err
	}
	_, initialCWD, err := a.readSQLConversation(a.ctx, execution.SessionID)
	if err != nil {
		return err
	}
	if err := validateRunnerLaunchDirectory(project, initialCWD); err != nil {
		return a.writeSQLLocked(a.ctx, func(st *sqliteSessionTx) error {
			execution.State, execution.Error = "interrupted", err.Error()
			execution.RecoveryBlocked = "workspace_unavailable"
			return st.putRunnerExecution(execution)
		})
	}
	if execution.Kind == "coordinator" && execution.Desired.ConfigurationGate && s.prepared[executionID] != revision(runnerCaptureJSON(execution.Desired)) {
		return s.captureNextTurnLocked(execution)
	}
	if revision(runnerCaptureJSON(execution.Desired)) != revision(runnerCaptureJSON(execution.Applied)) {
		s.applyDesiredLocked(execution)
		return nil
	}
	var meta Session
	var native nativeSession
	var payload submission
	var cwd string
	turnID := newID()
	err = a.writeSQLLocked(a.ctx, func(st *sqliteSessionTx) error {
		if err := st.runnerGroupOpen(execution.IntegrationID, execution.ProjectID, execution.InstanceKey); err != nil {
			return err
		}
		inputs := []runnerInput{}
		rows, err := st.tx.QueryContext(st.ctx, `SELECT body FROM runner_inputs WHERE execution_id=? AND state='retained' ORDER BY position LIMIT 32`, executionID)
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
			inputs = append(inputs, input)
		}
		if err := rows.Close(); err != nil {
			return err
		}
		actionable := false
		for _, input := range inputs {
			actionable = actionable || input.Mode != "passive"
		}
		if !actionable {
			return nil
		}
		execution.TurnID, execution.State = turnID, "running"
		execution.ProcessesDrained = false
		execution.ProcessDomains = []runnerProcessDomain{}
		execution.RecoveryBlocked = ""
		execution.NativeWaiting = false
		meta, err = st.sessionMeta(execution.SessionID)
		if err != nil {
			return err
		}
		cwd, err = st.conversationDirectory(meta)
		if err != nil {
			return err
		}
		native, err = st.native(meta.ID)
		if err != nil {
			return err
		}
		payload.Text = execution.Applied.Instructions + "\n\n"
		if native.ID == "" && len(execution.RecentConversation) > 0 {
			recent, err := json.Marshal(execution.RecentConversation)
			if err != nil {
				return err
			}
			payload.Text += "Retained conversation reference (not new input):\n" + string(recent) + "\n\n"
		}
		if context := s.preparedContext[executionID]; context != "" {
			payload.Text += "Application turn context:\n" + context + "\n\n"
		}
		for _, input := range inputs {
			input.State, input.TurnID = "claimed", turnID
			body, err := runnerJSON(input)
			if err != nil {
				return err
			}
			if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_inputs SET state='claimed',body=? WHERE id=? AND state='retained'`, body, input.ID); err != nil {
				return err
			}
			text := input.Text
			if len(input.Mentions) > 0 {
				var prepared queuedPayload
				if err := decodeSQLiteRecord(st.ctx, st.tx, "queue_payloads", input.ID, 2<<20, &prepared); err != nil {
					return err
				}
				text = prepared.Submission.piText()
			}
			var queued string
			if err := st.tx.QueryRowContext(st.ctx, `SELECT id FROM data_queued_messages WHERE record_id=? LIMIT 1`, input.ID).Scan(&queued); err == nil {
				if err := st.deleteRecord("queued_messages", queued); err != nil {
					return err
				}
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			payload.Text += "\nInput " + input.ID + " (" + input.SourceKind + ", source " + input.SourceID + "):\n" + text + "\n"
			e := event("runner_input", input.Text)
			e.ID = input.ID
			e.Data = map[string]any{"sourceKind": input.SourceKind, "sourceId": input.SourceID, "mode": input.Mode}
			if len(input.Mentions) > 0 {
				e.Data["mentions"] = input.Mentions
			}
			if _, err := st.appendEvent(meta.ID, e); err != nil {
				return err
			}
		}
		meta.Status, meta.UpdatedAt = "running", now()
		if err := st.putSessionMeta(meta); err != nil {
			return err
		}
		return st.putRunnerExecution(execution)
	})
	if err != nil || payload.Text == "" {
		return err
	}
	b, ok := a.binaries[meta.Harness]
	if !ok {
		return controlError(409, "Captured harness is unavailable.")
	}
	ctx, cancel := context.WithCancel(a.ctx)
	delete(s.prepared, executionID)
	delete(s.preparedContext, executionID)
	t := &turn{
		ctx: ctx, cancel: cancel, done: make(chan struct{}),
		runner: &runnerTurn{
			ExecutionID: execution.ID, TurnID: turnID, Capture: execution.Applied, RecoveryPolicy: execution.RecoveryPolicy,
		},
	}
	a.runs[meta.ID] = t
	a.wg.Add(1)
	go a.execute(t, meta, native, b, cwd, payload)
	return nil
}

func (s *runnerService) finishedLocked(p *adapter, failure error) {
	a := s.app
	err := a.persistSQLLocked(context.Background(), func(st *sqliteSessionTx) error {
		execution, err := st.runnerExecution(p.turn.runner.ExecutionID)
		if err != nil {
			return err
		}
		stopping := execution.State == "cancelling"
		execution.ProcessesDrained = p.processesDrained && !p.processDrainFailed && p.turn.stopErr == nil
		if !execution.ProcessesDrained {
			execution.RecoveryBlocked = "native_process_drainage_unconfirmed"
		}
		p.turn.runner.mu.Lock()
		if len(p.turn.runner.uncertainMCP) > 0 {
			execution.RecoveryBlocked = "mcp_result_unknown"
		}
		p.turn.runner.mu.Unlock()
		execution.State = "idle"
		var actionable bool
		if err := st.tx.QueryRowContext(st.ctx, `SELECT EXISTS(SELECT 1 FROM runner_inputs WHERE execution_id=? AND state='retained' AND json_extract(body,'$.mode')<>'passive')`, execution.ID).Scan(&actionable); err != nil {
			return err
		}
		if execution.Kind == "worker" && !actionable {
			execution.State = "native_completed"
		}
		var jobs int
		if err := st.tx.QueryRowContext(st.ctx, `SELECT count(*) FROM runner_jobs WHERE execution_id=? AND state IN ('starting','running','unknown')`, execution.ID).Scan(&jobs); err != nil {
			return err
		}
		if jobs > 0 {
			execution.State = "waiting_jobs"
		}
		var unanswered int
		if err := st.tx.QueryRowContext(st.ctx, `SELECT count(*) FROM runner_requests WHERE execution_id=? AND state='pending'`, execution.ID).Scan(&unanswered); err != nil {
			return err
		}
		if unanswered > 0 {
			execution.State = "interrupted"
			execution.Error = "Native context ended with unanswered requests; original callbacks are unavailable."
		}
		var effects int
		if err := st.tx.QueryRowContext(st.ctx, `SELECT count(*) FROM runner_calls WHERE execution_id=? AND json_extract(body,'$.state')='admitted'`, execution.ID).Scan(&effects); err != nil {
			return err
		}
		if effects > 0 {
			execution.State = "interrupted"
			execution.Error = "Application effects have not confirmed settlement."
		}
		if failure != nil || p.failed || !p.completed {
			execution.State = "interrupted"
			if failure != nil {
				execution.Error = failure.Error()
			} else if p.runnerFailureText != "" {
				execution.Error = p.runnerFailureText
			} else if execution.Error == "" {
				execution.Error = "Native turn ended without confirmed completion; inspect the retained native error events."
			}
			if execution.RecoveryPolicy == "resume_drained" && execution.RecoveryBlocked == "" && p.turn.ctx.Err() == nil {
				execution.RecoveryBlocked = "native_resume_failed"
				var network net.Error
				if errors.As(failure, &network) && network.Timeout() {
					execution.RecoveryBlocked = "native_transport_retry"
					delay := time.Duration(1<<min(execution.RecoveryAttempts+1, 6)) * time.Second
					execution.RecoveryRetryAfter = time.Now().Add(delay).UTC().Format(time.RFC3339Nano)
				}
			}
		}
		if p.turn.ctx.Err() != nil {
			if stopping {
				execution.State = "cancelling"
			} else {
				execution.State = "interrupted"
			}
		}
		if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_inputs SET state='unknown' WHERE execution_id=? AND state IN ('claimed','accepted')`, execution.ID); err != nil {
			return err
		}
		if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_requests SET state='unavailable' WHERE execution_id=? AND state='pending'`, execution.ID); err != nil {
			return err
		}
		return st.putRunnerExecution(execution)
	})
	if err == nil && p.turn.ctx.Err() == nil && failure == nil {
		_ = s.scheduleLocked(p.turn.runner.ExecutionID)
	}
}

func runnerCaptureJSON(c runnerCapture) []byte { body, _ := json.Marshal(c); return body }

func (s *runnerService) applyDesiredLocked(execution runnerExecution) {
	a := s.app
	if a.closing || a.runs[execution.SessionID] != nil {
		return
	}
	if pending, err := s.unsettledEffects(a.ctx, execution.ID, ""); err != nil || pending > 0 {
		return
	}
	var jobs int
	if err := a.store.readers.QueryRow(`SELECT count(*) FROM runner_jobs WHERE execution_id=? AND state IN ('starting','running','unknown')`, execution.ID).Scan(&jobs); err != nil || jobs > 0 {
		return
	}
	if err := a.writeSQLLocked(a.ctx, func(st *sqliteSessionTx) error {
		execution.State = "configuring"
		return st.putRunnerExecution(execution)
	}); err != nil {
		return
	}
	oldRuntime := a.runtimes[execution.SessionID]
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		var failure error
		if oldRuntime != nil {
			failure = oldRuntime.shutdown()
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		_ = a.persistSQLLocked(context.Background(), func(st *sqliteSessionTx) error {
			current, err := st.runnerExecution(execution.ID)
			if err != nil {
				return err
			}
			if current.State != "configuring" {
				return nil
			}
			if failure != nil {
				current.State = "interrupted"
				current.Error = failure.Error()
				return st.putRunnerExecution(current)
			}
			previous, err := st.sessionMeta(current.SessionID)
			if err != nil {
				return err
			}
			current.RecentConversation, err = st.runnerRecentConversation(current)
			if err != nil {
				return err
			}
			current.SessionID, current.Generation, current.Applied, current.State = newID(), runnerNext(current.Generation), execution.Desired, "idle"
			previous.ID, previous.Harness, previous.Model, previous.Effort = current.SessionID, current.Applied.Harness, current.Applied.Model, current.Applied.Effort
			previous.Status, previous.CreatedAt, previous.UpdatedAt = "idle", now(), now()
			previous.Capabilities = ConversationCapabilities{Desired: current.Applied.Selection}
			if err := st.createSession(previous, a.dir); err != nil {
				return err
			}
			return st.putRunnerExecution(current)
		})
		_ = s.scheduleLocked(execution.ID)
	}()
}

func (s *runnerService) call(ctx context.Context, p *adapter, name, callID string, raw json.RawMessage) (any, error) {
	for _, tool := range runnerJobTools() {
		if str(tool, "name") == name {
			return s.jobCall(ctx, p, name, callID, raw)
		}
	}
	if callID == "" {
		return nil, errors.New("Native tool call identity is required.")
	}
	key := revision([]byte(p.turn.runner.ExecutionID + "/" + p.turn.runner.TurnID + "/" + callID))
	digest := revision(append([]byte(name+"\n"), raw...))
	a := s.app
	var call runnerToolCall
	a.mu.Lock()
	err := a.writeSQLLocked(ctx, func(st *sqliteSessionTx) error {
		if a.runs[p.id] != p.turn || p.turn.ctx.Err() != nil {
			return controlError(409, "External native call belongs to an inactive turn.")
		}
		var body []byte
		var priorDigest string
		err := st.tx.QueryRowContext(ctx, `SELECT digest,body FROM runner_calls WHERE id=?`, key).Scan(&priorDigest, &body)
		if err == nil {
			if priorDigest != digest {
				return controlError(409, "Native call changed arguments.")
			}
			return json.Unmarshal(body, &call)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		call = runnerToolCall{ID: key, ExecutionID: p.turn.runner.ExecutionID, TurnID: p.turn.runner.TurnID, NativeCallID: callID, Name: name, Arguments: raw, State: "pending"}
		body, err = runnerJSON(call)
		if err != nil {
			return err
		}
		if _, err := st.tx.ExecContext(ctx, `INSERT INTO runner_calls VALUES(?,?,?,?,?,?)`, key, call.ExecutionID, call.TurnID, callID, digest, body); err != nil {
			return err
		}
		execution, err := st.runnerExecution(call.ExecutionID)
		if err != nil {
			return err
		}
		return st.runnerEvent(execution, "tool_call", call)
	})
	a.mu.Unlock()
	if err != nil {
		return nil, err
	}
	for {
		if call.State == "replied" {
			var value any
			err := json.Unmarshal(call.Reply, &value)
			if err == nil && call.IsError {
				err = errors.New("Application tool rejected the request: " + humanError(value))
			}
			return value, err
		}
		s.mu.Lock()
		wake := s.callWake[key]
		if wake == nil {
			wake = make(chan struct{}, 1)
			s.callWake[key] = wake
		}
		s.mu.Unlock()
		var body []byte
		if err := a.store.readers.QueryRowContext(ctx, `SELECT body FROM runner_calls WHERE id=?`, key).Scan(&body); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &call); err != nil {
			return nil, err
		}
		if call.State == "replied" {
			continue
		}
		select {
		case <-wake:
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(30 * time.Second):
			return nil, errors.New("Application tool reply timed out; inspect original call before recovery.")
		}
	}
}
