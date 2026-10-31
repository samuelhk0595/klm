# A — Agent catalog and configuration contract

Status: **CP1a/CP1b/CP2a approved; CP2b graph catalog/snapshot delivered for review**. Agent bindings,
manual/spawn creation, per-turn capture and Library cleanup are implemented.
Global graph snapshot v2/frozen delegation catalog is implemented in CP2b; actual
graph-child execution/finality stays with B under its separate manager assignment.
Current graph helper/format/evidence handoff is
`AGENTS_UNIFICATION_GRAPH_AGENT_SNAPSHOT_HANDOFF.md`. Storage handoff is documented in
`AGENTS_UNIFICATION_DELEGATION_STORAGE_HANDOFF.md`, with explicit schema 3 and
unwired recovery/runtime activation boundaries.

Owner/reply session: `e6684b09eac83ac3df478c50bf743022`.
Manager: `a05369eaa5c9151b61e93fb38cb35c52`.
Authority: `AGENTS_UNIFICATION_IMPLEMENTATION_PLAN.md`, not this document.

### CP0 coordination revision

Read `AGENTS_UNIFICATION_UI_CONTRACT.md` after the manager's coordination input.
The manager approved Settings -> Agents without projects, the combined project
list with origin, creation scope selection and read-only existing scope. These UX
decisions are incorporated; C retains exclusive TypeScript ownership and only its
independent CP1a preparation is released. A remains at document-only CP0.

The manager approved storage/identity/offline policies and the full eligible graph
delegation capture direction. A is the sole feature schema/store/domain/migration
integrator, including B's later supplied domain specifications. C owns frontend.
CP1a covered catalog/registry/lifecycle/outbox, pure resolver/types/persistence,
real catalog/CRUD/model APIs and fresh/v1 upgrade; it exposed no partial binding.
CP1b now supplies the working resolver/preparation/creation/settings paths below;
actual files, final DTOs and focused evidence are in the progress document.

This revision supersedes the initial A proposal in three material places: skills
and MCPs have independent override presence, override PATCH uses presence-aware
nullable values as directed by the manager, and only confirmed disabled/deleted
sources enable automatic fallback.
Unknown I/O/conflicts retain the capture but block a new complete turn. Transport
names below are A's merge proposal, pending manager review, not implemented APIs.

## 1. Inspected baseline before CP1a

- `engine/authoring.go:27` already defines the required specialist configuration
  (`AgentDefinition`). `authoringSlug` and `validAuthoringID` are the canonical
  normalization/validation functions. Reuse them, including Unicode normalization
  and Windows reserved-name checks.
- `loadAuthoring(Project)` reads only `.klm/agents/*.toml` and that project's graphs.
  It defaults `enabled` to true, disallows unknown TOML fields and checks the
  name/filename slug. Malformed files become catalog errors. It does **not** fully
  validate externally edited model, prompt or Library configuration.
- Current uniqueness, references, rename, disable and deletion checks are local
  to one project. Rename updates local graph YAML via `commitAuthoring`; referenced
  agents cannot currently be disabled/deleted. There is no persistent source identity.
- `configPath` rejects linked configuration files/directories, restricts generated
  relative paths and reports an unavailable project directory. Keep these boundaries.
  `readConfig` bounds individual files to 2 MiB; prompts are bounded to 128 KiB on save.
- SQLite is runtime authority. `sqliteDomains` explicitly inventories persisted
  fields, including split session metadata, selections and applied/preparing captures.
  Adding only a Go field or only fresh-database DDL does not upgrade existing stores.
  `sqlite_store.go` currently requires schema version 1 exactly; an additive,
  versioned upgrade path is necessary and must be coordinated with B.
- `api.go:createSession` accepts project/topic/title/harness/model; it does not yet
  accept effort or a saved agent. `newTopLevelSession` is also used by independent
  spawn. The general singleton and side chats have separate creation paths.
- `spawnArgs`, `sameSpawnRequest`, `spawnSelectionLocked` and `sqlSpawnSelection`
  retain explicit-human provenance, settings inheritance and sender-scoped receipts.
  `spawnSQLSession` validates outside the application lock and atomically rechecks
  settings/topic before creating the session, queue item, receipt and origin event.
  Accepted receipts are checked before a new creation attempt. Preserve this ordering.
- `inheritedCapabilities` / `sqliteSessionTx.inheritedSelection` use applied
  **selection**, falling back to saved selection. Do not copy `CapabilitySnapshot.Skills`:
  `prepareCapabilities` appends internal role skills to that snapshot after capture.
- `harness.go:execute` rereads turn settings, adds `conversationGuidance` and topic/
  graph context, then prepares capabilities and model defaults. This is the boundary
  seam; adding a specialist only in the initial user message would miss later edits
  and would duplicate instructions on some continuation paths.
- `selectSQLCapabilities` rereads the saved selection independently. It must consume
  the same resolved configuration token as model/prompt preparation for linked agents.
- `changeSessionSettings` / `changeSQLSettings`, the Library PATCH handlers and the
  general agent's settings tool currently write effective settings directly. They
  must record explicit overrides for bound sessions, even when the chosen value is
  equal to the current agent value.
- `captureGraphSnapshot` opens local agent files directly. Source validators assume
  project-relative TOML paths. `graphCatalog`, general graph-list tools, graph Save
  validation and graph-selection consumers also use `loadAuthoring`.
- Library deletion has durable filesystem/state outboxes in `library_api.go`,
  `sqlite_library_deletion.go`, `sqlite_outboxes.go` and maintained reference indexes
  in `sqlite_library_references.go`. It currently knows only local agent files and
  session desired selections; new bindings/fallbacks/globals must join that contract.
- `graph_adapter.go:bridgeTools` is the common role-catalog seam for all harnesses.
  General, normal, graph and subagent roles are deliberately different. A saved
  specialist must never select its administrative tools through configuration.
- `models.go:catalog` currently keys cache/flights by project ID + harness but
  discovery consumes CWD. Global authoring needs real engine-owned discovery with
  no project or session creation. Include CWD in the shared key to avoid mixing a
  general/global catalog or an activation workspace with another context. C's CP1a
  now explicitly separates demo models; an absent project still needs the new real
  global endpoint before global editing is wired.

## 2. Storage and visibility proposal

### 2.1 One authoritative definition per source

**Recommended global storage:** SQLite `global_agents`, scoped to this engine data
directory. It is independent of any selected project and naturally separates dev
and release engines. Do not create a second global TOML copy or put agent settings
in `engine_settings`. Local definitions remain filesystem-authoritative at their
existing paths. SQLite stores their identity/observed state, not an independently
editable copy of local configuration. Captured session/run configurations are
explicit immutable fallback/execution records, not alternate authoring sources.

The effective project catalog contains global definitions plus that project's
locals. A general-agent parent with no destination project sees globals only;
general `session_create` resolves against its explicit destination project.
Graph children resolve against their owning run capture, never live authoring.

Slugs are unique **engine-wide**, including disabled definitions and locals in other
projects. Proposed policy: removed-but-retained project registrations keep slug
reservations so restoring a project does not silently introduce a collision.
They do not appear as selectable agents. No scope move endpoint in this slice.

### 2.2 Registry/reconciliation

Add an `agent_sources` metadata domain with immutable `sourceUid`, scope, project ID,
current slug, relative locator, source state and last observed revision. Keep deleted
source tombstones. Globals store definition + source UID in `global_agents`; the
registry is a locator/lifecycle index, not a second configuration source.

Reconcile global rows and all retained registered project roots under `authoringMu`.
Read project metadata in bounded SQL pages; never load histories. Track completeness
and structured diagnostics. Do not persist duplicate external slugs as a successful
single owner: retain each source and mark every contender conflicted. The slug index
must be able to represent legacy conflicts rather than rejecting engine startup or
silently picking a row. Engine mutations serialize under the same lock and recheck
the registry/project fingerprint in their SQL transaction.

- An absent `.klm/agents` under an accessible project means an empty local catalog.
- An inaccessible project/configuration directory is **unknown**, not empty. Keep
  its known reservations; do not tombstone sources after a failed enumeration.
- To prove a new slug unused, all registered roots must have been checked. If any
  cannot be read, reject create/rename with `catalog_incomplete`, including when no
  agent was previously indexed there. Do not claim engine-wide uniqueness from a cache.
- Existing successfully reconciled sources can be displayed with completeness
  diagnostics. New creation/binding using a saved agent requires a complete uniqueness
  check. Default session creation remains available independently of catalog health.
- Engine cannot atomically control unrelated external filesystem writers. Re-read
  observed revisions/identity before admission; an interleaved external edit yields
  a conflict, never an automatic overwrite. Subsequent collisions become diagnostics.
- Invalid filenames still reserve any valid filename slug to prevent overwriting a
  malformed file. Name/filename disagreement and duplicate source UID are diagnostics.
- Project aliases pointing to the same physical configuration do not establish two
  valid local owners. Report conflicting project/source identities; never shadow.

### 2.3 Source identity, rename and same-slug recreation

`id` remains the name-derived slug. Add a separate opaque **source UID**, which is
not another user-facing agent ID. A session binds to this UID, not just a slug/path.

Global creation mints it in SQL. For local TOML, propose an optional top-level
`source_uid = "<opaque UID>"` metadata field, parsed by an authoring document wrapper
around `AgentDefinition`. Old TOMLs remain readable without it. A new API-created
local definition will include it once the graph decoder handoff permits marker
emission. CP1a assigns local source UID in the registry without introducing markers
into existing graph-readable TOMLs; no live binding is exposed in this slice.
This registry-observed UID is not a claim of durable incarnation tracking for an
unadopted legacy file. Before emitting/adopting a marker in CP1b, graph decoding,
offline snapshot validation/equality, authoring serializers and Library deletion/
outbox serializers must all preserve it. No snapshot reader may silently drop or
ignore the marker to claim compatibility. These integrations remain unreleased.
A legacy local definition selected for a **live**
session/child binding is adopted once through the authoring outbox: prepend the
metadata without rewriting existing content/comments and record its UID atomically
with the recoverable intent. Do not rewrite all files during a GET or startup.
If a legacy file is read-only, listing and graph capture still work; new live binding
reports `source_identity_required` until adoption can be completed. This behavior
needs manager approval before implementation.

The marker is needed because deleting/recreating an unmarked file with identical
bytes between scans is otherwise indistinguishable from an edit. Modification times,
path names, content hashes and OS file identity are insufficient for both atomic
editor replacement and portable rename identity.
If an external author deliberately preserves the same marker during an unobserved
delete/recreate, the engine cannot distinguish it from continued ownership; no
stronger filesystem-history guarantee is claimed. Observed absence/API deletion
tombstones remain authoritative and cannot be undone by copying that marker back.

| Operation | Identity/result |
| --- | --- |
| API rename | Preserve UID; slug/path changes; graph live references updated transactionally. Bound sessions resolve by UID. |
| External rename preserving marker + matching filename/name | Follow UID; do not rewrite graph YAML on a read; old graph slug is diagnosed until repaired. |
| Content edit preserving marker | Same source, new revision; eligible at next complete turn. |
| Disable | Keep source/reservation. Existing bindings retain the latest captured enabled configuration. New binding rejected. |
| Delete confirmed by successful scan/API | Tombstone UID. Existing bindings freeze; same slug can be used by a newly minted source. |
| Marker removed/replaced on a previously adopted path | Never rebind by slug. Ambiguous identity is `source_replaced` and blocks; only a successful complete reconciliation proving the old generation ended may tombstone it as `deleted` and enable fallback. New unmarked source needs fresh adoption. |
| Two files with the same UID | Conflict; neither wins. Retain existing capture for recovery, but block a new complete turn and any new source selection. |
| Unavailable directory / invalid TOML | Retain capture, show `source_unavailable` / `source_invalid`, block a new complete turn; do not activate deleted-source fallback. |
| Restored deleted file with tombstoned UID | Does not resurrect the old live link automatically; explicit replacement/new UID is required. |

Editing a definition while disabled does not update bound-session fallbacks. They
retain their latest captured usable configuration; enabling that same nondeleted UID
allows later boundary updates. Deleting and recreating must mint a fresh UID.

Fallback eligibility is explicit evidence, not a failed lookup. A confirmed disabled
row/file or a deleted/tombstoned generation permits existing bindings to use their
last capture. `source_unavailable`, `source_invalid`, `source_replaced` without a
confirmed tombstone, or conflicting live identities retain that capture for recovery
and display, but do not automatically execute it. Their `execution.allowed` is false
at a new complete-turn boundary. Active work and frozen run/invocation captures
continue under their admitted snapshot. No new approval dialog is introduced;
restoring access/valid identity and explicit retry recovers through ordinary controls.

Graph YAML is intentionally a slug reference for **future runs**. A same-slug new
source may therefore serve a newly started run; already captured runs never rebind.
Keep the existing API guard against disable/delete of a live graph-referenced agent,
expanded across projects for globals. External deletion can still invalidate future
runs; captured artifacts remain readable. Existing session references do not prevent
disable/delete. Global rename must inspect/rewrite all affected live graph files;
if references cannot be inspected or written, reject before mutation.

Multi-root rename/adoption operations need one SQL authoring intent with expected
revisions and per-root relative write sets, then file roll-forward, then registry
commit/publication. Reuse the existing journal/outbox patterns, including expected-
versus-already-written hash checks. No absolute arbitrary client paths, silent
conflict overwrite, historical event rewrite, or history-wide scan.

## 3. Shared types and a single resolver

Names below are proposed exact integration contracts. A owns their implementation;
B calls them and must not create a parallel specialist resolver.

```go
type AgentSourceRef struct {
    SourceUID string `json:"sourceUid"`
    ID        string `json:"id"`       // current/captured slug, not lookup identity
    Scope     string `json:"scope"`    // "global" | "project"
    ProjectID string `json:"projectId,omitempty"`
    ProjectName string `json:"projectName,omitempty"` // display projection, not identity
}

type AgentConfigOverrides struct {
    Model          *string         `json:"model,omitempty"`
    Effort         *string         `json:"effort,omitempty"`
    Prompt         *string         `json:"prompt,omitempty"` // specialist component only
    Skills         *[]string       `json:"skills,omitempty"`
    MCPConnections *[]MCPSelection `json:"mcpConnections,omitempty"`
}

type ResolvedAgentConfig struct {
    Harness   string              `json:"harness"`
    Model     string              `json:"model"`
    Effort    string              `json:"effort"`
    Prompt    string              `json:"prompt"`
    Selection CapabilitySelection `json:"selection"` // external positive lists only
}

type AgentCapture struct {
    Source   AgentSourceRef       `json:"source"`
    Revision string               `json:"revision"`
    Config   ResolvedAgentConfig  `json:"config"` // base before session overrides
}

type SessionAgentBinding struct {
    Source                AgentSourceRef       `json:"source"`
    Revision              string               `json:"revision"` // decimal CAS token
    LastCaptured          AgentCapture         `json:"lastCaptured"`
    Overrides             AgentConfigOverrides `json:"overrides"`
    InheritedCapabilities CapabilitySelection  `json:"inheritedCapabilities"`
    // Bound harness is Session.Harness, fixed when this identity is created.
}

type AgentResolution struct {
    Config          ResolvedAgentConfig `json:"config"`
    Capture         AgentCapture        `json:"capture"`
    SourceStatus    string              `json:"sourceStatus"`
    Blocked         *AgentConfigProblem `json:"blocked,omitempty"`
    Token           string              `json:"token"` // input/source fingerprint
}
```

`AgentConfigProblem` has `code`, `field`, English `message` and `hint`, plus optional
`validValues`. Source statuses: `linked`, `disabled`, `deleted`, `source_replaced`,
`source_unavailable`, `source_invalid`, `conflict`, `captured`. Status and blocked
are separate: disabled/deleted with valid fallback is allowed, incompatible pair
is blocked. Unknown/invalid/conflicting sources also block independently of model
compatibility. The approved UI contract includes the specialist prompt override
in creation/settings; it never exposes or overrides engine role instructions.

Proposed entry points in new `agent_resolution.go`:

```go
// Catalog loading is separate from effective-value composition and validation.
func (a *app) agentCatalog(ctx context.Context, projectID string) (AgentCatalog, error)
func (a *app) resolveNewAgent(ctx context.Context, req NewAgentRequest) (AgentResolution, error)
func (a *app) resolveBoundAgent(ctx context.Context, req BoundAgentRequest) (AgentResolution, error)
func resolveAgentConfig(base ResolvedAgentConfig, fixedHarness string,
    overrides AgentConfigOverrides, inherited CapabilitySelection) (ResolvedAgentConfig, error)
func (a *app) validateResolvedAgent(ctx context.Context, config ResolvedAgentConfig,
    projectID, executionCWD string) error
```

`NewAgentRequest`: project ID, exact name-or-slug selector, optional expected source
UID/revision from UI, optional creation harness override, configuration overrides,
captured inherited external selection, execution CWD and optional frozen graph
catalog. `BoundAgentRequest`: persisted binding, fixed session harness, execution
scope/CWD and optional frozen graph catalog. These are value inputs, not authority
objects. Role, parent ownership, authorization and workspace admission stay with
the creating caller. General has empty project ID; graph passes run-owned catalog.

`resolveNewAgent` requires exactly one enabled, valid, visible, unconflicted source.
An exact saved name is normalized through `authoringSlug`; do not use substring or
fuzzy matching. A selector matching a different expected UID returns a conflict.
`resolveBoundAgent` follows only the bound UID or frozen run record. The pure
composer applies agent base, inherited contribution, then explicit overrides.

Returned config is a value capture with no administrative role or private tokens.
Model validation uses the selected/fixed harness and real execution directory via
`a.catalog`/existing catalog validators; no provider/model reformatting or replacement.
Catalog discovery failure blocks preparation rather than silently accepting defaults.

### 3.1 Precedence and explicit reset

1. Current enabled source (or `LastCaptured` only on confirmed disable/deletion).
2. Captured inherited external contribution, unioned with the source's selection.
3. Explicit per-session overrides. Skills override replaces only composed skill IDs;
   MCP override replaces only composed connections/tool names. These two presences
   are independent, allowing users to remove one contribution without freezing the
   other. Model, effort and specialist prompt overrides are independently tracked.
4. Validate model + effort against the **fixed** harness. Changing agent harness
   does not change the existing session harness. A compatible resulting pair can
   continue; an incompatible one blocks until explicit user recovery.

Persistent override nil means follow the agent. A pointer to `""` is an explicit
empty/default value, **not** follow; a pointer to an empty skills/MCP slice means
select none of that family. For effort `""`, preserve current product meaning: use harness/model
default, not a newly invented `none` effort. Prompt override `""` removes only the
specialist component. An explicit value equal to the current base is still an override.

Manager's final transport decision is `{revision, overrides:{...}}`, with fields
`model`, `effort`, `prompt`, `skills`, `mcpConnections`. A present nonnull value sets
that explicit override; null clears it; omission leaves it unchanged. Reject unknown
fields and empty mutation. `agentOverrideValue[T]` implements presence-aware JSON
decode, distinguishing null from omission; a plain Go pointer is insufficient.
This supersedes the interim set/reset proposal sent during CP0 coordination.

| PATCH JSON | Meaning |
| --- | --- |
| field absent from `overrides` | Keep that override unchanged. |
| `"overrides":{"model":"provider/model"}` | Set explicit override. |
| `"overrides":{"model":"","effort":""}` | Set explicit harness/model defaults. |
| `"overrides":{"model":null}` | Clear only model override; follow agent again. |
| `"overrides":{"skills":[]}` | Select no skills; leave MCP override presence unchanged. |
| `"overrides":{"skills":null}` | Follow source + inherited skill contribution; leave MCPs unchanged. |
| `"overrides":{"mcpConnections":[]}` | Select no MCPs; leave skills override presence unchanged. |
| `"overrides":{"model":null,"effort":null,"prompt":null,"skills":null,"mcpConnections":null}` | Reset all configurable overrides against live or confirmed retained base; fixed harness does not change. |

At creation, omitted/null override fields follow the source. Empty
skills/MCP lists are valid explicit overrides and must serialize as `[]`, never null.
Creation harness override is fixed into `Session.Harness`; it is not a mutable
override that later agent edits can erase. No automatic effort reset when only model
changes: validate the resulting pair and return an actionable incompatibility.

### 3.2 Complete-turn capture and adapter handoff

Add a persisted session configuration child record `agent_binding` plus lightweight
`agentConfig` summary. Store large prompt/fallback payloads outside session metadata.
Optional `agent_turn_preparing` / `agent_turn_applied` child records distinguish
configuration admission from actual adapter application. Publish applied config with
`markCapabilitiesApplied`; never label `LastCaptured` or a merely valid preview as
the configuration that the harness already received. Coordinate these domains with B.
`Session` may carry a hydrated binding for internal use, but metadata writers must
strip it, and default bootstrap/list queries must not hydrate it. B's child session
creation must use the same helper to write this optional record.

Proposed adapter-facing helper:

```go
type AgentExecutionCapture struct {
    Config               ResolvedAgentConfig `json:"config"`
    ExternalCapabilities *CapabilitySnapshot `json:"externalCapabilities,omitempty"`
    Source               *AgentCapture       `json:"source,omitempty"`
    // Runtime owner supplies this for a frozen graph/automation/invocation.
    // Its authorization/activation ownership is validated by the caller.
}

type AgentTurnPreparation struct {
    Resolution           AgentResolution
    BindingRevision      string
    ExternalCapabilities *CapabilitySnapshot
}

func (a *app) prepareAgentTurn(ctx context.Context, sessionID string,
    frozen *AgentExecutionCapture) (*AgentTurnPreparation, error)
```

`AgentTurnPreparation` contains resolved config, input CAS token, source status and
the external capability snapshot after normal capture. A helper commits the new
base fallback/effective projection only after checking the same binding revision,
source revision, session ownership/Stop version and Library recovery state. Failed
validation does not replace the previous usable fallback. Binding, effective model/
effort and desired Library projection publish atomically. A turn uses that one
resolved configuration through prompting, model validation and Library preparation.

B inserts this after `execute` obtains the current session and before native process
preparation/model selection. The helper must run filesystem/catalog work without
holding `app.mu` or an open SQL writer transaction. Lock order remains authoring,
then Library, then `app.mu`, then SQL writer; do not call a locking wrapper from
inside an already-held lower-order lock. Source/binding changes during capture
return `configuration_changed`; no indefinite retry loop.

For bound sessions, `prepareCapabilities` / `selectSQLCapabilities` must take the
preparation's selection/snapshot rather than resolve again. `markCapabilitiesApplied`
still means the adapter has confirmed inventory, not merely that a source was read.
Keep internal skills separate until role-specific injection. Existing Default
sessions retain their normal capability path.

Specialist prompt is composed once per complete engine turn with role guidance and
the task. It never replaces `conversationGuidance`, graph Choice instructions,
Library policy, permission rules or leaf-role constraints. Do not persist the
specialist as a historical user event or splice new prompts into old events.

Steering, native tool/child continuations and B's wait/resume within the same logical
parent turn reuse its captured preparation. Send now resolves latest configuration
only after owned cancellation/drain. Queued messages resolve at actual execution,
not enqueue. An incompatibility retains queued input and exposes recovery instead
of repeatedly consuming/relaunching the queue; A returns the structured block, B
and the scheduler must honor it. This scheduler behavior needs explicit integration.

Graph/active automation captures outrank a live binding. For a reused automation
destination, capture its current bound effective configuration at automation
admission, then freeze it for the run and recovery. Ordinary turns resume live
resolution afterward. Do not add an automation-agent authoring UI in this slice.

## 4. HTTP contract for C

Extend `AgentRecord` with `source: AgentSourceRef`, `status` and optional structured
diagnostics; retain existing definition/ID/revision/graphReferences fields.
Add structured, bounded `references` with graph ID, project ID/name and graph name
(display strings alone are insufficient for global references). `graphReferences`
remains the old string array for compatibility; do not change its element type.
Internal `AgentCatalog` carries full definitions for resolver/graph compilation.
New paged reads return `AgentPage`: `agents: AgentSummary[]`, `diagnostics`,
`complete`, opaque `revision`, `hasMore`, `nextOffset`. A summary has source, slug,
name, description, enabled, harness/model/effort, authoring revision, source status,
`available` and optional `unavailableReason`; it omits prompt/large selections.
`available` certifies source eligibility/uniqueness, not final execution-directory
model verification, which creation/admission must still perform. A diagnostic has
code, message, scope/project/slug and related source identities; do not expose another
project's full prompt in conflict messages. Key rows by source UID, not slug alone.

Paginated endpoints support `search`, `offset` (default 0), `limit` (default 50,
maximum 100); apply visibility and search before pagination, with stable slug +
scope/project/source-identity ordering. The page carries an opaque catalog revision;
a client refreshing after invalidation replaces its contiguous loaded window, not
concatenates incompatible page revisions. Exact selected-source lookup works outside
the loaded page. Reference lists default to 50/max 100 with `referencesHasMore` and
`referencesNextOffset`. Detail `offset`/`limit` pages references
without repeating an unbounded list. Existing full authoring is a compatibility
endpoint: keep complete effective agents and graphs for existing graph editors;
do not silently paginate/truncate its agent array or turn a missing first-page row
into an unavailable graph reference. New management/pickers use the paged routes.

| Endpoint | Contract |
| --- | --- |
| `GET /api/agents?projectId=<optional>&search=&offset=&limit=` | Paged summaries; globals only when project omitted, globals + visible locals when supplied. Catalog/diagnostics, not permission to execute. No separate project-list alias is required. |
| `GET /api/agents/{item}?projectId=<optional>&sourceUid=<optional>` | Exact visible full definition/source/status/references. Optional sourceUid identifies a conflict contender without choosing an executable winner; stale UID rejects. Omitted project means globals. |
| `POST /api/agents` | Create global; `{agent: AgentDefinition}`. |
| `PATCH /api/agents/{item}` | Full global save `{agent, revision, sourceUid}`; revision CAS, slug can change. |
| `DELETE /api/agents/{item}` | `{revision, sourceUid}`; preserve bound fallback, enforce live graph-reference guard. |
| Existing project agent CRUD | Keep URLs/body, add `sourceUid` CAS when adopting/editing an identified source; local writes only. |
| `GET /api/projects/{id}/authoring` | Existing graph list plus effective global/local agents and diagnostics; graphs remain local. |
| `GET /api/models/{harness}?refresh=true` | Real engine-scoped `ModelCatalog`, no project/session prerequisite; same public model fields as existing discovery. |
| Existing `GET /api/projects/{id}/models/{harness}` | Preserve project-specific discovery/validation context. |
| `POST /api/sessions` | Implemented CP1b: existing fields plus optional `agent`, `agentSourceUid`, `agentRevision`, `agentOverrides`, `effort`. Default remains unbound; explicit empty model/effort intent uses nested agentOverrides. |
| `GET /api/sessions/{id}/agent` | Implemented CP1b: `SessionAgentState` below, including current/next config and execution eligibility; decimal binding revision. Full prompt only in detail. Unbound session returns `agent: null`. |
| `PATCH /api/sessions/{id}/agent` | Implemented CP1b: `{revision, overrides:{...independent nullable fields...}}`; null resets, omission preserves. No silent source rebind. Returns updated detail. |

`agent` absent/empty means Default; no binding record is created. A selected agent's
harness/model/effort/Library seed creation, then explicit settings win. Reject
duplicate specification of the same field in top-level settings and `agentOverrides`
instead of introducing competing precedence. `agentOverrides.skills` and
`agentOverrides.mcpConnections` are independently optional lists. Creation harness
override uses the existing top-level `harness`; `agentOverrides` excludes harness.
C can keep it in its editor draft, then map it to that single creation field.
Manual creation does not inherit any unrelated session. New session response uses
the existing session response shape plus compact `agentConfig`; unbound sessions
omit that summary. Exact JSON proposal:

```ts
type AgentOverrideField = 'model' | 'effort' | 'prompt' | 'skills' | 'mcpConnections';
type AgentExecutionEligibility = {
  allowed: boolean; code?: string; message?: string; fields?: AgentOverrideField[];
  token: string; // observed input fingerprint, never future permission to run
};
type AgentConfigSummary = {
  source: AgentSourceRef; name: string; sourceStatus: string;
  capturedRevision: string; pendingRevision?: string;
  overriddenFields: AgentOverrideField[]; pending: boolean;
  execution: AgentExecutionEligibility;
};
type SessionAgentState = {
  agent: null | (AgentConfigSummary & {
    revision: string; overrides: AgentConfigOverrides;
    inheritedCapabilities: CapabilitySelection;
    desired: ResolvedAgentConfig;
    applied?: AgentExecutionCapture & { revision: string };
  });
};
```

`sourceUid` (A) replaces C's candidate `sourceKey`; scope is the flat source-ref
object above, with project name as display metadata. `AgentConfigSummary` has no
prompt/large selection. `desired` retains the exact incompatible requested values;
do not substitute a provider default merely to render the detail. On unknown I/O,
the retained desired preview is labelled unverified and execution is blocked.
Reset on a disabled/deleted source uses its sanitized last capture; reset on unknown
or conflicted source updates override intent but does not clear its eligibility block.

Existing model/settings and capabilities PATCHes remain valid. On a bound session
they set corresponding overrides (including equal-value saves); capability edits
keep the existing selection CAS and update binding revision in the same transaction.
A legacy full capabilities PATCH explicitly pins **both** list families. Selective
edits/resets use the new agent PATCH so changing only skills does not pin MCPs.
Every change affecting effective skills/MCPs through either route increments the
capability selection revision and binding revision together. A model-only change
updates the binding revision; concurrent detail writes still conflict by that token.
The new agent PATCH performs reset-to-agent operations. Settings remain unavailable
for active automation/graph work. Bound harness cannot be changed by the legacy
harness endpoint; preserve the general singleton's existing first-message lock.
Selecting a source on an existing Default session, detaching, or switching saved
agents after creation is not required here; do not silently implement rebinding.

Errors: invalid input 400; unknown/invisible source 404; stale revision, conflict,
incomplete catalog, identity adoption/recovery or incompatible resulting pair 409;
unavailable model catalog 502. Bodies retain English message and add stable
`code`/`field`/`hint`, optional `conflicts`, `currentRevision`, `blockingReferences`
for C. Keep legacy English `error` in the response, rather than requiring every old
route to switch to a new message field. Deletion/disable keeps existing sessions
usable from fallback, subject to ordinary Library/provider availability.

### 4.1 Real global model catalog

Implement a thin A-owned global HTTP handler calling existing `a.catalog` with empty
project ID and a dedicated engine-owned `a.dir/workspaces/agent-catalog` directory.
Prepare that directory outside the state lock; do not register a project, call
`ensureGeneralSession`, create a native conversation, submit a prompt or choose a
selected/arbitrary project. Use the existing read-only inventory executors, installed
harness gates, connected-provider filtering, 45-second discovery deadline and
55-second HTTP write/client budget. Missing harness/catalog is an explicit error;
no fixtures or guessed defaults.

Global agent save validates against this engine context. Session/child creation
and graph execution revalidate against the actual destination CWD/harness, since
project-local provider/model config can differ. Expand cache/flight identity from
`ProjectID + harness` to `(ProjectID, harness, normalized execution CWD)` in one
shared helper; keep OS-appropriate path identity and existing cache expiry/cancel
semantics. Never put credentials into a cache key. This is a shared implementation
seam, not an independent second discovery implementation.
The manager subsequently released the finite `models.go` handoff. CP1a now uses
one shared `modelCatalogCacheKey` for cache and flights: JSON tuple of project
scope, harness and absolute/clean lexical CWD, case-normalized on Windows. Symlinks
are not resolved into a different discovery ancestry. Global authoring uses the
same existing cached discovery path; refresh, timeout, expiry, cancellation and
provider filtering semantics are preserved. No second discovery algorithm exists.

### 4.2 Exact scoped update keys

Reuse storage contract 3 and the existing multiplexed `subscribeScope` connection:

- Global watch is **`global/`**, i.e. `subscribeScope('global', '', ...)` (trailing
  slash intentional), not a new `agents` scope kind. Global management subscribes
  here. Project agent lists/graph catalogs subscribe to `global/` plus
  **`project/<projectId>`** because global sources affect project visibility.
- Binding dialogs/status projections also watch `global/` and their project where
  applicable, then reload only their mounted `/agent` detail/compact state when a
  source may have changed. Direct binding writes and turn-application updates use
  **`conversation/<sessionId>`** through existing scoped notices.
- A global mutation/reconciliation commits `global_agents`/`agent_sources` changes
  and publishes a post-COMMIT global invalidation. A local mutation also sets the
  affected project change. The current bus delivers global notices for any commit,
  so these catalog subscriptions can coalesce/debounce redundant refreshes; do not
  claim collection-filtered notices already exist or add a second EventSource.
- Source reconciliation discovered during explicit refresh/turn preparation also
  advances the registry revision and publishes the appropriate existing scope.
  No filesystem watcher/polling monitor is required. Direct external edits become
  visible on the next requested reconciliation/turn boundary, not magically at write.
- Reconnect/reset invalidates mounted loaded windows and exact selected-source
  detail. Cache keys include engine identity + explicit global/project context;
  empty project is not the old project-hook no-op or a demo context.

Only mounted metadata/configuration windows reload. Do not enumerate all linked
session histories or emit a model turn on a source edit. Source-summary reads expose
current observed status/pending revision without rewriting prior applied captures.
Default sessions need no migration that assigns a specialist or imports native config.

## 5. Spawn and agent discovery contract

Extend independent `session_spawn` and general `session_create` with:

- `agent`: exact saved name or slug in the destination scope.
- `agentOverrides`: `AgentConfigOverrides`; supports explicit empty values.
- `inheritParentCapabilities`: optional boolean, false for saved-agent creation.

Keep current top-level harness/model/effort for compatibility. Nonempty explicit
values override agent defaults; nested overrides carry empty/reset semantics.
Reject a field specified in both locations. For **no-agent** requests preserve
existing scalar inheritance/default semantics and external-selection copy. Thus
the new boolean is meaningful for saved-agent requests; reject it on no-agent
requests rather than silently changing the legacy no-agent policy.

YOLO is not an agent-definition field: retain current parent inheritance and explicit
user-requested override behavior. Saved specialist configuration is never evidence
of human creation/inheritance authorization. Tool text must say the boolean can be
true only when the user explicitly requests external inheritance.

When true, capture **all** parent's applied external selection, or saved selection
before first application, once at admission. Union skill IDs and connection IDs;
union explicit tool names for repeated connections. A selected connection with no
tools still allows its selected resources. Normalize/deduplicate and enforce current
100-skill/32-connection/512-tool limits on the final selection after per-family
overrides; without overrides the complete union must fit. Reject overflow rather than
truncating. Never copy private bridge tokens, internal skills, role tools, native
config or mutable future parent selection. Store this contribution separately in
the binding so future agent edits do not erase it or start following parent changes.

Extend new receipt payloads with captured source UID/revision and effective accepted
configuration (including inherited external selection). Accepted replay returns its
original receipt even if the definition has since changed/been deleted. Rejected
requests create no session/queue/receipt and reserve no operation ID. Compare the
full new request by value; current `left == right` cannot survive nested selection
slices. Do not rewrite old receipt JSON, historical spawn events or request hashes;
missing new fields have legacy semantics. Check existing receipt before consulting
live catalog or preparing capabilities.

The durable receipt retains its full capture; its acceptance/replay TOOL RESPONSE
is compact: `agent:{source,name,revision}` and
`configuration:{harness,model,effort}`, plus the existing accepted/session/message/
topic/YOLO identity fields. It never echoes specialist prompts, skill lists or MCP
catalogs. Normal, general and legacy callsites share this bounded response helper.

New private read-only tools:

- `agents_list({query?,cursor?,limit?})`: bounded visible summaries, source IDs,
  descriptions, harness/model/effort, availability and diagnostics. No prompt bodies
  in a large list; opaque cursor bound to scope/query/revision, max 50.
- `agent_get({agent})`: exactly one visible definition and selection metadata,
  source UID/revision and status; long prompt returned through bounded content.
- General variants accept an explicit optional `projectId`; omission means globals.
  Normal/side cannot enumerate another project's local definitions.

These tools can be shared with B's delegation-enabled parent roles. Graph parents
read only the frozen run catalog. Leaf children receive no agent-discovery/delegation
administration tools. No save/delete/install/selection mutation tool is implied.
Update `session_spawn_options` / `session_create_options` descriptions to direct
the agent to discovery; avoid embedding an unbounded second catalog in model pages.

## 6. B runtime and graph contract

### B call sites

1. New named child: resolve with `resolveNewAgent` using parent's allowed execution
   scope/CWD and saved agent settings; B supplies parent ownership and captures any
   explicitly requested external inheritance contribution.
2. Reused named child: load its `SessionAgentBinding`, pass the child's fixed harness
   to `resolveBoundAgent` at the new complete child invocation boundary. Reuse never
   changes native harness/history. Source deletion uses existing fallback.
3. Unnamed child: B uses current parent capture + explicit child settings, then the
   same pure composer/model validator; no invented saved source or mutable binding.
4. Leaf role and inherited Library selection are independent. A's resolver returns
   no role tools. B owns leaf catalog enforcement, waiting, Stop, durable outcomes,
   native suppression, permissions/questions and graph drainage.
5. Do not refresh configuration on a wait/resume within one invocation. B stores the
   admitted resolved config/token with invocation identity for idempotent replay.

### Graphs

Replace local-only agent lookup in Save/catalog/new-run capture with the effective
catalog. Local graphs can reference visible global agents by slug. Graph node
overrides and `compileGraph` continue resolving fixed captured values; do not run
the live session resolver when replaying/validating an old run.

Propose **version 2** new graph snapshots, with optional version absent treated as
legacy v1. Preserve v1 validation/hash rules and bytes without relabeling them.
V2 adds per-agent source identity and logical source locators:

- local bytes remain `.klm/agents/<slug>.toml`;
- global source bytes are canonical serialized TOML under `global/agents/<slug>.toml`
  (a logical captured origin, never a path to open against a project);
- origin metadata identifies global versus project and source UID, without absolute
  engine-private paths or credentials. Hash the exact captured sources and manifest.

Live revalidation routes each captured locator to its authoritative source/revision;
offline validation uses only captured sources. `validateGraphSnapshot` currently
hardcodes local source paths, so this is an explicit coordinated seam with B.

Graph saved-specialist delegation also needs a frozen catalog. Recommend capturing
the entire visible **eligible** catalog for delegation at actual run start, in a
separate `delegationAgents` map (configuration/source bytes plus external capability
captures). Otherwise a graph child named after a non-node agent would consult live
definitions and violate run-level immutability. Store invalid/unavailable candidates
as diagnostics rather than breaking an unrelated graph whose referenced nodes are
valid; discovering/selecting those candidates reports unavailable. Referenced graph
nodes retain strict preflight validation. Existing node `Agents`/`EffectiveAgents`
stay distinct from this delegation catalog for old validation invariants.

Legacy local files can be captured for a graph without adoption/writes: use a
run-scoped captured source key when no UID exists. Graph children bind to that
run-local record and never become live subscribers. At child execution, use the
actual activation workspace; selected relative MCP contracts still follow existing
execution-directory validation. No new graph support/platform guarantees are implied.

This broader graph delegation capture is a manager/B decision before CP1. If the
approved first slice restricts named graph delegation to already captured node
agents, expose that exact restriction; never silently fall back to live authoring.

## 7. Library deletion and persistence integrity

- Include global definitions, local documents (preserving source marker), explicit
  overrides, inherited contributions and frozen source fallback selections in
  deletion preview/reference indexes. Do not leave a hidden fallback capable of
  reintroducing a confirmed-deleted item on the next turn.
- Confirmed Library deletion removes **future** references from these mutable
  selections/base fallback selection components. It does not rewrite historical
  prompts, graph snapshots, old receipts or active invocation/turn captures.
- For each mutable binding component, remove the deleted ID while preserving its
  presence: an explicit skills/MCP override that becomes empty remains an explicit
  empty slice, not nil/follow. Scrub source defaults, inherited contribution and
  last-captured base selection as well as current desired projections. A later
  per-family Reset resolves against those sanitized components and cannot restore
  the deleted item. B's active frozen invocation remains unchanged; a later reuse
  resolves the sanitized binding rather than replaying its old applied selection.
- The deletion intent captures exact global-row/file/binding revisions, owns its
  complete write set and commits through recovery before publication. A stale
  settings dialog/Library PATCH/reset then conflicts by binding/selection revision;
  it cannot race a stale copied list back into the cleared configuration. Refuse
  agent/selection writes and new capture while the deletion outbox is unresolved.
  Unreadable authoritative roots prevent a misleading successful reference preview,
  using existing recovery/error semantics rather than silently skipping them.
- A disabled/deleted agent fallback preserves its configuration, not an entitlement
  to an item explicitly deleted from Library. Live invalid external skill edits
  still block capture; missing/offline MCP diagnostics follow current semantics.
- Inherited contributions contain only IDs/tool allowlists; credentials remain in
  existing engine-owned MCP credential/capture stores. Do not duplicate upstream
  secrets inside bindings or agent definitions.
- Add reference indexes for the new current binding components; don't hydrate all
  histories/frozen runs to compute deletion impact. Keep captured-reference indexes
  distinct from mutable-selection indexes, and retain credential liveness correctly.
- New schema must support fresh DBs **and** upgrades of existing authoritative DBs,
  initialize optional configuration records lazily/default absent, and validate
  cross-record source/binding ownership. Do not rebuild existing tables/history or
  rerun the legacy JSON importer. Coordinate one schema version/migration with B.

## 8. Proposed ownership and handoffs (not file claims)

| Files | Proposed owner / required handoff |
| --- | --- |
| New `engine/agent_catalog.go`, `agent_catalog_api.go`, `agent_resolution.go`, `agent_bindings.go`, `agent_tools.go`, `sqlite_agents.go` | A after CP1 release. Split implementation helpers here instead of expanding shared files. |
| `engine/authoring.go` | A: source document parsing, scope metadata, local CRUD and effective catalog adapters. Convert positional `AgentRecord{...}` constructors before adding fields. |
| `engine/session_collaboration.go`, `session_spawn.go`, `sqlite_spawn.go` | A: selectors/overrides/receipt comparison and atomic saved-agent creation. B should call helpers rather than edit these. |
| `engine/session_controls.go`, `sqlite_execution_settings.go`, `sqlite_capabilities_api.go`, Library deletion files | A requested: record overrides and clear future references; compatibility paths included. |
| `engine/store.go` | A is the sole feature store/schema integrator by manager assignment. CP1a adds optional binding and catalog domains; B supplies later domain specifications, no concurrent edits. |
| `engine/api.go` | A CP1a owns only its route/helper insertion (`registerAgentRoutes`); creation-by-agent and B lifecycle routing wait for later finite handoff. |
| `engine/models.go` | Finite A CP1a handoff completed: real global catalog reuses catalog(), with CWD-aware cache/flight identity. No synthetic project/session or duplicate model discovery. |
| `engine/harness.go` | B/manager: call A preparation helper at complete-turn boundary, compose specialist once and preserve preparation across parent continuation. A does not edit concurrently. |
| `engine/capabilities.go`, `sqlite_execution_context.go` | Manager arbitrates: consume A's captured effective selection while B wires child capture/isolation. A supplies selection/composition helpers. |
| `engine/graph_snapshot.go`, `graph_contract.go`, `graph_tools.go`, `graph_adapter.go` | Manager arbitrates: A owns proposed scope/capture lookup changes; B owns child snapshot/finality integration. Prefer sequential A then B or manager-applied thin hooks. |
| `engine/linked_bridge.go`, `general_agent_tools.go`, `sqlite_general_execution.go`, Pi tool projection/dispatch | B/manager shared catalog owner; A supplies `agentDiscoveryTools`, dispatch helper and spawn schema extension. Must project consistently to all three harnesses. |
| `engine/sqlite_domains.go`, `sqlite_schema.go`, `sqlite_store.go`, SQLite indexes/validation/upgrades, `sqlite_session_repository.go`, `sqlite_commands.go`, scoped updates | A is the manager-assigned sole feature persistence/migration integrator. CP1a schema 2 contains only A domains; later B contributions require a coordinated subsequent upgrade. |
| `engine/sqlite_execution_queue.go`, legacy `message_queue.go`, automation admission/snapshots | B/manager handoff: block invalid configuration before queue consumption, preserve frozen automation settings. |
| `clients/desktop/src/**` | C exclusively; A supplies JSON contract and backend response evidence. |
| Plan, coordination log, product/README consolidation | Manager. A updates only this stream document until released. |

The table retains later proposed handoffs, which are not released by CP1a. Actual
CP1a edits and current interfaces/check evidence are in
`AGENTS_UNIFICATION_CATALOG_PROGRESS.md`; frontend and runtime/graph/tool/scheduler
seams remain untouched by A.

## 9. Manager decisions requested before CP1

1. Approve globals in engine-owned SQLite, locals in compatible TOML, and immutable
   source UID distinct from public slug. Approve targeted metadata adoption on first
   live local binding, including explicit read-only/identity errors.
2. Approve retained removed-project reservations and conservative incomplete-catalog
   create/rename/binding rejection; choose a different explicit policy if desired.
3. Approve preserving the graph-reference disable/delete guard, global rename's
   cross-project write set, and the versioned global-source snapshot format.
4. Align B's frozen graph delegation catalog scope and `AgentTurnPreparation` /
   wait-continuation use with this shared resolver. Decide full eligible catalog
   versus an explicitly restricted captured-node first slice.
5. Align C on `agent`, `agentOverrides`, independent skill/MCP presence and explicit
   manager-approved nullable presence-aware overrides (superseding interim set/reset), compact `agentConfig`,
   paged/exact catalogs, real global models and exact existing SSE watch keys.
   Source I/O/conflict blocks are distinct from confirmed disabled/deleted fallback.
   No existing-session source switch is assumed.
6. Assign shared schema/store/API/harness/capabilities/tool/graph files and the single
   upgrade owner. Release only A's first implementation slice: catalog/identity +
   resolver + persistence/API types, before dependent runtime/frontend wiring.

## 10. CP0 verification and next evidence

Performed read-only source inspection and `git status --short` (10-second timeout,
exit 0). Existing modified `CLIENT_AUTH_IMPLEMENTATION_PROGRESS.md` and unrelated
untracked documents/probes were preserved. Only this contract was added by A.
Final `git status --short -- "AGENTS_UNIFICATION_CATALOG_CONTRACT.md"` (10-second
timeout, exit 0) reported this document as the expected new untracked file.
No packages installed, servers started, tests/builds run, or product files modified.

Coordination revision: read C's contract/current model hook, engine model discovery,
general workspace helper and existing SQL change bus + multiplexed client scope
helper. Updated only this document; C's existing CP1a edits remain untouched.
Ran `git diff --no-index --check -- NUL "AGENTS_UNIFICATION_CATALOG_CONTRACT.md"`
with a 10-second timeout: exit 0, no whitespace errors, only Git's LF-to-CRLF notice.

After approval, the first lightweight check should be one targeted Go compilation
covering the released backend slice, with temporary output outside the shared tree.
Later directed checks should exercise uniqueness/conflict/unknown roots, adoption
and same-slug replacement, CAS + explicit reset, disabled/deleted fallback,
fixed-harness incompatibility, saved-agent spawn receipt replay and external-only
inheritance. Full suites and the final Playwright session remain manager-controlled.

## CP1b implementation notes superseding CP0 candidate details

- Optional configuration authority is the existing `agent_bindings` domain, with
  desired configuration, last valid source capture, independent overrides, captured
  inherited contribution, applied configuration and eligibility. No additional
  `agent_turn_preparing` / `agent_turn_applied` tables were created: existing external
  `capability_preparing` / `capability_applied` records and one adapter preparation
  provide that handoff. Applied detail omits private external capture payloads.
- `resolveNewAgent`, `resolveBoundAgent`, `prepareAgentTurn`, `AgentResolution`,
  `AgentExecutionCapture` and `AgentTurnPreparation` are now concrete Go interfaces.
  See their definitions and the progress section for exact fields. Frozen managed
  execution requires an explicit capability capture even when selection is empty.
- Local marker compatibility gate is complete for this slice: wrapper parsing at
  local graph capture + offline validation, authoring and Library/outbox writers.
  Local CRUD can now emit markers; live adoption prefixes legacy bytes only after
  successful model/config validation and revision recheck. Missing adopted marker,
  conflict or I/O blocks; confirmed deletion/disable uses the retained valid base.
- Bound queue execution prepares before moving a queued item into history. Failure
  pauses it; Stop during preparation preserves it without inventing execution failure.
  Active adapters and active/recovered automation run snapshots retain their admitted
  configuration and external capture; they do not re-resolve during continuation.
- New independent creation tools use saved defaults/Library, optional exact source
  UID/revision and all-or-none explicit external inheritance. Existing no-agent
  inheritance/receipts/topics remain; replay is looked up before live source reads.
- Global definitions and all mutable binding components participate in Library
  references/removal. Explicit empty override presence survives cleanup; applied
  snapshots and historical receipts remain unchanged. V2's new mutable reference
  projections receive an authority-checked, paged, one-time metadata-version backfill;
  no B domains or schema user_version 3 were invented.
- Global graph execution/snapshot v2 and frozen delegation catalog remain CP2.
  Only the local wrapper parser/offline path changed in graph_snapshot.go.

**CP2b stops here. Await explicit manager review/correction or next release.**
