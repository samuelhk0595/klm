package main

import (
	"database/sql"
	"net/http"
)

func (a *app) patchSQLFavorite(w http.ResponseWriter, r *http.Request, favorite bool) {
	var key string
	a.mu.Lock()
	err := a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		s, err := st.sessionMeta(r.PathValue("id"))
		if err != nil {
			return err
		}
		if err := st.tx.QueryRowContext(st.ctx, `SELECT id FROM data_session_events WHERE owner_id=? AND record_id=? ORDER BY position DESC LIMIT 1`, s.ID, r.PathValue("eventID")).Scan(&key); err != nil {
			return err
		}
		if err := st.patchEventMetadata(s.ID, key, func(e *Event) error {
			if e.Type != "assistant" {
				return controlError(404, "Assistant message not found.")
			}
			if e.Status != "" && e.Status != "completed" {
				return controlError(409, "Only completed responses can be favorited.")
			}
			e.Favorite = favorite
			return nil
		}); err != nil {
			return err
		}
		s.UpdatedAt = now()
		return st.putSessionMeta(s)
	})
	a.mu.Unlock()
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = a.store.readSnapshot(r.Context(), func(tx *sql.Tx) error { return writeSQLiteRecordContent(r.Context(), tx, w, "session_events", key) })
}
