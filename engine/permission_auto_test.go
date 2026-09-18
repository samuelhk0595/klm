package main

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenCodeMCPPermissionIdentity(t *testing.T) {
	p := &adapter{app: &app{}, openCodeMCPPrefixes: openCodeMCPToolPrefixes(map[string]any{
		"agentdeck": map[string]any{"status": "connected"}, "my.server": map[string]any{"status": "disconnected"},
	})}
	for _, test := range []struct {
		tool string
		want bool
	}{
		{"agentdeck_notify", true}, {"my_server_query", true}, {"agentdeck_", false},
		{"mcp_unknown_tool", false}, {"other__tool", false}, {"read", false},
	} {
		if got := p.isOpenCodeMCPTool(test.tool); got != test.want {
			t.Errorf("%s: MCP=%v, want %v", test.tool, got, test.want)
		}
	}
}

func TestSkillNativePermissionReuse(t *testing.T) {
	p := &adapter{app: &app{}, gatedPermissions: map[string]string{"session/call": "skill"}}
	for _, test := range []struct {
		kind string
		want bool
	}{{"skill", true}, {"external_directory", true}, {"doom_loop", false}, {"bash", false}} {
		request := map[string]any{"sessionID": "session", "tool": map[string]any{"callID": "call"}, "permission": test.kind}
		if got := p.openCodePermissionApproved(request); got != test.want {
			t.Errorf("%s: approved=%v, want %v", test.kind, got, test.want)
		}
	}
}

func TestSkillResourceReadPolicy(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	cwd := filepath.Join(home, "project")
	for _, test := range []struct{ tool, path, want string }{
		{"read", ".agents/skills/example/SKILL.md", "allow"},
		{"read", ".claude/skills/example/references/guide.md", "allow"},
		{"read", ".config/opencode/skills/example/SKILL.md", "allow"},
		{"read", ".pi/agent/skills/example/SKILL.md", "allow"},
		{"read", ".agents/skills-copy/example/SKILL.md", ""},
		{"read", "unrelated/SKILL.md", ""},
		{"write", ".agents/skills/example/SKILL.md", "deny"},
	} {
		for _, harness := range []string{"opencode", "pi"} {
			req := Permission{Details: map[string]any{"toolName": test.tool, "input": map[string]any{"path": filepath.Join(home, filepath.FromSlash(test.path))}}}
			if got := permissionFactsFor(req, harness, cwd).decision; got != test.want {
				t.Errorf("%s %s %s: decision=%q, want %q", harness, test.tool, test.path, got, test.want)
			}
		}
	}
}

func TestSkillAndMCPAutomaticApproval(t *testing.T) {
	for _, test := range []struct {
		name, harness string
		req           Permission
	}{
		{"skill", "opencode", Permission{Kind: "skill"}},
		{"pi skill", "pi", Permission{Kind: "tool", Details: map[string]any{"toolName": "skill"}}},
		{"opencode MCP", "opencode", Permission{Kind: "agentdeck_notify", mcpTool: true}},
		{"codex MCP", "codex", Permission{Kind: "mcp"}},
	} {
		for _, denied := range []bool{false, true} {
			t.Run(test.name+map[bool]string{true: " saved deny", false: " default"}[denied], func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				active := &turn{ctx: ctx, cancel: cancel}
				a := &app{dir: t.TempDir(), state: permissionPersistenceState(), runs: map[string]*turn{"main": active}, historyJournal: map[string]*sessionJournal{}}
				a.state.session("main").Harness = test.harness
				scope := map[string]any{"tool": test.name}
				if denied {
					a.state.Grants = []permissionGrant{{ID: "deny", ProjectID: "project", Harness: test.harness, Kind: test.req.Kind, Key: permissionHash(scope), Decision: "deny"}}
				}
				p := &adapter{app: a, id: "main", turn: active, harness: test.harness}
				req := test.req
				req.SourceID = "call"
				calls := 0
				err := p.requestPermission(req, scope, func(allow bool) error {
					calls++
					if allow == denied {
						t.Errorf("allow=%v, saved deny=%v", allow, denied)
					}
					return nil
				})
				if err != nil || calls != 1 || len(a.state.session("main").Permissions) != 0 {
					t.Fatalf("request was not resolved automatically: calls=%d err=%v", calls, err)
				}
			})
		}
	}
}
