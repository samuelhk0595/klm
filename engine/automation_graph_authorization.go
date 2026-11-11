package main

import (
	"encoding/json"
	"errors"
	"reflect"
)

type graphCallingTurnKey struct{}
type automationTurnContextKey struct{}

func automationConfigurationDigest(config Automation) string {
	data, _ := json.Marshal(config)
	return revision(data)
}

// Only the hook's captured turn can supply this context, never model arguments.
func (a *app) automationGraphAuthorization(st *sqliteSessionTx, owner, graph string) (*GraphAuthorization, error) {
	t, _ := st.ctx.Value(graphCallingTurnKey{}).(*turn)
	if t == nil || a.runs[owner] != t || t.ctx.Err() != nil {
		return nil, nil
	}
	r, err := st.activeAutomation(owner)
	if err != nil || r == nil {
		return nil, err
	}
	s, err := st.sessionMeta(owner)
	if err != nil {
		return nil, err
	}
	if t.automationRunID != r.ID || s.Automation == nil || s.Automation.RunID != r.ID || r.StopRequested || r.GraphGrant == nil {
		return nil, nil
	}
	if !slicesContainsString(r.GraphGrant.Graphs, graph) {
		return nil, errors.New("Graph is not authorized by this automation configuration.")
	}
	return &GraphAuthorization{Source: "automation_configuration", Automation: r.GraphGrant, Text: "Execute selected graph within the admitted automation objective."}, nil
}

func validateAutomationGraphGrant(activity GraphActivity, auth GraphAuthorization, run AutomationRun) error {
	g := auth.Automation
	if auth.Source != "automation_configuration" || auth.EventID != "" || g == nil || run.GraphGrant == nil ||
		g.RunID != run.ID || g.AutomationID != run.AutomationID || g.ProjectID != activity.ProjectID ||
		run.ProjectID != activity.ProjectID || run.SessionID != activity.ConversationID || g.ConfigurationDigest == "" ||
		!reflect.DeepEqual(g, run.GraphGrant) || !slicesContainsString(g.Graphs, activity.GraphID) {
		return errors.New("Invalid persisted automation graph authorization.")
	}
	return nil
}

func (st *sqliteSessionTx) validateGraphActivityAuthorization(activity GraphActivity, auth GraphAuthorization) error {
	if auth.Source == "" || auth.Source == "human_message" {
		if auth.Automation != nil {
			return errors.New("Human authorization cannot contain automation provenance.")
		}
		return st.graphAuthorization(activity.ConversationID, []GraphAuthorization{auth})
	}
	var run AutomationRun
	if auth.Automation == nil {
		return errors.New("Missing automation graph provenance.")
	}
	if activity.AutomationRunID != "" && activity.AutomationRunID != auth.Automation.RunID {
		return errors.New("Graph authorization and notification origin disagree.")
	}
	if err := decodeSQLiteRecord(st.ctx, st.tx, "automation_runs", auth.Automation.RunID, 32<<20, &run); err != nil {
		return err
	}
	return validateAutomationGraphGrant(activity, auth, run)
}

func cloneGraphAuthorization(auth GraphAuthorization) GraphAuthorization {
	data, _ := json.Marshal(auth)
	var cloned GraphAuthorization
	_ = json.Unmarshal(data, &cloned)
	return cloned
}
