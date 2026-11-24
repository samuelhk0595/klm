package main

import (
	"database/sql"
	"encoding/json"
	"strings"
)

func (p *adapter) runnerNativeInput(text string, boundary bool) error {
	if p.turn.runner == nil {
		return nil
	}
	a := p.app
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.writeSQLLocked(p.turn.ctx, func(st *sqliteSessionTx) error {
		rows, err := st.tx.QueryContext(st.ctx, `SELECT id,body FROM runner_inputs WHERE execution_id=? AND state='claimed'`, p.turn.runner.ExecutionID)
		if err != nil {
			return err
		}
		ids := []string{}
		for rows.Next() {
			var id string
			var body []byte
			if err := rows.Scan(&id, &body); err != nil {
				rows.Close()
				return err
			}
			var input runnerInput
			if err := json.Unmarshal(body, &input); err != nil {
				rows.Close()
				return err
			}
			if input.TurnID == p.turn.runner.TurnID && (boundary || strings.Contains(text, "Input "+id+" (")) {
				ids = append(ids, id)
			}
		}
		if err := rows.Close(); err != nil {
			return err
		}
		for _, id := range ids {
			if err := st.runnerInputState(id, "included"); err != nil {
				return err
			}
		}
		if len(ids) > 0 {
			execution, err := st.runnerExecution(p.turn.runner.ExecutionID)
			if err != nil {
				return err
			}
			return st.runnerEvent(execution, "input_included", map[string]any{"input_ids": ids, "turn_id": p.turn.runner.TurnID})
		}
		return nil
	})
}

func (st *sqliteSessionTx) runnerSteeringState(inputID, state, turnID string) error {
	var body []byte
	err := st.tx.QueryRowContext(st.ctx, `SELECT body FROM runner_inputs WHERE id=?`, inputID).Scan(&body)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	var input runnerInput
	if err := json.Unmarshal(body, &input); err != nil {
		return err
	}
	input.State, input.TurnID = state, turnID
	body, err = runnerJSON(input)
	if err != nil {
		return err
	}
	_, err = st.tx.ExecContext(st.ctx, `UPDATE runner_inputs SET state=?,body=? WHERE id=?`, state, body, inputID)
	return err
}
