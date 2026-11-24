package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

const projectGitIdentitySchema = `
CREATE TABLE IF NOT EXISTS project_git_identities(
 id TEXT PRIMARY KEY,
 project_id TEXT NOT NULL REFERENCES data_projects(id),
 alias TEXT NOT NULL COLLATE NOCASE,
 login TEXT NOT NULL,
 is_default INTEGER NOT NULL CHECK(is_default IN (0,1)),
 credential_version TEXT NOT NULL,
 row_version TEXT NOT NULL,
 created_at INTEGER NOT NULL,
 encrypted_token BLOB NOT NULL,
 UNIQUE(project_id,alias)
) STRICT;
CREATE UNIQUE INDEX IF NOT EXISTS project_git_identity_default
 ON project_git_identities(project_id) WHERE is_default=1;
CREATE INDEX IF NOT EXISTS project_git_identity_order
 ON project_git_identities(project_id,created_at,id);
`

// Every writer sharing the Linux encryption key must use the same inventory.
const existingProtectedCredentialsSQL = `SELECT EXISTS(SELECT 1 FROM data_mcp_credentials)
 OR EXISTS(SELECT 1 FROM runner_requests WHERE reply IS NOT NULL)
 OR EXISTS(SELECT 1 FROM project_git_identities)`

type GitIdentity struct {
	ID        string `json:"id"`
	Alias     string `json:"alias"`
	Login     string `json:"login"`
	IsDefault bool   `json:"isDefault"`
}

type ProjectGitIdentityReference struct {
	ID                string `json:"id"`
	Alias             string `json:"alias"`
	CredentialVersion string `json:"credentialVersion"`
}

// Nil is a missing legacy/general capture. A nonnil empty capture is frozen shared mode.
// This contract contains no credentials and may be persisted with its owning work.
type ProjectGitIdentitySnapshot struct {
	ProjectID  string                        `json:"projectId"`
	DefaultID  string                        `json:"defaultId"`
	Identities []ProjectGitIdentityReference `json:"identities"`
}

func (s *ProjectGitIdentitySnapshot) Configured() bool {
	return s != nil && len(s.Identities) != 0
}

type projectGitIdentityRecord struct {
	GitIdentity
	CredentialVersion string
	RowVersion        string
	EncryptedToken    []byte
}

func validProjectGitAlias(alias string) bool {
	if len(alias) < 1 || len(alias) > 40 {
		return false
	}
	for i := range len(alias) {
		c := alias[i]
		alphanumeric := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
		if !alphanumeric && (i == 0 || c != '_' && c != '-') {
			return false
		}
	}
	return true
}

func requireGitIdentityProject(ctx context.Context, q sqliteQuery, projectID string) error {
	var active bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS(
 SELECT 1 FROM data_projects WHERE id=? AND x_removed<>'true')`, projectID).Scan(&active)
	if err != nil {
		return err
	}
	if !active {
		return sqliteReject(404, "Project not found.")
	}
	return nil
}

func readProjectGitIdentity(ctx context.Context, q sqliteQuery, projectID, id string) (projectGitIdentityRecord, error) {
	var row projectGitIdentityRecord
	err := q.QueryRowContext(ctx, `SELECT id,alias,login,is_default,credential_version,row_version,encrypted_token
 FROM project_git_identities WHERE project_id=? AND id=?`, projectID, id).Scan(
		&row.ID, &row.Alias, &row.Login, &row.IsDefault,
		&row.CredentialVersion, &row.RowVersion, &row.EncryptedToken,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return row, sqliteReject(404, "Git identity not found.")
	}
	return row, err
}

func listProjectGitIdentityMetadata(ctx context.Context, q sqliteQuery, projectID string) ([]GitIdentity, error) {
	identities := []GitIdentity{}
	rows, err := q.QueryContext(ctx, `SELECT id,alias,login,is_default FROM project_git_identities
 WHERE project_id=? ORDER BY created_at,id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var identity GitIdentity
		if err := rows.Scan(&identity.ID, &identity.Alias, &identity.Login, &identity.IsDefault); err != nil {
			return nil, err
		}
		identities = append(identities, identity)
	}
	return identities, rows.Err()
}

func (a *app) captureProjectGitIdentities(ctx context.Context, projectID string) (*ProjectGitIdentitySnapshot, error) {
	if projectID == "" {
		return nil, nil
	}
	if a.store == nil {
		return nil, errors.New("Project Git identity storage is unavailable.")
	}
	snapshot := &ProjectGitIdentitySnapshot{ProjectID: projectID, Identities: []ProjectGitIdentityReference{}}
	err := a.store.readSnapshot(ctx, func(tx *sql.Tx) error {
		if err := requireGitIdentityProject(ctx, tx, projectID); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, `SELECT id,alias,credential_version,is_default
 FROM project_git_identities WHERE project_id=? ORDER BY created_at,id`, projectID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var ref ProjectGitIdentityReference
			var isDefault bool
			if err := rows.Scan(&ref.ID, &ref.Alias, &ref.CredentialVersion, &isDefault); err != nil {
				return err
			}
			snapshot.Identities = append(snapshot.Identities, ref)
			if isDefault {
				snapshot.DefaultID = ref.ID
			}
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if snapshot.Configured() && snapshot.DefaultID == "" {
		return nil, errors.New("Project Git identity default is unavailable.")
	}
	return snapshot, nil
}

func (a *app) resolveProjectGitIdentity(
	ctx context.Context,
	snapshot *ProjectGitIdentitySnapshot,
	alias string,
) (string, error) {
	if !snapshot.Configured() {
		return "", errors.New("No configured project Git identity was captured.")
	}
	var selected *ProjectGitIdentityReference
	for i := range snapshot.Identities {
		ref := &snapshot.Identities[i]
		matches := alias != "" && strings.EqualFold(alias, ref.Alias)
		if matches || alias == "" && ref.ID == snapshot.DefaultID {
			selected = ref
			break
		}
	}
	if selected == nil {
		return "", errors.New("Git identity is not in the captured project selection.")
	}
	if a.store == nil {
		return "", errors.New("Project Git identity storage is unavailable.")
	}
	var encrypted []byte
	err := a.store.readSnapshot(ctx, func(tx *sql.Tx) error {
		if err := requireGitIdentityProject(ctx, tx, snapshot.ProjectID); err != nil {
			return err
		}
		row, err := readProjectGitIdentity(ctx, tx, snapshot.ProjectID, selected.ID)
		if err != nil {
			return err
		}
		if row.CredentialVersion != selected.CredentialVersion {
			return errors.New("Captured Git identity credential was replaced; admit new work to use it.")
		}
		encrypted = row.EncryptedToken
		return nil
	})
	if err != nil {
		return "", err
	}
	return a.decryptProjectGitToken(encrypted)
}

func (a *app) decryptProjectGitToken(encrypted []byte) (string, error) {
	plain, err := unprotectMCPCredential(encrypted, mcpCredentialStorage{Directory: a.dir, Existing: true})
	if err != nil {
		return "", errors.New("Cannot decrypt project Git identity; restore the matching credential key.")
	}
	defer clear(plain)
	var envelope struct {
		Kind  string `json:"kind"`
		Token string `json:"token"`
	}
	if json.Unmarshal(plain, &envelope) != nil || envelope.Kind != "project-git-identity-v1" {
		return "", errors.New("Invalid project Git credential envelope.")
	}
	if !validProjectGitToken(envelope.Token) {
		return "", errors.New("Invalid stored project Git credential.")
	}
	return envelope.Token, nil
}
