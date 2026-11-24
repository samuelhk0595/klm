package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunnerColdOwnedJobProcess(t *testing.T) {
	if os.Getenv("KLM_RUNNER_COLD_JOB_CHILD") != "1" {
		t.Skip("owned cold-job subprocess")
	}
	dir := os.Getenv("KLM_RUNNER_COLD_JOB_DIR")
	file, err := os.OpenFile(filepath.Join(dir, "runner-output", "cold-job.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	cmd := exec.Command("powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "Write-Output 'RETAINED-BEFORE-CRASH'; Start-Sleep -Seconds 60; Write-Output 'AFTER-SLEEP'")
	cmd.Stdout, cmd.Stderr = file, file
	owner, domain, err := prepareRunnerOwnedProcess(cmd, dir)
	if err != nil {
		t.Fatal(err)
	}
	owner.beforeResume = func() error { domain.Contained = true; domain.PendingStarts = 0; return nil }
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if err := owner.Attach(); err != nil {
		t.Fatal(err)
	}
	defer func() { owner.Close(); cmd.Wait() }()
	deadline := time.Now().Add(10 * time.Second)
	for {
		data, _ := os.ReadFile(filepath.Join(dir, "runner-output", "cold-job.log"))
		if strings.Contains(string(data), "RETAINED-BEFORE-CRASH") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("owned job output deadline")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err := json.NewEncoder(os.Stdout).Encode(domain); err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, os.Stdin)
}

func TestRunnerColdOwnedJobRetainsOutput(t *testing.T) {
	a, execution := runnerRegressionApp(t, "worker")
	if err := os.Mkdir(filepath.Join(a.dir, "runner-output"), 0700); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(exe, "-test.run=^TestRunnerColdOwnedJobProcess$", "-test.timeout=30s")
	child.Env = append(os.Environ(), "KLM_RUNNER_COLD_JOB_CHILD=1", "KLM_RUNNER_COLD_JOB_DIR="+a.dir)
	stdin, err := child.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	stdout, err := child.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer child.Process.Kill()
	domains := make(chan runnerProcessDomain, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var domain runnerProcessDomain
			if json.Unmarshal(scanner.Bytes(), &domain) == nil {
				domains <- domain
				return
			}
		}
	}()
	var domain runnerProcessDomain
	select {
	case domain = <-domains:
	case <-time.After(15 * time.Second):
		t.Fatal("owned process identity deadline")
	}
	job := runnerJob{ID: "cold-job", ExecutionID: execution.ID, TurnID: execution.TurnID, State: "unknown", Domain: &domain}
	if _, err := a.store.writer.Exec(`INSERT INTO runner_jobs VALUES(?,?,?,?,?,?,?)`, job.ID, execution.ID, execution.TurnID, "cold-job-call", "digest", job.State, runnerRawJSON(job)); err != nil {
		t.Fatal(err)
	}
	if err := child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = child.Wait()
	if err := a.reconcileRunnerJobs(context.Background(), &execution); err != nil {
		t.Fatal(err)
	}
	if execution.RecoveryBlocked != "" {
		t.Fatal(execution.RecoveryBlocked, execution.Error)
	}
	var body []byte
	var state string
	if err := a.store.readers.QueryRow(`SELECT state,body FROM runner_jobs WHERE id=?`, job.ID).Scan(&state, &body); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &job); err != nil {
		t.Fatal(err)
	}
	if state != "interrupted" || job.ExitCode != nil || job.Digest == "" || job.Bytes == 0 {
		t.Fatalf("cold job was lost or fabricated success: %s %s", state, body)
	}
	data, err := os.ReadFile(filepath.Join(a.dir, "runner-output", "cold-job.log"))
	if err != nil || !strings.Contains(string(data), "RETAINED-BEFORE-CRASH") || strings.Contains(string(data), "AFTER-SLEEP") {
		t.Fatalf("retained output is invalid: %q %v", data, err)
	}
	// Reconciliation of the same durable intent is observation-only.
	if err := a.reconcileRunnerJobs(context.Background(), &execution); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(a.dir, "runner-output", "cold-job.log"))
	if string(after) != string(data) {
		t.Fatal("cold command was replayed")
	}
}
