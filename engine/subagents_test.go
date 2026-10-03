package main

import (
	"context"
	"testing"
)

func TestSubagentSessionsPersistBesideSideAgent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dir := t.TempDir()
	a := &app{dir: dir, historyJournal: map[string]*sessionJournal{}, state: diskState{
		Version:  2,
		Native:   map[string]nativeSession{},
		Projects: []Project{{ID: "project", Folders: []string{}}},
		Sessions: []Session{
			{ID: "main", ProjectID: "project", Title: "Main", Workspace: "Ungrouped", Harness: "opencode", Status: "running", Events: []Event{}},
			{ID: "side", ParentID: "main", Role: sessionRoleSideAgent, ProjectID: "project", Title: "Side agent", Workspace: "Ungrouped", Harness: "opencode", Status: "idle", Events: []Event{}},
		},
	}}
	// Preserve the startup base; subsequent mutations must recover from the journal.
	if err := saveState(dir, &a.state); err != nil {
		t.Fatal(err)
	}
	p := &adapter{app: a, turn: &turn{ctx: ctx, cancel: cancel}, id: "main", harness: "opencode", keys: map[string]string{}, subagents: map[string]*adapter{}}
	first, err := p.ensureSubagent("native-one", "Explore engine", "openai/model", "high")
	if err != nil {
		t.Fatal(err)
	}
	second, err := p.ensureSubagent("native-two", "Explore desktop", "openai/model", "high")
	if err != nil {
		t.Fatal(err)
	}
	if first == nil || second == nil || first.id == second.id {
		t.Fatal("subagents were not created independently")
	}
	if err := p.startSubagent(first, "task-one"); err != nil {
		t.Fatal(err)
	}
	if err := p.put("task-one", "subagent", "Explore engine", "", "running", false, map[string]any{"childSessionId": first.id, "subagentExecutionId": "task-one"}); err != nil {
		t.Fatal(err)
	}
	if err := p.finishSubagent(first, "completed"); err != nil {
		t.Fatal(err)
	}
	if got := a.state.session(first.id); got == nil || got.Role != sessionRoleSubagent || got.Status != "idle" {
		t.Fatalf("unexpected first subagent state: %#v", got)
	}
	if got := a.state.session("main").Events[0].Status; got != "completed" {
		t.Fatalf("parent event status = %q, want completed", got)
	}
	if err := p.closeSubagents(false, false); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := replayStreamJournal(dir, &loaded); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{first.id, second.id} {
		if child := loaded.session(id); child == nil || child.ParentID != "main" || child.Role != sessionRoleSubagent || child.Status != "idle" {
			t.Fatalf("subagent identity or relationship was not recovered: %#v", child)
		}
	}
	if parent := loaded.session("main"); parent == nil || len(parent.Events) == 0 || parent.Events[0].Status != "completed" {
		t.Fatal("completed parent event was not recovered")
	}
	if loaded.linked("main") == nil || loaded.linked("main").ID != "side" {
		t.Fatal("subagents displaced the side agent relationship")
	}
}

func subagentTestAdapter(t *testing.T, a *app, harness string) *adapter {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	p := &adapter{app: a, turn: &turn{ctx: ctx, cancel: cancel}, id: "main", harness: harness, keys: map[string]string{}}
	t.Cleanup(func() {
		if err := p.closeSubagents(false, false); err != nil {
			t.Error(err)
		}
		cancel()
	})
	return p
}

func subagentTestApp(t *testing.T, harness string) *app {
	t.Helper()
	a := &app{dir: t.TempDir(), state: diskState{Version: 2, Native: map[string]nativeSession{},
		Projects: []Project{{ID: "project", Folders: []string{}}},
		Sessions: []Session{{ID: "main", ProjectID: "project", Harness: harness, Status: "running", Events: []Event{}}}}}
	if err := saveState(a.dir, &a.state); err != nil {
		t.Fatal(err)
	}
	return a
}

func TestSubagentReusedExecutionPreservesHistory(t *testing.T) {
	for _, statuses := range [][2]string{{"completed", "cancelled"}, {"error", "completed"}, {"completed", "error"}} {
		t.Run(statuses[0]+"/"+statuses[1], func(t *testing.T) {
			a := subagentTestApp(t, "opencode")
			p := subagentTestAdapter(t, a, "opencode")
			child, err := p.ensureSubagent("native-child", "Explore", "", "")
			if err != nil {
				t.Fatal(err)
			}
			for index, execution := range []string{"first", "second"} {
				if err := p.startSubagent(child, execution); err != nil {
					t.Fatal(err)
				}
				if err := p.put(execution, "subagent", "Explore", "", "running", false, map[string]any{"childSessionId": child.id, "subagentExecutionId": execution}); err != nil {
					t.Fatal(err)
				}
				if err := child.put(execution, "reasoning", "", "Working", "running", false, nil); err != nil {
					t.Fatal(err)
				}
				if index == 1 {
					// Both registration and an old completed-part replay must leave
					// the second execution running, including after a first failure.
					same, err := p.ensureSubagent("native-child", "Explore", "", "")
					if err != nil || same != child {
						t.Fatalf("did not reuse child: %v", err)
					}
					if err := p.startSubagent(child, "first"); err != nil {
						t.Fatal(err)
					}
					if err := p.finishSubagentExecution(child, "first", "completed"); err != nil {
						t.Fatal(err)
					}
				}
				if a.state.session(child.id).Status != "running" || a.state.session("main").Events[index].Status != "running" {
					t.Fatal("new execution was not running")
				}
				if err := p.finishSubagentExecution(child, execution, statuses[index]); err != nil {
					t.Fatal(err)
				}
			}
			for index, status := range statuses {
				if got := a.state.session("main").Events[index].Status; got != status {
					t.Fatalf("parent cycle %d = %q, want %q", index, got, status)
				}
				if got := a.state.session(child.id).Events[index].Status; got != status {
					t.Fatalf("child cycle %d = %q, want %q", index, got, status)
				}
			}
		})
	}
}

func TestSubagentRestoresObservationWithoutReactivation(t *testing.T) {
	a := subagentTestApp(t, "opencode")
	p := subagentTestAdapter(t, a, "opencode")
	child, err := p.ensureSubagent("native-child", "Explore", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.startSubagent(child, "first"); err != nil {
		t.Fatal(err)
	}
	if err := p.put("first", "subagent", "Explore", "", "running", false, map[string]any{"childSessionId": child.id, "subagentExecutionId": "first"}); err != nil {
		t.Fatal(err)
	}
	if err := p.finishSubagent(child, "completed"); err != nil {
		t.Fatal(err)
	}
	next := subagentTestAdapter(t, a, "opencode")
	if err := next.restoreSubagents(); err != nil {
		t.Fatal(err)
	}
	restored := next.subagents["native-child"]
	if restored == nil || restored.id != child.id || a.state.session(child.id).Status != "idle" {
		t.Fatal("observation did not restore the idle child")
	}
	if err := next.startSubagent(restored, "first"); err != nil {
		t.Fatal(err)
	}
	if a.state.session(child.id).Status != "idle" {
		t.Fatal("historical replay reactivated the child")
	}
	// A reused child can emit text before the parent's task metadata arrives.
	if err := restored.put("early", "assistant", "", "New work", "running", false, nil); err != nil {
		t.Fatal(err)
	}
	if err := next.startSubagent(restored, "second"); err != nil {
		t.Fatal(err)
	}
	if err := next.finishSubagentExecution(restored, "second", "completed"); err != nil {
		t.Fatal(err)
	}
	if e := a.state.session(child.id).Events[0]; e.Status != "completed" || str(e.Data, "subagentExecutionId") != "second" {
		t.Fatalf("early observation was not correlated and finalized: %#v", e)
	}
}

func TestCodexSubagentSequentialTurns(t *testing.T) {
	a := subagentTestApp(t, "codex")
	p := subagentTestAdapter(t, a, "codex")
	c := &codexInteractive{p: p, threadID: "main-thread", children: map[string]*codexSubagent{}}
	call := func(id, tool string) map[string]any {
		return map[string]any{"id": id, "type": "collabAgentToolCall", "tool": tool, "status": "completed", "receiverThreadIds": []any{"child-thread"}}
	}
	turn := func(id, status string) map[string]any {
		return map[string]any{"threadId": "child-thread", "turn": map[string]any{"id": id, "status": status}}
	}
	if err := c.collaboration(call("spawn", "spawnAgent")); err != nil {
		t.Fatal(err)
	}
	child := c.children["child-thread"]
	if a.state.session(child.adapter.id).Status != "running" {
		t.Fatal("spawn receipt was interpreted as child completion")
	}
	if err := c.childFrame(child, "turn/started", turn("turn-one", "inProgress")); err != nil {
		t.Fatal(err)
	}
	if err := c.childFrame(child, "turn/completed", turn("turn-one", "completed")); err != nil {
		t.Fatal(err)
	}
	// Simulate the next parent turn: children must be observable before sendInput.
	next := &codexInteractive{p: subagentTestAdapter(t, a, "codex"), threadID: "main-thread", children: map[string]*codexSubagent{}}
	if err := next.restoreChildren(); err != nil {
		t.Fatal(err)
	}
	child = next.children["child-thread"]
	if err := next.childFrame(child, "turn/started", turn("turn-two", "inProgress")); err != nil {
		t.Fatal(err)
	}
	if err := next.collaboration(call("input", "sendInput")); err != nil {
		t.Fatal(err)
	}
	if err := next.collaboration(call("spawn", "spawnAgent")); err != nil {
		t.Fatal(err)
	}
	if err := next.childFrame(child, "turn/completed", turn("turn-one", "completed")); err != nil {
		t.Fatal(err)
	}
	if a.state.session(child.adapter.id).Status != "running" || len(a.state.session("main").Events) != 2 {
		t.Fatal("reused turn was completed early or duplicated its call")
	}
	if err := next.childFrame(child, "turn/completed", turn("turn-two", "interrupted")); err != nil {
		t.Fatal(err)
	}
	if events := a.state.session("main").Events; events[0].Status != "completed" || events[1].Status != "cancelled" {
		t.Fatalf("turn results did not preserve history: %#v", events)
	}
}

func TestSubagentLegacyCallReplayRemainsIdle(t *testing.T) {
	for _, harness := range []string{"opencode", "codex"} {
		t.Run(harness, func(t *testing.T) {
			a := subagentTestApp(t, harness)
			p := subagentTestAdapter(t, a, harness)
			child, err := p.ensureSubagent("native-child", "Explore", "", "")
			if err != nil {
				t.Fatal(err)
			}
			data := map[string]any{"childSessionId": child.id, "nativeAgentId": "native-child", "nativeCallId": "call"}
			execution := "codex/subagent/call/native-child"
			if harness == "opencode" {
				data["messageID"], data["partID"] = "message", "part"
				execution = "opencode/message/part"
			}
			if err := p.put(execution, "subagent", "Explore", "", "completed", false, data); err != nil {
				t.Fatal(err)
			}
			next := subagentTestAdapter(t, a, harness)
			if err := next.restoreSubagents(); err != nil {
				t.Fatal(err)
			}
			restored := next.subagents["native-child"]
			if err := next.startSubagent(restored, execution); err != nil {
				t.Fatal(err)
			}
			if a.state.session(child.id).Status != "idle" || restored.subagentResults[execution] != "completed" || next.keys[execution] != p.keys[execution] {
				t.Fatal("legacy replay reopened a completed cycle or lost its event identity")
			}
		})
	}
}

func TestCodexSubagentCompletionBeforeCallReceipt(t *testing.T) {
	a := subagentTestApp(t, "codex")
	p := subagentTestAdapter(t, a, "codex")
	adapter, err := p.ensureSubagent("child-thread", "Explore", "", "")
	if err != nil {
		t.Fatal(err)
	}
	child := newCodexSubagent(adapter)
	c := &codexInteractive{p: p, threadID: "main-thread", children: map[string]*codexSubagent{"child-thread": child}}
	params := map[string]any{"threadId": "child-thread", "turn": map[string]any{"id": "early-turn", "status": "completed"}}
	if err := c.childFrame(child, "turn/started", params); err != nil {
		t.Fatal(err)
	}
	if err := c.childFrame(child, "turn/completed", params); err != nil {
		t.Fatal(err)
	}
	call := map[string]any{"id": "late-receipt", "tool": "sendInput", "status": "completed", "receiverThreadIds": []any{"child-thread"}}
	if err := c.collaboration(call); err != nil {
		t.Fatal(err)
	}
	if a.state.session(adapter.id).Status != "idle" || len(a.state.session("main").Events) != 1 || a.state.session("main").Events[0].Status != "completed" {
		t.Fatal("late tool receipt started a phantom child cycle")
	}
	call["id"] = "next-input"
	if err := c.collaboration(call); err != nil {
		t.Fatal(err)
	}
	if a.state.session(adapter.id).Status != "running" || len(a.state.session("main").Events) != 2 {
		t.Fatal("next input was incorrectly bound to the previous cycle")
	}
}
