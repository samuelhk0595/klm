package main

import (
	"strings"
	"testing"
)

func TestAutomationWebhookRouting(t *testing.T) {
	triggers := []AutomationWebhook{
		{EventType: "pull_request", Action: "opened"},
		{EventType: "pull_request", Action: "ready_for_review"},
		{EventType: "issues", Action: "opened"},
		{EventType: "issues", Action: "closed"},
		{EventType: "issues", Action: "deleted"},
		{EventType: "issues", Action: "edited"},
		{EventType: "issues", Action: "assigned"},
		{EventType: "issues", Action: "milestoned"},
		{EventType: "issue_comment", Action: "created"},
		{EventType: "issue_comment", Action: "edited"},
		{EventType: "issue_comment", Action: "deleted"},
	}
	for _, trigger := range triggers {
		t.Run(trigger.EventType+"/"+trigger.Action, func(t *testing.T) {
			d := diskState{}
			for i, config := range triggers {
				a := Automation{ID: config.EventType + "/" + config.Action, ProjectID: "p", Mode: "webhook", Enabled: true, Webhook: &triggers[i]}
				if i == 0 {
					a.Webhook = nil // Existing records retain opened, not all events.
				}
				d.Automations = append(d.Automations, a)
			}
			e := WebhookStoredEvent{ID: "event", ProjectID: "p", Provider: "github", EventType: trigger.EventType, Action: trigger.Action, ReceivedSequence: 2,
				Payload: `{"number":42,"issue":{"number":17},"repository":{"name":"repo","owner":{"type":"Organization","login":"org"}}}`}
			automationWebhook(&d, e)
			automationWebhook(&d, e) // Repeated delivery must not enqueue twice.
			if len(d.AutomationTriggers) != 1 || d.AutomationTriggers[0].AutomationID != trigger.EventType+"/"+trigger.Action {
				t.Fatalf("event routed to wrong automations: %+v", d.AutomationTriggers)
			}
			vars := d.AutomationTriggers[0].Variables
			if vars["organization.name"] != "org" || vars["repository.name"] != "repo" {
				t.Fatalf("missing repository context: %v", vars)
			}
			if trigger.EventType == "issues" || trigger.EventType == "issue_comment" {
				if vars["issue.number"] != "17" || vars["pr.number"] != "" {
					t.Fatalf("issue number must come from issue.number: %v", vars)
				}
			} else if vars["pr.number"] != "42" || vars["issue.number"] != "" {
				t.Fatalf("incorrect PR references: %v", vars)
			}
			d.AutomationTriggers = nil
			e.ProjectID = "other"
			automationWebhook(&d, e)
			e.ProjectID, e.ReceivedSequence = "p", 0
			automationWebhook(&d, e)
			if len(d.AutomationTriggers) != 0 {
				t.Fatal("wrong-project or pre-boundary event was admitted")
			}
		})
	}
}

func TestAutomationWebhookConfiguration(t *testing.T) {
	v := Automation{Mode: "webhook", Harness: "pi", Name: "Task", Prompt: "Handle issue {{issue.number}}", Webhook: &AutomationWebhook{EventType: "issues", Action: "opened"}}
	if err := validateAutomationConfig(&v); err != nil {
		t.Fatal(err)
	}
	v.Prompt = "Handle PR {{pr.number}}"
	if err := validateAutomationConfig(&v); err == nil {
		t.Fatal("issue trigger accepted an unavailable PR reference")
	}
	v.Webhook = nil
	if err := validateAutomationConfig(&v); err != nil {
		t.Fatal("legacy PR configuration rejected:", err)
	}
	v.Webhook = &AutomationWebhook{EventType: "pull_request", Action: "reopened"}
	if err := validateAutomationConfig(&v); err == nil {
		t.Fatal("unsupported action accepted")
	}
	v.Webhook = &AutomationWebhook{EventType: "issue_comment", Action: "created"}
	v.Prompt = "Handle comment on issue or PR {{issue.number}}"
	if err := validateAutomationConfig(&v); err != nil {
		t.Fatal("comment configuration rejected:", err)
	}
	v.SessionMode = "existing"
	v.SessionID = "session"
	v.FollowUpPrompt = "Handle PR {{pr.number}}"
	if err := validateAutomationConfig(&v); err == nil {
		t.Fatal("comment follow-up accepted an unavailable PR reference")
	}
	v.FollowUpPrompt = v.Prompt
	if err := validateAutomationConfig(&v); err != nil {
		t.Fatal("comment follow-up rejected:", err)
	}
}

func TestAutomationWebhookCommentTargets(t *testing.T) {
	for _, action := range []string{"created", "edited", "deleted"} {
		for _, target := range []string{"issue", "pull_request"} {
			t.Run(action+"/"+target, func(t *testing.T) {
				marker := ""
				if target == "pull_request" {
					marker = `,"pull_request":{"url":"https://api.github.com/repos/org/repo/pulls/17"}`
				}
				e := WebhookStoredEvent{ID: "comment", ProjectID: "p", Provider: "github", EventType: "issue_comment", Action: action, ReceivedSequence: 2,
					Payload: `{"number":99,"issue":{"number":17` + marker + `},"comment":{"body":"external content"},"repository":{"name":"repo"}}`}
				d := diskState{Automations: []Automation{{ID: "comments", ProjectID: "p", Mode: "webhook", Enabled: true, Webhook: &AutomationWebhook{EventType: "issue_comment", Action: action}}}}
				automationWebhook(&d, e)
				if len(d.AutomationTriggers) != 1 {
					t.Fatalf("comment did not enqueue: %+v", d.AutomationTriggers)
				}
				vars := d.AutomationTriggers[0].Variables
				if len(vars) != 3 || vars["issue.number"] != "17" || vars["organization.name"] != "" || vars["repository.name"] != "repo" {
					t.Fatalf("unexpected comment references: %v", vars)
				}
				d.AutomationTriggers = nil
				e.EventType = "pull_request_review_comment"
				automationWebhook(&d, e)
				if len(d.AutomationTriggers) != 0 {
					t.Fatal("inline review comment triggered a general-comment automation")
				}
			})
		}
	}
	e := WebhookStoredEvent{Provider: "github", EventType: "issue_comment", Action: "created", Payload: `{"number":99,"repository":{"name":"repo"}}`}
	if automationWebhookVariables(e) != nil {
		t.Fatal("comment without issue.number used a top-level number")
	}
	e.Action = "opened"
	if supportedAutomationWebhook(e.EventType, e.Action) {
		t.Fatal("unsupported comment action accepted")
	}
}

func TestAutomationWebhookPendingTriggerEdit(t *testing.T) {
	previous := Automation{ID: "review", Mode: "webhook", Enabled: true}
	d := diskState{AutomationTriggers: []AutomationTrigger{{ID: "queued-pr", AutomationID: previous.ID, Source: "webhook", Variables: map[string]string{"pr.number": "42"}}}}
	for _, config := range []AutomationWebhook{
		{EventType: "issues", Action: "opened"},
		{EventType: "pull_request", Action: "ready_for_review"},
		{EventType: "issue_comment", Action: "created"},
	} {
		next := previous
		next.Webhook = &config
		if err := d.validateAutomationTriggerEdit(previous, next); err == nil {
			t.Fatalf("queued event allowed trigger change to %+v", config)
		}
		if err := d.validateAutomationTriggerEdit(next, previous); err == nil {
			t.Fatal("reverse trigger change allowed with queued work")
		}
	}
	next := previous
	next.Webhook = &AutomationWebhook{EventType: "issues", Action: "opened"}
	d.AutomationTriggers[0].Consumed = true
	if err := d.validateAutomationTriggerEdit(previous, next); err != nil {
		t.Fatal("consumed work prevented trigger edit:", err)
	}
	d.AutomationTriggers[0].Consumed = false
	d.AutomationTriggers[0].AutomationID = "other"
	if err := d.validateAutomationTriggerEdit(previous, next); err != nil {
		t.Fatal("another automation's queue prevented edit:", err)
	}
	d.AutomationTriggers[0].AutomationID = previous.ID
	next = previous
	next.Mode = "schedule"
	if err := d.validateAutomationTriggerEdit(previous, next); err == nil {
		t.Fatal("activation change reinterpreted queued webhook work")
	}
	next = previous
	next.Enabled = false
	next.Prompt = "Updated task"
	next.Webhook = &AutomationWebhook{EventType: "pull_request", Action: "opened"}
	if err := d.validateAutomationTriggerEdit(previous, next); err != nil {
		t.Fatal("pause/prompt edit or explicit legacy default was blocked:", err)
	}
}

func TestAutomationWebhookRunReferences(t *testing.T) {
	prVariables := map[string]string{"pr.number": "42", "organization.name": "", "repository.name": "repo"}
	issueVariables := map[string]string{"issue.number": "17", "organization.name": "org", "repository.name": "repo"}
	for _, tc := range []struct {
		prompt    string
		variables map[string]string
		missing   string
	}{
		{"Handle PR {{pr.number}} in {{organization.name}}/{{repository.name}}", prVariables, ""},
		{"Handle issue {{issue.number}}", issueVariables, ""},
		{"Handle issue {{issue.number}}", prVariables, "{{issue.number}}"},
		{"Handle PR {{pr.number}}", issueVariables, "{{pr.number}}"},
		{"Handle issue {{issue.number}}", nil, "{{issue.number}}"},
	} {
		err := validateAutomationRunPrompt(tc.prompt, "webhook", tc.variables)
		if tc.missing == "" && err != nil {
			t.Fatalf("valid references rejected: %v", err)
		}
		if tc.missing != "" && (err == nil || !strings.Contains(err.Error(), tc.missing)) {
			t.Fatalf("missing reference %s not reported: %v", tc.missing, err)
		}
	}
}
