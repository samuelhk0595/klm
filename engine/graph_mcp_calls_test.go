package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func mcpBoundaryAdapter(harness string) *adapter {
	return &adapter{harness: harness, turn: &turn{ctx: context.Background()}, completed: true, nativeSettled: true, processesDrained: true,
		graph: &graphAdapterBinding{state: "open", nativeSession: "root", stop: make(chan struct{}), pluginReady: make(chan struct{}),
			hooks:         GraphAdapterHooks{Node: true, ReserveChoice: func(context.Context, string, json.RawMessage) (any, error) { return nil, nil }, Transition: func(context.Context, string) error { return nil }},
			openCodeTools: map[string]bool{"read": true, "klm_linked_graph_submit_choice": true}}}
}

func mcpGate(t *testing.T, p *adapter, op, session, call, tool string) error {
	t.Helper()
	raw, _ := json.Marshal(map[string]any{"operation": op, "sessionID": session, "callID": call, "tool": tool})
	return p.openCodeGraphGate(raw)
}

func mcpChoice(t *testing.T, p *adapter) {
	t.Helper()
	if p.harness == "opencode" {
		if err := mcpGate(t, p, "admit", "root", "choice-call", "klm_linked_graph_submit_choice"); err != nil {
			t.Fatal(err)
		}
	}
	value, err := p.callGraphTool(context.Background(), "graph_submit_choice", "choice-rpc", json.RawMessage(`{"choice":{"origin":"engine","id":"blocked"},"payload":{"reason":"test"}}`))
	if err != nil {
		t.Fatal(err)
	}
	text, _ := json.Marshal(value)
	p.graphChoiceToolSettled(string(text))
}

func assertMCPStop(t *testing.T, p *adapter, want bool) {
	t.Helper()
	select {
	case <-p.graph.stop:
		if !want {
			t.Fatal("interrupted while a parallel MCP call was pending")
		}
	default:
		if want {
			t.Fatal("settled calls did not release the native interrupt")
		}
	}
}

func TestGraphMCPBoundaryParallelChoice(t *testing.T) {
	p := mcpBoundaryAdapter("opencode")
	p.setOpenCodeGraphTools([]string{"read", "arbitrary_server_custom_tool"}, map[string]any{"arbitrary.server": map[string]any{"status": "connected"}, "agentdeck": map[string]any{"status": "connected"}})
	// Names come from resolved native hooks, not an allowlist of configured servers.
	for _, session := range []string{"root", "subtask"} {
		if err := mcpGate(t, p, "admit", session, "same-call-id", "arbitrary_server_custom_tool"); err != nil {
			t.Fatal(err)
		}
	}
	mcpChoice(t, p) // Must return its own result without waiting on itself.
	if len(p.graph.mcpCalls) != 2 {
		t.Fatal("Choice entered the remote wait set or child calls collapsed")
	}
	assertMCPStop(t, p, false)
	if err := mcpGate(t, p, "admit", "root", "late", "agentdeck_notify"); err == nil {
		t.Fatal("new call admitted after seal")
	}
	if len(p.graph.mcpCalls) != 2 {
		t.Fatal("before-rejected call counted as sent")
	}
	if err := mcpGate(t, p, "response", "root", "same-call-id", "arbitrary_server_custom_tool"); err != nil {
		t.Fatal(err)
	}
	assertMCPStop(t, p, false)
	if err := mcpGate(t, p, "response", "subtask", "same-call-id", "arbitrary_server_custom_tool"); err != nil {
		t.Fatal(err)
	}
	assertMCPStop(t, p, true)
	result := p.finishGraphAdapter(nil)
	if result.Outcome != "choice" || !result.ToolCallsSettled || result.ExternalWorkConfirmed {
		t.Fatalf("wrong boundary: %+v", result)
	}
}

func TestGraphMCPBoundaryOpenCodeErrorAndBeforeRejection(t *testing.T) {
	p := mcpBoundaryAdapter("opencode")
	if err := mcpGate(t, p, "admit", "root", "call", "external_tool"); err != nil {
		t.Fatal(err)
	}
	p.observeOpenCodeMCPPart(map[string]any{"type": "tool", "sessionID": "root", "callID": "call", "state": map[string]any{"status": "error", "error": "transport cancelled"}})
	mcpChoice(t, p)
	result := p.finishGraphAdapter(errors.New("stream disconnected"))
	var uncertain *graphUnconfirmedError
	if !errors.As(result.Error, &uncertain) || result.ToolCallsSettled || result.ChoiceState == "accepted" {
		t.Fatalf("local abort became remote completion: %+v", result)
	}
	if !result.ProcessesDrained {
		t.Fatal("remote uncertainty must remain separate from owned process drainage")
	}

	p = mcpBoundaryAdapter("opencode")
	if err := mcpGate(t, p, "admit", "root", "call", "external_tool"); err != nil {
		t.Fatal(err)
	}
	if err := mcpGate(t, p, "not_started", "root", "call", "external_tool"); err != nil {
		t.Fatal(err)
	}
	p.observeOpenCodeMCPPart(map[string]any{"type": "tool", "sessionID": "root", "callID": "call", "state": map[string]any{"status": "error"}})
	mcpChoice(t, p)
	if result := p.finishGraphAdapter(nil); result.Outcome != "choice" {
		t.Fatalf("before-denied call retained uncertainty: %+v", result)
	}
}

func TestGraphMCPBoundaryCodexResponsesAndCancellation(t *testing.T) {
	p := mcpBoundaryAdapter("codex")
	frame := func(session, turn, call, status string, result any) map[string]any {
		return map[string]any{"threadId": session, "turnId": turn, "item": map[string]any{"type": "mcpToolCall", "server": "shared-server", "tool": "enqueue", "id": call, "status": status, "result": result}}
	}
	p.observeCodexMCP("item/started", frame("root", "turn", "call", "inProgress", nil))
	p.observeCodexMCP("item/started", frame("child", "turn", "call", "inProgress", nil))
	mcpChoice(t, p)
	assertMCPStop(t, p, false)
	// An async 'task started' response ends the call; no remote-task polling.
	p.observeCodexMCP("item/completed", frame("root", "turn", "call", "completed", map[string]any{"content": []any{map[string]any{"type": "text", "text": "task started"}}}))
	assertMCPStop(t, p, false)
	p.observeCodexMCP("item/completed", frame("child", "turn", "call", "cancelled", nil))
	assertMCPStop(t, p, true)
	if len(p.graph.unresolvedMCPCallsLocked()) != 1 {
		t.Fatal("local cancellation was accepted as a returned response")
	}
	// Real error response, unlike an error string/local abort, settles the call.
	p.observeCodexMCP("item/completed", frame("child", "turn", "call", "failed", map[string]any{"content": []any{}, "isError": true}))
	p.observeCodexMCP("item/completed", frame("child", "turn", "call", "failed", nil)) // stale duplicate cannot undo proof
	result := p.finishGraphAdapter(nil)
	if result.Outcome != "choice" || !result.ToolCallsSettled {
		t.Fatalf("returned error did not settle: %+v", result)
	}
}

func TestGraphMCPBoundaryUnobservedChildAndPendingExit(t *testing.T) {
	p := mcpBoundaryAdapter("codex")
	p.noteCodexChildren(map[string]any{"type": "collabAgentToolCall", "tool": "spawnAgent", "receiverThreadIds": []any{"child"}})
	mcpChoice(t, p)
	result := p.finishGraphAdapter(nil)
	var uncertain *graphUnconfirmedError
	if !errors.As(result.Error, &uncertain) {
		t.Fatal("uncovered child thread was accepted")
	}

	p = mcpBoundaryAdapter("opencode")
	if err := mcpGate(t, p, "admit", "root", "call", "external_tool"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p.turn.ctx = ctx
	result = p.finishGraphAdapter(ctx.Err())
	if !errors.As(result.Error, &uncertain) || len(result.UncertainToolCalls) != 1 {
		t.Fatal("process exit erased an outstanding call")
	}
}

func TestGraphMCPBoundaryColdChildProof(t *testing.T) {
	for _, returned := range []bool{true, false} {
		p := mcpBoundaryAdapter("codex")
		p.noteCodexChildren(map[string]any{"type": "collabAgentToolCall", "tool": "spawnAgent", "receiverThreadIds": []any{"child"}})
		var response any
		if returned {
			response = map[string]any{"content": []any{map[string]any{"type": "text", "text": "background task started"}}}
		}
		p.auditCodexChildTurn("child", map[string]any{"id": "child-turn", "items": []any{map[string]any{"type": "mcpToolCall", "id": "child-call", "server": "shared", "tool": "enqueue", "status": "completed", "result": response}}})
		p.graph.codexAudited = map[string]bool{"child": true}
		mcpChoice(t, p)
		result := p.finishGraphAdapter(nil)
		if (result.Outcome == "choice") != returned {
			t.Fatalf("cold response=%v: %+v", returned, result)
		}
	}
	// A reused legacy thread with no timestamp and an unreturned item is not
	// proof that all this activation's calls finished.
	p := mcpBoundaryAdapter("codex")
	p.noteCodexChildren(map[string]any{"type": "collabAgentToolCall", "tool": "sendInput", "receiverThreadIds": []any{"child"}})
	p.auditCodexChildTurn("child", map[string]any{"id": "unknown-turn", "items": []any{map[string]any{"type": "mcpToolCall", "id": "unknown-call", "server": "shared", "tool": "run"}}})
	p.graph.codexAudited = map[string]bool{"child": true}
	mcpChoice(t, p)
	if result := p.finishGraphAdapter(nil); result.Outcome != "finality_unconfirmed" {
		t.Fatalf("missing coverage accepted: %+v", result)
	}
}
