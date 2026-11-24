package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRunnerIssueFieldsOrganizations(t *testing.T) {
	client := &http.Client{Transport: gitIdentityRoundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/user/memberships/orgs" || r.URL.Query().Get("state") != "active" || r.Header.Get("Authorization") != "Bearer fixture-token" {
			t.Fatal("unexpected memberships request")
		}
		header := http.Header{}
		login := "Acme"
		switch r.URL.Query().Get("page") {
		case "1":
			header.Set("Link", `<https://api.github.com/user/memberships/orgs?page=2>; rel="next"`)
			login = "Zebra"
		case "2":
		default:
			t.Fatal("unexpected memberships page")
		}
		return &http.Response{StatusCode: 200, Header: header, Body: io.NopCloser(strings.NewReader(`[{"state":"active","organization":{"login":"` + login + `"}}]`))}, nil
	})}
	orgs, err := loadRunnerIssueFieldOrganizations(context.Background(), client, "fixture-token")
	if err != nil || strings.Join(orgs, ",") != "acme,zebra" {
		t.Fatalf("memberships: %v (%v)", orgs, err)
	}
}

func TestRunnerIssueFieldsCatalog(t *testing.T) {
	state, issueStatus, calls := "active", 200, 0
	client := &http.Client{Transport: gitIdentityRoundTripper(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "api.github.com" || r.URL.Scheme != "https" || r.Method != "GET" || r.Header.Get("Authorization") != "Bearer fixture-token" {
			t.Fatal("unexpected GitHub catalog request")
		}
		switch r.URL.Path {
		case "/user/memberships/orgs/acme":
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"organization":{"login":"Acme"},"state":"` + state + `"}`))}, nil
		case "/orgs/acme/issue-fields":
			return &http.Response{StatusCode: issueStatus, Body: io.NopCloser(strings.NewReader(`[{"id":1,"data_type":"text"},{"id":2,"name":"Progress","data_type":"single_select","options":[{"id":9007199254740992,"name":"Ready"},{"id":9007199254740993,"name":"Blocked"}]}]`))}, nil
		default:
			t.Fatal("unexpected GitHub catalog path")
			return nil, nil
		}
	})}
	catalog, err := loadRunnerIssueFields(context.Background(), client, "acme", "fixture-token")
	if err != nil || len(catalog.Fields) != 1 || len(catalog.Fields[0].Options) != 2 || catalog.Fields[0].Options[0].ID != "9007199254740992" || catalog.Fields[0].Options[1].ID != "9007199254740993" {
		t.Fatalf("normalized catalog: %+v (%v)", catalog, err)
	}
	state, calls = "pending", 0
	if _, err := loadRunnerIssueFields(context.Background(), client, "acme", "fixture-token"); err == nil || calls != 1 {
		t.Fatal("pending membership reached organization catalog")
	}
	state, issueStatus = "active", 403
	if _, err := loadRunnerIssueFields(context.Background(), client, "acme", "fixture-token"); err == nil || !strings.Contains(err.Error(), "permissions and SSO") {
		t.Fatalf("permission diagnostic: %v", err)
	}
}
