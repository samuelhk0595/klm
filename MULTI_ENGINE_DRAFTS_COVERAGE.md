# Native editable-state coverage

CP2b implementation matrix, 2026-10-30, approved by coordinator static inspection
including its corrections. CP3 passed coordinator static inspection. Independent
directed native validation and CP4-01 retest are recorded with their actual subset
and limits in `MULTI_ENGINE_PLAYWRIGHT_VALIDATION.md`; they do not establish every
field/variant or final human acceptance. Paths below are relative to
`clients/desktop/src/`.

All keys are inside the same DPAPI document scoped to desktop profile/app ID,
canonical endpoint and confirmed database identity. They are not web-storage keys.
`form.*` holds editable values/bases/necessary receipts; `ui.*` holds current
locators. Several keys may exist for different entities without overwriting one
another. No historical-version log is stored.

## Shared lifecycle

- `features/drafts/draftStore.ts`: shared `useEngineDraft`, 400 ms serialized
  debounce, confirmed `flushDrafts`, native-only clear/discard. Defaults are kept
  per key; explicit base retention does not replace an existing key. Setters and
  asynchronous discard closures are bound to their namespace. Generic validation
  checks outer containers only, not a schema inferred from defaults; optional
  clearing and RAW child values are allowed. Directory variants, elicitation
  dictionaries, graphs and questions use small feature-specific predicates.
  Invalid data is not replaced with a default document.
- `App.tsx`: desktop editable content waits for authenticated bootstrap and native
  draft hydration. Root editor locators are restored from `ui.*`, including view,
  session/topic/project create/edit, project settings section and Settings. Rename
  locators retain only ID/title, not conversation history. Project editor locators
  retain only project fields, not topic/history windows.
- Close/X and section changes preserve native values. Explicit Cancel/Discard and
  confirmed save clear the corresponding edit. Browser/mobile keep their existing
  local-state behavior. Restoration never calls a submit/apply/authorization or
  execution handler.

## Matrix

| Path / surface | Protected fields and key | Base / operation identity | Cleanup | Locator / return |
| --- | --- | --- | --- | --- |
| `App.tsx`, chat adapter from CP2a/R1 | `chat.composers`, `chat.sources`, `chat.submissions`; normal/general/side text, mentions and passages | Session ID, original delivery ID/mode/composer key | Shared HTTP/queue/history acceptance; explicit restore/discard | Workspace navigation plus session/side destination; no automatic resend |
| `features/workspace/CreateSessionDialog.tsx` | `form.session.new:<project>:<topic>.title/harness` | Project/topic destination in key/locator | Confirmed Create or Cancel | `ui.createSession`; component keyed by project/topic |
| `features/workspace/RenameSessionDialog.tsx` | `form.session.rename:<id>.title/baseTitle` | Original title; fresh session read before native rename detects a changed title | Confirmed Rename or Cancel | `ui.renameSession` ID/title only |
| `features/workspace/TopicDialog.tsx` | `form.topic.new:<project>`, `form.topic.edit:<project>:<topic>`, `form.topic.choose:<project>:<session>`; mode, per-mode name/color, selection and chosen-topic metadata | Retained topic base/revision; exact topic read before native Edit save when revision exists | Confirmed create/edit/move or Cancel | `ui.createTopic`, `ui.editTopic`, `ui.chooseTopic`; per-mode keys keep create text when returning from Choose |
| `features/projects/ProjectDialog.tsx` | `form.project:<id-or-new>.name/folder/icon/browsing` | Original project fields in root locator; latest known project metadata checked before native Edit | Confirmed Save/Create or Cancel | `ui.addProject` / `ui.editProject`; directory modal state retained |
| `features/projects/DirectoryBrowser.tsx` | `<project-form>.directory.path/location/selected/query/descending/edit/folderName/uncertain` | Current listing target and folder-action parent/item; existing uncertainty flag | Explicit folder-action cancel, confirmed navigation after mutation, or parent discard | Read-only roots/listing hydration does not overwrite typed path, selection or folder editor; no create/rename/delete on restore |
| `features/projects/ProjectGeneralSettings.tsx` | `form.projectGeneral:<id>.name/icon/base` | Retained name/icon base; detects changed current project props before Save | Confirmed Save or native Discard | `ui.projectSettings` + `ui.projectSettingsSection:<project>` |
| `features/projects/ProjectSettingsDialog.tsx` | Current General/Webhooks/Tasks/Automations section | Project ID | Selection updates only; closing keeps child drafts | Same project/section; unavailable project ID is not silently purged |
| `features/agents/AgentEditorDialog.tsx`, `AgentsPage.tsx` | `form.agent:<project>:<agent-or-new>.values/revision`; name, description, harness, model, effort, prompt, skills/MCP selections | Original authoring-file revision retained and sent on PATCH | Confirmed Save/Create or Cancel; X keeps edit | `ui.agents:<project>.editing`; editor keyed by agent/new |
| `features/graphs/GraphsPage.tsx`, `GraphDetailsDialog.tsx` | `form.graphDetails:<project>:new.name/description`; graph-scoped details name/description | Returned staged graph ID retained in current editing locator | Confirmed graph-draft creation, local Apply into canvas, or Cancel | `ui.graphs:<project>.creating/editing` |
| `features/graphs/GraphCanvas.tsx` | `form.graph:<project>:<graph>.base/nodes/edges/definition/viewport/layout/layoutError/configuring/detailsOpen` | Original graph file revision; latest confirmed base replaces it after Save; current scene retains local IDs/ports/handles | Save updates base/clears the explicit definition edit while preserving current scene and unrelated unapplied child drafts; native Discard removes the full graph edit | Current graph/node/details/viewport; no render callbacks, DOM or React Flow instance serialized |
| `features/graphs/AgentNodePanel.tsx` | Name, agent and overrides already controlled by the protected canvas node data | Canvas graph/node ID | Part of explicit canvas Save/Discard | Canvas `configuring` locator; no independent local editable state |
| `features/graphs/ChoiceNodePanel.tsx` | `<graph>.node:<id>.choice`; name/description, terminal, input/output fields, policy | Same original graph revision in parent | Apply into canvas clears this subedit; Cancel discards node subedit fields; X preserves | `features/graphs/draftScope.ts` supplies stable graph/node context |
| `features/graphs/TerminalNodePanel.tsx` | `<graph>.node:<id>.terminal`; name, raw command and output mappings | Parent graph revision/node | Same Apply/Cancel/X rule | Same node context |
| `features/graphs/ForkNodePanel.tsx` | `<graph>.node:<id>.fork/expanded`; branches, worktree/branch names, output mappings | Parent graph revision plus stable branch IDs | Same Apply/Cancel/X rule | Branch-scoped context preserves independent branch previews |
| `features/graphs/JoinNodePanel.tsx` | `<graph>.node:<id>.join`; agent, prompt, output branch | Parent graph revision/node | Same Apply/Cancel/X rule | Same node context |
| `features/graphs/ChoiceOutputEditor.tsx`, `ChoiceOutputPreview.tsx` | Node/branch `.payloadField:<field-id>` and `.preview.task/result/input`; raw JSON remains a string | Graph/node/branch/field IDs | Explicit node discard; raw preview inputs are not erased by unrelated graph Save | Current scene IDs retained to avoid canonical-ID orphaning |
| `features/webhooks/WebhooksSettings.tsx` | `form.webhook:<project>:<id-or-new>` name/retention/enabled/changeSecret/secret; `form.webhooks:<project>.error` | Original editor metadata includes identity/updatedAt; this API has no CAS | Confirmed mutation callback or explicit Cancel; generated secret is explicitly retained at first mount | `ui.webhooks:<project>.editor/selected`; transport uncertainty survives without auto retry |
| `features/activity-board/ProjectTaskConnections.tsx` | `form.taskConnection:<project>.values/revision/needsReload/catalog/labelName`; provider, scope, PAT/token, mappings and pending label text | Original config revision and connection ID | Confirmed Save, explicit Cancel; native Discard and reload only by user action | Active connection draft belongs to its project; missing connection retains native fields and blocks saving |
| `features/activity-board/BoardSettingsDialog.tsx` | `form.taskPhases.phase/revision/needsReload/mappings/catalogs` | Phase ID and original config revision | Confirmed Save or Cancel; missing phase is retained on reload | `ui.taskPhases`; one active phase editor, other entity changes require explicit Cancel |
| `features/activity-board/ActivityBoardPage.tsx` | `form.taskMove.attempt/receipt/unknown/missing/notice`; item/version, target phase, operation ID | Original move operation ID, config revision and observed external version | Existing receipt reconciliation/cancel controls; only one current attempt | Task view/project/sprint locators; attempt persisted before dispatch; restore never moves an item |
| `features/automations/AutomationEditor.tsx` | `form.automation:<project>:<id-or-new>.base/values/step/topicMode/topicName/topicColor/promptField` | Original automation revision and creation operation ID; full activation/schedule/session destination/prompts/harness/model/effort/YOLO/graph/Library selections | Parent confirmed Save result or explicit Discard; closing project settings preserves | `ui.automation:<project>.selected/editing`; catalog load still happens when restoring an open editor |
| `features/automations/AutomationsSettings.tsx` Run now | `form.automationRun:<project>` current target/revision/operation ID plus persisted error | Original ID reused for uncertain retry; fresh local non-dispatch/known rejection releases it; exact handler revision conflict proves no receipt | Confirmed response or proven rejection; a later NEW explicit action can use fresh revision/ID; no auto request | Existing automation detail/history controls; auth/missing target/other conflicts cannot disprove earlier uncertainty; no universal ledger |
| `features/settings/SettingsDialog.tsx` Transcription | `form.transcription.settings/base/apiKey/removeKey`; raw model/language/vocabulary and pending secret | Saved base settings; fresh comparison before Save is best-effort client reconciliation, not CAS | Confirmed Save clears sensitive edit keys; native Discard and reload explicitly removes the edit | `ui.settings`, `ui.settingsSection`; GET cannot overwrite an existing draft |
| Same Settings dialog, native connection management | The native control uses the shared saved-engine UI; older `form.settings.engineURL` data is not purged | Current immutable endpoint | Native selection uses the global confirmed transition; browser keeps its legacy Connect | Same Settings section opens the global manager; no independent endpoint selector |
| `features/settings/EngineAccessSettings.tsx` | `form.access.origins/proxies/base/restarting` | Saved access-config base; restart recovery flag is not a stored confirmation/consent | Confirmed Save or native Discard and reload; reconnect only reads status | Same Engine section; confirmation dialog itself is not persisted; no auto Restart |
| `features/automations/EngineTimeSettings.tsx` | `form.engineTime` current time zone + revision | Original engine settings revision retained in edited object | Confirmed Save or native Discard and reload | Settings Chat section |
| `features/library/LibraryPanel.tsx` / MCP editor | `form.mcp:<id-or-new>` name/config/raw args/environment/headers/bearer/clientSecret/secret-change flags/revision | Original MCP revision; editor locator projects only ID/name/config/revision/credential flag, not tool catalogs | Confirmed Save or Cancel; closing Settings keeps edit | `ui.libraryEditing`, `ui.libraryAuthorization`; saved OAuth URL is only a link and never starts OAuth on restore |
| `features/library/SessionLibraryDialog.tsx`, `features/chat/SessionOptions.tsx` | `form.sessionLibrary:<session>.base/selection` including desired skills/MCP/tools | Original capabilities revision, never replaced by a later GET while editing | Confirmed Save or explicit Discard/reload confirmation; native X/Esc preserves | `ui.sessionLibrary:<session>`; automation read-only rule remains active |
| `features/chat/QuestionCard.tsx` | `form.question:<session>:<question>:<items-fingerprint>` selected/custom answers, including secret input | Question ID + exact item shape; minimum array/item validation | Confirmed reply/dismiss; disappeared question keeps its saved entry | Only a still-pending matching question mounts; hydration does not reply |
| `features/library/MCPElicitationCard.tsx` | `form.elicitation:<session>:<question>:<schema-fingerprint>` raw field values | Question ID/schema identity | Confirmed explicit accept/decline/cancel only | Only matching pending elicitation renders; no automatic authorization or external-page open |

## CP3 connection fields and transition coverage

| Path / surface | Protected fields / identity | Completion and return |
| --- | --- | --- |
| `features/engines/connectionDrafts.ts`, `EngineConnectionDialog.tsx` | Profile-owned null-database document; Add slot keyed by runtime endpoint, Edit/Pair-again by registration ID; local ID/name/icon/URL, Client name, raw code, code URL association and attempted-code marker | Close preserves; confirmed metadata/connection completion or explicit Discard clears that form only. Failed pre-dispatch attempt does not burn the local marker; uncertain dispatch is not replayed. |
| `features/engines/EngineProvider.tsx` registry recovery | Raw replacement JSON draft and exact original registry strings, protected before explicit replacement | Restore validated JSON or explicitly start an empty list; never silent fallback. Originals are deduplicated only for explicit recovery, not automatic per-edit snapshots. |
| `features/engines/switchEngine.ts`, `localActivities.ts` | Actual in-memory save/answer/chat/layout/icon/Tasks/automation promise leases; recording/native-picker blockers; no persisted generic mutation ledger | Freeze -> finish local work -> detach/stop reads/SSE/renewal -> seal/flush entity + connection drafts -> commit -> drain credential cleanup -> reload. Waiting is cancellable without server cancellation/rollback claims. |

## Boundaries and pending work

- CP3 is implemented and statically approved. Directed native rounds concluded
  with CP4-01 corrected/retested; unexercised paths and human acceptance remain
  pending. Pre-auth forms never hydrate entity drafts before
  auth/DB confirmation and do not replace the active entity namespace.
- Mutation/audio/layout guards now coordinate the global transition. They do not
  introduce universal server timeouts, replay or a receipt API. Native credential
  cleanup/public-registry failure is not advertised as an atomic rollback.
- Missing/archived/conflicting entities keep their draft entries and original
  bases. They are never automatically reapplied. Existing entity restoration or
  reopening recovers their values; no generic orphan-draft explorer was added.
- Session/project/webhook/access/transcription APIs do not all provide atomic CAS
  or queryable creation receipts. Local checks and retained fields cannot invent
  these guarantees. Transport uncertainty remains explicit/manual, with original
  concrete IDs reused where the API already supports them.
- Run now checks its receipt before the exact `Automation changed.` revision
  rejection, so that conflict releases the attempt. Auth/missing automation or
  project responses may happen before receipt lookup; an earlier uncertain ID is
  retained in those cases. If that target stays unavailable there is no separate
  public receipt query in the existing API. No automatic abandonment is assumed.
- Theme/background/feed preferences are already immediate local preferences;
  browser-only Focus widgets/photos and cosmetic filters did not receive a new
  persistence abstraction. Uncommitted HarnessSettings work was not imported.

## Evidence

Implementation checks `npx tsc --noEmit` and `git diff --check` passed. The independent
tester subsequently exercised real native/live paths at `72a959f9` and focused
CP4-01 retest paths at `d2519de1`. See `MULTI_ENGINE_PLAYWRIGHT_VALIDATION.md` for
outcomes, environment and explicit unverified rows. This matrix is implementation
coverage, not a claim that every field/variant passed runtime acceptance. No full
suite or installation acceptance is claimed.

CP2b correction R1/R2 additionally passed the two focused pure-function files
`draftShape.test.ts` / `runNowAttempt.test.ts` (nine cases). They cover valid
variant/RAW/default transitions and definite-versus-uncertain Run-now outcomes;
they do not constitute application or native runtime validation.
