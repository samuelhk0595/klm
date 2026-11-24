package main

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Path is the initial launch directory, not an execution boundary. Legacy Git
// fields retain their wire encoding and supply provenance only.
type runnerWorkspace struct {
	Path                string `json:"path"`
	Base                string `json:"base"`
	BaseCommit          string `json:"base_commit"`
	InitialHEAD         string `json:"initial_head"`
	ProjectRoot         string `json:"project_root"`
	CommonGitDir        string `json:"common_git_dir"`
	WorktreeGitDir      string `json:"worktree_git_dir"`
	DirectoryIdentity   string `json:"directory_identity"`
	CommonGitIdentity   string `json:"common_git_identity"`
	WorktreeGitIdentity string `json:"worktree_git_identity"`
}

func runnerCanonicalDirectory(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errors.New("Initial CWD must be an absolute existing directory.")
	}
	path, err := existingDirectory(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

func runnerGit(ctx context.Context, cwd string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "git", append([]string{"-C", cwd}, args...)...)
	// Preserve executable/system lookup but exclude inherited repository overrides.
	// Inspection has no admitted process owner. Keep it local even in a partial
	// clone; actual worker commands use the execution's frozen identity capture.
	command.Env = append(runnerGitEnvironment(), "GIT_NO_LAZY_FETCH=1")
	output, err := command.Output()
	if err != nil {
		return "", errors.New("Read-only Git workspace inspection failed.")
	}
	return strings.TrimSpace(string(output)), nil
}

func inspectRunnerWorkspace(ctx context.Context, project Project, supplied runnerWorkspace) (*runnerWorkspace, error) {
	if project.Removed {
		return nil, controlError(404, "Project not found.")
	}
	path := supplied.Path
	if path == "" {
		path = project.Folder
	}
	path, err := runnerCanonicalDirectory(path)
	if err != nil {
		return nil, err
	}
	// Observations are best effort. Git availability, repository membership, base
	// resolution and filesystem identities never gate preparation or continuation.
	gitPath := func(cwd, flag string) string {
		value, err := runnerGit(ctx, cwd, "rev-parse", "--path-format=absolute", flag)
		if err != nil {
			return ""
		}
		value, _ = runnerCanonicalDirectory(value)
		return value
	}
	projectRoot := gitPath(project.Folder, "--show-toplevel")
	common := gitPath(path, "--git-common-dir")
	gitDir := gitPath(path, "--git-dir")
	var base string
	if supplied.Base != "" && len(supplied.Base) <= 1024 {
		base, _ = runnerGit(ctx, path, "rev-parse", "--verify", "--end-of-options", supplied.Base+"^{commit}")
	}
	head, _ := runnerGit(ctx, path, "rev-parse", "--verify", "HEAD^{commit}")
	directoryIdentity, _ := runnerDirectoryIdentity(path)
	var commonIdentity, worktreeIdentity string
	if common != "" {
		commonIdentity, _ = runnerDirectoryIdentity(common)
	}
	if gitDir != "" {
		worktreeIdentity, _ = runnerDirectoryIdentity(gitDir)
	}
	return &runnerWorkspace{
		Path: path, Base: supplied.Base, BaseCommit: base, InitialHEAD: head,
		ProjectRoot: projectRoot, CommonGitDir: common, WorktreeGitDir: gitDir,
		DirectoryIdentity: directoryIdentity, CommonGitIdentity: commonIdentity,
		WorktreeGitIdentity: worktreeIdentity,
	}, nil
}

func validateRunnerLaunchDirectory(project Project, path string) error {
	if project.Removed {
		return controlError(404, "Project not found.")
	}
	_, err := runnerCanonicalDirectory(path)
	return err
}
