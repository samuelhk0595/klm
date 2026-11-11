package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestV062AuthorizationBoundaries(t *testing.T) {
	grant := &AutomationGraphGrant{ProjectID: "project", AutomationID: "automation", RunID: "run", Revision: 2, ConfigurationDigest: "captured-config", Graphs: []string{"selected"}}
	run := AutomationRun{ID: "run", AutomationID: "automation", ProjectID: "project", SessionID: "owner", Status: "succeeded", GraphGrant: grant}
	activity := GraphActivity{ProjectID: "project", ConversationID: "owner", GraphID: "selected"}
	auth := GraphAuthorization{Source: "automation_configuration", Automation: grant, Text: "Captured objective"}
	if err := validateAutomationGraphGrant(activity, auth, run); err != nil {
		t.Fatal("terminal owner invalidated an admitted grant", err)
	}
	for _, invalid := range []GraphActivity{
		{ProjectID: "other-project", ConversationID: "owner", GraphID: "selected"},
		{ProjectID: "project", ConversationID: "other-owner", GraphID: "selected"},
		{ProjectID: "project", ConversationID: "owner", GraphID: "unselected"},
	} {
		if validateAutomationGraphGrant(invalid, auth, run) == nil {
			t.Fatal("cross-scope grant accepted", invalid)
		}
	}
	forged := cloneGraphAuthorization(auth)
	forged.Automation.ConfigurationDigest = "different-config"
	if validateAutomationGraphGrant(activity, forged, run) == nil {
		t.Fatal("different configuration provenance accepted")
	}
	forged = cloneGraphAuthorization(auth)
	forged.EventID = "invented-human"
	if validateAutomationGraphGrant(activity, forged, run) == nil {
		t.Fatal("synthetic human reference accepted")
	}
	var args graphInvokeArgs
	if err := decodeGraphTool(json.RawMessage(`{"operationId":"call","graphId":"selected","objective":"Objective","task":"Task"}`), &args); err != nil {
		t.Fatal("owning-turn wire omission rejected", err)
	}
	if err := decodeGraphTool(json.RawMessage(`{"authorization":[{"source":"automation_configuration","automation":{"runId":"run"}}]}`), &args); err == nil {
		t.Fatal("model-supplied automation provenance accepted")
	}
	if err := decodeGraphTool(json.RawMessage(`{"authorization":[{"eventId":"human","text":"Scope"}]}`), &args); err != nil {
		t.Fatal("legacy human authorization rejected", err)
	}
}

func TestV062ShellAndGHCompatibility(t *testing.T) {
	var legacy GraphNode
	if err := yaml.Unmarshal([]byte("type: terminal\ncommand: echo old\n"), &legacy); err != nil {
		t.Fatal(err)
	}
	if terminalShell(legacy) != "powershell" {
		t.Fatal("legacy shell changed")
	}
	legacy.Shell = "bash"
	encoded, err := yaml.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	var restored GraphNode
	if err := yaml.Unmarshal(encoded, &restored); err != nil || restored.Shell != "bash" {
		t.Fatal("shell lost on YAML round trip", err)
	}
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("GH_CONFIG_DIR", "")
	projected := ghConfigurationEnvironment([]string{"XDG_CONFIG_HOME=isolated", "GH_TOKEN=synthetic", "GH_HOST=example.test"})
	if !strings.Contains(strings.Join(projected, "\n"), "GH_CONFIG_DIR="+filepath.Join(root, "gh")) {
		t.Fatal("isolated XDG redirected gh")
	}
	t.Setenv("GH_CONFIG_DIR", filepath.Join(root, "explicit"))
	projected = ghConfigurationEnvironment(projected)
	if !strings.Contains(strings.Join(projected, "\n"), "GH_CONFIG_DIR="+filepath.Join(root, "explicit")) {
		t.Fatal("explicit gh directory lost precedence")
	}
	if !strings.Contains(strings.Join(projected, "\n"), "GH_TOKEN=synthetic\nGH_HOST=example.test") {
		t.Fatal("native token/host selection changed")
	}
}
