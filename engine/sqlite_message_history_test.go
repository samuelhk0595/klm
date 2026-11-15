package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
)

func messageHistoryStore(t *testing.T, events []Event) *sqliteStore {
	t.Helper()
	ctx := context.Background()
	state := sqliteLegacyFixture()
	state.Sessions[0].Events = events
	store, path := sqliteTestStore(t)
	body, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := stageLegacyCheckpoint(ctx, store, bytes.NewReader(body)); err != nil {
		t.Fatal(err)
	}
	if err := importStagedDomains(ctx, store); err != nil {
		t.Fatal(err)
	}
	if err := store.openReaders(ctx, path); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestSQLiteMessageHistoryPages(t *testing.T) {
	ctx := context.Background()
	events := []Event{}
	for index := 0; index < 30; index++ {
		kind := []string{"user", "assistant", "agent_prompt"}[index%3]
		events = append(events, Event{ID: fmt.Sprintf("message-%d", index), Type: kind, Text: "Message"})
		for tool := 0; tool < 10; tool++ {
			events = append(events, Event{ID: fmt.Sprintf("work-%d-%d", index, tool), Type: "tool", Text: "Work"})
		}
	}
	store := messageHistoryStore(t, events)
	tail, err := store.readMessageHistory(ctx, "a", "", "", 25, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if tail.Unit != "messages" || len(tail.Events) != 275 || tail.Before != "55" || tail.After != "329" || !tail.HasOlder || tail.HasNewer {
		t.Fatalf("message tail did not retain work events: unit=%s events=%d before=%s after=%s", tail.Unit, len(tail.Events), tail.Before, tail.After)
	}
	older, err := store.readMessageHistory(ctx, "a", tail.Before, "", 25, "", "")
	if err != nil || len(older.Events) != 55 || older.Before != "0" || older.After != "54" || older.HasOlder {
		t.Fatalf("older message page: %+v, %v", older, err)
	}
	newer, err := store.readMessageHistory(ctx, "a", "", older.After, 25, "", "")
	if err != nil || newer.Before != tail.Before || newer.After != tail.After || len(newer.Events) != len(tail.Events) {
		t.Fatalf("forward message page lost continuity: %+v, %v", newer, err)
	}
	bounded, err := store.readMessageHistory(ctx, "a", "", "", 10, "55", "165")
	if err != nil || len(bounded.Events) != 110 || bounded.Before != "55" || bounded.After != "164" {
		t.Fatalf("bounded window refresh: %+v, %v", bounded, err)
	}
	legacy, err := store.readHistory(ctx, "a", "", "", 25)
	if err != nil || len(legacy.Events) != 25 || legacy.Unit != "" {
		t.Fatal("legacy event readers changed", err)
	}
	for index, entry := range tail.Events {
		if entry.Ordinal != strconv.Itoa(55+index) || entry.Event == nil || entry.Event.ID != events[55+index].ID {
			t.Fatal("message page omitted or reordered an intermediate event")
		}
	}
}

func TestSQLiteMessageHistoryLongWorkContinuation(t *testing.T) {
	ctx := context.Background()
	events := []Event{{ID: "user", Type: "user", Text: "Start"}}
	for index := 0; index < 2200; index++ {
		events = append(events, Event{ID: fmt.Sprintf("tool-%d", index), Type: "tool", Text: "Work"})
	}
	events = append(events, Event{ID: "answer", Type: "assistant", Text: "Done"})
	store := messageHistoryStore(t, events)
	after, visited := "", 0
	for visited < len(events) {
		page, err := store.readMessageHistory(ctx, "a", "", after, 25, "0", strconv.Itoa(len(events)))
		if err != nil || len(page.Events) == 0 || len(page.Events) > sqliteMessageHistoryEventLimit {
			t.Fatalf("invalid transport continuation: %d rows, %v", len(page.Events), err)
		}
		encoded, _ := json.Marshal(page)
		if len(encoded) > sqliteHistoryPageBytes {
			t.Fatal("message page exceeded the byte budget")
		}
		for _, entry := range page.Events {
			if entry.Ordinal != strconv.Itoa(visited) || entry.Event == nil || entry.Event.ID != events[visited].ID {
				t.Fatal("long work continuation skipped an event")
			}
			visited++
		}
		after = page.After
	}
	tail, err := store.readMessageHistory(ctx, "a", "", "", 25, "", "")
	if err != nil || tail.After != strconv.Itoa(len(events)-1) || len(tail.Events) != sqliteMessageHistoryEventLimit || !tail.HasOlder {
		t.Fatal("bounded transport no longer opens at the real tail", err)
	}
}
