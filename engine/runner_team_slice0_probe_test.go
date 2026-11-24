package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Explicitly commissioned, retained, real-provider workload. It never uses the
// installed engine, its ports, its database, or its control endpoint.
func TestRunnerTeamSlice0Process(t *testing.T) {
	if os.Getenv("KLM_TEAM_SLICE0_CHILD") != "1" {
		t.Skip("owned Slice 0 child only")
	}
	root := os.Getenv("KLM_TEAM_SLICE0_ROOT")
	ctx, cancel := context.WithCancel(context.Background())
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
	node, err := exec.LookPath("node.exe")
	if err != nil {
		t.Fatal(err)
	}
	b := binary{path: filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "OpenAI", "Codex", "bin", "codex.exe"), node: node}
	a := &app{
		ctx: ctx, dir: dir, store: store, sqlChanges: &sqliteChangeBus{},
		runs: map[string]*turn{}, runtimes: map[string]*sessionRuntime{},
		binaries:  map[string]binary{"codex": b},
		harnesses: []Harness{{ID: "codex", Available: true, Compatible: true, InstalledVersion: "0.153.4", RequiredVersion: "0.153.4"}},
	}
	if err := a.initializeRunner(ctx); err != nil {
		t.Fatal(err)
	}
	a.clientAuth, err = newClientAuthStorage(dir, store)
	if err != nil {
		t.Fatal(err)
	}
	project := Project{ID: "slice0-project", Name: "Owned Slice 0", Folder: filepath.Join(root, "repository")}
	if _, err := store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
		return st.putRoot("projects", project.ID, project)
	}); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: a.routes()}
	go server.Serve(listener)
	a.mu.Lock()
	err = a.runner.resumeUnclaimedLocked()
	a.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("SLICE0 http://" + listener.Addr().String())
	_, _ = io.Copy(io.Discard, os.Stdin)
	cancel()
	a.wg.Wait()
	a.shutdownSessionRuntimes()
	_ = server.Close()
	if err := store.close(); err != nil {
		t.Fatal(err)
	}
}

func TestRunnerTeamSlice0(t *testing.T) {
	if os.Getenv("KLM_RUN_TEAM_SLICE0") != "1" {
		t.Skip("explicit real-provider/restart workload")
	}
	if !developmentBuild {
		t.Fatal("owned dev authentication bypass is required")
	}
	root := os.Getenv("KLM_TEAM_SLICE0_ROOT")
	if !filepath.IsAbs(root) {
		t.Fatal("supply an absolute new retained evidence directory")
	}
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	rows := []map[string]any{}
	record := func(kind string, value any) {
		t.Helper()
		rows = append(rows, map[string]any{"at": now(), "kind": kind, "value": value})
		body, _ := json.MarshalIndent(rows, "", "  ")
		if err := os.WriteFile(filepath.Join(root, "evidence.json"), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	git := func(cwd string, args ...string) {
		t.Helper()
		command := exec.Command("git", args...)
		command.Dir, command.Env = cwd, runnerGitEnvironment()
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("explicit Foreman Git preparation: %s: %v", output, err)
		}
	}
	repository := filepath.Join(root, "repository")
	// The focused driver acts as Foreman and explicitly chooses existing HEAD.
	// Clone existing objects; this workload creates no commits.
	git(root, "clone", "--shared", "--no-checkout", filepath.Dir(mustWorkingDirectory(t)), repository)
	worktree := filepath.Join(root, "worktree")
	git(repository, "worktree", "add", "--detach", worktree, "HEAD")
	if err := os.MkdirAll(filepath.Join(repository, ".klm", "agents"), 0700); err != nil {
		t.Fatal(err)
	}
	document := agentDocument{SourceUID: newID(), AgentDefinition: AgentDefinition{
		Name: "Slice0 member", Enabled: true, DefaultHarness: "codex",
		Model: "gpt-6-astra", Effort: "low", Prompt: "Perform the supplied harmless file task and retain its result. Never delegate, access credentials or use the network.",
	}}
	agentBytes, err := toml.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	agentPath := filepath.Join(repository, ".klm", "agents", "slice0-member.toml")
	if err := os.WriteFile(agentPath, agentBytes, 0600); err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	authPath := filepath.Join(home, ".codex", "auth.json")
	auth, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatal("existing protected Codex provider authentication is unavailable")
	}
	originalHash := sha256.Sum256(auth)
	// Native authentication stays outside project/config/evidence trees.
	protectedHome, err := os.MkdirTemp(filepath.Join(os.Getenv("LOCALAPPDATA"), "Temp", "opencode"), "slice0-native-auth-")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(protectedHome, "auth.json"), auth, 0600); err != nil {
		t.Fatal(err)
	}
	defer func() {
		// Retain the native session database/transcripts for later inspection.
		// Only the temporary provider auth file is removed at handoff.
		removeErr := os.Remove(filepath.Join(protectedHome, "auth.json"))
		after, err := os.ReadFile(authPath)
		record("protected_auth_cleanup", map[string]any{
			"original_unchanged":     err == nil && sha256.Sum256(after) == originalHash,
			"temporary_auth_removed": removeErr == nil || os.IsNotExist(removeErr),
			"retained_native_home":   protectedHome,
		})
	}()
	t.Setenv("CODEX_HOME", protectedHome)
	t.Setenv("KLM_DEV_DISABLE_CLIENT_AUTH", "1")
	t.Setenv("KLM_USER_SKILLS_DIR", filepath.Join(root, "skills"))
	if err := os.Mkdir(filepath.Join(root, "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	var command *exec.Cmd
	var stdin io.WriteCloser
	base, owner := "", ""
	start := func() {
		t.Helper()
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		command = exec.Command(executable, "-test.run=^TestRunnerTeamSlice0Process$", "-test.timeout=180s")
		command.Env = append(os.Environ(), "KLM_TEAM_SLICE0_CHILD=1")
		stdin, err = command.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		stdout, err := command.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		command.Stderr = os.Stderr
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		ready := make(chan string, 1)
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				if strings.HasPrefix(scanner.Text(), "SLICE0 ") {
					ready <- strings.TrimPrefix(scanner.Text(), "SLICE0 ")
				}
			}
		}()
		select {
		case base = <-ready:
			record("owned_engine_started", map[string]any{"pid": command.Process.Pid, "base": base})
		case <-time.After(20 * time.Second):
			_ = command.Process.Kill()
			t.Fatal("owned engine startup deadline")
		}
	}
	stop := func() {
		t.Helper()
		if command == nil {
			return
		}
		_ = stdin.Close()
		err := command.Wait()
		record("owned_engine_stopped", map[string]any{"pid": command.Process.Pid, "clean": err == nil})
		command = nil
		if err != nil {
			t.Fatal(err)
		}
	}
	start()
	defer stop()
	client := &http.Client{Timeout: 15 * time.Second}
	request := func(method, path string, body any, target any) {
		t.Helper()
		r, err := http.NewRequest(method, base+path, bytes.NewReader(runnerRawJSON(body)))
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-KLM-Native", "1")
		r.Header.Set("X-KLM-Read-Contract", "3")
		r.Header.Set("X-KLM-Runner-ID", owner)
		response, err := client.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		data, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
		if err != nil || response.StatusCode >= 400 {
			record("http_failure", map[string]any{"path": path, "status": response.StatusCode, "body": string(data)})
			t.Fatalf("%s: %d %s %v", path, response.StatusCode, data, err)
		}
		if target != nil {
			if err := json.Unmarshal(data, target); err != nil {
				t.Fatal(err)
			}
		}
	}
	var connection struct {
		ID string `json:"integration_id"`
	}
	request("POST", "/api/runner/connections", map[string]string{"operation_id": "slice0-connect", "key": "slice0-owned-integration"}, &connection)
	owner = connection.ID
	// A later production role can require a larger application tool manifest.
	// The runner-only cold workload can explicitly reuse an earlier immutable
	// role capture rather than claiming that its minimal manifest implements Teams.
	if source := os.Getenv("KLM_TEAM_SLICE0_ROLE_CAPTURE"); source != "" {
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		var captured []struct {
			Kind  string          `json:"kind"`
			Value json.RawMessage `json:"value"`
		}
		if err := json.Unmarshal(data, &captured); err != nil {
			t.Fatal(err)
		}
		promptDir := filepath.Join(root, "captured-role-fixture")
		if err := os.Mkdir(promptDir, 0700); err != nil {
			t.Fatal(err)
		}
		for _, entry := range captured {
			name := ""
			if entry.Kind == "member_capture" {
				name = "team-member.md"
			}
			if entry.Kind == "embedded_leader_capture" {
				name = "team-leader.md"
			}
			if name == "" {
				continue
			}
			var operation runnerOperation
			var preparation runnerPreparation
			if err := json.Unmarshal(entry.Value, &operation); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(operation.Result, &preparation); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(promptDir, name), []byte(preparation.Capture.Instructions), 0600); err != nil {
				t.Fatal(err)
			}
		}
		t.Setenv("WERBERT_PROMPTS_DIR", promptDir)
		record("immutable_role_fixture", map[string]string{"source": source})
	}
	var foremanCommand *exec.Cmd
	var foremanInput io.WriteCloser
	var foremanOutput chan map[string]any
	startForeman := func(resume bool) map[string]any {
		t.Helper()
		probe := os.Getenv("KLM_TEAM_SLICE0_FOREMAN_PROBE")
		if !filepath.IsAbs(probe) {
			t.Fatal("supply the separately compiled owned Foreman process probe")
		}
		foremanCommand = exec.Command(probe, "-test.run=^TestForemanTeamSlice0Process$", "-test.timeout=180s")
		foremanCommand.Env = append(os.Environ(),
			"KLM_TEAM_SLICE0_FOREMAN_CHILD=1", "KLM_TEAM_SLICE0_ENDPOINT="+base,
			"KLM_TEAM_SLICE0_OWNER="+owner, fmt.Sprintf("KLM_TEAM_SLICE0_FOREMAN_RESUME=%d", map[bool]int{true: 1, false: 0}[resume]),
		)
		var err error
		foremanInput, err = foremanCommand.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		stdout, err := foremanCommand.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		foremanCommand.Stderr = os.Stderr
		if err := foremanCommand.Start(); err != nil {
			t.Fatal(err)
		}
		foremanOutput = make(chan map[string]any, 4)
		output := foremanOutput
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				var message map[string]any
				if json.Unmarshal(scanner.Bytes(), &message) == nil {
					output <- message
				}
			}
		}()
		select {
		case ready := <-foremanOutput:
			record("owned_foreman_started", map[string]any{"pid": foremanCommand.Process.Pid, "resume": resume})
			return ready
		case <-time.After(20 * time.Second):
			_ = foremanCommand.Process.Kill()
			t.Fatal("owned Foreman process readiness deadline")
			return nil
		}
	}
	stopForeman := func() {
		t.Helper()
		if foremanCommand == nil {
			return
		}
		_ = foremanInput.Close()
		err := foremanCommand.Wait()
		record("owned_foreman_stopped", map[string]any{"pid": foremanCommand.Process.Pid, "clean": err == nil})
		foremanCommand = nil
		if err != nil {
			t.Fatal(err)
		}
	}
	roles := startForeman(false)
	defer stopForeman()
	manifest := []map[string]any{{"name": "retain_result", "description": "Retain this dispatch result and the exact file content.", "inputSchema": map[string]any{
		"type": "object", "properties": map[string]any{"phase": map[string]any{"type": "string"}, "content": map[string]any{"type": "string"}},
		"required": []string{"phase", "content"}, "additionalProperties": false,
	}}}
	role := roles["member_role"].(string)
	prepare := map[string]any{
		"operation_id": "slice0-member-capture", "project_id": "slice0-project", "instance_key": "slice0-instance",
		"workspace": map[string]string{"path": worktree, "base": "HEAD"}, "recovery_policy": "resume_drained",
		"agent": "slice0-member", "source_uid": document.SourceUID, "agent_revision": revision(agentBytes),
		"capture": runnerCapture{Instructions: role, Manifest: manifest},
	}
	var prepared runnerOperation
	request("POST", "/api/runner/captures", prepare, &prepared)
	record("member_capture", prepared)
	leaderRole := roles["leader_role"].(string)
	leaderManifest := []map[string]any{{
		"name": "member_send", "description": "Dispatch initial work or a correction to this instance's one saved member; returns promptly.",
		"inputSchema": map[string]any{
			"type": "object", "properties": map[string]any{"task": map[string]any{"type": "string"}},
			"required": []string{"task"}, "additionalProperties": false,
		},
	}}
	var leaderPrepared runnerOperation
	request("POST", "/api/runner/captures", map[string]any{
		"operation_id": "slice0-leader-capture", "project_id": "slice0-project", "instance_key": "slice0-instance",
		"workspace": map[string]string{"path": worktree, "base": "HEAD"}, "recovery_policy": "resume_drained",
		"capture": runnerCapture{
			Harness: "codex", Model: "gpt-6-astra", Effort: "low", Instructions: leaderRole, Manifest: leaderManifest,
		},
	}, &leaderPrepared)
	record("embedded_leader_capture", leaderPrepared)
	// Mutating the source after preparation must not change this instance.
	document.Prompt = "MUTATED SOURCE MUST NEVER EXECUTE"
	changedBytes, _ := toml.Marshal(document)
	if err := os.WriteFile(agentPath, changedBytes, 0600); err != nil {
		t.Fatal(err)
	}
	var leaderAdmitted struct {
		Result runnerExecution `json:"result"`
	}
	request("POST", "/api/runner/work", map[string]any{
		"operation_id": "slice0-leader-admit", "project_id": "slice0-project", "key": "slice0-leader",
		"preparation_id": leaderPrepared.ID,
		"prompt":         "Call member_send once directing your member to write slice0-retained.txt in the current workspace with exactly INITIAL on its own line, using owned shell/job tools, read it back, call retain_result with phase initial and actual content, then end its turn. No other files or delegation. End your own turn immediately after dispatch; do not poll.",
	}, &leaderAdmitted)
	leaderID := leaderAdmitted.Result.ID
	workID := ""
	record("embedded_leader_admitted", leaderAdmitted.Result)
	cursor := "0"
	resultPhases := map[string]int{}
	processCalls := func() {
		var events struct {
			Events []runnerEvent `json:"events"`
			Next   string        `json:"next_after"`
		}
		request("GET", "/api/runner/events?after="+cursor, nil, &events)
		for _, event := range events.Events {
			if event.Type != "tool_call" {
				continue
			}
			var call runnerToolCall
			if err := json.Unmarshal(event.Data, &call); err != nil {
				t.Fatal(err)
			}
			if call.Name == "member_send" {
				var args struct {
					Task string `json:"task"`
				}
				if err := json.Unmarshal(call.Arguments, &args); err != nil || call.ExecutionID != leaderID {
					t.Fatal("leader dispatch ownership/arguments", err)
				}
				if workID == "" {
					var admitted struct {
						Result runnerExecution `json:"result"`
					}
					request("POST", "/api/runner/work", map[string]any{
						"operation_id": "slice0-member-admit", "project_id": "slice0-project", "key": "slice0-member",
						"preparation_id": prepared.ID, "prompt": args.Task,
					}, &admitted)
					workID = admitted.Result.ID
					record("member_admitted", admitted.Result)
				} else {
					request("POST", "/api/runner/work/"+workID+"/inputs", runnerInput{
						OperationID: "dispatch-" + call.ID, SourceKind: "application_event", SourceID: call.ID,
						Mode: "steer", Text: args.Task,
					}, nil)
					record("leader_directed_correction", call)
				}
				request("POST", "/api/runner/tool-calls/"+call.ID+"/reply", map[string]any{
					"operation_id": "reply-" + call.ID, "result": map[string]any{"dispatched": true, "work_id": workID},
				}, nil)
				continue
			}
			if call.Name != "retain_result" {
				continue
			}
			request("POST", "/api/runner/tool-calls/"+call.ID+"/reply", map[string]any{
				"operation_id": "retain-" + call.ID, "result": map[string]any{"retained": true, "result": json.RawMessage(call.Arguments)},
			}, nil)
			var result struct {
				Phase string `json:"phase"`
			}
			if err := json.Unmarshal(call.Arguments, &result); err != nil {
				t.Fatal(err)
			}
			resultPhases[result.Phase]++
			record("retained_member_result", call)
		}
		cursor = events.Next
	}
	previousTurn := ""
	originalNativeID := ""
	waitSettled := func(label string) runnerExecution {
		t.Helper()
		deadline := time.Now().Add(80 * time.Second)
		for time.Now().Before(deadline) {
			processCalls()
			if workID == "" {
				var leader struct {
					Execution runnerExecution `json:"execution"`
				}
				request("GET", "/api/runner/work/"+leaderID, nil, &leader)
				if leader.Execution.State == "interrupted" {
					record("leader_native_blocker", leader.Execution)
					var history any
					request("GET", "/api/runner/work/"+leaderID+"/history", nil, &history)
					record("leader_native_history", history)
					t.Fatal(leader.Execution.Error)
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}
			var current struct {
				Execution runnerExecution `json:"execution"`
				Native    struct {
					ID string `json:"id"`
				} `json:"native_context"`
			}
			request("GET", "/api/runner/work/"+workID, nil, &current)
			if current.Execution.State == "native_completed" && current.Execution.TurnID != previousTurn {
				record(label, current.Execution)
				record(label+"_native_context", current.Native)
				if originalNativeID == "" {
					originalNativeID = current.Native.ID
				}
				if current.Native.ID == "" || originalNativeID != current.Native.ID {
					t.Fatal("original native context was replaced")
				}
				previousTurn = current.Execution.TurnID
				return current.Execution
			}
			if current.Execution.State == "interrupted" {
				record(label+"_blocked", current.Execution)
				var history any
				request("GET", "/api/runner/work/"+workID+"/history", nil, &history)
				record("native_history", history)
				t.Fatalf("%s: %s (%s)", label, current.Execution.Error, current.Execution.RecoveryBlocked)
			}
			time.Sleep(100 * time.Millisecond)
		}
		t.Fatal(label + " provider deadline")
		return runnerExecution{}
	}
	initial := waitSettled("initial_work_settled")
	if strings.Contains(initial.Applied.Instructions, "MUTATED SOURCE") || resultPhases["initial"] != 1 {
		t.Fatal("immutable saved-member work/result was not established")
	}
	correction := runnerInput{
		OperationID: "slice0-correction", SourceKind: "application_event", SourceID: "leader-correction-1", Mode: "steer",
		Text: "The member retained its initial result. Use member_send to correct the SAME saved member: inspect slice0-retained.txt, preserve INITIAL and append exactly CORRECTED on a second line only if absent, read it back, retain_result with phase corrected and actual content, then end its turn. End your own turn after dispatch.",
	}
	// Persist the original continuation in Werbert, then replace only the owned
	// W process. Its production outbox resumes this exact operation automatically.
	if err := json.NewEncoder(foremanInput).Encode(map[string]any{
		"operation_id": correction.OperationID, "path": "/api/runner/work/" + leaderID + "/inputs", "body": correction,
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case ack := <-foremanOutput:
		record("foreman_continuation_durably_queued", ack)
	case <-time.After(10 * time.Second):
		t.Fatal("Foreman outbox durable acknowledgement deadline")
	}
	stopForeman()
	startForeman(true)
	corrected := waitSettled("same_member_correction_settled")
	if resultPhases["corrected"] != 1 {
		t.Fatal("leader correction did not retain its result")
	}
	var correctionReceipt runnerInput
	request("GET", "/api/runner/work/"+leaderID+"/inputs?source_kind=application_event&source_id=leader-correction-1", nil, &correctionReceipt)
	record("original_foreman_continuation_inclusion", correctionReceipt)
	stopForeman()
	if initial.SessionID != corrected.SessionID || initial.Generation != corrected.Generation {
		t.Fatal("correction replaced the original participant/session")
	}
	// A live native continuation is interrupted by a separate owned K process
	// restart; the existing result/file must survive without replay.
	request("POST", "/api/runner/work/"+workID+"/inputs", runnerInput{
		OperationID: "slice0-restart-work", SourceKind: "application_event", SourceID: "leader-restart-check", Mode: "steer",
		Text: "Read slice0-retained.txt with an owned shell command. Then reason briefly about whether its original two lines remain intact before calling retain_result with phase recovered and the actual content. Preserve the file exactly. If a runner recovery event arrives, inspect history and the existing file first; do not replay prior edits.",
	}, nil)
	time.Sleep(1500 * time.Millisecond)
	var interrupted struct {
		Execution runnerExecution `json:"execution"`
	}
	request("GET", "/api/runner/work/"+workID, nil, &interrupted)
	if interrupted.Execution.State != "running" {
		t.Fatal("the owned restart did not interrupt active native work")
	}
	record("active_original_work_before_klm_restart", interrupted.Execution)
	if os.Getenv("KLM_TEAM_SLICE0_CRASH") == "1" {
		deadline := time.Now().Add(10 * time.Second)
		for len(interrupted.Execution.ProcessDomains) == 0 || !interrupted.Execution.ProcessDomains[0].Contained {
			if time.Now().After(deadline) {
				t.Fatal("durable native containment admission deadline")
			}
			time.Sleep(50 * time.Millisecond)
			request("GET", "/api/runner/work/"+workID, nil, &interrupted)
		}
		pid := command.Process.Pid
		if err := command.Process.Kill(); err != nil {
			t.Fatal(err)
		}
		_ = command.Wait()
		_ = stdin.Close()
		command = nil
		record("owned_engine_abruptly_killed", map[string]any{"pid": pid, "domains": interrupted.Execution.ProcessDomains})
	} else {
		stop()
	}
	start()
	recovered := waitSettled("original_session_after_klm_restart")
	if resultPhases["initial"] != 1 || resultPhases["corrected"] != 1 || resultPhases["recovered"] != 1 {
		t.Fatalf("missing or replayed retained results: %v", resultPhases)
	}
	var recoveryInput runnerInput
	request("GET", "/api/runner/work/"+workID+"/inputs?source_kind=application_event&source_id=recovery/"+interrupted.Execution.TurnID, nil, &recoveryInput)
	if recoveryInput.State != "included" {
		t.Fatal("automatic original-turn recovery input was not included")
	}
	record("original_recovery_input_included", recoveryInput)
	if recovered.SessionID != initial.SessionID || recovered.Generation != initial.Generation {
		t.Fatal("restart replaced the original participant/session")
	}
	content, err := os.ReadFile(filepath.Join(worktree, "slice0-retained.txt"))
	if err != nil || strings.ReplaceAll(strings.TrimSpace(string(content)), "\r\n", "\n") != "INITIAL\nCORRECTED" {
		t.Fatalf("retained artifact changed or replayed: %q %v", content, err)
	}
	record("retained_artifact", map[string]any{"path": filepath.Join(worktree, "slice0-retained.txt"), "content": string(content)})
	var history any
	request("GET", "/api/runner/work/"+workID+"/history", nil, &history)
	record("native_history", history)
	stop()
}

func mustWorkingDirectory(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return cwd
}
