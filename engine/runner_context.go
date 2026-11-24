package main

import (
	"encoding/json"
	"slices"
)

// Read stored previews, not full message bodies or tool histories. These records
// are reference context; they never enter the input/authorization ledger again.
func (st *sqliteSessionTx) runnerRecentConversation(execution runnerExecution) ([]runnerConversationEntry, error) {
	rows, err := st.tx.QueryContext(st.ctx, `SELECT d.record_id,CAST(d.position AS TEXT),d.kind,
 CASE WHEN d.kind='runner_input' THEN coalesce(r.source_kind,'unattributed_input')
 ELSE 'native_assistant' END,coalesce(r.source_id,d.record_id),
 coalesce(g.text_preview,e.preview,'')
 FROM data_session_events d
 LEFT JOIN runner_inputs r ON r.id=d.record_id AND r.execution_id=?
 LEFT JOIN event_content_publications p ON p.event_key=d.id
 LEFT JOIN event_content_generations g ON g.id=p.generation_id
 LEFT JOIN event_stream_state e ON e.event_id=d.id
 WHERE d.owner_id=? AND d.kind IN ('runner_input','assistant')
 ORDER BY d.position DESC LIMIT 24`, execution.ID, execution.SessionID)
	if err != nil {
		return nil, err
	}
	entries := []runnerConversationEntry{}
	for rows.Next() {
		entry := runnerConversationEntry{SessionID: execution.SessionID}
		if err := rows.Scan(
			&entry.EventID, &entry.Ordinal, &entry.Kind, &entry.SourceKind, &entry.SourceID, &entry.Text,
		); err != nil {
			rows.Close()
			return nil, err
		}
		entry.Text = boundedText(entry.Text, 1024)
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	slices.Reverse(entries)
	entries = append(slices.Clone(execution.RecentConversation), entries...)
	if len(entries) > 24 {
		entries = entries[len(entries)-24:]
	}
	for len(entries) > 0 {
		body, err := json.Marshal(entries)
		if err != nil {
			return nil, err
		}
		if len(body) <= 32<<10 {
			break
		}
		entries = entries[1:]
	}
	return entries, nil
}
