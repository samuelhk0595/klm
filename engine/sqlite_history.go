package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"strconv"
	"unicode/utf8"
)

const sqliteHistoryPageBytes = 2 << 20
const sqliteMessageHistoryEventLimit = 1024

type sqliteEventContent struct {
	Key    string `json:"key"`
	Digest string `json:"digest"`
	Bytes  int64  `json:"bytes"`
}

type sqliteHistoryEvent struct {
	Ordinal  string              `json:"ordinal"`
	Revision string              `json:"revision"`
	ID       string              `json:"id"`
	Event    *Event              `json:"event,omitempty"`
	Content  *sqliteEventContent `json:"content,omitempty"`
	Preview  *sqliteEventPreview `json:"preview,omitempty"`
}

type sqliteEventPreview struct {
	Type           string `json:"type"`
	Title          string `json:"title,omitempty"`
	Status         string `json:"status,omitempty"`
	CreatedAt      string `json:"createdAt"`
	ConsultationID string `json:"consultationId,omitempty"`
	Text           string `json:"textPreview,omitempty"`
}

type sqliteHistoryPage struct {
	Unit      string               `json:"unit,omitempty"`
	SessionID string               `json:"sessionId"`
	Revision  string               `json:"revision"`
	Events    []sqliteHistoryEvent `json:"events"`
	Before    string               `json:"before"`
	After     string               `json:"after"`
	HasOlder  bool                 `json:"hasOlder"`
	HasNewer  bool                 `json:"hasNewer"`
	Count     string               `json:"count"`
}

// Indexed keyset query; no work-group expansion, cursor scan or transcript copy.
// A metadata query enforces both caps before reading any event payload. Oversized
// records remain lossless behind the digest-bound, base64 byte-content contract.
// before/after are exclusive absolute ordinals; empty selects the latest tail.
func (s *sqliteStore) readHistory(ctx context.Context, sessionID, before, after string, limit int) (sqliteHistoryPage, error) {
	return s.readHistoryWindow(ctx, sessionID, before, after, limit, false, "", "")
}

// Message-mode cursors remain absolute event ordinals. Events between messages
// are never filtered out; byte/record continuations may split a large interval.
// from/end restrict revalidation to the reader's already loaded event range.
func (s *sqliteStore) readMessageHistory(ctx context.Context, sessionID, before, after string, limit int, from, end string) (sqliteHistoryPage, error) {
	return s.readHistoryWindow(ctx, sessionID, before, after, limit, true, from, end)
}

func (s *sqliteStore) readHistoryWindow(ctx context.Context, sessionID, before, after string, limit int, messages bool, from, end string) (sqliteHistoryPage, error) {
	page := sqliteHistoryPage{SessionID: sessionID, Events: []sqliteHistoryEvent{}}
	if limit < 1 || limit > 50 || before != "" && (after != "" || from != "") || s.readers == nil {
		return page, errors.New("invalid history request")
	}
	bound := int64(math.MaxInt64)
	var err error
	if before != "" {
		bound, err = strconv.ParseInt(before, 10, 64)
	} else if after != "" {
		bound, err = strconv.ParseInt(after, 10, 64)
	}
	if err != nil || bound < 0 {
		return page, errors.New("invalid history cursor")
	}
	tx, err := s.readers.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return page, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, `SELECT value FROM store_meta WHERE key='revision'`).Scan(&page.Revision); err != nil {
		return page, err
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM data_sessions WHERE id=?)`, sessionID).Scan(&exists); err != nil {
		return page, err
	}
	if !exists {
		return page, sql.ErrNoRows
	}
	var count int64
	if err := tx.QueryRowContext(ctx, `SELECT size FROM record_collections WHERE parent_table='sessions' AND parent_id=? AND field='events'`, sessionID).Scan(&count); err != nil {
		return page, err
	}
	page.Count = strconv.FormatInt(count, 10)
	forward := after != "" || from != ""
	query := `SELECT id,record_id,position,payload_bytes,digest,revision,kind,status,created_at,correlation_id FROM data_session_events WHERE owner_id=? AND position<? ORDER BY position DESC LIMIT ?`
	if forward {
		query = `SELECT id,record_id,position,payload_bytes,digest,revision,kind,status,created_at,correlation_id FROM data_session_events WHERE owner_id=? AND position>? ORDER BY position ASC LIMIT ?`
	}
	args := []any{sessionID, bound, limit}
	if messages {
		page.Unit = "messages"
		lower, upper := int64(0), count
		if before != "" {
			upper = min(upper, bound)
		}
		if after != "" {
			lower = min(count, bound)
			if lower < count {
				lower++
			}
		}
		if from != "" {
			value, parseErr := strconv.ParseInt(from, 10, 64)
			if parseErr != nil || value < 0 {
				return page, controlError(400, "Invalid history range.")
			}
			lower = max(lower, value)
		}
		if end != "" {
			value, parseErr := strconv.ParseInt(end, 10, 64)
			if parseErr != nil || value < lower {
				return page, controlError(400, "Invalid history range.")
			}
			upper = min(upper, value)
		}
		if lower < upper {
			boundary, boundaryErr := historyMessageBoundary(ctx, tx, sessionID, lower, upper, limit, forward)
			if boundaryErr != nil {
				return page, boundaryErr
			}
			if forward {
				upper = boundary
			} else {
				lower = boundary
			}
		}
		order := "DESC"
		if forward {
			order = "ASC"
		}
		query = `SELECT id,record_id,position,payload_bytes,digest,revision,kind,status,created_at,correlation_id FROM data_session_events WHERE owner_id=? AND position>=? AND position<? ORDER BY position ` + order + ` LIMIT ?`
		args = []any{sessionID, lower, upper, sqliteMessageHistoryEventLimit}
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return page, err
	}
	type meta struct {
		key, id, digest, revision string
		position, size            int64
		preview                   sqliteEventPreview
	}
	metadata := make([]meta, 0, limit)
	for rows.Next() {
		var m meta
		if err := rows.Scan(&m.key, &m.id, &m.position, &m.size, &m.digest, &m.revision, &m.preview.Type, &m.preview.Status, &m.preview.CreatedAt, &m.preview.ConsultationID); err != nil {
			rows.Close()
			return page, err
		}
		metadata = append(metadata, m)
	}
	err = errors.Join(rows.Err(), rows.Close())
	if err != nil {
		return page, err
	}
	// Reserve room for envelope metadata and JSON escaping overhead. Payload
	// bytes are stored encoded JSON; no giant text is decoded just to size it.
	remaining := int64(sqliteHistoryPageBytes - 64<<10)
	for _, m := range metadata {
		entry := sqliteHistoryEvent{ID: m.id, Ordinal: strconv.FormatInt(m.position, 10), Revision: m.revision}
		if m.size+4096 <= remaining {
			var buffer bytes.Buffer
			if err := writeSQLiteRecordContent(ctx, tx, &buffer, "session_events", m.key); err != nil {
				return page, err
			}
			var event Event
			decoder := json.NewDecoder(bytes.NewReader(buffer.Bytes()))
			decoder.UseNumber()
			if err := decoder.Decode(&event); err != nil {
				return page, err
			}
			entry.Event = &event
		} else {
			entry.Content = &sqliteEventContent{Key: m.key, Digest: m.digest, Bytes: m.size}
			entry.Preview = &m.preview
			var title, preview string
			err := tx.QueryRowContext(ctx, `SELECT coalesce(json_extract(CAST(g.header AS TEXT),'$.title'),json_extract(CAST(e.header AS TEXT),'$.title'),''),coalesce(g.text_preview,e.preview,'') FROM data_session_events d LEFT JOIN event_content_publications p ON p.event_key=d.id LEFT JOIN event_content_generations g ON g.id=p.generation_id LEFT JOIN event_stream_state e ON e.event_id=d.id WHERE d.id=?`, m.key).Scan(&title, &preview)
			if err != nil {
				return page, err
			}
			entry.Preview.Title = boundedText(title, 512)
			if len(preview) > 4096 {
				end := 4096
				for end > 0 && !utf8.RuneStart(preview[end]) {
					end--
				}
				preview = preview[:end]
			}
			entry.Preview.Text = preview
		}
		encoded, err := json.Marshal(entry)
		if err != nil {
			return page, err
		}
		if int64(len(encoded)+1) > remaining {
			break
		}
		remaining -= int64(len(encoded) + 1)
		page.Events = append(page.Events, entry)
	}
	if !forward {
		slices.Reverse(page.Events)
	}
	if len(page.Events) > 0 {
		first, _ := strconv.ParseInt(page.Events[0].Ordinal, 10, 64)
		last, _ := strconv.ParseInt(page.Events[len(page.Events)-1].Ordinal, 10, 64)
		page.Before, page.After = strconv.FormatInt(first, 10), strconv.FormatInt(last, 10)
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM data_session_events WHERE owner_id=? AND position<?),EXISTS(SELECT 1 FROM data_session_events WHERE owner_id=? AND position>?)`, sessionID, first, sessionID, last).Scan(&page.HasOlder, &page.HasNewer); err != nil {
			return page, err
		}
	}
	return page, tx.Commit()
}

func historyMessageBoundary(ctx context.Context, tx *sql.Tx, owner string, lower, upper int64, limit int, forward bool) (int64, error) {
	order := "DESC"
	if forward {
		order = "ASC"
	}
	// Each branch uses (owner_id,kind,position). At most 3*(limit+1) message
	// ordinals are merged; a long tool group does not cause a payload/cursor scan.
	branch := `SELECT position FROM (SELECT position FROM data_session_events WHERE owner_id=? AND kind=? AND position>=? AND position<? ORDER BY position ` + order + ` LIMIT ?)`
	query := `SELECT position FROM (` + branch + ` UNION ALL ` + branch + ` UNION ALL ` + branch + `) ORDER BY position ` + order + ` LIMIT ?`
	args := []any{}
	for _, kind := range []string{"user", "assistant", "agent_prompt"} {
		args = append(args, owner, kind, lower, upper, limit+1)
	}
	args = append(args, limit+1)
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	positions := []int64{}
	for rows.Next() {
		var position int64
		if err := rows.Scan(&position); err != nil {
			rows.Close()
			return 0, err
		}
		positions = append(positions, position)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return 0, err
	}
	if forward {
		if len(positions) > limit {
			return positions[limit], nil
		}
		return upper, nil
	}
	if len(positions) > limit {
		return positions[limit-1], nil
	}
	return lower, nil
}
