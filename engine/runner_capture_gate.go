package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// This fixed lifecycle request lets the owning application re-read its definition
// immediately before a complete turn. It is never offered as a model tool.
func (s *runnerService) captureNextTurnLocked(execution runnerExecution) error {
	a := s.app
	call := runnerToolCall{ID: newID(), ExecutionID: execution.ID, TurnID: newID(), NativeCallID: "capture", Name: "__capture_next_turn", Arguments: json.RawMessage(`{}`), State: "pending"}
	err := a.writeSQLLocked(a.ctx, func(st *sqliteSessionTx) error {
		execution.State = "capture_pending"
		if err := st.putRunnerExecution(execution); err != nil {
			return err
		}
		body, err := runnerJSON(call)
		if err != nil {
			return err
		}
		if _, err := st.tx.ExecContext(st.ctx, `INSERT INTO runner_calls VALUES(?,?,?,?,?,?)`, call.ID, call.ExecutionID, call.TurnID, call.NativeCallID, "capture-v1", body); err != nil {
			return err
		}
		return st.runnerEvent(execution, "capture_request", call)
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	wake := make(chan struct{}, 1)
	s.callWake[call.ID] = wake
	s.mu.Unlock()
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
		defer cancel()
		var desired runnerCapture
		var turnContext string
		var failure error
		for {
			var body []byte
			if err := a.store.readers.QueryRowContext(ctx, `SELECT body FROM runner_calls WHERE id=?`, call.ID).Scan(&body); err != nil {
				failure = err
				break
			}
			if err := json.Unmarshal(body, &call); err != nil {
				failure = err
				break
			}
			if call.State == "replied" {
				var reply struct {
					Capture runnerCapture `json:"capture"`
					Error   string        `json:"error"`
					Context string        `json:"context"`
				}
				if err := json.Unmarshal(call.Reply, &reply); err != nil {
					failure = err
					break
				}
				if reply.Error != "" {
					failure = errors.New(reply.Error)
					break
				}
				if len(reply.Context) > 128<<10 {
					failure = errors.New("Application turn context exceeds 128 KiB.")
					break
				}
				turnContext = reply.Context
				project, err := a.store.readProjectMeta(ctx, execution.ProjectID)
				if err != nil {
					failure = err
					break
				}
				desired, failure = a.validateRunnerCapture(ctx, reply.Capture, project)
				break
			}
			select {
			case <-wake:
			case <-ctx.Done():
				failure = ctx.Err()
			}
			if failure != nil {
				break
			}
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		_ = a.persistSQLLocked(context.Background(), func(st *sqliteSessionTx) error {
			current, err := st.runnerExecution(execution.ID)
			if err != nil {
				return err
			}
			if current.State != "capture_pending" {
				return nil
			}
			if failure != nil {
				current.State = "blocked"
				current.Error = failure.Error()
				return st.putRunnerExecution(current)
			}
			current.Desired, current.State, current.Error = desired, "idle", ""
			s.prepared[current.ID] = revision(runnerCaptureJSON(desired))
			s.preparedContext[current.ID] = turnContext
			return st.putRunnerExecution(current)
		})
		if failure == nil {
			_ = s.scheduleLocked(execution.ID)
		}
	}()
	return nil
}
