package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func fixtureGitVerifier(_ context.Context, token string) (string, error) {
	switch token {
	case "klm-synthetic-github-alice":
		return "synthetic-alice", nil
	case "klm-synthetic-github-bob":
		return "synthetic-bob", nil
	default:
		return "", errors.New("synthetic transport error including " + token)
	}
}

func gitIdentityRequest(t *testing.T, handler http.Handler, method, path string, body any, want int) []byte {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, "http://localhost"+path, bytes.NewReader(encoded))
	req.RemoteAddr = "127.0.0.1:12345"
	if method != http.MethodDelete {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-KLM-Read-Contract", sqliteReadContract)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != want {
		t.Fatalf("%s %s: want %d, got %d %s", method, path, want, res.Code, res.Body.String())
	}
	return res.Body.Bytes()
}

func gitIdentityList(t *testing.T, body []byte) []GitIdentity {
	t.Helper()
	var result struct {
		Identities []GitIdentity `json:"identities"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil || result.Identities == nil {
		t.Fatalf("invalid identity list: %s (%v)", body, err)
	}
	var raw map[string][]map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	for _, identity := range raw["identities"] {
		if len(identity) != 4 {
			t.Fatalf("public identity must have exactly four metadata fields: %s", body)
		}
	}
	return result.Identities
}

func TestProjectGitIdentityHTTPContract(t *testing.T) {
	if !developmentBuild {
		t.Skip("real router fixture uses dev auth bypass")
	}
	t.Setenv("KLM_DEV_DISABLE_CLIENT_AUTH", "1")
	t.Setenv("KLM_DEV_SYNTHETIC_GIT_IDENTITIES", "1")
	a, session := sqliteAutomationApp(t)
	var err error
	a.clientAuth, err = newClientAuthStorage(a.dir, a.store)
	if err != nil {
		t.Fatal(err)
	}
	handler := a.routes()
	base := "/api/projects/" + session.ProjectID + "/git-identities"
	call := func(method, suffix string, body any, status int) []byte {
		return gitIdentityRequest(t, handler, method, base+suffix, body, status)
	}
	if len(gitIdentityList(t, call("GET", "", nil, 200))) != 0 {
		t.Fatal("new project must be shared")
	}
	call("POST", "/test", map[string]string{"token": "klm-synthetic-github-alice"}, 200)
	call("POST", "/test", map[string]string{"token": "rejected-synthetic-token"}, 400)
	call("POST", "", map[string]string{"alias": "_bad", "token": "klm-synthetic-github-alice"}, 400)
	call("POST", "", map[string]string{"alias": "bad name", "token": "klm-synthetic-github-alice"}, 400)
	call("POST", "", map[string]string{"alias": strings.Repeat("x", 41), "token": "klm-synthetic-github-alice"}, 400)
	call("POST", "", map[string]string{"alias": "work", "token": strings.Repeat("x", 4097)}, 400)
	list := gitIdentityList(t, call("POST", "", map[string]any{
		"alias": "Work_1", "token": "klm-synthetic-github-alice", "makeDefault": false,
	}, 201))
	first := list[0]
	if !first.IsDefault || first.Login != "synthetic-alice" {
		t.Fatal("first registration must be default and use verified login", first)
	}
	call("POST", "", map[string]string{"alias": "WORK_1", "token": "klm-synthetic-github-bob"}, 409)
	before, err := a.existingProjectGitIdentity(a.ctx, session.ProjectID, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	call("POST", "/test", map[string]string{"identityId": first.ID}, 200)
	call("PATCH", "/"+first.ID, map[string]string{"token": "rejected-synthetic-token"}, 400)
	after, err := a.existingProjectGitIdentity(a.ctx, session.ProjectID, first.ID)
	if err != nil || before.RowVersion != after.RowVersion || !bytes.Equal(before.EncryptedToken, after.EncryptedToken) {
		t.Fatal("testing or invalid replacement changed the saved credential", err)
	}
	if bytes.Contains(after.EncryptedToken, []byte("klm-synthetic-github-alice")) {
		t.Fatal("token was stored in plaintext")
	}
	call("POST", "/test", map[string]string{"identityId": first.ID, "token": "klm-synthetic-github-bob"}, 400)
	call("DELETE", "/"+first.ID, nil, 200)
	if len(gitIdentityList(t, call("GET", "", nil, 200))) != 0 {
		t.Fatal("last removal must restore shared mode")
	}
	// The same routes still pass through paired-client authentication.
	t.Setenv("KLM_DEV_DISABLE_CLIENT_AUTH", "0")
	a.clientAuth, err = newClientAuthStorage(a.dir, a.store)
	if err != nil {
		t.Fatal(err)
	}
	gitIdentityRequest(t, a.routes(), "GET", base, nil, 401)
}

func TestProjectGitIdentitySnapshotRevocation(t *testing.T) {
	a, session := sqliteAutomationApp(t)
	mux := http.NewServeMux()
	a.projectGitIdentityRoutes(mux, fixtureGitVerifier)
	base := "/api/projects/" + session.ProjectID + "/git-identities"
	call := func(method, suffix string, body any, status int) []byte {
		return gitIdentityRequest(t, mux, method, base+suffix, body, status)
	}
	shared, err := a.captureProjectGitIdentities(a.ctx, session.ProjectID)
	if err != nil || shared == nil || shared.Configured() {
		t.Fatal("shared capture is not an explicit nonnil empty snapshot", err)
	}
	general, err := a.captureProjectGitIdentities(a.ctx, "")
	if err != nil || general != nil {
		t.Fatal("general capture must be absent", err)
	}
	encoded, _ := json.Marshal(shared)
	var restored *ProjectGitIdentitySnapshot
	if json.Unmarshal(encoded, &restored) != nil || restored == nil || restored.Configured() {
		t.Fatal("shared capture lost its presence during serialization")
	}
	first := gitIdentityList(t, call("POST", "", map[string]string{
		"alias": "Work", "token": "klm-synthetic-github-alice",
	}, 201))[0]
	old, err := a.captureProjectGitIdentities(a.ctx, session.ProjectID)
	if err != nil || !old.Configured() {
		t.Fatal("configured capture missing", err)
	}
	list := gitIdentityList(t, call("POST", "", map[string]any{
		"alias": "Personal", "token": "klm-synthetic-github-bob", "makeDefault": true,
	}, 201))
	second := list[1]
	if list[0].IsDefault || !second.IsDefault {
		t.Fatal("explicit default was not changed atomically", list)
	}
	call("PATCH", "/"+first.ID, map[string]string{"alias": "Renamed"}, 200)
	token, err := a.resolveProjectGitIdentity(a.ctx, old, "work")
	if err != nil || token != "klm-synthetic-github-alice" {
		t.Fatal("alias rename retargeted the admitted capture", err)
	}
	token, err = a.resolveProjectGitIdentity(a.ctx, old, "")
	if err != nil || token != "klm-synthetic-github-alice" {
		t.Fatal("default change retargeted the admitted capture", err)
	}
	if _, err := a.resolveProjectGitIdentity(a.ctx, old, "Personal"); err == nil {
		t.Fatal("old capture acquired a new alias")
	}
	call("PATCH", "/"+first.ID, map[string]string{"token": "klm-synthetic-github-bob"}, 200)
	if _, err := a.resolveProjectGitIdentity(a.ctx, old, ""); err == nil {
		t.Fatal("replaced credential version still resolves")
	}
	call("DELETE", "/"+second.ID, nil, 200)
	list = gitIdentityList(t, call("GET", "", nil, 200))
	if len(list) != 1 || !list[0].IsDefault || list[0].ID != first.ID {
		t.Fatal("removing default did not promote oldest remaining stable identity", list)
	}
	current, err := a.captureProjectGitIdentities(a.ctx, session.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	project, err := a.store.readProjectMeta(a.ctx, session.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	project.Removed = true
	_, err = a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
		return st.putRoot("projects", project.ID, project)
	})
	if err != nil {
		t.Fatal(err)
	}
	call("GET", "", nil, 404)
	call("POST", "/test", map[string]string{"identityId": first.ID}, 404)
	if _, err := a.resolveProjectGitIdentity(a.ctx, current, ""); err == nil {
		t.Fatal("removed project credentials still resolve")
	}
	if _, err := a.captureProjectGitIdentities(a.ctx, project.ID); err == nil {
		t.Fatal("removed project still admits a capture")
	}
	project.Removed = false
	_, err = a.store.sessionTransaction(a.ctx, func(st *sqliteSessionTx) error {
		return st.putRoot("projects", project.ID, project)
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.resolveProjectGitIdentity(a.ctx, current, ""); err != nil {
		t.Fatal("project restoration lost retained credentials", err)
	}
	call("DELETE", "/"+first.ID, nil, 200)
	if _, err := a.resolveProjectGitIdentity(a.ctx, current, ""); err == nil {
		t.Fatal("removed credential silently resolved or entered shared mode")
	}
	if shared.Configured() || len(shared.Identities) != 0 {
		t.Fatal("admitted shared capture changed after configuration")
	}
	encoded, _ = json.Marshal(current)
	if bytes.Contains(encoded, []byte("klm-synthetic")) || bytes.Contains(encoded, []byte("encrypted")) {
		t.Fatal("snapshot contains credential data")
	}
}

func TestProjectGitIdentityStaleVerification(t *testing.T) {
	a, session := sqliteAutomationApp(t)
	mux := http.NewServeMux()
	base := "/api/projects/" + session.ProjectID + "/git-identities"
	var onVerify func()
	a.projectGitIdentityRoutes(mux, func(ctx context.Context, token string) (string, error) {
		if onVerify != nil {
			hook := onVerify
			onVerify = nil
			hook()
		}
		return fixtureGitVerifier(ctx, token)
	})
	list := gitIdentityList(t, gitIdentityRequest(t, mux, "POST", base, map[string]string{
		"alias": "work", "token": "klm-synthetic-github-alice",
	}, 201))
	path := base + "/" + list[0].ID
	onVerify = func() {
		gitIdentityRequest(t, mux, "PATCH", path, map[string]string{"token": "klm-synthetic-github-bob"}, 200)
	}
	gitIdentityRequest(t, mux, "PATCH", path, map[string]string{"token": "klm-synthetic-github-alice"}, 409)
	snapshot, err := a.captureProjectGitIdentities(a.ctx, session.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	token, err := a.resolveProjectGitIdentity(a.ctx, snapshot, "")
	if err != nil || token != "klm-synthetic-github-bob" {
		t.Fatal("stale verifier overwrote a newer credential", err)
	}
	body := gitIdentityRequest(t, mux, "PATCH", path, map[string]string{"token": "secret-in-transport-error"}, 503)
	if bytes.Contains(body, []byte("secret-in-transport-error")) || bytes.Contains(body, []byte("transport")) {
		t.Fatal("public response exposed verifier details")
	}
}

func TestProjectGitIdentityStartupMigration(t *testing.T) {
	a, session := sqliteAutomationApp(t)
	if _, err := a.store.writer.Exec(`DROP TABLE project_git_identities; PRAGMA user_version=6;`); err != nil {
		t.Fatal(err)
	}
	if err := a.store.enableRuntime(a.ctx); err != nil {
		t.Fatal("upgrade existing runtime", err)
	}
	mux := http.NewServeMux()
	a.projectGitIdentityRoutes(mux, fixtureGitVerifier)
	base := "/api/projects/" + session.ProjectID + "/git-identities"
	gitIdentityRequest(t, mux, "POST", base, map[string]string{
		"alias": "work", "token": "klm-synthetic-github-alice",
	}, 201)
	if err := a.store.enableRuntime(a.ctx); err != nil {
		t.Fatal("idempotent startup", err)
	}
	var version int
	if err := a.store.readers.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != 7 {
		t.Fatal("schema version", version, err)
	}
	var existing bool
	if err := a.store.readers.QueryRow(existingProtectedCredentialsSQL).Scan(&existing); err != nil || !existing {
		t.Fatal("Git identity missing from existing-key inventory", err)
	}
	if len(gitIdentityList(t, gitIdentityRequest(t, mux, "GET", base, nil, 200))) != 1 {
		t.Fatal("startup lost identities")
	}
	if err := a.store.readSnapshot(a.ctx, func(tx *sql.Tx) error {
		return requireGitIdentityProject(a.ctx, tx, session.ProjectID)
	}); err != nil {
		t.Fatal("migration damaged existing project", err)
	}
}

func TestProjectGitIdentityDefaultRemovalStableTie(t *testing.T) {
	a, session := sqliteAutomationApp(t)
	mux := http.NewServeMux()
	a.projectGitIdentityRoutes(mux, fixtureGitVerifier)
	base := "/api/projects/" + session.ProjectID + "/git-identities"
	for _, alias := range []string{"first", "second", "third"} {
		gitIdentityRequest(t, mux, "POST", base, map[string]any{
			"alias": alias, "token": "klm-synthetic-github-alice", "makeDefault": true,
		}, 201)
	}
	if _, err := a.store.writer.Exec(`UPDATE project_git_identities SET created_at=1`); err != nil {
		t.Fatal(err)
	}
	var defaultID, oldestID string
	if err := a.store.readers.QueryRow(`SELECT id FROM project_git_identities WHERE is_default=1`).Scan(&defaultID); err != nil {
		t.Fatal(err)
	}
	if err := a.store.readers.QueryRow(`SELECT id FROM project_git_identities WHERE id<>? ORDER BY id LIMIT 1`, defaultID).Scan(&oldestID); err != nil {
		t.Fatal(err)
	}
	list := gitIdentityList(t, gitIdentityRequest(t, mux, "DELETE", base+"/"+defaultID, nil, 200))
	defaults := 0
	for _, identity := range list {
		if identity.IsDefault {
			defaults++
			if identity.ID != oldestID {
				t.Fatal("oldest stable-ID tie did not become default")
			}
		}
	}
	if defaults != 1 {
		t.Fatal("removal left an invalid default count", defaults)
	}
}

type gitIdentityRoundTripper func(*http.Request) (*http.Response, error)

func (f gitIdentityRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestProjectGitIdentityVerifierBounds(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	for _, test := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{name: "account", status: 200, body: `{"login":"synthetic-alice"}`, want: "synthetic-alice"},
		{name: "unauthorized", status: 401, body: `secret-provider-details`},
		{name: "redirect", status: 302, body: `secret-provider-details`},
		{name: "oversized", status: 200, body: strings.Repeat("x", (64<<10)+1)},
		{name: "invalid-login", status: 200, body: `{"login":"unsafe\nlogin"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			http.DefaultTransport = gitIdentityRoundTripper(func(req *http.Request) (*http.Response, error) {
				calls++
				if req.URL.String() != "https://api.github.com/user" || req.Header.Get("Authorization") != "Bearer synthetic-token" {
					t.Fatal("verifier changed its fixed target or credential header")
				}
				deadline, ok := req.Context().Deadline()
				if !ok || time.Until(deadline) > 10*time.Second {
					t.Fatal("verifier request is not time-bounded")
				}
				return &http.Response{
					StatusCode: test.status,
					Header:     http.Header{"Location": []string{"https://redirect.invalid/user"}},
					Body:       io.NopCloser(strings.NewReader(test.body)),
					Request:    req,
				}, nil
			})
			login, err := verifyProjectGitToken(context.Background(), "synthetic-token")
			if test.want != "" && (err != nil || login != test.want) {
				t.Fatal("valid account verification failed", err)
			}
			if test.want == "" && err == nil {
				t.Fatal("invalid upstream result was accepted")
			}
			if calls != 1 {
				t.Fatal("verifier followed an upstream redirect")
			}
			if err != nil && strings.Contains(err.Error(), "secret-provider-details") {
				t.Fatal("verifier exposed upstream response bytes")
			}
		})
	}
}
