# Delegation CP1b — concrete types and persistence/runtime handoff

Status: additive B contract implemented, **not registered or activated**. Runtime
and persistence integration require CP2 release. A owns shared schema/store/API
inserts and base agent binding; B's future isolated delegation transactions/reads/
recovery files are not automatically assigned to A.

## 1. Implemented type and protocol authority

Authoritative Go source is now:

- `engine/delegation_records.go`: `DelegationBatchRecord`,
  `DelegationInvocationRecord`, `DelegationChildRecord`, `DelegationGraphScope`,
  `DelegationNativeCall`, `DelegationFinality` and structural/pure transition helpers.
- `engine/delegation_tools.go`: versioned request/schema, private RPC envelopes,
  all-terminal aggregation, delivery/result validation and deterministic hashing.
- `engine/delegation_codex.go`: native dynamic specification/manifest, exact-bound
  call decoding, response encoding and new-versus-existing-native eligibility.
- `engine/delegation_prompts.go`: leaf-role specialist/task composition only.
- `engine/delegation_assets.go`, `delegation-{protocol,pi,opencode}.mjs`: unregistered
  embedded assets and a pure renderer returning private-profile files to its caller.

The old illustrative Go snippets in CP0 are superseded by these concrete types.
Invocation configuration is **`Capture AgentExecutionCapture` from A**, not another
config/binding type. It contains A's resolved model/prompt/selection, source capture
and captured external Library. Child binding/overrides remain A's Session-owned
`agent_bindings` record. B's child ledger does not duplicate that authority.

No `init()` registration, route insertion, adapter hook, worker goroutine, native
delegation suppression in the current engine or SQLite writer is introduced.

Manager correction B-R1 is incorporated in the additive assets: fatal/uncertain
execution state is latched independently of active Promise bookkeeping and is not
reset by `finally` or native abort HTTP completion. The owning runtime must dispose
the failed execution/transport after its actual shutdown/drain. Before-model hooks
must use `assertCanContinue()` / `canContinue()`, not `isWaiting()` alone. Pi guards
both before-agent-start and before-provider-request; OpenCode guards its message
transform even while a native abort request remains pending. This is locally
hermetic-regression-tested, not a claim of product or native timing validation.

## 2. Exact root/domain request for A

Add under A's single coordinated additive schema migration, with empty defaults:

```go
DelegationBatches     []DelegationBatchRecord     `json:"delegationBatches,omitempty"`
DelegationInvocations []DelegationInvocationRecord `json:"delegationInvocations,omitempty"`
DelegationChildren    map[string]DelegationChildRecord `json:"delegationChildren,omitempty"`
```

Domain mappings:

| Root / collection | Table | Identity |
| --- | --- | --- |
| `delegationBatches` | `delegation_batches` | Fresh batch `ID`; parent-scoped operation receipt. |
| Batch `calls` child collection | `delegation_calls` | Native request/call aliases, split rows owned by batch. Include the initial native call as well as retry aliases. |
| `delegationInvocations` | `delegation_invocations` | Fresh invocation `ID`, one execution per child lease. |
| `delegationChildren` | `delegation_children` | Map key and record `ID` equal the hidden child Session.ID. This is ownership metadata, intentionally not a new globally distinct session identity. |

The batch's `NativeCall` is its immutable initial admission correlation summary.
`Calls` is the separately stored alias collection: persist each accepted native
call correlation, including retries returning an already accepted operation. A
new RPC ID is not a new task. Native response IDs are transport correlation and
must preserve numeric **0**; do not use string nonemptiness as their validity test.

The three roots must join schema-domain inventory, chunk handling, import defaults,
new-store creation, validation and scoped revisions. Do not populate managed rows
from legacy native task/subagent event names. No whole history arrays in metadata.

### Required projected/index fields

These records use **`state`**, not the generic `status` field. Eligibility queries
must use the explicit state projection, not assume `data_*.status` was populated.

- Batches: `parentSessionId`, `ownerConversationId`, `parentTurnId`, `operationId`,
  `requestHash`, `state`, `delivery`, `outcome`, `revision`, `createdAt`, `endedAt`,
  `graph.runId`, `graph.activationId`, `graph.turnId`, `graph.workspaceId`,
  `graph.snapshotDigest`.
- Calls: `parentTurnId`, `sessionId`, `turnId`, `callId` (batch via owner_id).
- Invocations: `batchId`, `key`, `ordinal`, `childSessionId`, `parentSessionId`,
  `parentTurnId`, `state`, `outcome`, `dispatchAttempted`, `nativeAccepted`,
  `nativeSessionId`, `nativeTurnId`, `resultDigest`, `capture.config.harness`,
  `graph.runId`, `graph.activationId`, `graph.turnId`, `graph.workspaceId`.
- Children: `parentSessionId`, `origin`, `harness`, `executionCwd`, `graphRunId`,
  `workspaceId`, `activeInvocationId`, `quarantined`.

Indexes/uniqueness:

1. Unique `(parentSessionId, operationId)` for batch receipts. Same canonical
   decoded request hash attaches to existing work/result; changed hash conflicts,
   including after delivery. Never dedupe by task text/title/agent slug.
2. Unique `(parentTurnId, sessionId, turnId, callId)` over all accepted call aliases;
   use validated native/engine epochs for transports without native turn IDs.
3. Unique `(batchId,key)` and `(batchId,ordinal)`, contiguous ordinals `0..N-1`.
4. Unique active invocation per child; batch admission reserves all child leases in
   one transaction. At most one unsettled batch/delivery per parent, including
   native parallel tool calls. State checks and uniqueness must share the writer.
5. Index batch membership/order, active parent batches, active child leases,
   owner created-time/ID pages, graph activation child set and recovery eligibility.

Large task/result/config/snapshot fields stay chunked and do not enter indexes.
Revisions are canonical decimal strings in JSON. Pure transition helpers do not
perform CAS/revision increments; the eventual repository owns those atomic updates.

## 3. Native Codex compatibility metadata

Manager relayed the user's explicit decision **new native threads only**, reference
`85e3f4d3845b2f7b80dd2e1e812c0f90` in the manager conversation (not a local human
event/tool grant). Close legacy-adoption research; no migration/fork/DB/transcript
rewrite or pin change is included.

Request a persisted optional **`DelegationCodexManifest`** on `native_sessions`
(for example `delegation,omitempty`, name to be agreed with A's shared-file owner).
It records schema `Version`, exact `ToolName` and `SpecDigest` only, no private token.

- If native ID is absent, the future adapter can create a new native thread with
  the v1 spec and persist its manifest when the matching native ID is accepted.
- Existing native ID + no manifest remains ordinary chat with delegation unavailable,
  including an existing general agent; never infer support from session age, current
  binary version or an ordinary successful resume.
- Matching manifest resumes with a compatible handler and the saved native schema.
  Incompatible manifest blocks this capability; do not silently replace specs.
- Child leaves never register the parent dynamic tool, even when their native thread
  is new. Graph/platform/managed-profile checks remain separate mandatory gates.

Frozen v1 **complete function spec** digest:

`2b9a6f688babd8f22a33a631ab0a707c77c050a9ad6a3214c9d390b4a83902c0`

`delegationCodexSpecV1Digest` is a golden guard. Editing its name/schema/description
without an explicit version/compatibility change makes manifest preparation fail.
CP1a's simpler task-only fixture spec was never installed as a product capability;
it does not authorize relabeling a native product manifest as v1.

## 4. Structural invariants and transaction boundaries

Concrete `Validate()` methods cover local shape/version/state/digest invariants.
`validateDelegationOwnership` additionally checks the supplied Session relationship;
the repository must validate referenced stored rows, actual run/activation/workspace/
snapshot, source status/config eligibility and same-child lease under the same lock.

- Parents: normal main, side, general or graph node only; no subagent-as-parent.
  Side ownership cannot be confused with top-level project membership. General has
  empty project; graph children remain `role=subagent`, with parent=node Session,
  and graph ownership lives in the explicit ledger, not a Choice grant on the child.
- Fixed child harness/canonical cwd and exact parent identity; no own topic, graph
  selector, independent spawn receipt or inherited administrative tools/internal skills.
- Invocation capture is immutable for that execution. Source I/O/invalid/conflict
  blocks preparation; only confirmed disabled/deleted sources use A's validated
  fallback. Inherited external selection is creation-only explicit human opt-in.
- Dispatch attempt intent persists before native prompt bytes can be sent; native
  acceptance is separately correlated. A crash in that window is not proof of no work.
- Terminal completion requires native acceptance plus native/process/tool settlement.
  Result digest includes outcome/task outcome/text/error/finality. Terminal content
  changes fail validation. Native completion never invents semantic satisfaction.
- `buildDelegationResult` rejects missing/running members, identity/order/scope
  mismatches and duplicate children. Failed siblings still join; mixed batches
  return an error outcome, not `success=true`.
- Unconfirmed finality cannot be encoded as a native result: `CanContinue=false`
  requires the owning parent runtime to stop. `finishDelegationChild` quarantines
  and retains the uncertain active lease rather than making it reusable.
- Settled batch/result digest precedes native sending. Delivery moves through
  pending -> sending -> delivered or uncertain/cancelled, never skips the native
  correlation boundary or automatically retries uncertain consumption.

Needed repository/runtime entrypoints (names not yet wired):

1. Atomic admit/attach receipt + all child identities/leases + invocation rows +
   parent cards, with parent/activation/Stop-version CAS; return replay vs dispatch.
2. CAS prepare/dispatch intent and correlated native-accepted update.
3. Per-child drainage/terminal result + lease/quarantine update + visible event.
4. Read/join complete membership; persist consolidated result and digest atomically.
5. Native call/result delivery correlation; preserve alias receipt identity on retry.
6. Parent cancellation intent, owned child shutdown outside locks, then final evidence.
7. Bounded read/detail/history/request projection and existing scoped invalidations.

No native process, capability discovery, model call or wait inside SQL/app/graph
locks. Future B service/read/recovery files can implement these under CP2 once A's
shared root/schema/hooks are merged; they need not all be implemented by A.

## 5. Bounded private wrapper protocol

Only parent role-bound wrappers call these private methods; no public execution
endpoint/model wait tool is added:

| Method | Contract |
| --- | --- |
| `klm/delegation/ready` | `{version,specDigest}` -> `DelegationReadyReply`; exact v1 digest, max children 8, max concurrency 4, active-batch limit 1. Attest before native suppression. |
| `klm/delegation/gate` | `{tool}` -> `{allowed:true}` only when owner admits that dispatch; false/error blocks it even under YOLO. |
| `klm/delegation/admit` | `DelegationAdmitRequest` -> `DelegationAdmissionReceipt`; authenticate owner/role, validate all-or-none, persist receipt before dispatch; identical retry attaches. |
| `klm/delegation/wait` | `DelegationWaitRequest` -> `DelegationWaitReply`; <=8s server wait and <=10s wrapper fetch; only waiting, consolidated settled result, or stop-without-result. |
| `klm/delegation/cancel` | `{batchId}` cancellation intent for the authenticated owner; wrapper best-effort call <=1s, owning turn cancellation remains authoritative. |

JSON-RPC `error.data=DelegationRejection{accepted:false,code}` is only for a proven
atomic **pre-admission** rejection (`invalid_request`, `batch_busy`,
`operation_conflict`, `child_busy`, `configuration_blocked`). That error can reach
the model without aborting a different admitted batch. Ambiguous transport or
post-admission failures must not claim accepted=false; wrappers abort the parent
so an exception cannot release it while work remains. No automatic admission retry.

The accepted=false recovery marker is valid only for the admission RPC. A later
wait error cannot reuse it to reopen continuation. Uncertain alias/admission delivery
arms the same permanent local stop latch; a late original success cannot clear it.
Alias responses are counted until settled, and known-batch cancel is deduplicated
and bounded. Invalid input/proven pre-admission alias rejection leaves other valid
admitted work intact. Owner Stop/drain remains authoritative if cancel transport
fails or the batch ID was never received.

Request schema is final v1 `{operationId,children}` with 1–8 children; concurrency
4 is engine-owned, not a caller knob. Child keys/tasks required; exact reuse ID,
saved agent slug, harness/model/effort and creation-only boolean inheritance optional.
No async/background/depth/session-creation/workspace/path fields are accepted.

Native consolidated output currently has a **32KiB** fail-closed budget, below the
pinned OpenCode generic output-byte truncator. Oversized output is rejected, never
silently truncated into an incomplete batch. Before CP2 activation, the collector
must produce bounded all-member summaries/content references and attest effective
native output limits so all member outcomes remain represented. Full child history
and persisted results retain the existing chunked-content policy.

## 6. Restart/recovery order required before readiness

Within A's shared recovery coordinator, expose hooks for B's eventual bounded
delegation recovery; do not start a replay worker just because new domains exist.

1. Establish interrupted owning graph/automation/chat epochs. Read child evidence
   before any graph recovery could treat stale Choice as accepted/drained.
2. In bounded writer pages, mark nonterminal invocations interrupted and preserve
   uncertain dispatch/native/tool evidence; settle parent batch as interruption/
   uncertainty, not success. Keep terminal results immutable.
3. Retain quarantined child leases if physical/tool settlement is unknown; no user/
   parent assertion of success resolves physical uncertainty. Pause unsent parent
   input and clear replacement/continuation scheduling intents.
4. Pending native delivery is cancelled/interrupted; sending without transcript
   evidence is uncertain. Delivered receipts stay delivered. Repeating recovery is
   idempotent and starts no harness, MCP, model prompt, child or result delivery.
5. Only then enable normal readiness/scheduling. A later explicit parent turn may
   read an old receipt or request a new operation using an owned reusable child;
   it cannot revive the old interrupted operation automatically.

Existing v1 graph snapshot bytes/validation remain intact; new v2 captures the full
eligible visible catalog at run start. Choice reservation and child admission must
share a linearization point, and acceptance additionally needs every child/tool/
process/delivery boundary settled. No Linux graph support is implied.
