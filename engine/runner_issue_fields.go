package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type runnerIssueFieldOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type runnerIssueField struct {
	ID      string                   `json:"id"`
	Name    string                   `json:"name"`
	Options []runnerIssueFieldOption `json:"options"`
}

type runnerIssueFieldCatalog struct {
	Organization string             `json:"organization"`
	Fields       []runnerIssueField `json:"fields"`
}

func (a *app) runnerIssueFieldsHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if clientIdentity(r) == "" {
		fail(w, 403, "A paired runner integration is required.")
		return
	}
	if _, err := a.runnerOwner(r); err != nil {
		sqlHTTPError(w, err)
		return
	}
	listing := strings.HasSuffix(r.URL.Path, "/issue-field-organizations")
	values := r.URL.Query()
	org := strings.ToLower(values.Get("organization"))
	if (listing && r.URL.RawQuery != "") || (!listing && (len(values) != 1 || len(values["organization"]) != 1 || !validIssueFieldsOrganization(org))) {
		fail(w, 400, "Select one GitHub organization before loading Issue fields.")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	project, err := a.store.readProjectMeta(ctx, r.PathValue("projectId"))
	if err != nil || project.Removed {
		fail(w, 409, "The linked KLM project is unavailable. Check the Factory's project binding.")
		return
	}
	snapshot, err := a.captureProjectGitIdentities(ctx, project.ID)
	if err != nil || !snapshot.Configured() {
		fail(w, 422, "No usable default Git identity. Check Project settings → Git identities in KLM.")
		return
	}
	token, err := a.resolveProjectGitIdentity(ctx, snapshot, "")
	if err != nil {
		fail(w, 422, "Cannot use the default project Git identity. Check it in KLM and retry.")
		return
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	if listing {
		organizations, err := loadRunnerIssueFieldOrganizations(ctx, client, token)
		if err != nil {
			sqlHTTPError(w, err)
			return
		}
		respond(w, 200, map[string][]string{"organizations": organizations})
		return
	}
	catalog, err := loadRunnerIssueFields(ctx, client, org, token)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 200, catalog)
}

func validIssueFieldsOrganization(org string) bool {
	return validGitHubIdentityLogin(org) && !strings.HasPrefix(org, "-") && !strings.HasSuffix(org, "-") && !strings.Contains(org, "--")
}

type runnerIssueFieldMembership struct {
	State        string
	Organization struct{ Login string }
}

func loadRunnerIssueFieldOrganizations(ctx context.Context, client *http.Client, token string) ([]string, error) {
	orgs := []string{}
	seen := map[string]bool{}
	for page := 1; page <= 100; page++ {
		var next bool
		body, err := runnerIssueFieldsGitHub(ctx, client, "/user/memberships/orgs?state=active&per_page=100&page="+strconv.Itoa(page), token, 2<<20, &next)
		if err != nil {
			return nil, err
		}
		var rows []runnerIssueFieldMembership
		if json.Unmarshal(body, &rows) != nil || rows == nil {
			return nil, controlError(502, "GitHub returned an invalid organization list.")
		}
		for _, row := range rows {
			org := strings.ToLower(row.Organization.Login)
			if row.State != "active" || !validIssueFieldsOrganization(org) {
				return nil, controlError(502, "GitHub returned invalid organization membership data.")
			}
			if !seen[org] {
				seen[org] = true
				orgs = append(orgs, org)
			}
		}
		if !next {
			sort.Strings(orgs)
			return orgs, nil
		}
	}
	return nil, controlError(502, "The GitHub organization list exceeded the pagination limit.")
}

func loadRunnerIssueFields(ctx context.Context, client *http.Client, org, token string) (runnerIssueFieldCatalog, error) {
	result := runnerIssueFieldCatalog{Organization: org, Fields: []runnerIssueField{}}
	malformed := controlError(502, "GitHub returned an incomplete or invalid Issue field catalog. Retry or check the organization in GitHub.")
	ownerBody, err := runnerIssueFieldsGitHub(ctx, client, "/user/memberships/orgs/"+org, token, 64<<10, nil)
	if err != nil {
		return result, err
	}
	var membership runnerIssueFieldMembership
	if json.Unmarshal(ownerBody, &membership) != nil || membership.State != "active" || !strings.EqualFold(membership.Organization.Login, org) {
		return result, controlError(422, "The default project Git identity is not an active member of this organization.")
	}
	body, err := runnerIssueFieldsGitHub(ctx, client, "/orgs/"+org+"/issue-fields", token, 4<<20, nil)
	if err != nil {
		return result, err
	}
	if !utf8.Valid(body) {
		return result, malformed
	}
	// RawMessage avoids float64 coercion for both field and option IDs.
	var fields []struct {
		ID       json.RawMessage
		Name     string
		DataType string `json:"data_type"`
		Options  []struct {
			ID   json.RawMessage
			Name string
		}
	}
	if json.Unmarshal(body, &fields) != nil || fields == nil {
		return result, malformed
	}
	seen := map[string]bool{}
	for _, field := range fields {
		id, ok := runnerIssueFieldID(field.ID)
		if !ok || seen[id] || field.DataType == "" {
			return result, malformed
		}
		seen[id] = true
		if field.DataType != "single_select" {
			continue
		}
		if !runnerIssueFieldLabel(field.Name) || field.Options == nil {
			return result, malformed
		}
		entry := runnerIssueField{ID: id, Name: field.Name, Options: []runnerIssueFieldOption{}}
		options := map[string]bool{}
		for _, option := range field.Options {
			id, ok := runnerIssueFieldID(option.ID)
			if !ok || options[id] || !runnerIssueFieldLabel(option.Name) {
				return result, malformed
			}
			options[id] = true
			entry.Options = append(entry.Options, runnerIssueFieldOption{ID: id, Name: option.Name})
		}
		result.Fields = append(result.Fields, entry)
	}
	return result, nil
}

func runnerIssueFieldsGitHub(ctx context.Context, client *http.Client, path, token string, limit int64, next *bool) ([]byte, error) {
	unavailable := controlError(502, "GitHub Issue fields are unavailable. Check the KLM engine's GitHub connectivity and retry.")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com"+path, nil)
	if err != nil {
		return nil, unavailable
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "KLM")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	res, err := client.Do(req)
	if err != nil {
		return nil, unavailable
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusUnauthorized:
		return nil, controlError(422, "GitHub rejected the default project Git identity. Check its credential in KLM and retry.")
	case http.StatusForbidden, http.StatusNotFound:
		return nil, controlError(422, "GitHub denied organization discovery or Issue field access. Check the default identity's organization permissions and SSO (OAuth/classic: read:org).")
	case http.StatusTooManyRequests:
		return nil, controlError(429, "GitHub rate-limited Issue field discovery. Retry later.")
	case http.StatusOK:
	default:
		return nil, unavailable
	}
	if next != nil {
		*next = strings.Contains(strings.Join(res.Header.Values("Link"), ","), `rel="next"`)
	} else if len(res.Header.Values("Link")) != 0 {
		return nil, controlError(502, "GitHub returned a continued Issue field catalog; partial catalogs are unavailable.")
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil || int64(len(body)) > limit {
		return nil, unavailable
	}
	return body, nil
}

func runnerIssueFieldID(raw json.RawMessage) (string, bool) {
	id := string(raw)
	if len(id) > 0 && id[0] == '"' && json.Unmarshal(raw, &id) != nil {
		return "", false
	}
	if len(id) == 0 || len(id) > 20 || id[0] < '1' || id[0] > '9' {
		return "", false
	}
	for i := range len(id) {
		if id[i] < '0' || id[i] > '9' {
			return "", false
		}
	}
	_, err := strconv.ParseUint(id, 10, 64)
	return id, err == nil
}

func runnerIssueFieldLabel(s string) bool {
	return s != "" && utf8.RuneCountInString(s) <= 256 && !strings.ContainsFunc(s, unicode.IsControl)
}
