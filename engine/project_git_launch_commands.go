package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// execute is a same-process fixture seam. Production always runs an owned native child.
type projectGitInvocation struct {
	name    string
	args    []string
	env     []string
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	execute func(*exec.Cmd) error
}

func (i projectGitInvocation) run(cmd *exec.Cmd) error {
	if i.execute != nil {
		return i.execute(cmd)
	}
	return cmd.Run()
}

func projectGitRun(i projectGitInvocation) error {
	if i.name == "klm-git-credential" {
		return projectGitCredential(i)
	}
	alias := projectGitEnv(i.env, projectGitAliasEnv)
	if i.name == "klm-git-identity" {
		if len(i.args) == 1 && i.args[0] == "list" {
			reply, err := projectGitCall(i.env, "/list", projectGitLaunchRequest{})
			if err != nil {
				return err
			}
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, reply.Metadata, "", "  "); err != nil {
				return errors.New("invalid captured Git identity metadata")
			}
			_, err = fmt.Fprintln(i.stdout, pretty.String())
			return err
		}
		if len(i.args) < 2 || (i.args[1] != "git" && i.args[1] != "gh") {
			return errors.New("usage: klm-git-identity list | <alias> git|gh [arguments]")
		}
		alias, i.name, i.args = i.args[0], i.args[1], i.args[2:]
	}
	i.env = projectGitSetEnv(i.env, projectGitAliasEnv, alias)
	if i.name == "gh" {
		if err := projectGitCheckGH(i.args); err != nil {
			return err
		}
	}
	if i.name == "git" {
		if _, err := projectGitCheckGit(i.args); err != nil {
			return err
		}
	}
	reply, err := projectGitCall(i.env, "/launch", projectGitLaunchRequest{Tool: i.name, Alias: alias})
	if err != nil {
		return err
	}
	if i.name == "gh" {
		return projectGitRunGH(i, reply)
	}
	defer func() {
		_, _ = projectGitCall(i.env, "/release", projectGitLaunchRequest{Grant: reply.Grant})
	}()
	return projectGitRunGit(i, reply)
}

func projectGitCheckGH(args []string) error {
	denied := errors.New("this GitHub CLI operation is disabled for configured Git identities")
	if len(args) == 0 {
		return nil
	}
	// Only built-ins are admitted: extensions, aliases, account mutation/token output,
	// and host-shell entrypoints must never receive the selected credential.
	command := ""
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "-R" || arg == "--repo" {
			index++
			continue
		}
		if strings.HasPrefix(arg, "--repo=") || strings.HasPrefix(arg, "-R") {
			continue
		}
		command = arg
		break
	}
	switch command {
	case "api", "repo", "pr", "issue", "release", "run", "workflow", "gist", "search", "status",
		"label", "project", "ruleset", "variable", "secret", "org", "cache", "help", "version", "--version", "-v", "--help", "-h":
	default:
		return denied
	}
	for index, arg := range args {
		lower := strings.ToLower(arg)
		switch {
		case lower == "--verbose", strings.HasPrefix(lower, "--verbose="), lower == "--debug",
			strings.HasPrefix(lower, "--debug="), lower == "--web", lower == "-w":
			return denied
		case strings.HasPrefix(lower, "--hostname="):
			if strings.TrimPrefix(lower, "--hostname=") != "github.com" {
				return denied
			}
		case lower == "--hostname":
			if index+1 == len(args) || !strings.EqualFold(args[index+1], "github.com") {
				return denied
			}
		case lower == "--header", strings.HasPrefix(lower, "--header="),
			strings.HasPrefix(arg, "-H"):
			// Header injection can replace Authorization or the request destination.
			return denied
		case strings.HasPrefix(lower, "//"):
			return denied
		case strings.Contains(lower, "://"):
			u, err := url.Parse(arg)
			if err == nil && u.Host != "" && (u.Scheme != "https" || !strings.EqualFold(u.Host, "github.com")) {
				return denied
			}
		}
	}
	// gh api accepts absolute endpoint URLs; github.com is the only normal CLI host,
	// but api.github.com is its fixed API endpoint, also supplied by gh internally.
	return nil
}

func projectGitRunGH(i projectGitInvocation, reply projectGitLaunchReply) error {
	dir, err := os.MkdirTemp(reply.Directory, "gh-command-")
	if err != nil {
		return errors.New("cannot create isolated GitHub CLI configuration")
	}
	defer os.RemoveAll(dir)
	env := projectGitCleanEnvironment(i.env)
	// Preserve only the private launch authority for nested managed Git commands.
	for _, key := range []string{projectGitURLEnv, projectGitBearerEnv, projectGitAliasEnv} {
		env = projectGitSetEnv(env, key, projectGitEnv(i.env, key))
	}
	env = projectGitSetEnv(env, "GH_CONFIG_DIR", dir)
	env = projectGitSetEnv(env, "GH_HOST", "github.com")
	env = projectGitSetEnv(env, "GH_TOKEN", reply.Token)
	env = projectGitSetEnv(env, "GH_PROMPT_DISABLED", "1")
	// gh may otherwise pass its environment to an editor/browser/pager process.
	blocked := filepath.Join(reply.Directory, "bin", projectGitExecutableName("klm-git-blocked"))
	for _, key := range []string{"GH_EDITOR", "GH_BROWSER", "GH_PAGER", "EDITOR", "VISUAL", "BROWSER", "PAGER"} {
		env = projectGitSetEnv(env, key, projectGitShellQuote(blocked))
	}
	cmd := exec.Command(reply.Executable, i.args...)
	stdout := newProjectGitRedactor(i.stdout, reply.Token)
	stderr := newProjectGitRedactor(i.stderr, reply.Token)
	cmd.Env, cmd.Stdin, cmd.Stdout, cmd.Stderr = env, i.stdin, stdout, stderr
	err = i.run(cmd)
	return errors.Join(err, stdout.flush(), stderr.flush())
}

// Return just the repository-selection options needed by the private config probe.
func projectGitCheckGit(args []string) ([]string, error) {
	selection := []string{}
	denied := errors.New("this Git operation is disabled for configured Git identities")
	for _, arg := range args {
		if strings.HasPrefix(arg, "--config") {
			return nil, denied
		}
		if strings.Contains(arg, "://") {
			u, err := url.Parse(arg)
			if err == nil && u.User != nil && u.Scheme != "ssh" {
				return nil, errors.New("configured Git identities do not permit embedded URL credentials")
			}
		}
	}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if !strings.HasPrefix(arg, "-") {
			name := strings.ToLower(arg)
			if strings.HasPrefix(name, "credential") || name == "shell" || name == "daemon" {
				return nil, denied
			}
			if name == "config" {
				if _, err := projectGitConfigQuery(args[index+1:]); err != nil {
					return nil, err
				}
			}
			if name == "clone" {
				for _, option := range args[index+1:] {
					if strings.HasPrefix(option, "-c") {
						return nil, denied
					}
				}
			}
			return selection, nil
		}
		switch {
		case arg == "-C", arg == "--git-dir", arg == "--work-tree", arg == "--namespace":
			if index+1 == len(args) {
				return nil, denied
			}
			selection = append(selection, arg, args[index+1])
			index++
		case strings.HasPrefix(arg, "--git-dir="), strings.HasPrefix(arg, "--work-tree="),
			strings.HasPrefix(arg, "--namespace="), arg == "--bare":
			selection = append(selection, arg)
		case arg == "--no-pager", arg == "--paginate", arg == "--no-optional-locks",
			arg == "--literal-pathspecs", arg == "--glob-pathspecs", arg == "--noglob-pathspecs",
			arg == "--icase-pathspecs", arg == "--no-replace-objects", arg == "--no-lazy-fetch",
			arg == "--version", arg == "-v", arg == "--help", arg == "-h":
		default:
			// Reject -c, --config-env, --exec-path and unknown global options.
			return nil, denied
		}
	}
	return selection, nil
}

func projectGitRunGit(i projectGitInvocation, reply projectGitLaunchReply) error {
	selection, err := projectGitCheckGit(i.args)
	if err != nil {
		return err
	}
	probeArgs := append(append([]string{}, selection...), "config", "--null", "--list", "--includes")
	probe := exec.Command(reply.Executable, probeArgs...)
	probe.Env = projectGitCleanEnvironment(i.env)
	var config bytes.Buffer
	probe.Stdout = &config
	probe.Stderr = io.Discard
	if err := i.run(probe); err != nil {
		return errors.New("cannot inspect Git configuration for identity isolation")
	}
	if err := projectGitCheckConfigOutput(i.args, config.Bytes()); err != nil {
		return err
	}
	helper := filepath.Join(reply.Directory, "bin", projectGitExecutableName("klm-git-credential"))
	configArgs, err := projectGitConfiguration(config.Bytes(), helper)
	if err != nil {
		return err
	}
	env := projectGitCleanEnvironment(i.env)
	for _, key := range []string{projectGitURLEnv, projectGitBearerEnv, projectGitAliasEnv} {
		env = projectGitSetEnv(env, key, projectGitEnv(i.env, key))
	}
	env = projectGitSetEnv(env, projectGitHelperEnv, reply.Grant)
	// Keep an empty gh account context even for native Git descendants that lose
	// our PATH entry. The managed gh shim still resolves its own selected token.
	env = projectGitSetEnv(env, "GH_CONFIG_DIR", filepath.Join(reply.Directory, "gh"))
	env = projectGitSetEnv(env, "GH_PROMPT_DISABLED", "1")
	// libcurl's optional .netrc lookup must not authenticate before our helper.
	// The safe effective user preferences are reapplied as process-only -c settings.
	for _, key := range []string{"HOME", "USERPROFILE", "CURL_HOME"} {
		env = projectGitSetEnv(env, key, reply.Directory)
	}
	env = projectGitSetEnv(env, "GIT_CONFIG_NOSYSTEM", "1")
	env = projectGitSetEnv(env, "GIT_CONFIG_GLOBAL", os.DevNull)
	env = projectGitSetEnv(env, "GIT_TERMINAL_PROMPT", "0")
	env = projectGitSetEnv(env, "GCM_INTERACTIVE", "Never")
	env = projectGitSetEnv(env, "GIT_ASKPASS", filepath.Join(reply.Directory, "bin", projectGitExecutableName("klm-git-blocked")))
	cmd := exec.Command(reply.Executable, append(configArgs, i.args...)...)
	cmd.Env, cmd.Stdin, cmd.Stdout, cmd.Stderr = env, i.stdin, i.stdout, i.stderr
	return i.run(cmd)
}

func projectGitConfiguration(data []byte, helper string) ([]string, error) {
	args := []string{}
	// Local/worktree configs remain on disk, but their auth-affecting settings are
	// overridden at highest precedence. Never copy those values into files or argv.
	for _, entry := range bytes.Split(data, []byte{0}) {
		if len(entry) == 0 {
			continue
		}
		key, value, _ := strings.Cut(string(entry), "\n")
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "url.") && (strings.HasSuffix(lower, ".insteadof") || strings.HasSuffix(lower, ".pushinsteadof")) {
			return nil, errors.New("configured Git identities require Git URL rewrites to be removed from effective configuration")
		}
		if strings.HasPrefix(lower, "remote.") && (strings.HasSuffix(lower, ".url") || strings.HasSuffix(lower, ".pushurl")) {
			u, err := url.Parse(value)
			if err == nil && u.User != nil && u.Scheme != "ssh" {
				return nil, errors.New("configured Git identities do not permit embedded remote credentials")
			}
		}
		if strings.HasPrefix(lower, "credential.") {
			value = ""
			switch lower[strings.LastIndex(lower, ".")+1:] {
			case "username":
				value = "x-access-token"
			case "interactive":
				value = "false"
			case "usehttppath":
				value = "true"
			}
			args = append(args, "-c", key+"="+value)
			continue
		}
		if strings.HasPrefix(lower, "alias.") {
			args = append(args, "-c", key+"=")
			continue
		}
		if strings.HasPrefix(lower, "http.") {
			last := lower[strings.LastIndex(lower, ".")+1:]
			original := value
			value = ""
			switch last {
			case "sslverify":
				value = "true"
			case "followredirects", "savecookies", "emptyauth":
				value = "false"
			case "extraheader", "cookiefile", "proxy", "sslcert", "sslkey":
			case "sslbackend", "sslcainfo", "sslcapath", "version", "postbuffer", "lowspeedlimit", "lowspeedtime",
				"maxrequests", "minsessions", "useragent", "schannelusesslcainfo", "schannelcheckrevoke":
				// Preserve ordinary transport/CA preferences, including Git for
				// Windows' system http.sslBackend, without accepting client auth.
				value = original
			default:
				// Unknown persistent transport/auth customization must fail closed,
				// rather than silently selecting ambient client credentials.
				return nil, errors.New("configured Git identities do not support this persistent HTTP configuration")
			}
			args = append(args, "-c", key+"="+value)
			continue
		}
		if strings.HasPrefix(lower, "remote.") && strings.HasSuffix(lower, ".proxy") {
			args = append(args, "-c", key+"=")
			continue
		}
		if projectGitSafePreference(lower) {
			args = append(args, "-c", key+"="+value)
		}
	}
	for _, value := range []string{
		"credential.helper=", "credential.https://github.com.helper=",
		"credential.username=x-access-token", "credential.interactive=false", "credential.useHttpPath=true",
		"http.extraHeader=", "http.cookieFile=", "http.saveCookies=false", "http.sslVerify=true",
		"http.followRedirects=false", "http.proxy=", "protocol.allow=never",
		"protocol.https.allow=always", "protocol.file.allow=always", "protocol.ssh.allow=always",
		"core.askPass=", "credential.helper=!" + projectGitShellQuote(helper),
		"credential.https://github.com.helper=!" + projectGitShellQuote(helper),
	} {
		args = append(args, "-c", value)
	}
	return args, nil
}

func projectGitSafePreference(key string) bool {
	switch key {
	case "user.name", "user.email", "user.signingkey", "commit.gpgsign", "tag.gpgsign",
		"gpg.format", "gpg.program", "gpg.ssh.program", "gpg.ssh.allowedsignersfile",
		"core.autocrlf", "core.eol", "core.ignorecase", "core.filemode", "core.longpaths",
		"core.excludesfile", "core.attributesfile", "safe.directory",
		"pull.rebase", "pull.ff", "push.default", "push.autosetupremote", "init.defaultbranch":
		return true
	}
	return false
}

// gh reads remote metadata through Git. Permit those reads and ordinary local
// preferences, but never config dumps, auth queries, includes or account stores.
func projectGitConfigQuery(args []string) (func(string) bool, error) {
	denied := errors.New("Git configuration access is restricted to non-authentication project preferences")
	values := []string{}
	regex := false
	for _, arg := range args {
		switch arg {
		case "--null", "-z", "--local", "--get", "--get-all":
		case "--get-regexp":
			regex = true
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, denied
			}
			values = append(values, arg)
		}
	}
	if len(values) == 0 || len(values) > 2 {
		return nil, denied
	}
	if !regex {
		if !projectGitPublicConfigKey(strings.ToLower(values[0])) {
			return nil, denied
		}
		return func(key string) bool { return strings.EqualFold(key, values[0]) }, nil
	}
	pattern := strings.TrimPrefix(values[0], "^")
	if !strings.HasPrefix(pattern, `remote\.`) && !strings.HasPrefix(pattern, `branch\.`) {
		return nil, denied
	}
	re, err := regexp.Compile(values[0])
	if err != nil {
		return nil, denied
	}
	return re.MatchString, nil
}

func projectGitPublicConfigKey(key string) bool {
	if projectGitSafePreference(key) {
		return true
	}
	last := key[strings.LastIndex(key, ".")+1:]
	if strings.HasPrefix(key, "remote.") {
		switch last {
		case "url", "pushurl", "gh-resolved", "fetch", "mirror", "prune", "tagopt":
			return true
		}
	}
	if strings.HasPrefix(key, "branch.") {
		switch last {
		case "remote", "merge", "pushremote", "rebase":
			return true
		}
	}
	return false
}

func projectGitCheckConfigOutput(args []string, data []byte) error {
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "-C" || arg == "--git-dir" || arg == "--work-tree" || arg == "--namespace" {
			index++
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if arg != "config" {
			return nil
		}
		match, err := projectGitConfigQuery(args[index+1:])
		if err != nil {
			return err
		}
		for _, entry := range bytes.Split(data, []byte{0}) {
			key, _, _ := strings.Cut(string(entry), "\n")
			if match(key) && !projectGitPublicConfigKey(strings.ToLower(key)) {
				return errors.New("Git configuration query includes authentication or execution settings")
			}
		}
		return nil
	}
	return nil
}

func projectGitShellQuote(path string) string {
	// Git invokes ! helpers through its POSIX shell even on Windows. Backslashes
	// are not shell-safe path separators there; single-quote every executable path.
	path = filepath.ToSlash(path)
	return "'" + strings.ReplaceAll(path, "'", "'\"'\"'") + "'"
}

func projectGitCredential(i projectGitInvocation) error {
	if len(i.args) != 1 {
		return errors.New("private Git credential helper")
	}
	// Git calls store/erase after authentication. Neither may write any account store.
	if i.args[0] == "store" || i.args[0] == "erase" {
		return nil
	}
	grant := projectGitEnv(i.env, projectGitHelperEnv)
	if i.args[0] != "get" || grant == "" {
		return errors.New("private Git credential helper requires an active Git command")
	}
	fields := map[string]string{}
	scanner := bufio.NewScanner(io.LimitReader(i.stdin, 64*1024))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		key, value, ok := strings.Cut(line, "=")
		if ok {
			fields[key] = value
		}
	}
	if scanner.Err() != nil {
		return errors.New("invalid private Git credential request")
	}
	reply, err := projectGitCall(i.env, "/credential", projectGitLaunchRequest{
		Grant: grant, Proto: fields["protocol"], Host: fields["host"],
	})
	if err != nil {
		_, _ = io.WriteString(i.stdout, "quit=true\n\n")
		return err
	}
	_, err = fmt.Fprintf(i.stdout, "username=x-access-token\npassword=%s\n\n", reply.Token)
	return err
}
