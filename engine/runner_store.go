package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
)

const runnerSchema = `
CREATE TABLE IF NOT EXISTS runner_integrations(id TEXT PRIMARY KEY,client_id TEXT NOT NULL,external_key TEXT NOT NULL UNIQUE) STRICT;
CREATE TABLE IF NOT EXISTS runner_executions(id TEXT PRIMARY KEY,integration_id TEXT NOT NULL,project_id TEXT NOT NULL,kind TEXT NOT NULL,external_key TEXT NOT NULL,session_id TEXT NOT NULL UNIQUE,body BLOB NOT NULL,UNIQUE(integration_id,kind,external_key)) STRICT;
CREATE TABLE IF NOT EXISTS runner_operations(integration_id TEXT NOT NULL,id TEXT NOT NULL,digest TEXT NOT NULL,body BLOB NOT NULL,PRIMARY KEY(integration_id,id)) STRICT;
CREATE TABLE IF NOT EXISTS runner_inputs(id TEXT PRIMARY KEY,execution_id TEXT NOT NULL,source_kind TEXT NOT NULL,source_id TEXT NOT NULL,position INTEGER NOT NULL,state TEXT NOT NULL,body BLOB NOT NULL,UNIQUE(execution_id,source_kind,source_id)) STRICT;
CREATE INDEX IF NOT EXISTS runner_inputs_owner ON runner_inputs(execution_id,state,position);
CREATE TABLE IF NOT EXISTS runner_calls(id TEXT PRIMARY KEY,execution_id TEXT NOT NULL,turn_id TEXT NOT NULL,native_call_id TEXT NOT NULL,digest TEXT NOT NULL,body BLOB NOT NULL,UNIQUE(execution_id,turn_id,native_call_id)) STRICT;
CREATE TABLE IF NOT EXISTS runner_jobs(id TEXT PRIMARY KEY,execution_id TEXT NOT NULL,turn_id TEXT NOT NULL,native_call_id TEXT NOT NULL,digest TEXT NOT NULL,state TEXT NOT NULL,body BLOB NOT NULL,UNIQUE(execution_id,turn_id,native_call_id)) STRICT;
CREATE TABLE IF NOT EXISTS runner_events(seq INTEGER PRIMARY KEY AUTOINCREMENT,id TEXT NOT NULL UNIQUE,integration_id TEXT NOT NULL,execution_id TEXT NOT NULL,type TEXT NOT NULL,at TEXT NOT NULL,body BLOB NOT NULL) STRICT;
CREATE INDEX IF NOT EXISTS runner_events_owner ON runner_events(integration_id,seq);
CREATE TABLE IF NOT EXISTS runner_generations(execution_id TEXT NOT NULL,generation TEXT NOT NULL,session_id TEXT NOT NULL UNIQUE,PRIMARY KEY(execution_id,generation)) STRICT;
CREATE TABLE IF NOT EXISTS runner_requests(id TEXT PRIMARY KEY,execution_id TEXT NOT NULL,native_id TEXT NOT NULL,kind TEXT NOT NULL,state TEXT NOT NULL,body BLOB NOT NULL,reply BLOB) STRICT;
`

func runnerJSON(value any) ([]byte, error) {
	body, err := json.Marshal(value)
	if err == nil && len(body) > 2<<20 {
		return nil, errors.New("Runner record exceeds 2 MiB.")
	}
	return body, err
}

func (st *sqliteSessionTx) runnerExecution(id string) (runnerExecution, error) {
	var result runnerExecution
	var body []byte
	err := st.tx.QueryRowContext(st.ctx, `SELECT body FROM runner_executions WHERE id=?`, id).Scan(&body)
	if err == nil {
		err = json.Unmarshal(body, &result)
	}
	return result, err
}

func (s *sqliteStore) readRunnerExecution(ctx context.Context, id string) (runnerExecution, error) {
	var result runnerExecution
	err := s.readSnapshot(ctx, func(tx *sql.Tx) error {
		var err error
		result, err = (&sqliteSessionTx{ctx: ctx, tx: tx}).runnerExecution(id)
		return err
	})
	return result, err
}

func (st *sqliteSessionTx) putRunnerExecution(execution runnerExecution) error {
	execution.Revision = st.changes.Revision
	body, err := runnerJSON(execution)
	if err != nil {
		return err
	}
	_, err = st.tx.ExecContext(st.ctx, `INSERT INTO runner_executions VALUES(?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET session_id=excluded.session_id,body=excluded.body`,
		execution.ID, execution.IntegrationID, execution.ProjectID, execution.Kind,
		execution.Key, execution.SessionID, body)
	if err != nil {
		return err
	}
	if _, err := st.tx.ExecContext(st.ctx, `INSERT OR IGNORE INTO runner_generations VALUES(?,?,?)`, execution.ID, execution.Generation, execution.SessionID); err != nil {
		return err
	}
	return st.runnerEvent(execution, "execution", execution)
}

func (st *sqliteSessionTx) runnerEvent(execution runnerExecution, kind string, value any) error {
	st.changes.Invalidated["runner"] = true
	body, err := runnerJSON(value)
	if err != nil {
		return err
	}
	_, err = st.tx.ExecContext(st.ctx, `INSERT INTO runner_events(id,integration_id,execution_id,type,at,body) VALUES(?,?,?,?,?,?)`,
		newID(), execution.IntegrationID, execution.ID, kind, now(), body)
	return err
}

func (st *sqliteSessionTx) runnerReceipt(owner, id, digest string, result any) (runnerOperation, error) {
	if id == "" || len(id) > 200 {
		return runnerOperation{}, controlError(400, "A bounded operation_id is required.")
	}
	var previous runnerOperation
	var body []byte
	err := st.tx.QueryRowContext(st.ctx, `SELECT body FROM runner_operations WHERE integration_id=? AND id=?`, owner, id).Scan(&body)
	if err == nil {
		if err := json.Unmarshal(body, &previous); err != nil {
			return previous, err
		}
		if previous.Digest != digest {
			return previous, controlError(409, "Operation identity was reused with another body.")
		}
		return previous, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return previous, err
	}
	if result == nil {
		return previous, sql.ErrNoRows
	}
	encoded, err := runnerJSON(result)
	if err != nil {
		return previous, err
	}
	previous = runnerOperation{ID: id, Digest: digest, State: "recorded", Result: encoded}
	body, err = runnerJSON(previous)
	if err != nil {
		return previous, err
	}
	_, err = st.tx.ExecContext(st.ctx, `INSERT INTO runner_operations VALUES(?,?,?,?)`, owner, id, digest, body)
	return previous, err
}

// External mutations retain their own ordered notices in the same transaction.
// Ordinary publication receives only changes to ordinary records.
func (st *sqliteSessionTx) retainRunnerChanges() error {
	for id, positions := range st.changes.Sessions {
		session, err := st.sessionMeta(id)
		if err != nil {
			return err
		}
		if !externalSession(session) {
			continue
		}
		execution, err := st.runnerExecution(session.RunnerID)
		if err != nil {
			return err
		}
		if err := st.runnerEvent(execution, "history_changed", map[string]any{"ordinals": positions, "source_session_id": session.ID}); err != nil {
			return err
		}
		for _, table := range []string{"pending_questions", "pending_permissions"} {
			if err := st.eachMatching(table, "owner_id=?", []any{id}, func(key string) error {
				var nativeID string
				if err := st.tx.QueryRowContext(st.ctx, `SELECT record_id FROM data_`+table+` WHERE id=?`, key).Scan(&nativeID); err != nil {
					return err
				}
				var value map[string]any
				if err := decodeSQLiteRecord(st.ctx, st.tx, table, key, 1<<20, &value); err != nil {
					return err
				}
				requestID := revision([]byte(execution.ID + "/" + table + "/" + nativeID))
				var previousState string
				err := st.tx.QueryRowContext(st.ctx, `SELECT state FROM runner_requests WHERE id=?`, requestID).Scan(&previousState)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return err
				}
				if previousState == "callback_lost" || previousState == "unavailable" {
					// A genuinely re-established native callback gets a fresh human
					// request identity; an old protected answer is never rebound.
					requestID = revision([]byte(requestID + "/" + execution.TurnID))
				}
				body, err := runnerJSON(map[string]any{"id": requestID, "execution_id": execution.ID, "native_id": nativeID, "kind": table, "request": value, "state": "pending"})
				if err != nil {
					return err
				}
				result, err := st.tx.ExecContext(st.ctx, `INSERT OR IGNORE INTO runner_requests VALUES(?,?,?,?,?,?,NULL)`, requestID, execution.ID, nativeID, table, "pending", body)
				if err != nil {
					return err
				}
				if n, _ := result.RowsAffected(); n > 0 {
					return st.runnerEvent(execution, "request", json.RawMessage(body))
				}
				return nil
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *app) ordinaryRunnerChanges(changes sqliteSessionChanges) sqliteSessionChanges {
	delete(changes.Invalidated, "runner")
	external := false
	ordinary := false
	for id := range changes.Sessions {
		session, err := a.store.readSessionMeta(context.Background(), id)
		if err == nil && externalSession(session) {
			external = true
			delete(changes.Sessions, id)
		} else {
			ordinary = true
		}
	}
	if external && !ordinary {
		// An external session can share a transaction with an ordinary
		// invalidation-only change (for example refreshed harness inventory).
		for table := range changes.Invalidated {
			if table == "sessions" || table == "queue_payloads" || sqliteDomainByTable()[table].parentTable == "sessions" {
				delete(changes.Invalidated, table)
			}
		}
		return sqliteSessionChanges{Revision: changes.Revision, Invalidated: changes.Invalidated}
	}
	// Session-scoped operational records are represented by Sessions as well.
	for table, ids := range changes.Records {
		kept := []string{}
		for _, id := range ids {
			var owner string
			err := a.store.readers.QueryRow(`SELECT session_id FROM runner_executions WHERE session_id=?`, id).Scan(&owner)
			if errors.Is(err, sql.ErrNoRows) {
				kept = append(kept, id)
			}
		}
		if len(kept) == 0 {
			delete(changes.Records, table)
		} else {
			changes.Records[table] = kept
		}
	}
	if external && len(changes.Sessions) == 0 && len(changes.Records) == 0 {
		changes.Projects = map[string]bool{}
		changes.Invalidated = map[string]bool{}
	}
	return changes
}

func runnerNext(value string) string {
	n, _ := strconv.ParseUint(value, 10, 64)
	return strconv.FormatUint(n+1, 10)
}
