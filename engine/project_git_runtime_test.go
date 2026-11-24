package main

import "testing"

func TestProjectGitCodexShellPolicy(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy map[string]any
		safe   bool
	}{
		{"defaults", nil, true},
		{"other variables", map[string]any{"inherit": "all", "set": map[string]any{"EXAMPLE": "value"}}, true},
		{"core", map[string]any{"inherit": "core"}, false},
		{"filtered", map[string]any{"exclude": []any{"KLM_*"}}, false},
		{"path override", map[string]any{"set": map[string]any{"Path": "unmanaged"}}, false},
		{"GitHub override", map[string]any{"set": map[string]any{"gh_token": "synthetic"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := verifyCodexProjectGitPolicy(map[string]any{"shell_environment_policy": tc.policy}); (err == nil) != tc.safe {
				t.Fatalf("policy check = %v, want compatible = %v", err, tc.safe)
			}
		})
	}
}
