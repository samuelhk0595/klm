package main

import (
	"bytes"
	"encoding/json"
	"slices"
)

func runnerAnswerDecode(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return controlError(400, "Invalid native answer fields or types.")
	}
	return nil
}

func (a *app) validateRunnerAnswerLocked(sessionID, nativeID, kind string, raw json.RawMessage) error {
	t := a.runs[sessionID]
	if t == nil || t.ctx.Err() != nil {
		return controlError(409, "The original native request is no longer active.")
	}
	if kind == "pending_permissions" {
		var answer struct {
			Decision string `json:"decision"`
		}
		if err := runnerAnswerDecode(raw, &answer); err != nil {
			return err
		}
		pending := t.approvals[nativeID]
		if pending == nil || pending.busy {
			return controlError(409, "The exact permission request is unavailable or resolving.")
		}
		if !slices.Contains([]string{"once", "reject", "session"}, answer.Decision) || !slices.Contains(pending.request.Decisions, answer.Decision) {
			return controlError(400, "Choose an offered request or native-session decision.")
		}
		return nil
	}
	pending := t.questions[nativeID]
	if pending == nil || pending.busy {
		return controlError(409, "The exact question is unavailable or resolving.")
	}
	if pending.request.Elicitation == nil {
		var answer struct {
			Answers   [][]string `json:"answers"`
			Cancelled bool       `json:"cancelled"`
		}
		if err := runnerAnswerDecode(raw, &answer); err != nil {
			return err
		}
		if !validSQLQuestionAnswers(pending.request, answer.Answers, answer.Cancelled) {
			return controlError(400, "Answer every question using its offered choices or allowed custom answer.")
		}
		return nil
	}
	var answer struct {
		Action  string         `json:"action"`
		Content map[string]any `json:"content"`
	}
	if err := runnerAnswerDecode(raw, &answer); err != nil {
		return err
	}
	if pending.elicitReply == nil {
		return controlError(409, "The original elicitation callback is unavailable.")
	}
	if !slices.Contains([]string{"accept", "decline", "cancel"}, answer.Action) {
		return controlError(400, "Choose accept, decline or cancel.")
	}
	if answer.Action == "accept" && pending.request.Elicitation.Mode == "form" {
		schema, err := resolveMCPSchema(pending.request.Elicitation.Schema)
		if err != nil || schema.Validate(answer.Content) != nil {
			return controlError(400, "Complete the form using its requested field types and options.")
		}
	} else if len(answer.Content) > 0 {
		return controlError(400, "Only accepted forms may include content.")
	}
	return nil
}
