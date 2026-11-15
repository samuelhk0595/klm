package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"time"
)

func (st *sqliteSessionTx) runnerInputState(id, state string) error {
	var body []byte
	err := st.tx.QueryRowContext(st.ctx, `SELECT body FROM runner_inputs WHERE id=?`, id).Scan(&body)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var input runnerInput
	if err := json.Unmarshal(body, &input); err != nil {
		return err
	}
	input.State = state
	body, err = runnerJSON(input)
	if err != nil {
		return err
	}
	_, err = st.tx.ExecContext(st.ctx, `UPDATE runner_inputs SET state=?,body=? WHERE id=?`, state, body, id)
	return err
}

func (a *app) runnerRemoveInput(w http.ResponseWriter, r *http.Request) {
	execution, err := a.runnerOwned(r)
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
	digest := revision([]byte("remove/" + execution.ID + "/" + r.PathValue("inputId")))
	a.mu.Lock()
	defer a.mu.Unlock()
	var receipt runnerOperation
	err = a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		var err error
		receipt, err = st.runnerReceipt(execution.IntegrationID, body.OperationID, digest, nil)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		var data []byte
		var state string
		if err := st.tx.QueryRowContext(st.ctx, `SELECT state,body FROM runner_inputs WHERE execution_id=? AND id=?`, execution.ID, r.PathValue("inputId")).Scan(&state, &data); err != nil {
			return err
		}
		var input runnerInput
		if err := json.Unmarshal(data, &input); err != nil {
			return err
		}
		if state != "retained" || input.Mode != "queue" || input.SourceKind != "human" {
			return controlError(409, "Only an unclaimed queued human input can be removed.")
		}
		if err := st.runnerInputState(input.ID, "removed"); err != nil {
			return err
		}
		receipt, err = st.runnerReceipt(execution.IntegrationID, body.OperationID, digest, map[string]any{"removed": input.ID})
		return err
	})
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 200, receipt)
}

func (a *app) runnerConfiguration(w http.ResponseWriter, r *http.Request) {
	execution, err := a.runnerOwned(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	if execution.Kind != "coordinator" {
		fail(w, 409, "Worker captures are immutable.")
		return
	}
	var body struct {
		OperationID      string        `json:"operation_id"`
		ExpectedRevision string        `json:"expected_revision"`
		Capture          runnerCapture `json:"capture"`
	}
	if !decode(w, r, &body) {
		return
	}
	project, err := a.store.readProjectMeta(r.Context(), execution.ProjectID)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	capture, err := a.validateRunnerCapture(r.Context(), body.Capture, project)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	encoded, _ := json.Marshal(body)
	digest := revision(append([]byte(execution.ID+"/configuration/"), encoded...))
	var receipt runnerOperation
	fresh := false
	a.mu.Lock()
	err = a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		var err error
		receipt, err = st.runnerReceipt(execution.IntegrationID, body.OperationID, digest, nil)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		execution, err = st.runnerExecution(execution.ID)
		if err != nil {
			return err
		}
		if execution.Revision != body.ExpectedRevision {
			return controlError(409, "Configuration revision changed.")
		}
		fresh = true
		execution.Desired = capture
		if err := st.putRunnerExecution(execution); err != nil {
			return err
		}
		receipt, err = st.runnerReceipt(execution.IntegrationID, body.OperationID, digest, map[string]any{"saved": true, "applied": false})
		return err
	})
	if err == nil && fresh && revision(runnerCaptureJSON(execution.Desired)) != revision(runnerCaptureJSON(execution.Applied)) {
		a.runner.applyDesiredLocked(execution)
	}
	a.mu.Unlock()
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 202, receipt)
}

func (a *app) runnerToolReply(w http.ResponseWriter, r *http.Request) {
	owner, err := a.runnerOwner(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var body struct {
		OperationID string          `json:"operation_id"`
		Result      json.RawMessage `json:"result"`
		IsError     bool            `json:"is_error"`
	}
	if !decode(w, r, &body) {
		return
	}
	encoded, _ := json.Marshal(body)
	digest := revision(append([]byte("tool/"+r.PathValue("callId")+"/"), encoded...))
	var receipt runnerOperation
	a.mu.Lock()
	err = a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		var err error
		receipt, err = st.runnerReceipt(owner, body.OperationID, digest, nil)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		var data []byte
		if err := st.tx.QueryRowContext(st.ctx, `SELECT body FROM runner_calls WHERE id=?`, r.PathValue("callId")).Scan(&data); err != nil {
			return err
		}
		var call runnerToolCall
		if err := json.Unmarshal(data, &call); err != nil {
			return err
		}
		execution, err := st.runnerExecution(call.ExecutionID)
		if err != nil {
			return err
		}
		if execution.IntegrationID != owner {
			return sql.ErrNoRows
		}
		if call.State == "replied" && (!bytes.Equal(call.Reply, body.Result) || call.IsError != body.IsError) {
			return controlError(409, "Tool call already has a different reply.")
		}
		call.Reply, call.State, call.IsError = body.Result, "replied", body.IsError
		data, err = runnerJSON(call)
		if err != nil {
			return err
		}
		if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_calls SET body=? WHERE id=?`, data, call.ID); err != nil {
			return err
		}
		if err := st.runnerEvent(execution, "tool_reply", call); err != nil {
			return err
		}
		receipt, err = st.runnerReceipt(owner, body.OperationID, digest, map[string]any{"call_id": call.ID, "state": "recorded"})
		return err
	})
	a.mu.Unlock()
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	a.runner.mu.Lock()
	wake := a.runner.callWake[r.PathValue("callId")]
	if wake != nil {
		select {
		case wake <- struct{}{}:
		default:
		}
	}
	a.runner.mu.Unlock()
	respond(w, 202, receipt)
}

func (a *app) runnerStopHTTP(w http.ResponseWriter, r *http.Request) {
	execution, err := a.runnerOwned(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var body struct {
		OperationID      string `json:"operation_id"`
		ExpectedTurnID   string `json:"expected_turn_id"`
		ExpectedRevision string `json:"expected_revision"`
	}
	if !decode(w, r, &body) {
		return
	}
	encoded, _ := json.Marshal(body)
	digest := revision(append([]byte(execution.ID+"/stop/"), encoded...))
	var receipt runnerOperation
	a.mu.Lock()
	var stopped *turn
	fresh := false
	err = a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		var err error
		receipt, err = st.runnerReceipt(execution.IntegrationID, body.OperationID, digest, nil)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		fresh = true
		execution, err = st.runnerExecution(execution.ID)
		if err != nil {
			return err
		}
		if body.ExpectedTurnID != "" && execution.TurnID != body.ExpectedTurnID {
			return controlError(409, "Observed turn changed.")
		}
		if body.ExpectedRevision != "" && execution.Revision != body.ExpectedRevision {
			return controlError(409, "Execution changed.")
		}
		execution.State = "cancelling"
		if err := st.putRunnerExecution(execution); err != nil {
			return err
		}
		if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_inputs SET state='paused' WHERE execution_id=? AND state='retained'`, execution.ID); err != nil {
			return err
		}
		stopped = a.runs[execution.SessionID]
		receipt, err = st.runnerReceipt(execution.IntegrationID, body.OperationID, digest, map[string]any{"requested": true, "execution_id": execution.ID})
		return err
	})
	if err == nil && fresh {
		if stopped != nil {
			stopped.cancel()
		}
		a.wg.Add(1)
	}
	a.mu.Unlock()
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	if fresh {
		go func() { defer a.wg.Done(); a.runner.stopOwned(execution, stopped) }()
	}
	respond(w, 202, receipt)
}

func (s *runnerService) stopOwned(execution runnerExecution, t *turn) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var failure error
	if t == nil && execution.Error != "" {
		failure = errors.New(execution.Error)
	}
	s.mu.Lock()
	jobs := []*runnerLiveJob{}
	for _, job := range s.jobs {
		if job.record.ExecutionID == execution.ID {
			jobs = append(jobs, job)
		}
	}
	s.mu.Unlock()
	for _, job := range jobs {
		job.mu.Lock()
		job.cancelRequested = true
		job.mu.Unlock()
		if err := job.owner.Stop(); err != nil {
			failure = errors.Join(failure, err)
		}
		select {
		case <-job.done:
			if !job.owner.Drained() {
				failure = errors.Join(failure, errors.New("Owned job process drainage is unconfirmed."))
			}
		case <-ctx.Done():
			failure = errors.Join(failure, ctx.Err())
		}
	}
	if t != nil {
		select {
		case <-t.done:
			failure = errors.Join(failure, t.stopErr)
		case <-ctx.Done():
			failure = errors.Join(failure, ctx.Err())
		}
	}
	for failure == nil {
		pending, err := s.unsettledEffects(ctx, execution.ID, "")
		if err != nil {
			failure = err
			break
		}
		if pending == 0 {
			break
		}
		select {
		case <-time.After(25 * time.Millisecond):
		case <-ctx.Done():
			failure = ctx.Err()
		}
	}
	s.app.mu.Lock()
	defer s.app.mu.Unlock()
	_ = s.app.persistSQLLocked(context.Background(), func(st *sqliteSessionTx) error {
		current, err := st.runnerExecution(execution.ID)
		if err != nil {
			return err
		}
		var unsettledJobs bool
		if err := st.tx.QueryRowContext(st.ctx, `SELECT EXISTS(SELECT 1 FROM runner_jobs WHERE execution_id=? AND state IN ('starting','running','unknown'))`, execution.ID).Scan(&unsettledJobs); err != nil {
			return err
		}
		if unsettledJobs {
			failure = errors.Join(failure, errors.New("Persisted owned job cessation is unconfirmed."))
		}
		current.State = "stopped"
		current.Error = ""
		if current.Kind == "worker" {
			current.State = "cancelled"
		}
		if failure != nil {
			current.State = "interrupted"
			current.Error = failure.Error()
		}
		return st.putRunnerExecution(current)
	})
	if failure == nil && execution.Kind == "coordinator" {
		_ = s.scheduleLocked(execution.ID)
	}
}

func (a *app) runnerHistory(w http.ResponseWriter, r *http.Request) {
	execution, err := a.runnerOwned(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	sessionID, err := a.runnerHistorySession(r, execution)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	page, err := a.store.readHistory(r.Context(), sessionID, r.URL.Query().Get("before"), r.URL.Query().Get("after"), 25)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 200, page)
}

func (a *app) runnerHistorySession(r *http.Request, execution runnerExecution) (string, error) {
	id := r.URL.Query().Get("source_session_id")
	if id == "" {
		id = execution.SessionID
	}
	var found bool
	err := a.store.readers.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM runner_generations WHERE execution_id=? AND session_id=?)`, execution.ID, id).Scan(&found)
	if err == nil && !found {
		err = sql.ErrNoRows
	}
	return id, err
}

func (a *app) runnerContent(w http.ResponseWriter, r *http.Request) {
	execution, err := a.runnerOwned(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	id, err := a.runnerHistorySession(r, execution)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	r.SetPathValue("id", id)
	a.sqlEventContent(w, r)
}

func (a *app) runnerRequestReply(w http.ResponseWriter, r *http.Request) {
	owner, err := a.runnerOwner(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var body struct {
		OperationID string          `json:"operation_id"`
		Answer      json.RawMessage `json:"answer"`
	}
	if !decode(w, r, &body) {
		return
	}
	var execution runnerExecution
	var nativeID, kind, state string
	var receipt runnerOperation
	encoded, _ := json.Marshal(body)
	digest := revision(append([]byte("request/"+r.PathValue("requestId")+"/"), encoded...))
	var prior []byte
	if err := a.store.readers.QueryRowContext(r.Context(), `SELECT body FROM runner_operations WHERE integration_id=? AND id=?`, owner, body.OperationID).Scan(&prior); err == nil {
		var previous runnerOperation
		if err := json.Unmarshal(prior, &previous); err != nil {
			sqlHTTPError(w, err)
			return
		}
		if previous.Digest != digest {
			fail(w, 409, "Reply operation body changed.")
			return
		}
		respond(w, 202, previous)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		sqlHTTPError(w, err)
		return
	}
	protectedReply, protectErr := a.protectRunnerReply(body.Answer)
	if protectErr != nil {
		sqlHTTPError(w, protectErr)
		return
	}
	a.mu.Lock()
	err = a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		var err error
		receipt, err = st.runnerReceipt(owner, body.OperationID, digest, nil)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		var executionID string
		if err := st.tx.QueryRowContext(st.ctx, `SELECT execution_id,native_id,kind,state FROM runner_requests WHERE id=?`, r.PathValue("requestId")).Scan(&executionID, &nativeID, &kind, &state); err != nil {
			return err
		}
		execution, err = st.runnerExecution(executionID)
		if err != nil {
			return err
		}
		if execution.IntegrationID != owner {
			return sql.ErrNoRows
		}
		if state != "pending" {
			return controlError(409, "This exact native request is no longer pending.")
		}
		if err := a.validateRunnerAnswerLocked(execution.SessionID, nativeID, kind, body.Answer); err != nil {
			return err
		}
		if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_requests SET state='reply_recorded',reply=? WHERE id=? AND reply IS NULL`, protectedReply, r.PathValue("requestId")); err != nil {
			return err
		}
		receipt, err = st.runnerReceipt(owner, body.OperationID, digest, map[string]string{"state": "recorded"})
		return err
	})
	a.mu.Unlock()
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	if nativeID == "" {
		respond(w, 202, receipt)
		return
	}
	request := r.Clone(r.Context())
	request.Body = http.NoBody
	request.SetPathValue("id", execution.SessionID)
	request.Body = ioBody(body.Answer)
	recorder := httptest.NewRecorder()
	if kind == "pending_permissions" {
		var decision struct {
			Decision string `json:"decision"`
		}
		_ = json.Unmarshal(body.Answer, &decision)
		if decision.Decision != "once" && decision.Decision != "reject" && decision.Decision != "session" {
			fail(w, 400, "External permissions retain request or native-session scope only.")
			return
		}
		request.SetPathValue("permissionID", nativeID)
		a.permissionDecision(recorder, request)
	} else {
		request.SetPathValue("questionID", nativeID)
		var answer map[string]any
		_ = json.Unmarshal(body.Answer, &answer)
		if _, elicitation := answer["action"]; elicitation {
			a.answerMCPElicitation(recorder, request)
		} else {
			a.answerQuestion(recorder, request)
		}
	}
	state = "delivered"
	if recorder.Code >= 400 {
		state = "delivery_failed"
	}
	a.mu.Lock()
	_ = a.persistSQLLocked(context.Background(), func(st *sqliteSessionTx) error {
		_, err := st.tx.ExecContext(st.ctx, `UPDATE runner_requests SET state=? WHERE id=?`, state, r.PathValue("requestId"))
		if err != nil {
			return err
		}
		return st.runnerEvent(execution, "request_reply", map[string]string{"id": r.PathValue("requestId"), "state": state})
	})
	a.mu.Unlock()
	respond(w, 202, receipt)
}

type runnerBody struct{ *bytes.Reader }

func (runnerBody) Close() error     { return nil }
func ioBody(data []byte) runnerBody { return runnerBody{Reader: bytes.NewReader(data)} }
