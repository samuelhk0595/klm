package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type runnerJob struct {
	ProjectGit        *ProjectGitIdentitySnapshot `json:"project_git,omitempty"`
	Command           string                      `json:"command"`
	Directory         string                      `json:"directory"`
	Shell             string                      `json:"shell"`
	EnvironmentDigest string                      `json:"environment_digest"`
	Error             string                      `json:"error,omitempty"`
	ContentID         string                      `json:"content_id,omitempty"`
	Digest            string                      `json:"digest,omitempty"`
	Bytes             int64                       `json:"bytes"`
	ID                string                      `json:"id"`
	ExecutionID       string                      `json:"execution_id"`
	TurnID            string                      `json:"turn_id"`
	State             string                      `json:"state"`
	PID               int                         `json:"pid,omitempty"`
	Output            string                      `json:"output"`
	ExitCode          *int                        `json:"exit_code,omitempty"`
	Domain            *runnerProcessDomain        `json:"process_domain,omitempty"`
}
type runnerLiveJob struct {
	cancelRequested bool
	mu              sync.Mutex
	record          runnerJob
	owner           *ownedProcess
	done            chan struct{}
	output          cappedBuffer
	file            *os.File
}

func (j *runnerLiveJob) Write(data []byte) (int, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	_, _ = j.output.Write(data)
	n, err := j.file.Write(data)
	j.record.Bytes += int64(n)
	return n, err
}

func runnerJobTools() []map[string]any {
	stringField := map[string]any{"type": "string"}
	tools := []map[string]any{}
	for _, name := range []string{"shell_start", "job_read", "job_wait", "job_cancel"} {
		properties := map[string]any{"job_id": stringField, "wait_ms": map[string]any{"type": "integer", "minimum": 0, "maximum": 5000}}
		required := []string{"job_id"}
		if name == "shell_start" {
			properties = map[string]any{"command": stringField, "wait_ms": map[string]any{"type": "integer", "minimum": 0, "maximum": 5000}}
			required = []string{"command"}
		}
		tools = append(tools, map[string]any{"name": name, "description": "Owned shell job operation. Foreground and job waits are capped at 5000 ms; a running job continues in the same process. Never poll for completion; completion is delivered as an event.", "inputSchema": map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}})
	}
	return tools
}

func (s *runnerService) saveJob(job runnerJob) error {
	return s.saveJobState(job, false)
}

func (s *runnerService) saveJobState(job runnerJob, admission bool) error {
	a := s.app
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.persistSQLLocked(context.Background(), func(st *sqliteSessionTx) error {
		execution, err := st.runnerExecution(job.ExecutionID)
		if err != nil {
			return err
		}
		if admission {
			if execution.CancelRequested || execution.State != "running" || execution.TurnID != job.TurnID {
				return errors.New("Owned shell process admission is sealed.")
			}
			if err := st.runnerGroupOpen(execution.IntegrationID, execution.ProjectID, execution.InstanceKey); err != nil {
				return err
			}
		}
		body, err := runnerJSON(job)
		if err != nil {
			return err
		}
		if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_jobs SET state=?,body=? WHERE id=?`, job.State, body, job.ID); err != nil {
			return err
		}
		return st.runnerEvent(execution, "job", job)
	})
}

func (s *runnerService) jobCall(ctx context.Context, p *adapter, name, callID string, raw json.RawMessage) (any, error) {
	started := time.Now()
	var args struct {
		Command string `json:"command"`
		JobID   string `json:"job_id"`
		WaitMS  *int   `json:"wait_ms"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, err
	}
	wait := 5000
	if args.WaitMS != nil {
		wait = max(0, min(5000, *args.WaitMS))
	}
	// Reserve framing time inside the external role's five-second ceiling.
	wait = min(wait, 4800)
	if name == "shell_start" {
		if p.turn.projectGit == nil {
			return nil, errors.New("Shell admission has no owning turn Git identity capture.")
		}
		if args.Command == "" || len(args.Command) > 64<<10 || callID == "" {
			return nil, errors.New("A bounded command and native call identity are required.")
		}
		release, err := s.app.acquireHarnessConsumer()
		if err != nil {
			return nil, err
		}
		retained := false
		var command *exec.Cmd
		var owner *ownedProcess
		settled := func() bool {
			return command == nil || command.Process == nil || owner != nil && owner.Drained()
		}
		defer func() {
			if !retained && settled() {
				release()
			}
		}()
		node := ""
		if b, ok := s.app.binarySnapshot()[p.harness]; ok {
			node = b.node
		}
		environment := skillExecutionEnv(runtimeEnvironment(os.Environ(), node))
		shell, argv := "sh", []string{"-c", args.Command}
		if runtime.GOOS == "windows" {
			shell = "powershell.exe"
			argv = []string{"-NoLogo", "-NoProfile", "-NonInteractive", "-Command", args.Command}
		}
		// Only captured references enter the digest. Projected authority, payloads
		// and ambient credentials never become durable job metadata or hashes.
		intent := runnerJob{
			Command: args.Command, Directory: p.cwd, Shell: shell, ProjectGit: p.turn.projectGit,
			EnvironmentDigest: revision(runnerRawJSON(p.turn.projectGit)),
		}
		args.JobID = revision([]byte(p.turn.runner.ExecutionID + "/" + p.turn.runner.TurnID + "/" + callID))
		digest := revision(raw)
		fresh := false
		a := s.app
		a.mu.Lock()
		err = a.writeSQLLocked(ctx, func(st *sqliteSessionTx) error {
			execution, err := st.runnerExecution(p.turn.runner.ExecutionID)
			if err != nil {
				return err
			}
			if execution.CancelRequested || execution.State != "running" || execution.TurnID != p.turn.runner.TurnID {
				return errors.New("Shell admission belongs to a sealed turn.")
			}
			if err := st.runnerGroupOpen(execution.IntegrationID, execution.ProjectID, execution.InstanceKey); err != nil {
				return err
			}
			var previous string
			err = st.tx.QueryRowContext(ctx, `SELECT digest FROM runner_jobs WHERE id=?`, args.JobID).Scan(&previous)
			if err == nil {
				if previous != digest {
					return controlError(409, "Shell call arguments changed.")
				}
				return nil
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			job := intent
			job.ID, job.ExecutionID, job.TurnID, job.State = args.JobID, p.turn.runner.ExecutionID, p.turn.runner.TurnID, "starting"
			body, err := runnerJSON(job)
			if err != nil {
				return err
			}
			_, err = st.tx.ExecContext(ctx, `INSERT INTO runner_jobs VALUES(?,?,?,?,?,?,?)`, job.ID, job.ExecutionID, job.TurnID, callID, digest, job.State, body)
			fresh = err == nil
			return err
		})
		a.mu.Unlock()
		if err != nil {
			return nil, err
		}
		if fresh {
			cmd := exec.Command(shell, argv...)
			command = cmd
			cmd.Dir = p.cwd
			var domain *runnerProcessDomain
			environment, cleanup, err := a.projectGitEnvironment(environment, p.turn.projectGit)
			if err != nil {
				intent.ID, intent.ExecutionID, intent.TurnID = args.JobID, p.turn.runner.ExecutionID, p.turn.runner.TurnID
				intent.State, intent.Error = "failed", err.Error()
				_ = s.saveJob(intent)
				return nil, err
			}
			cmd.Env = environment
			projectGitRetained := false
			defer func() {
				if !projectGitRetained && settled() {
					cleanup()
				}
			}()
			if p.turn.runner.RecoveryPolicy == "resume_drained" {
				owner, domain, err = prepareRunnerOwnedProcess(cmd, a.dir)
			} else {
				owner, err = prepareOwnedProcess(cmd)
			}
			if err != nil {
				return nil, err
			}
			live := &runnerLiveJob{record: intent, owner: owner, done: make(chan struct{})}
			live.record.ID, live.record.ExecutionID, live.record.TurnID, live.record.State = args.JobID, p.turn.runner.ExecutionID, p.turn.runner.TurnID, "running"
			outputDir := filepath.Join(s.app.dir, "runner-output")
			if err := os.MkdirAll(outputDir, 0700); err != nil {
				owner.Close()
				return nil, err
			}
			output, err := os.OpenFile(filepath.Join(outputDir, args.JobID+".log"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if err != nil {
				owner.Close()
				return nil, err
			}
			live.file = output
			live.record.ContentID = args.JobID
			live.record.Domain = domain
			if domain != nil {
				pending := live.record
				pending.State = "starting"
				pending.Domain.PendingStarts = 1
				if err := s.saveJob(pending); err != nil {
					output.Close()
					owner.Close()
					return nil, err
				}
				owner.beforeResume = func() error {
					live.mu.Lock()
					live.record.Domain.Contained = true
					live.record.Domain.PendingStarts = 0
					record := live.record
					live.mu.Unlock()
					return s.saveJobState(record, true)
				}
			}
			cmd.Stdout, cmd.Stderr = live, live
			// A cancelled call cannot start an unowned job after Stop's admission barrier.
			s.mu.Lock()
			if err := p.turn.ctx.Err(); err != nil {
				s.mu.Unlock()
				output.Close()
				owner.Close()
				s.failedJobStart(live, err)
				return nil, err
			}
			if err := cmd.Start(); err != nil {
				s.mu.Unlock()
				output.Close()
				owner.Close()
				s.failedJobStart(live, err)
				return nil, err
			}
			if err := owner.Attach(); err != nil {
				owner.Stop()
				cmd.Wait()
				s.mu.Unlock()
				output.Close()
				owner.Close()
				s.failedJobStart(live, err)
				return nil, err
			}
			live.record.PID = cmd.Process.Pid
			s.jobs[args.JobID] = live
			s.mu.Unlock()
			live.mu.Lock()
			initial := live.record
			live.mu.Unlock()
			if err := s.saveJob(initial); err != nil {
				owner.Stop()
				cmd.Wait()
				owner.Close()
				output.Close()
				s.failedJobStart(live, err)
				close(live.done)
				return nil, err
			}
			s.app.wg.Add(1)
			retained = true
			projectGitRetained = true
			go func() {
				defer s.app.wg.Done()
				defer func() {
					if owner.Drained() {
						cleanup()
						release()
					}
				}()
				err := cmd.Wait()
				live.mu.Lock()
				code := cmd.ProcessState.ExitCode()
				live.record.State = "completed"
				if err != nil {
					live.record.State = "failed"
				}
				if live.cancelRequested {
					live.record.State = "cancelled"
				}
				live.record.ExitCode = &code
				live.record.Output = live.output.String()
				if syncErr := live.file.Sync(); syncErr != nil {
					live.record.State = "failed"
				}
				_ = live.file.Close()
				if file, err := os.Open(filepath.Join(outputDir, args.JobID+".log")); err == nil {
					hash := sha256.New()
					if _, err := io.Copy(hash, file); err == nil {
						live.record.Digest = hex.EncodeToString(hash.Sum(nil))
					}
					file.Close()
				}
				record := live.record
				live.mu.Unlock()
				if closeErr := owner.Close(); closeErr != nil || !owner.Drained() {
					live.mu.Lock()
					live.record.State = "unknown"
					live.record.Error = "Owned job process drainage is unconfirmed."
					record = live.record
					live.mu.Unlock()
				}
				s.jobCompleted(record)
				close(live.done)
			}()
		}
	}
	s.mu.Lock()
	live := s.jobs[args.JobID]
	s.mu.Unlock()
	if live == nil {
		var body []byte
		var state string
		if err := s.app.store.readers.QueryRowContext(ctx, `SELECT body,state FROM runner_jobs WHERE id=? AND execution_id=?`, args.JobID, p.turn.runner.ExecutionID).Scan(&body, &state); err != nil {
			return nil, err
		}
		var record runnerJob
		if err := json.Unmarshal(body, &record); err != nil {
			return nil, err
		}
		record.State = state
		return record, nil // An uncertain start is never permission to spawn again.
	}
	if live.record.ExecutionID != p.turn.runner.ExecutionID {
		return nil, errors.New("Job belongs to another execution.")
	}
	if name == "job_cancel" {
		live.mu.Lock()
		live.cancelRequested = true
		live.mu.Unlock()
		if err := live.owner.Stop(); err != nil {
			return nil, err
		}
	}
	if name == "shell_start" || name == "job_wait" || name == "job_cancel" {
		remaining := time.Duration(wait)*time.Millisecond - time.Since(started)
		if remaining > 0 {
			timer := time.NewTimer(remaining)
			defer timer.Stop()
			select {
			case <-live.done:
			case <-timer.C:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	live.mu.Lock()
	defer live.mu.Unlock()
	result := live.record
	result.Output = live.output.String()
	return result, nil
}

func (s *runnerService) failedJobStart(live *runnerLiveJob, cause error) {
	live.mu.Lock()
	live.record.State = "unknown"
	live.record.Error = cause.Error()
	if live.owner.Drained() {
		live.record.State = "failed"
	}
	record := live.record
	live.mu.Unlock()
	_ = s.saveJob(record)
}

func (s *runnerService) jobCompleted(job runnerJob) {
	a := s.app
	a.mu.Lock()
	defer a.mu.Unlock()
	var execution runnerExecution
	var continuation bool
	err := a.persistSQLLocked(context.Background(), func(st *sqliteSessionTx) error {
		var err error
		execution, err = st.runnerExecution(job.ExecutionID)
		if err != nil {
			return err
		}
		text, err := runnerJSON(job)
		if err != nil {
			return err
		}
		if _, err := st.tx.ExecContext(st.ctx, `UPDATE runner_jobs SET state=?,body=? WHERE id=?`,
			job.State, text, job.ID); err != nil {
			return err
		}
		if err := st.runnerEvent(execution, "job", job); err != nil {
			return err
		}
		switch execution.State {
		case "cancelling", "cancelled", "stopped", "interrupted":
			return nil
		}
		if job.State == "unknown" {
			execution.State, execution.Error = "interrupted", job.Error
			return st.putRunnerExecution(execution)
		}
		if a.ctx.Err() != nil {
			return nil
		}
		input := runnerInput{
			ID: "job/" + job.ID, SourceID: job.ID, SourceKind: "job_result", Mode: "steer", Text: string(text),
		}
		if err := s.enqueueTx(st, execution, input); err != nil {
			return err
		}
		continuation = true
		// Retained completion and job cessation commit together. A duplicate
		// callback whose result was already included must not reopen settled work.
		if execution.State == "native_completed" {
			var retained bool
			if err := st.tx.QueryRowContext(st.ctx,
				`SELECT EXISTS(SELECT 1 FROM runner_inputs WHERE execution_id=? AND state='retained'
 AND json_extract(body,'$.mode')<>'passive')`, execution.ID).Scan(&retained); err != nil {
				return err
			}
			if retained {
				execution.State = "idle"
				return st.putRunnerExecution(execution)
			}
		}
		return nil
	})
	if err != nil {
		return
	}
	if job.State == "unknown" {
		if t := a.runs[execution.SessionID]; t != nil {
			t.stopErr = errors.New(job.Error)
			t.cancel()
		}
		return
	}
	if !continuation {
		return
	}
	a.wakeSteeringLocked(execution.SessionID)
	_ = s.scheduleLocked(execution.ID)
}
