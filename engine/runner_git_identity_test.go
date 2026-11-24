package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestRunnerGitIdentityDurableContext(t *testing.T) {
	for _, mode := range []string{"configured", "shared", "legacy"} {
		t.Run(mode, func(t *testing.T) {
			a, execution := runnerRegressionApp(t, "worker")
			foreign := Project{ID: newID(), Name: "Other project at runner CWD", Folder: t.TempDir()}
			if _, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
				return st.putRoot("projects", foreign.ID, foreign)
			}); err != nil {
				t.Fatal(err)
			}
			mux := http.NewServeMux()
			a.projectGitIdentityRoutes(mux, fixtureGitVerifier)
			add := func(projectID, alias, token string, makeDefault bool) GitIdentity {
				return gitIdentityList(t, gitIdentityRequest(
					t,
					mux,
					"POST",
					"/api/projects/"+projectID+"/git-identities",
					map[string]any{"alias": alias, "token": token, "makeDefault": makeDefault},
					201,
				))[0]
			}
			add(foreign.ID, "Work", "klm-synthetic-github-bob", true)
			var original GitIdentity
			if mode == "configured" {
				original = add(execution.ProjectID, "Work", "klm-synthetic-github-alice", true)
			}
			if mode != "legacy" {
				if err := a.bindRunnerProjectGit(a.ctx, &execution); err != nil {
					t.Fatal(err)
				}
			}
			admitted := execution.ProjectGit
			prepared := runnerPreparation{
				ProjectID: execution.ProjectID, InstanceKey: "frozen-instance", ProjectGit: admitted,
				Workspace: &runnerWorkspace{Path: foreign.Folder}, Capture: execution.Applied,
			}
			execution.Workspace = prepared.Workspace
			execution.PreparationID, execution.InstanceKey = "frozen-preparation", prepared.InstanceKey
			execution.State = "running"
			raw := json.RawMessage(`{"command":"fixture must never execute","wait_ms":0}`)
			callID := "frozen-shell-call"
			job := runnerJob{
				ID:          revision([]byte(execution.ID + "/" + execution.TurnID + "/" + callID)),
				ExecutionID: execution.ID, TurnID: execution.TurnID, State: "completed",
				Command: "fixture must never execute", Directory: foreign.Folder,
				ProjectGit: admitted, EnvironmentDigest: revision(runnerRawJSON(admitted)),
			}
			if _, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
				if _, err := st.runnerReceipt(execution.IntegrationID, execution.PreparationID, "same-preparation", prepared); err != nil {
					return err
				}
				if err := st.putRunnerExecution(execution); err != nil {
					return err
				}
				meta, err := st.sessionMeta(execution.SessionID)
				if err != nil {
					return err
				}
				meta.ExecutionCWD = foreign.Folder
				if err := st.putSessionMeta(meta); err != nil {
					return err
				}
				_, err = st.tx.ExecContext(st.ctx, `INSERT INTO runner_jobs VALUES(?,?,?,?,?,?,?)`,
					job.ID, job.ExecutionID, job.TurnID, callID, revision(raw), job.State, runnerRawJSON(job),
				)
				return err
			}); err != nil {
				t.Fatal(err)
			}

			// Change the registered project's current configuration after admission.
			// Neither the foreign CWD nor this edit may retarget an existing capture.
			if mode != "configured" {
				add(execution.ProjectID, "Work", "klm-synthetic-github-alice", true)
			}
			add(execution.ProjectID, "Personal", "klm-synthetic-github-bob", true)
			if mode == "configured" {
				gitIdentityRequest(
					t,
					mux,
					"PATCH",
					"/api/projects/"+execution.ProjectID+"/git-identities/"+original.ID,
					map[string]string{"alias": "Renamed"},
					200,
				)
			}
			loaded, err := a.store.readRunnerExecution(a.ctx, execution.ID)
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest("GET", "http://localhost", nil)
			loadedPreparation, err := a.readRunnerPreparation(request, execution.IntegrationID, execution.PreparationID)
			if err != nil {
				t.Fatal(err)
			}
			for _, capture := range []*ProjectGitIdentitySnapshot{loaded.ProjectGit, loadedPreparation.ProjectGit} {
				if !reflect.DeepEqual(capture, admitted) {
					t.Fatal("durable owner changed the admitted capture")
				}
			}
			if loadedPreparation.Workspace.Path != foreign.Folder {
				t.Fatal("custom runner directory was not retained")
			}
			p := &adapter{app: a, cwd: foreign.Folder, turn: &turn{
				ctx: a.ctx, projectGit: loaded.ProjectGit,
				runner: &runnerTurn{ExecutionID: loaded.ID, TurnID: loaded.TurnID},
			}}
			if mode == "legacy" {
				if _, err := a.runner.jobCall(a.ctx, p, "shell_start", callID, raw); err == nil || !strings.Contains(err.Error(), "Git identity capture") {
					t.Fatal("unbound legacy running turn was allowed to launch a shell", err)
				}
				recovery := loaded
				recovery.ProcessesDrained = true
				if _, err := a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
					return a.runner.recoverDrainedTx(st, &recovery)
				}); err != nil || recovery.RecoveryBlocked != "git_identity_capture_missing" {
					t.Fatal("legacy recovery acquired a new identity context", err)
				}
			}
			// This is the same binding seam used by execution admission and resumed
			// turn scheduling. Only a missing legacy capture acquires current refs.
			if err := a.bindRunnerProjectGit(a.ctx, &loaded); err != nil {
				t.Fatal(err)
			}
			if mode == "legacy" {
				current, err := a.captureProjectGitIdentities(a.ctx, execution.ProjectID)
				if err != nil || !reflect.DeepEqual(loaded.ProjectGit, current) {
					t.Fatal("first new legacy admission did not capture its registered project", err)
				}
			} else if !reflect.DeepEqual(loaded.ProjectGit, admitted) {
				t.Fatal("resumed execution refreshed its admitted capture")
			}
			if loaded.ProjectGit.ProjectID != execution.ProjectID {
				t.Fatal("runner Git ownership followed its CWD")
			}
			if mode == "configured" {
				for _, alias := range []string{"", "Work"} {
					token, err := a.resolveProjectGitIdentity(a.ctx, loaded.ProjectGit, alias)
					if err != nil || token != "klm-synthetic-github-alice" {
						t.Fatal("admitted default/alias followed settings or the foreign CWD", err)
					}
				}
			}
			p.turn.projectGit = loaded.ProjectGit
			result, err := a.runner.jobCall(a.ctx, p, "shell_start", callID, raw)
			if err != nil {
				t.Fatal(err)
			}
			reused := result.(runnerJob)
			if !reflect.DeepEqual(reused, job) || len(a.runner.jobs) != 0 || a.harnessConsumers.Load() != 0 {
				t.Fatal("durable shell replay changed context or started a process")
			}
			for _, record := range []any{loaded, loadedPreparation, reused} {
				encoded := runnerRawJSON(record)
				for _, forbidden := range []string{"klm-synthetic", "encrypted_token", "launch_bearer", "helper_bearer", "launch_url"} {
					if bytes.Contains(bytes.ToLower(encoded), []byte(forbidden)) {
						t.Fatalf("durable runner record contains secret/authority data: %s", forbidden)
					}
				}
			}
			assertRunnerGitReferenceFields(t, loaded.ProjectGit)
		})
	}
}

func assertRunnerGitReferenceFields(t *testing.T, capture *ProjectGitIdentitySnapshot) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(runnerRawJSON(capture), &fields); err != nil || len(fields) != 3 {
		t.Fatal("Git capture must contain exactly project/default/identity references", err)
	}
	for _, key := range []string{"projectId", "defaultId", "identities"} {
		if _, ok := fields[key]; !ok {
			t.Fatal("unexpected Git capture contract")
		}
	}
	var refs []map[string]json.RawMessage
	if err := json.Unmarshal(fields["identities"], &refs); err != nil || refs == nil {
		t.Fatal("identity references must be an explicit list", err)
	}
	for _, ref := range refs {
		if len(ref) != 3 || ref["id"] == nil || ref["alias"] == nil || ref["credentialVersion"] == nil {
			t.Fatal("identity capture includes fields other than stable references")
		}
	}
}
