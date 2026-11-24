package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The lease links this test executable as its real helper. Only the explicitly
// marked native fixture children dispatch the CLI instead of the Go test runner.
func init() {
	if os.Getenv("KLM_NATIVE_GIT_TEST_SHIM") == "1" {
		if handled, code := projectGitLaunchCLI(); handled {
			os.Exit(code)
		}
	}
}

func TestProjectGitLaunchNativeGit(t *testing.T) {
	git, err := projectGitLookPath("git", os.Environ())
	if err != nil {
		t.Skip("native Git is not available")
	}
	root := t.TempDir()
	home, repo := filepath.Join(root, "synthetic home"), filepath.Join(root, "repo")
	templates := filepath.Join(root, "empty templates")
	for _, dir := range []string{home, templates} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	global := filepath.Join(home, ".gitconfig")
	marker := filepath.Join(root, "ambient-helper-used")
	ambient := "!f() { printf used >> " + projectGitShellQuote(marker) +
		"; printf 'username=ambient\\npassword=ambient-secret\\n'; }; f"
	globalData := fmt.Sprintf("[credential]\n helper = %s\n helper = %s\n username = ambient\n"+
		"[credential \"https://github.com\"]\n helper = %s\n username = scoped\n"+
		"[credential \"https://GitHub.com/example\"]\n helper = %s\n username = path-scoped\n useHttpPath = false\n"+
		"[core]\n hooksPath = %s\n"+
		"[http \"https://github.com\"]\n extraHeader = Authorization: ambient-secret\n",
		strconv.Quote(ambient), strconv.Quote(ambient), strconv.Quote(ambient), strconv.Quote(ambient),
		strconv.Quote(filepath.ToSlash(filepath.Join(home, "custom hooks"))),
	)
	if err := os.WriteFile(global, []byte(globalData), 0600); err != nil {
		t.Fatal(err)
	}
	// Deliberately copy no account environment, real HOME, XDG config or Git config.
	env := []string{
		"PATH=" + projectGitEnv(os.Environ(), "PATH"), "HOME=" + home, "USERPROFILE=" + home,
		"XDG_CONFIG_HOME=" + filepath.Join(home, "xdg"), "TEMP=" + root, "TMP=" + root,
		"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=" + global, "GIT_TERMINAL_PROMPT=0",
		"KLM_NATIVE_GIT_TEST_SHIM=1",
	}
	for _, key := range []string{"SYSTEMROOT", "WINDIR", "COMSPEC", "PATHEXT"} {
		if value := projectGitEnv(os.Environ(), key); value != "" {
			env = append(env, key+"="+value)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	initCommand := exec.CommandContext(ctx, git, "init", "--template="+templates, repo)
	initCommand.Env = env
	if output, err := initCommand.CombinedOutput(); err != nil {
		t.Fatalf("isolated git init failed: %v: %s", err, output)
	}
	localPath := filepath.Join(repo, ".git", "config")
	localData, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatal(err)
	}
	localData = append(localData, []byte(fmt.Sprintf(
		"\n[remote \"origin\"]\n url = https://github.com/example/repo.git\n fetch = +refs/heads/*:refs/remotes/origin/*\n"+
			"[credential \"https://github.com/example/repo.git\"]\n helper = %s\n username = most-specific\n",
		strconv.Quote(ambient),
	))...)
	if err := os.WriteFile(localPath, localData, 0600); err != nil {
		t.Fatal(err)
	}
	projected, cleanup, err := newProjectGitLease(env, json.RawMessage(`{"projectId":"native-fixture"}`),
		func(context.Context, string) (string, error) { return "synthetic-native-token", nil })
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	// Test helper quoting with spaces and an apostrophe using an actual Git shell.
	dir := filepath.Dir(projectGitEnv(projected, "GH_CONFIG_DIR"))
	quotedBin := filepath.Join(dir, "helper's directory")
	if err := os.Mkdir(quotedBin, 0700); err != nil {
		t.Fatal(err)
	}
	helperName := projectGitExecutableName("klm-git-credential")
	if err := projectGitLinkExecutable(filepath.Join(dir, "bin", helperName), filepath.Join(quotedBin, helperName)); err != nil {
		t.Fatal(err)
	}
	credentialInput := ""
	run := func(args ...string) (string, error) {
		var output, diagnostic bytes.Buffer
		err := projectGitRun(projectGitInvocation{
			name: "git", args: append([]string{"-C", repo}, args...), env: projected,
			stdin: strings.NewReader(""), stdout: &output, stderr: &diagnostic,
			execute: func(cmd *exec.Cmd) error {
				probe := strings.Contains(strings.Join(cmd.Args, " "), "config --null --list")
				if probe {
					// Production inspects effective machine settings; this fixture
					// explicitly forbids reading any existing machine credentials.
					cmd.Env = projectGitSetEnv(cmd.Env, "GIT_CONFIG_NOSYSTEM", "1")
					cmd.Env = projectGitSetEnv(cmd.Env, "GIT_CONFIG_GLOBAL", global)
				} else {
					if projectGitEnv(cmd.Env, "GH_CONFIG_DIR") != projectGitEnv(projected, "GH_CONFIG_DIR") ||
						projectGitEnv(cmd.Env, "GH_PROMPT_DISABLED") != "1" {
						return errors.New("Git descendants lost the isolated gh account context")
					}
					for index, arg := range cmd.Args {
						if strings.HasPrefix(strings.ToLower(arg), "core.hookspath=") {
							return errors.New("global hook configuration was elevated into command-line scope")
						}
						cmd.Args[index] = strings.ReplaceAll(arg,
							projectGitShellQuote(filepath.Join(dir, "bin", helperName)),
							projectGitShellQuote(filepath.Join(quotedBin, helperName)),
						)
					}
					if credentialInput != "" {
						// Public git credential commands remain blocked. Only this
						// narrow native seam probes Git's private helper protocol.
						cmd.Args = append(cmd.Args[:len(cmd.Args)-1], "credential", "fill")
						cmd.Stdin = strings.NewReader(credentialInput)
					}
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				native := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
				native.Env, native.Dir = cmd.Env, cmd.Dir
				native.Stdin, native.Stdout, native.Stderr = cmd.Stdin, cmd.Stdout, cmd.Stderr
				return native.Run()
			},
		})
		if err != nil {
			return output.String(), fmt.Errorf("native managed Git: %w: %s", err, diagnostic.String())
		}
		return output.String(), nil
	}
	if output, err := run("status", "--porcelain"); err != nil || output != "" {
		t.Fatalf("configured status failed: %v", err)
	}
	if output, err := run("config", "--get", "remote.origin.url"); err != nil ||
		strings.TrimSpace(output) != "https://github.com/example/repo.git" {
		t.Fatalf("gh remote URL discovery failed: %v", err)
	}
	if output, err := run("config", "--null", "--get-regexp", `^remote\..*\.(url|gh-resolved)$`); err != nil ||
		!strings.Contains(output, "remote.origin.url") {
		t.Fatalf("gh remote regexp discovery failed: %v", err)
	}
	if _, err := run("config", "branch.fixture.remote", "origin"); err != nil {
		t.Fatalf("gh upstream metadata update failed: %v", err)
	}
	// Restore the intentional local metadata write before checking machinery-only mutation.
	if err := os.WriteFile(localPath, localData, 0600); err != nil {
		t.Fatal(err)
	}
	for _, host := range []string{"github.com", "GitHub.com", "other.example"} {
		credentialInput = "protocol=https\nhost=" + host + "\npath=example/repo.git\n\n"
		output, err := run("status")
		if host == "other.example" {
			if err == nil || strings.Contains(output, "password=") {
				t.Fatal("non-GitHub credentials did not fail closed")
			}
		} else if err != nil || !strings.Contains(output, "username=x-access-token\n") ||
			!strings.Contains(output, "password=synthetic-native-token\n") || strings.Contains(output, "ambient") {
			t.Fatalf("scoped helper ordering/host case failed for %s: %v", host, err)
		}
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("ambient credential helper ran")
	}
	for path, before := range map[string][]byte{global: []byte(globalData), localPath: localData} {
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatal("managed launch changed source Git configuration")
		}
	}
	// A filtered PATH cannot make an ordinary gh process rediscover saved accounts:
	// its remaining GH_CONFIG_DIR exists and contains no hosts/account configuration.
	entries, err := os.ReadDir(projectGitEnv(projected, "GH_CONFIG_DIR"))
	if err != nil || len(entries) != 0 || projectGitEnv(projected, "GH_TOKEN") != "" ||
		projectGitEnv(projected, "GH_PROMPT_DISABLED") != "1" {
		t.Fatal("outer GitHub CLI fallback context was not empty and noninteractive")
	}
}
