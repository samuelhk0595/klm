# Engine-owned delegation — stream B / CP0

Status: **CP1b additive types/protocol/assets implemented without product activation**.
CP1a local fixtures passed; the user approved Codex support for new native threads
only. Existing Codex histories remain ordinary chat without delegation. Concrete
types and current ownership/hook requirements are now authoritative in
`AGENTS_UNIFICATION_DELEGATION_PERSISTENCE_HANDOFF.md` and the new Go source.
CP1b is still awaiting manager approval; its B-R1 local fatal-stop regression was
corrected with a persistent transport latch and focused hermetic revalidation.
See `AGENTS_UNIFICATION_DELEGATION_PROGRESS.md` for exact fixture evidence and
unexecuted scenarios. This document is not runtime acceptance. Read together with `AGENTS_UNIFICATION_IMPLEMENTATION_PLAN.md` and
stream A's catalog contract when the manager merges the seams.

- Worker conversation: `e80963a1b2bdb9c249b765921feb7d28`.
- Manager conversation: `a05369eaa5c9151b61e93fb38cb35c52`.
- Scope released: CP0/CP1a and CP1b additive compilable types, pure helpers and
  unregistered assets only; no durable service, product runtime or frontend wiring.
- No sessions/subagents spawned, graphs run, providers called, installations,
  commits, production data changes or production restarts performed.

Manager integration decision received after the initial CP0 report: A is the
exclusive persistence owner (`store.go`, `api.go` A inserts, SQLite domains/schema/
store/indexes/migration and persistence integration). B supplies requirements here
and does not edit those files. Harness/adapters/graph execution/capabilities/private
tool dispatch remain reserved for B's **future unreleased** slice. The confirmed
resolver and graph decisions below supersede conflicting earlier CP0 proposals.

## 1. Decision summary

1. Execute children through the existing `execute` / `runPi` / `runOpenCode` /
   `runCodex` adapters, with hidden `role=subagent` sessions and separately owned
   child turns. Do not use a native task/subagent executor or build a model loop.
2. Expose one logical **blocking batch** operation, `subagent_run`. Its public
   return contains every admitted child's terminal outcome; there is no `async`,
   background mode, model-visible wait/poll tool, or notification subscription.
3. **Do not implement the blocking wait as a long `klm_linked` MCP request.** Use
   an owned Pi sequential tool, an owned OpenCode local plugin tool, and Codex
   app-server dynamic tools. Admission/read transport is bounded; the native tool
   promise/request remains pending while an engine-owned batch runs.
4. Keep one complete parent engine turn, role/capability snapshot, queue reservation
   and graph activation alive across the join. A native idle/completion event is
   not permission to finish the engine turn while a batch or delivery is unsettled.
5. Suppress native delegation by effective profile/tool inventory, including under
   YOLO, and enforce leaf-only dispatch on the engine, not just in child prompts.
6. Reuse A's catalog/resolver/binding types. B does not own another agent registry.
7. Codex pending dynamic-call feasibility passed the bounded CP1a fixture. The
   user then chose **new native threads only**; old native histories (including
   general) keep normal chat without this capability. No adoption/fork/migration
   research or implementation remains in scope. Production batching remains unverified.

## 2. Inspected implementation and concrete gaps

| Existing seam | Observation relevant to implementation |
| --- | --- |
| `engine/subagents.go` | Hidden native children use the parent's harness/turn; `ensureSubagent` is observation/projection, not engine-owned execution admission. `closeSubagents` derives status from the parent and cannot substitute for joining actual independent child adapters. Preserve it for historical/native projection. |
| `engine/harness.go:168-397` | `execute` already owns capability preparation, native invocation, streams, MCP drain and final persistence. It removes `a.runs[id]` and closes `t.done` at the end. Insert join/delivery checks **before** bridge unbind, capability teardown, graph settlement and `done`. A child must not get ordinary conversation guidance. |
| `engine/linked_bridge.go:15-22,149-159,288-329` | Finite server/client/write deadlines. Returning `action=yield` from `linked_ask` is cooperative guidance plus later scheduling; it does not prevent another model step/tool in the original turn. Do not copy this as the delegation barrier. |
| `engine/pi-permissions.ts` | Existing registered tools are sequential; `tool_call` runs before permission/internal bypasses; `turn_end`/`before_agent_start` take steering. Current tools/call fetch has a timeout. New blocking local tool must bypass that long-call path and guard steering/continuation boundaries. |
| `engine/pi_interactive.go`, `pi_subagents.go`, `pi-subagents.mjs`, `pi-subagent-factory.mjs` | Current managed Pi delegates to the installed native package, including background runs. It is not the new runtime. Stop loading this package in managed execution after replacement; retain historical readers. Its policy-only external snapshot/attribution is a useful pattern, not a cross-harness executor. |
| `engine/opencode-permissions-plugin.mjs` | Owned tool-before and pre-model `experimental.chat.messages.transform` hooks exist. YOLO rewrites permissions/tools to allow: native-task and leaf restrictions must be reapplied **after** that rewrite and verified, or YOLO re-enables delegation. |
| `engine/opencode_library.go:584-608,632-710` | Root build/plan policy and selected skill gates are projected and inspected. Native `task` is not yet replaced. Preserve root policy while removing native task/subagent routes. |
| `engine/codex_library.go:50-101` | Native multiagent/agents already disabled. Owned MCP servers have finite `tool_timeout_sec`; no reliable general pre-tool gate is provided by this adapter. |
| `engine/codex_interactive.go` | Existing app-server RPC loop supports server requests, permissions/questions, native thread resume and tool transcript settlement. `item/tool/call` needs a new explicit handler. Do not block the frame reader while a child runs. |
| `engine/runtime.go` | Ordinary runtimes can retain user workloads across turns. Children need stricter invocation teardown: do not retain live child processes after reporting terminal results. Preserve native identity/history rather than a warm child process. |
| `engine/sqlite_execution_requests.go:118-143` | Public child request targets are refused; graph routing currently follows graph-node sessions. Add an exact owned-child request projection/resolver rather than opening direct child chat controls. |
| `engine/sqlite_validate.go:121-189`, `store.go:254-280` | Existing ownership validation rejects graph-parent children, child `ExecutionCWD`, and `subagent` with graph fields. Explicitly model the new relation; do not globally loosen ParentID validation. |
| `engine/graph_adapter.go`, `graph_mcp_calls.go` | Choice reservation/seal/drain/accept uses parent-local process and MCP evidence. Cross-harness engine children are outside that current domain. They need an activation-owned join/drain ledger. |
| `engine/sqlite_graph_execution.go` | Activation has a durable `CurrentTurnID` and exact workspace. Extend binding/finality evidence, not graph scheduling through chat Stop or another graph invocation. |
| `engine/sqlite_schema.go`, `sqlite_domains.go`, `sqlite_indexes.go`, `sqlite_store.go` | SQLite schema version is 1; runtime domain DDL is created at initial store construction. Editing the reflected domain list alone does not upgrade an existing store. A coordinated additive migration is required. |
| Frontend `engine.ts`, `ConversationEvents.tsx`, `SubagentView.tsx` | Existing `subagent` events and read-only child history are reusable. Current card calls every non-running child `completed`; this must use invocation outcome. Hidden graph histories require graph-owned routes, not arbitrary conversation access. |

`PI_SUBAGENTS_IMPLEMENTATION_PROGRESS.md` explicitly records unvalidated model,
permission, Stop and background delivery paths. Its load probe is not acceptance
of the new design.

## 3. Blocking operation and ownership

Proposed single logical request (same schema on all native wrappers):

```ts
type SubagentRunRequest = {
  operationId: string; // nonblank, <=200 bytes; stable for a retry
  children: Array<{
    key: string; // unique within this request; correlation, not a global identity
    task: string; // self-contained, nonblank, <=128 KiB
    title?: string; // <=200 characters
    childSessionId?: string; // exact prior child of THIS parent, optional reuse
    agentId?: string; // A's globally unique slug; never an authorization grant
    harness?: 'pi' | 'opencode' | 'codex';
    model?: string;
    effort?: string;
    inheritParentCapabilities?: boolean; // false; explicit human opt-in only
  }>;
};
```

Manager-approved capacity: 1–8 children per call, four simultaneously executing
children per parent; excess admitted children stay in the same batch's bounded
dispatch queue. These are infrastructure limits, not a background product mode.
Only one outstanding batch per parent engine
turn is admitted. A concurrent *different* batch request gets a deterministic
`batch_busy` rejection; an identical retry attaches to the same pending result.
Use `children[]` for intended parallelism instead of multiple independent calls.

Admission is all-or-none: validate roles, size, unique keys/child identities,
resolver results, source lifecycle, models, workspace and capacity before admitting
any work. Persist all identities/reservations and the receipt atomically before
dispatch. Preparation failures after admission are child outcomes, not a rollback
of already executed siblings. Continue/join siblings; do not fail-fast past them.

Owner identity is a tuple, not a client-supplied path:

- `parentSessionId`, durable `parentTurnId`, `ownerConversationId`;
- optional `graphRunId`, `activationId`, graph `turnId`, `workspaceId`, snapshot digest;
- child session ID, batch ID, invocation ID and native request/call ID.

`ownerConversationId` is the actual ordinary/side/general chat displaying requests;
for a graph node it is the graph's owning main chat. It does not change who owns
the graph or give that chat Stop authority over an independent graph.

Children are leaf sessions. They cannot be entered into ordinary input queues,
used as independent-session destinations, or reused by another parent. A child
identity has fixed harness and cwd, at most one active invocation, and serial native
history. No lookup/reuse by title, slug similarity, native recency or project alone.
Missing/corrupt native history is an error, not an implicit replacement.

New ordinary/side children inherit the effective parent's execution directory;
general children use its dedicated workspace. Graph children use the exact recorded
activation workspace, including Fork/Join isolation already chosen by the graph.
No caller-supplied child directory/worktree/base revision is introduced.

## 4. Enforcing the join without a long MCP/HTTP call

### Common invariant

`subagent_run` has **one native tool call/result pair**. The result is unavailable
until all admitted invocations are terminal and their owned work/drain evidence is
settled. The parent does not receive an early admission receipt as a tool result.
Consequently its next native model iteration cannot consume a partial batch.

Inside the owned wrapper only, use private short admission and bounded revision
wait/read operations. Proposed wait: <=8 seconds server deadline / <=10 seconds
client deadline, immediately woken by a durable revision change. Reissue only the
read wait under the same batch, with a cancellable wrapper lifetime. This is
transport maintenance, not model polling, task retry or admission replay. Keep a
lease/watchdog and cancellation; broken ownership/transport cancels the batch and
fails the parent rather than letting it continue while children run. No HTTP call
or MCP tools/call remains open for a model task's whole lifetime.

The parent engine turn keeps a join reservation even if a harness incorrectly
returns idle/final. The adapter must reject early completion, stop/cancel and join
children, and produce a failure; it must not accept the premature assistant text
as a successful completed task. Do not merely wait in `execute` after allowing
native execution to continue unguarded.

The local wrapper has a separate persistent fatal-stop latch as well as active
call bookkeeping. Post-admission failure/uncertain admission or alias delivery
closes it before any cancellation await. Promise cleanup cannot reopen tool or
model continuation; hooks use `assertCanContinue()` rather than `isWaiting()`.
There is no local reset before the owner actually ends/disposes that execution.
Proven pre-admission rejection remains recoverable and cannot cancel another valid
batch. See B-R1 in the progress document for the preserved manager regression and
alias/late-result/mock-hook cases. This does not replace durable owner drain.

### Pi 0.85.1

- Register an owned `subagent_run` tool with `executionMode: sequential`; its async
  `execute` does short admission + revision waits and resolves with all results.
- Keep the Pi RPC reader running for UI requests, Stop, child completion and stats.
  Child execution is Go-managed, not `pi-subagents` or an SDK child of Pi.
- Enforce a parent barrier at tool dispatch and steering/model boundaries before
  bypass/YOLO logic. While joining, permit only the owning tool's transport and
  human request replies; new unrelated tool admissions fail/park, not approve.
- A Promise rejection does not drop ownership. First cancel/join/drain in the
  runtime, then release the native failure result or terminate the parent.
- Refresh only at a complete engine-turn boundary; internal wait/compaction does
  not capture new agent or Library selections.

### OpenCode 1.18.35

- Supply an owned **local plugin tool** named `subagent_run`, using the pinned
  native plugin tool registration/schema mechanism. Its Promise stays pending;
  only its internal short RPCs have transport deadlines. Do not expose the same
  operation simultaneously as an MCP tool and a native tool.
- Extend owned before-tool/shell hooks with the same parent barrier. Extend
  `experimental.chat.messages.transform` to check the barrier before taking
  steering or dispatching the next model snapshot. A second parallel call cannot
  bypass the engine's batch reservation.
- Verify effective inventory, plugin handshake, denied native `task` and child
  leaf restrictions before the first prompt, after YOLO profile rewrites.
- CP1a found config-only task suppression insufficient on the installed pin.
  The proven fixture uses the public `chat.message` hook to set
  `output.message.tools.task=false` before persistence, plus an unconditional
  pre-execution rejection. Future wiring must preserve that per-input restriction
  through steering/native continuation and verify the actual model surface, not
  confuse `/experimental/tool/ids` registry presence with model exposure.
- Do not return the existing idle path at `opencode_interactive.go:1057` or final
  reconciliation while an invocation/delivery is outstanding. Unexpected idle is
  failure + cancellation/drain, not a synthetic successful delegation result.
- Confirm in the pinned plugin API that local tool execution has no shorter
  implicit deadline. If this fails, CP1 stops at a compatibility blocker rather
  than silently returning an early receipt.

### Codex 0.153.4: app-server dynamic tool, not MCP

The installed binary's experimental JSON schema was exported during CP0:

- `ThreadStartParams.dynamicTools` accepts function specs with `name`,
  `description`, `inputSchema`, `type: function`, optional `deferLoading`.
- `ServerRequest` includes `item/tool/call`.
- `DynamicToolCallParams` includes `threadId`, `turnId`, `callId`, `tool`,
  `arguments` and optional namespace.
- `DynamicToolCallResponse` requires `contentItems` (`inputText` for our JSON
  result) and `success`.

Register `subagent_run` as a direct dynamic function at new thread creation.
Handle its `item/tool/call` asynchronously in the existing native reader. Retain
the exact JSON-RPC response ID; send the dynamic response only after the batch
settles. The reader must continue processing permissions/questions and Stop rather
than blocking on child completion. This wait does not traverse MCP's timeout.

**Required CP1 feasibility proof:** a bounded disposable protocol fixture must
show that an outstanding dynamic tool request prevents the next model interaction
and native successful completion, survives a wait longer than the ordinary linked
MCP deadline, and can be cancelled without automatic re-execution. Schema presence
alone does not establish timeout/cancellation/scheduling behavior. Native turn
timeouts or parent process death must cancel/drain all children and fail the
batch, never release the parent to continue independently. Do not increase linked
MCP timeouts as a substitute.

CP1a continuation evidence now confirms this pending wait/result/cancel/resume
mechanism in a deterministic local-provider fixture for threads born with the
tool: one 155,010ms wait, one consolidated output, same-ID persisted tool spec
across process restart and no automatic replay after pending interruption. It does
not establish existing-thread adoption or engine-owned child/graph integration.
Persisted tool specs are immutable native history context: fresh per-process
handlers must support the stored name/schema rather than assume that redeclaring
a new schema at thread start will update a resumed thread. Full fixture evidence
and spec digest are in `AGENTS_UNIFICATION_DELEGATION_PROGRESS.md`.

**Existing-thread compatibility boundary:** the exported `ThreadResumeParams` does NOT
advertise `dynamicTools`, and the exported client request catalog has no obvious
set-dynamic-tools method. Thread creation support does not establish registration
on an existing persisted native thread. CP1a's directed exact-pin investigation
did not identify a supported config/update seam. The manager obtained the user's
explicit **new native threads only** decision (manager human message reference
`85e3f4d3845b2f7b80dd2e1e812c0f90`, not a local authorization event). Preserve old
native histories unchanged, expose the concise capability/detail limitation, and
register/persist compatible v1 handlers only for new native threads. Do not send
undocumented fields, fork/replace histories, upgrade pins or reopen adoption research.

**Parallel-call boundary:** a pending native dynamic tool blocks the *next model
iteration*, not necessarily sibling built-ins already emitted in the same native
response. Codex does not offer the Pi/OpenCode pre-tool gate. Track and drain those
siblings; do not describe dynamic tools as a universal pre-execution lock. Multiple
dynamic delegation calls still serialize through the engine reservation. Choice
cannot be reserved/accepted during the batch, even if emitted as a sibling call.
Manager approved this boundary: same-response siblings may remain in flight and
must be tracked/drained; no later model decision, partial result, independent batch
or Choice is allowed before the consolidated join. Characterize in the probe if
an already emitted sibling begins only after admission; do not claim zero effects
after admission. Interrupt-after-receipt has a race and is not a substitute.

### Pause/continuation alternatives considered

- `linked_ask`-style yield plus a later prompt is insufficient: native continuation
  remains possible before settlement and delivery is then a new scheduling problem.
- Pi abort / OpenCode abort / Codex interrupt followed by resume is useful for
  **cancellation/recovery**, not the primary wait. It risks incomplete tool/result
  pairs, replayed calls and queued steering. Do not automatically restart after an
  ambiguous interrupt or transport failure.
- Suspending only a parent PID is not a harness/tool/child-process seal. Killing the
  parent before dispatching children avoids overlap but loses safe native result
  delivery; no such fallback is approved by this contract.
- MCP progress/HTTP keepalive is not proof of extending native client deadlines.
  Keep execution waits outside MCP; resource/result reads remain bounded.

## 5. Durable records, state transitions and receipts

Proposed additive domains (types live in `delegation_records.go`, not `store.go`):

| Domain | Required data |
| --- | --- |
| `delegation_batches` | ID, parent/owner/turn IDs, operation ID, canonical input hash, ordered child keys, parent native correlation, graph ownership, status, revision, timestamps, cancellation/error, delivery state and digest. |
| `delegation_invocations` | ID, batch/key/child IDs, fixed harness/cwd, A's captured resolved config/binding revision, external snapshot reference, execution state, native acceptance/turn ID, task outcome, result/error, process/tool finality evidence, timestamps. |
| `delegation_children` | Child session ID, exact parent ID, managed-vs-legacy provenance, creation identity, reference to A's session-owned binding/config record, fixed harness/cwd, graph run/workspace restriction and active invocation lease. No duplicate binding/config authority. |

Histories stay in `session_events`; native identity stays in `native_sessions`.
Prompts/results and snapshots use the existing chunk/content machinery rather
than a growing array in one session metadata row. Version/revision counters are
decimal strings in public JSON. Capture task, resolved settings and child identity
before any native prompt dispatch.

State machines:

```text
batch: admitted -> preparing -> running -> joining -> settled
                          \-> cancelling -> joining -> settled
delivery: pending -> sending -> delivered
                       \-> uncertain
          pending -> cancelled             (Stop/restart before delivery)

invocation: admitted -> preparing -> running -> draining -> terminal
                         \-> draining ---------------------> terminal
```

`terminal` carries an explicit execution outcome:
`completed | failed | cancelled | interrupted | finality_unconfirmed`.
Admission/preparation failure uses `failed`, with native acceptance false. Track
task outcome separately: `satisfied | unsatisfied | unknown`. Normal native
completion establishes only execution completion, not semantic satisfaction.

`taskOutcome` is optional and is supplied only when the actual execution provides
an explicit attributable semantic outcome. No new report/classification tool,
evaluation loop or completion reminder is introduced to populate this field.
Absent evidence means unavailable, not invented success or inferred failure from
assistant wording. The parent assesses the returned work. Native execution errors
remain factual errors regardless of any available semantic outcome.

A settled batch always returns all children in original key order. Derive
`outcome=completed` only if all executions completed; otherwise preserve
`failed/cancelled/interrupted/finality_unconfirmed` with deterministic precedence
(uncertain finality first). Include every per-child outcome; mixed results never
collapse into a successful tool response. Native `success=false`/`isError=true`
for execution failures or reported semantic failure, while retaining full JSON
results so the parent can recover. `completed` must not be named `succeeded` when
task satisfaction is unavailable or unknown.

Idempotency:

- Unique `(parentSessionId, operationId)` receipt for the logical request, across
  parent turns; exact canonical request retries return/attach to the original
  batch and result. Changed payload under that key conflicts. A new intentional
  task/retry needs a new operation ID; never dedupe by task text alone.
- Also persist native `(parentTurnId, nativeSession/turn/callId)` correlation so
  reconnect/replayed native tool events cannot start duplicate batches.
- Reserve one active invocation per child atomically. The same child cannot occur
  twice in a batch, or run concurrently through another operation ID.
- Commit admission and visible cards before dispatch; commit terminal result and
  result digest before releasing its wrapper. Sending a native result is not
  confirmation that it entered native history. Mark delivered only after the
  adapter's correlated tool-result event/persisted-history evidence.
- Ambiguous delivery is `uncertain`. Do not auto-reprompt/replay it; a later explicit
  parent recovery can read the existing batch/result under its exact IDs. Never
  promise exactly-once model consumption across process loss; promise no automatic
  duplicate execution and durable correlated outcomes.

## 6. Resolver seam, prompts and capability isolation

A's CP0 contract became available during this inspection. B adopts its proposed
names/types rather than introducing a parallel resolver API:

```go
func (a *app) resolveNewAgent(ctx context.Context, req NewAgentRequest) (AgentResolution, error)
func (a *app) resolveBoundAgent(ctx context.Context, req BoundAgentRequest) (AgentResolution, error)
func (a *app) prepareAgentTurn(ctx context.Context, sessionID string,
    frozen *AgentExecutionCapture) (*AgentTurnPreparation, error)
```

Reuse A's `AgentSourceRef`, `SessionAgentBinding`, `AgentConfigOverrides`,
`ResolvedAgentConfig`, `AgentCapture` and captured external snapshot. Graph requests
pass A's optional frozen catalog; child invocation dispatch supplies the frozen
`AgentExecutionCapture` to `prepareAgentTurn`. Commit A's optional `agent_binding`
through its helper when creating a bound child; store only references in B's child
ownership ledger. A's lock order (authoring, Library, app, SQL writer) applies.

Required output: stable source incarnation (not slug alone), revision/digest,
source lifecycle status, prompt/description, fixed harness, validated model/effort,
merged external selection, persisted explicit overrides and last accepted config.
Pointers/optional fields preserve absence versus explicit empty/reset values.

Integration semantics: skill IDs and `mcpConnections` are independent override
fields; a skill edit must not freeze/reset MCPs, and vice versa. For PATCH,
omitted=unchanged, null=reset-to-agent and empty list=explicit select-none. Preserve
presence separately from the value; do not use a plain Go pointer decoder to
conflate omitted with null. This manager decision supersedes earlier candidate
`reset` spellings or a single full-Selection override in CP0 drafts. Inheritance
remains all-or-none, external only, false by default, explicit-human opt-in, and
creation-only: reuse does not add another parent contribution or live-link changes.

- Saved specialist: resolve through A; use the specialist's Library by default.
- Unnamed new child: fallback uses the parent's resolved harness/model/effort and
  a dedicated general child prompt. Cross-harness explicit override resolves target
  harness defaults rather than copying an incompatible model ID. External Library
  defaults to **empty**, with no internal roles/skills inherited.
- `inheritParentCapabilities=true`: only when the human explicitly requested full
  external inheritance. It adds all of the parent's applied external contribution
  (saved selection before any application), captured once, not individual toggles
  or a live link. Internal skills/tokens/tools are always excluded. This field is
  intent/traceability, not semantic proof that Library content authorized it.
- Reuse: refresh bound definition only between invocations/complete parent turns,
  retain fixed harness and explicit overrides. Incompatible model/effort blocks
  admission. Only confirmed disabled/deleted sources use A's validated last capture;
  new creation from them fails. Unknown source I/O, invalid source and conflicts
  retain stored capture for history but BLOCK preparation, not activate deleted
  fallback. No silent same-slug rebind after recreation. A single model/prompt/
  external-Library execution capture is reused throughout this invocation's wait/
  native continuation; no resolver refresh during the join.

Child prompt composition: engine-owned leaf role/scope and outcome instructions +
captured specialist prompt (or unnamed fallback) + explicit self-contained task +
selected external Library context. Parent history, orchestration prompt, topic
classification instructions, general-agent installation manager, graph Choice
contract and human event IDs are not injected as inherited child context.
Repository instructions remain available through the normal harness policy;
this is not a repository/filesystem sandbox.

Each child gets its own adapter, capability snapshot, MCP execution/gateway and
private per-invocation token. Never reuse `p.mcp`, a parent bridge token or parent
adapter pointer to implement child tool dispatch. Engine-to-provider credentials
remain handled by the selected harness's normal engine-user integration, not
copied from the parent's prompt/environment/profile. Build child environment from
approved harness inputs; strip parent KLM turn/role/graph/gateway credentials.

Leaf catalog is selected external Library and native user question facilities.
No linked/session/general/project/graph/automation/delegation
tools, internal skill manager, dynamic parent tool or discovery-to-admin route.
Enforce both catalog filtering and authenticated role checks, even under YOLO.

## 7. Stop, requests, queue, restart and reuse

- Derive child turn cancellation from the live parent turn, not a transient HTTP
  request. Stop cancels the batch, closes further admission, cancels all child
  turns and shuts down every child process domain and MCP worker. Keep ownership
  until all teardowns have returned or a bounded infrastructure deadline records
  finality uncertainty. Native interrupted/idle is insufficient shutdown evidence.
- Each child invocation ends its process container. This intentionally prevents
  a child dev server/browser from continuing as unjoined background work. Preserve
  the reusable native history and identity; a later invocation starts a new managed
  profile. Child process termination does not prove remote task cancellation.
- Join cannot be a blocking loop holding `app.mu`, graph binding mutex, a SQL write
  transaction or the native frame reader. Persist under locks, execute/wait outside
  locks; completion callbacks wake the join by IDs. Publish events after commit.
- Send now follows existing replacement semantics but waits for **parent and all
  children**. Aggregate child stop errors into `t.stopErr`. Any uncertainty blocks
  replacement; explicit Stop clears a pending replacement. Do not revive a stopped
  batch when a late child response arrives.
- Ordinary messages accepted during join stay queued for the next native model
  boundary after batch result settlement. They do not steer children, answer human
  requests, change snapshots or trigger a second parent turn. The parent remains
  running/working while waiting; automation/consultation schedulers cannot claim it.
- Preserve archive-as-metadata semantics. Archive is not Stop and does not make
  a child independent. A batch remains bound to its already running parent.
- Route permission/question records through `ownerConversationId`, with explicit
  `{childSessionId, invocationId, batchId, parentSessionId, graphRunId?, activationId?}`
  attribution. Native replies execute on the exact child turn/source request,
  not whichever child currently has the same tool name. Keep native grant scope
  and engine-visible owner/session/project/global rules intact. General children
  have no project grant lifetime. YOLO never answers questions or elicitation.
- Parent live permission policy applies to children. A Codex child that started
  danger-full-access also participates in the existing no-safe-downgrade rule:
  reject disabling parent YOLO until that child ends; do not falsely show policy
  narrowed while the native child still has the wider sandbox.
- Restart recovery runs before readiness: nonterminal invocations become interrupted
  (or finality_unconfirmed when teardown/remote evidence is unresolved); terminal
  child results remain immutable. Mark pending/sending delivery interrupted or
  uncertain as appropriate and pause parent unsent queue. No automatic child
  dispatch, native tool replay, result prompt delivery or replacement turn occurs.
- Reuse after failure/restart requires an explicit new parent invocation with an
  exact owned child ID and a new operation ID. Reuse is rejected if history/cwd/
  harness/ownership cannot be validated. Old operation IDs remain read-only receipts.

## 8. Graph snapshot and sealing contract

Approved integration direction: **snapshot v2** captures the COMPLETE visible
eligible delegation catalog at actual run start: global agents plus the run
project's local agents, with source identities, exact captured configuration and
external capability captures. It is not restricted to agents referenced by nodes,
not captured when scheduling, and never supplemented by live lookup later.

Use A's separate `delegationAgents` map keyed by captured slug plus source UID/
run-local source key; retain node `Agents`/`EffectiveAgents` invariants. V2's manifest
includes a version, catalog digest, source revision/identity, global/local logical
locator, definition content hash, external snapshot hash, completeness and bounded
diagnostics. No credentials, live tokens or absolute private source paths. Local
legacy files need no identity-adoption write to be captured: use the run-scoped key.

Concrete bounds proposal for the manager/A to integrate: at most 256 eligible
definitions and 16 MiB aggregate canonical definition/prompt bytes, retaining the
existing per-definition/per-prompt bounds. Reuse the existing bounded capability
capture and content-addressed skill machinery; do not inline duplicate skill trees.
At most 100 diagnostic entries plus a truncation flag/count (diagnostics only may
truncate). An eligible-catalog overflow fails v2 preflight with a structured
`delegation_catalog_too_large` error; never truncate eligible entries silently.

Unavailable/disabled/invalid/conflicting candidates are frozen as diagnostics,
not substituted into the eligible map. Referenced node agents retain strict
preflight errors; an unrelated unusable candidate need not break an otherwise
valid graph. Unknown roots/source I/O must set completeness false and diagnostics;
the snapshot must not claim that unknown candidates were eligible or absent.
Discovery reports the frozen incompleteness and selection of a missing/diagnosed
specialist fails without checking the live catalog. Thus every child either uses
its exact admitted run capture or fails, independent of later source changes.

Preserve v1 bytes, validation and hashes; absent version remains v1. Never relabel
or backfill historical snapshots from live authoring. V1 has no complete delegation
catalog: named specialist delegation requiring v2 is explicitly unavailable there,
while historical node/child reads remain unchanged. This is backward-compatible
history handling, not permission to create new v1 runs for this feature.

B's graph child resolution uses only that captured artifact. No new skill or MCP
selection is captured at child start. Explicit model/effort overrides still validate
against the fixed child harness; their resolved values are persisted per invocation.
Inherited external contribution is the parent's run-captured external snapshot,
only under the same explicit opt-in. Ordinary reusable bindings update between
turns; graph run snapshots never follow agent edits mid-run.

Graph-child sessions remain `role=subagent`, `ParentID=<node session>` and carry a
separate managed-child ownership record. Do not label them graph nodes or hand
them `GraphAdapterHooks.Node=true`. They have no Choice capability. Every invocation
records current activation/turn/workspace/snapshot identity. A reusable node session
may reuse its child only in the same run/workspace, while rebinding each invocation
to the current activation. No cross-run/workspace reuse by recency.

Use a leaf execution-domain binding distinct from the node Choice binding so every
child adapter (including Codex) uses immediate owned-process teardown, graph-compatible
tool boundary tracking and external MCP drainage. Current `role != graph_node`
runtime retention and `graphNode()` shortcuts need explicit integration: setting a
graph parent pointer alone is neither containment nor a gate.

Admission and Choice reservation share a linearization point:

1. If a child batch/delivery is active, `graph_submit_choice` rejects without
   reserving/sealing, including parallel native calls. Parent must consume the full
   batch result first. Never block Choice while holding the graph lock.
2. If Choice is already reserved/sealed, reject new child admission.
3. Choice acceptance requires parent native/tool settlement + parent process drain
   + all child invocation terminals + confirmed child process/MCP/tool drainage +
   no delivery pending + no unresolved child finality evidence.
4. Extend `GraphAdapterResult`/persisted activation evidence to carry child-settlement
   evidence and uncertainty. Query the durable activation ledger at reservation
   and final transitions; a parent-local empty map is not proof of no children.
   Recheck v2 catalog/snapshot digest and activation epoch at admission; definition
   edits, Library edits and source deletion after run start cannot reopen or change
   an already captured activation/child. Replay of an old Choice cannot attach to
   a later activation merely because the node session/history is reused.
5. A failed child task can be reported and handled by the parent; it need not make
   every later Choice impossible. **Uncertain physical/tool finality does** block
   acceptance and must remain `finality_unconfirmed`, not a parent decision override.

Chat Stop still does not stop independent graphs. Graph ending/shutdown cancels its
activation children and joins them before releasing workspace usage or dispatching
successor activations. `CheckGraphAdapter` remains mandatory for every harness used
inside the activation; ordinary Linux managed chats do not imply Linux graph support.
Do not broaden `GraphAdapterCapabilities` or label any path runtime-validated from
schema export/compilation. Returned MCP responses still do not prove detached remote
work finished; retain existing limitations.

## 9. Shared frontend/API event contract requested from C

Reuse `Event.type='subagent'` per invocation, with stable ID
`delegation/<invocationId>`, `childSessionId` and explicit data:

```ts
type DelegationInvocationData = {
  managed: true;
  batchId: string; invocationId: string; childSessionId: string;
  parentSessionId: string; ownerConversationId: string; key: string;
  harness: 'pi' | 'opencode' | 'codex'; model: string; effort: string;
  agentId?: string; agentRevision?: string;
  executionState: 'admitted' | 'preparing' | 'running' | 'draining' | 'terminal';
  outcome?: 'completed' | 'failed' | 'cancelled' | 'interrupted' | 'finality_unconfirmed';
  taskOutcome?: 'unknown' | 'satisfied' | 'unsatisfied';
  error?: string;
};
type DelegationBatchView = {
  id: string; revision: string; parentSessionId: string;
  state: 'admitted' | 'preparing' | 'running' | 'joining' | 'cancelling' | 'settled';
  total: number; terminal: number; remaining: number; waiting: boolean;
  delivery: 'pending' | 'sending' | 'delivered' | 'uncertain' | 'cancelled';
  outcome?: DelegationInvocationData['outcome'];
};
```

B proposes the following concrete CP0 transport for manager approval (not multiple
candidate authorities):

- `Session.delegation?: DelegationBatchView & { actions: DelegationActions }` is a
  compact active-batch projection, without member transcripts or member arrays.
- `DelegationActions` has booleans `send`, `queue`, `steer`, `interrupt`, `settings`,
  `stop`; native execution eligibility is still checked by the engine. `settings`
  means the existing settings action, not permission to mutate an active snapshot.
- Parent history includes one `subagent` batch event with
  `data={managed:true,kind:'batch',...DelegationBatchView}` and stable ID
  `delegation-batch/<batchId>`, plus the invocation events above with
  `data.kind='invocation'`. Preserve the legacy `childSessionId` branch unchanged.
- `GET /api/sessions/{ownerId}/delegations?offset=&limit=` lists batches for the
  exact parent (normal, side or general). Response:
  `{items:DelegationBatchView[],revision:string,hasMore:boolean,nextOffset:number}`.
  Newest-created first with stable ID ties; default/max page size 50.
- `GET /api/sessions/{ownerId}/delegations/{batchId}?offset=&limit=` returns
  `{batch:DelegationBatchView,items:DelegationChildView[],revision:string,hasMore:boolean,nextOffset:number}`,
  members in original request key order. Child view extends invocation data with
  title, `invocationNumber`, `reused`, `terminal:boolean`, optional latest activity
  and graph IDs; large task/result content is absent from this summary.
- Child metadata/history for normal/side/general parents uses existing exact-ID
  `/api/sessions/{childId}` and `/history` reads with validated ownership. No child
  composer, queue or public execution route is introduced.
- For graph children, use explicit owner-validated reads under
  `/api/sessions/{ownerId}/graph/runs/{runId}/activations/{activationId}/delegations`
  and `.../delegations/{batchId}` with the same paging envelopes. Exact child
  metadata/history uses `.../delegations/{batchId}/children/{childId}` and
  `.../children/{childId}/history` (existing history page/content-descriptor shape).
  These are proposed additive graph-owned routes, not arbitrary graph conversation
  access. Existing node activity includes batch/child navigation IDs for each
  activation. Validate run/activation/batch/child relations and project availability.

Use the raw execution fields above as the authority. C's visible `needs_input`
derives from exact pending requests; `stopping` from cancellation/drain;
`execution_failed`/`shutdown_unconfirmed` map to failed/uncertain execution.
`task_failed` is available only if an actual explicit `taskOutcome=unsatisfied`
exists; B adds no evaluator or new tool to manufacture it.

Extend existing pending Permission/Question projections with this typed origin;
native callback/request storage remains singular:

```ts
type DelegatedRequestOrigin = {
  type: 'subagent';
  kind: 'permission' | 'question'; // elicitation remains a question payload
  replySessionId: string; requestId: string;
  ownerConversationId: string; parentSessionId: string; childSessionId: string;
  childTitle: string; harness: 'pi' | 'opencode' | 'codex'; agentName?: string;
  batchId: string; invocationId: string;
  graph?: { runId: string; nodeId: string; nodeName: string; activationId: string };
};
```

The reply tuple is exactly `(replySessionId, requestId, kind)` from the projection.
Proposed implementation: replySessionId is the visible owner route and the engine
resolves that exact tuple to its live child turn. Public clients do not infer target,
grant ownership or role from event text. Attribute requests from the admitted
invocation before any child history/metadata has been opened. Pending pages and
graph request projection include these origins/counts with existing bounded
pagination; reply invalidates owner metadata/request pages and child state.

Keep `Session.status` unchanged. During join, `waiting=true`, `remaining>0`, send/
queue allowed where the existing parent kind permits them, immediate native steering
paused, Stop allowed; Send now allowed only where the existing parent kind permits
it and always uses expanded child drain. Settings retain existing active-turn rules
and do not alter this batch. Graph-node/automation restrictions take precedence.
`remaining=0` with result delivery pending still means the parent is waiting; it
does not imply permission to run another model iteration or release a graph Choice.

Reuse storage contract 3 scoped invalidations and history cursors; batch/terminal
commits invalidate the owner metadata/history and changed child metadata/history,
plus graph activity when applicable. Do not broadcast complete histories or
credentials/snapshots. Child status cards must show actual failure/interruption/
uncertainty, not infer `completed` from non-running status. Legacy `subagent` cards
without `managed` remain readable with their historical native provenance.

## 10. File ownership and handoff proposal

Nothing in this section authorizes edits before the manager's release.

**B runtime new-file proposal:** `engine/delegation_records.go`, `delegation.go`,
`delegation_tools.go`, `delegation_bridge.go`, `delegation_prompts.go`,
`graph_delegation.go`, and owned local-tool wrapper
assets if needed (`pi-delegation.ts`, `opencode-delegation-plugin.mjs`). Keep the
runtime service small: admission, child adapter dispatch, join, result/receipt.

**B primary existing runtime files, after handoff:** `subagents.go`, `runtime.go`,
`pi_interactive.go`, `pi-permissions.ts`, `pi_subagents.go` (disable native loading),
`opencode_interactive.go`, `opencode-permissions-plugin.mjs`,
`opencode_permission_gate.go`, `opencode_library.go`, `codex_interactive.go`,
`codex_library.go`, `graph_adapter.go`, `graph_mcp_calls.go`,
`sqlite_graph_execution.go`, graph transitions/projections, `permissions.go`,
`questions.go`, `sqlite_execution_requests.go`, `sqlite_execution_stop.go`,
`sqlite_execution_finish.go`, `sqlite_execution_queue.go`, `steering.go` and
`sqlite_execution_context.go` as actually required. Existing native-history
projection is preserved; no wholesale deletion of old modules is planned.

**Manager-assigned ownership and required seams:**

| File/seam | Ownership / integration |
| --- | --- |
| `store.go` | A ONLY. B provides the record/root requirements in sections 5 and 10.1; no B edits. |
| `harness.go`, `main.go` | CP1b read-only for B. A temporarily owns thin resolver hooks in harness.go; main.go remains untouched. B runtime integration requires later explicit handoff/release. |
| `api.go`, `sqlite_api.go`, route catalogs | A owns persistence/API inserts. B supplies section 9's exact route/projection requirements, no edits to A-owned files. Runtime-only route helpers require later explicit assignment. |
| `linked_bridge.go`, `graph_adapter.go:bridgeTools`, `general_agent_tools.go` | A temporarily integrates private discovery/spawn in its owned files. B only adds unregistered protocol/schema assets now; actual delegation dispatch waits for explicit handoff. No child-report tool or duplicate exposure. |
| `capabilities.go`, `sqlite_execution_context.go` | A temporarily owns thin resolver/capture hooks. B does not edit them in CP1b. Later runtime gets one prepared invocation/run snapshot, no second live graph capture. |
| SQLite shared schema/store/domains/indexes/migration/base binding | A ONLY. B's concrete root/types/index requirements are in the handoff. Future isolated B delegation transaction/read/recovery files may be released in CP2; they are not automatically assigned to A. Existing stores upgrade additively, preserving history without JSON fallback. |
| Graph snapshots/catalog | A captures resolvable specialist definitions/external snapshots; B consumes that immutable bundle and owns activation child finality. Sequence the shared types. |
| `clients/desktop/src/engine.ts` and components | C owns frontend types/UI; B supplies JSON DTOs/projections and error codes. |

Validation invariants to give the migration owner: graph-child parent is a real
node session in the same project/run; exact workspace/cwd; child has only subagent
role, no own topic/graph selector/independent creation receipt; general empty-project
ownership is allowed only below the general parent; no subagent-as-parent; unique
active child invocation; admitted batch has all referenced identities; immutable
receipt hash and delivery correlation; legacy native children remain valid.

### 10.1 Persistence payload supplied to A (requirements, not released edits)

The three domains in section 5 are the requested root inventories. Proposed
`diskState` compatibility field names: `delegationBatches`, `delegationInvocations`,
`delegationChildren`; fresh-store/import defaults are empty, never reconstructed
from legacy native tool names. A chooses reflected roots versus explicit SQL
tables in the single approved migration, preserving existing storage conventions.

Minimal routing/indexed columns (large content stays in chunked records):

- Batches: `parentSessionId`, `ownerConversationId`, `parentTurnId`, `operationId`,
  `requestHash`, `state`, `delivery`, `revision`, `createdAt`, `updatedAt`, optional
  graph run/activation/turn/workspace IDs and `snapshotDigest`.
- Invocations: `batchId`, `key`, `childSessionId`, `parentSessionId`, `state`,
  `outcome`, `nativeAccepted`, native session/turn/call correlation, ordinal within
  batch, graph epoch/workspace references and `resultDigest`.
- Children: `childSessionId`, `parentSessionId`, `origin=engine`, fixed `harness`,
  canonical `executionCwd`, optional graph run/workspace restriction and active
  invocation ID/lease. Binding/overrides remain in A's session-owned record only.
- Large invocation payload fields: self-contained task, A's frozen
  `AgentExecutionCapture`/binding revision, external snapshot reference, result/
  error and structured finality evidence (`nativeSettled`, `processesDrained`,
  `toolCallsSettled`, unresolved external calls). No native/source credentials.
- Unique constraints: `(parentSessionId,operationId)` and
  `(parentTurnId,nativeSessionId,nativeTurnId,nativeCallId)` for batch receipt;
  `(batchId,key)` and `(batchId,ordinal)` for membership; one nonterminal invocation
  per child; single active batch per parent. Missing native correlation components
  use a validated explicit epoch/key rather than empty values that collide.
- Child history events need indexed batch/invocation/child correlation for bounded
  reads; do not add historical event rewrites merely to populate new columns.

Transaction boundaries A must expose to runtime callers: admit receipt + all child
identities + invocation reservations + parent cards atomically; CAS prepare/dispatch
ownership; terminal result + immutable digest + card update atomically; correlated
delivery transition; cancellation intent before physical teardown; finality evidence
after teardown. No native launch or model request inside a SQL writer transaction.
Runtime callbacks use exact IDs + parent/activation epoch and cannot update a reused
child's newer invocation. A persisted admission is not replay authority after crash.

Recovery order requirement: after existing graph/automation recovery has established
interrupted owning epochs, settle delegation records in bounded pages before normal
session readiness/scheduling. Mark every nonterminal invocation interrupted, retain
unresolved finality evidence, settle batch/counters without success and pause parent
unsent queue. Release a lease for reuse only with confirmed physical/tool settlement;
uncertain children remain quarantined even if their recorded outcome is terminal.
Do not resolve that uncertainty by a user/parent assertion of success. Pending delivery becomes
cancelled/interrupted; sending without transcript confirmation becomes uncertain.
Already terminal results/delivered receipts are immutable. Repeat recovery is
idempotent and launches no harnesses, tools, prompts or MCPs. If graph recovery needs
child evidence earlier, collect it in a read-only prepass and finalize in the same
ordered recovery coordinator; do not let graph recovery accept a stale Choice.

Validation additions: every engine child ownership row refers to a `subagent` session
and an allowed non-leaf parent of the same project (or explicit general empty-project
case); graph ownership matches existing run/node/session plus exact activation/turn/
workspace/snapshot. Subagents have no own graph Choice/selection, topic or independent
spawn receipt. Batch count/order/parent references and result/delivery digests agree;
terminal success cannot coexist with unresolved drain evidence; bound source UID is
not inferred from slug. Keep old native children valid/readable without engine-owned
rows and without claiming managed admission/finality they never recorded.

### 10.2 Historical CP0 shape sketch (superseded by concrete CP1b source)

This initial sketch is retained as design history. Use actual
`engine/delegation_records.go` and the dedicated persistence handoff for current
fields/types and ownership. The following were not compiled runtime types or an
alternate resolver. They use strings/standard `encoding/json.RawMessage` so A can
integrate independently of runtime types not yet implemented. A owns the eventual
concrete record/type names, migration and writers. `ExecutionCapture` serializes
A's resolved immutable capture exactly; B does not reinterpret it as another agent
configuration authority. Large fields use existing chunked storage.

```go
type DelegationGraphScope struct {
    RunID, NodeID, ActivationID, TurnID, WorkspaceID, SnapshotDigest string
}
type DelegationNativeCall struct {
    ParentTurnID, SessionID, TurnID, CallID string
}
type DelegationBatchRecord struct {
    ID, ParentSessionID, OwnerConversationID, ParentTurnID string
    OperationID, RequestHash, Revision string
    NativeCall DelegationNativeCall
    Graph *DelegationGraphScope
    State, Delivery, Outcome string
    ExpectedChildren int
    CreatedAt, UpdatedAt, EndedAt string
    ResultDigest, Error string
    CancellationRequested bool
}
type DelegationInvocationRecord struct {
    ID, BatchID, Key, ChildSessionID, ParentSessionID string
    Ordinal int
    Graph *DelegationGraphScope
    State, Outcome, TaskOutcome string // TaskOutcome optional/unknown, no evaluator
    Harness, ExecutionCWD, BindingRevision string
    ExecutionCapture json.RawMessage // frozen A capture; opaque at this seam
    Task, Result, Error, ResultDigest string
    NativeAccepted bool
    NativeSessionID, NativeTurnID, NativeCallID string
    NativeSettled, ProcessesDrained, ToolCallsSettled bool
    UncertainCalls json.RawMessage // bounded existing tool-finality observations
    CreatedAt, StartedAt, EndedAt string
}
type DelegationChildRecord struct {
    ChildSessionID, ParentSessionID, Origin string // Origin="engine"
    Harness, ExecutionCWD string
    BindingOwnerSessionID string // same child; A's optional agent_binding authority
    GraphRunID, WorkspaceID string // optional reuse restrictions, not a Choice grant
    ActiveInvocationID string
    Quarantined bool // finality unknown forbids reuse, even after terminal record
    CreatedAt string
}
```

The unique/index/recovery rules of 10.1 are normative for this handoff. Per-native
call retry correlation may need a small batch-owned receipt/correlation child
collection so multiple native delivery attempts map to the same operation without
overwriting the first receipt. A must retain accepted call aliases rather than
treating a new MCP/RPC response ID as a new task. No history-sized arrays in batch
metadata; memberships/results remain separate invocation rows.

## 11. Evidence, risks and requested release

### CP0 lightweight evidence

Commands executed (all bounded, no model or server execution):

1. `git status --short` — recorded pre-existing unrelated changes/untracked documents.
2. `Get-Command codex -ErrorAction SilentlyContinue | Format-List Name,Source,CommandType`
   — found `C:\Users\Samuel\AppData\Local\Programs\OpenAI\Codex\bin\codex.exe`.
3. `& "C:\Users\Samuel\AppData\Local\Programs\OpenAI\Codex\bin\codex.exe" --version`
   — `codex-cli 0.153.4`.
4. Same executable with `app-server generate-json-schema --help` — supported bounded
   protocol export, with experimental fields optional.
5. After `Test-Path -LiteralPath "C:\Users\Samuel\AppData\Local\Temp\opencode"`,
   exported with `app-server generate-json-schema --experimental --out
   "C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-cp0-codex-01534-schema"`
   — exit 0. Read exported ThreadStart/ThreadResume, ServerRequest and dynamic
   request/response schemas. Temporary generated files are outside the repository.
6. `git diff --no-index --check -- NUL "AGENTS_UNIFICATION_DELEGATION_CONTRACT.md"`
   — no whitespace errors (only Git's normal LF/CRLF normalization notice).

Source reads also covered the product/plan/instructions and the listed runtime,
SQLite, graph, permission and UI seams, plus A/C's newly available CP0 contracts.
Concurrent source changes observed in final status belong to other workstreams;
B changed only this document and did not inspect/revert their in-flight diffs.
Upstream `main` source was reference-only;
it is not evidence for compatibility with the installed pin. No build/tests were
needed for this documentation-only CP0; no runtime behavior has been validated.

### Risks that must not be hidden by a successful build

1. Codex schema permits dynamic tools at start, not on existing-thread adoption.
   The approved compatibility boundary keeps old native histories without this
   capability; never lose that distinction in creation/resume/availability wiring.
2. Dynamic tools do not prove universal serial tool dispatch; same-response siblings
   on Codex need explicit semantics and real fixture evidence. There is no existing
   fail-closed all-tools hook to claim otherwise.
3. Native/local tool timeout or unexpected parent finality must cancel/join, never
   convert the system into fire-and-forget. Verify actual pin behavior before enable.
4. Existing SQLite stores will not acquire new domains by editing creation DDL;
   A/B must share one real additive migration and ownership validator.
5. Reusing ordinary child runtimes would retain workloads and violate all-results
   join/graph drain. Child teardown evidence must aggregate into parent Stop/Choice.
6. Reusing a parent gateway/token/policy adapter leaks role powers or misattributes
   requests. New child execution needs independently bound capabilities and exact
   owner projection; current child-target rejection cannot simply be removed.
7. YOLO OpenCode rewriting can re-enable native delegation. Codex child sandbox
   downgrade limits also affect the parent UI/control response.
8. Capturing only graph-node agents leaves dynamic saved-specialist selection live
   and mutable. Run-time capture of the resolvable catalog needs A's explicit seam.

### Current checkpoint / requested manager review

- Wait/sibling semantics and new-native-thread-only compatibility are approved.
  Do not imply a universal Codex pre-tool gate from the bounded fixture.
- Persistence ownership is now assigned to A; confirm section 10.1/10.2's payload/seams
  and remaining runtime ownership at release. Resolver integration decisions and
  complete frozen v2 catalog and unnamed empty-Library fallback are incorporated;
  proposed capture bounds still need approval.
- CP1b additive types/assets/pure helpers are delivered. The latest combined
  package check is blocked by a concurrent A-owned `sqlite_agents.go` json import
  error, documented in progress. Durable service/SQL/runtime activation is CP2
  and remains unreleased. No real graphs/upstream providers or product enable.

Worker stops after sending this CP0 report. Only an explicit manager instruction
releases further inspection/implementation; no polling or waiting command is used.
