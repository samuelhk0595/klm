package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRunnerDurableBoundary(t *testing.T) {
	if !developmentBuild {
		t.Skip("uses the explicit isolated dev authentication bypass")
	}
	t.Setenv("KLM_DEV_DISABLE_CLIENT_AUTH", "1")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dir := t.TempDir()
	store, err := openOrMigrateSQLite(ctx, dir, prepareSQLiteImport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer store.close()
	a := &app{ctx: ctx, dir: dir, store: store, sqlChanges: &sqliteChangeBus{}, runs: map[string]*turn{}, runtimes: map[string]*sessionRuntime{}, binaries: map[string]binary{}}
	if err := a.initializeRunner(ctx); err != nil {
		t.Fatal(err)
	}
	a.clientAuth, err = newClientAuthStorage(dir, store)
	if err != nil {
		t.Fatal(err)
	}
	project := Project{ID: newID(), Name: "Disposable", Folder: t.TempDir()}
	execution := runnerExecution{ID: newID(), IntegrationID: newID(), ProjectID: project.ID, Key: "one-assignment", Kind: "worker", SessionID: newID(), Generation: "1", State: "idle"}
	session := Session{ID: execution.SessionID, Origin: runnerOrigin, IntegrationID: execution.IntegrationID, RunnerID: execution.ID, Role: "runner_worker", ProjectID: project.ID, Harness: "codex", Model: "explicit-model", YOLO: true, Status: "idle", CreatedAt: now(), UpdatedAt: now()}
	a.mu.Lock()
	err = a.writeSQLLocked(ctx, func(st *sqliteSessionTx) error {
		if err := st.putRoot("projects", project.ID, project); err != nil {
			return err
		}
		if err := st.createSession(session, dir); err != nil {
			return err
		}
		if err := st.putRunnerExecution(execution); err != nil {
			return err
		}
		_, err := st.runnerReceipt(execution.IntegrationID, "accepted", "same-body", map[string]string{"execution_id": execution.ID})
		return err
	})
	a.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
		_, err := st.runnerReceipt(execution.IntegrationID, "accepted", "different-body", nil)
		return err
	}); err == nil {
		t.Fatal("changed operation body was accepted")
	}
	if _, err := store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
		changed := session
		changed.Origin = ""
		return st.putSessionMeta(changed)
	}); err == nil {
		t.Fatal("external origin was mutable")
	}
	if _, err := store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
		other := execution
		other.ID = newID()
		other.SessionID = newID()
		return st.putRunnerExecution(other)
	}); err == nil {
		t.Fatal("Assignment key admitted a second execution")
	}
	handler := a.routes()
	a.runs[session.ID] = &turn{runner: &runnerTurn{ExecutionID: execution.ID}}
	a.runs["ordinary-turn"] = &turn{}
	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest("GET", "http://localhost/api/health", nil))
	var healthBody struct {
		RunningSessions int `json:"runningSessions"`
	}
	if err := json.Unmarshal(health.Body.Bytes(), &healthBody); err != nil || health.Code != 200 || healthBody.RunningSessions != 1 {
		t.Fatalf("ordinary health count exposed external work: %d %s (%v)", health.Code, health.Body.String(), err)
	}
	delete(a.runs, session.ID)
	delete(a.runs, "ordinary-turn")
	for _, path := range []string{"/api/sessions/" + session.ID, "/api/sessions/" + session.ID + "/history", "/api/sessions/" + session.ID + "/metadata", "/api/sessions/" + session.ID + "/models", "/api/sessions/" + session.ID + "/quota"} {
		request := httptest.NewRequest("GET", "http://localhost"+path, nil)
		request.Header.Set("X-KLM-Read-Contract", "3")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != 404 {
			t.Fatalf("hidden exact ID: %s -> %d", path, response.Code)
		}
	}
	request := httptest.NewRequest("GET", "http://localhost/api/projects/"+project.ID+"/sessions", nil)
	request.Header.Set("X-KLM-Read-Contract", "3")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 200 || bytes.Contains(response.Body.Bytes(), []byte(session.ID)) {
		t.Fatalf("ordinary inventory exposed external session: %d %s", response.Code, response.Body.String())
	}
	subscription, err := a.sqlChanges.subscribe(sqliteChangeScope{Kind: "global"})
	if err != nil {
		t.Fatal(err)
	}
	defer a.sqlChanges.unsubscribe(subscription)
	a.mu.Lock()
	err = a.writeSQLLocked(ctx, func(st *sqliteSessionTx) error {
		_, err := st.appendEvent(session.ID, event("assistant", "Private native evidence"))
		return err
	})
	a.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	select {
	case notice := <-subscription.notices:
		t.Fatalf("external evidence published ordinary notice: %s", notice)
	case <-time.After(25 * time.Millisecond):
	}
	var count int
	if err := store.readers.QueryRow(`SELECT count(*) FROM runner_events WHERE execution_id=? AND type='history_changed'`, execution.ID).Scan(&count); err != nil || count == 0 {
		t.Fatalf("runner evidence was not retained: %d %v", count, err)
	}
	page, err := store.readHistory(ctx, session.ID, "", "", 25)
	if err != nil || len(page.Events) != 1 {
		t.Fatalf("native evidence lost: %+v %v", page, err)
	}
	encoded, _ := json.Marshal(page)
	if !bytes.Contains(encoded, []byte("Private native evidence")) {
		t.Fatal("native transcript was replaced by a summary")
	}
}
