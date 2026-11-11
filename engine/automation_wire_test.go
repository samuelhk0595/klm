package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAutomationSaveJSONContract(t *testing.T) {
	for _, mode := range []string{"schedule", "webhook"} {
		for _, sessionMode := range []string{"new_per_run", "reuse_created", "existing"} {
			t.Run(mode+"/"+sessionMode, func(t *testing.T) {
				// Start with the complete response document that the editor copies,
				// including engine-owned metadata and string revision/sequence values.
				v := Automation{ID: "automation", ProjectID: "project", Revision: 9007199254740993,
					AfterSequence: 9007199254740995, CreationRequest: "fingerprint", OperationID: "operation",
					Name: "Task", Prompt: "Handle the task", FollowUpPrompt: "Continue the task",
					Mode: mode, SessionMode: sessionMode, Harness: "pi", TopicID: "topic",
					Schedule: AutomationSchedule{Recurrence: "daily", Time: "09:00", EveryDays: 2, Weekdays: []int{1}, MonthDay: 1},
					Webhook:  &AutomationWebhook{EventType: "issue_comment", Action: "created"}}
				data, err := json.Marshal(map[string]any{"automation": automationSQLView(v)})
				if err != nil {
					t.Fatal(err)
				}
				var body sqlSaveAutomationBody
				res := httptest.NewRecorder()
				if !decode(res, httptest.NewRequest("PATCH", "/", bytes.NewReader(data)), &body) {
					t.Fatalf("editor round-trip rejected: %s", res.Body.String())
				}
				if body.Automation.Revision != "9007199254740993" || body.Automation.AfterSequence != "9007199254740995" || *body.Automation.Webhook != *v.Webhook {
					t.Fatalf("editor values changed: %+v", body.Automation)
				}
			})
		}
	}
}

func TestAutomationSaveJSONErrors(t *testing.T) {
	for _, tc := range []struct {
		name, body, message string
	}{
		{"numeric revision", `{"automation":{"revision":1}}`, "Invalid JSON type for automation.revision: expected string."},
		{"numeric sequence", `{"automation":{"afterSequence":1}}`, "Invalid JSON type for automation.afterSequence: expected string."},
		{"presentation field", `{"automation":{"webhook":{"eventType":"issue_comment","action":"created","label":"private value"}}}`, `Unknown JSON request field "label".`},
		{"syntax", `{"automation":,}`, "Invalid JSON request body near byte 15."},
		{"empty", "", "A JSON request body is required."},
		{"too large", `{"automation":{"prompt":"` + strings.Repeat("x", 512<<10) + `"}}`, "JSON request body exceeds 512 KiB."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			var body sqlSaveAutomationBody
			if decode(res, httptest.NewRequest("POST", "/", strings.NewReader(tc.body)), &body) {
				t.Fatal("invalid request accepted")
			}
			var result map[string]string
			if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if res.Code != 400 || result["error"] != tc.message || strings.Contains(res.Body.String(), "private value") {
				t.Fatalf("unexpected diagnostic: %d %s", res.Code, res.Body.String())
			}
		})
	}
}
