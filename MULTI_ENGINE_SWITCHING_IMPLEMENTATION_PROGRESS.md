# Saved engines implementation progress

## Scope and checkpoint state

Desktop Windows/Tauri first: saved engines, one immutable active endpoint per page
load, and a full reload when switching. The approved feature plan is
`MULTI_ENGINE_SWITCHING_PLAN.md`.

- CP0 approved by the coordinator.
- CP1 approved after the coordinator's static inspection; no runtime acceptance implied.
- CP2a, including R1, approved by the coordinator's static inspection.
- CP2b and its R1/R2 corrections approved by coordinator static inspection.
- CP3 UI/transition/recovery implemented; awaiting coordinator inspection and later independent runtime validation.
- Worktree: `C:\Users\Samuel\Documents\Projects\Personal\klm-multi-engine`.
- Branch: `feat/desktop-multi-engine`.
- Base: `952ff2bffae89d9b646ea9202b0187b65c0b5069`, head of PR #39,
  `fix/harness-large-events` targeting `release/0.6.0`.
- No merge has been performed. Feature-branch commit/publication state is recorded in the CP3 checkpoint report.

## CP1 — Registration, resolution and authentication

### Registration and migration

`clients/desktop/src/features/engines/engineRegistry.ts` stores version 1 public
metadata in one localStorage record, `klm.engines.v1:dev` or
`klm.engines.v1:release`. Tauri app identities/origins and native credential keys
retain their existing isolation too.

Each entry has a local stable ID, a trimmed required name (up to 80 Unicode code
points), a canonical URL, and an optional predefined icon ID. There is no entry
count limit. Validation rejects duplicate IDs/URLs, unknown icons, unsupported
versions, and a dangling active ID. Equal names and distinct URL aliases do not
merge entries. Only whitelisted public metadata is serialized.

A missing record imports the previously resolved desktop URL once. Its initial
name comes from the hostname. The legacy URL, native credentials, endpoint/DB
preference keys and engine identity are not rewritten. A valid record with
`activeEngineId: null` is an intentional unselected state, even if it has zero
entries; it is not eligible for migration again.

Invalid/unreadable registry or desktop legacy input fails closed. Initialization
does not replace it with an empty record or silently connect to localhost.
Migration succeeds only after the record is written and read back successfully.
An unsuccessful initialization leaves no usable runtime endpoint. Normal requests
then fail before native credential lookup or fetch, allowing the existing
connection gate to display the error.

### Endpoint resolution

The existing normalizer moved unchanged in behavior to the pure leaf module
`clients/desktop/src/engineURL.ts`, re-exported by `platform.ts`. This avoids a
platform/registry import cycle and preserves HTTPS/loopback, explicit ports and
path-prefix policy.

Desktop resolution happens once during module initialization, before auth,
`ENGINE_URL`, and engine identity capture the endpoint. Later registry edits cannot
retarget any request from that page. On desktop only, `VITE_ENGINE_URL` is a seed
when the registry is absent; an existing saved selection takes precedence. Browser,
Focus and mobile retain their legacy override/URL behavior and create no registry.
The download support introduced by PR #39 is unchanged.

The legacy desktop connection forms can re-pair/retry only the already-active
endpoint at this checkpoint. They cannot select another endpoint and reload ahead
of CP2's confirmed draft flush. CP3 will route them through the single selector
and transition flow. Error/unselected recovery UI is also part of CP3.

### Authentication separation

`clientAuth.ts` now separates active re-pairing from two desktop candidate helpers:

- `pairCandidateClient`: redeem the candidate's code, store its native credential
  under that canonical endpoint, and verify it without selecting it.
- `verifyCandidateClient`: retry verification from the already-saved credential,
  without consuming another code after a later verification/registry failure.

Candidate handshakes never use the active bearer or DB ID and never change active
auth generation/blocking or start an inventory/update stream. Native handshakes
reject redirects. Verification requires a successful JSON client-session response.

Active re-pairing finishes the initial keyring read, invalidates older responses,
waits for pending deletion before storing a replacement, and serializes attempts.
Failure restores prior usable auth state while keeping pre-pair responses obsolete;
an already-denied state remains denied. Existing Credential Manager storage is
unchanged; no credentials enter the new registry.

## Evidence and remaining validation

CP1 checks in `clients/desktop`:

- `npm ci --ignore-scripts --no-audit --no-fund`: installed worktree-local
  dependencies from its existing lockfile; no lockfile change.
- `npx tsc --noEmit`: passed.
- `npx vitest run src/features/engines/engineRegistry.test.ts src/platform.test.ts src/clientAuth.test.ts`:
  passed, 3 files / 31 cases, using mocked fetch and native keyring.
- `git diff --check`: passed for tracked changes.

These focused module checks cover canonical URL identity/deduplication, migration,
explicitly removed selection, invalid/unavailable storage, immutable endpoint,
desktop override seeding versus browser policy, candidate isolation/recovery, and
active late-response/keyring-deletion ordering. They are not desktop/runtime or
human acceptance evidence. No full suite, application startup, Playwright, native
keyring operation, production engine connection, installer, or VPS operation ran.

Native protected drafts and all surface adapters must precede exposing engine
switching. Publication through
The initial unscoped `gh` calls could not find authentication in the harness's
projected config. The coordinator later resolved that lookup in CP3 using the
existing user login, as recorded below; no token/config was copied or inspected.

## CP2a — Protected storage and chat drafts

### Native file and format

`clients/desktop/src-tauri/src/draft_storage.rs` adds only three native commands:
`draft_read`, `draft_write`, and `draft_delete`. The renderer supplies a validated
`{endpoint, databaseId}` namespace and JSON document, never a path.

Files live under the desktop's own Tauri `app_data_dir()/drafts/`, on Windows
normally `%APPDATA%/<app-identifier>/drafts/<namespace-hash>.dpapi`. This is separate
from engine data and from the embedded static assets served by Focus. The hash
includes app identifier, canonical endpoint, database ID and scope discriminator;
its bytes are also DPAPI entropy. `databaseId: null` is reserved for future
pre-connection forms and cannot collide with an entity namespace. CP2a entity
adapters always require a confirmed database ID.

DPAPI uses CurrentUser with UI forbidden, never LocalMachine. Writes encrypt
before creating a temporary file in that same private directory, sync ciphertext,
then use `MoveFileExW(REPLACE_EXISTING | WRITE_THROUGH)` for atomic replacement.
Only ciphertext reaches disk, including temporary files. Native operations are
serialized. Missing file is the sole empty-vault condition; access errors,
decryption failure, invalid UTF-8 and oversized data are explicit failures. The
document cap is 32 MiB, with bounded encrypted-file reads; exceeding it is an
error rather than truncation. Non-Windows native protection returns unsupported.

The plaintext document inside DPAPI is `{version: 1, entries: {...}}`. It keeps
current editable values and necessary pending-delivery records, not historical
snapshots. An empty entry map can delete the namespace file through the same
serialized flush path. The existing native credential store is unchanged.

### Store, hydration and flush ordering

`features/drafts/draftStore.ts` contains one module-owned store and a small typed
`useEngineDraft` hook. All mounted native editors share it; browser/mobile use
their prior component-local state instead. The generic hook does not persist its
default value on mount or after an async read.

Desktop bootstrap in `App.tsx` confirms `/api/state.engineId`, binds the identity,
then loads and validates the protected document before loading conversation UI.
Entity draft reads/updates reject an unconfirmed or mismatching namespace.
Desktop persists its cached identity only after native legacy migration succeeds,
so failed first adoption can still import the original records after reload. A
separate inventory identity keeps database-reset handling pending across a failed
native read, instead of mixing old session windows on the next bootstrap retry.
Replacing the database at the same URL first flushes any old dirty state under
the old namespace, then loads the replacement's separate file. Old entity data is
never hydrated into the replacement. Failed old flush retains its data/context
for Retry; failed new read/decryption does not become an empty vault.

Edits update shared memory synchronously and debounce native persistence by 400 ms.
Only one IPC write is in flight; edits arriving during it become the next latest
snapshot. `flushDrafts()` cancels the debounce and awaits all dirty revisions,
propagating errors. It must succeed before a future destructive transition.
Failed writes retain dirty memory and expose the existing inline error/Retry
pattern; editing does not silently discard that data or fall back to web storage.
No generic recovery/transaction framework or recovery UI was added.

### Chat coverage and pending deliveries

Current protected entries:

- `chat.composers`: drafts by main/general session ID and `side:<mainId>`,
  preserving text and file/directory mentions.
- `chat.sources`: selected main-chat passages for each side composer.
- `chat.submissions`: original delivery IDs, draft, sources, mode, status and
  originating composer key for current pending submissions.
- `chat.legacyMigration`: one bounded receipt that prevents reimport of obsolete
  web-storage delivery records after migration.

`features/chat/chatDrafts.ts` validates these shapes before hydration. Prior
`sending`/`accepted`/`uncertain` submissions reload as `uncertain`, retaining the
same IDs; definite failures remain failed. Hydration never sends or retries them.
HTTP, history and queue confirmations share the same local acceptance routine;
explicit Retry delivery / Restore draft controls remain available for unresolved
records.

Before desktop HTTP submission, the current draft and pending ID are flushed to
native storage. Failure at this point prevents dispatch and retains the composer.
The native composer and side sources now remain until actual acceptance; an
unconfirmed network outcome keeps both and its original pending ID. Acceptance
clears only a matching composer payload and the submitted source references, so
new edits are not overwritten. Restore draft also works when that same retained
draft is already visible, and does not duplicate its source passages. Equal
pending payloads cannot be submitted as a new implicit delivery.

`MessageComposer.clearOnSend` defaults to the prior behavior. Only actual desktop
chat entry points disable eager clearing; browser/mobile and the design-system
example retain their previous behavior. All session/overlay/general views use the
same App-owned hooks rather than separate editor stores. Audio devices, DOM,
callbacks, refs and recordings are not serialized; completed transcription text
uses the normal draft path. Transition-time handling of active native interactions
belongs to CP3, before exposing the switch.

For existing desktop pending records, current DB-scoped sessionStorage data is
merged by delivery ID into the native document. An unscoped legacy key is adopted
only during the first confirmed identity binding. The native migration marker and
records are flushed before old keys are removed; write/parse/cleanup failures
retain recoverable data and stop bootstrap completion. Desktop identity migration
no longer copies these secret-bearing records into another web-storage key, and
desktop calls to the old plaintext read/save helpers are rejected. Browser/mobile
retain the legacy helpers unchanged in behavior.

### CP2a evidence and next checkpoint

- `npx tsc --noEmit`: passed after the storage/chat integration.
- `cargo check`: passed in the isolated worktree, compiling the real Windows
  DPAPI/file replacement code and IPC handlers. The baseline embeds `../dist`
  even during check; an ignored temporary marker supplied that otherwise-absent
  asset directory for compilation and was removed afterward. No frontend runtime
  or built installer was produced or launched.
- `git diff --check`: passed for tracked changes.

No application, native storage operation, full suite, Playwright, live engine,
VPS or installer was exercised. Compiler success does not establish DPAPI runtime,
failure-recovery, reload or end-to-end chat acceptance. No new tests/review agents
were added or run during initial CP2a delivery; the subsequent explicitly requested
R1 function check is recorded below.

Await coordinator inspection and explicit CP2b release. All other form adapters,
including their base revisions and sensitive fields, remain required. CP3 remains
blocked; it must also use neutral registry-confirmation failure copy instead of
claiming a persistent selection rollback after a failed readback.

### CP2a correction R1 — Shared acceptance from HTTP, queue and history

The coordinator's static inspection found that queue/history reconciliation removed
the pending receipt without also clearing its retained composer and sent sources.
A lost POST response/reload could therefore leave submitted text as an ordinary
draft after its duplicate guard disappeared.

`features/chat/chatAcceptance.ts` now computes acceptance as one pure operation
over the three chat maps. `App.tsx` invokes it for both successful POST responses
and IDs observed independently in queue/history. It uses the original composer key,
or the side session's parent for legacy receipts. Matching compares text and
individual mention fields, independent of JSON property ordering. Only matching
composer content is cleared; newer edits survive. Sent sources are subtracted by
value and occurrence count, keeping new passages and newly-added duplicate copies.

Cleanup is computed before receipt removal, then `draftStore.setDraftValues`
publishes all three resulting maps in one revision/debounced native document.
There is no nested store write inside a React state updater. After a receipt is
removed, repeated HTTP/queue/history confirmations return the unchanged state,
including when a new identical draft was entered later. A late POST failure does
not reclassify or report a delivery already confirmed by queue/history.

R1 touched only `App.tsx`, `draftStore.ts`, `chatDrafts.ts` (comparison helper moved),
the new pure `chatAcceptance.ts` / `chatAcceptance.test.ts`, and this progress file.
Native storage, engine APIs and browser/mobile behavior were not changed.

Checks:

- `npx tsc --noEmit`: passed.
- `npx vitest run src/features/chat/chatAcceptance.test.ts`: passed, one file /
  six pure-function cases (reload/history ACK without POST, new draft, legacy side
  queue ACK with new sources, HTTP ACK, repeated ACKs, unrelated receipts).
- No app/native runtime, network engine, Playwright or full suite was exercised;
  no Rust changes required repeating the earlier Cargo check.

R1 and CP2a were subsequently approved by the coordinator. No runtime acceptance
was implied by that static inspection.

## CP2b — Remaining editable surfaces

The existing store/hook now covers the remaining forms in the approved inventory.
The field/key/base/cleanup/locator matrix is in `MULTI_ENGINE_DRAFTS_COVERAGE.md`.
No Go/native APIs, harness-installation files, engine selector or pairing UI were
added in this checkpoint.

Hook additions remain local and small: a per-key fallback, explicit retained base
values, a native-only clear/prefix discard, outer-container checks with optional
feature predicates, and namespace-bound setters/discard callbacks. Changing an
entity/key cannot reuse the previous entity's initial value. A callback from an
obsolete database context cannot write into the replacement namespace. Editable
desktop content mounts only after bootstrap and draft context are ready.

GET initialization checks existing draft keys. Agents/graphs/MCP/Library/Tasks/
automations/time-zone forms retain their original revisions/operation IDs. Where
the API lacks CAS, saved base values and best-effort client checks do not constitute
atomic server-side conflict prevention. Explicit discard/reload actions preserve
the existing server contract; no silent rebase is performed on mounted drafts.

Native close/section navigation keeps field values; existing explicit Cancel or
Discard actions and confirmed save/apply paths clear the corresponding data.
Graph subeditor Apply moves values into the protected canvas state before clearing
that subeditor. The current graph scene/local node IDs stay alongside the latest
saved base so canonical server Choice IDs cannot orphan raw subeditor/preview
inputs after a reload. This is one current scene, not a sequence of snapshots.

Existing concrete Tasks move IDs/receipts, automation creation/run operation IDs,
directory uncertainty and layout/restart recovery flags are retained. Restoring
them does not replay a mutation. Run-now recovery reuses the original request ID;
queue/history chat acceptance from R1 remains unchanged. JSON text fields are kept
as strings, including invalid JSON. API keys, PATs, MCP headers/environment/bearer/
client secrets and secret answers use the native cofre only.

Browser/mobile continue using local React state, their existing Cancel/default
behavior and legacy connection policy. New error/discard controls are native-only.
Native missing/archived/conflicting entities do not cause automatic submission or
purging of their draft keys. Existing entity restoration/reopening can recover the
stored fields; no generic draft explorer/recovery UI was introduced.

Checks: `npx tsc --noEmit` and `git diff --check` passed. TypeScript was repeated
only after final changes to scene retention/base guards and namespace fallbacks.
No app, native IPC, live engine, Playwright, full suite, Cargo recheck or additional
review was run for this TypeScript-only checkpoint. CP3 still must wire transition
guards for in-flight mutations/audio, the pre-connection namespace, recovery
selector and confirmed flush before changing the active registry entry.

### CP2b corrections R1/R2

R1 removes field-schema inference from initial values. The pure
`features/drafts/draftShape.ts` helper checks only the outer kind and allows
optional/nullable clearing; object keys, optional fields and child default types
are not a schema. Thus roots -> directory does not require `initial`, numeric MCP
defaults can become RAW strings, and collapsing a branch initialized with a string
ID can be cleared. Fork collapse uses null so JSON persistence retains the
collapsed state instead of omitting an undefined entry. Directory target variants have an explicit small
predicate; elicitation requires only a values dictionary. Existing graph/question
predicates remain feature-specific. Stored RAW values are neither parsed nor
replaced with defaults by this check.

R2 separates new Run-now attempts from earlier uncertain ones. A fresh attempt
that fails before entering the request, or gets a known pre-admission 400/401/403/
404/409 response, is removed from memory/protected state. After refreshing, a new
explicit action can allocate a new ID with the current revision. There is no new
automatic request. A transport failure, 408/429, 5xx or unknown outcome retains
the original ID/revision.

For an earlier uncertain ID, the exact `409 "Automation changed."` response also
permits release: `runSQLAutomationNow` searches the manual-operation receipt
before that revision check. Generic 4xx is insufficient evidence for an older
request. In particular `st.automation`/auth can return 401/403/404 before receipt
lookup, and another 409 may refer to identity/project state. Those remain
uncertain. If that target stays unavailable, the existing API does not expose a
separate receipt query to resolve it safely; no automatic abandonment/new ID or
new recovery API was invented.

Contract evidence: `engine/sqlite_automations_api.go` lines 595-640 and
`engine/sqlite_app.go` error mapping; no Go code changed. The feature-local pure
`runNowAttempt.ts` classifier is called only with status/message from
`EngineRequestError`; ordinary transport error text is not treated as a rejection.

Checks: `npx tsc --noEmit` passed; `npx vitest run
src/features/drafts/draftShape.test.ts src/features/automations/runNowAttempt.test.ts`
passed (two pure-function files / nine cases), including the optional Fork
collapse refinement. Targeted tracked-file `git diff --check` passed. No
application, live API, native operation, full suite or review agent was run.
CP3 remains blocked.

## CP3 — Saved-engine UI, transitions and recovery

Implemented under the coordinator's explicit CP3 release, after CP2b corrections
were approved. This section supersedes the earlier checkpoint's blocked status.

### UI and metadata

- Native `EngineProvider` lives outside `ClientConnection` and the App draft gate.
  It owns connection dialogs and the recovery selector even on cold/offline/401,
  empty selection, invalid registry or a failed workspace render.
- `EngineSwitcher` places an always-visible avatar between Theme and Settings in
  Project Rail. The same control is available outside the rail while loading or
  disconnected. The searchable/scrollable list uses only local metadata, without
  an entry-count cap, inactive-engine probes or online claims. Menu/list keyboard
  navigation, focus return and English labels reuse the design-system controls.
- Add/Edit/Remove use required local names, predefined Lucide icons or Unicode
  initials. Client name remains a separate paired-device field. Edit cannot
  silently change an endpoint. Current selection is a no-op by immutable runtime
  URL, even if a previous readback failed after a durable selection write.
- Settings' native Connect form is replaced by the same management flow.
  Browser/Focus/mobile keep their legacy URL selection and presentation.

### One transition sequence

`switchEngine.ts` freezes user events/new actions, waits for registered actual
local activities and allows Cancel switch during that wait. It then synchronously
detaches the workspace/auth subtree, pauses/aborts GET transport, closes scoped
SSE/reconnect timers, waits for any native draft hydration, seals late entity edits,
and confirms entity plus pre-auth flushes. Only then does it commit the registry
selection/local operation, drain native credential deletion, stop transport and
reload. Endpoint/auth/identity modules remain immutable for the current page.

Actual save/answer handlers, chat submission queues, icon preparation, task and
automation handlers and graph layout queues hold in-memory activity leases through
their draft/state cleanup. Native mutation RPCs provide a small fallback lease for
other real HTTP operations. No server receipt ledger, generic timeout/retry or
job-completion monitor was introduced. Automatic general/side initialization is
suppressed while changing. A pending server agent/graph is not itself a local
activity and keeps running after selection changes.

Recording/microphone prompts and native/icon pickers are explicit blockers until
the user finishes/cancels them. Finite transcription is awaited through append of
its unsent text. Cancel switch never cancels a sent mutation, Stop/Restart or the
remote agent; it restores the old UI without claiming rollback. Storage/commit
failure unseals/remounts the same runtime endpoint. Registry failure copy is neutral
because setItem can have succeeded before readback failed.

Cold/error pages with no new entity edits do not require a ready entity context
to leave. Existing unreadable files stay untouched; any already-dirty memory is
still flushed under its original namespace, and failures stop destructive reload.

### Protected pre-auth and pairing

`connectionDrafts.ts` keeps one profile-owned DPAPI document in the reserved
`{endpoint: 'http://localhost', databaseId: null}` storage namespace. That endpoint
is a storage discriminator, never a network fallback. Add slots include the actual
runtime endpoint; Edit/Pair-again slots include the saved registration ID. This
does not switch or overwrite the active endpoint/database draft document.

Raw connection fields/codes, code-to-URL association and uncertain pairing markers
are protected/debounced. Public storage, metadata, duplicates and protected
persistence are checked before redeeming a code. A known failure before actual
pair dispatch releases its attempted-code marker; uncertain dispatch is not
replayed. Candidate pairing leaves active auth generation/credential untouched.
Use saved access can verify the already-saved native credential after verification
or registration failure, without consuming another code. Active repair occurs
only after workspace quiescence and retains keyring deletion/generation ordering.
Confirmed completion/discard clears the corresponding form, not other drafts.

Removal cleans only that endpoint's local credential and registration. Active
removal flushes/detaches and reloads to explicit unselected state; another entry is
not silently chosen. Drafts/server data/work remain. If credential deletion
succeeds and the later registry write fails, re-pairing may be necessary; no local
or remote rollback is asserted.

### Registry recovery and evidence

Invalid/unreadable registry is never overwritten by a fallback. Recovery permits
validated pasted JSON or an explicit empty list only after the exact original is
preserved in the protected pre-auth document. Explicit recovery originals are
deduplicated there (not an automatic draft-history stream), with a local Copy
preserved original action. The recovery textarea itself is a protected draft.

`product.md`, `clients/desktop/README.md` and the coverage matrix are updated.
`MULTI_ENGINE_DEV_VALIDATION.md` gives an isolated dev-tag/data/skills/native-app-ID/
WebView2-profile/CDP procedure for the independent tester. No such processes were
started. Final compilation/diff/commit/publication results belong to the CP3 report;
there is no native pairing/DPAPI/UI/Playwright runtime acceptance in this session.

CP3 final lightweight check: `npx tsc --noEmit` passed; `git diff --check` passed
with only the repository's normal LF/CRLF notices. Native Rust source is unchanged
since its successful CP2a `cargo check`, so that check was not repeated. No app,
engine, Playwright, full suite, installer or additional review was run.

Before publication, remote heads still matched the approved stack:
`fix/harness-large-events` = `952ff2bffae89d9b646ea9202b0187b65c0b5069`,
`release/0.6.0` = `89d0cb45ac0b049826db8f664dce5393a4b899d3`.
The coordinator resolved the misleading unscoped `gh` authentication result at
the user's request: the harness projects `XDG_CONFIG_HOME`, while the existing
user login belongs to `C:\Users\Samuel\AppData\Roaming\GitHub CLI`. Commands use
temporary `GH_CONFIG_DIR` only around `gh`, restored in `finally`; global/harness
XDG and Git config remain unchanged. No hosts.yml/token/keyring content is read or
copied. The coordinator confirmed account `samuelhk0595` and the same sole PR #39.
Feature commit/DRAFT PR references belong to the checkpoint report. No merge,
automatic rebase, deployment or installer is authorized by publication.
