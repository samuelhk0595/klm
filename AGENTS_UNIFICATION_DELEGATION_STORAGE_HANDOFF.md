# A CP2a — delegation storage/schema handoff

Status: schema/repository foundation delivered for manager review; **execution and
startup recovery wiring are not activated by this slice**. Manager retains the
checkpoint gate. A owns `engine/sqlite_delegation.go` and
`engine/sqlite_validate_delegation.go` exclusively. B owns its new runtime/service
files (including `delegation_repository.go` / `delegation_prepare.go`) and must not
duplicate the readers declared below.

## Schema 3 and upgrades

- `sqliteSchemaVersion` / fresh `PRAGMA user_version` is **3**.
- The previous `upgradeAgentSchema` still implements the original v1 -> v2 step.
  It accepts an already supported v2/v3 store without recreating A tables.
- `upgradeDelegationSchema` requires v2 and creates only the four new tables through
  the same selected-domain DDL factory used by a fresh store, then commits version 3.
  `enableRuntime` invokes the steps in order only after existing authority verification.
- Existing v2 bindings/native metadata/history are not decoded, rewritten or
  backfilled by this migration. No managed records or Codex manifests are inferred
  from legacy native children/events. V1 defaults for the new roots are absent/empty.

Authoritative concrete B types map as follows:

| Field | Domain | Layout |
| --- | --- | --- |
| diskState.delegationBatches | delegation_batches | Independent chunked roots; metadata excludes Calls |
| batch.calls | delegation_calls | Split alias rows, `owner_id=batch.ID`; array positions |
| diskState.delegationInvocations | delegation_invocations | Independent chunked roots; frozen A capture/task/result |
| diskState.delegationChildren | delegation_children | Map key = record.ID = hidden Session.ID |
| nativeSession.delegation | native_sessions payload | Optional `*DelegationCodexManifest`; no new root/table |

All normal writes remain inside `sqliteStore.sessionTransaction` / existing app
post-COMMIT publication. No alternate DB, model loop, native launcher or global worker
was added. Optional binding authority remains `agent_bindings`.

## Exact projections

`state` is projected as **`x_state`**. Generic `status` remains empty for these
records and is not an execution eligibility predicate. Nested dots become underscores
in the existing projection machinery. Large task/result/config/capture fields are
not indexed.

```text
delegation_batches:
 parentSessionId ownerConversationId parentTurnId operationId requestHash state
 delivery outcome revision endedAt graph.runId graph.activationId graph.turnId
 graph.workspaceId graph.snapshotDigest

delegation_calls:
 parentTurnId sessionId turnId callId delivery

delegation_invocations:
 batchId key ordinal childSessionId parentSessionId parentTurnId state outcome
 dispatchAttempted nativeAccepted nativeSessionId nativeTurnId resultDigest
 capture.config.harness graph.runId graph.activationId graph.turnId graph.workspaceId

delegation_children:
 parentSessionId origin harness executionCwd graphRunId workspaceId
 activeInvocationId quarantined
```

Generic `created_at` / `updated_at` and routing columns are retained. Delegation
writes route `project_id` from the actual parent, and `session_id`/`parent_id` to
the engine parent. For calls, the **native** session ID is `x_sessionId`, not that
engine routing column. Parent/child/visible graph owner scope notices are marked
without fetching any history. Batch DeliveryTurnID and invocation EventStart/EventEnd
stay payload fields; alias Delivery is independently projected.

## Exact additional indexes

All tables retain the generic domain indexes. Feature indexes are:

```sql
-- Batch receipt and serialization
UNIQUE data_delegation_batches_record_unique(record_id)
UNIQUE data_delegation_batches_operation(x_parentSessionId,x_operationId)
UNIQUE data_delegation_batches_active_parent(x_parentSessionId)
 WHERE x_state<>'settled' OR x_delivery IN ('pending','sending','uncertain')
       OR x_outcome='finality_unconfirmed'
data_delegation_batches_owner_page(x_ownerConversationId,created_at DESC,id)
data_delegation_batches_graph_epoch
 (x_graph_runId,x_graph_activationId,x_graph_turnId,x_state,id)
data_delegation_batches_recovery(x_state,x_delivery,id)
data_delegation_batches_pending_recovery(id)
 WHERE x_state<>'settled' OR x_delivery IN ('pending','sending')

-- Immutable native identity, mutable delivery
UNIQUE data_delegation_calls_native_alias
 (x_parentTurnId,x_sessionId,x_turnId,x_callId)
data_delegation_calls_delivery(owner_id,x_parentTurnId,x_delivery,position)

-- Membership and leases
UNIQUE data_delegation_invocations_record_unique(record_id)
UNIQUE data_delegation_invocations_member_key(x_batchId,x_key)
UNIQUE data_delegation_invocations_member_order(x_batchId,x_ordinal)
UNIQUE data_delegation_invocations_active_child(x_childSessionId)
 WHERE x_state<>'terminal'
data_delegation_invocations_batch_order(x_batchId,CAST(x_ordinal AS INTEGER),id)
data_delegation_invocations_graph_epoch
 (x_graph_runId,x_graph_activationId,x_graph_turnId,x_state,id)
data_delegation_invocations_recovery(x_state,id)
data_delegation_invocations_pending_recovery(id) WHERE x_state<>'terminal'

UNIQUE data_delegation_children_record_unique(record_id)
UNIQUE data_delegation_children_lease(x_activeInvocationId)
 WHERE x_activeInvocationId<>''
data_delegation_children_parent(x_parentSessionId,created_at,id)
data_delegation_children_graph_workspace(x_graphRunId,x_workspaceId,id)
```

Contiguous membership is a transaction-final invariant: exactly ExpectedChildren
rows, ordered ordinals 0..N-1, unique child identities; UNIQUE alone does not prove
contiguity. SQL unique constraints serialize conflicting operations/native aliases/
active parent batches/child leases under the normal single writer.

## Readers/writers B can reuse

All following methods are A-owned on `*sqliteSessionTx`:

```go
delegationBatch(id string) (DelegationBatchRecord, error)
delegationChild(id string) (DelegationChildRecord, error)
delegationInvocation(id string) (DelegationInvocationRecord, error)
delegationMembers(batchID string) ([]DelegationInvocationRecord, error)
saveDelegationBatch(record DelegationBatchRecord, expectedRevision string) (DelegationBatchRecord, error)
appendDelegationCall(batchID string, call DelegationNativeCall) error
putDelegationChild(record DelegationChildRecord) error
putDelegationInvocation(record DelegationInvocationRecord) error
putNativeDelegationMetadata(sessionID string, metadata nativeSession) error
```

`delegationMembers` reads at most 9 to detect overflow, never silently truncates
an oversized batch, and returns a maximum of 8 in ordinal order. Batch reads are
metadata-only: Calls are split and must be read through their own bounded collection.
`saveDelegationBatch` owns the counter (initial "0", exact expected CAS then +1),
stores the initial NativeCall alias as well as supplied retry aliases, and strips
Calls from the metadata row. `appendDelegationCall` is idempotent for the exact
4-field tuple in the same batch; another batch conflicts. Alias row key hashing
uses only that identity tuple, excluding mutable Delivery. String CallID **"0"**
is valid and roundtrips.

`validateDelegationWrite` is also called by generic `putRecord`, so bypassing these
convenience writers does not bypass shape/immutability checks. Whole-struct alias
comparison is deliberately not used: ParentTurnID/SessionID/TurnID/CallID are immutable;
Delivery can follow the ordinary pending/sending/delivered or cancelled/uncertain
paths. Delivered/uncertain/cancelled aliases cannot reset to pending. Batch
DeliveryTurnID is not part of settled result identity. Event range payload remains
on the invocation; terminal invocation content is immutable.

After a callback has inserted all related identities/leases, `sessionTransaction`
calls `validateDelegationChanges` before COMMIT. It validates changed batches and
their bounded members/children; ordinary session streaming does not scan old ledgers.
Transactions exceeding the existing change-ID budget fall back to streaming domain
validation. A growing Calls alias collection is streamed in pages, not copied into
every metadata update.

Do not treat successful storage as execution authorization or native acceptance.
B still owns request admission hashing, dispatch intent, physical evidence, native
responses, result delivery and parent/graph linearization. No callback launches or
waits for a process inside SQL/app locks.

## Ownership and immutable evidence

- Managed child has exactly one recorded parent, role=subagent, same project (or
  exact general empty-project ownership), fixed harness and canonical allowed cwd.
- Side parent must belong to a real normal main in that project. General parent
  is genuinely engine-owned and projectless. Subagents cannot become parents.
- Child has no own topic, graph fields/selector or independent spawn receipt.
  A general child can use only its general owner's directory; project children use
  the owning project; graph children use their explicitly associated run workspace.
- Graph ledger binds actual run/project, node/session, activation, turn, workspace
  and captured snapshot revision. Active invocation epoch checks use the current
  activation; historical graph child cwd is checked against its retained workspace,
  not a later cwd of a reusable node. No cross-run/workspace child reuse is permitted.
- Legacy ParentID invariants are unchanged unless an actual managed ledger row is
  present and its strict new validation succeeds. This is not a universal ParentID
  or ExecutionCWD exception.
- Fixed parent/harness/role scope cannot mutate after managed ownership exists.
  Graph parents cannot move cwd with an active/quarantined child lease.
- Invocation capture/task/membership are immutable after admission; dispatch/native
  acceptance evidence cannot be erased. Terminal result content cannot be edited.
  Settled batch result identity is immutable while legitimate delivery/CAS metadata
  can change. Delegation records/aliases are not deleted by state transitions.
- Uncertain terminal finality retains the child lease and quarantine. No helper
  marks drainage confirmed or clears it by a user/parent assertion. Later physical
  settlement evidence beyond the current concrete record protocol needs an explicit
  B/manager contract; storage does not invent it.
- Captured invocation Library/credential references are projected into the existing
  maintained capture indexes. Active invocations and retained quarantined leases
  keep referenced credentials live without opening historical capture payloads.

## Optional Codex manifest

`nativeSession.Delegation` is optional JSON `delegation,omitempty` containing exactly
B's Version/ToolName/SpecDigest. Existing nil records remain nil and their bytes are
not migrated. The writer forbids installing a manifest onto an existing native ID
that lacked one, and forbids dropping/changing its manifest/thread identity after
capture. A new manifest requires a native ID, structurally valid metadata, a Codex
parent role, and no leaf-role registration.

Persistence retains structurally valid future/incompatible manifest versions so
the handler can report capability unavailability rather than rewriting evidence.
B must run its existing exact compatibility check before registering/handling a
dynamic spec. The schema never registers that spec or treats a manifest as a tool
permission grant.

## Recovery seam — NOT wired in CP2a

```go
type DelegationRecoveryHook func(context.Context, *sqliteStore) error
(*sqliteStore).pendingDelegationRecovery(ctx) (bool,error)
(*sqliteStore).delegationRecoveryPage(ctx, afterID string, limit int) ([]DelegationBatchRecord,error)
(*sqliteStore).runDelegationRecoveryBarrier(ctx, hook DelegationRecoveryHook) error
```

Page size is 1..100; pending means nonsettled batches, pending/sending delivery or
nonterminal invocations. The barrier invokes only an explicitly supplied B hook and
refuses to pass if pending work remains. A nil hook does not settle anything or
create successful finality. Already terminal uncertainty/quarantine can remain as
history without automatic replay/reuse; the active-parent/lease guards still apply.

Required future coordinator insertion (B owns runtime/main/readiness files):

1. Read child evidence before graph owner recovery could accept stale Choice/drain.
2. Establish interrupted graph/automation/chat epochs using existing coordinator.
3. B's bounded recovery hook invalidates unfinished work and native delivery,
   preserving uncertainty/quarantine and pausing parent input without dispatch.
4. Call `runDelegationRecoveryBarrier` before scheduling/readiness.

No registration, coordinator edit or replay worker is added by CP2a. The schema
alone does **not** activate delegation. The pending recovery hook must be wired and
reviewed before B execution activation; no readiness guarantee is claimed before it.
# Stacked publication version note

The original handoff below used standalone catalog2/delegation3. Publication above
PR #40 preserves #39 content2, then catalog3/delegation4. Standalone layouts are
explicitly refused rather than silently reinterpreted. See
`AGENTS_UNIFICATION_PUBLICATION.md` for the integrated chain and focused checks.
