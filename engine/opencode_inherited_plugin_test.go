package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOpenCodeReplacesInheritedKLMGate(t *testing.T) {
	old := "file:///C:/Users/test/AppData/Roaming/klm/engine/.opencode-graph-123/klm-policy.mjs"
	userPlugin := "file:///C:/custom/klm-policy.mjs"
	config, _ := json.Marshal(map[string]any{"plugin": []string{old, userPlugin, "user-plugin@1.0.0"}, "model": "provider/model"})
	t.Setenv("OPENCODE_CONFIG_CONTENT", string(config))
	p := &adapter{app: &app{dir: t.TempDir()}, bridge: &linkedBridge{url: "http://127.0.0.1:1/mcp", token: "test-token"}, yolo: true}
	encoded, cleanup, err := p.openCodeGraphConfig()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	var result struct {
		Plugin []string `json:"plugin"`
		Model  string   `json:"model"`
	}
	if err := json.Unmarshal([]byte(encoded), &result); err != nil {
		t.Fatal(err)
	}
	if result.Model != "provider/model" || len(result.Plugin) != 3 || result.Plugin[0] != userPlugin || result.Plugin[1] != "user-plugin@1.0.0" || result.Plugin[2] == old || !strings.HasSuffix(result.Plugin[2], "/klm-policy.mjs") {
		t.Fatalf("inherited gate was not replaced while preserving user config: %+v", result)
	}
}
