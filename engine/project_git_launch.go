package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	projectGitURLEnv    = "KLM_GIT_LAUNCH_URL"
	projectGitBearerEnv = "KLM_GIT_LAUNCH_BEARER"
	projectGitHelperEnv = "KLM_GIT_HELPER_BEARER"
	projectGitAliasEnv  = "KLM_GIT_IDENTITY"
)

// A lease belongs to the owning native process, never to a caller-supplied project ID.
// The caller must retain cleanup until that process and its owned children drain.
func (a *app) projectGitEnvironment(
	env []string,
	snapshot *ProjectGitIdentitySnapshot,
) ([]string, func(), error) {
	noop := func() {}
	if snapshot == nil || !snapshot.Configured() {
		return env, noop, nil
	}
	// Freeze the reference-only capture rather than retaining a caller-owned pointer.
	metadata, err := json.Marshal(snapshot)
	if err != nil {
		return nil, noop, errors.New("cannot encode Git identity capture")
	}
	var captured ProjectGitIdentitySnapshot
	if err := json.Unmarshal(metadata, &captured); err != nil {
		return nil, noop, errors.New("cannot copy Git identity capture")
	}
	resolve := func(ctx context.Context, alias string) (string, error) {
		return a.resolveProjectGitIdentity(ctx, &captured, alias)
	}
	return newProjectGitLease(env, metadata, resolve)
}

type projectGitResolver func(context.Context, string) (string, error)

type projectGitLease struct {
	mu       sync.Mutex
	closed   bool
	grants   map[string]string
	bearer   string
	metadata json.RawMessage
	resolve  projectGitResolver
	git      string
	gh       string
	dir      string
}

type projectGitLaunchRequest struct {
	Tool  string `json:"tool"`
	Alias string `json:"alias"`
	Grant string `json:"grant"`
	Host  string `json:"host"`
	Proto string `json:"protocol"`
}

type projectGitLaunchReply struct {
	Executable string          `json:"executable,omitempty"`
	Directory  string          `json:"directory,omitempty"`
	Token      string          `json:"token,omitempty"`
	Grant      string          `json:"grant,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

func newProjectGitLease(
	env []string,
	metadata json.RawMessage,
	resolve projectGitResolver,
) ([]string, func(), error) {
	noop := func() {}
	// Resolve before adding our executable directory to PATH. No fallback to shims.
	git, err := projectGitLookPath("git", env)
	if err != nil {
		return nil, noop, errors.New("configured Git identity requires an installed Git executable")
	}
	gh, _ := projectGitLookPath("gh", env) // Git-only installations remain usable.
	executable, err := os.Executable()
	if err != nil {
		return nil, noop, err
	}
	dir, err := os.MkdirTemp("", "klm-git-launch-")
	if err != nil {
		return nil, noop, err
	}
	failed := true
	defer func() {
		if failed {
			_ = os.RemoveAll(dir)
		}
	}()
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		return nil, noop, err
	}
	if err := os.Mkdir(filepath.Join(dir, "gh"), 0700); err != nil {
		return nil, noop, err
	}
	for _, name := range []string{"git", "gh", "klm-git-identity", "klm-git-credential", "klm-git-blocked"} {
		if err := projectGitLinkExecutable(executable, filepath.Join(bin, projectGitExecutableName(name))); err != nil {
			return nil, noop, err
		}
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, noop, err
	}
	lease := &projectGitLease{
		grants: make(map[string]string), bearer: projectGitRandom(), metadata: metadata,
		resolve: resolve, git: git, gh: gh, dir: dir,
	}
	server := &http.Server{
		Handler: lease, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second,
		IdleTimeout: 15 * time.Second, MaxHeaderBytes: 4096,
	}
	go func() { _ = server.Serve(listener) }()
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			lease.mu.Lock()
			lease.closed = true
			clear(lease.grants)
			lease.mu.Unlock()
			_ = server.Close()
			_ = os.RemoveAll(dir)
		})
	}
	projected := projectGitHarnessEnvironment(env)
	projected = projectGitSetEnv(projected, "PATH", bin+string(os.PathListSeparator)+projectGitEnv(env, "PATH"))
	projected = projectGitSetEnv(projected, projectGitURLEnv, "http://"+listener.Addr().String())
	projected = projectGitSetEnv(projected, projectGitBearerEnv, lease.bearer)
	projected = projectGitSetEnv(projected, "GH_CONFIG_DIR", filepath.Join(dir, "gh"))
	projected = projectGitSetEnv(projected, "GH_PROMPT_DISABLED", "1")
	failed = false
	return projected, cleanup, nil
}

func (l *projectGitLease) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	authorized := subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+l.bearer)) == 1
	if r.Method != http.MethodPost || !authorized || r.Header.Get("Origin") != "" {
		http.Error(w, "Git launch authorization required", http.StatusForbidden)
		return
	}
	var request projectGitLaunchRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid Git launch request", http.StatusBadRequest)
		return
	}
	l.mu.Lock()
	closed := l.closed
	l.mu.Unlock()
	if closed {
		http.Error(w, "Git launch lease expired", http.StatusGone)
		return
	}
	reply, err := l.reply(r.Context(), r.URL.Path, request)
	if err != nil {
		// Resolver errors may contain storage/provider diagnostics. Never relay them.
		http.Error(w, "selected Git identity is unavailable or this operation is not permitted", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reply)
}

func (l *projectGitLease) reply(
	ctx context.Context,
	path string,
	r projectGitLaunchRequest,
) (projectGitLaunchReply, error) {
	empty := projectGitLaunchReply{}
	switch path {
	case "/list":
		return projectGitLaunchReply{Metadata: l.metadata}, nil
	case "/release":
		l.mu.Lock()
		delete(l.grants, r.Grant)
		l.mu.Unlock()
		return empty, nil
	case "/credential":
		if r.Proto != "https" || !strings.EqualFold(r.Host, "github.com") {
			return empty, errors.New("credential host denied")
		}
		l.mu.Lock()
		token, ok := l.grants[r.Grant]
		l.mu.Unlock()
		if !ok {
			return empty, errors.New("credential grant expired")
		}
		return projectGitLaunchReply{Token: token}, nil
	case "/launch":
		if r.Tool != "git" && r.Tool != "gh" {
			return empty, errors.New("unsupported tool")
		}
		if r.Tool == "gh" && l.gh == "" {
			return empty, errors.New("GitHub CLI is not installed")
		}
		token, err := l.resolve(ctx, r.Alias)
		if err != nil || token == "" || strings.ContainsAny(token, "\r\n\x00") {
			return empty, errors.New("selected identity unavailable")
		}
		if r.Tool == "gh" {
			return projectGitLaunchReply{Executable: l.gh, Directory: l.dir, Token: token}, nil
		}
		grant := projectGitRandom()
		l.mu.Lock()
		defer l.mu.Unlock()
		if l.closed {
			return empty, errors.New("Git launch lease expired")
		}
		l.grants[grant] = token
		return projectGitLaunchReply{Executable: l.git, Directory: l.dir, Grant: grant}, nil
	default:
		return empty, errors.New("unknown launch operation")
	}
}

func projectGitRandom() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("cannot generate Git launch capability")
	}
	return hex.EncodeToString(b[:])
}

func projectGitExecutableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func projectGitLinkExecutable(source, dest string) error {
	if err := os.Link(source, dest); err == nil {
		return nil
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	return errors.Join(copyErr, closeErr)
}

func projectGitLookPath(name string, env []string) (string, error) {
	for _, dir := range filepath.SplitList(projectGitEnv(env, "PATH")) {
		if dir == "" || !filepath.IsAbs(dir) {
			continue
		}
		candidate := filepath.Join(dir, projectGitExecutableName(name))
		info, err := os.Stat(candidate)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		if runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0 {
			continue
		}
		if strings.Contains(strings.ToLower(dir), "klm-git-launch-") {
			continue
		}
		return candidate, nil
	}
	return "", exec.ErrNotFound
}

func projectGitEnv(env []string, name string) string {
	for i := len(env) - 1; i >= 0; i-- {
		key, value, found := strings.Cut(env[i], "=")
		if found && strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

func projectGitSetEnv(env []string, name, value string) []string {
	result := make([]string, 0, len(env)+1)
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.EqualFold(key, name) {
			result = append(result, entry)
		}
	}
	return append(result, name+"="+value)
}

func projectGitHarnessEnvironment(env []string) []string {
	return filterProjectGitEnvironment(env, true)
}

func projectGitCleanEnvironment(env []string) []string {
	return filterProjectGitEnvironment(env, false)
}

func filterProjectGitEnvironment(env []string, harness bool) []string {
	result := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		key = strings.ToUpper(key)
		switch {
		case strings.HasPrefix(key, "GH_"), strings.HasPrefix(key, "GITHUB_"),
			strings.HasPrefix(key, "GIT_CONFIG"), strings.HasPrefix(key, "GIT_TRACE"),
			strings.HasPrefix(key, "GIT_CURL"), strings.HasPrefix(key, "GIT_SSL"),
			strings.HasPrefix(key, "GIT_HTTP"),
			strings.HasPrefix(key, "KLM_GIT_"):
			continue
		case key == "GIT_ASKPASS", key == "SSH_ASKPASS", key == "GIT_EXEC_PATH",
			key == "GIT_TERMINAL_PROMPT", key == "GCM_INTERACTIVE", key == "GCM_TRACE",
			key == "GIT_ALLOW_PROTOCOL", key == "GIT_PROTOCOL_FROM_USER", key == "GIT_PROXY_COMMAND",
			key == "CURL_HOME", key == "NETRC", key == "SSLKEYLOGFILE":
			continue
		case key == "HTTP_PROXY", key == "HTTPS_PROXY", key == "ALL_PROXY", key == "NO_PROXY":
			if !harness {
				continue
			}
		}
		result = append(result, entry)
	}
	return result
}

func projectGitCall(env []string, path string, request projectGitLaunchRequest) (projectGitLaunchReply, error) {
	var reply projectGitLaunchReply
	base, err := url.Parse(projectGitEnv(env, projectGitURLEnv))
	if err != nil || base == nil {
		return reply, errors.New("Git identity launch context is missing")
	}
	validURL := base.Scheme == "http" && base.Hostname() == "127.0.0.1" && base.Port() != ""
	if !validURL || base.User != nil || base.Path != "" || base.RawQuery != "" || base.Fragment != "" {
		return reply, errors.New("invalid Git identity launch endpoint")
	}
	body, err := json.Marshal(request)
	if err != nil {
		return reply, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base.String()+path, bytes.NewReader(body))
	if err != nil {
		return reply, errors.New("cannot create Git identity request")
	}
	req.Header.Set("Authorization", "Bearer "+projectGitEnv(env, projectGitBearerEnv))
	req.Header.Set("Content-Type", "application/json")
	transport := &http.Transport{Proxy: nil}
	defer transport.CloseIdleConnections()
	client := &http.Client{
		Transport:     transport,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	response, err := client.Do(req)
	if err != nil {
		return reply, errors.New("Git identity launch lease is unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return reply, errors.New("selected Git identity is unavailable or this operation is not permitted")
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 128*1024)).Decode(&reply); err != nil {
		return projectGitLaunchReply{}, errors.New("invalid Git identity launch response")
	}
	return reply, nil
}

// Runs before normal engine CLI dispatch, including argv[0] routing for Windows .exe shims.
func projectGitLaunchCLI() (bool, int) {
	name, args, handled := projectGitLaunchRoute(os.Args[0], os.Args[1:])
	if !handled {
		return false, 0
	}
	if name == "klm-git-blocked" {
		fmt.Fprintln(os.Stderr, "Interactive editors, browsers and pagers are disabled for configured Git identities")
		return true, 1
	}
	err := projectGitRun(projectGitInvocation{
		name: name, args: args, env: os.Environ(),
		stdin: os.Stdin, stdout: os.Stdout, stderr: os.Stderr,
	})
	if err == nil {
		return true, 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return true, exit.ExitCode()
	}
	fmt.Fprintln(os.Stderr, "Git identity:", err)
	return true, 1
}

func projectGitLaunchRoute(executable string, args []string) (string, []string, bool) {
	name := strings.TrimSuffix(strings.ToLower(filepath.Base(executable)), ".exe")
	if len(args) > 0 && args[0] == "klm-git-identity" {
		name, args = args[0], args[1:]
	}
	switch name {
	case "git", "gh", "klm-git-identity", "klm-git-credential", "klm-git-blocked":
		return name, args, true
	default:
		return name, args, false
	}
}
