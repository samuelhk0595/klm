package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type graphInvokeArgs struct {
	OperationID   string                 `json:"operationId"`
	GraphID       string                 `json:"graphId"`
	Objective     string                 `json:"objective"`
	Task          string                 `json:"task"`
	Authorization []GraphAuthorization   `json:"authorization"`
	Workspace     GraphWorkspaceDecision `json:"workspace"`
	Dependencies  []string               `json:"dependencies"`
	Priority      int                    `json:"priority"`
}

type graphActivityUpdateArgs struct {
	ActivityID      string                  `json:"activityId"`
	OperationID     string                  `json:"operationId"`
	ExpectedVersion uint64                  `json:"expectedVersion"`
	Action          string                  `json:"action"`
	Task            string                  `json:"task"`
	Correction      string                  `json:"correction"`
	Workspace       *GraphWorkspaceDecision `json:"workspace"`
	UserEventID     string                  `json:"userEventId"`
	Priority        int                     `json:"priority"`
	Dependencies    []string                `json:"dependencies"`
}

func graphAuthorizationSchema() map[string]any {
	return map[string]any{
		"type": "array", "minItems": 1,
		"description": "References to real user messages authorizing this graph/activity in the owning main conversation. This records activity scope, not semantic proof of consent or harness/tool permission grants.",
		"items": map[string]any{
			"type": "object", "additionalProperties": false, "required": []string{"eventId", "text"},
			"properties": map[string]any{
				"eventId": map[string]any{"type": "string", "minLength": 1, "description": "Exact eventId of the authorizing user message from recentUserMessages or retrieved main history. Never invent an ID."},
				"text":    map[string]any{"type": "string", "minLength": 1, "description": "Nonblank description of the activity scope authorized by that user message; not a workspace-default justification or a tool grant."},
			},
		},
	}
}

func graphWorkspaceSchema() map[string]any {
	return map[string]any{
		"type": "object", "additionalProperties": false, "required": []string{},
		"description": "Invocation workspace policy, never an arbitrary path. On invoke, omission of workspace or mode means original: the registered project folder with its existing files. Do not ask for confirmation or justify this product default. Retry still needs the user's explicit fresh/reuse decision when a prior attempt exists. On non-retry updates, omission preserves the saved workspace.",
		"properties": map[string]any{
			"mode":                 map[string]any{"type": "string", "enum": []string{"original", "new_worktree"}, "default": "original", "description": "original uses Project.Folder as-is; new_worktree explicitly requests isolation. Omission defaults to original, not to a previous attempt's mode. Explicit invalid values are rejected."},
			"attempt":              map[string]any{"type": "string", "enum": []string{"fresh", "reuse"}, "description": "Required for retry after a prior run: the user's fresh/reuse decision. No automatic default. fresh preserves existing files/worktrees; reuse needs sourceRunId and reuse.initial."},
			"authorizationEventId": map[string]any{"type": "string", "description": "Optional legacy reference to a user workspace decision in this conversation. Not required for any mode, and never needed to justify original. If nonempty, it must be a real user event; empty legacy values are accepted."},
			"baseRevision":         map[string]any{"type": "string", "description": "Explicitly user-requested local revision for new_worktree. Omitted/empty uses the source project folder's local HEAD at creation. A retry retains a previously requested revision; no reset or copy of dirty files."},
			"sourceRunId":          map[string]any{"type": "string", "minLength": 1, "description": "Required for reuse: terminal source run in this conversation, whose recorded workspace IDs supply the associations."},
			"reuse": map[string]any{
				"type": "object", "required": []string{}, "additionalProperties": map[string]any{"type": "string", "minLength": 1},
				"description": "Explicit association-to-workspace-ID map from graph_get_run, only with attempt=reuse. Must include initial; other keys are node:<nodeId>:<occurrence> for Join or node:<nodeId>:<occurrence>:branch:<branchId> for Fork. Values are recorded workspace IDs, never paths. Do not infer by names or recency.",
			},
		},
	}
}

// These checks cover the two activity tools' nested wire contract. Keep the Go
// decoder strict as well; it alone cannot report unknown fields with their path,
// and treats explicit null/empty enum strings like omitted values.
func checkGraphActivityArgumentShape(raw json.RawMessage, invoke bool) error {
	fields := []string{"activityId", "operationId", "expectedVersion", "action", "task", "correction", "workspace", "userEventId", "priority", "dependencies"}
	if invoke {
		fields = []string{"operationId", "graphId", "objective", "task", "authorization", "workspace", "dependencies", "priority"}
	}
	args, err := graphArgumentObject(raw, "arguments", fields)
	if err != nil {
		return err
	}
	if auth, ok := args["authorization"]; ok {
		var entries []json.RawMessage
		if json.Unmarshal(auth, &entries) != nil || entries == nil {
			return errors.New("authorization: expected a nonempty array of {eventId: string, text: string} objects.")
		}
		for i, entry := range entries {
			path := fmt.Sprintf("authorization[%d]", i)
			item, err := graphArgumentObject(entry, path, []string{"eventId", "text"})
			if err != nil {
				return err
			}
			for _, key := range []string{"eventId", "text"} {
				if _, err := graphArgumentString(item[key], path+"."+key, true); err != nil {
					return err
				}
			}
		}
	}
	if workspace, ok := args["workspace"]; ok {
		item, err := graphArgumentObject(workspace, "workspace", []string{"mode", "attempt", "authorizationEventId", "baseRevision", "sourceRunId", "reuse"})
		if err != nil {
			return err
		}
		for _, key := range []string{"mode", "attempt", "authorizationEventId", "baseRevision", "sourceRunId"} {
			value, exists := item[key]
			if !exists {
				continue
			}
			// Empty legacy authorizationEventId remains accepted for compatibility.
			text, err := graphArgumentString(value, "workspace."+key, key == "sourceRunId")
			if err != nil {
				return err
			}
			if key == "mode" && text != "original" && text != "new_worktree" {
				return errors.New("workspace.mode: expected original or new_worktree; omit mode to use original.")
			}
			if key == "attempt" && text != "fresh" && text != "reuse" {
				return errors.New("workspace.attempt: expected fresh or reuse; retry requires the user's explicit decision.")
			}
		}
		if reuse, exists := item["reuse"]; exists {
			associations, err := graphArgumentObject(reuse, "workspace.reuse", nil)
			if err != nil {
				return err
			}
			for _, key := range sortedGraphKeys(associations) {
				if _, err := graphArgumentString(associations[key], "workspace.reuse."+key, true); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func graphArgumentObject(raw json.RawMessage, path string, fields []string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || object == nil {
		return nil, fmt.Errorf("%s: expected an object; accepted fields: %s.", path, graphArgumentFields(fields))
	}
	if fields != nil {
		for _, key := range sortedGraphKeys(object) {
			if !slicesContainsString(fields, key) {
				prefix := path + "."
				if path == "arguments" {
					prefix = ""
				}
				return nil, fmt.Errorf("%s%s: unknown field; %s accepts {%s}.", prefix, key, path, graphArgumentFields(fields))
			}
		}
	}
	return object, nil
}

func graphArgumentFields(fields []string) string {
	if fields == nil {
		return "association keys mapped to workspace ID strings"
	}
	return strings.Join(fields, ", ")
}

func graphArgumentString(raw json.RawMessage, path string, nonblank bool) (string, error) {
	var text *string
	if json.Unmarshal(raw, &text) != nil || text == nil {
		return "", fmt.Errorf("%s: expected a string.", path)
	}
	if nonblank && strings.TrimSpace(*text) == "" {
		return "", fmt.Errorf("%s: provide a nonblank string.", path)
	}
	return *text, nil
}

func normalizeGraphWorkspaceDecision(decision *GraphWorkspaceDecision) {
	if decision.Mode == "" {
		decision.Mode = "original"
	}
}

// Add only the missing default on legacy v2 records. Preserve decisions,
// authorization references, workspace IDs and creation/reuse provenance.
func normalizeGraphWorkspaceRecords(d *diskState) {
	for i := range d.GraphActivities {
		normalizeGraphWorkspaceDecision(&d.GraphActivities[i].Workspace)
	}
	for i := range d.GraphRuns {
		normalizeGraphWorkspaceDecision(&d.GraphRuns[i].Workspace)
	}
}

func validateGraphAuthorization(d *diskState, owner string, authorization []GraphAuthorization) error {
	if len(authorization) == 0 {
		return errors.New("authorization: provide at least one {eventId, text} reference to the user message authorizing this graph/activity; graph selection or the original-workspace default is not activity authorization.")
	}
	for i, auth := range authorization {
		if !graphUserEvent(d, owner, auth.EventID) {
			return fmt.Errorf("authorization[%d].eventId: must reference a real user message in this main conversation authorizing the activity; use recentUserMessages or retrieve main history, never invent an ID.", i)
		}
		if strings.TrimSpace(auth.Text) == "" {
			return fmt.Errorf("authorization[%d].text: describe the activity scope authorized by the referenced user message (nonblank string).", i)
		}
	}
	return nil
}

func graphActivityHasRun(d *diskState, activityID string) bool {
	for _, run := range d.GraphRuns {
		if run.ActivityID == activityID {
			return true
		}
	}
	return false
}
