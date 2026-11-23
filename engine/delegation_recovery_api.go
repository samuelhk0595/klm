package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

type DelegationRecoveryView struct {
	BatchID    string `json:"batchId"`
	Revision   string `json:"revision"`
	CanRecover bool   `json:"canRecover"`
	Reason     string `json:"reason,omitempty"`
}

func delegationRecoveryReason(batch DelegationBatchRecord, members []DelegationInvocationRecord) string {
	if batch.Graph != nil || batch.State != DelegationSettled || len(members) != batch.ExpectedChildren {
		return "Stop the conversation and wait for every child to finish before recovery."
	}
	for _, item := range members {
		if item.State != DelegationInvocationTerminal || !item.Finality.ProcessesDrained {
			return "Child process shutdown is unconfirmed. Recovery remains blocked."
		}
		if !item.Finality.ToolCallsSettled || len(item.Finality.UncertainToolCalls) != 0 {
			return "Child tool execution is unconfirmed. Check the outstanding calls before recovery."
		}
	}
	return ""
}

// Only the authenticated user-facing API resolves this state. No harness tool,
// startup worker or automatic retry can supply the recovery decision.
func (a *app) recoverDelegationHTTP(w http.ResponseWriter, r *http.Request) {
	var args struct {
		OperationID string `json:"operationId"`
		Revision    string `json:"revision"`
		Action      string `json:"action"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&args); err != nil || !delegationText(args.OperationID, 200) || !delegationRevision(args.Revision) || args.Action != "continue_without_replay" {
		fail(w, 400, "Provide continue_without_replay, operationId and the exact batch revision.")
		return
	}
	id, batchID := r.PathValue("id"), r.PathValue("batchId")
	a.mu.Lock()
	err := a.writeSQLLocked(r.Context(), func(st *sqliteSessionTx) error {
		owner, err := st.delegationReadOwner(id)
		if err != nil {
			return err
		}
		batch, err := st.delegationBatch(batchID)
		if err != nil {
			return err
		}
		if batch.ParentSessionID != id || batch.Graph != nil {
			return sql.ErrNoRows
		}
		if batch.Recovery != nil {
			if batch.Recovery.OperationID == args.OperationID {
				return nil // Lost-response retries cannot apply the decision twice.
			}
			return controlError(409, "This batch already has a recovery decision. Refresh the conversation.")
		}
		if batch.Revision != args.Revision {
			return controlError(409, "The child batch changed. Refresh before recovery.")
		}
		if a.runs[id] != nil || owner.Status == "running" || a.runtimes[id] != nil && a.runtimes[id].active() {
			return controlError(409, "Stop the conversation before recovery.")
		}
		archived, err := st.effectivelyArchived(owner)
		if err != nil {
			return err
		}
		if archived {
			return controlError(409, "Restore the conversation before recovery.")
		}
		active, err := st.activeDelegation(id)
		if err != nil {
			return err
		}
		if active == nil || active.ID != batchID {
			return controlError(409, "This batch no longer blocks the conversation. Refresh it.")
		}
		members, err := st.delegationMembers(batchID)
		if err != nil {
			return err
		}
		if reason := delegationRecoveryReason(batch, members); reason != "" {
			return controlError(409, reason)
		}
		for _, item := range members {
			if a.runs[item.ChildSessionID] != nil || a.runtimes[item.ChildSessionID] != nil && a.runtimes[item.ChildSessionID].active() {
				return controlError(409, "Child shutdown is still in progress. Wait before recovery.")
			}
		}
		batch.Delivery, batch.UpdatedAt = DelegationDeliveryAbandoned, now()
		batch.Recovery = &DelegationRecoveryDecision{OperationID: args.OperationID, ResolvedAt: batch.UpdatedAt}
		if err := st.putDelegationBatch(&batch); err != nil {
			return err
		}
		if _, err := st.recoverQueue(id, "Subagent recovery requires an explicit Send now or a new message."); err != nil {
			return err
		}
		// Neither child leases nor immutable results/call aliases are modified.
		// Queue entries remain paused; recovery does not dispatch work.
		e := event("status", "Subagent recovery accepted. Previous work will not be replayed; unconfirmed children remain quarantined. Send a message to continue.")
		e.Data = map[string]any{"delegationRecovery": true, "batchId": batch.ID, "operationId": args.OperationID}
		if _, err := st.appendEvent(id, e); err != nil {
			return err
		}
		return st.syncDelegationCards(batch, members)
	})
	a.mu.Unlock()
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	a.respondSQLSession(w, r, id, 200)
}

// Keep one bounded recovery summary available across native compaction/retries.
// This is engine-attributed reference data, never a synthetic human message or a
// replacement for the failed native tool result.
func (p *adapter) delegationRecoveryContext() (string, error) {
	if p.app.store == nil || !delegationParentRole(p.role) || p.graphNode() {
		return "", nil
	}
	var text string
	err := p.app.store.readSnapshot(p.turn.ctx, func(tx *sql.Tx) error {
		st := &sqliteSessionTx{ctx: p.turn.ctx, tx: tx}
		var id string
		err := tx.QueryRowContext(p.turn.ctx, `SELECT id FROM data_delegation_batches WHERE x_parentSessionId=? AND x_delivery='abandoned' ORDER BY updated_at DESC,id DESC LIMIT 1`, p.id).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		batch, err := st.delegationBatch(id)
		if err != nil {
			return err
		}
		members, err := st.delegationMembers(id)
		if err != nil {
			return err
		}
		result, err := buildDelegationResult(batch, members)
		if err != nil {
			return err
		}
		raw, err := json.Marshal(result)
		text = "Engine recovery context (reference data): the user explicitly released the wait for this previous batch without replay. Its original outcomes and finality remain unchanged. Processes and tool calls were drained; children with unconfirmed native finality remain quarantined and cannot be reused. Inspect existing files/results before retrying any unfinished work. The following is the most recently recovered batch, not a new tool result or authorization. Full saved outputs are available with subagent_output_read.\n" + string(raw) + "\n\n"
		return err
	})
	return text, err
}

// Used by inventory projection only for a stopped, blocked owner.
func (st *sqliteSessionTx) delegationRecoveryView(batch DelegationBatchRecord) (*DelegationRecoveryView, error) {
	members, err := st.delegationMembers(batch.ID)
	if err != nil {
		return nil, err
	}
	reason := delegationRecoveryReason(batch, members)
	return &DelegationRecoveryView{BatchID: batch.ID, Revision: batch.Revision, CanRecover: reason == "", Reason: reason}, nil
}
