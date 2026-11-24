package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunnerLauncherProtectedControlLF(t *testing.T) {
	t.Setenv("KLM_DEV_DISABLE_CLIENT_AUTH", "")
	a, _ := runnerRegressionApp(t, "coordinator")
	var err error
	a.clientAuth, err = newClientAuthStorage(a.dir, a.store)
	if err != nil {
		t.Fatal(err)
	}
	a.controlReady = true
	shutdown := make(chan struct{})
	listener, err := serveEngineControl(a.dir, a, shutdown)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	literal := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
	// Extract only the launcher's actual Control function. Its top-level startup
	// statements are never evaluated; this smoke opens no HTTP/model listener.
	script := `$ErrorActionPreference='Stop'; $utf8=[Text.UTF8Encoding]::new($false); ` +
		`$sid=[Security.Principal.WindowsIdentity]::GetCurrent().User; $engineDir=` + literal(a.dir) + `; ` +
		`$tokens=$null; $errors=$null; ` +
		`$ast=[Management.Automation.Language.Parser]::ParseFile(` +
		literal(filepath.Join(cwd, "validation", "Start-ForemanJourney.ps1")) + `,[ref]$tokens,[ref]$errors); ` +
		`if($errors.Count){throw 'Launcher syntax invalid'}; ` +
		`$fn=$ast.Find({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst] -and $n.Name -eq 'Control'},$true); ` +
		`Invoke-Expression $fn.Extent.Text; ` +
		`$discovery=(Control 'discover') | ConvertFrom-Json; ` +
		`if(-not $discovery.ready -or $discovery.runner_protocol -ne '1' -or -not $discovery.engine_id){throw 'Discovery failed'}; ` +
		`$pair=Control 'pair'; if($pair -notmatch '^pair [A-Za-z0-9_-]+$'){throw 'Pairing framing failed'}; ` +
		`$pair=$null; if((Control 'stop') -ne 'ok'){throw 'Stop framing failed'}; ` +
		`'PASS: protected discover, pair (redacted), stop using launcher LF; no application listener.'`
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(
		ctx, "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", script,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bounded launcher control smoke failed: %v\n%s", err, output)
	}
	select {
	case <-shutdown:
	default:
		t.Fatal("LF Stop did not reach its protected control owner")
	}
	t.Log(strings.TrimSpace(string(output)))
}
