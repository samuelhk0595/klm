# Agents unification — shared frontend CP0 contract

Status: CP0 UX, CP1a foundation and corrected CP1b management approved by manager.
CP2a manual creation/binding settings is implemented pending manager review.
Management routes/shapes below
match the reviewed backend. Binding/session/delegation sections remain proposals,
not claims about implemented behavior. Implementation/evidence is
tracked in `AGENTS_UNIFICATION_UI_PROGRESS.md`.

Owner: stream C, conversation `413fd7541b23359f48df60d266d93775`.
Manager: `a05369eaa5c9151b61e93fb38cb35c52`.
Authority: the approved product contract in
`AGENTS_UNIFICATION_IMPLEMENTATION_PLAN.md`, CP0 instruction and manager's explicit
CP1a release within that implementation scope.

## Manager review and CP1a boundary

- Approved: Settings -> Agents for globals without a project; project list combines
  globals+locals with origin; creation scope This project/Global; existing scope
  read-only. These screens are not part of the CP1a implementation release.
- Released CP1a only: production agent types/slug extraction, complete definition
  preservation on enable/disable and explicit demo-vs-production model sourcing.
- CP1a subsequently approved after manager inspected the complete diff. CP1b then
  released real global/local management, scope-aware authoring/model discovery,
  server pagination/search/exact details and existing scoped invalidation.
- CP1b excluded binding/creation/child runtime UI. CP2a specifically releases the
  delivered session creation/binding DTOs described in section 3.3 and their shared
  frontend integration. Child batch/graph-global UI remains unreleased. No backend
  changes or provider/browser probes are released to C.
- Approved catalog merge: public slug plus opaque `sourceUid`; globals stored in
  SQLite, locals remain TOML. A owns paginated/exact catalog reads, real global model
  discovery and the concrete invalidation key. Final wire shapes await A's report.
- Approved override semantics: omitted field = unchanged; `null` = reset/follow;
  explicit empty string/list = explicit empty override. Skills and MCP connections
  are independently overridden/reset. Server decoding is presence-aware. This
  supersedes CP0's candidate `set` / `reset` arrays.
- Deleted/disabled identity must come from A's authoritative source lifecycle.
  Catalog conflicts and unknown filesystem/I/O failures are not deletion. C must
  not infer fallback or execution eligibility from an error string/failed lookup.
  Unknown I/O, invalid or conflicted sources preserve captured history but block
  new preparation; confirmed deleted/disabled sources may continue their last
  validated configuration. Harness remains fixed.
- Child semantic results are rendered only if B actually supplies them. No new
  semantic classifier/evaluation is introduced to manufacture `task_failed`.
- Later Codex child-tool availability must use factual backend capability: only
  newly instrumented native threads can expose `subagent_run`; old native threads
  preserve history/normal chat without it, including an old general conversation.
  Harness/model alone cannot establish availability; no permanent tutorial is added.

## 1. Inspected CP0 baseline and concrete gaps

- `features/graphs/ProjectAuthoring.tsx` supplies `AuthoringProject` and a complete
  project authoring catalog to Agents/Graphs. `graphs/catalog.ts` caches by project
  and reloads after its own mutations; it currently ignores an empty project ID.
- `features/graphs/files.ts` owns `AgentFile`/`AuthoringCatalog`, while the production
  agent shape and slug helper originate in `features/agents/demo.ts`. Move shared
  production types/helpers to a real module; keep visual fixtures isolated.
- `AgentsPage.tsx` filters loaded project agents locally, derives all mutation URLs
  from the displayed project, and only knows string graph references. Its enable/
  disable payload currently omits skills/MCPs. Preserve the complete definition on
  toggles when wiring the new contract.
- `AgentEditorDialog.tsx` already edits prompt, harness, model, effort and Library.
  Library is hidden without `AuthoringProject`. `useAgentModels` currently returns
  visual fixture models without a project: that fallback must not feed global UI.
- `CreateSessionDialog.tsx` currently collects only title+harness. `App.tsx` posts
  those settings plus project/topic to `/api/sessions`; manual creation has no parent.
- `Session.model`/`effort` and `resolvedModel`/`resolvedEffort` do not distinguish
  agent-inherited settings from overrides. `ModelPicker` currently substitutes
  catalog defaults for unsupported values and saves model+effort together. This
  cannot serve the linked-agent incompatibility/recovery contract unchanged.
- `SessionLibraryDialog.tsx` uses revision-checked desired/applied selection and
  complete-turn pending copy. It has no agent-follow/reset state today.
- `SubagentView.tsx` reuses paginated `ConversationHistory` and shows model/effort;
  it is read-only. `ConversationEvents.SubagentActivity` treats all nonrunning
  activity as completed. New failure/cancellation/uncertainty must be explicit.
- `PermissionCard`/`QuestionCard` already accept `origin` and target a session+exact
  request ID; graph requests already have a separate projected origin. Keep those
  controls and their offered decisions rather than introducing another reply flow.
- `scopedConnection.ts` implements contract-3 scoped invalidation. Histories use
  bounded pages and decimal-string revisions; unopened child histories stay unloaded.
  `App.tsx`, `FocusSessionChat.tsx` and `SideChatPanel.tsx` are composition seams,
  including shared browser/mobile web presentations.

## 2. Screens and interactions

### 2.1 Global/local authoring without a new navigation system

Approved and implemented placement: an **Agents** section in global Settings
navigation, using the existing agents list/editor in a global authoring context.
Keep **Library** as the existing skill/MCP management surface. Settings is reachable
with no project selected, making global agent management independent of projects.
The project **Agents** page remains, displaying globals plus this project's locals.

- Every row and picker option shows **Global** or **Project · <name>** in secondary
  metadata, alongside harness/model/effort. Scope is server-supplied, not inferred
  from where the row is displayed. Disabled sources remain visible for management.
- New agent in project Agents defaults to this project and offers **Scope** with
  **This project** / **Global**. Global Settings creates global agents directly.
  Scope of an existing definition is read-only in this slice; migration/move is not
  implicitly implemented as delete+create.
- Editing a global row from a project targets its global URL. Inline scope metadata
  makes clear that this is the same shared definition. No cloning or shadowing.
- Keep the existing read-only **Identifier** slug preview. Known loaded duplicates
  provide immediate feedback, but server uniqueness is authoritative across every
  project/scope, including unloaded sources. Never download every project's catalog
  just to validate a name.
- A collision returns an actionable English error with the conflicting scope and
  ID. Existing duplicate/unreadable sources show real server diagnostics. The UI
  never chooses one conflicting definition or rewrites a name automatically.
- Preserve graph-reference blocking/rename behavior supplied by A. Reference rows
  need project/name/ID, not just a display string, for a global agent used in other
  projects. Linked sessions do not themselves block disabling/deleting a source:
  existing bindings retain their captured configuration per approved lifecycle.
- Render only real inventories. The global model picker requires an engine-scoped
  model catalog; it must not borrow an arbitrary project or use demo fixtures.

### 2.2 Manual session creation

Extend the existing dialog with a searchable **Agent** picker before Harness:

1. **Default** is selected initially and preserves the current title+harness flow.
2. The picker lists available global+current-project agents with origin and real
   configuration metadata. Disabled/conflicted/unavailable definitions cannot be
   selected for creation. Loading failure leaves Default usable and offers Retry.
3. Selecting an agent shows its defaults and an expandable **Overrides** area for
   harness, model, effort, prompt and external Library selection. Untouched fields
   are omitted from the request; displaying a default does not create an override.
4. A field edited by the user is marked **Override**, with **Use agent value** to
   clear that override. Agent changes recompute only untouched fields. Selecting
   Default removes the agent binding and its agent-only draft overrides; preserve
   title/topic and the explicit ordinary harness choice.
5. A precreation harness override is allowed. Its exact model/effort combination
   must be explicitly compatible before Create; do not silently translate a model
   ID or choose another effort when changing harness.
6. Creation captures the resulting harness permanently. Creation does not send a
   message/start work. Existing topic/no-topic navigation remains unchanged.
7. No parent-capability inheritance toggle is shown: manual creation has no parent.
   The independent-spawn tool uses the engine's explicit default-false boolean;
   this is not a manual session preference or a general Library default.

Show per-field origins only where they help edit/recover. Do not add explanatory
paragraphs to the ordinary Default flow.

### 2.3 Linked-session identity, overrides and recovery

Add a compact specialist identity action to Session options / status, opening a
shared **Agent settings** dialog. Reuse it in main, side and Focus/mobile surfaces
where a binding exists. Default/no-agent sessions keep their current controls.

- Show specialist name, slug, source scope, fixed harness and effective model/effort.
  Distinguish **Following agent**, **Overrides**, **Changes next turn**, and
  **Saved configuration** (source deleted/disabled/unavailable) from execution state.
- The dialog edits prompt/model/effort/Library overrides, not the saved definition.
  Provide per-field **Use agent value** and **Reset overrides**. Reset cannot switch
  the harness or silently rebind a different source generation. A captured-only
  source resets against its retained capture and remains labelled accordingly.
- ModelPicker must save only the fields deliberately changed. Choosing effort must
  not freeze the inherited model; a model change requiring an effort choice must
  ask for/explicitly capture that compatible effort. Empty string is an explicit
  harness-default override, distinct from following the agent.
- Library edits replace the session's explicit skill/MCP lists. Skills and MCPs have
  independent override presence; explicit empty lists mean select none. The session
  dialog has **Use agent skills** / **Use agent MCPs** reset actions. Do not mutate
  the source agent when saving session Library.
- Active turn shows its applied snapshot; future edits are labelled next-turn.
  The engine, not client timers/catalog refresh, resolves the boundary. A source
  harness change does not change an existing session's harness.
- If the next effective model/effort is incompatible with the fixed harness, show
  **Agent settings need attention** plus exact requested values and **Choose model**
  / effort recovery. Keep the invalid value visible. Block new execution locally
  using server eligibility, and keep Stop and exact pending-request replies usable.
  Engine admission must enforce the block for other clients/tools/queues as well.
- Do not silently resume a previously rejected send/uncertain operation after a
  recovery save. Preserve its draft and let the user explicitly send again.
- Deleted/disabled source is not itself an execution error for an existing binding;
  show saved-source state in details and retain the last capture. A new same-slug
  definition never changes that binding automatically. Ordinary source edits and
  confirmed deleted/disabled-source fallback require no new permission/approval flow.
  Conflict/unknown I/O must retain their distinct server status and authoritative
  eligibility; unknown I/O, invalid sources and conflicts block new preparation
  while retaining capture for history. A catalog failure alone never activates
  deleted-source fallback.

No new postcreation specialist switching/detach workflow is proposed in this slice.
Such operations would require a separately approved identity/history contract.

### 2.4 Engine-owned subagents

Keep child entry inside the parent history and the existing subagent tabs/Focus
back navigation. Children never become independent session rows or Happening now
cards, and get no manual composer, new-session, graph or delegation controls.

- A batch card displays **Subagents · 1/3 complete**, child specialist/title,
  harness, effective model/effort and factual current state. Parent shows
  **Waiting for subagents · 2 remaining** until all admitted children terminate.
- Child states distinguish preparing/running, **Needs input**, **Stopping**,
  **Completed**, **Task failed**, **Execution failed**, **Cancelled**,
  **Interrupted**, and **Shutdown unconfirmed** as exposed by B. Completion means
  terminal execution, not semantic task success; do not invent a successful outcome
  when the runtime has not assessed one.
- Opening a child loads only its bounded history and metadata. Show **Reused
  subagent** and invocation boundaries/tasks/results for multiple uses of the same
  child; one child's repeated use must not create multiple independent child tabs.
  Pending batch members can be listed from summaries without fetching transcripts.
- Keep native historical children readable through the existing role/path with a
  **Native subagent** origin when the backend can establish it. Unknown legacy
  origin stays unknown; missing new metadata must not fabricate engine-owned state.
- Parent Stop remains the visible stop action and cancels the active batch. Closing
  a child tab/sheet does not stop it. No foreground/background mode selector exists.
- Waiting/continuation and allowed queue/steering/Send now actions come from B's
  authoritative action flags. No frontend mechanism can release a parent early.
- Side/general parents use the same batch component and attribution. Graph-node
  activity must link a child within its run/activation context and expose projected
  requests in the owning chat. It must not require making a graph node a normal
  top-level session. Exact graph-child navigation/read seam is required from B.

### 2.5 Questions and permissions

Reuse PermissionCard, QuestionCard and MCPElicitationCard. Add a typed origin label:
**Subagent · <specialist/title> · <harness>**, with graph/node prefix when present.
Requests must remain answerable in the visible owning main/side/general chat even
when the child's tab/history is unopened. Requests arising before child metadata
must carry enough attribution to render immediately.

Every projection includes an exact reply target session/request pair. Do not send
a child request ID to an assumed parent route; do not infer grant ownership from
the visible chat. Dedupe by reply-target session ID + kind + request ID. Preserve
provided decisions, scope labels, elicitation semantics and human-only questions.
After a reply, invalidate the owning projection as well as the child, so parent
cards/counts cannot remain stale. Support bounded additional request pages.

## 3. A contract: implemented management and future binding

Management types/routes were read directly in `engine/agent_types.go`,
`engine/agent_catalog_api.go` and A's HTTP progress section. Sections 3.1–3.2 below
describe that delivered surface. Section 3.3 remains future binding design.

### 3.1 Catalog shapes

```ts
type AgentSource = {
  sourceUid: string; id: string; scope: 'global' | 'project';
  projectId?: string; projectName?: string;
};
type AgentReference = {
  kind: 'graph'; id: string; name: string; projectId: string; projectName: string;
};
type AgentRecord = AgentDefinition & {
  id: string; revision: string; source: AgentSource; status: string;
  available: boolean; unavailableReason?: string;
  graphReferences?: string[]; references: AgentReference[] | null;
  referencesHasMore?: boolean;
  referencesNextOffset?: number;
};
// AgentDefinition retains name, description, enabled, defaultHarness, model,
// effort, prompt, skills and mcpConnections (explicit connection/tool names).
type AgentPage = {
  revision: string; agents: AgentSummary[];
  diagnostics: AgentDiagnostic[] | null; complete: boolean;
  hasMore: boolean; nextOffset: number;
};
// Summary includes id/name/description/enabled/harness/model/effort/revision,
// source/status/available/unavailableReason and referenceCount; not prompt/Library.
// AgentDiagnostic contains code/message and optional source/conflicts.
```

`sourceUid` is the approved opaque anti-rebinding token, separate from the public
slug ID. Globals use SQLite and local definitions retain TOML. Rename continuity,
delete/recreate and external-file replacement must have defined behavior before C
wires actions. React row identity must include scope/source identity while conflicts
are being diagnosed, rather than keying duplicate legacy sources by slug alone.

### 3.2 Implemented HTTP surface

| Surface | Implemented route / behavior |
| --- | --- |
| Project visible agent list | `GET /api/agents?projectId={id}&search=&offset=&limit=`; globals + this project's locals, filtering before pagination |
| Global management | `GET /api/agents?search=&offset=&limit=`; globals only, works without projects |
| Exact source details | `GET /api/agents/{slug}?projectId=&sourceUid=&offset=&limit=`; full definition/source and bounded references; local source supplies projectId, all exact reads supply sourceUid |
| Global mutation | `POST /api/agents` with `{agent}`; `PATCH /api/agents/{slug}` with `{agent,revision,sourceUid}`; `DELETE` with `{revision,sourceUid}`; global update/delete require exact UID |
| Local mutation | `/api/projects/{id}/agents[/{slug}]`; same full definition/update identity body, preserving TOML compatibility; C always supplies sourceUid on identified mutations |
| Global models | `GET /api/models/{harness}` with existing ModelCatalog shape; no project directory required |
| Project models | Existing `GET /api/projects/{id}/models/{harness}`; exact authoring source scope chooses the endpoint |

`/api/projects/{id}/authoring` retains its complete arrays and now includes globals
plus additive source/status/reference metadata. New management uses its own bounded
summary list and exact detail reads rather than loading complete authoring arrays.
List/reference pages default to 50/max 100. C requests 50 and exposes Previous/Next;
exact details and More references are not restricted to the loaded summary page.

Graph Save/execution consumers are not global-enabled in this checkpoint. The
graph-authoring entry retains project-local agents (and legacy metadata-less rows)
so its existing picker cannot offer an unsupported new global reference. Agent
management never filters globals out of the project list. Read-only chat GraphView
still receives the complete authoring catalog. New global graph references,
execution/snapshots and cross-project diagnostic effects need the CP2 handoff;
none is claimed ready by this UI slice. Captured graph history is not rewritten.

Mutations provide structured error data in addition to English `error`:
`code`, `field?`, `hint?`, `conflicts?`, `currentRevision?`, `blockingReferences?` for
duplicate slug, stale source/revision, source unavailable, compatibility and graph
reference failures. EngineRequestError now retains the response details. Agent UI
shows the message, hint and source-scoped conflicts/graph references; stale editors
keep their original revision until explicit discard/reload. Generic errors remain
displayable for old routes. Saves/toggles/deletes are finite and never auto-retried.
Explicit discard uses an editor reset nonce after the exact read, clearing draft,
error and reference-page state even if source revision is unchanged. Incidental
SSE/list rerenders do not reset the editor. Reload reads are aborted/ignored after
leaving the editor/scope or changing engine identity.

### 3.3 Implemented CP2a session creation and binding transport

```ts
type AgentOverrides = {
  model?: string; effort?: string; prompt?: string;
  skills?: string[]; mcpConnections?: MCPSelection[];
};
type AgentOverridePatch = { [Field in keyof AgentOverrides]?: AgentOverrides[Field] | null };
type AgentConfiguration = {
  harness: Harness['id']; model: string; effort: string; prompt: string;
  selection: CapabilitySelection;
};
type SessionAgent = {
  revision: string; source: AgentSource;
  name: string;
  sourceStatus: string; // Preserve backend states; eligibility is independent.
  overrides: AgentOverrides;
  capturedRevision: string;
  desired: AgentConfiguration;
  pendingRevision?: string;
  overriddenFields: string[];
  inheritedCapabilities: CapabilitySelection;
  applied?: { config: AgentConfiguration; revision: string;
    source?: { name?: string; source: AgentSource; revision: string; config: AgentConfiguration } };
  pending: boolean;
  execution: { allowed: boolean; code?: string; message?: string;
    fields?: string[]; token: string };
};
type SessionCreateInput = {
  projectId: string; title: string; topicId?: string;
  harness?: Harness['id']; model?: string; effort?: string;
  agent?: string; agentSourceUid?: string; agentRevision?: string;
  agentOverrides?: AgentOverrides;
};
```

Concrete routes read in `agent_bindings.go`, `agent_resolution.go`, `agent_types.go`
and A's Definitive HTTP progress section:

- `POST /api/sessions`: fields above. C sends top-level harness only if explicitly
  overridden at creation, and uses nested `agentOverrides` for all chosen specialist
  fields. Untouched defaults are never materialized. No field is sent top-level and
  nested at once. Default retains the original title/topic/harness flow; no message
  is sent. No manual parent inheritance toggle exists. Exact source lookup must
  settle before Create is enabled, and unavailable sources cannot be chosen.
- `GET /api/sessions/{id}/agent` returns `{agent:null}` or `{agent:SessionAgent}`.
  `Session.agentConfig` is the compact summary without override prompts/selection.
  Existing mounted bound chats share detail reads; unopened sessions/children do
  not acquire histories or bindings through this UI. Applied is a nested execution
  capture, not a flat configuration.
- `PATCH /api/sessions/{id}/agent` sends `{revision,overrides}`. Omission means keep,
  null means follow/reset, empty string/list is explicit. Skills and MCPs remain
  independent. C's settings form saves only deliberately edited/reset fields;
  effort edits never pin model. Reset model/effort does not change fixed harness.
- Existing `/settings` pins fields supplied for bound sessions; existing full
  `/capabilities` PATCH pins both families. C routes bound model/Library UI through
  `/agent` instead, preventing accidental pinning of untouched fields. Unbound UI
  retains its current endpoints.

Source statuses include linked, disabled, deleted, source_invalid,
source_unavailable, conflict and exceptional source_replaced/captured. C preserves
them and uses execution.allowed/message/token; errors never become deleted state.
The composer/queue block new sending while settings are blocked/unavailable;
Stop, exact pending replies, drafts and queue removal remain usable. Recovery does
not replay a send or uncertain request. General identity/initial harness is unchanged.

Shared binding reads are keyed by engine identity and session, subscribe to the
existing global/project/conversation scopes, debounce/coalesce concurrent notices,
and do not apply configuration. Only explicit Save PATCHes do that. Mutation epochs
keep pre-save GETs from overwriting a saved configuration. Key change/unmount stops
readers. Source updates refresh the live desired/applied/eligibility presentation
while preserving an editor's edit-start base/draft until explicit discard/reload.
When running, bound model controls show applied.config; desired settings are labelled
next-turn. No model/effort is silently substituted if a pair becomes incompatible.

Library impact renders new `global_agent` and `agent_binding` reference kinds as
Global agent and Agent-linked session. Captured runs/active turns retain their
backend snapshots. Parent inheritance, when present in a binding, is displayed as
captured external contribution only; the UI neither enables it manually nor includes
internal role capabilities. Specialist prompt contents are never role authorization.

## 4. Proposed B contract: child lifecycle, event and request projections

```ts
type AgentScope = { kind: 'global' } |
  { kind: 'project'; projectId: string; projectName: string };
type ChildState = 'preparing' | 'running' | 'needs_input' | 'stopping' |
  'completed' | 'task_failed' | 'execution_failed' | 'cancelled' |
  'interrupted' | 'shutdown_unconfirmed';
type DelegationChildSummary = {
  sessionId: string; parentSessionId: string;
  origin: 'engine';
  title: string; agent?: { id: string; name: string; scope: AgentScope };
  harness: Harness['id']; model: string; effort: string;
  batchId: string; invocationId: string; invocationNumber: number;
  reused: boolean; state: ChildState; terminal: boolean;
  taskOutcome?: 'satisfied' | 'unsatisfied' | 'unknown';
  latestActivity?: string; error?: string;
  graph?: { runId: string; nodeId: string; activationId: string };
};
type DelegationBatchSummary = {
  id: string; parentSessionId: string; revision: string;
  state: string; total: number; terminal: number;
  waiting: boolean; children: DelegationChildSummary[];
  hasMore: boolean; nextOffset?: number;
};
type DelegatedRequestOrigin = {
  kind: 'subagent'; parentSessionId: string; childSessionId: string;
  childTitle: string; agentName?: string; harness: Harness['id'];
  batchId: string; invocationId: string;
  replySessionId: string; requestId: string;
  graph?: { runId: string; nodeId: string; nodeName: string; activationId: string };
};
```

Candidate reads: `GET /api/sessions/{parentId}/delegations?offset=&limit=` for
batches/current child summaries, exact `/delegations/{batchId}` for bounded members,
and existing child `/api/sessions/{childId}` + `/history` for transcripts. Invocation
task/result history can be normal child events with stable typed batch/invocation
data; a second transcript API is unnecessary. B must confirm the actual routes,
whether direct child GET is supported for every parent role, and graph activation
navigation. Do not infer ownership/access from untrusted event text.

Reuse `EngineEvent.type = 'subagent'` with a discriminated typed payload for
engine-owned batches/invocations. Preserve the legacy `childSessionId` event branch
unchanged. Each new payload needs stable batch/invocation/child IDs, explicit state,
terminal/outcome and a compact effective configuration. Large tasks/results use
existing event content descriptors, not oversized SSE payloads. Batch acceptance,
result delivery and settled completion must not be conflated into one event.

The parent's compact summary must include current waiting count and authoritative
action eligibility (send/queue/steer/interrupt/settings/Stop). Keep Session.status's
existing union where practical and add a delegation state instead of making child
waiting look idle. A parent can remain running while waiting, but the UI must not
display a model-thinking animation or label that misrepresents its wait state.

Requests may extend PermissionRequest/QuestionRequest with the typed origin above
or use a dedicated delegation request projection; choose one in the merge. Required:
exact reply routing, immediate attribution, parent counts, paginated overflow and
postreply parent invalidation. Graph projection must retain both graph and child
origins and project ownership. Questions/elicitation never become automatic answers.

## 5. Scoped updates and bounded loading

- Keep storage contract 3 and decimal-string database revisions. Authoring file
  revisions remain opaque equality tokens, not numeric revisions to BigInt-compare.
- Delivered catalog invalidation uses existing `global/` and
  `project/<projectId>` keys. Mounted project management/authoring watches both;
  global management watches `global/`. Binding invalidation is future work.
- Mounted catalogs subscribe through `subscribeScope`, retain their loaded window
  on refresh and refetch after reconnect/reset. Global scope cache keys must not
  collide with the hook's current empty-project no-op. Scope all new caches to
  engine identity/URL and clear them on engine switch.
- Management-page invalidations coalesce behind at most one in-flight read for
  the same key. Notices mark it dirty and cause one follow-up after settlement;
  the completed page is published and stays actionable during background refresh.
  Key changes/unmount abort and ignore the old read, including its queued refresh.
- B invalidates parent conversation metadata/history for batch state/request changes
  even when no child is open. Child conversation invalidation updates open history.
  Do not subscribe to every child or fetch histories to compute batch progress.
- Keep paginated history, requests, references and batch members. Server performs
  scoped search before pagination. Detail/creation controls must resolve a selected
  exact source even if it is outside the currently loaded list page.
- Do not add a second EventSource transport, polling monitor, fabricated progress
  percentage or automatic independent-session completion subscription.

## 6. Proposed implementation inventory / C ownership

Only this document was changed at CP0. CP1a/CP1b changes are recorded separately in the
UI progress document. After the applicable release C owns the frontend files below;
A/B should publish Go/JSON contracts, not edit these TypeScript files concurrently.
Final subset depends on the approved slice.

| Files | Work |
| --- | --- |
| `clients/desktop/src/engine.ts` | Shared summary/event/request types, proposed typed errors and API helpers |
| `features/agents/types.ts` (new), `demo.ts` | Production agent/source/override types and slug helper separated from fixtures |
| `features/agents/api.ts`, `useAgentPage.ts`, `AgentProblemMessage.tsx` (new in CP1b) | Delivered management routes, bounded scoped page state, structured failures |
| `features/agents/AgentPicker.tsx`, `AgentOverridesEditor.tsx`, `SessionAgentDialog.tsx`, `binding.ts`, `AgentMessageComposer.tsx` (new in CP2a) | Real creation selection, field-specific binding settings, shared scoped detail/eligibility and composer integration |
| `features/agents/catalog.tsx` | Explicit global/project authoring context, real model discovery in both scopes |
| `features/agents/AgentsPage.tsx`, `AgentEditorDialog.tsx`, `AgentInUseDialog.tsx`, `DeleteAgentDialog.tsx`, `agents.css` | Scope-aware list/editor/mutations, collision/reference recovery |
| `features/agents/AgentPicker.tsx` (new) | Shared real source picker, paged search and origin metadata |
| `features/agents/AgentOverridesEditor.tsx` (new) | Reusable field-level explicit/reset controls |
| `features/agents/SessionAgentDialog.tsx` (new) | Session binding, effective settings and recovery |
| `features/graphs/files.ts`, `catalog.ts`, `ProjectAuthoring.tsx` | New production type imports and global+local catalog integration |
| `features/graphs/AgentNodePanel.tsx`, `JoinNodePanel.tsx`, `NodeConfigurationPanel.tsx`, `GraphRunActivity.tsx` | Scope labels and actual graph-child activity navigation, only as needed |
| `features/settings/SettingsDialog.tsx` | Global Agents management entry |
| `features/workspace/CreateSessionDialog.tsx`, `App.tsx` | Agent-aware create payload, binding eligibility/controls and composition |
| `features/library/api.ts`, `SessionLibraryDialog.tsx`, `LibraryPanel.tsx`, `session-library.css` | Provenance/reset, global references and recovery response rendering |
| `features/chat/ModelPicker.tsx`, `modelCatalog.ts`, `SessionOptions.tsx`, `SessionStatusBar.tsx` | Explicit override writes, fixed-harness recovery and specialist identity |
| `features/chat/ConversationEvents.tsx`, `SubagentView.tsx`, `SessionEvent.tsx`, `styles.css` | Typed batch/child activity and failure/reuse/history display; stylesheet is at src root |
| `features/chat/DelegationActivity.tsx`, `DelegationRequests.tsx` (new) | Shared compact batch renderer and bounded requests if B uses a separate projection |
| `features/chat/PendingRequestPages.tsx`, `GraphRequestPages.tsx`, `PermissionCard.tsx`, `QuestionCard.tsx`, `SessionNavigation.tsx` | Exact origin/reply targeting and parent/graph-child navigation |
| `features/chat/SideChatPanel.tsx`, `features/workspace/FocusSessionChat.tsx` | Consistent side/Focus/mobile sheet composition |
| `features/chat/sessionState.ts`, `scopedConnection.ts`, existing shared chat CSS | Summary merging, approved scope invalidation and responsive state presentation if needed |

All relative feature paths above are under `clients/desktop/src`. Design-system
primitives remain generic: reuse Button, IconButton, Input, SearchSelect, RadioGroup,
Slider, Toggle, Menu, ListTile, BottomSheet and existing tokens. No Flutter change
is currently necessary; it hosts this same web frontend. Design-system demo callers
may need type-only compatibility, but production must never fall back to fixtures.

## 7. Remaining integration after CP2a

1. Resolved by manager: Settings -> Agents as global management entry and project
   scope selector approved and delivered in CP1b.
2. Resolved for management: A delivered reviewed catalog/model/exact-read/CRUD APIs,
   typed errors, global references and sourceUid identity. C CP1b wires these only.
3. Resolved for CP2a: A delivered actual creation/binding summary/detail/write shapes,
   fixed-harness eligibility and model/Library override semantics; C wires them.
4. Resolved for UI: existing global/project/conversation invalidations and new
   Library reference kinds are wired. Real source lifecycle/deletion/snapshot
   behavior remains subject to the integrated backend and CP4 runtime evidence.
5. B: finalize batch/child/invocation states, task outcome availability, parent action
   flags, typed timeline payload and compact/paged read routes across all parents.
6. B: finalize request projection/reply targets and graph activation child navigation;
   preserve legacy native-child reads without guessed new origin metadata.
7. Manager: CP1b is approved. CP2a creation/binding UI awaits review; child runtime
   UI/global graph selectors remain held until a specific release after their
   actual A/B interfaces are reviewed.

## 8. CP0 evidence and later validation boundary

Read `AGENTS.md`, `product.md` (product definition, persistence, shared UI, Library,
permissions and child behavior), the implementation plan and the existing frontend
seams named above. Initial `git status --short` (10-second timeout, exit 0) showed
preexisting changed/untracked documentation/probes; these are preserved.
`git diff --no-index --check -- NUL "AGENTS_UNIFICATION_UI_CONTRACT.md"`
(10-second timeout, exit 0) found no whitespace errors; Git emitted its normal
LF-to-CRLF working-copy notice.

CP0 was document inspection/design only. No TypeScript compilation, test suite,
runtime, provider call, installation or browser validation was run or is claimed.
CP1a lightweight check uses the installed TypeScript compiler directly with
`--noEmit --pretty false --project tsconfig.json` from `clients/desktop` and a finite
timeout; outcome is recorded in the progress document. The separate final
runtime/Playwright session owns actual desktop/narrow
UI execution evidence after manager review; compilation cannot establish behavior.

Checkpoint exit: send this contract and decisions to manager once, then end the
turn. Resume only on explicit release/correction instruction; no polling or waiting
process is needed.
