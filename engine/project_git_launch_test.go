package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func projectGitFixtureEnvironment(t *testing.T) []string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"git", "gh"} {
		if err := os.WriteFile(filepath.Join(dir, projectGitExecutableName(name)), []byte("synthetic executable"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	return []string{
		"PATH=" + dir, "GH_TOKEN=ambient-account", "GITHUB_TOKEN=other-account",
		"GH_DEBUG=api", "GIT_TRACE=1", "GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=credential.helper",
		"GIT_CONFIG_VALUE_0=ambient-store", "HTTPS_PROXY=http://ambient-proxy", "GIT_ALLOW_PROTOCOL=ssh",
	}
}

func projectGitFixtureLease(t *testing.T, resolve projectGitResolver) []string {
	t.Helper()
	metadata := json.RawMessage(`{"projectId":"captured","defaultId":"one","identities":[{"alias":"work"}]}`)
	env, cleanup, err := newProjectGitLease(projectGitFixtureEnvironment(t), metadata, resolve)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	return env
}

func TestProjectGitLaunchConcurrentAccounts(t *testing.T) {
	// Every invocation goes through the real private transport, but no installed
	// Git/gh executable, credential file or external endpoint is ever invoked.
	accounts := []struct {
		project string
		alias   string
		token   string
	}{
		{project: "alpha", alias: "", token: "alpha-default-secret"},
		{project: "alpha", alias: "review", token: "alpha-review-secret"},
		{project: "beta", alias: "", token: "beta-default-secret"},
		{project: "beta", alias: "review", token: "beta-review-secret"},
	}
	leases := map[string][]string{}
	for _, project := range []string{"alpha", "beta"} {
		leases[project] = projectGitFixtureLease(t, func(_ context.Context, alias string) (string, error) {
			for _, account := range accounts {
				if account.project == project && account.alias == alias {
					return account.token, nil
				}
			}
			return "", errors.New("unknown alias")
		})
	}
	var wg sync.WaitGroup
	for _, account := range accounts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var output, diagnostic bytes.Buffer
			invocation := projectGitInvocation{
				name: "klm-git-identity", args: []string{account.alias, "gh", "api", "user"},
				env: leases[account.project], stdin: strings.NewReader(""), stdout: &output, stderr: &diagnostic,
				execute: func(cmd *exec.Cmd) error {
					if got := projectGitEnv(cmd.Env, "GH_TOKEN"); got != account.token {
						return fmt.Errorf("wrong account for %s/%s: got %q", account.project, account.alias, got)
					}
					if projectGitEnv(cmd.Env, "GH_HOST") != "github.com" || projectGitEnv(cmd.Env, "GH_DEBUG") != "" {
						return errors.New("host or diagnostic isolation failed")
					}
					if projectGitEnv(cmd.Env, projectGitAliasEnv) != account.alias {
						return errors.New("nested Git lost explicit alias")
					}
					if _, err := os.Stat(projectGitEnv(cmd.Env, "GH_CONFIG_DIR")); err != nil {
						return err
					}
					_, err := io.WriteString(cmd.Stdout, "synthetic account result\n")
					if err != nil {
						return err
					}
					// Both CLI streams may receive split diagnostic writes.
					_, _ = io.WriteString(cmd.Stdout, account.token[:5])
					_, _ = io.WriteString(cmd.Stdout, account.token[5:]+"\n")
					_, err = io.WriteString(cmd.Stderr, account.token+"\n")
					return err
				},
			}
			if err := projectGitRun(invocation); err != nil {
				t.Error(err)
			}
			if strings.Contains(output.String(), "secret") || strings.Contains(output.String(), "ambient") ||
				!strings.Contains(output.String(), "[redacted]") || strings.Contains(diagnostic.String(), account.token) ||
				!strings.Contains(diagnostic.String(), "[redacted]") {
				t.Errorf("credential leaked into command output: %q", output.String())
			}
		}()
	}
	wg.Wait()
}

func TestProjectGitLaunchNoFallback(t *testing.T) {
	var mu sync.Mutex
	revoked := false
	env := projectGitFixtureLease(t, func(_ context.Context, alias string) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		if revoked || alias == "missing" {
			return "", errors.New("provider diagnostics containing secret must not escape")
		}
		return "selected-secret", nil
	})
	for _, name := range []string{
		"GH_TOKEN", "GITHUB_TOKEN", "GH_DEBUG", "GIT_TRACE", "GIT_CONFIG_COUNT", "GIT_ALLOW_PROTOCOL",
	} {
		if projectGitEnv(env, name) != "" {
			t.Fatalf("ambient %s survived projection", name)
		}
	}
	if projectGitEnv(env, "HTTPS_PROXY") == "" || projectGitEnv(projectGitCleanEnvironment(env), "HTTPS_PROXY") != "" {
		t.Fatal("harness proxy must be preserved for the provider, then removed for managed Git commands")
	}
	for _, alias := range []string{"missing", ""} {
		if alias == "" {
			mu.Lock()
			revoked = true
			mu.Unlock()
		}
		for _, tool := range []string{"gh", "git"} {
			called := false
			err := projectGitRun(projectGitInvocation{
				name: tool, args: []string{"status"}, env: projectGitSetEnv(env, projectGitAliasEnv, alias),
				stdin: strings.NewReader(""), stdout: io.Discard, stderr: io.Discard,
				execute: func(*exec.Cmd) error { called = true; return nil },
			})
			if err == nil || called || strings.Contains(err.Error(), "secret") {
				t.Fatalf("%s/%s did not fail closed: called=%v err=%v", tool, alias, called, err)
			}
		}
	}
	mu.Lock()
	revoked = false
	mu.Unlock()
	calls := 0
	insufficient := errors.New("selected account lacks repository access")
	err := projectGitRun(projectGitInvocation{
		name: "gh", args: []string{"api", "repos/example/private"}, env: env,
		stdin: strings.NewReader(""), stdout: io.Discard, stderr: io.Discard,
		execute: func(*exec.Cmd) error { calls++; return insufficient },
	})
	if !errors.Is(err, insufficient) || calls != 1 {
		t.Fatalf("insufficient credentials retried or fell back: calls=%d err=%v", calls, err)
	}
}

func TestProjectGitLaunchGitPrivateCredentials(t *testing.T) {
	env := projectGitFixtureLease(t, func(context.Context, string) (string, error) { return "selected-secret", nil })
	env = projectGitSetEnv(env, "GIT_SSH_COMMAND", "ssh -F synthetic-config")
	var output bytes.Buffer
	var gitEnv []string
	var grant string
	err := projectGitRun(projectGitInvocation{
		name: "git", args: []string{"-C", "synthetic-repo", "fetch", "origin"}, env: env,
		stdin: strings.NewReader(""), stdout: &output, stderr: io.Discard,
		execute: func(cmd *exec.Cmd) error {
			if strings.Contains(strings.Join(cmd.Args, " "), "config --null --list") {
				_, err := io.WriteString(cmd.Stdout,
					"credential.helper\nambient-store\x00http.extraheader\nAuthorization: ambient-secret\x00"+
						"http.sslbackend\nschannel\x00user.name\nFixture User\x00"+
						"remote.origin.url\nhttps://github.com/example/repo\x00")
				return err
			}
			gitEnv = cmd.Env
			grant = projectGitEnv(cmd.Env, projectGitHelperEnv)
			if grant == "" || strings.Contains(strings.Join(cmd.Env, "\n"), "selected-secret") {
				return errors.New("Git environment did not isolate credential grant")
			}
			if projectGitEnv(cmd.Env, "GIT_SSH_COMMAND") != "ssh -F synthetic-config" {
				return errors.New("unrelated SSH command preference was discarded")
			}
			flags := strings.Join(cmd.Args, " ")
			if strings.Contains(flags, "ambient-secret") || !strings.Contains(flags, "credential.helper=!") ||
				!strings.Contains(flags, "http.followRedirects=false") || !strings.Contains(flags, "protocol.ssh.allow=always") {
				return errors.New("Git config precedence was not sanitized")
			}
			for _, host := range []string{"evil.example", "github.com.evil.example", "github.com:443", "github.com"} {
				for _, protocol := range []string{"http", "https"} {
					var private bytes.Buffer
					err := projectGitRun(projectGitInvocation{
						name: "klm-git-credential", args: []string{"get"}, env: cmd.Env,
						stdin:  strings.NewReader("protocol=" + protocol + "\nhost=" + host + "\n\n"),
						stdout: &private, stderr: io.Discard,
					})
					allowed := protocol == "https" && host == "github.com"
					if allowed != (err == nil) || allowed != strings.Contains(private.String(), "password=selected-secret") {
						return fmt.Errorf("wrong helper scope for %s/%s: %v", protocol, host, err)
					}
				}
			}
			_, err := io.WriteString(cmd.Stdout, "fetch result\n")
			return err
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "secret") {
		t.Fatal("private credentials reached ordinary command output")
	}
	if _, err := projectGitCall(gitEnv, "/credential", projectGitLaunchRequest{
		Grant: grant, Proto: "https", Host: "github.com",
	}); err == nil {
		t.Fatal("Git grant survived command completion")
	}
	for _, args := range [][]string{{"credential", "fill"}, {"credential-manager", "get"}, {"config", "--list"},
		{"-c", "credential.helper=ambient", "fetch"}, {"clone", "--config=http.extraHeader=ambient", "https://github.com/a/b"},
		{"fetch", "https://ambient@github.com/a/b"}} {
		if _, err := projectGitCheckGit(args); err == nil {
			t.Errorf("unsafe Git command admitted: %q", args)
		}
	}
	if _, err := projectGitCheckGit([]string{"config", "--null", "--get-regexp", `^remote\..*\.(url|gh-resolved)$`}); err != nil {
		t.Fatalf("gh remote discovery blocked: %v", err)
	}
	if _, err := projectGitCheckGit([]string{"branch", "--set-upstream-to=origin/main", "main"}); err != nil {
		t.Fatalf("gh upstream command blocked: %v", err)
	}
	for _, args := range [][]string{{"-R", "example/repo", "pr", "view"}, {"--repo=example/repo", "issue", "view", "1"}, {"api", "-h"}} {
		if err := projectGitCheckGH(args); err != nil {
			t.Fatalf("normal gh command blocked: %q: %v", args, err)
		}
	}
}

func TestProjectGitLaunchLifetimeSharedAndDiscovery(t *testing.T) {
	original := []string{"PATH=unchanged", "GH_TOKEN=legacy-shared-token"}
	for _, snapshot := range []*ProjectGitIdentitySnapshot{nil, {ProjectID: "shared"}} {
		projected, cleanup, err := (*app)(nil).projectGitEnvironment(original, snapshot)
		if err != nil || &projected[0] != &original[0] {
			t.Fatalf("shared environment changed: %v", err)
		}
		cleanup()
	}
	metadata := json.RawMessage(`{"projectId":"captured","identities":[{"alias":"work","id":"one"}]}`)
	env, cleanup, err := newProjectGitLease(projectGitFixtureEnvironment(t), metadata,
		func(context.Context, string) (string, error) { return "selected-secret", nil })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	var output bytes.Buffer
	if err := projectGitRun(projectGitInvocation{
		name: "klm-git-identity", args: []string{"list"}, env: env, stdout: &output,
	}); err != nil || !strings.Contains(output.String(), "work") || strings.Contains(output.String(), "secret") {
		t.Fatalf("alias discovery failed: %v: %s", err, output.String())
	}
	bin := filepath.SplitList(projectGitEnv(env, "PATH"))[0]
	for _, name := range []string{"git", "gh", "klm-git-identity", "klm-git-credential"} {
		if _, err := os.Stat(filepath.Join(bin, projectGitExecutableName(name))); err != nil {
			t.Fatal(err)
		}
	}
	cleanup()
	cleanup()
	if _, err := projectGitCall(env, "/list", projectGitLaunchRequest{}); err == nil {
		t.Fatal("lease survived process cleanup")
	}
	if _, err := os.Stat(bin); !os.IsNotExist(err) {
		t.Fatalf("shim files survived cleanup: %v", err)
	}
	for _, args := range [][]string{{"auth", "token"}, {"auth", "login"}, {"config", "set", "host", "evil"},
		{"api", "user", "--verbose=true"}, {"api", "user", "--hostname", "evil.example"},
		{"api", "https://evil.example/user"}, {"api", "//evil.example/user"}, {"extension", "exec", "x"}} {
		if err := projectGitCheckGH(args); err == nil {
			t.Errorf("unsafe gh command admitted: %q", args)
		}
	}
	want := `'C:/Program Files/it'"'"'s Git/helper.exe'`
	if got := projectGitShellQuote(`C:/Program Files/it's Git/helper.exe`); got != want {
		t.Fatalf("Git shell helper quoting: got %s want %s", got, want)
	}
	for _, executable := range []string{"GH.EXE", "gh", "gh.exe"} {
		name, args, handled := projectGitLaunchRoute(executable, []string{"api", "user"})
		if !handled || name != "gh" || strings.Join(args, " ") != "api user" {
			t.Errorf("executable routing failed for %s", executable)
		}
	}
	if name, args, handled := projectGitLaunchRoute("engine-dev.exe", []string{"klm-git-identity", "list"}); !handled || name != "klm-git-identity" || len(args) != 1 || args[0] != "list" {
		t.Fatal("explicit engine CLI routing failed")
	}
}
