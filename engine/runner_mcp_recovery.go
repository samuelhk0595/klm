package main

import (
	"context"
	"errors"
	"strings"
)

// Library effects use the existing call ledger, but only KLM can settle these
// internal records. No native argument/credential environment is projected here.
func (p *adapter) retainRunnerMCP(callID, tool, state string, result any) error {
	if p.turn.runner == nil || p.turn.runner.RecoveryPolicy != "resume_drained" {
		return nil
	}
	a := p.app
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.persistSQLLocked(context.Background(), func(st *sqliteSessionTx) error {
		execution, err := st.runnerExecution(p.turn.runner.ExecutionID)
		if err != nil {
			return err
		}
		id := revision([]byte(execution.ID + "/" + p.turn.runner.TurnID + "/mcp/" + callID))
		if state == "mcp_pending" {
			if execution.CancelRequested || execution.State != "running" || p.turn.ctx.Err() != nil {
				return errors.New("Library effect admission is sealed.")
			}
			if err := st.runnerGroupOpen(execution.IntegrationID, execution.ProjectID, execution.InstanceKey); err != nil {
				return err
			}
		}
		parts := strings.SplitN(tool, "/", 3)
		if len(parts) != 3 {
			return errors.New("Invalid retained Library tool identity.")
		}
		call := runnerToolCall{
			ID: id, ExecutionID: execution.ID, TurnID: p.turn.runner.TurnID, NativeCallID: "mcp/" + callID,
			Name: "__mcp/" + tool, State: state, Arguments: runnerRawJSON(map[string]string{"connection_id": parts[0], "operation": parts[1], "tool": parts[2]}),
		}
		if result != nil {
			call.Reply = runnerRawJSON(result)
		}
		data, err := runnerJSON(call)
		if err != nil {
			return err
		}
		_, err = st.tx.ExecContext(st.ctx, `INSERT INTO runner_calls VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET body=excluded.body`,
			call.ID, call.ExecutionID, call.TurnID, call.NativeCallID, revision([]byte(tool)), data)
		if err != nil {
			return err
		}
		return st.runnerEvent(execution, "mcp_effect", call)
	})
}
