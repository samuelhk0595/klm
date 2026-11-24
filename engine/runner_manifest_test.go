package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRunnerLegacyManifestProjection(t *testing.T) {
	var capture runnerCapture
	if err := json.Unmarshal([]byte(`{"manifest":[
		{"name":"team_read","description":"Read the captured instance.","inputSchema":{"type":"object","properties":{},"required":null,"additionalProperties":false}},
		{"name":"team_member_send","inputSchema":{"type":"object","properties":{"task":{"type":"string"}},"required":["task"],"additionalProperties":false}},
		{"name":"optional_read","inputSchema":{"type":"object","properties":{}}}
	]}`), &capture); err != nil {
		t.Fatal(err)
	}
	original, err := json.Marshal(capture)
	if err != nil {
		t.Fatal(err)
	}
	p := &adapter{role: "runner_worker", turn: &turn{runner: &runnerTurn{Capture: capture}}}
	tools := p.bridgeTools()
	if len(tools) != len(capture.Manifest)+len(runnerJobTools()) {
		t.Fatal("projection changed the tool roster")
	}
	for i, tool := range capture.Manifest {
		want, err := json.Marshal(tool)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			want = bytes.Replace(want, []byte(`"required":null`), []byte(`"required":[]`), 1)
		}
		got, err := json.Marshal(tools[i])
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s projection: got %s, want %s", str(tool, "name"), got, want)
		}
	}
	// Projecting and using the schema must not rewrite the retained capture.
	object(tools[0]["inputSchema"])["required"] = []string{"unexpected"}
	after, err := json.Marshal(capture)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, after) {
		t.Fatal("projection mutated the original capture")
	}
}
