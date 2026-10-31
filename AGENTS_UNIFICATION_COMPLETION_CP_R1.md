# Completion owner R — CP-R1

Status: **candidate delivered; stopped for manager review**. No CP-R2 release is
implied. Owner: `92467a25db7be6788388129d542b401f`. Manager:
`a05369eaa5c9151b61e93fb38cb35c52`.

## Implemented

- Registered authenticated private ready/gate/admit/prepared/wait/cancel dispatch.
  Preparation remains temporary, owner-bound and bounded; children start only after
  the existing atomic admission commits. A preparation/capture remains authoritative
  through the whole child invocation, including explicit empty Library snapshots.
- Loaded owned Pi/OpenCode wrappers in actual adapter profiles. Pi's old native
  factory is not loaded for an owned parent. OpenCode composes the existing policy
  plugin with the replacement, suppresses task per input and after YOLO config
  rewriting, and keeps the unconditional before-tool gate. Leaf profiles register
  neither delegation nor administrative tools.
- Codex registers the frozen v1 dynamic spec only on new thread creation, persists
  the manifest atomically with the first native ID, and uses the same handler on
  compatible resume. Existing histories without a manifest retain normal chat
  without subagent_run. Dynamic requests wait in an owned goroutine while the native
  reader continues. No legacy adoption, history migration or pin changes.
- Consolidated results enter sending before return; actual native result events
  correlate each alias and confirm delivery. Replay aliases retain independent
  fences without resetting historical aliases or immutable results. Premature root
  completion fails; cleanup cancels/joins children. JS fatal-stop remains latched,
  including engine gate RPC failures, and both wrappers gate model boundaries.
- Steering is deferred during preparation/join/delivery and wakes after correlated
  delivery. Existing parent Stop/Send now cancellation reaches child contexts; only
  confirmed shutdown releases replacement input.
- Child permissions consult live owner YOLO. A full-access Codex child blocks owner
  downgrade until completion/Stop. Exact permission/question/elicitation replies
  route through the owner to the active child request; public child replies remain
  rejected. Child request changes invalidate the owner's projection.
- Thin MCP callbacks record actual send and classify received response, explicit
  not-sent/protocol error, or uncertainty. Client closure is not response evidence;
  existing graph tracking is preserved.
- Startup calls runDelegationRecoveryBarrier with recoverDelegationRuntime before
  ordinary/graph recovery, scheduling and readiness. Dispatched interrupted work
  keeps uncertainty/quarantine; delivery is cancelled/uncertain and input paused.
  Pending replay aliases participate too. No replay or replacement execution.

## Actual HTTP/DTO handoff

Authenticated read-only routes below `/api/sessions/{id}/delegation`:

| Route | Response |
| --- | --- |
| `GET /batches?offset=0&limit=20` | DelegationBatchPage: items, string revision, hasMore, nextOffset; limit max50 |
| `GET /batches/{batchId}/members?offset=0&limit=20` | DelegationMemberPage: batch, items, revision, hasMore, nextOffset |
| `GET /batches/{batchId}/invocations/{invocationId}/output?digest=...&offset=0` | DelegationOutputPage; UTF-8 content <=16KiB, string offset/nextOffset/totalBytes |
| `GET /children?afterId=...` | DelegationOwnedChildView identities, reusable/quarantined/readOnly; items/hasMore/nextAfterId, max50 |
| `GET /requests?afterId=...` | DelegationRequestView: ownerSessionId, childSessionId, batchId, invocationId, kind, actual request; max50 |

DelegationChildView includes effective harness/model/effort, source labels,
executionState/outcome, reused and string invocationNumber. Terminal members expose
`resultReference { invocationId, digest, totalBytes }` for the output route. Batch
actions include send/queue/steer/interrupt/settings/stop with waiting/delivery fences.

Reply using the existing **owner** routes and actual pending request ID:
`POST /api/sessions/{ownerId}/permissions/{permissionID}` and
`POST /api/sessions/{ownerId}/questions/{questionID}/reply`. Request/grant ownership
remains with the child. Existing GET session/history/content/events routes provide
the read-only child timeline; no child composer/mutation endpoint was added.

Private model read `subagent_output_read` accepts batchId, invocationId, digest and
optional decimal-string offset, scoped to this parent. It returns only bounded
output, never captured config/credentials. subagent_run is a local wrapper/dynamic
tool, not an external MCP tool waiting beyond the bridge timeout.

## Files touched by R

New: engine/delegation_api.go, delegation_codex_runtime.go, delegation_recovery.go,
delegation_requests.go, delegation_parent_smoke_test.go, delegation_runtime_smoke_test.go.

Runtime/private tools: delegation_bridge.go, delegation_delivery.go,
delegation_service.go, delegation_repository.go, delegation_projection.go,
delegation_child.go, delegation_prompts.go, delegation_assets.go,
delegation-protocol.mjs, delegation-pi.mjs, delegation-opencode.mjs, harness.go,
linked_bridge.go, graph_adapter.go (ordinary output catalog only), runtime.go,
pi_interactive.go, pi-permissions.ts, opencode_interactive.go,
opencode_permission_gate.go, codex_interactive.go, mcp_client.go, live_yolo.go.

Thin hooks: api.go, sqlite_runtime.go, sqlite_execution_requests.go,
sqlite_execution_queue.go, sqlite_record_writes.go, sqlite_delegation_recovery.go.
All code paths above are under engine/. A's sqlite_delegation.go and
sqlite_validate_delegation.go were preserved. No frontend or graph-node
activation/Choice implementation in this checkpoint.

## Commands and evidence

All Go commands ran from engine; no full suite ran.

1. First recovery check:
   `go test -tags dev -run '^$' -timeout 45s .`
   Tool bound60s; exit0, package0.142s, no tests run.
2. After initial integration, the same compile-only target:
   tool bound60s; exit0, package0.124s, no tests run.
3. New central native-parent check:
   `go test -tags dev -run '^TestDelegationNativeParentSmoke$' -count=1 -timeout 55s -v .`
   Tool bound65s; exit0, test10.14s, package10.274s. Actual installed Pi root wrapper
   and two actual child adapters used a deterministic loopback provider: four local
   requests, consolidated result in the next parent input, delivered native alias.
4. After engine model-boundary gates and alias/control/API/recovery integration:
   `go test -tags dev -run '^(TestDelegationNativeParentSmoke|TestDelegationRuntimeRequestsRecovery)$' -count=1 -timeout 65s -v .`
   Tool bound75s; exit0, native parent9.03s, request/recovery1.15s, package10.320s.
   The local control fixture verified owner request projection/exact reply, public
   child reply rejection, restart uncertainty/quarantine, and registered batch/
   member/children/request HTTP reads. It used no provider/native process.

Native opt-in: KLM_RUN_DELEGATION_NATIVE_PROBE=1, exact installed Pi package
`C:\Users\Samuel\AppData\Roaming\npm\node_modules\@earendil-works\pi-coding-agent`
and Node `C:\Program Files\nodejs\node.exe`. HOME/profile/skills/model config,
SQLite and native histories were temporary; credential/proxy envs were cleared.
Owned runtime shutdown and Go fixture cleanup were used. No installed engine/user
data changed, package installation, commit/push or real graph execution.

The old TestDelegationNativeBatchSmoke was **not rerun**. Its earlier join/reuse/Stop
evidence remains attributed to the interrupted checkpoint. The final post-smoke
edits only clarify source comments and the owned-tool exception to native-only
collaboration guidance; they do not change the v1 schema/digest or execution code.

## Boundaries / next release

- New engine integration has an actual native Pi central smoke. OpenCode/Codex
  engine paths, side/general combinations, actual MCP/providers and visual behavior
  still need runtime/manual acceptance. Prior protocol feasibility is not relabeled
  as this checkpoint's integration evidence.
- CP-R2 requires manager review/release: frozen graph lookup/compose, activation/
  workspace/Choice join+drain+delivery, then frontend batch/requests/history/reuse
  against these real routes. No CP-R2 advancement from report acceptance alone.
- CP-R3 and the separately created final runtime/Playwright session remain later.
  R will not create that session or claim visual validation from fixtures.
