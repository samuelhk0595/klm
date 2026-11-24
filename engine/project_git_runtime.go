package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

func projectGitSnapshotKey(snapshot *ProjectGitIdentitySnapshot) string {
	if !snapshot.Configured() {
		return ""
	}
	data, _ := json.Marshal(snapshot)
	return revision(data)
}

type runtimeGitLease struct {
	values  map[string]string
	cleanup func()
}

// Reject custom shell policy filters instead of broadening a user's Codex
// environment policy or risking a shell that can resolve the machine Git first.
func verifyCodexProjectGitPolicy(config map[string]any) error {
	if config == nil {
		return errors.New("Codex did not return its shell environment configuration.")
	}
	policy := object(config["shell_environment_policy"])
	bad := errors.New("Codex shell environment policy is incompatible with managed project Git identities. Use the default shell inheritance without custom filters or Git launch overrides.")
	if inherit := str(policy, "inherit"); inherit != "" && inherit != "all" {
		return bad
	}
	for _, key := range []string{"include_only", "exclude"} {
		if policy[key] == nil {
			continue
		}
		if err := protocolArrayEach(policy[key], func(any) error { return bad }); err != nil {
			return bad
		}
	}
	if len(object(policy["filters"])) != 0 || policy["experimental_use_profile"] == true {
		return bad
	}
	for key := range object(policy["set"]) {
		if strings.HasPrefix(strings.ToUpper(key), "KLM_GIT_") ||
			strings.HasPrefix(strings.ToUpper(key), "GIT_") ||
			strings.HasPrefix(strings.ToUpper(key), "GH_") ||
			strings.HasPrefix(strings.ToUpper(key), "GITHUB_") || strings.EqualFold(key, "PATH") {
			return bad
		}
	}
	return nil
}

// The runtime's workload can outlive an individual harness process. Reuse the
// lease for an identical capture and retire it only at runtime shutdown.
func (t *turn) projectNativeGitEnvironment(env []string, runtime *sessionRuntime) ([]string, func(), error) {
	noop := func() {}
	key := projectGitSnapshotKey(t.projectGit)
	if key == "" {
		return env, noop, nil
	}
	if runtime == nil {
		return t.projectGitApp.projectGitEnvironment(env, t.projectGit)
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.closed {
		return nil, noop, context.Canceled
	}
	lease, found := runtime.projectGitLeases[key]
	if !found {
		projected, cleanup, err := runtime.app.projectGitEnvironment(env, t.projectGit)
		if err != nil {
			return nil, noop, err
		}
		lease = runtimeGitLease{values: map[string]string{}, cleanup: cleanup}
		for _, name := range []string{"PATH", "GH_CONFIG_DIR", "GH_PROMPT_DISABLED", projectGitURLEnv, projectGitBearerEnv} {
			lease.values[name] = projectGitEnv(projected, name)
		}
		if runtime.projectGitLeases == nil {
			runtime.projectGitLeases = map[string]runtimeGitLease{}
		}
		runtime.projectGitLeases[key] = lease
		return projected, noop, nil
	}
	env = projectGitHarnessEnvironment(env)
	for name, value := range lease.values {
		env = projectGitSetEnv(env, name, value)
	}
	return env, noop, nil
}
