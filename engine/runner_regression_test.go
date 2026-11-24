package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func runnerRegressionApp(t *testing.T, kind string) (*app, runnerExecution) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	dir := t.TempDir()
	store, err := openOrMigrateSQLite(ctx, dir, prepareSQLiteImport, nil)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	a := &app{
		ctx: ctx, dir: dir, store: store, sqlChanges: &sqliteChangeBus{},
		runs: map[string]*turn{}, runtimes: map[string]*sessionRuntime{}, binaries: map[string]binary{},
	}
	t.Cleanup(func() { cancel(); a.wg.Wait(); store.close() })
	if err := a.initializeRunner(ctx); err != nil {
		t.Fatal(err)
	}
	project := Project{ID: newID(), Name: "Runner regression", Folder: t.TempDir()}
	execution := runnerExecution{
		ID: newID(), IntegrationID: newID(), ProjectID: project.ID, Key: newID(),
		Kind: kind, SessionID: newID(), Generation: "1", TurnID: "original-turn", State: "idle",
	}
	execution.Applied = runnerCapture{Harness: "codex", Model: "explicit-model", Instructions: "Original capture"}
	execution.Desired = execution.Applied
	meta := Session{
		ID: execution.SessionID, Origin: runnerOrigin, IntegrationID: execution.IntegrationID,
		RunnerID: execution.ID, Role: "runner_" + kind, ProjectID: project.ID,
		Harness: "codex", Model: "explicit-model", YOLO: true, Status: "idle", CreatedAt: now(), UpdatedAt: now(),
	}
	_, err = store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
		if err := st.putRoot("projects", project.ID, project); err != nil {
			return err
		}
		if err := st.createSession(meta, dir); err != nil {
			return err
		}
		return st.putRunnerExecution(execution)
	})
	if err != nil {
		t.Fatal(err)
	}
	return a, execution
}

func runnerRegressionJob(t *testing.T, a *app, execution runnerExecution, state string) runnerJob {
	t.Helper()
	job := runnerJob{ID: newID(), ExecutionID: execution.ID, TurnID: execution.TurnID, State: state}
	body, err := runnerJSON(job)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.store.writer.Exec(`INSERT INTO runner_jobs VALUES(?,?,?,?,?,?,?)`,
		job.ID, job.ExecutionID, job.TurnID, newID(), "command-digest", job.State, body)
	if err != nil {
		t.Fatal(err)
	}
	return job
}

func TestRunnerJobCompletionWakeAndSettlement(t *testing.T) {
	for _, initial := range []string{"running", "native_completed"} {
		t.Run(initial, func(t *testing.T) {
			a, execution := runnerRegressionApp(t, "worker")
			execution.State = initial
			if _, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
				return st.putRunnerExecution(execution)
			}); err != nil {
				t.Fatal(err)
			}
			job := runnerRegressionJob(t, a, execution, "running")
			job.State = "completed"
			active := &turn{
				ctx: a.ctx, steerWake: make(chan struct{}, 1),
				runner: &runnerTurn{ExecutionID: execution.ID, TurnID: execution.TurnID},
			}
			// Hold the native dispatch slot: this regression never launches a harness.
			a.runs[execution.SessionID] = active
			a.runner.jobCompleted(job)
			select {
			case <-active.steerWake:
			default:
				t.Fatal("completion did not wake the active native steering consumer")
			}
			a.mu.Lock()
			a.runner.finishedLocked(&adapter{app: a, id: execution.SessionID, turn: active, completed: true}, nil)
			a.mu.Unlock()
			current, err := a.store.readRunnerExecution(a.ctx, execution.ID)
			if err != nil || current.State != "idle" {
				t.Fatalf("pending result became unschedulable: state=%s err=%v", current.State, err)
			}
			var pending int
			if err := a.store.readers.QueryRow(`SELECT count(*) FROM runner_inputs WHERE execution_id=? AND source_kind='job_result' AND state='retained'`, execution.ID).Scan(&pending); err != nil || pending != 1 {
				t.Fatalf("completion input lost: count=%d err=%v", pending, err)
			}
			var terminal int
			if err := a.store.readers.QueryRow(`SELECT count(*) FROM runner_events WHERE execution_id=? AND type='execution' AND json_extract(body,'$.state')='native_completed'`, execution.ID).Scan(&terminal); err != nil {
				t.Fatal(err)
			}
			if initial == "running" && terminal != 0 {
				t.Fatal("settlement published terminal work before the pending job result")
			}
		})
	}
}

func TestRunnerCancelRetainsPersistedUncertainty(t *testing.T) {
	for _, state := range []string{"starting", "running", "unknown", "lost_native", "completed"} {
		t.Run(state, func(t *testing.T) {
			a, execution := runnerRegressionApp(t, "worker")
			execution.State = "cancelling"
			if state == "lost_native" {
				execution.Error = "Restarted native execution has not confirmed cessation."
			} else {
				runnerRegressionJob(t, a, execution, state)
			}
			if _, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
				return st.putRunnerExecution(execution)
			}); err != nil {
				t.Fatal(err)
			}
			a.runner.stopOwned(execution, nil) // No reconstructed process or native-turn handle.
			current, err := a.store.readRunnerExecution(a.ctx, execution.ID)
			if err != nil {
				t.Fatal(err)
			}
			if state == "completed" {
				if current.State != "cancelled" {
					t.Fatalf("positively settled job blocked cancellation: %+v", current)
				}
				return
			}
			if current.State != "interrupted" || current.Error == "" {
				t.Fatalf("missing handles erased retained uncertainty: %+v", current)
			}
		})
	}
}

func TestRunnerGenerationConversation(t *testing.T) {
	a, execution := runnerRegressionApp(t, "coordinator")
	_, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
		for i := range 28 {
			if _, err := st.appendEvent(execution.SessionID, event("assistant", fmt.Sprintf("Older answer %d", i))); err != nil {
				return err
			}
		}
		input := runnerInput{ID: "constraint", SourceKind: "human", SourceID: "actual-human-source", Mode: "steer", Text: "Preserve the release branch; do not deploy."}
		if err := a.runner.enqueueTx(st, execution, input); err != nil {
			return err
		}
		if err := st.runnerInputState(input.ID, "included"); err != nil {
			return err
		}
		e := event("runner_input", input.Text)
		e.ID = input.ID
		if _, err := st.appendEvent(execution.SessionID, e); err != nil {
			return err
		}
		_, err := st.appendEvent(execution.SessionID, event("assistant", "I will preserve that operational constraint."))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	execution.Desired.Instructions = "New capture, same conversation"
	a.mu.Lock()
	a.runner.applyDesiredLocked(execution)
	a.mu.Unlock()
	a.wg.Wait()
	current, err := a.store.readRunnerExecution(a.ctx, execution.ID)
	if err != nil || current.SessionID == execution.SessionID || current.Generation != "2" {
		t.Fatalf("generation did not rotate: %+v %v", current, err)
	}
	encoded, _ := json.Marshal(current.RecentConversation)
	if len(current.RecentConversation) > 24 || len(encoded) > 32<<10 {
		t.Fatal("generation context is unbounded")
	}
	if !strings.Contains(string(encoded), "operational constraint") || !strings.Contains(string(encoded), "do not deploy") {
		t.Fatal("fresh generation lost the recent operational exchange")
	}
	var found bool
	for _, entry := range current.RecentConversation {
		if entry.EventID == "constraint" {
			found = entry.SourceKind == "human" && entry.SourceID == "actual-human-source" && entry.SessionID == execution.SessionID
		}
	}
	if !found {
		t.Fatal("old human provenance was lost or promoted into a new input")
	}
	var inputs int
	if err := a.store.readers.QueryRow(`SELECT count(*) FROM runner_inputs WHERE execution_id=?`, execution.ID).Scan(&inputs); err != nil || inputs != 1 {
		t.Fatalf("reference context created input authorization: %d %v", inputs, err)
	}
}

func TestRunnerOrdinaryInvalidationOnly(t *testing.T) {
	for _, connected := range []bool{false, true} {
		t.Run(fmt.Sprintf("connected=%t", connected), func(t *testing.T) {
			a, execution := runnerRegressionApp(t, "coordinator")
			if connected {
				if _, err := a.store.writer.Exec(`INSERT INTO runner_integrations VALUES(?,?,?)`, execution.IntegrationID, "paired-client", "application-key"); err != nil {
					t.Fatal(err)
				}
			}
			subscription, err := a.sqlChanges.subscribe(sqliteChangeScope{Kind: "global"})
			if err != nil {
				t.Fatal(err)
			}
			defer a.sqlChanges.unsubscribe(subscription)
			for _, mixed := range []bool{false, true} {
				a.mu.Lock()
				err := a.writeSQLLocked(a.ctx, func(st *sqliteSessionTx) error {
					st.changes.Invalidated["harnesses"] = true
					if mixed {
						_, err := st.appendEvent(execution.SessionID, event("assistant", "Private runner evidence"))
						return err
					}
					return nil
				})
				a.mu.Unlock()
				if err != nil {
					t.Fatal(err)
				}
				select {
				case notice := <-subscription.notices:
					if strings.Contains(string(notice), execution.SessionID) {
						t.Fatal("ordinary notice exposed hidden execution")
					}
				default:
					t.Fatalf("ordinary harness invalidation dropped (mixed external=%t)", mixed)
				}
			}
		})
	}
}
