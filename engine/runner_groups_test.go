package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunnerGroupFenceSurvivesDelayedAdmission(t *testing.T) {
	a, execution := runnerRegressionApp(t, "worker")
	execution.InstanceKey = "owned-instance"
	execution.TurnID = ""
	if _, err := a.store.writer.Exec(`INSERT INTO runner_integrations VALUES(?,?,?)`, execution.IntegrationID, "", "group-fence-client-key"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
		if err := st.putRunnerExecution(execution); err != nil {
			return err
		}
		_, err := st.runnerReceipt(execution.IntegrationID, "prepared-member", "capture-digest", runnerPreparation{
			ProjectID: execution.ProjectID, InstanceKey: execution.InstanceKey, Workspace: &runnerWorkspace{Path: "unused-sealed-workspace"},
		})
		return err
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	a.runnerRoutes(mux)
	request := func(method, path string, body any) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "http://localhost"+path, bytes.NewReader(runnerRawJSON(body)))
		r.Header.Set("X-KLM-Runner-ID", execution.IntegrationID)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	cancel := request("POST", "/api/runner/groups/owned-instance/cancel", map[string]string{"operation_id": "cancel-instance", "project_id": execution.ProjectID})
	if cancel.Code != 202 {
		t.Fatal(cancel.Code, cancel.Body.String())
	}
	a.wg.Wait()
	// Reinitialize the durable runner exactly as a fresh engine does. Old source
	// receipts remain readable, but a delayed new admission cannot claim them.
	if err := a.initializeRunner(context.Background()); err != nil {
		t.Fatal(err)
	}
	late := request("POST", "/api/runner/work", map[string]string{
		"operation_id": "late-worker", "key": "late-worker", "project_id": execution.ProjectID, "preparation_id": "prepared-member", "prompt": "must never run",
	})
	if late.Code != 409 || !strings.Contains(late.Body.String(), "durably cancelled") {
		t.Fatal(late.Code, late.Body.String())
	}
	prepared := request("POST", "/api/runner/captures", map[string]string{"operation_id": "late-capture", "project_id": execution.ProjectID, "instance_key": execution.InstanceKey})
	if prepared.Code != 409 || !strings.Contains(prepared.Body.String(), "durably cancelled") {
		t.Fatal(prepared.Code, prepared.Body.String())
	}
	input := request("POST", "/api/runner/work/"+execution.ID+"/inputs", runnerInput{OperationID: "late-input", SourceKind: "application_event", SourceID: "late-input", Mode: "steer", Text: "must never run"})
	if input.Code != 409 {
		t.Fatal(input.Code, input.Body.String())
	}
	read := request("GET", "/api/runner/groups/owned-instance?project_id="+execution.ProjectID, nil)
	var group runnerGroup
	if read.Code != 200 || json.Unmarshal(read.Body.Bytes(), &group) != nil || !group.CancelRequested || len(group.Executions) != 1 || group.Executions[0].State != "cancelled" {
		t.Fatal(read.Code, read.Body.String())
	}
	if a.runs[execution.SessionID] != nil {
		t.Fatal("cancelled instance acquired a native turn")
	}
	var admitted int
	if err := a.store.readers.QueryRow(`SELECT count(*) FROM runner_executions`).Scan(&admitted); err != nil || admitted != 1 {
		t.Fatal("delayed admission created a replacement", admitted, err)
	}
}

func TestRunnerColdReplyReconciliation(t *testing.T) {
	a, execution := runnerRegressionApp(t, "worker")
	if err := a.bindRunnerProjectGit(a.ctx, &execution); err != nil {
		t.Fatal(err)
	}
	execution.State = "interrupted"
	execution.RecoveryPolicy = "resume_drained"
	execution.ProcessesDrained = true
	execution.InstanceKey = "reply-instance"
	call := runnerToolCall{ID: "retained-call", ExecutionID: execution.ID, TurnID: execution.TurnID, NativeCallID: "native-call", Name: "retain_result", State: "admitted", Arguments: json.RawMessage(`{"file":"retained.txt"}`)}
	if _, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
		if err := st.putRunnerExecution(execution); err != nil {
			return err
		}
		if err := st.putRoot("native_sessions", execution.SessionID, nativeSession{ID: "original-native-thread"}); err != nil {
			return err
		}
		if err := a.runner.enqueueTx(st, execution, runnerInput{ID: "unknown-original", SourceKind: "application_event", SourceID: "original-source", Mode: "steer", Text: "original work"}); err != nil {
			return err
		}
		if err := st.runnerInputState("unknown-original", "unknown"); err != nil {
			return err
		}
		_, err := st.tx.Exec(`INSERT INTO runner_calls VALUES(?,?,?,?,?,?)`, call.ID, call.ExecutionID, call.TurnID, call.NativeCallID, "digest", runnerRawJSON(call))
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := a.recoverRunnerExecutions(a.ctx); err != nil {
		t.Fatal(err)
	}
	blocked, err := a.store.readRunnerExecution(a.ctx, execution.ID)
	if err != nil || blocked.RecoveryBlocked != "application_call_unsettled" {
		t.Fatal(blocked, err)
	}
	call.State = "replied"
	call.Reply = json.RawMessage(`{"retained":true,"marker":"completed-before-crash"}`)
	if _, err := a.store.writer.Exec(`UPDATE runner_calls SET body=? WHERE id=?`, runnerRawJSON(call), call.ID); err != nil {
		t.Fatal(err)
	}
	// The same reconciliation used by the timer observes the original committed
	// reply; it neither substitutes a worker nor replays the unknown source.
	if err := a.reconcileRunnerExecutions(a.ctx, true); err != nil {
		t.Fatal(err)
	}
	recovered, err := a.store.readRunnerExecution(a.ctx, execution.ID)
	if err != nil || recovered.State != "idle" || recovered.SessionID != execution.SessionID || recovered.Generation != execution.Generation {
		t.Fatal(recovered, err)
	}
	var originalState string
	if err := a.store.readers.QueryRow(`SELECT state FROM runner_inputs WHERE id='unknown-original'`).Scan(&originalState); err != nil || originalState != "unknown" {
		t.Fatal("unknown input was replayed", originalState, err)
	}
	var recovery []byte
	if err := a.store.readers.QueryRow(`SELECT body FROM runner_inputs WHERE execution_id=? AND source_id=?`, execution.ID, "recovery/"+execution.TurnID).Scan(&recovery); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(recovery, []byte("completed-before-crash")) || !bytes.Contains(recovery, []byte("original_callback_consumption_confirmed")) {
		t.Fatal("committed result provenance was lost", string(recovery))
	}
	if _, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
		seal := runnerGroup{InstanceKey: execution.InstanceKey, ProjectID: execution.ProjectID, CancelRequested: true, Executions: []runnerExecution{}}
		_, err := st.runnerReceipt(execution.IntegrationID, runnerGroupRecordID(seal.InstanceKey), "seal", seal)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	a.mu.Lock()
	err = a.runner.scheduleLocked(execution.ID)
	a.mu.Unlock()
	if err == nil || a.runs[execution.SessionID] != nil {
		t.Fatal("a recovery claim crossed the durable group fence")
	}
	var count int
	if err := a.store.readSnapshot(a.ctx, func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT count(*) FROM runner_inputs WHERE source_id=?`, "recovery/"+execution.TurnID).Scan(&count)
	}); err != nil || count != 1 {
		t.Fatal("duplicate recovery source", count, err)
	}
}

func TestRunnerGroupPreparedOnlyAndProjectCollision(t *testing.T) {
	a, execution := runnerRegressionApp(t, "worker")
	if _, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
		seal := runnerGroup{InstanceKey: "prepared-only", ProjectID: execution.ProjectID, CancelRequested: true, Executions: []runnerExecution{}}
		_, err := st.runnerReceipt(execution.IntegrationID, runnerGroupRecordID(seal.InstanceKey), "seal", seal)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
		if err := st.runnerGroupOpen(execution.IntegrationID, execution.ProjectID, "prepared-only"); err == nil {
			t.Fatal("prepared-only instance reopened")
		}
		if _, err := st.runnerGroup(execution.IntegrationID, "other-project", "prepared-only"); err == nil {
			t.Fatal("group key changed projects")
		}
		other, err := st.runnerGroup("another-integration", execution.ProjectID, "prepared-only")
		if err != nil {
			return err
		}
		if other.CancelRequested || len(other.Executions) > 0 {
			t.Fatal("another integration observed private instance state")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
