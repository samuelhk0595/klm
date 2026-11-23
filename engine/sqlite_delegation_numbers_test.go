package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDelegationInvocationNumericCaptureValidation(t *testing.T) {
	store, ctx, project, parent := delegationStorageFixture(t)
	child, owned, batch, item := delegationStorageRecords(parent, project.Folder, "numbers")
	selected := MCPSelection{ConnectionID: strings.Repeat("a", 32), Tools: []string{"memory_search"}}
	item.Capture.Config.Selection.MCPConnections = []MCPSelection{selected}
	external := item.Capture.ExternalCapabilities
	external.Selection = copySelection(item.Capture.Config.Selection)
	external.Connections = []MCPConnectionSnapshot{{
		ID: selected.ConnectionID, Name: "Schema fixture", Selection: selected,
		Catalog: &MCPCatalog{Tools: []*mcp.Tool{{
			Name: "memory_search",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{
				"limit":   map[string]any{"type": "integer", "maximum": json.Number("50")},
				"large":   map[string]any{"maximum": json.Number("9007199254740993")},
				"precise": map[string]any{"multipleOf": json.Number("0.123456789012345678901")},
			}},
			OutputSchema: map[string]any{"type": "object", "properties": map[string]any{
				"score": map[string]any{"minimum": json.Number("1e-400")},
			}},
		}}},
	}}
	sealCapabilitySnapshot(external)
	if _, err := store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
		return persistDelegationStorageFixture(st, project.Folder, child, owned, batch, item)
	}); err != nil {
		t.Fatal(err)
	}
	// Exercise the same read -> state/dispatch-only write used by markDispatch.
	if _, err := store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
		next, err := st.delegationInvocation(item.ID)
		if err != nil {
			return err
		}
		next.State, next.DispatchAttempted = DelegationInvocationPreparing, true
		return st.putDelegationInvocation(next)
	}); err != nil {
		t.Fatal("unchanged numeric capture rejected", err)
	}
	for _, test := range []struct {
		name   string
		mutate func(*DelegationInvocationRecord)
	}{
		{"task", func(next *DelegationInvocationRecord) { next.Task += " changed" }},
		{"owner", func(next *DelegationInvocationRecord) { next.ParentSessionID = "other-parent" }},
		{"config", func(next *DelegationInvocationRecord) { next.Capture.Config.Model = "other/model" }},
		{"schema-number", func(next *DelegationInvocationRecord) {
			tool := next.Capture.ExternalCapabilities.Connections[0].Catalog.Tools[0]
			properties := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
			// Adjacent integers above 2^53 must not compare equal through float64.
			properties["large"].(map[string]any)["maximum"] = json.Number("9007199254740992")
			sealCapabilitySnapshot(next.Capture.ExternalCapabilities)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
				next, err := st.delegationInvocation(item.ID)
				if err != nil {
					return err
				}
				test.mutate(&next)
				return st.putDelegationInvocation(next)
			})
			if err == nil || !strings.Contains(err.Error(), "Admitted delegation task/configuration/ownership is immutable.") {
				t.Fatalf("expected immutable capture/task/ownership rejection, got %v", err)
			}
		})
	}
	// Finalization uses the same gate. Terminal no-op writes must also compare
	// equally, while the persisted generic schema numbers retain their exact text.
	if _, err := store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
		finality := DelegationFinality{NativeSettled: true, ProcessesDrained: true, ToolCallsSettled: true}
		if err := st.finishDelegationItem(item.ID, DelegationFailed, finality, "", "Fixture failure", 0); err != nil {
			return err
		}
		next, err := st.delegationInvocation(item.ID)
		if err != nil {
			return err
		}
		tool := next.Capture.ExternalCapabilities.Connections[0].Catalog.Tools[0]
		properties := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
		for property, want := range map[string]string{"limit": "50", "large": "9007199254740993", "precise": "0.123456789012345678901"} {
			key := "maximum"
			if property == "precise" {
				key = "multipleOf"
			}
			if got := properties[property].(map[string]any)[key]; got != json.Number(want) {
				t.Fatalf("%s lost numeric precision/type: %T %v", property, got, got)
			}
		}
		output := tool.OutputSchema.(map[string]any)["properties"].(map[string]any)
		if got := output["score"].(map[string]any)["minimum"]; got != json.Number("1e-400") {
			t.Fatalf("output schema lost numeric precision/type: %T %v", got, got)
		}
		return st.putDelegationInvocation(next)
	}); err != nil {
		t.Fatal("numeric capture finalization/terminal no-op rejected", err)
	}
}
