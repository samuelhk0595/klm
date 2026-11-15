package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"
)

func TestRunnerNativeCodexOwnedBoundary(t *testing.T) {
	if os.Getenv("KLM_RUN_FACTORY_NATIVE_PROBE") != "1" {
		t.Skip("explicit real-provider probe")
	}
	if !developmentBuild {
		t.Fatal("probe requires isolated dev authentication bypass")
	}
	if runtime.GOOS != "windows" {
		t.Skip("this opt-in receipt targets the pinned Windows profiles")
	}
	harness := os.Getenv("KLM_FACTORY_PROBE_HARNESS")
	if harness == "" {
		harness = "codex"
	}
	model, version, askTool := "gpt-6-astra", "0.153.4", "request_user_input"
	originalHome, appdata, localappdata := os.Getenv("USERPROFILE"), os.Getenv("APPDATA"), os.Getenv("LOCALAPPDATA")
	originalAuthPath := filepath.Join(originalHome, ".codex", "auth.json")
	if harness == "pi" {
		originalAuthPath = filepath.Join(originalHome, ".pi", "agent", "auth.json")
		model, version, askTool = "openai-codex/gpt-6-astra", "0.85.1", "ask_user"
	}
	if harness == "opencode" {
		originalAuthPath = filepath.Join(originalHome, ".local", "share", "opencode", "auth.json")
		model, version, askTool = "openai/gpt-6-astra", "1.18.35", "question"
	}
	auth, err := os.ReadFile(originalAuthPath)
	if err != nil {
		t.Fatal("native authentication unavailable")
	}
	authHash := sha256.Sum256(auth)
	root := t.TempDir()
	home := filepath.Join(root, harness)
	if err := os.MkdirAll(home, 0700); err != nil {
		t.Fatal(err)
	}
	authPath := filepath.Join(home, "auth.json")
	nativeAuth := auth
	if harness == "opencode" {
		var providers map[string]json.RawMessage
		if err := json.Unmarshal(auth, &providers); err != nil {
			t.Fatal(err)
		}
		nativeAuth, _ = json.Marshal(map[string]json.RawMessage{"openai": providers["openai"]})
	}
	if err := os.WriteFile(authPath, nativeAuth, 0600); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(authPath)
		after, err := os.ReadFile(originalAuthPath)
		if err != nil || sha256.Sum256(after) != authHash {
			t.Error("original native auth changed")
		}
	}()
	if err := os.MkdirAll(filepath.Join(root, "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", home)
	if harness == "pi" {
		t.Setenv("PI_CODING_AGENT_DIR", home)
	}
	if harness == "opencode" {
		t.Setenv("XDG_DATA_HOME", root)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
		t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
		t.Setenv("XDG_STATE_HOME", filepath.Join(root, "state"))
		t.Setenv("HOME", filepath.Join(root, "isolated-home"))
		t.Setenv("USERPROFILE", filepath.Join(root, "isolated-home"))
		t.Setenv("OPENCODE_CONFIG_CONTENT", `{"enabled_providers":["openai"],"provider":{"openai":{"models":{"gpt-6-astra":{"name":"GPT-6 Astra"}}}}}`)
		t.Setenv("OPENCODE_DISABLE_MODELS_FETCH", "true")
		t.Setenv("OPENCODE_CONFIG", "")
		t.Setenv("OPENCODE_CONFIG_DIR", "")
	}
	t.Setenv("KLM_USER_SKILLS_DIR", filepath.Join(root, "skills"))
	t.Setenv("KLM_DEV_DISABLE_CLIENT_AUTH", "1")
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()
	dir := filepath.Join(root, "engine")
	if err := os.MkdirAll(filepath.Join(dir, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := prepareProtocolStaging(dir); err != nil {
		t.Fatal(err)
	}
	store, err := openOrMigrateSQLite(ctx, dir, prepareSQLiteImport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer store.close()
	node, err := exec.LookPath("node.exe")
	if err != nil {
		t.Fatal(err)
	}
	b := binary{path: filepath.Join(localappdata, "Programs", "OpenAI", "Codex", "bin", "codex.exe"), node: node}
	if harness == "pi" {
		b.path = node
		b.args = []string{filepath.Join(appdata, "npm", "node_modules", "@earendil-works", "pi-coding-agent", "dist", "cli.js")}
	}
	if harness == "opencode" {
		b.path = filepath.Join(appdata, "npm", "node_modules", "opencode-ai", "bin", "opencode.exe")
	}
	a := &app{ctx: ctx, dir: dir, store: store, sqlChanges: &sqliteChangeBus{}, runs: map[string]*turn{}, runtimes: map[string]*sessionRuntime{}, binaries: map[string]binary{harness: b}, harnesses: []Harness{{ID: harness, Available: true, Compatible: true, InstalledVersion: version, RequiredVersion: version}}}
	if err := a.initializeRunner(ctx); err != nil {
		t.Fatal(err)
	}
	defer func() { cancel(); a.wg.Wait(); a.shutdownSessionRuntimes() }()
	a.clientAuth, err = newClientAuthStorage(dir, store)
	if err != nil {
		t.Fatal(err)
	}
	project := Project{ID: newID(), Name: "Native runner probe", Folder: filepath.Join(root, "project")}
	if err := os.MkdirAll(filepath.Join(project.Folder, ".klm", "agents"), 0700); err != nil {
		t.Fatal(err)
	}
	document := agentDocument{SourceUID: newID(), AgentDefinition: AgentDefinition{Name: "Native probe", Enabled: true, DefaultHarness: harness, Model: model, Effort: "low", Prompt: "Perform only the harmless supplied native feasibility task. Never delegate, edit files or access the network."}}
	agentBytes, err := toml.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	agentPath := filepath.Join(project.Folder, ".klm", "agents", "native-probe.toml")
	if err := os.WriteFile(agentPath, agentBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.sessionTransaction(ctx, func(st *sqliteSessionTx) error { return st.putRoot("projects", project.ID, project) }); err != nil {
		t.Fatal(err)
	}
	handler := a.routes()
	owner := ""
	request := func(method, path string, body any) []byte {
		t.Helper()
		encoded, _ := json.Marshal(body)
		r := httptest.NewRequest(method, "http://localhost"+path, bytes.NewReader(encoded))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-KLM-Native", "1")
		r.Header.Set("X-KLM-Read-Contract", "3")
		if owner != "" {
			r.Header.Set("X-KLM-Runner-ID", owner)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code >= 400 {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w.Body.Bytes()
	}
	var connection struct {
		ID string `json:"integration_id"`
	}
	if err := json.Unmarshal(request("POST", "/api/runner/connections", map[string]string{"operation_id": "probe-connect", "key": "native-probe-" + newID()}), &connection); err != nil {
		t.Fatal(err)
	}
	owner = connection.ID
	manifest := []map[string]any{{"name": "probe_record", "description": "Record the harmless probe marker.", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"marker": map[string]any{"type": "string"}}, "required": []string{"marker"}, "additionalProperties": false}}}
	var coordinator struct {
		Result runnerExecution `json:"result"`
	}
	if err := json.Unmarshal(request("POST", "/api/runner/conversations", map[string]any{"operation_id": "probe-conversation", "key": "probe-conversation", "project_id": project.ID, "capture": runnerCapture{Harness: harness, Model: model, Effort: "low", Manifest: manifest, Instructions: "Authorized harmless native runner check. Never delegate or use files/network. Use the owned shell/job tools, not native shell paths. Questions require an actual human response."}}), &coordinator); err != nil {
		t.Fatal(err)
	}
	command := "Write-Output ('START:'+$PID); Start-Sleep -Seconds 60; Write-Output ('END:'+$PID)"
	request("POST", "/api/runner/conversations/"+coordinator.Result.ID+"/inputs", runnerInput{OperationID: "probe-human", SourceKind: "human", SourceID: "original-probe-request", Mode: "steer", Text: "Call shell_start once with command " + command + ". After it returns running, call probe_record with marker native-owned-marker, then ask the human to choose the required final label Alpha or Beta using " + askTool + ". Do not answer the question, wait for the job, or poll."})
	var worker struct {
		Result runnerExecution `json:"result"`
	}
	workerRequest := map[string]any{"operation_id": "probe-assignment", "key": "one-assignment", "project_id": project.ID, "agent": "native-probe", "source_uid": document.SourceUID, "agent_revision": revision(agentBytes), "prompt": "Before any other work, ask the human to choose the required final label Alpha or Beta using native " + askTool + ". Do not answer the question or use other tools.", "capture": runnerCapture{Manifest: []map[string]any{}, Instructions: "Harmless worker question probe; no delegation or files/network."}}
	if err := json.Unmarshal(request("POST", "/api/runner/work", workerRequest), &worker); err != nil {
		t.Fatal(err)
	}
	var duplicate struct {
		Result runnerExecution `json:"result"`
	}
	if err := json.Unmarshal(request("POST", "/api/runner/work", workerRequest), &duplicate); err != nil || duplicate.Result.ID != worker.Result.ID {
		t.Fatal("worker retry did not return the original admission", err)
	}
	marker := false
	questions := map[string]bool{}
	wait := func(label string, condition func() bool) {
		t.Helper()
		deadline := time.NewTimer(50 * time.Second)
		defer deadline.Stop()
		for {
			if condition() {
				return
			}
			select {
			case <-time.After(50 * time.Millisecond):
			case <-ctx.Done():
				for _, id := range []string{coordinator.Result.ID, worker.Result.ID} {
					current, _ := store.readRunnerExecution(context.Background(), id)
					t.Logf("execution %s state=%s error=%s", id, current.State, current.Error)
					page, _ := store.readHistory(context.Background(), current.SessionID, "", "", 25)
					for _, item := range page.Events {
						if item.Event != nil && (item.Event.Type == "error" || item.Event.Type == "status" || item.Event.Type == "assistant") {
							t.Logf("native %s: %s", item.Event.Type, boundedText(item.Event.Text, 2048))
						}
					}
				}
				t.Fatal(label, "native probe context expired")
			case <-deadline.C:
				for _, id := range []string{coordinator.Result.ID, worker.Result.ID} {
					current, _ := store.readRunnerExecution(context.Background(), id)
					t.Logf("execution %s state=%s error=%s", id, current.State, current.Error)
					page, _ := store.readHistory(context.Background(), current.SessionID, "", "", 25)
					for _, item := range page.Events {
						if item.Event != nil && (item.Event.Type == "error" || item.Event.Type == "status" || item.Event.Type == "assistant") {
							t.Logf("native %s: %s", item.Event.Type, boundedText(item.Event.Text, 2048))
						}
					}
				}
				t.Fatal(label, "deadline")
			}
		}
	}
	wait("native operational call and exact questions", func() bool {
		rows, err := store.readers.Query(`SELECT body FROM runner_calls`)
		if err != nil {
			t.Fatal(err)
		}
		calls := []runnerToolCall{}
		for rows.Next() {
			var body []byte
			var call runnerToolCall
			if err := rows.Scan(&body); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(body, &call); err != nil {
				t.Fatal(err)
			}
			calls = append(calls, call)
		}
		rows.Close()
		for _, call := range calls {
			if call.Name == "probe_record" && call.State == "pending" {
				marker = bytes.Contains(call.Arguments, []byte("native-owned-marker"))
				request("POST", "/api/runner/tool-calls/"+call.ID+"/reply", map[string]any{"operation_id": "reply-" + call.ID, "result": map[string]bool{"recorded": true}})
			}
		}
		rows, err = store.readers.Query(`SELECT execution_id FROM runner_requests WHERE state='pending'`)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				t.Fatal(err)
			}
			questions[id] = true
		}
		rows.Close()
		return marker && questions[coordinator.Result.ID] && questions[worker.Result.ID]
	})
	current, err := store.readRunnerExecution(ctx, coordinator.Result.ID)
	if err != nil {
		t.Fatal(err)
	}
	request("POST", "/api/runner/conversations/"+current.ID+"/stop", map[string]string{"operation_id": "stop-coordinator", "expected_turn_id": current.TurnID})
	wait("Foreman Stop", func() bool {
		current, _ := store.readRunnerExecution(ctx, coordinator.Result.ID)
		return current.State == "stopped"
	})
	var pending int
	if err := store.readers.QueryRow(`SELECT count(*) FROM runner_requests WHERE execution_id=? AND state='pending'`, worker.Result.ID).Scan(&pending); err != nil || pending != 1 {
		t.Fatalf("Foreman Stop affected independent worker request: %d %v", pending, err)
	}
	request("POST", "/api/runner/work/"+worker.Result.ID+"/cancel", map[string]string{"operation_id": "cancel-worker"})
	wait("exact work cancellation", func() bool {
		current, _ := store.readRunnerExecution(ctx, worker.Result.ID)
		return current.State == "cancelled"
	})
	var running int
	if err := store.readers.QueryRow(`SELECT count(*) FROM runner_jobs WHERE state IN ('starting','running')`).Scan(&running); err != nil || running != 0 {
		t.Fatal("owned job drainage was not confirmed", err)
	}
	after, err := os.ReadFile(agentPath)
	if err != nil || !bytes.Equal(after, agentBytes) {
		t.Fatal("saved-Agent input changed", err)
	}
	t.Log(fmt.Sprintf("actual %s operational marker, two exact unanswered native questions, original worker receipt, Foreman-only Stop and independent work cancellation; project=%s", harness, project.ID))
}
