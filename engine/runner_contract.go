package main

import "encoding/json"

const runnerOrigin = "external_runner"

type runnerCapture struct {
	ExternalRevision  string              `json:"external_revision,omitempty"`
	ConfigurationGate bool                `json:"configuration_gate,omitempty"`
	Harness           string              `json:"harness"`
	Model             string              `json:"model"`
	Effort            string              `json:"effort"`
	Instructions      string              `json:"instructions"`
	Selection         CapabilitySelection `json:"selection"`
	Manifest          []map[string]any    `json:"manifest"`
	YOLO              bool                `json:"yolo"`
	Capabilities      *CapabilitySnapshot `json:"capabilities,omitempty"`
}

type runnerExecution struct {
	RecentConversation []runnerConversationEntry `json:"recent_conversation,omitempty"`
	NativeWaiting      bool                      `json:"native_waiting"`
	ID                 string                    `json:"id"`
	IntegrationID      string                    `json:"integration_id"`
	ProjectID          string                    `json:"project_id"`
	Key                string                    `json:"key"`
	Kind               string                    `json:"kind"`
	SessionID          string                    `json:"session_id"`
	Generation         string                    `json:"generation"`
	TurnID             string                    `json:"turn_id"`
	Revision           string                    `json:"revision"`
	State              string                    `json:"state"`
	Desired            runnerCapture             `json:"desired"`
	Applied            runnerCapture             `json:"applied"`
	Error              string                    `json:"error,omitempty"`
	Source             *AgentCapture             `json:"source,omitempty"`
}

type runnerConversationEntry struct {
	SessionID  string `json:"source_session_id"`
	EventID    string `json:"event_id"`
	Ordinal    string `json:"ordinal"`
	Kind       string `json:"kind"`
	SourceKind string `json:"source_kind"`
	SourceID   string `json:"source_id"`
	Text       string `json:"text_preview"`
}

type runnerOperation struct {
	ID     string          `json:"operation_id"`
	Digest string          `json:"digest"`
	State  string          `json:"state"`
	Result json.RawMessage `json:"result"`
}

type runnerInput struct {
	Mentions    []Mention `json:"mentions,omitempty"`
	prepared    *submission
	ID          string `json:"id"`
	OperationID string `json:"operation_id"`
	SourceID    string `json:"source_id"`
	SourceKind  string `json:"source_kind"`
	Mode        string `json:"mode"`
	Text        string `json:"text"`
	State       string `json:"state"`
	TurnID      string `json:"turn_id"`
}

type runnerEvent struct {
	ID          string          `json:"id"`
	Cursor      string          `json:"cursor"`
	ExecutionID string          `json:"execution_id"`
	Type        string          `json:"type"`
	At          string          `json:"at"`
	Data        json.RawMessage `json:"data"`
}

type runnerToolCall struct {
	IsError      bool            `json:"is_error,omitempty"`
	ID           string          `json:"id"`
	ExecutionID  string          `json:"execution_id"`
	TurnID       string          `json:"turn_id"`
	NativeCallID string          `json:"native_call_id"`
	Name         string          `json:"name"`
	Arguments    json.RawMessage `json:"arguments"`
	State        string          `json:"state"`
	Reply        json.RawMessage `json:"reply,omitempty"`
}

func externalSession(s Session) bool { return s.Origin == runnerOrigin }
