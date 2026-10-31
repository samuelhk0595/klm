# Agents unification — frontend progress

Owner: stream C, conversation `413fd7541b23359f48df60d266d93775`.
Manager: `a05369eaa5c9151b61e93fb38cb35c52`.

## CP1a — frontend foundation

Status: **approved by manager after direct diff review**. No blocking correction.
Scope is the manager's explicit CP1a subset following CP0 UX approval. No dependent
CP1b work is released. The remaining candidate API/types/events in
`AGENTS_UNIFICATION_UI_CONTRACT.md` are still proposals.

### Implemented

1. Added `features/agents/types.ts` as the production owner of `AgentDefinition`,
   `Agent` and `agentSlug`. The slug implementation is unchanged. The definition
   contains the existing prompt/harness/model/effort/skills/MCP settings; identity,
   revision and graph-reference metadata stay in the record type.
2. Updated all production consumers that previously imported agent types/slug from
   `agents/demo.ts`. No `agents/demo` imports remain under shared frontend source.
3. Fixed AgentsPage's enable/disable definition payload to include `skills` and
   `mcpConnections`, preserving the existing arrays, connection IDs and tool names
   while changing only `enabled` in the submitted configuration. No metadata fields
   are added to the authoring definition payload.
4. Removed the catalog hook's implicit fixture fallback for a missing project.
   Without a project or explicit preview provider, the hook returns no models and
   an English context-unavailable error when a harness is selected. It makes no
   guessed global/project request.
5. Retained demo agent/model data and `modelsForHarness` in `demo.ts`. Added the
   ready-to-use `demoAgentModels` map and an explicit `DemoAgentModels` context in
   `catalog.tsx`. Preview callers can opt in with
   `<DemoAgentModels value={demoAgentModels}>...</DemoAgentModels>`. Project context
   always uses the real project model endpoint, even inside a preview provider.
   There are no current in-repository callers of these agent demo records; no live
   authoring screen is wrapped in a preview provider.
6. Updated the UI contract with the approved navigation/scope decisions, CP1a
   boundary, authoritative lifecycle/eligibility dependency and optional-only
   rendering of B's actual semantic results.

### Files changed by C in CP1a

All feature paths below are under `clients/desktop/src`:

- `features/agents/types.ts` — new production types/slug module.
- `features/agents/demo.ts` — fixtures consume production types; explicit demo map.
- `features/agents/catalog.tsx` — explicit preview injection; no implicit fixtures.
- `features/agents/AgentEditorDialog.tsx` — production types/slug import.
- `features/agents/AgentsPage.tsx` — complete toggle definition including Library.
- `features/graphs/agent-node.ts` — production type import.
- `features/graphs/AgentNodePanel.tsx` — production type import.
- `features/graphs/ChoiceNodePanel.tsx` — production slug import.
- `features/graphs/files.ts` — production type/slug imports.
- `features/graphs/GraphCanvas.tsx` — production type import.
- `features/graphs/GraphDetailsDialog.tsx` — production slug import.
- `features/graphs/GraphsPage.tsx` — production type import.
- `features/graphs/GraphView.tsx` — production type import.
- `features/graphs/intermediate-choice.ts` — production slug import.
- `features/graphs/JoinNodePanel.tsx` — production type import.
- `AGENTS_UNIFICATION_UI_CONTRACT.md` — manager decisions and checkpoint boundary.
- `AGENTS_UNIFICATION_UI_PROGRESS.md` — this evidence/report document.

No backend, Flutter, product definition, manager plan/coordination, dependency or
generated output file was edited by C. Preexisting unrelated documentation/probes
remain preserved.

### Checks and outcomes

- `git status --short` from repository root, timeout **10000 ms**, exit **0**:
  recorded the preexisting work before edits; frontend files above were clean then.
- `git diff -- "clients/desktop/src/features/agents" "clients/desktop/src/features/graphs"`
  from repository root, timeout **10000 ms**, exit **0**: inspected the scoped diff.
  Git printed normal LF-to-CRLF working-copy notices.
- Dedicated content search under `clients/desktop/src` for `agents/demo` returned
  no matches after import updates. The catalog has no runtime import of demo data.
- **TypeScript check:**
  `& ".\node_modules\.bin\tsc.cmd" --noEmit --pretty false --project "tsconfig.json"`
  from `clients/desktop`, timeout **120000 ms**, exit **0**, no diagnostics.
  Uses already installed dependencies; no installation, watch, emitted build files
  or build-info changes were needed.
- Final scoped status:
  `git status --short -- "clients/desktop" "AGENTS_UNIFICATION_UI_CONTRACT.md" "AGENTS_UNIFICATION_UI_PROGRESS.md"`
  from repository root, timeout **10000 ms**, exit **0**: exactly the 15 frontend
  files and two UI documents listed above; no generated output changes.

No test suite, browser/runtime/provider execution or installation was performed.
The TypeScript result establishes compilation/types, not user-visible acceptance.

### Manager/human validation and next gate

- Review the small extraction/import diff, explicit preview boundary and exact
  skills/MCP preservation in the toggle payload.
- Later runtime acceptance should toggle an unreferenced agent with skills and
  explicit MCP tools and confirm the saved selections remain unchanged; exercise
  real project model discovery and explicit preview injection separately.
- Global management, source binding/recovery, agent session creation and child UI
  wait for A/B interface consolidation and a new explicit release. Source deletion
  is never inferred from conflict/I/O, and semantic failure is never invented.
- Report CP1a once to manager, then end the turn without polling.

## Post-CP1a catalog merge update

Manager supplied additional contract decisions without releasing CP1b. Updated
only this progress file and the UI contract to record:

- `sourceUid` is the approved opaque identity beside public slug; globals use
  SQLite and locals TOML.
- Presence-aware override transport: omitted = unchanged, `null` = reset/follow,
  empty string/list = explicit override; skills/MCPs independently reset.
- Unknown I/O, invalid and conflict retain capture for readable history but block
  new preparation. Confirmed deleted/disabled may use last validated configuration.
- A must supply real paginated/exact catalog/model APIs and a concrete invalidation
  key; final wire types still await its CP1a report. No candidate endpoint is wired.

No source code changed for this update. The successful CP1a TypeScript result above
remains the latest compilation evidence; no overlapping check or runtime was run.
The original CP1a report was accepted/queued by manager's session. Send one concise
merge-update acknowledgment and stop; CP1b still requires a new explicit release.

## CP1b — real global/local agent management

Status: C-R1/C-R2 manager corrections delivered; awaiting manager revalidation.
Manager explicitly released management only after reviewing A's actual API and
R1/R2 corrections. No session/binding/runtime UI is included in this release.

### Implemented behavior

- Settings -> Agents manages real globals without requiring a selected/registered
  project. Project Agents displays globals plus its locals with server-provided
  origin, status and configuration metadata; no global is hidden from management.
- Search is debounced, sent to the server before pagination, and resets to page 1.
  Previous/Next read bounded 50-row pages. Refresh/invalidation retains the current
  page/filter; an empty later page keeps Previous available.
- Every row action reads the exact source by slug + sourceUid (and the source's
  projectId for a local). Editor/toggle/delete never depend on a full loaded catalog
  or infer existing-source scope from the currently displayed project. Search can
  find sources outside the original loaded page, followed by exact full detail GET.
- New project-view agent offers Scope This project/Global. Global Settings creates
  globals; existing Scope is plain read-only origin text. The editor keeps Library
  available globally and chooses the actual global/project model endpoint according
  to selected creation scope or existing source metadata.
- CRUD uses full definitions and exact update/delete revision/sourceUid, including
  locals. Enable/disable first reads the full definition, preserves Library and
  compares against the displayed revision before inverting state; a stale row asks
  for Refresh rather than toggling a different newly changed state.
- Structured backend failures retain English error/hint, conflicting exact sources,
  and project-qualified blocking references. Editor errors retain the user's draft;
  Discard changes and reload is explicit. No save/delete/uncertain request is
  automatically replayed. Global/project writes have finite timeout signals.
- Editor displays structured graph references and loads additional bounded reference
  pages through exact GET, guarding the agent revision. Mutation admission remains
  server-authoritative if the live reference set changes between reads.
- Child edit/delete dialogs stop close/cancel propagation to the containing Settings
  dialog and disable cancellation/duplicate confirmation while their mutation runs.
- Management's mounted bounded page is its isolated cache; it subscribes to existing
  `global/` and, for projects, `project/<id>` invalidation via `subscribeScope`.
  Existing complete authoring caches now isolate by engine URL/identity/generation
  and subscribe to those same shared scopes. Model caches isolate by engine and
  explicit global/project scope. No additional EventSource implementation, polling
  monitor or fixture fallback was introduced. Explicit preview injection remains.

### Actual transport wired

- `GET /api/agents?projectId=&search=&offset=&limit=` -> summary page with diagnostics,
  complete, opaque revision, hasMore/nextOffset. Omit projectId for global Settings.
- `GET /api/agents/{slug}?projectId=&sourceUid=&offset=&limit=` -> full definition,
  flat source `{sourceUid,id,scope,projectId?,projectName?}`, status/availability and
  bounded structured references. Local source supplies its own projectId.
- Global create `POST /api/agents` `{agent}`; PATCH/DELETE `/api/agents/{slug}` with
  `{agent,revision,sourceUid}` / `{revision,sourceUid}`.
- Local create/update/delete retain `/api/projects/{sourceProjectId}/agents[/slug]`
  with the same full body and exact identity for existing sources.
- Global model discovery `GET /api/models/{harness}`; project model discovery keeps
  `/api/projects/{sourceProjectId}/models/{harness}`. C did not execute either path
  against real providers in this phase.

All names were checked against `engine/agent_catalog_api.go`, `agent_types.go` and
`AGENTS_UNIFICATION_CATALOG_PROGRESS.md` HTTP section, not CP0 candidate shapes.
`EngineRequestError.details` preserves the structured response without changing
existing error message/status handling elsewhere.

### Files changed in this checkpoint

Under `clients/desktop/src`:

- `engine.ts` — retain structured error details (no new session/runtime APIs).
- `features/agents/types.ts` — actual source/summary/detail/page/diagnostic types.
- `features/agents/api.ts` **new** — exact read and scope-aware management mutations.
- `features/agents/useAgentPage.ts` **new** — bounded page fetch and shared invalidation.
- `features/agents/AgentProblemMessage.tsx` **new** — English structured error display.
- `features/agents/AgentsPage.tsx` — paged real management, exact details and CRUD.
- `features/agents/AgentEditorDialog.tsx` — Scope, real models, global Library,
  structured references and explicit stale reload.
- `features/agents/DeleteAgentDialog.tsx` — nested-dialog propagation/busy controls.
- `features/agents/catalog.tsx` — explicit global model scope and engine-isolated cache.
- `features/agents/agents.css` — shared responsive management/detail styling.
- `features/settings/SettingsDialog.tsx` — global Agents entry.
- `features/graphs/ProjectAuthoring.tsx` — separate paged management from full graph
  catalog consumers; preserve local graph authoring selection pending CP2.
- `features/graphs/catalog.ts` — engine-isolated cache and existing scoped invalidation.

Also updated this document and `AGENTS_UNIFICATION_UI_CONTRACT.md`. CP1a import-only
changes remain intact. C did not edit backend/Flutter/product/manager documents.

### Graph impact reserved for CP2

A's existing `/api/projects/{id}/authoring` response now includes globals with
metadata and keeps complete arrays. Management no longer consumes that large array.
Graph Save/execution is not yet global-enabled by A/B. ProjectAuthoring therefore
passes only project/local or legacy metadata-less rows to the existing GraphsPage
picker, retaining local use without offering a known unsupported global Save path.
This filter applies only to graph authoring, never to agent management. Existing
read-only chat GraphView still receives the complete catalog through its current
hook; CP1b does not change saved graph definitions, captured history or execution.

CP2 must integrate actual global graph save/catalog/snapshot execution contracts and
their diagnostics, including externally authored global references. Current catalog
errors still follow existing graph recovery/blocking behavior; this checkpoint does
not promise complete global graph support or silently rewrite an unavailable node.

### Verification

- `git status --short`, repository root, timeout **10000 ms**, exit **0** before work:
  recorded CP1a and concurrent A/B/manager changes; preserved all unrelated work.
- `git diff --stat -- "clients/desktop/src"`, repository root, timeout **10000 ms**,
  exit **0**. Its tracked summary includes retained CP1a changes; new untracked C
  modules are itemized above. Git emitted normal LF-to-CRLF notices.
- **One TypeScript check:**
  `& ".\node_modules\.bin\tsc.cmd" --noEmit --pretty false --project "tsconfig.json"`
  from `clients/desktop`, timeout **120000 ms**, exit **0**, no diagnostics.
  No install, watch, build output or build-info mutation was required.
- `git diff --check -- "clients/desktop/src"`, repository root, timeout **10000 ms**,
  exit **0**, no whitespace error (normal LF-to-CRLF notices only).
- `git status --short -- "clients/desktop" "AGENTS_UNIFICATION_UI_CONTRACT.md" "AGENTS_UNIFICATION_UI_PROGRESS.md"`,
  repository root, timeout **10000 ms**, exit **0**: cumulative C frontend changes
  include the retained CP1a slice plus the explicit CP1b inventory, with no generated
  output changes.

No runtime, provider discovery, real agent mutation, Playwright, full suite or
backend test was executed by C. Actual UI/HTTP behavior remains for the separately
released runtime/manual acceptance stage. TypeScript is compilation evidence only.

### Next gate

Report this CP1b implementation and stop for manager review/correction. Do not add
agent session creation/binding/settings/subagent UI or invoke candidate endpoints.
Later Codex `subagent_run` visibility must come from factual backend capabilities
on newly instrumented native threads, never harness/model inference; old histories
and ordinary chat remain readable/usable without the tool. No such control is added
in CP1b. A/B's further actual interfaces and another explicit release are required.

## CP1b manager review — C-R1/C-R2 corrections

The manager did not approve the initial CP1b delivery and supplied two focused
regressions in `features/agents/AgentsPage.manager.test.tsx`. Attributed manager
evidence: the corrected-setup run reached both product defects, **2 failed / 3.75s**
(same-revision discard kept the draft; invalidations aborted the active read).
The earlier missing-jsdom-showModal setup failure was a fixture issue, not product
failure evidence. The supplied polyfill and both original assertions are preserved.

### Corrected behavior

- **C-R1:** AgentsPage increments an editor reset nonce only after the human's
  explicit discard/reload has read the source successfully. It remounts/reset all
  local editor state (draft/error/reference list and reference cursor), including
  when the UID/revision is unchanged. Mere scoped refresh or incidental rerender
  does not increment the nonce or discard an in-progress draft.
- The discard GET has its own AbortController, aborted on leaving/replacing the
  editor or unmounting its project/engine scope. A late response must also match
  the captured engine identity/generation before it can publish; it cannot reopen
  an editor after navigation or apply a previous engine's reload.
- **C-R2:** useAgentPage now owns one reader per engine/scope/query/page key.
  Same-key invalidations/manual refreshes mark `dirty` while in flight and do not
  abort/recreate that read. Its result is published, then at most one follow-up is
  started for all notices accumulated during that read. Notices during the next
  read coalesce the same way; there is no timer/polling monitor.
- Key changes/unmount still abort the active read, drop its pending dirty refresh
  and ignore stale completions even if a transport/mock returns after cancellation.
  Guards check both latest rendered key and current engine identity/generation.
- A settled page remains visible and its row/menu/page actions remain usable while
  the background refresh is pending; exact GET and revision checks still protect
  mutations. Status distinguishes Refreshing from the initial Loading state.

### Correction files

Only the approved findings and their focused evidence were edited in this round:

- `clients/desktop/src/features/agents/AgentsPage.tsx`
- `clients/desktop/src/features/agents/useAgentPage.ts`
- `clients/desktop/src/features/agents/AgentsPage.manager.test.tsx` — integrated the
  manager-owned file, preserved its original two regressions and added assertions
  for cleared error, one follow-up/read publication, draft preservation, usable
  cached rows, late reload after scope exit and project/search/page/engine changes
  plus unmount cancellation.
- `AGENTS_UNIFICATION_UI_CONTRACT.md`
- `AGENTS_UNIFICATION_UI_PROGRESS.md`

No backend, session binding/runtime UI, graph expansion or other source was changed.

### Focused verification

Commands from `clients/desktop`, using installed dependencies:

1. `& ".\node_modules\.bin\vitest.cmd" run "src/features/agents/AgentsPage.manager.test.tsx"`
   — tool timeout **90000 ms**, exit **0**, initial correction **8 passed / 3.11s**.
2. `& ".\node_modules\.bin\tsc.cmd" --noEmit --pretty false --project "tsconfig.json"`
   — tool timeout **120000 ms**, exit **0**, no diagnostics.
3. After addressing the related remaining background-refresh action lockout and
   adding its focused assertion, reran the same regression command only:
   **9 passed / 3.87s**, exit **0**, timeout **90000 ms**.
4. Reran the same TypeScript command for that final source/test change:
   timeout **120000 ms**, exit **0**, no diagnostics.

The second pair follows an actual source/assertion change, not repeated broad
verification. These mocked DOM regressions did not contact engine/provider data or
exercise real browser/runtime/Playwright behavior. No suite-wide test run or install.

Report the correction once and stop for manager revalidation; no later checkpoint
is self-approved or released by these passing checks.

## CP1b approval recorded on authorized resumption

Manager approved CP1b after directly reading the corrected sources/assertions and
personally running `vitest.cmd run src/features/agents/AgentsPage.manager.test.tsx`:
**9 passed, exit 0, 2.95s** (manager-attributed evidence). No TypeScript rerun was
needed at approval. This remains mocked component evidence, not CP4 browser/runtime.

## CP2a — manual agent creation and linked-session settings

Status: implemented and targeted checks complete; awaiting manager review.
Release: frontend creation/binding/settings only, using A's reviewed definitive
HTTP shapes. Concurrent backend receipt/automation-capture corrections were not
awaited, consulted or edited by C. No child-batch or global-graph UI is released.

### Implemented

- Create session has a real globals+locals picker using paged server search and
  exact slug/sourceUid detail reads. Unavailable sources cannot be selected; an
  in-progress exact lookup blocks Create so the previous/default source is not
  accidentally submitted. Default keeps ordinary title/topic/harness behavior.
- POST uses `agent`, `agentSourceUid`, `agentRevision`, optional nested
  `agentOverrides`. Displayed source defaults do not create overrides. Top-level
  harness is omitted while following the source, and is sent only for an explicit
  creation override; model/effort/prompt/skills/MCPs use the nested object only.
  Explicit overrides survive choosing another specialist; returning to Default
  removes agent-only fields while retaining title/topic/ordinary harness choice.
  Creation sends no message and exposes no parent inheritance toggle.
- Shared AgentOverridesEditor preserves full prompt text and independently edits/
  resets model, effort, prompt, skills and MCPs. Empty strings/lists are explicit;
  null reset/follow is sent only by the session settings path, while creation reset
  removes that draft override. Clear skills/Clear MCPs can explicitly pin empty
  selections even when inherited selections were already empty.
- SessionAgentDialog consumes `{agent:...}` detail and the nested
  `applied.config` execution capture, not a flat applied configuration. It displays
  specialist/source origin/status, fixed harness, pending/next-turn settings,
  overrides, confirmed saved deleted/disabled fallback, full applied prompt/Library
  and any captured external parent contribution. It offers per-field reset and
  explicit whole-override reset; Library-only reset cannot touch model/prompt.
- Bound ModelPicker opens that shared field-specific editor. Effort-only edits
  send effort only. Selecting an incompatible model retains the requested effort
  and requires an explicit compatible effort/default choice, rather than silently
  substituting it. No postcreation source switch/detach/harness control was added.
- Bound Session Library uses the same `/agent` revision/partial PATCH path, avoiding
  the existing full `/capabilities` PATCH which would pin both families. Unbound
  session/model/Library behavior keeps its existing endpoints.
- Main, side and Focus/mobile-web composers reuse AgentMessageComposer and shared
  identity/recovery controls. New sends/Send now are blocked for unavailable or
  denied eligibility; drafts, Stop, request replies and queue removal remain usable.
  Recovery does not replay sends or uncertain deliveries. Side initial harness
  choices are hidden when a binding fixes it. General identity/initial harness is
  otherwise unchanged; child/graph-node sessions acquire no public editing control.
- Bound views share an engine/session-keyed detail reader with existing
  global/project/conversation scope subscriptions, 250ms event debounce and one
  in-flight read plus coalesced dirty refresh. It never applies settings on GET.
  Explicit PATCH increments a client epoch so an older GET cannot overwrite its
  saved result; last-listener cleanup/key changes abort/drop old readers. There is
  no polling monitor, extra EventSource or automatic configuration-apply loop.
  Resubscription reinstates the same entry after StrictMode cleanup; stale cleanup
  cannot delete a newer entry. This matches the application's actual StrictMode root.
- Live detail/eligibility updates preserve an editor's edit-start draft. Changed
  revision/token/source state requires explicit discard/reload before saving.
  Running bound model controls show applied config; desired config stays next-turn.
- Model catalogs are isolated by engine identity and actual session/scope rather
  than mixing native histories/workspaces. No fixture fallback exists in production.
- Library deletion impact labels `global_agent` as Global agent and `agent_binding`
  as Agent-linked session. No removal or recovery operation was executed by C.

### Definitive transport

Read directly from `agent_bindings.go`, `agent_resolution.go`, `agent_types.go` and
A's Definitive HTTP progress section. `Session.agentConfig` is compact metadata;
`GET /api/sessions/{id}/agent` returns full desired/overrides/provenance and optional
`applied: {config,source?,revision,...}`. Binding writes are
`PATCH /api/sessions/{id}/agent` `{revision,overrides}` with independent presence,
null and explicit-empty semantics. Source status is preserved without treating I/O,
invalid or conflict states as deleted. Eligibility uses the backend's allowed/code/
message/fields/token, not harness/model heuristics. Manual POST uses the exact
creation fields above, never a CP0 candidate route or nested agent object.

### Files changed in CP2a

Product files under `clients/desktop/src` (19):

- `App.tsx`, `engine.ts` — actual create DTO, compact agentConfig, composer/send gates.
- `features/agents/types.ts`, `agents.css` — definitive DTOs and shared visual tokens.
- `features/agents/AgentPicker.tsx` **new** — paged eligible selection/exact fetch.
- `features/agents/AgentOverridesEditor.tsx` **new** — field-level edits/reset.
- `features/agents/SessionAgentDialog.tsx` **new** — settings/provenance/recovery.
- `features/agents/binding.ts` **new** — shared scoped detail cache and partial writes.
- `features/agents/AgentMessageComposer.tsx` **new** — shared identity/eligibility UI.
- `features/workspace/CreateSessionDialog.tsx`, `FocusSessionChat.tsx`.
- `features/chat/ModelPicker.tsx`, `modelCatalog.ts`, `SessionOptions.tsx`,
  `SessionStatusBar.tsx`, `SideChatPanel.tsx`, `MessageQueue.tsx`.
- `features/library/SessionLibraryDialog.tsx`, `LibraryPanel.tsx`.

Focused tests/fixture (4 new): `features/agents/CreateAgentSession.test.tsx`,
`SessionAgentDialog.test.tsx`, `binding.test.tsx`, `sessionAgentFixtures.ts`.
Fixtures are explicitly test-only imports, not production defaults. Also updated
the UI contract and this progress file. Backend/Flutter/manager documents and
CP1b's provisional local-only graph selector filter remain unchanged.

### Checks and outcomes

Commands from `clients/desktop`, using installed dependencies; no watch/install:

1. `& ".\node_modules\.bin\vitest.cmd" run "src/features/agents/CreateAgentSession.test.tsx" "src/features/agents/SessionAgentDialog.test.tsx" "src/features/agents/binding.test.tsx"`
   — timeout **90000ms**, **8 passed / 2 failed, 3.17s**. The two failures were
   jsdom fixture setup (`ResizeObserver` absent in Menu), not a provider/runtime
   defect. Added/restored a test-local observer stub without changing assertions.
2. Reran only the affected file:
   `& ".\node_modules\.bin\vitest.cmd" run "src/features/agents/SessionAgentDialog.test.tsx"`
   — timeout **90000ms**, exit **0**, **5 passed / 2.80s**.
3. `& ".\node_modules\.bin\tsc.cmd" --noEmit --pretty false --project "tsconfig.json"`
   — timeout **120000ms**, initial nonzero result with two introduced diagnostics:
   required SessionCapabilityPicker `general` prop and an unknown-valued JSX guard.
   Fixed with explicit external-only `general={false}` and a boolean guard.
   Directed rerun: exit **0**, no diagnostics.
4. After fixing a further source-selection-in-flight submission race and adding its
   focused assertion, reran only creation tests:
   `& ".\node_modules\.bin\vitest.cmd" run "src/features/agents/CreateAgentSession.test.tsx"`
   — timeout **90000ms**, exit **0**, **4 passed / 4.42s**. The unchanged binding
   tests passed **2 cases** in step 1; the five settings cases passed in step 2.
5. Same TypeScript command after that actual creation change:
   timeout **120000ms**, exit **0**, no diagnostics.
6. After inspecting `main.tsx` and correcting same-entry StrictMode resubscription,
   added the corresponding cache-publication/eligibility assertion and ran only:
   `& ".\node_modules\.bin\vitest.cmd" run "src/features/agents/binding.test.tsx"`
   — timeout **90000ms**, exit **0**, **3 passed / 3.85s**. Repeated the same
   TypeScript command for that actual hook/test change: timeout **120000ms**,
   exit **0**, no diagnostics (final compilation evidence). No unrelated test was
   repeated. Final targeted evidence is 4 creation + 5 settings + 3 reader cases,
   passed in their separate executions, not one asserted combined run.

Additional repository commands, timeout **10000ms** each, exit **0**:
`git diff --check -- "clients/desktop/src"` (no whitespace errors; normal LF/CRLF
notices), `git status --short -- "clients/desktop"` (cumulative stream changes;
no generated output), and scoped `git diff` for App/ModelPicker/CreateSessionDialog.

Meaningful assertions cover unmaterialized defaults, no duplicate override fields,
title/topic/Default preservation, source lookup admission, effort-only PATCH,
explicit incompatible-effort recovery, independent Library clear/reset, nested
applied prompt, draft retention on source updates, shared/coalesced reads and stale
  GET vs saved PATCH ordering. These are mocked component/transport-state checks,
not real engine/provider/native/browser behavior. No full suite, Playwright,
installation, backend execution, production data or provider call was used.

### Checkpoint boundary

Report CP2a once and stop for manager review. Child batch UI, graph-global selectors,
permission-origin projections and factual Codex child capability remain later
releases. Browser/desktop/narrow behavior and live API/harness lifecycle remain CP4
runtime/manual acceptance; neither compilation nor these fixtures establishes them.

## CP2b — enable the delivered global graph catalog

CP2a was approved by the manager after the main-path review; its reported checks
were accepted without requesting another matrix. Manager then approved A's global
graph Save/capture backend and released this one-line frontend adjustment.

- `clients/desktop/src/features/graphs/ProjectAuthoring.tsx`: removed the temporary
  global-source filter. GraphsPage now receives the existing complete global+local
  catalog unchanged, retaining IDs/source metadata and existing local/read-only flows.
- No new type, endpoint, configuration, batch/subagent UI or graph execution.
- Inspected the affected file and applied the single-line diff. No new check/test
  was run, as explicitly permitted for this import/filter-only slice.

CP2b supersedes the earlier local-only filter notes in this document/UI contract.
Report the small change and stop; runtime browser evidence and child UI remain
separate released work.
