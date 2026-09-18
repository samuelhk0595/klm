package main

import (
	"path/filepath"
	"testing"
)

func TestPermissionPolicyFileBoundaries(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	for _, test := range []struct{ name, tool, path, decision string }{
		{"read inside", "read", "src/file.ts", "allow"},
		{"write inside", "write", "src/file.ts", "allow"},
		{"read outside", "read", "../other/file.ts", ""},
		{"write outside", "write", "../other/file.ts", "deny"},
		{"sibling prefix", "write", "../workspace-copy/file.ts", "deny"},
		{"unresolved expansion", "read", "$HOME/file.ts", ""},
		{"missing write path", "write", "", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			facts := permissionFactsFor(Permission{Details: map[string]any{"toolName": test.tool, "input": map[string]any{"path": test.path}}}, "pi", root)
			if facts.decision != test.decision {
				t.Fatalf("decision = %q, want %q", facts.decision, test.decision)
			}
		})
	}
}

func TestPermissionPolicyCommandTargets(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "workspace")
	for _, test := range []struct {
		command, cwd, decision string
		dangerous              bool
	}{
		{"rm -rf dist", root, "", true},
		{"Remove-Item ../other -Recurse -Force", root, "deny", true},
		{"rm file.txt", parent, "deny", true},
		{"rm dist; echo done", root, "", true},
		{"python cleanup.py", root, "", false},
		{"git status", root, "allow", false},
		{"git status", parent, "", false},
	} {
		facts := permissionFactsFor(Permission{Command: test.command, Path: test.cwd}, "pi", root)
		if facts.decision != test.decision || facts.dangerous != test.dangerous {
			t.Errorf("%q in %q: got %q dangerous=%v", test.command, test.cwd, facts.decision, facts.dangerous)
		}
	}
}

func TestPermissionPolicyNPMReadScope(t *testing.T) {
	dir := t.TempDir()
	cache := filepath.Join(dir, "cache")
	t.Setenv("npm_config_cache", cache)
	read := func(cwd, path string) permissionFacts {
		return permissionFactsFor(Permission{Details: map[string]any{"toolName": "read", "input": map[string]any{"path": path}}}, "pi", cwd)
	}
	first := read(filepath.Join(dir, "first"), filepath.Join(cache, "package-a", "index.js"))
	second := read(filepath.Join(dir, "second"), filepath.Join(cache, "package-b", "index.js"))
	if first.key == "" || first.key != second.key || first.decision != "" {
		t.Fatal("NPM reads must share a stable, initially unapproved cache scope across projects")
	}
	write := permissionFactsFor(Permission{Details: map[string]any{"toolName": "write", "input": map[string]any{"path": filepath.Join(cache, "package-a", "index.js")}}}, "pi", filepath.Join(dir, "first"))
	if write.key == first.key || write.decision != "deny" {
		t.Fatal("NPM read approval must not authorize writes")
	}
}

func TestPermissionPolicyDenyPrecedence(t *testing.T) {
	s := &Session{ID: "session", ProjectID: "project"}
	grants := []permissionGrant{
		{Key: "key", Kind: "read", ProjectID: "project", SessionID: "session", Decision: "allow"},
		{Key: "key", Kind: "read", Decision: "deny"},
	}
	if got := rememberedPermission(grants, s, "", "read", "key"); got != "deny" {
		t.Fatalf("global deny lost to session allow: %q", got)
	}
	grants[1].ProjectID = "other"
	if got := rememberedPermission(grants, s, "", "read", "key"); got != "allow" {
		t.Fatalf("unrelated project deny leaked: %q", got)
	}
	if got := rememberedPermission(grants, s, "", "write", "key"); got != "" {
		t.Fatalf("read scope leaked into writes: %q", got)
	}
}

func TestPermissionPolicyPatchScopes(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	for _, test := range []struct {
		patch, decision string
		dangerous       bool
	}{
		{"*** Begin Patch\n*** Add File: src/new.ts\n+export {};\n*** End Patch", "allow", false},
		{"*** Begin Patch\n*** Delete File: src/old.ts\n*** End Patch", "", true},
		{"*** Begin Patch\n*** Add File: ../other/new.ts\n+export {};\n*** End Patch", "deny", false},
	} {
		facts := permissionFactsFor(Permission{Details: map[string]any{"toolName": "apply_patch", "input": map[string]any{"patchText": test.patch}}}, "opencode", root)
		if facts.decision != test.decision || facts.dangerous != test.dangerous {
			t.Errorf("patch decision = %q, dangerous = %v", facts.decision, facts.dangerous)
		}
	}
}
