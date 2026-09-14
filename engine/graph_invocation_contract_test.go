package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// Boundary-only regression checks: no app, persistence I/O, Git, harness or run.
func TestGraphInvocationDecoder(t *testing.T) {
	for _, test := range []struct {
		name, raw, want string
		update          bool
	}{
		{"minimal", `{"operationId":"call-1","graphId":"checks","objective":"Check work","task":"Check the current work","authorization":[{"eventId":"user-1","text":"Run checks"}]}`, "", false},
		{"empty workspace", `{"workspace":{}}`, "", false},
		{"legacy", `{"workspace":{"mode":"original","authorizationEventId":"user-1"}}`, "", false},
		{"legacy empty reference", `{"workspace":{"authorizationEventId":""}}`, "", false},
		{"isolated", `{"workspace":{"mode":"new_worktree","baseRevision":"HEAD"}}`, "", false},
		{"reuse update", `{"workspace":{"attempt":"reuse","sourceRunId":"run-1","reuse":{"initial":"ws-1","node:join:2":"ws-2","node:fork:1:branch:a":"ws-3"}}}`, "", true},
		{"path", `{"workspace":{"path":"C:/other"}}`, "workspace.path: unknown field; workspace accepts {mode, attempt, authorizationEventId, baseRevision, sourceRunId, reuse}", false},
		{"eventId", `{"workspace":{"eventId":"user-1"}}`, "workspace.eventId", true},
		{"event_id", `{"workspace":{"event_id":"user-1"}}`, "workspace.event_id", false},
		{"userEventId", `{"workspace":{"userEventId":"user-1"}}`, "workspace.userEventId", false},
		{"scope", `{"authorization":[{"scope":"task"}]}`, "authorization[0].scope: unknown field; authorization[0] accepts {eventId, text}", false},
		{"graphId", `{"authorization":[{"graphId":"checks"}]}`, "authorization[0].graphId", false},
		{"missing event", `{"authorization":[{"text":"Run checks"}]}`, "authorization[0].eventId", false},
		{"blank text", `{"authorization":[{"eventId":"user-1","text":"  "}]}`, "authorization[0].text", false},
		{"null event", `{"authorization":[{"eventId":null,"text":"Run checks"}]}`, "authorization[0].eventId", false},
		{"wrong text type", `{"authorization":[{"eventId":"user-1","text":5}]}`, "authorization[0].text", false},
		{"authorization object", `{"authorization":{"eventId":"user-1","text":"Run checks"}}`, "authorization: expected a nonempty array", false},
		{"null authorization", `{"authorization":null}`, "authorization: expected a nonempty array", false},
		{"null workspace", `{"workspace":null}`, "workspace: expected an object", false},
		{"bad mode", `{"workspace":{"mode":"auto"}}`, "workspace.mode: expected original or new_worktree", true},
		{"blank mode", `{"workspace":{"mode":""}}`, "workspace.mode: expected original or new_worktree", false},
		{"null mode", `{"workspace":{"mode":null}}`, "workspace.mode: expected a string", false},
		{"bad attempt", `{"workspace":{"attempt":"clean"}}`, "workspace.attempt", true},
		{"bad association", `{"workspace":{"reuse":{"initial":{"path":"x"}}}}`, "workspace.reuse.initial: expected a string", true},
		{"null root", `null`, "arguments: expected an object", false},
		{"extra root", `{"path":"x"}`, "path: unknown field", false},
		{"trailing JSON", `{} {}`, "Invalid graph tool arguments", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var target any = &graphInvokeArgs{}
			if test.update {
				target = &graphActivityUpdateArgs{}
			}
			err := decodeGraphTool(json.RawMessage(test.raw), target)
			if test.want == "" && err != nil || test.want != "" && (err == nil || !strings.Contains(err.Error(), test.want)) {
				t.Fatalf("decode error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestGraphInvocationDefaults(t *testing.T) {
	d := diskState{Sessions: []Session{{ID: "main", Events: []Event{{ID: "user-1", Type: "user"}, {ID: "assistant-1", Type: "assistant"}}}}}
	for _, decision := range []GraphWorkspaceDecision{{}, {Mode: "original"}, {Mode: "new_worktree"}, {AuthorizationEventID: "user-1"}} {
		normalizeGraphWorkspaceDecision(&decision)
		if decision.Mode == "" {
			t.Fatal("missing mode was not normalized")
		}
		if err := validateGraphWorkspaceDecision(&d, "main", decision, false); err != nil {
			t.Fatal(err)
		}
		if err := validateGraphWorkspaceDecision(&d, "main", decision, true); err == nil || !strings.Contains(err.Error(), "workspace.attempt") {
			t.Fatalf("default mode bypassed the retry decision: %v", err)
		}
		decision.Attempt = "fresh"
		if err := validateGraphWorkspaceDecision(&d, "main", decision, true); err != nil {
			t.Fatal(err)
		}
	}
	for _, decision := range []GraphWorkspaceDecision{{Mode: "invalid"}, {Attempt: "invalid"}, {AuthorizationEventID: "invented"}, {Attempt: "reuse"}, {Reuse: map[string]string{"initial": "ws-1"}}} {
		if err := validateGraphWorkspaceDecision(&d, "main", decision, false); err == nil {
			t.Fatalf("invalid decision accepted: %+v", decision)
		}
	}
	for _, auth := range [][]GraphAuthorization{nil, {{EventID: "invented", Text: "Run"}}, {{EventID: "assistant-1", Text: "Run"}}, {{EventID: "user-1", Text: " "}}} {
		if err := validateGraphAuthorization(&d, "main", auth); err == nil || !strings.Contains(err.Error(), "authorization") {
			t.Fatalf("invalid activity authorization accepted: %+v (%v)", auth, err)
		}
	}
	if err := validateGraphAuthorization(&d, "main", []GraphAuthorization{{EventID: "user-1", Text: "Run checks"}}); err != nil {
		t.Fatal(err)
	}
	legacy := GraphWorkspaceDecision{Mode: "new_worktree", Attempt: "reuse", AuthorizationEventID: "user-1", BaseRevision: "abc", SourceRunID: "run-1", Reuse: map[string]string{"initial": "ws-1"}}
	d.GraphActivities = []GraphActivity{{}, {Workspace: legacy}}
	d.GraphRuns = []GraphRun{{InitialWorkspaceID: "existing"}, {Workspace: legacy}}
	normalizeGraphWorkspaceRecords(&d)
	if d.GraphActivities[0].Workspace.Mode != "original" || d.GraphRuns[0].Workspace.Mode != "original" || d.GraphRuns[0].InitialWorkspaceID != "existing" || !reflect.DeepEqual(d.GraphActivities[1].Workspace, legacy) || !reflect.DeepEqual(d.GraphRuns[1].Workspace, legacy) {
		t.Fatal("normalization lost existing decisions/provenance or omitted the default")
	}
}

func TestGraphInvocationSchema(t *testing.T) {
	auth := graphAuthorizationSchema()
	item := auth["items"].(map[string]any)
	if auth["minItems"] != 1 || item["additionalProperties"] != false || !reflect.DeepEqual(item["required"], []string{"eventId", "text"}) {
		t.Fatal("authorization schema must publish a strict {eventId, text} contract")
	}
	properties := item["properties"].(map[string]any)
	for _, key := range []string{"eventId", "text"} {
		if properties[key].(map[string]any)["type"] != "string" {
			t.Fatalf("authorization.%s must be a string", key)
		}
	}
	workspace := graphWorkspaceSchema()
	properties = workspace["properties"].(map[string]any)
	if workspace["additionalProperties"] != false || len(workspace["required"].([]string)) != 0 || len(properties) != 6 {
		t.Fatal("workspace must publish all six optional accepted fields")
	}
	mode := properties["mode"].(map[string]any)
	if mode["default"] != "original" || !reflect.DeepEqual(mode["enum"], []string{"original", "new_worktree"}) {
		t.Fatal("workspace mode default/enum mismatch")
	}
	reuse := properties["reuse"].(map[string]any)
	if reuse["additionalProperties"].(map[string]any)["type"] != "string" {
		t.Fatal("reuse associations must map to string workspace IDs")
	}
}
