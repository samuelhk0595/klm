package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type projectGitIdentityVerifier func(context.Context, string) (string, error)

func validProjectGitToken(token string) bool {
	if len(token) == 0 || len(token) > 4096 {
		return false
	}
	for i := range len(token) {
		if token[i] <= 32 || token[i] >= 127 {
			return false
		}
	}
	return true
}

// The URL is deliberately fixed. Tests inject a verifier into route registration,
// never an endpoint or transport chosen by runtime configuration.
func verifyProjectGitToken(ctx context.Context, token string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", sqliteReject(502, "GitHub account verification failed.")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "KLM")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return "", sqliteReject(502, "GitHub account verification failed.")
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return "", sqliteReject(400, "GitHub could not verify this token.")
	}
	if res.StatusCode != http.StatusOK {
		return "", sqliteReject(502, "GitHub account verification failed.")
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, (64<<10)+1))
	if err != nil || len(body) > 64<<10 {
		return "", sqliteReject(502, "GitHub account verification failed.")
	}
	var account struct {
		Login string `json:"login"`
	}
	if json.Unmarshal(body, &account) != nil || !validGitHubIdentityLogin(account.Login) {
		return "", sqliteReject(502, "GitHub account verification failed.")
	}
	return account.Login, nil
}

func validGitHubIdentityLogin(login string) bool {
	if len(login) == 0 || len(login) > 39 {
		return false
	}
	for i := range len(login) {
		c := login[i]
		letterOrDigit := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
		if !letterOrDigit && c != '-' {
			return false
		}
	}
	return true
}

func (a *app) projectGitIdentityRoutes(mux *http.ServeMux, verify projectGitIdentityVerifier) {
	mux.HandleFunc("GET /api/projects/{id}/git-identities", a.listProjectGitIdentities)
	mux.HandleFunc("POST /api/projects/{id}/git-identities/test", func(w http.ResponseWriter, r *http.Request) {
		a.testProjectGitIdentity(w, r, verify)
	})
	mux.HandleFunc("POST /api/projects/{id}/git-identities", func(w http.ResponseWriter, r *http.Request) {
		a.saveProjectGitIdentity(w, r, verify)
	})
	mux.HandleFunc("PATCH /api/projects/{id}/git-identities/{identityId}", func(w http.ResponseWriter, r *http.Request) {
		a.saveProjectGitIdentity(w, r, verify)
	})
	mux.HandleFunc("DELETE /api/projects/{id}/git-identities/{identityId}", a.deleteProjectGitIdentity)
}

func decodeProjectGitIdentity(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(target) != nil || d.Decode(&struct{}{}) != io.EOF {
		fail(w, 400, "Invalid Git identity request body.")
		return false
	}
	return true
}

func (a *app) readProjectGitIdentitySettings(ctx context.Context, projectID string) ([]GitIdentity, error) {
	if a.store == nil {
		return nil, sqliteReject(503, "Project Git identity storage is unavailable.")
	}
	identities := []GitIdentity{}
	err := a.store.readSnapshot(ctx, func(tx *sql.Tx) error {
		if err := requireGitIdentityProject(ctx, tx, projectID); err != nil {
			return err
		}
		var err error
		identities, err = listProjectGitIdentityMetadata(ctx, tx, projectID)
		return err
	})
	return identities, err
}

func respondProjectGitIdentities(w http.ResponseWriter, status int, identities []GitIdentity) {
	respond(w, status, struct {
		Identities []GitIdentity `json:"identities"`
	}{Identities: identities})
}

func gitIdentityHTTPError(w http.ResponseWriter, err error) {
	var request *sqliteRequestError
	if errors.As(err, &request) {
		fail(w, request.Status, request.Message)
		return
	}
	// Storage/DPAPI/transport errors never expose their implementation details.
	fail(w, 503, "Project Git identity is unavailable.")
}

func (a *app) listProjectGitIdentities(w http.ResponseWriter, r *http.Request) {
	identities, err := a.readProjectGitIdentitySettings(r.Context(), r.PathValue("id"))
	if err != nil {
		gitIdentityHTTPError(w, err)
		return
	}
	respondProjectGitIdentities(w, 200, identities)
}

func (a *app) existingProjectGitIdentity(ctx context.Context, projectID, id string) (projectGitIdentityRecord, error) {
	var row projectGitIdentityRecord
	if a.store == nil {
		return row, sqliteReject(503, "Project Git identity storage is unavailable.")
	}
	err := a.store.readSnapshot(ctx, func(tx *sql.Tx) error {
		if err := requireGitIdentityProject(ctx, tx, projectID); err != nil {
			return err
		}
		var err error
		row, err = readProjectGitIdentity(ctx, tx, projectID, id)
		return err
	})
	return row, err
}

func (a *app) testProjectGitIdentity(w http.ResponseWriter, r *http.Request, verify projectGitIdentityVerifier) {
	var body struct {
		Token      *string `json:"token"`
		IdentityID *string `json:"identityId"`
	}
	if !decodeProjectGitIdentity(w, r, &body) {
		return
	}
	if (body.Token == nil) == (body.IdentityID == nil) {
		fail(w, 400, "Supply either token or identityId.")
		return
	}
	projectID := r.PathValue("id")
	if _, err := a.readProjectGitIdentitySettings(r.Context(), projectID); err != nil {
		gitIdentityHTTPError(w, err)
		return
	}
	var token string
	if body.Token != nil {
		token = *body.Token
	}
	var saved projectGitIdentityRecord
	if body.IdentityID != nil {
		var err error
		saved, err = a.existingProjectGitIdentity(r.Context(), projectID, *body.IdentityID)
		if err == nil {
			token, err = a.decryptProjectGitToken(saved.EncryptedToken)
		}
		if err != nil {
			gitIdentityHTTPError(w, err)
			return
		}
	}
	if !validProjectGitToken(token) {
		fail(w, 400, "A token of at most 4096 non-whitespace ASCII characters is required.")
		return
	}
	login, err := verify(r.Context(), token)
	if err != nil {
		gitIdentityHTTPError(w, err)
		return
	}
	if body.IdentityID != nil {
		current, err := a.existingProjectGitIdentity(r.Context(), projectID, saved.ID)
		if err != nil {
			gitIdentityHTTPError(w, err)
			return
		}
		if current.RowVersion != saved.RowVersion {
			fail(w, 409, "Git identity changed during verification. Reload and try again.")
			return
		}
	} else if _, err := a.readProjectGitIdentitySettings(r.Context(), projectID); err != nil {
		gitIdentityHTTPError(w, err)
		return
	}
	respond(w, 200, map[string]string{"login": login})
}

type projectGitIdentityMutation struct {
	Alias       *string `json:"alias"`
	Token       *string `json:"token"`
	MakeDefault *bool   `json:"makeDefault"`
}

func (a *app) saveProjectGitIdentity(w http.ResponseWriter, r *http.Request, verify projectGitIdentityVerifier) {
	var body projectGitIdentityMutation
	if !decodeProjectGitIdentity(w, r, &body) {
		return
	}
	create := r.Method == http.MethodPost
	missingCreateFields := create && (body.Alias == nil || body.Token == nil)
	invalidAlias := body.Alias != nil && !validProjectGitAlias(*body.Alias)
	invalidToken := body.Token != nil && !validProjectGitToken(*body.Token)
	if missingCreateFields || invalidAlias || invalidToken {
		fail(w, 400, "Supply a valid alias and token. Aliases use 1–40 ASCII letters, digits, _ or -, starting with a letter or digit.")
		return
	}
	if body.Alias == nil && body.Token == nil && body.MakeDefault == nil {
		fail(w, 400, "Supply alias, token or makeDefault.")
		return
	}
	projectID := r.PathValue("id")
	if _, err := a.readProjectGitIdentitySettings(r.Context(), projectID); err != nil {
		gitIdentityHTTPError(w, err)
		return
	}
	row := projectGitIdentityRecord{}
	if !create {
		var err error
		row, err = a.existingProjectGitIdentity(r.Context(), projectID, r.PathValue("identityId"))
		if err != nil {
			gitIdentityHTTPError(w, err)
			return
		}
	}
	expectedVersion := row.RowVersion
	if create {
		row.ID = newID()
	}
	if body.Alias != nil {
		row.Alias = *body.Alias
	}
	if body.Token != nil {
		login, err := verify(r.Context(), *body.Token)
		if err != nil {
			gitIdentityHTTPError(w, err)
			return
		}
		row.Login = login
		row.CredentialVersion = newID()
		plain, _ := json.Marshal(struct {
			Kind  string `json:"kind"`
			Token string `json:"token"`
		}{Kind: "project-git-identity-v1", Token: *body.Token})
		var existing bool
		err = a.store.readers.QueryRowContext(r.Context(), existingProtectedCredentialsSQL).Scan(&existing)
		if err == nil {
			row.EncryptedToken, err = protectMCPCredential(plain, mcpCredentialStorage{Directory: a.dir, Existing: existing})
		}
		clear(plain)
		if err != nil {
			gitIdentityHTTPError(w, err)
			return
		}
	}
	identities := []GitIdentity{}
	a.mu.Lock()
	err := a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		if err := requireGitIdentityProject(st.ctx, st.tx, projectID); err != nil {
			return err
		}
		if !create {
			current, err := readProjectGitIdentity(st.ctx, st.tx, projectID, row.ID)
			if err != nil {
				return err
			}
			if current.RowVersion != expectedVersion {
				return sqliteReject(409, "Git identity changed during verification. Reload and try again.")
			}
		}
		var duplicate bool
		err := st.tx.QueryRowContext(st.ctx, `SELECT EXISTS(SELECT 1 FROM project_git_identities
 WHERE project_id=? AND alias=? COLLATE NOCASE AND id<>?)`, projectID, row.Alias, row.ID).Scan(&duplicate)
		if err != nil {
			return err
		}
		if duplicate {
			return sqliteReject(409, "A Git identity with this alias already exists.")
		}
		var count int
		if err := st.tx.QueryRowContext(st.ctx, `SELECT count(*) FROM project_git_identities WHERE project_id=?`, projectID).Scan(&count); err != nil {
			return err
		}
		makeDefault := body.MakeDefault != nil && *body.MakeDefault
		row.IsDefault = row.IsDefault || count == 0 || makeDefault
		if row.IsDefault {
			if _, err := st.tx.ExecContext(st.ctx, `UPDATE project_git_identities SET is_default=0,row_version=?
 WHERE project_id=? AND is_default=1 AND id<>?`, newID(), projectID, row.ID); err != nil {
				return err
			}
		}
		_, err = st.tx.ExecContext(st.ctx, `INSERT INTO project_git_identities(
 id,project_id,alias,login,is_default,credential_version,row_version,created_at,encrypted_token)
 VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET alias=excluded.alias,login=excluded.login,
 is_default=excluded.is_default,credential_version=excluded.credential_version,
 row_version=excluded.row_version,encrypted_token=excluded.encrypted_token`,
			row.ID, projectID, row.Alias, row.Login, row.IsDefault,
			row.CredentialVersion, newID(), time.Now().UnixNano(), row.EncryptedToken,
		)
		if err != nil {
			return err
		}
		st.changes.Projects[projectID] = true
		identities, err = listProjectGitIdentityMetadata(st.ctx, st.tx, projectID)
		return err
	})
	a.mu.Unlock()
	if err != nil {
		gitIdentityHTTPError(w, err)
		return
	}
	status := http.StatusOK
	if create {
		status = http.StatusCreated
	}
	respondProjectGitIdentities(w, status, identities)
}

func (a *app) deleteProjectGitIdentity(w http.ResponseWriter, r *http.Request) {
	if a.store == nil {
		fail(w, 503, "Project Git identity storage is unavailable.")
		return
	}
	projectID := r.PathValue("id")
	identities := []GitIdentity{}
	a.mu.Lock()
	err := a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		if err := requireGitIdentityProject(st.ctx, st.tx, projectID); err != nil {
			return err
		}
		row, err := readProjectGitIdentity(st.ctx, st.tx, projectID, r.PathValue("identityId"))
		if err != nil {
			return err
		}
		if _, err := st.tx.ExecContext(st.ctx, `DELETE FROM project_git_identities WHERE project_id=? AND id=?`, projectID, row.ID); err != nil {
			return err
		}
		if row.IsDefault {
			_, err = st.tx.ExecContext(st.ctx, `UPDATE project_git_identities SET is_default=1,row_version=?
 WHERE id=(SELECT id FROM project_git_identities WHERE project_id=? ORDER BY created_at,id LIMIT 1)`, newID(), projectID)
			if err != nil {
				return err
			}
		}
		st.changes.Projects[projectID] = true
		identities, err = listProjectGitIdentityMetadata(st.ctx, st.tx, projectID)
		return err
	})
	a.mu.Unlock()
	if err != nil {
		gitIdentityHTTPError(w, err)
		return
	}
	respondProjectGitIdentities(w, 200, identities)
}
