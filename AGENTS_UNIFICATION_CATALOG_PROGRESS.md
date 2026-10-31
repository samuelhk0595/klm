# A — Catalog CP1a delivery

Publication note: schema numbers below describe the original standalone source.
The stacked delivery reserves predecessor content schema2, catalog schema3 and
delegation schema4; see `AGENTS_UNIFICATION_PUBLICATION.md`. Historical checks
below were not rerun or relabelled as the publication chain.

Status: **CP1b/CP2a approved; CP2b global graph/frozen catalog delivered for review**.
Reply session: `e6684b09eac83ac3df478c50bf743022`.
Manager: `a05369eaa5c9151b61e93fb38cb35c52`.
Contract: `AGENTS_UNIFICATION_CATALOG_CONTRACT.md`.

## Implemented slice

- Global definitions are authoritative SQLite records, separate from engine
  settings. Locals remain compatible TOML files. The registry stores source identity,
  locator/revision and lifecycle, without an independently editable local definition.
- Engine-wide slug claims include disabled and retained removed-project sources.
  Conflicts are diagnosed rather than shadowed. Unknown directories retain their
  reservations and prevent create/rename. Default session creation is untouched.
- Opaque source UID survives API rename; deletion retains a tombstone and same-slug
  recreation gets a new UID. Optional `source_uid` wrapper metadata is parsed.
  Restored tombstoned markers remain conflicted across repeated reconciliation.
  Malformed/unknown/conflicting sources are not relabelled as deleted.
- GET reconciliation can persist **observed registry metadata**, but does not adopt
  a marker or write source TOMLs. No startup marker migration was added. Existing
  legacy authoring-journal recovery behavior on its compatibility surface remains.
- Global/local mutations preflight source revisions and graph references. A durable
  catalog intent owns multi-project file writes; recovery checks expected/already-
  written hashes and rolls forward authoring only. A global rename can rewrite live
  graph YAML references across projects. Referenced definitions cannot be disabled
  or deleted. A failed accepted write freezes further storage mutation for recovery,
  following the existing graph-authoring model.
- Pure configuration composition applies source plus captured external inheritance,
  then independent model/effort/prompt/skills/MCP overrides. The fixed harness input
  wins. Positive selections are copied/unioned/normalized and bounded; no internal
  role tools or credentials are copied.
- Presence-aware override decoding supports omitted = unchanged, null = reset,
  empty string/list = explicit override. Skills/MCP presence is independent.
  This supersedes the interim CP0 set/reset transport proposal.
- Optional binding storage/helper types exist, but **no route, creation option or
  tool exposes a partially integrated binding**. There is no runtime-success stub.

## HTTP contract implemented for C

All routes retain normal authenticated API wrapping and English error text.

| Route | Shape/behavior |
| --- | --- |
| `GET /api/agents?projectId=&search=&offset=&limit=` | Globals when project absent; globals + that project's locals when supplied. Search before paging; default 50/max 100. Returns `agents` summaries, `diagnostics`, `complete`, opaque `revision`, `hasMore`, `nextOffset`. |
| `GET /api/agents/{slug}?projectId=&sourceUid=&offset=&limit=` | Exact full source/configuration. Optional UID disambiguates distinct-source slug conflicts; duplicate UID ambiguity still rejects. References use offset/limit default 50/max 100. |
| `POST /api/agents` | Global create `{agent: AgentDefinition}`; source UID is engine-generated. |
| `PATCH /api/agents/{slug}` | Global full save `{agent,revision,sourceUid}`. UID is required to protect stale same-slug editors. |
| `DELETE /api/agents/{slug}` | `{revision,sourceUid}`, required exact global lifetime. |
| Existing project agent POST/PATCH/DELETE | Same URLs and complete-definition body; SQLite path uses unified uniqueness/lifecycle/outbox. Optional `sourceUid` is checked when supplied; legacy local callers remain compatible. |
| `GET /api/projects/{id}/authoring` | Existing complete graphs/agents arrays retained, now including globals and additive source/status/reference metadata. No first-page truncation. |
| `GET /api/models/{harness}?refresh=true` | Real installed-harness discovery with 45-second context in engine-owned `workspaces/agent-catalog`; no synthetic project, singleton creation, prompt or thread. |

Summary fields: `id`, `name`, `description`, `enabled`, `defaultHarness`, `model`,
`effort`, `revision`, `source`, `status`, `available`, `unavailableReason`,
`referenceCount`. Full rows additionally retain AgentDefinition and legacy string
`graphReferences`, and add structured `references`, `referencesHasMore`,
`referencesNextOffset`. `source` is flat: `sourceUid`, public slug `id`,
`scope: global|project`, optional `projectId`/`projectName`.

`available` is source/uniqueness eligibility, not a claim that a future destination's
model/effort or runtime preparation has been validated. Actual saves validate the
real harness model catalog. Following the finite `models.go` handoff, global/project
model discovery uses the same cached implementation with scope/harness/absolute
clean CWD identity. Windows case normalization is applied, lexical aliases retain
their discovery ancestry, and cache and flights share exactly the same key. The
initial uncached global workaround has been removed; existing refresh/timeout/
expiry/cancellation behavior is preserved.

Errors preserve `error` and add `code`, optional `field`, `hint`, `conflicts`,
`currentRevision`, `blockingReferences`. Typical conflicts are `agent_conflict`,
`source_changed`, `catalog_incomplete`, `agent_in_use`, `references_unavailable`,
`source_identity_reused`, `source_identity_ambiguous`, `recovery_pending`.

Scoped invalidation reuses existing bus/connection keys: `global/`,
`project/<projectId>` and eventually binding updates on `conversation/<sessionId>`.
Registry/global commits already publish through existing post-COMMIT notices;
local changes also mark the project. C must subscribe mounted project catalogs to
global + project and clear engine-scoped caches on switch. External file changes
are observed at requested reconciliation, not by an added filesystem watcher.

## SQLite contract

- Schema **2 is A's CP1a revision**, not the completed A+B feature schema.
- Added domains: `global_agents`, `agent_sources`, `agent_catalog_changes` and optional
  session child `agent_bindings`. Global slug has a SQL uniqueness index; registry
  projections support source UID/slug/project/lifecycle without choosing conflicts.
- The table factory emits the same selected-domain DDL for fresh and upgrade paths.
  `openSQLiteStore` accepts supported v1 for inspection, but does not mutate it.
  `enableRuntime`, after existing authority/manifest validation, upgrades transactionally
  then recovers accepted catalog authoring intents before normal runtime starts.
- Upgrade creates only the additive domains and advances `user_version`; it does
  not decode/rebuild old sessions, histories, receipts, snapshots or native records.
  Optional binding rows remain absent. Session metadata excludes hydrated binding.
- A is the sole schema/store/domain/index/migration integrator. B must supply its
  actual domain specification/patch; a subsequent version will integrate it. No B
  domain or final v2 feature schema was invented.

Catalog bounds: 2,000 sources, 2,000 graph-reference records, 16 MiB aggregate
configuration scan content, 20,000 retained registry records, diagnostics capped at
200; new list/reference pages max 100. Exceeding a bound is an explicit incomplete/
capacity error, never silently accepted uniqueness. An outbox is capped at 8 MiB
and each project journal write set at the existing 2 MiB limit.

## Files changed by A

New:

- `engine/agent_types.go`
- `engine/agent_catalog.go`
- `engine/agent_catalog_api.go`
- `engine/agent_outbox.go`
- `engine/sqlite_agents.go`
- `engine/sqlite_agent_upgrade.go`
- `engine/agent_catalog_storage_probe_test.go`
- `AGENTS_UNIFICATION_CATALOG_PROGRESS.md`

Modified:

- `engine/api.go` — one route-registration helper insertion only.
- `engine/authoring.go` — optional document parsing and unified SQL CRUD/catalog adapters.
- `engine/store.go` — explicit catalog roots and optional binding field.
- `engine/sqlite_domains.go` — domains and selected additive DDL factory.
- `engine/sqlite_indexes.go` — new-domain projections/global slug uniqueness.
- `engine/sqlite_schema.go`, `engine/sqlite_store.go` — versioned fresh/v1 path and recovery gate.
- `engine/sqlite_session_repository.go` — keep hydrated binding outside metadata.
- `AGENTS_UNIFICATION_CATALOG_CONTRACT.md` — manager decisions and actual boundaries.

Frontend C work and unrelated documents/probes were preserved. A did not edit
`harness.go`, `capabilities.go`, `sqlite_execution_context.go`, graph runtime/snapshot/
adapter files, linked/general tool catalogs, scheduler, automations or spawn files.

## Verification performed

Commands from `engine` unless noted; every command had a finite timeout.

1. `gofmt -w` on A's explicitly named Go files — 10 seconds, exit 0.
2. `Test-Path -LiteralPath "C:\Users\Samuel\AppData\Local\Temp\opencode"`
   — 10 seconds, `True`, before build output creation.
3. `go build -tags dev -o "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-agents-cp1a.exe" .`
   — 120 seconds, exit 0. Final rerun followed actual source corrections.
4. `go test -tags dev -run '^TestAgentCatalogCP1aStorageProbe$' -count=1 -timeout 60s .`
   — tool timeout 90 seconds, exit 0, final output `ok klm/engine 2.433s`.
   Earlier runs passed at 1.624s and 3.047s; directed reruns followed identity/cache/
   stable-catalog corrections, not broad verification or a full suite.
5. `git diff --check -- "engine/api.go" "engine/authoring.go" "engine/store.go" "engine/sqlite_domains.go" "engine/sqlite_indexes.go" "engine/sqlite_schema.go" "engine/sqlite_store.go" "engine/sqlite_session_repository.go"`
   from project root — 10 seconds, exit 0; only Git LF/CRLF notices.

The isolated probe constructs fresh and synthetic v1 fixtures (removing exactly the
four new A domains from an otherwise unchanged fixture), upgrades/reopens them,
checks schema/SQLite integrity, compares old session/event/receipt payload digests,
verifies absent bindings, creates/renames/recreates a global source, recovers a durable
catalog deletion intent, exercises local marker tombstones across repeated reads,
and checks override omission/null/empty differentiation. It uses temporary data,
no provider/model calls, real user data, server, package installation or agent turn.

## Explicit next handoffs / remaining validation

- No agent-based session creation/spawn, live source adoption, next-turn application,
  settings/reset API or Library deletion integration is released yet. Types/helpers
  must not be mistaken for executed runtime behavior.
- CP1a does not emit new local markers because `graph_snapshot.go` still has a strict
  AgentDefinition decoder. Graph decoder/snapshot v2 handoff must precede marker
  emission/live adoption. Existing markerless local graph files remain compatible.
- Required CP1b compatibility gate before any marker emission/adoption: update the
  graph document decoder, offline captured-source validation/equality, all relevant
  authoring serializers and Library deletion/outbox serializers to preserve the
  wrapper marker. Never silently strip or ignore it in snapshots. Those files
  were not edited in CP1a. Markerless legacy files have only registry-observed
  identity in this slice, not a promised durable external filesystem incarnation;
  no live binding uses that provisional identity. API rename continuity is an
  observed engine transaction, not proof of an unobserved external rename/recreate.
- Global authoring/catalog metadata is implemented; graph Save/catalog/execution
  consumers outside authoring and graph v2/delegation capture still await B merge.
  This delivery does not claim global graph execution works yet.
- Full eligible graph delegation capture is approved but not implemented here.
  Invalid candidates must remain bounded diagnostics without breaking valid nodes.
- The CWD-aware model-cache handoff was completed in the supplement below.
- Private discovery/spawn schemas, Library deletion reference indexes/sanitization,
  scheduler admission and runtime/automation boundaries need later explicit release.
- Real harness/provider model discovery HTTP paths, UI, filesystem ACL/offline failure,
  crash/restart with multi-root filesystem writes and human acceptance are not claimed
  from compilation or the isolated storage fixture. Manager review/correction and the
  separately authorized final runtime/Playwright session remain required.

## Coordination follow-up: B requirements received

After the CP1a report, the manager supplied B's future persistence direction.
Read sections 5, 6, 8 and 10 of `AGENTS_UNIFICATION_DELEGATION_CONTRACT.md`.
This follow-up changes documentation only and does not release runtime/schema work.

- Reserved future domains are `delegation_batches`, `delegation_invocations` and
  `delegation_children`. B is limited to disposable pinned feasibility probes;
  its illustrative shapes are not final schema types or permission to add records.
- No B definitions, fake rows or alternative configuration binding were added to
  CP1a. Histories/native identities stay in their existing domains. B child records
  will reference A's optional session-owned binding; frozen invocation captures are
  execution evidence, not another editable settings authority.
- Concrete version seam: fixtures now genuinely have schema 2 with A's domains.
  B integration must introduce the next explicit upgrade (and cover both original
  v1 -> current and already-created v2 -> next), not amend v2 DDL in place or rely
  solely on fresh-store initialization. A remains the sole persistence integrator.
- Concrete naming seam: implemented optional SQL table is `agent_bindings` (plural),
  with `Session.agentBinding` as the split field. B's singular `agent_binding` in
  prose is conceptual; callers should use A's `agentBinding` / `putAgentBinding`
  helpers instead of constructing a guessed SQL locator. Runtime preparation/
  resolve-new/resolve-bound APIs remain proposed until their later released slice.
- Graph-child validation must explicitly extend the existing owner rules: exact
  graph node parent/run/activation/workspace and leaf role, with no Choice grant,
  cross-parent reuse or duplicate lease. Those existing validators/runtime gates
  were not relaxed by CP1a. Startup recovery order and finality quarantine also
  require B's exact coordinator handoff before implementation.
- Approved runtime constraints are one active batch, 1–8 members and max 4
  concurrent children. No new semantic report/classification tool is included;
  final text and factual execution outcome are consumed by the parent.
- Full visible eligible graph specialist catalog remains the capture direction.
  B's proposed per-run capture bounds are separate from A's management catalog
  bounds; overflow must reject preflight explicitly, never truncate the eligible
  set or fall back to live authoring.
- The pinned Codex dynamic-tools start/resume feasibility blocker does not affect
  the delivered catalog. No runtime support is inferred from the A build/probe.

No additional build/test was run for that documentation-only B follow-up.

## CP1a supplement — finite model-cache handoff

The manager subsequently released `models.go` for the CWD-aware key correction.
Implemented `modelCatalogCacheKey(Session, cwd) (string, error)`: the key is an
unambiguous JSON tuple of project scope, harness and absolute clean lexical CWD;
Windows case folding matches existing project-directory identity. Discovery paths
are not resolved through symlinks, which could change inherited configuration.
Both cached catalog lookup and in-flight discovery use this key. All other catalog
options and lifecycle behavior stay on the existing implementation. Global
authoring now calls that same cache path, so agent-catalog and general-agent
workspaces cannot share cached/in-flight results merely because ProjectID is empty.

Additional files changed:

- `engine/models.go` — shared key helper and its use before cache/flight admission.
- `engine/agent_catalog_api.go` — remove the now-unnecessary uncached workaround.
- `engine/agent_catalog_storage_probe_test.go` — isolated cache/flight CWD separation
  probe, including an equivalent lexical path; no native discovery is launched.
- `engine/general_agent_tools_test.go`, `engine/sqlite_automation_api_test.go`,
  `engine/sqlite_automations_execution_test.go`,
  `engine/latency_investigation_bench_test.go` — mechanically update existing
  manually seeded cache fixtures to the new complete identity.

Checks for this actual code handoff:

- `gofmt -w` on the seven named Go files: timeout 10 seconds, exit 0.
- Temporary output parent `Test-Path` repeated: timeout 10 seconds, `True`.
- `go build -tags dev -o "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-agents-cp1a.exe" .`:
  timeout 120 seconds, exit 0.
- `go test -tags dev -run '^(TestAgentCatalogCP1aStorageProbe|TestAgentCatalogModelCacheScope|TestGeneralToolScopeAndAtomicSettings|TestSQLiteAutomationManagementHTTP|TestSQLiteSpawnFrozenSettingsAndReceipt)$' -count=1 -timeout 90s .`:
  tool timeout 120 seconds. First run exposed a preexisting nil-app fixture in
  `TestGeneralToolScopeAndAtomicSettings` before it reached its model-cache check:
  `bridgeTools` requires the general adapter's app. Initialized only that test
  fixture with `app: &app{}`; no production role code changed. Directed rerun passed:
  `ok klm/engine 3.321s`, exit 0. No benchmark or full suite was run.

Nullable PATCH semantics are already aligned with the final manager decision:
omitted unchanged, null reset/follow, empty explicit override, independently for
skills and MCPs. There is no intermediate set/reset implementation to remove.
No binding/runtime/graph/schema expansion accompanied this cache handoff.

## CP1a manager review — R1/R2 corrections

The manager rejected CP1a pending two concrete regression fixes. Original failure
evidence is attributed to the manager's targeted reproductions:

- R1 `TestAgentCatalogManagerRejectsDuplicateLifetimeMutation`: mutation returned
  nil despite two live definitions sharing a source UID (approximately 1.1 seconds).
- R2 `TestAgentCatalogManagerValidatesEffectiveSelectionAfterOverrides`: a
  100-source + 100-inherited skill union failed its intermediate limit before an
  explicit empty skills override could replace it (approximately 0.14 seconds).

Corrections are confined to the approved findings:

- `engine/agent_catalog_api.go`: before any update/delete/rename, admission checks
  all effective source locators and retained nondeleted registry reservations for
  the exact UID. Multiple owners return HTTP 409 `source_identity_conflict`, with
  bounded structured conflict references. This is independent of the capped display
  diagnostics. An unreadable known locator cannot silently resolve a copied UID.
  Slug-only collision with distinct exact UIDs is not blanket-blocked by this gate;
  existing safe rename/preflight rules still apply. No rejected operation writes an
  authoring source or reserves an outbox intent.
- `engine/agent_types.go`: positive-list union is composition only; independent
  family overrides replace it, then the full effective selection is normalized and
  validated once. Oversized source+inherited unions replaced by explicit empty
  skills/MCP overrides no longer fail prematurely. Unoverridden families and unions
  with no override remain subject to all existing limits; no entries are truncated.
- `engine/agent_catalog_manager_probe_test.go`: manager-owned reproductions were
  released for integration. Applied gofmt and preserved every original assertion;
  added rename/delete + unchanged-file checks for R1 and both-family limit controls
  for R2, including errors in an unoverridden family.
- Contract/progress now state final-effective validation and that the finite
  `models.go` handoff is already complete, rather than describing it as unreleased.
  No new model-discovery change was required in this correction round.

Verification:

```text
gofmt -w "agent_types.go" "agent_catalog_api.go" "agent_catalog_manager_probe_test.go"
  workdir: engine; timeout 10 seconds; exit 0

go test -tags dev -run '^(TestAgentCatalogManagerRejectsDuplicateLifetimeMutation|TestAgentCatalogManagerValidatesEffectiveSelectionAfterOverrides|TestAgentCatalogCP1aStorageProbe|TestAgentCatalogModelCacheScope)$' -count=1 -timeout 60s .
  workdir: engine; tool timeout 90 seconds; exit 0
  ok klm/engine 2.719s
```

This single directed execution covered both review regressions, the storage fresh/
v1 upgrade probe and model-cache/flight CWD isolation. It compiled the affected Go
product code; no separate redundant build or full suite was run. No runtime, graph,
Library, schema, frontend or B probe files were changed in this correction round.
CP1a is not self-approved; manager revalidation is still required.

## CP1b — working saved-agent session binding

The manager approved CP1a after independently rerunning the four directed checks
(exit 0, 2.682 seconds, manager evidence), then explicitly released CP1b. The work
below is a new slice, not a repetition of that approved CP1a verification.

### Implemented behavior

- Concrete `resolveNewAgent`, `resolveBoundAgent`, `prepareAgentTurn` APIs now share
  the same pure effective-selection composer and real model validator. Fixed
  harness/model compatibility is enforced without provider/model translation or
  native-history replacement. Explicit per-field overrides survive source edits.
- Source UID adoption happens only on a live creation operation, after successful
  configuration/model validation and a revision recheck. It prefixes `source_uid`
  while retaining original legacy bytes/comments/BOM. Local CRUD now also emits the
  wrapper marker. A removed adopted marker is a source error, not implicit rebinding.
- Local graph capture parsing and offline snapshot validation explicitly decode the
  wrapper. Old source text/hash/record shapes are retained; no graph v2 field or live
  global graph execution was added. Authoring and both Library writer paths preserve
  the marker. A marked v1-format fixture validates without changing captured bytes.
- Binding records contain desired configuration, last valid base capture, independent
  overrides, captured inherited contribution, source status/block and applied config.
  Large settings remain outside session metadata; `agentConfig` is the compact
  metadata projection. No new preparing/applied configuration domains were invented.
- A complete turn resolves model/prompt/positive selection together, captures external
  Library once, CAS-checks the binding/source and commits its preparing snapshot.
  `execute` and `prepareCapabilities` consume that same preparation. Specialist prompt
  is added once alongside existing role/topic/Library guidance. Saved configuration
  is explicitly described to the model as not human authorization.
- Actual adapter inventory confirmation records applied config through
  `markCapabilitiesApplied`. Desired changes made during active work do not alter its
  capture. Frozen managed execution requires its explicit capability snapshot even
  for an empty selection; nil cannot silently become legacy additive configuration.
- Source invalid/conflict/missing marker/I/O blocks new preparation. Confirmed
  disabled/deleted/replaced origin retains the last valid capture with normal model/
  Library validation. A same-slug new UID never relinks the old binding; source moves
  across scope cannot silently import another project's definition.
- Bound queued execution reserves the owner turn, prepares outside app/SQL locks,
  then atomically moves exact queued inputs to history. Incompatibility pauses them
  instead of consuming/relaunching them. Stop during preparation preserves input,
  clears preparing state and returns idle rather than inventing execution failure.
- Manual creation chooses Default or a saved agent. Saved creation does not send a
  message. Independent spawn and general-create support an exact saved name/slug,
  optional source UID/revision, explicit overrides and creation-only all-or-none
  external inheritance. Without an agent, existing inheritance remains unchanged.
- With an agent, its Library is the default. Explicit true inheritance captures the
  entire parent's applied external selection (saved before application), unions it
  with source lists, then applies independent overrides. No internal tools/skills,
  parent role, live future selection, bridge token or history is copied.
- Independent receipts retain original request/history shape when new fields are
  absent. New receipts capture source/configuration. Replay returns the receipt before
  consulting a changed/deleted source. Human event checks, topic inheritance, general
  explicit project/no-topic rules and independent lifecycle remain in the old gate.
- Existing model/effort settings route records overrides. Existing full Library
  PATCH explicitly pins both families; the agent PATCH can edit/reset either one.
  Binding/selection CAS prevents a stale dialog from restoring old choices. Prompt/
  Library intent may be saved for next turn while work is active; model/effort changes
  wait for the current turn. Active automation settings remain read-only.
- Global definitions, inherited contributions, last-valid fallback, explicit
  overrides and desired selections all join Library reference/removal handling.
  Confirmed removal sanitizes every mutable future component while preserving
  explicit empty override presence. Applied snapshots and old receipts stay intact.
- V2 stores that already contain CP1a global definitions receive a paged one-time
  `agent_binding_reference_version=1` projection backfill after authority validation.
  Only mutable reference indexes are backfilled; no histories/snapshots/receipts or
  schema user_version 3/B domains are created.
- Automation reuse refreshes an idle bound destination outside app/SQL locks before
  normal admission rechecks. The admitted engine-owned specialist capture is stored
  in its run/recovery snapshot; continuation uses that frozen config/Library rather
  than live source edits. User automation Save cannot inject that internal field.

### Definitive HTTP and private-tool contract for C/B

```text
POST /api/sessions
  Existing fields + agent?: string, agentSourceUid?: string,
  agentRevision?: string, agentOverrides?: AgentConfigOverrides, effort?: string.
  Top-level harness is creation-only. Nonempty top-level model/effort may override;
  duplicate nested specification rejects. Explicit empty/default values use nested
  agentOverrides. Existing {session,revision,eventCount,more} response is retained,
  with session.agentConfig compact metadata. Default has no binding.

GET /api/sessions/{id}/agent
  {agent:null} when unbound, otherwise:
  {agent:{source,name,sourceStatus,capturedRevision,pendingRevision?,
    overriddenFields,pending,execution:{allowed,code?,message?,fields?,token},
    revision,overrides,inheritedCapabilities,desired,applied?}}
  Binding revision is a decimal string. Applied detail omits private external
  capture inventory/credential references; source specialist prompt is detail only.

PATCH /api/sessions/{id}/agent
  {revision:"<binding CAS>",overrides:{model?,effort?,prompt?,skills?,mcpConnections?}}
  Omitted unchanged; null follows source; empty string/list explicitly overrides.
  Responds with the same detail shape. No source switch/detach/harness mutation.

PATCH /api/sessions/{id}/settings
  Existing shape/behavior; model/effort become explicit binding overrides when bound.

PATCH /api/sessions/{id}/capabilities
  Existing decimal selection CAS/full selection shape; bound sessions pin BOTH
  skills and MCPs atomically with binding revision. Per-family reset uses /agent.
```

`AgentConfigOverrides` contains independently optional model/effort/prompt strings,
`skills: string[]`, and `mcpConnections: MCPSelection[]`. Resolved `desired` is
`{harness,model,effort,prompt,selection:{skills,mcpConnections}}`. Source status uses
`linked`, `disabled`, `deleted`, `source_invalid`, `source_unavailable`, `conflict`
and exceptional `source_replaced`/`captured`; status is independent of eligibility.
Compact `Session.agentConfig` contains no full prompt or Library snapshot.

`session_spawn` / general `session_create` add `agent`, `agentSourceUid`,
`agentRevision`, `agentOverrides`, `inheritParentCapabilities`. The inheritance
boolean is meaningful only with an agent, defaults false and needs the user's
explicit request. YOLO remains the existing sender/default/explicit setting, not
part of a saved definition. Creation errors retain accepted=false/code/field/hint.
New read-only `agents_list` and `agent_get` use bounded existing page/fragment helpers;
normal/side scopes stay local, general can supply an explicit destination project.
They are not added to graph or leaf catalogs in this slice.

Library impact projection retains `LibraryReference` fields and adds factual kinds
`global_agent` (public slug, no project) and `agent_binding` (session/project identity)
alongside existing current selections and captured references. C needs to render
these new kinds when wiring bound/global deletion impact.

### Concrete Go/runtime handoff

- `NewAgentRequest`: ProjectID, Agent, SourceUID, Revision, Harness, ExecutionCWD,
  Overrides and InheritedCapabilities. `BoundAgentRequest`: persisted Binding,
  FixedHarness, ProjectID and ExecutionCWD.
- `AgentResolution`: Config, base Capture, SourceStatus, optional Blocked and Token.
  `AgentTurnPreparation`: Resolution, BindingRevision, ExternalCapabilities.
- `AgentExecutionCapture`: Config, optional Source, Revision, optional
  ExternalCapabilities. Frozen managed execution must supply ExternalCapabilities
  even when empty; do not reconstruct it from a live selection during continuation.
- `execute(..., preparations ...*AgentTurnPreparation)` accepts an already admitted
  preparation, avoiding a second source lookup. The adapter keeps it through the
  logical turn. B must preserve it during join/resume rather than opening a new
  configuration boundary.
- `agentBinding` / `putAgentBinding` remain the optional SQL authority helpers;
  child ownership records must reference them, not duplicate editable settings.
- `harness.go`, `capabilities.go`, private bridge/general catalogs and queue/runtime
  seams are ready for a **manager-assigned** B handoff after review. B must preserve
  the pre-consumption bound queue gate, source/selection CAS and existing role/tool
  isolation. No automatic ownership transfer is asserted by this report.
- A retains sole schema/store/domain/index/migration ownership. B's new record/helper
  files were preserved and were not edited, imported into schema or activated by A.

### CP1b files

New: `engine/agent_resolution.go`, `agent_bindings.go`, `agent_turn.go`,
`agent_automation.go`, `agent_library.go`, `agent_tools.go`,
`agent_binding_probe_test.go`.

Updated A files: `agent_types.go`, `agent_catalog.go`, `agent_catalog_api.go`,
`sqlite_agents.go`, `sqlite_store.go`, `store.go`, `api.go`, `history.go`,
`session_collaboration.go`, `session_spawn.go`, `sqlite_spawn.go`,
`sqlite_execution_settings.go`, `sqlite_capabilities_api.go`, `sqlite_commands.go`,
`sqlite_app.go`, `linked_bridge.go`, `general_agent_tools.go`,
`sqlite_general_execution.go`, `harness.go`, `capabilities.go`,
`sqlite_execution_queue.go`, `automations.go`, `automations_api.go`,
`sqlite_automations_api.go`, `sqlite_automations_worker.go`, `authoring.go`,
`library_api.go`, `sqlite_library_deletion.go`, `sqlite_library_references.go`,
`sqlite_outboxes.go`, and the local-only parser/offline portion of
`graph_snapshot.go`. Contract/progress were updated. `sqlite_execution_context.go`
needed no edit because the prepared-capability branch bypasses its live reread.

### Focused verification and actual evidence

All commands used finite timeouts; workdir was `engine` unless stated.

- `gofmt -w` on explicitly named changed Go files: 10 seconds, exit 0.
- Initial targeted compilation identified a misplaced snapshot branch, a duplicate
  local err declaration and imports made unused by wrapper serialization. These
  introduced compile errors were corrected before successful execution evidence.
- One added HTTP fixture initially decoded the contract-3 decimal string capability
  revision into Go's internal uint64 Session field. Corrected the fixture to decode
  only the compact HTTP fields it asserts; no production wire contract was weakened.
- Comprehensive directed command:

```text
go test -tags dev -run '^(TestAgentBinding.*|TestSQLiteSpawnFrozenSettingsAndReceipt|TestSQLiteAutomationManagementHTTP|TestGeneralToolScopeAndAtomicSettings)$' -count=1 -timeout 90s .
tool timeout 120 seconds; exit 0; final comprehensive run: ok klm/engine 6.864s
```

This covered manual/helper + HTTP creation, source adoption preserving original
bytes, independent override/reset/CAS, next-turn model/effort capture, fixed-harness
block, missing-marker/copied-UID/I/O block, queue retention, deleted fallback,
default-false and explicit true external inheritance, receipt replay after source
deletion, real Library authoring/outbox marker preservation and hidden-reference
cleanup, mutable-reference backfill, local v1-format snapshot/offline hash stability,
and frozen automation continuation. The three existing directed tests cover ordinary
spawn receipts, automation management and general tool/settings scope.

After actual final isolation/Stop changes, affected checks only:

```text
go test -tags dev -run '^(TestAgentBindingCreationOverridesAndLifecycle|TestAgentBindingAutomationKeepsAdmittedCapture)$' -count=1 -timeout 60s .
tool timeout 90 seconds; exit 0; ok klm/engine 3.066s

go test -tags dev -run '^TestAgentBindingSpawnReceiptAndLibraryMarker$' -count=1 -timeout 60s .
tool timeout 90 seconds; exit 0; ok klm/engine 1.580s
```

The last test additionally exercised saved-agent general-create dispatch with an
actual fixture human event and explicit project/no-topic result. Tests compile all
current Go product sources, so no redundant separate build was needed. No full suite
or benchmark was run. `git diff --check -- "engine"` from root: timeout 10 seconds,
exit 0, only Git LF/CRLF notices. No packages, servers, providers, installed data,
commits or pushes were used. Fixtures use temporary SQL/authoring/skill roots, a fake
read-only pinned package manifest and cached model data; no native harness starts.

A final explicit-default edge was corrected and checked in the affected HTTP
fixture: if the catalog cannot name its native default effort, `effort:""` remains
an explicit native-default override rather than an invalid pair or an invented
replacement. A nonempty unsupported effort still blocks. Directed command
`go test -tags dev -run '^TestAgentBindingHTTPAndReferenceBackfill$' -count=1 -timeout 45s .`
(tool timeout 60 seconds) passed, exit 0, `ok klm/engine 1.432s`.

### Boundary for manager validation / CP2

Manager integration follow-up received after the CP1b report: the exact available
`TestDelegationContracts` target was run once against the current shared package:

```text
go test -tags dev -run '^TestDelegationContracts$' -count=1 -timeout 45s .
workdir engine; tool timeout 60 seconds; exit 0; ok klm/engine 0.106s
```

This confirms the package compiles with B's additive records/tools/Codex/assets and
A's concrete `AgentExecutionCapture` type. The transient `sqlite_agents.go`
undefined-json error reported during concurrent editing is absent in this completed
check. No B source, domain, native manifest or activation was edited/integrated by
A. A passing Go contract target is not evidence that the JS fail-closed latch or
native continuation was fixed; those remain B/manager review boundaries.

Actual native provider messaging, generated managed-profile execution, retained
history resumption, client UX, real permissions/ACLs and process/lifecycle behavior
are not established by these fixtures. Manager review/correction and the final
separate runtime/Playwright checkpoint remain required. Global graph v2/frozen full
delegation catalog and B child activation/finality are not part of CP1b. Codex's
new-thread-only delegation decision remains B's separate boundary; A did not migrate
threads, change pins or edit Codex adapters. No schema B domains were invented.

## CP1b manager review — A-B1 / A-B2 corrections

Approval was retained after the manager reproduced two failures in
`engine/agent_binding_manager_probe_test.go` (both failed, 1.343 seconds, manager
evidence). The earlier automation probe prepopulated a run snapshot; it proved
continuation reuse but did not prove capture retention in real admission.

### A-B1 — compact acceptance/replay response

The original 128 KiB valid specialist prompt appeared twice in the tool result,
producing 262,651 bytes against the 64 KiB tool bound. `spawnReceiptResult` now
projects only `agent:{source,name,revision}` and
`configuration:{harness,model,effort}`. Existing accepted/session/message/topic/YOLO
and general identity fields remain, and the durable receipt still stores its full
capture unchanged. No prompt, skills or MCP catalog is echoed or truncated into a
possibly unusable receipt. All normal/general/legacy creation and replay paths use
this one helper. SessionAgentState/public creation DTOs are unchanged.

### A-B2 — retain the actual admission capture

`refreshAutomationAgentLocked` now returns its concrete preparation rather than
discarding it, together with destination/harness/CWD, automation + trigger digests,
trigger identity and Stop epoch. `AgentTurnPreparation.SelectionRevision` is read
inside the same SQL transaction that commits the prepared binding/selection.

Before consuming a trigger, admission rechecks:

- Exact destination, fixed harness/CWD, current binding and prepared desired digest.
- Binding CAS and the selection CAS recorded by preparation.
- Automation/trigger identity + content digests and the destination Stop epoch.

Changed inputs leave the trigger unconsumed. A normal new bound run can be admitted
only with that preparation's explicit external snapshot; the SAME snapshot is
persisted as `run.Capabilities` alongside its specialist configuration. Dispatch and
continuation read the admitted snapshot. `prepareAutomationAgent` no longer calls
`captureCapabilities` to fill nil on a saved-agent run. Existing captured recovery
data is reused; a missing old/partial saved-agent capture returns explicit
`automation_capture_missing`, requiring a fresh admission decision instead of silently
using current mutable Library content. Legacy unbound automation behavior remains.

### Focused regression coverage

The manager's original assertions were preserved and formatted. Added controls:

- Large valid prompt + large bounded selection/catalog metadata still produce a
  compact result with `accepted` and `sessionId`, for normal/general response shapes.
- Actual `admitSQLAutomationLocked` with a selected fixture MCP connection persists
  the prepared snapshot. Editing that connection's revision/URL after admission
  leaves dispatch's snapshot revision/URL unchanged.
- Removing the admitted external capture from an old/partial run is an explicit
  failure, never a permission to recapture live configuration.

One directed selection after these corrections:

```text
gofmt -w "session_collaboration.go" "agent_types.go" "agent_turn.go" "agent_automation.go" "sqlite_automations_worker.go" "agent_binding_manager_probe_test.go"
  timeout 10 seconds, exit 0

go test -tags dev -run '^(TestAgentBinding.*|TestSQLiteSpawnFrozenSettingsAndReceipt|TestSQLiteAutomationManagementHTTP|TestGeneralToolScopeAndAtomicSettings|TestDelegationContracts)$' -count=1 -timeout 90s .
  workdir engine; tool timeout 120 seconds; exit 0
  ok klm/engine 10.752s
```

This compiled the combined A+B package and covered the two review regressions plus
affected binding/automation/spawn/general contracts. No redundant build, full suite,
provider/server or B source change occurred. Public leaf/graph controls were not
loosened; internal bound-source read helpers remain generic for engine-owned callers.

Files changed this correction round: `session_collaboration.go`, `agent_types.go`,
`agent_turn.go`, `agent_automation.go`, `sqlite_automations_worker.go`, the authorized
manager probe, and A's contract/progress. Shared runtime files remain ready only for
a manager-assigned handoff after revalidation; A still owns persistence/schema.
No CP2/global graph/delegation domain or migration was added.

## CP2a — delegation storage/schema foundation

The manager approved CP1b after revalidation (manager-owned focused execution:
exit 0, 3.215 seconds), then released the separate CP2a storage checkpoint. No global
graph v2/catalog work or B execution activation was included.

**Exact schema/index/query/helper/recovery handoff:**
`AGENTS_UNIFICATION_DELEGATION_STORAGE_HANDOFF.md`.

Implemented:

- Fresh schema version 3; preserved original v1 -> v2 A step; explicit v2 -> v3
  selected-domain upgrade. Existing binding/native/history bytes remain unchanged.
- Concrete B roots `delegationBatches`, `delegationInvocations`, `delegationChildren`
  plus split `batch.calls` -> `delegation_calls`. No legacy native-child backfill.
- Explicit `x_state` projections and operation/native-alias/member-key/member-order/
  active-parent/active-child/retained-lease uniqueness. Indexed owner pages, graph
  epochs, membership order, alias delivery and bounded recovery candidates.
- Optional `nativeSession.delegation` manifest roundtrip. Existing native history
  cannot gain one retroactively; instrumented manifest/thread identity cannot be
  silently dropped/replaced. Structurally valid incompatible metadata remains
  evidence; B's handler, not storage, decides tool compatibility/availability.
- A-owned typed readers/writers/CAS helpers and generic-write shape/immutability
  checks. Transaction-final validation checks related identities/ownership/contiguous
  membership/leases atomically. Ordinary event streaming does not scan old ledgers.
- Exact ordinary/side/general/graph ownership, allowed canonical cwd/workspace,
  project/parent roles and no child-owned topic/graph selector/spawn receipt. Legacy
  ParentID exceptions are not universally expanded; only validated managed rows
  receive the graph-child case. History cwd uses the retained workspace while active
  epoch validation uses the actual activation/turn.
- Accepted aliases compare only ParentTurnID/SessionID/TurnID/CallID. Delivery is
  mutable with normal transitions, and numeric-zero CallID survives. Batch
  DeliveryTurnID and invocation EventStart/EventEnd remain payload, not indexes.
- Captured invocation reference indexes and credential retention for active or
  quarantined leases reuse normal SQL authority. No helper fabricates drain evidence
  or clears an uncertain lease by assertion.
- Metadata-only pending/recovery page + readiness-barrier seams exist. They are
  **not wired into B-owned startup/runtime/readiness** in this checkpoint. Nil hook
  never settles work; pending work makes the explicit barrier fail. B must add its
  evidence prepass and post-owner-epoch interruption hook before activation.

Ownership arbitration followed: `sqlite_delegation.go` and
`sqlite_validate_delegation.go` remain A-only. Current A contents were preserved;
B's new `delegation_repository.go` / `delegation_prepare.go` were not edited.
Existing readers `delegationBatch`, `delegationInvocation`, `delegationChild` and
`delegationMembers` are published in the handoff and must not be duplicated.

### Files changed in CP2a

New: `engine/sqlite_delegation_upgrade.go`, `sqlite_delegation.go`,
`sqlite_validate_delegation.go`, `sqlite_delegation_recovery.go`,
`sqlite_delegation_probe_test.go`, and the A-owned storage handoff document.

Updated: `store.go`, `sqlite_domains.go`, `sqlite_schema.go`, `sqlite_store.go`,
`sqlite_agent_upgrade.go`, `sqlite_indexes.go`, `sqlite_record_writes.go`,
`sqlite_session_repository.go`, `sqlite_validate.go`, `sqlite_prepare.go`,
`sqlite_library_references.go`, `sqlite_outboxes.go`, and the existing
`agent_catalog_storage_probe_test.go` to reconstruct genuine old schemas and expect
the current explicit final schema version. Contract/progress were updated.

No edits to B-owned main/harness/runtime/adapters/private dispatch/API registration/
queue/stop/finish/requests, public child role guards, frontend or graph runtime.
No B service, model loop, native launch or scheduling worker was implemented by A.

### Directed evidence

All commands had finite timeouts, no real data/providers/native execution.

1. gofmt explicitly named A files: timeout 10 seconds, exit 0.
2. Initial focused execution exposed a new index name collision with the generic
   domain factory (`*_identity` already existed). Renamed new UNIQUE indexes to
   `*_record_unique`; no table/data workaround or IF NOT EXISTS hid the conflict.
3. The full session validator's preexisting TEMP table writes cannot run on a
   query-only reader. The fixture was corrected to use its intended writer
   transaction, without weakening ownership assertions or production read-only DBs.
4. Final complete directed selection:

```text
go test -tags dev -run '^(TestDelegationStorage.*|TestAgentCatalogCP1aStorageProbe|TestDelegationContracts)$' -count=1 -timeout 90s .
workdir engine; tool timeout 120 seconds; exit 0; ok klm/engine 10.252s
```

Coverage: fresh schema, synthetic v1/v2 reconstruction, preserved binding/native/
session digests, empty managed defaults, optional manifest roundtrip, no legacy
adoption/no manifest drop, explicit state rather than status, CallID zero and mutable
Delivery, conflicting parent/child leases, contiguous-member gaps, foreign cwd,
leaf-as-parent and own-selector rejection, valid side/general/exact graph ownership,
wrong node activation and unchanged legacy graph ParentID rejection. Pending recovery
with nil hook fails rather than pretending successful cleanup.

After strengthening ONLY the upgrade fixture to close/reopen actual v1/v2 files and
exercise the authority-checked `enableRuntime` upgrade path as well as idempotent
steps, reran its affected target:

```text
go test -tags dev -run '^TestDelegationStorageSchemaUpgradeAndManifest$' -count=1 -timeout 60s .
tool timeout 90 seconds; exit 0; ok klm/engine 3.943s
```

`git diff --check` over A's changed tracked Go files: timeout 10 seconds, exit 0,
only Git LF/CRLF notices. No redundant full suite/build, installation, foreground
server, real migration, commit or push. The tests compile the current combined
package, including B's additive files, without editing their sources.

Schema-ready is not feature activation or runtime acceptance. Pending review/handoff:
B recovery wiring and physical evidence, runtime/service integration, any future
late-finality release evidence contract, manager review, and CP2b global graph
snapshot/catalog release. A remains sole shared storage/migration owner.

## CP2b — global graph agents and frozen run specialists

The manager approved CP2a storage, then released this distinct graph artifact/catalog
slice. Exact DTO/helper/ownership/evidence handoff:
`AGENTS_UNIFICATION_GRAPH_AGENT_SNAPSHOT_HANDOFF.md`.

Implemented global + local graph Save/compile/new capture/catalog input metadata.
New SQL snapshots are v2, with explicit origins/canonical global node sources and
the complete eligible visible specialist catalog (including non-node definitions),
resolved config/Library captures, digest/completeness/diagnostics. Node execution
Agents/EffectiveAgents remain separate. Legacy v1 raw source/hash handling stays
unchanged, and graph reads do not adopt local markers.

Catalog limits are 256 eligible entries / 16 MiB serialized entry metadata; overflow
fails preflight rather than truncates. Unusable candidates are diagnosed, while a
referenced node remains strict. Snapshot-only lookup/composition never falls back
to live authoring. Explicit parent inheritance uses only this run's parent node
capture, rejects internal or uncaptured assets/tools, and preserves contract identity.
Existing reservation revalidation checks the whole catalog/global sources and
captured Library. Reference/credential projection includes the new frozen captures.

Existing scheduler, runner, Choice, worktree and snapshot-store mechanisms were not
redesigned or edited. No graph child execution/tool registration was activated.
C's temporary graph global-agent picker filter can be removed after backend review:
actual Save + new capture now support those visible definitions.

Verification: `go test -tags dev -run '^(TestGraphFrozenGlobalSaveCaptureAndLookup|TestAgentBindingLocalMarkerSnapshotValidation)$' -count=1 -timeout 90s .`
from engine, tool timeout 120 seconds: exit 0, `ok klm/engine 3.093s`. One central
controlled fixture covers Save/compile/global v2 capture, non-node frozen lookup,
invalid candidate diagnosis, no marker adoption, no live inheritance/rebind and
offline immutability; the existing v1-format test covers compatibility. Initial
compile-only bytes-name shadow was fixed. Gofmt and tracked-file diff whitespace
check passed; no full suite/redundant build/provider/graph execution was run.

New files: `graph_snapshot_v2.go`, `graph_frozen_agents.go`,
`graph_frozen_capabilities.go`, `graph_frozen_agents_probe_test.go`, graph handoff doc.
Updated: `graph_snapshot.go`, `graph_contract.go`, `graph_tools.go` (catalog only),
`authoring.go` (Save only), `sqlite_library_references.go`, contract/progress.
No B/runtime/API/bridge/graph_adapter/execution/transition/worktree/mcp_client/
agent_tools/frontend files changed. B helper signatures and caller responsibilities
are in the handoff; actual activation and finality remain separate manager/B work.

**Stopped at CP2b. Await explicit manager review/correction or next release.**
