# SQLite and on-demand loading — 0.6.0 implementation plan

## 1. Status, scope, and execution boundary

**Status: planning only.** Repository inspection performed on 2026-10-27 at HEAD
`64d45ef`. This document is the implementation handoff, not evidence that migration,
compilation, runtime behavior, or browser acceptance has passed. No application,
test, build, installer, or migration was run while writing this plan.

Read before implementation: `AGENTS.md`, `product.md`, this plan, and the current
progress documents for any domain being changed. Memory was queried before
architecture work (`SQLite OR persistence OR journal OR "on-demand"`); it returned
no matching pages. Repository code is the evidence for the findings below.

The user's requested outcome is **KLM release 0.6.0 using SQLite**, not SQLite
library version 0.6.0. The engine remains the persistence/execution owner; harnesses
remain responsible for executing coding work and their native continuations.

### Required outcomes

1. Durable, queryable SQLite records replace application checkpoint/journal writes.
2. Initial UI load retrieves projects, harness availability, and lightweight activity.
3. Selecting a project loads that project's topics/session summaries on demand.
4. Opening a conversation loads its current operational detail and a bounded latest
   history page; earlier content loads only on demand. Side, general, native-child,
   floating, mobile-sheet, and graph-node views follow the same principle.
5. Active turns, grants, requests, queues, graph scheduling, automations, webhook
   release, and recovery work with **zero UI clients**. UI subscriptions never decide
   what work exists or is eligible to execute.
6. Existing installations upgrade without losing supported domain records, history,
   native identity, receipts, credentials, or referenced filesystem artifacts.
7. Preserve the **exact original `state.json` bytes**, including whitespace and
   encoding, throughout 0.6.0 after success. Preserve the exact original journals
   too, because the checkpoint alone is not the latest logical state. Deletion of
   that backup belongs only to an explicitly designed future update.
8. Crash/retry migration, legacy journal replay, downgrade fencing, SQLite authority,
   and bounded resource use have concrete acceptance evidence.
9. Implementation includes an actual isolated dev-engine run and Playwright checks,
   followed by **two independent review/fix rounds** requested by the user.

### Resolved user clarification

The user clarified **“Validate before real migration”**: proceed with implementation,
isolated development validation, and the two coordinator-owned independent review
rounds. Leave the installed engine and all real release/development user data
untouched. Only fresh, absolute scratch `KLM_DEV_DATA_DIR` and `KLM_USER_SKILLS_DIR`
paths may be used for runtime/migration probes. No installation or real migration
is authorized by this implementation request.

### Worktree ownership

The brief warned of many preexisting modifications. The actual inspection returned:

```text
 M CLIENT_AUTH_IMPLEMENTATION_PROGRESS.md
?? PI_SUBAGENTS_IMPLEMENTATION_PLAN.md
?? PI_SUBAGENTS_IMPLEMENTATION_PROGRESS.md
```

Treat these as user/coordinator work and preserve them. Many recent features are
already present in tracked source, including Pi native subagents and client auth;
their persistence must be included even when not shown by `git status`. Recheck
status/diff at implementation start and before each handoff; coordinate overlapping
edits rather than reverting or resetting. No commits or pushes are authorized.

## 2. Repository findings that constrain the design

| Current code | Finding and required response |
| --- | --- |
| `engine/store.go:96–147` | `diskState` contains all domains, including every session's `Events`. Treat this as the legacy serialization inventory, not the new live store. |
| `engine/store.go:190–335` | `loadState` reads the whole file, validates global slices, normalizes records, and can **write a v1→v2 checkpoint**. Never run it against the original migration source. |
| `engine/transactions.go:95–101,340–440` | Transactions clone/diff the global object, scan all sessions, validate graph/topic state, append a journal frame, then publish. A SQL implementation of `saveState` alone would retain global eager loading and write amplification. |
| `engine/stream.go:124–179`, `engine/usage.go` | Streaming and usage have separate journal-writing paths in addition to `commitLocked`. All must converge on SQL transactions. Keep 40 ms stream batching and commit-before-publication. |
| `engine/stream_journal.go:68–192` | Replay is sealed segment first, active segment second; CRC-framed records are revision ordered. The reader opens files **read/write and truncates incomplete tails**. Migration requires a read-only variant. |
| `engine/transactions.go:312–337` | Legacy transaction replay preserves histories by session ID across slice reorder, applies non-history patches first, then event patches. Positional patches cannot be naively translated into ID-based UPSERTs. |
| `engine/abandoned_tasks.go:110–185` | Startup cleanup precedes loading and can rewrite checkpoint/remove journals. Capture originals before this code or replace it with nonmutating import transforms. Do not resurrect retired task authorization. |
| `engine/main.go:97–255` | Startup loads/replays globally, rolls forward filesystem outboxes, interrupts work, writes state, migrates topics, indexes all event IDs/webhooks, and validates all graph snapshots. Replace this pipeline, not merely its first read. |
| `engine/history.go:17–104` | Existing `SessionSummary` includes queue text, capability snapshots, requests, and graph projection; `lastMessageAt` scans history. It is too heavy for global bootstrap. |
| `engine/history.go:196–252` | Nominal 50-event pagination expands backward for work/consultation groups, potentially to the beginning of a huge conversation. Cursor resolution also scans events. Enforce hard page bounds and support split groups. |
| `engine/history.go:279–290`, `engine/graph_activity_api.go` | Export copies a whole history; node activity copies/merges whole activation/native ranges. Convert to streaming export and bounded activity pages. |
| `engine/api.go:205–229`, `engine/updates.go:78–149` | Both `/api/state` and SSE inventory enumerate all visible session summaries; SSE keeps per-client global inventory and can regenerate it after a commit. Remove global inventory dependence. |
| `engine/message_identity.go` | Startup builds message/event indexes for all histories; duplicate acceptance checks also scan sessions. Replace with indexed existence/receipt queries and transaction constraints. |
| `engine/general_agent.go`, `engine/linked.go`, `engine/general_agent_tools.go:678+`, `engine/graph_tools.go`, `engine/subagents.go` | Lock inference, source passages, authorization, read/search, and parent-child synchronization inspect event slices. Replace with exact ID/range queries, not runtime hydration of entire sessions. |
| `engine/analytics_records.go`, `engine/analytics_transactions.go`, `engine/analytics.go` | Accounting has nested growing span/usage histories, global validation and report loops. Store observations as rows, query/aggregate on demand, and load only open executions for recovery. |
| `engine/webhooks.go`, `engine/webhooks_worker.go`, `engine/github_activity*.go` | Raw retained webhook payloads and derived grouping currently live in global maps/projections. Pagination after rebuilding all payloads is not sufficient. |
| `engine/graph_snapshot.go:454+`, `engine/library_api.go:416+` | Filesystem roll-forward recovery writes `state.json` directly. Migrate outbox ownership to SQL while keeping filesystem operations recoverable. |
| `engine/client_auth.go`, `engine/engine_access_settings.go` | Separate `client-auth.json` and `engine-access.json` persist pairing/last-use and restart-applied access policy. Inventory and preserve these too; do not silently re-pair users or revert saved empty allowlists to environment defaults. |
| `clients/desktop/src/App.tsx:391–422,485–533` | Polls `/api/state` every 30 seconds; multiplex stream rebuilds global inventory. Subagent event discovery can subscribe to unopened child histories. Rewrite acquisition/subscription scope. |
| `clients/desktop/src/features/chat/sessionState.ts` | Summaries and histories share `Session[]`; previously opened histories accumulate without eviction. Preserve race handling but separate caches and bound them. |
| `HistoryLoader.tsx`, `ConversationHistory.tsx`, `SideChatPanel.tsx` | Earlier pages prepend into one array; intersection observers can continue fetching while mounted. Make loading visibility-aware and use a bounded page window. |
| `App.tsx`, `FocusWorkspace.tsx`, `focusLayout.ts`, `SessionNavigation.tsx` | Navigation/restoration assumes absence from global inventory means unavailable. Sparse loading requires exact-ID lookup and explicit availability, not destructive reconciliation against an incomplete list. |
| `engine/cli.go:12–25,93–106` | Dev data override already exists; startup readiness timeout is only 45 seconds. A large first migration can exceed it and needs explicit migration progress/readiness handling. |
| `engine/platform_windows.go:18–59` | Existing same-volume replace uses `MoveFileExW(REPLACE_EXISTING\|WRITE_THROUGH)` and an exclusive data-directory lock. Reuse these guarantees, test interruptions at the publication boundaries. |

## 3. Target storage and runtime architecture

### 3.1 SQLite ownership and connection policy

Add a narrowly scoped storage layer under `engine` (for example
`sqlite_store.go`, `sqlite_schema.go`, `sqlite_migration.go`, `sqlite_legacy.go`,
and domain-specific repository files). Keep the current package layout unless a
package split is independently necessary; this is not an unrelated modularization.

* Database path: `<engine-data-dir>/engine.sqlite3`.
* Prefer `database/sql` with a pinned **pure-Go SQLite driver**, provisionally
  `modernc.org/sqlite`, to preserve the current Windows build without a new C
  toolchain or separately shipped DLL. Confirm a maintained version compatible
  with `engine/go.mod`'s Go 1.24.2 and the actual build toolchain before pinning;
  record that choice in the implementation progress. Do not silently bump Go.
* One serialized write connection/transaction gate; a small explicitly capped
  reader pool (initial target four connections). Configure foreign keys, busy
  timeout, and cache limits on **every** connection, not only the first opened one.
* Runtime WAL mode, `synchronous=FULL`, `foreign_keys=ON`, bounded busy timeout,
  disk-backed temporary work for large sorts, and explicit total cache budget.
  Use short read transactions and bounded pages so readers cannot indefinitely
  pin WAL growth. Checkpoint based on WAL size/time; report checkpoint failures.
* Schema version (`PRAGMA user_version`) is independent of product version and
  legacy `diskState.Version`. Verify an application identifier, schema version,
  and migration metadata before serving requests. Reject newer schemas.
* SQL transactions replace both transaction frames and streaming journal frames.
  No JSON dual-writing after cutover, no background regeneration of `state.json`,
  and no fallback to legacy data when SQL open/query/recovery fails.
* Open SQL only after the existing data-directory lock and fixed port reservation.
  Close/checkpoint it on orderly shutdown after writers/workers finish. A crash
  relies on SQLite recovery, not an application-written duplicate WAL.

### 3.2 Persist rows, not a serialized object graph

The following is the required mapping inventory. Table names are proposed; record
identities and domain semantics are existing contracts. Each growing collection is
its own rows. Small per-record structured payloads may remain JSON where they are
not list/filter/recovery keys. **No database-wide, project-wide, conversation-wide,
or domain-history JSON blob.** A whole session with `events` is not a row payload.

| Legacy domain | Proposed rows / important keys |
| --- | --- |
| Version, topic schema, journal/graph revisions, engine settings | `store_meta`, `engine_settings`, `migration_runs`; retain imported revisions and explicit initialized/schema flags. |
| Projects, topics, folders/archive migration, order | `projects`, `topics`, `project_topic_order`; explicit project display order; persist topic grouping sort key based on newest-created member including archived members. Preserve removed projects. |
| Session metadata and hierarchy | `sessions` keyed by ID, nullable project for engine-owned general/children, parent, role, topic, graph IDs, timestamps, settings, durable harness lock, event count, last message timestamp, per-session revision. |
| Session events, text/data, sources, usage | `session_events(session_id, ordinal, event_id, type, status, consultation_id, created_at, revision, ...)`, separately fetched payload columns/table, `session_sources`, current `session_usage`; indexes on `(session_id,ordinal)` and `(session_id,event_id)`. Preserve original order/IDs and all event data/favorites. |
| Queue, prepared submissions, native delivery/acceptance | `queued_messages`, `queue_payloads`, `accepted_messages`; preserve order, client IDs/fingerprints, mode, status, prepared references and uncertainty. |
| Message identity reservations | Indexed identity lookup/reservations matching existing global collision checks, plus sender/session-scoped receipts. Do not impose a new global uniqueness rule on legacy native event IDs without checking existing data semantics; retain composite event keys and query collisions as today. |
| Native continuation | `native_sessions` stores exact IDs/paths; native JSONL and external harness histories remain filesystem/harness owned. |
| Pending permissions/questions and saved rules | `pending_permissions`, `pending_questions`, `permission_grants`; preserve owner/native IDs, scope and offered decisions. In-process waiters are not serialized. |
| Spawn/send/control receipts and consultations | `session_spawns`, `session_commands`, `consultations`, indexed by sender/operation and delivery/status; preserve paired visible events and real human authorization references. |
| Graph state | `graph_activities`, dependencies, assessments, operation receipts, `graph_runs`, per-run snapshots, `graph_activations`, activation events, deliveries, join rounds/arrivals, workspaces/uses, notifications. Separate append-growing nested collections and payloads from small scheduler metadata. |
| Filesystem authoring recovery | `graph_catalog_changes` and ordered writes; `library_deletions` and per-file intent/receipts. Keep content hashes/preconditions and exact paths. |
| MCP and Library | Connections, protected credential BLOBs, desired selection, applied/preparing snapshots and captured graph capabilities; per-selection rows where queried. Immutable skill content remains in `capabilities/skills`; user/internal skill directories remain sources of truth. |
| Webhooks | Connections/secrets, immutable event payloads, pending rows, releases, quota counters and both independent sequences; indexes for delivery dedupe, expiry, project/release order and pending work. |
| GitHub activity | Derived compact event/group/member keys sufficient for indexed cards before pagination; raw event stays authoritative. Rebuild a derived index in bounded batches, never by loading all raw payloads at startup or on each feed read. |
| Tasks board | Config/phases/mappings, connections, encrypted credentials, cached items, sync status and move receipts; keep revision/uncertain-provider semantics and removed-project connections. |
| Automations | Definitions, triggers/FIFO, runs, snapshots, recoveries, missed intervals, per-occurrence decisions and operation receipts; index pending/due/destination ownership. Normalize growing recovery decision/receipt collections. |
| Analytics | Executions, spans, usage deltas, counter baselines, graph timings and waits/reasons; append observations separately, query aggregates by project/time/owner. Preserve unknown/partial measurements rather than invent backfill. |
| Transcription | Settings, write-only key and vocabulary entries; preserve existing secret behavior. |
| Separate auth/access files | `paired_clients` and `engine_access_settings`, including a saved-config-present flag. Import hashes/last-use exactly; preserve Windows credential protection on other domains and active-vs-saved restart semantics. |

Before freezing schema, make a checked inventory against **every JSON field in
`diskState`, `Session`, graph/automation/accounting nested types**, and all independent
data-dir file readers/writers. Each field needs a destination, a filesystem-owned
exception, or a documented legacy-only transformation. Detect unmapped fields as
an error; do not drop them through a permissive decoder.

Other filesystem-owned data remains intact: prompts/internal skills, `.klm` graph,
agent and layout authoring, worktrees/workspace artifacts, native sessions, protected
external harness credentials, local browser drafts/layout and Flutter saved hosts.
SQLite holds references/outbox intents where appropriate. It must not import user
project files or replace native harness storage. Existing retired Tasks fields stay
preserved in immutable legacy backups without becoming executable records or grants.

Indexes must match actual access paths: active/unfinished status, project+role+
archive+created order, topic membership, parent/child role, event ID/ordinal,
consultation/request ownership, sender operation IDs, pending notifications,
automation destination and due time, webhook expiry/dedupe/release, analytics time.
Use foreign keys and transaction constraints for integrity; avoid cascading away
archived/removed history. Validate ownership rules that SQL foreign keys alone do
not capture (general-agent exclusivity, native-child role, graph scope, topic project).

### 3.3 Transaction API: explicit read/write sets

Replace `commitLocked(func(*diskState))` callbacks with repository transactions that
load only records needed for the operation and stage explicit inserts/updates.
Keep `app.mu` initially for runtime handle coordination and command serialization;
do not introduce concurrent writers while changing persistence semantics.

Suggested operations include `GetSessionMeta`, `GetEvent`, `ReadEventRange`,
`AppendEvent`, `UpdateEvent`, `AcceptMessage`, `GetPendingQueue`, `PutReceipt`,
`GetActiveGraph`, and domain-specific graph/automation transitions. A transaction
returns a compact change set (changed project/session IDs, ordinals, revisions,
request/status changes) for publication **after COMMIT**. It must not build a full
`diskState` and diff it, even temporarily in production runtime.

Important atomic units to preserve:

* User acceptance + durable harness lock + user event or queued message + prepared
  payload + accepted fingerprint; sender receipt + spawned session + creation card.
* Consultation request/status + both participants' visible projections.
* Stream event upserts + owning summary/count/revision; usage + accounting delta
  and native counter baseline in the same SQL commit.
* Graph authorization/admission + activity/run/reservation and notification;
  Choice sealing/settlement + deliveries/join arrivals under existing rules.
* Permission grant/reply intent/provenance and native reply recovery; SQL rollback
  must not leave an in-memory grant effective after a failed commit.
* Webhook acceptance + pending row + dedupe/quota/sequence changes; release +
  sequence; expiry/deletion removes all dependent visibility atomically.
* Automation trigger admission + FIFO/idempotency; reused-destination claim;
  recovery decision + its receipt + exact selected runs.
* Board move intent/reconciliation and authoring/Library outbox intents.

No native process, tool call, HTTP/provider call, file authoring or model dispatch
runs inside the SQL transaction. Commit intent first, perform external effects,
then settle; uncertain effects retain existing explicit recovery behavior. Add
lock-order documentation for `app.mu`, writer gate, auth mutex, authoring/library
locks; auth transactions must not call back into `app.mu` while holding the writer.

On storage failure retain `failStorageLocked`'s fail-closed/read-only cancellation
semantics. Domain rejection is a rollback, not a permanent storage error. No SSE
publication or successful acceptance response before durable commit. Unknown commit
outcomes must be reconciled through operation IDs, not blind effect retries.

### 3.4 Active state without global history

`app` keeps live process/turn handles (`runs`, `runtimes`, `graphRuns`, waiters,
gateway ownership) and bounded active-record caches. Cold metadata/history resides
only in SQL. Even an active session must not hydrate its entire past conversation.

* Load exact native continuation/settings and current pending work on admission.
* Resolve authorization/source passages by session+event ID, and nearby context by
  bounded ordinal range. Do not let cache eviction change authorization outcomes.
* Replace `len(Events)` dependencies with persisted event counts/ordinal boundaries;
  use exact current-turn IDs/ranges, including graph `NativeEventStart/End`.
* Track current reasoning title from current-turn changes; clear on turn end/start.
* Replace `messageIDs`/`streamIndexes` lifetime-global maps with SQL lookups and a
  bounded current-turn event-ID cache. Cache a currently edited event, not all IDs.
* Query unfinished graph/automation/consultation/queue/outbox work by indexed status,
  in batches. Scheduled but inactive work stays in SQL until due/admitted.
* Active graph snapshot compilation is demand-driven and evictable after execution;
  do not run `validateGraphRecordsCached` over every historical snapshot at startup.
* General-agent/collaboration tool reads retain current scope and bounded output,
  but also bound records read. Literal search may scan scoped SQL text on demand;
  do not fetch all results to Go before applying a limit. Support cancellation and
  an indexed continuation scan budget for expensive substring search.
* Native OpenCode/Pi/Codex reconciliation is an explicit per-running-session adapter
  operation. Preserve existing native paging/input caps (`opencode_history_projection.go`)
  and incrementally upsert its output; do not attach a whole native transcript to
  the engine's session cache. Audit adapter `Events` loops as part of this conversion.

## 4. Migration, backup, fencing, and recovery protocol

### 4.1 Files and durable markers

Proposed layout (all below the chosen engine data directory):

```text
engine.sqlite3                         # authoritative only after committed cutover
engine.sqlite3-wal / engine.sqlite3-shm # SQLite-owned runtime files
state.json                             # downgrade fence after publication, NOT backup
storage-cutover.json                    # small versioned cutover/recovery manifest
legacy-backups/<migration-id>/
  state.json                           # exact original checkpoint bytes
  stream.sealed                        # exact original bytes, if present
  stream.journal                       # exact original bytes, if present
  client-auth.json                      # exact original, if present
  engine-access.json                    # exact original, if present
  manifest.json                        # hashes, sizes, presence, source/schema versions
.sqlite-migration-<id>/                 # owned, unpublished staging artifacts only
```

The backup's file named `state.json` is the **original**, not a reserialized/replayed
checkpoint. Explain this path distinction in the release documentation. No 0.6.0
startup, cleanup, retention, uninstall, successful-retry, or checkpoint path may
delete/replace that backup. Do not set a cleanup timer. Protect the backup directory
with the same per-user data ACLs; it contains secrets. Hashes/diagnostics never log
payloads or credentials. Keep earlier unrelated backups untouched.

### 4.2 State machine and publication order

Use the existing exclusive data-directory lock for the entire migration. Do not
start workers or serve application/control readiness until cutover and recovery
are complete. A separate non-ready migration status may be exposed only through
the protected local control path.

1. **Detect before legacy mutation.** Inspect DB/manifest/fence state before
   `removeAbandonedTasks`, `loadState`, replay, auth writes, or outbox recovery.
   Existing committed SQLite always wins. An unsupported DB schema or broken
   authoritative DB fails closed; it never triggers fresh initialization/import.
2. **Capture immutable source.** Stream-copy originals to a new backup directory
   using exclusive create, flush files, and atomically publish its manifest only
   after lengths/hashes verify. Record absent files explicitly. Keep root originals
   untouched in this phase. An incomplete backup is not a reusable source; preserve
   it for diagnosis and create/recover only clearly migration-owned artifacts.
3. **Import into staging.** Build schema and rows in a same-volume staging database;
   replay preserved journals using read-only readers. Apply legacy compatibility
   transforms in staging. Store source identity, checkpoint/final replay revision,
   counts/digests and `prepared` migration status. Batched import commits are allowed
   only while staging is unreachable by runtime. Retry may rebuild staging from the
   immutable source; no partially imported DB is served.
4. **Validate prepared state.** Referential/semantic checks, per-domain counts and
   normalized digests, `foreign_key_check` and SQLite integrity check. Compare
   logical replay result before restart interruption separately from the expected
   post-recovery state. Verify preserved source hashes again. No filesystem outbox
   effects are executed merely to validate an import.
5. **Publish prepared DB safely.** Use a closed/checkpointed single-file staging DB
   (e.g. DELETE journal mode during construction). Never rename a DB while its
   committed data still lives only in a temporary WAL. Flush/close then publish
   with same-volume replacement, without overwriting any committed authoritative DB.
6. **Fence old engines.** Publish a durable `storage-cutover.json` containing DB ID,
   backup ID and source hashes, then atomically replace the *live* `state.json`
   with a small sentinel the supported old JSON readers reject (unsupported version
   and explicit unknown storage-format field). Keep the sentinel free of abandoned
   Tasks keys so old pre-load cleanup cannot transform it. The exact backup is
   untouched. Use a fence even on a fresh install so an older engine cannot see a
   missing checkpoint and initialize an empty competing store.
7. **Commit authority.** After the fence is durable, open the published DB, verify
   its identity/source manifest, and commit `cutover=committed` in SQL. The prepared
   DB is not eligible for normal writes before this step. Atomic manifest updates
   mirror status for diagnostics; SQL's committed marker governs authority.
8. **Recover using SQL; then become ready.** Enable runtime WAL settings, perform
   idempotent startup recovery and required filesystem outbox roll-forward, start
   workers, and only then report ready. No post-cutover write touches the original
   JSON/journals or their backup. Root legacy journals may remain inert in 0.6.0;
   no new code replays/appends them after the committed marker.

Do not rely on a new sidecar marker alone to fence old binaries: they do not know
about it. Do not rename away `state.json` without replacing it with a rejecting
sentinel: old `loadState` treats missing state as a fresh install.

### 4.3 Crash/retry decision table

| Durable state on restart | Required behavior |
| --- | --- |
| Legacy sources, no complete backup or staging DB | Start capture; no loss or mutation of originals. |
| Complete backup and incomplete staging, no fence | Revalidate current root source hashes. Rebuild/resume only matching source. If an old engine wrote meanwhile, capture a new backup generation/import; preserve earlier backups. |
| Prepared published DB, live original still present | No authority yet. Verify source fingerprints and manifest; rebuild if changed, otherwise finish fence/cutover. Never serve prepared rows. |
| Fence installed, DB prepared | Old engine is blocked. Finish verified cutover from the prepared DB; if unavailable/corrupt, rebuild only this pre-authority generation from its verified preserved source. |
| SQL committed, manifest status lagging | SQL wins; repair marker/fence identity before readiness, then recover SQL. No legacy replay. |
| SQL committed, old files edited or extra journals present | Do not merge/reimport them. Keep/report conflict and preserve evidence; SQL remains authoritative. |
| Cutover manifest/fence exists but DB missing/corrupt and authority cannot be proven absent | Fail closed with explicit recovery instructions. Never silently restore the old checkpoint. |
| Newer/unrecognized schema or mismatched DB/manifest ID | Refuse startup without mutation. |

Re-read and hash the live sources under lock before fencing. A migration interrupted
before the fence may have been followed by an old-engine run; a stale backup is not
proof of the latest state. Record phases durably enough that recovery can distinguish
pre-authority rebuild from destructive post-authority rollback. Test ordering at
each flush/rename/commit boundary on Windows, including denied rename/disk-full.

### 4.4 Legacy replay correctness and memory feasibility

Extract a **read-only** framed reader from `stream_journal.go`; do not call its
`O_RDWR`/`Truncate` reader on backup or source. Report the last complete offset and
ignored incomplete tail without altering bytes. Preserve these exact rules:

* Checkpoint `GraphRevision` is the replay lower bound. Read `stream.sealed`, then
  `stream.journal`; skip already-checkpointed revisions, require each applied
  revision to be exactly the next one. Validate complete frames even when skipped
  according to current compatibility behavior.
* Format 0: append-text versus replace-text, merged event data, title/status,
  usage and optional analytics mutation, session timestamp and revision semantics.
* Format 1: whole transaction, nested paths, map delete, null/replacement versus
  merge, slice resize, new records and reordered sessions. Preserve histories by
  **session identity** before applying event patches at their new positions.
* Partial final header/body is an uncommitted tail; complete bad CRC, bad JSON,
  invalid size/format/path, missing session or revision gap is fatal. No guessed
  repair, partial frame application, or skipping of complete corrupt records.
* Match historical distinction between `GraphRevision` and `GraphViewRevision`.
  Preserve webhook and other independent sequences; do not renumber durable IDs.
* Apply retired Tasks compatibility filtering to checkpoint **and journal patches**
  without erasing unrelated mutations. Handle v1/v2, topic migration, old child-role
  defaults, general harness lock inference, Library preparation clearing, and graph
  workspace normalization in explicit ordered transforms with evidence.

Migration must not be implemented as “load all rows into `diskState`, save to SQL,
then keep that object.” For large legacy files use a streaming decoder and
temporary SQL staging positions:

1. Decode top-level fields and nested growing arrays/maps record by record, including
   session events, graph activation events, analytics spans/deltas and webhook data.
   Do not `Decode(&Session)` with its entire `Events` array or decode the complete
   analytics object as one allocation. A bounded single record/legacy frame is the
   memory unit; the existing maximum journal frame is 256 MiB and must be accounted
   for as an explicit import-only upper bound, not normal runtime memory.
2. Keep legacy collection positions in staging tables so positional patches can be
   applied in order. Implement the known patch grammar using typed per-record
   compatibility values; spill large collection replacements to staging rows.
   For a session reorder, retain the pre-frame ID→history association, apply all
   metadata patches, rebind histories by the resulting IDs, then apply event patches.
3. Replay each frame atomically in staging and advance its applied revision in that
   transaction. Rebuild staging on ambiguous import state rather than trying to
   salvage half a frame. Drop staging-only positional tables before authority.
4. Port whole-state validation into local validation plus streaming/SQL relational
   checks. Existing full legacy replay can be a **small-fixture test oracle** on
   disposable copies, not the production startup/replay implementation.

This compatibility layer is one of the largest pieces of work. Prove it on legacy
reorder/map/nested-patch fixtures before migrating all runtime writers. If an
unsupported legacy shape is discovered, stop with the originals intact and add
support; never “simplify” by losing records or silently disabling their domains.

### 4.5 Startup recovery after cutover

Move `main.go` recovery into indexed, idempotent SQL operations, with a durable
startup/recovery generation if multiple batches are required. Do not expose a
partially recovered engine to commands. Repeating after a crash must not duplicate
interruption events, notifications, missed intervals, or recovery receipts.

* Running chats become interrupted/error with the correct attributed event. Native
  children update their parent invocation by exact event ID. Clear dead in-process
  permission/question waiters using current restart semantics; preserve historical
  request events and saved grants. Do not pretend a restarted native tool is waiting.
* Queue recovery uses existing pause/uncertain distinctions. Never auto-redeliver
  `sending`/uncertain prompts or reset accepted-message fingerprints.
* Consultations preserve correlation and become interrupted as currently specified;
  graph running/starting/ending states, active activations, pending deliveries,
  joins/reserved workspaces settle per `interruptGraphState`. Retain scheduled
  activities and notification IDs; preserve the existing notification delivery policy.
* Accounting closes only open executions/timings, keeping unmeasured tails unknown.
  Automation running work fails with durable recovery, missed schedules retain saved
  calendar/time zone/occurrence IDs, and delivering recovery notices become uncertain.
* Webhook pending release, TTL/quota and board startup reconciliation run headlessly.
  Expired events are invisible immediately even if physical cleanup uses batches.
* Graph catalog/Library outboxes execute their existing idempotent file actions after
  authority, then settle in SQL. A failure remains recoverable and does not discard
  the outbox or rewrite the backup. Synthetic tests must use scratch project/skill
  paths because recovery can touch paths outside the data directory.

Extend `klm start`/installer readiness handling for a long migration. Provide bounded
status/progress and a useful error path; the existing 45-second timer must not label
a healthy migration a failed engine while an orphan worker continues changing data.
Any cancellation is cooperative at safe boundaries. Preserve the private local
control channel and do not add a LAN stop/migration-admin endpoint.

## 5. On-demand read API and notification contract

Use an explicit API storage/read-contract version so an independently installed old
desktop does not interpret a partial inventory as “all sessions deleted.” The new
frontend checks compatibility before starting normal polling/streams and offers an
English update/recovery error if incompatible. An old unversioned read contract may
receive a clear upgrade-required response; never preserve it with a full-history or
global-session inventory fallback. Document independent engine/desktop update order.

### Suggested endpoints/projections

| Surface | Contract and query behavior |
| --- | --- |
| Bootstrap (`GET /api/state` revised/versioned) | Lightweight project rows, harness availability, protocol/version, global revision, general-agent identity/lock metadata, per-project activity aggregate. No all-session IDs, topics for every project, queue bodies, graph snapshots, capability content, or history. |
| Lightweight activity (`GET /api/activity`) | Compact Happening now rows, default 50/max 100 with cursor, plus project aggregates. Preserve current eligibility and Failed → Needs attention → Working → Idle precedence. Ordinary eligible top-level sessions only; exclude removed/archived/topic-archived/general/side/graph-node. |
| Project selection | `GET /api/projects/{id}/topics` and `/sessions` with bounded keyset pages (default 100/max 200), archive/search/topic scopes and explicit completeness/cursor. No reads for other projects. Empty topics and topic order remain available. |
| Exact ID lookup | Small metadata batch endpoint (e.g. max 64 IDs), and/or `GET /api/sessions/{id}` returning detail without automatic all-history. Used for restored Focus windows, spawn links, overlays, automation history and native-child navigation. Missing/inaccessible is explicit, distinct from not-yet-loaded. |
| Open conversation | Load operational detail/settings, pending queue/permissions/questions/usage and latest bounded `history` page; native child summaries separately scoped to parent. General chat hydration only when its chat surface is opened/restored. |
| History | Preserve stable absolute ordinals and event identity, implement indexed keyset cursor with default/max 50 events and a byte budget. Return revision/count/hasMore and both-direction page boundaries if cache eviction needs navigation forward. |
| Graph | Lightweight selection/progress separate from captured definition on Graph view; node activity and activation list paginated only while its panel is open. Preserve run/node/activation ownership and ordering for merged native/activation events. |
| Export | Explicit user action streams a consistent SQL snapshot in bounded chunks; cancellation closes the reader. Avoid constructing a whole `Session`/JSON buffer or adding it to browser chat cache. |
| Agent tools | Exact-message/context lookup, bounded read/search and delta-fragment contracts backed by scoped SQL; preserve private role capabilities and authorization provenance. |
| Feed/board/automation/analytics | Existing feature-on-demand APIs query normalized rows/derived indexes. They must not recreate a global eager domain cache during engine startup or each request. |

Page ordering and completeness need explicit design, not arbitrary slicing of the
current sidebar. Keep current newest-created session order and stable ID ties;
topic order uses persisted manual order or the newest-created member including
archived members. Compute group counts/order in SQL. Topic-aware paged queries and
server-side search must find matches beyond loaded pages without loading all sessions.
Changing topic/membership must invalidate affected page cursors/revisions and refresh
the loaded window, not duplicate/omit rows. Recency remains independent of ordering.

History grouping must tolerate a page cutting a work or consultation group. Emit
group IDs/continuation metadata, then merge adjacent segments in
`ConversationEvents.tsx` without walking backward to fetch the entire group.
For a single oversized event, return a bounded preview plus an explicit versioned
content continuation/download path. Preserve full text/data in SQL and export;
never permanently truncate migrated content. Separate payload reads from event
metadata so a page-byte limit does not first allocate every giant payload. Define
text offsets/JSON-fragment encoding and revision checks; no split invalid UTF-8 or
malformed partially decoded JSON. Full event retrieval must itself be chunked.

### SSE: scopes rather than global inventory

Refactor `updates.go`, `history.go`, `notifyAllLocked`, and frontend
`updatesConnection.ts` together:

* Global scope carries project/activity aggregates and compact activity changes.
* Project scopes carry summary upserts/removals/invalidation only for selected or
  explicitly retained project lists. Invalidation may mark a cached project stale
  without loading it. Do not enumerate every session ID to detect inventory changes.
* Conversation scopes carry operational detail and event deltas only for opened
  history consumers. Separate metadata-only child subscriptions from history scope.
* Use commit change sets and indexed impacted-owner queries. Publish topic/project
  archive/removal to all relevant scopes, including overlaid/floating conversations.
* Register subscriptions before snapshot capture, use one consistent read revision,
  and deduplicate races. Preserve per-conversation cursors; a global current cursor
  cannot skip a newly opened chat's first page.
* Keep delta retention bounded by bytes/count and **total engine memory**, with
  eviction when no consumer remains. Expired cursor/process restart triggers a
  bounded scoped reset, not all-engine rehydration. Metadata-only transactions need
  revisions too; `updatedAt` alone is not sufficient for ordering.
* Document revision representation: use decimal strings/opaque cursors in the new
  contract for uint64 values, or enforce an explicit safe-integer range. Do not
  silently round imported revision/sequence values in JavaScript or signed SQL.
* Slow consumers cannot block durable writes; bound pending output and disconnect
  with resumable cursors when necessary. Preserve client revocation/stream teardown.

## 6. Frontend loading, cache, and existing UI behavior

Primary files: `clients/desktop/src/App.tsx`, `engine.ts`,
`features/chat/sessionState.ts`, `updatesConnection.ts`, `HistoryLoader.tsx`,
`ConversationHistory.tsx`, `ConversationEvents.tsx`, `SideChatPanel.tsx`,
`features/workspace/{WorkspaceSidebar,FocusWorkspace,FocusSessionChat}.tsx`,
`features/workspace/focusLayout.ts`, `features/projects/ProjectRail.tsx`,
`features/chat/SessionNavigation.tsx`, `features/klm-agent/*`, and graph activity UI.

Introduce a small shared query/cache layer; avoid a speculative application-wide
framework rewrite. Separate:

1. Bootstrap/project/activity records.
2. Project-scoped paged topic/session lists and their completeness/version.
3. Session detail cache keyed by engine identity + session ID.
4. History page windows, per-event revisions and per-stream cursors.
5. Drafts and local navigation/layout, independent of history eviction.

Required behavior:

* Replace `/api/state` global polling and whole-inventory replacement. Initial Home
  loads no project list or chat history. A restored selected project/chat is an
  explicit selection and may load only that scope. Poll/focus fallback only refreshes
  the visible loaded scopes/window; it never broadens the inventory.
* Use abort/generation guards on project changes, history pages and reconnects.
  Keep the current protections against HTTP revision N preceding SSE N−1; mutation
  responses must not advance the contiguous SSE cursor past missing updates.
* Subscription ownership is shared/reference-counted across main chat, side panel,
  KLM Agent dock, overlay, Focus windows and mobile stacked sheets. Closing one
  consumer does not unsubscribe another or Stop the engine's work. Covered mobile
  sheets keep the documented mounted/subscribed behavior; distinguish them from
  unopened histories. A minimized already-open chat may retain its bounded cache.
* Pending request indicators, rail LEDs, Happening now and Stop use compact active
  state/detail endpoints, not scanning message arrays. Do not subscribe to native
  child history merely because a parent event says the child is running.
* Resolve saved window/session IDs explicitly before pruning layout. Do not confuse
  unloaded, loading, missing, forbidden/unavailable, removed and failed-to-fetch.
  Preserve saved geometry/drafts across cache eviction and temporary disconnection.
* Spawn cards, automation history and “Open in project” must navigate by exact ID
  even when that project's list has never loaded. Expand its correct topic after
  loading the metadata; no requirement to page through all preceding sessions.
* Use a bounded LRU for closed history/detail caches and a bounded visible page
  window for open sessions. Proposed defaults: 10 history pages/500 events per
  consumer window, 8 MiB of retained event payload per session and 64 MiB globally,
  with smaller chunk/detail budgets for oversized bodies. Metadata and drafts have
  separate policies. Tune only with measured evidence, not by removing limits.
* Evict pages away from the reading anchor; permit loading newer as well as older
  pages. Keep a compact live tail/notification while reading far back. Do not drop
  drafts or pending input intent; request/queue state is independent of event cache.
* History observers are active only for a visible reading surface and never recursively
  drain the whole transcript because a collapsed group leaves a sentinel visible.
  Keep scroll anchors stable on prepend, reset, group continuation and eviction.
* Reuse current components/tokens and English UI labels. Keep explicit loading/retry
  states; avoid presenting an unloaded conversation as an empty one.

## 7. Delivery phases and gates

This is a substantial persistence/runtime refactor. Deliver in dependency order;
do not release a halfway state that hides eager loading behind SQL. Suggested file
names below are implementation locations, not claims that they already exist.

### Phase 0 — reconcile constraints and freeze contracts

* Resolve the incomplete “Before applying” condition with the coordinator/user.
* Recheck worktree, ownership and current HEAD; record baseline in a new
  `SQLITE_0_6_0_IMPLEMENTATION_PROGRESS.md` without modifying other owners' evidence.
* Finish field/file/reader/writer inventory; explicitly track every `a.state`/`d.*`
  collection scan, `.Events` access, `saveState` and `appendStreamRecord` production call.
* Freeze SQL schema mapping, protocol compatibility/version, oversized-content
  contract, backup paths, migration state machine and performance budgets.
* Pin the driver after compatibility/build feasibility inspection. Identify fixture
  sources for every persisted domain and a matching pre-SQL downgrade reader/binary.

**Gate:** no unmapped current domain or unresolved implementation precondition.

### Phase 1 — storage primitives and read-only legacy import

* Add SQL open/config/schema/transaction infrastructure and explicit repository APIs.
* Implement immutable capture/manifests, streaming checkpoint import and read-only
  journal replay with positional staging; add typed/relational validation.
* Implement prepared/cutover metadata, sentinel publication and resume decision table.
* Add injected filesystem/SQL failure boundaries and small-fixture legacy replay oracle.

**Gate:** scoped migration tests prove exact backup bytes, logical equivalence and
safe retry before migration is wired into production startup.

### Phase 2 — session writes and indexed reads

* Convert `store.go`, `transactions.go`, `stream.go`, `usage.go`,
  `message_identity.go`, `message_queue.go`, normal session controls/commands/spawn,
  topics/project registration, permissions/questions and linked consultations.
* Replace implicit event slices with exact/range repository methods across adapters,
  general-agent tools and subagents. Preserve native mapping and acceptance receipts.
* Implement hard-bounded history/read/export paths and compact summary projections.

**Gate:** no session mutation persists by global clone/diff; cold history is not
resident and mutation failure never publishes partial state.

### Phase 3 — graphs, background domains, auth and recovery

* Convert graph records/scheduler/activation/notification and filesystem outboxes;
  retain private bridge permissions, sealing and native grant ownership.
* Convert Library/MCP, webhook ingestion/release/expiry/feed projection, Tasks board,
  automations/recoveries and analytics, including growing nested records.
* Import and switch auth/access settings writes to SQL with correct lock ordering.
  Preserve saved policy/environment bootstrap and client credential expiry semantics.
* Replace `main.go` startup/shutdown recovery with indexed SQL recovery and ensure
  all headless workers query SQL independently of client connections.
* Remove production global history/message indexes and checkpoint worker; retire
  JSON journal writes. Keep legacy format code only in the importer/test oracle.

**Gate:** all current persistence domains use their assigned owner; SQL authority is
exclusive and startup touches historical payloads only for explicit migration.

### Phase 4 — scoped API/SSE and frontend acquisition

* Implement bootstrap/activity/project/detail/history/lookup contracts and indexed
  SQL projections. Add role/scope checks to all new endpoints.
* Replace all-session SSE snapshots with scoped change sets/reset contracts.
* Refactor frontend caches, navigation/restoration, paged sidebar/search and history
  windows across main/side/general/Focus/mobile/graph activity surfaces.
* Update feature consumers that previously searched `engine.sessions`, including
  Project settings automation choices, child navigation, recency and rail activity.

**Gate:** network/query instrumentation proves untouched projects and unopened
histories are absent from initial/requested-scope traffic; visible UI still works.

### Phase 5 — focused checks, actual dev run, Playwright

Run only the scoped checks in section 8, record results and fix failures. Follow
the isolated run procedure in section 9. Record query/byte/cache measurements and
browser evidence, not just screenshots or compilation. Obtain human validation of
the delivered behavior; mark any unavailable native execution cases explicitly.

### Phase 6 — two independent review/fix rounds

After implementation and its first local/browser validation, the coordinator
arranges the two reviews explicitly requested by the user. The planner does not
spawn agents. A reviewer must be independent of the implementing author; if that
cannot be arranged, report the gate as pending rather than calling a self-check an
independent review.

* **Round 1:** review the actual changed migration/transaction/runtime/API/frontend
  code and evidence against data preservation, cutover crash states, scoped reads
  and headless operation. Record concrete findings with files/reproduction.
* Implement fixes; run only affected scoped checks and relevant browser cases.
* **Round 2:** independently review the updated implementation, verify round-one
  fixes and look for remaining migration/recovery, eager-loading and UI race gaps.
* Fix round-two findings; rerun affected checks and finalize the evidence ledger.
  A second pass over the unchanged original code is not a completed second round.

**Gate:** both reviews and their fixes are evidenced; no unresolved data-loss,
authority, scope, compilation or boundedness issue. Do not launch extra review
rounds or full suites without a further request.

### Phase 7 — version/release documentation and handoff

* Verify/update version sources listed below; document format/protocol changes and
  supported old-install upgrade path in `engine/README.md` and relevant delivery docs.
* Update `product.md` only for delivered behavior, not projected acceptance.
* Publish measured results, pending manual installer/LAN/native checks and backup
  recovery instructions in the new progress document. No commit/push/install.

## 8. Meaningful scoped verification to implement and run later

The user explicitly requested migration/recovery checks, compilation, local
execution and Playwright. These targeted checks implement that scope; they do not
authorize all repository tests. Existing tests in `transactions_test.go`,
`stream_journal_test.go`, `permission_persistence_test.go`, `message_queue_test.go`,
`message_identity_test.go`, `updates_test.go`, topic/spawn/automation/auth tests are
semantic references. Update selected fixtures for SQL; do not keep a production
global-state compatibility adapter just to avoid updating tests.

| Test group | Assertions that matter |
| --- | --- |
| Legacy import equivalence | v1/v2, format 0/1 mixed journals, sealed+active, covered-prefix duplicates, new/reordered sessions, edited favorites/text/data, nested replacement/null/delete, usage/accounting and independent sequence preservation. Compare per-record normalized digests/IDs/order against the isolated old replay oracle. |
| All domains preserved | Populate every mapped domain, including removed/archive history, side/general/Pi child/graph node/native paths, pending queue payloads, grants, receipts, graph snapshots/workspaces/outboxes, webhook payload bytes/secrets/TTL, board credentials/moves, automation recovery sets, analytics and auth/access saved-empty policy. Verify SQL then reopen and read representative APIs. |
| Immutable source | Hash checkpoint and both journals before/after success, restart, failed import/retry and old-engine rejection. Assert original checkpoint exact bytes even when compatibility transforms/replay change logical state; no reserialized “backup.” |
| Damaged legacy input | Incomplete final header/body is ignored without truncating source; complete bad CRC/JSON/format/path/size/gap fails with no authority. Unknown fields/schema and invalid ownership fail rather than dropping data. |
| Crash boundary matrix | Inject interruption after each capture flush/manifest, batch import, prepared DB close/publish, cutover manifest, sentinel replace, authority commit and recovery batch. Reopen converges once or fails closed; no duplicate events/receipts or old-state fallback. |
| Old-engine writes before fence | Interrupt prepared import, change legacy data using a disposable pre-SQL fixture, retry: new source generation is imported, earlier backup retained; stale prepared DB is not committed. |
| Authority/downgrade | Old loader/binary rejects live sentinel before executing old cleanup/work; backup unchanged. After SQL-only mutation, edit/root-add stale JSON/journal and restart: SQL result persists. Missing/corrupt committed DB fails closed. Fresh install also fences downgrade. |
| SQL atomicity/storage failure | Inject busy/full/I/O/commit failure into acceptance+receipt+queue, grant replies, graph transition and webhook event+pending. No partial rows or in-memory/SSE success; accepted operation retries return original receipt. Verify foreign keys and reopen. |
| Restart recovery idempotence | No UI connected: running chats/children/graph nodes settle correctly, queue uncertainty survives, consultations correlate, grants retained, pending native waiters handled correctly, graph notifications stable, automation interrupted/missed decisions do not auto-run, accounting not double-counted. Kill during recovery and repeat. |
| Background work and outboxes | Webhook pending release/expiry, automation admission/destination serialization, board uncertain moves and graph/Library roll-forward operate with no client. Use temporary project/skill paths and deterministic transports. No provider/model side effects in storage fixtures. |
| Indexed bounded reads | Grow unrelated histories 10× at fixed active set. Inspect SQL plans and returned/scanned record instrumentation: bootstrap never reads event payloads; history reads only its range; mutation reads/writes only related rows. No full `diskState` reconstruction/validation per commit. |
| Paging correctness | >50 adjacent work events, consultation spanning many pages, equal timestamps, event edits while paging, huge event, invalid cursor, mixed activation/native ordering. Hard row/byte caps, lossless eventual content, no duplicate/skipped IDs. |
| Frontend state/races | Extend focused `sessionState.test.ts` and `updatesConnection.test.ts`: out-of-order mutation/SSE, new scope at current revision, scoped reset, stale project response, cursor expiry, page eviction/reload, unloaded ID navigation and shared subscription teardown. |

Suggested command shapes (final test names chosen by implementer):

```text
go test -tags dev -run '^(TestSQLiteMigration|TestSQLiteRecovery|TestSQLiteAtomicity|TestSQLitePaging|TestSQLiteBoundedReads)' .
go build -tags dev -o <isolated-root>/bin/klm-sqlite-dev.exe .
npx.cmd vitest run src/features/chat/sessionState.test.ts src/features/chat/updatesConnection.test.ts <new-scoped-cache-test-path>
npm.cmd run build
```

Run Go commands from `engine`, frontend commands from `clients/desktop`. A frontend
build includes the TypeScript check; do not redundantly run overlapping checks by
default. Compile the release Go path as a separate relevant check if build-tag/CLI
code changed; never run that release binary against the user's installed data.
Only run Rust/Flutter checks if their source/build configuration changes. Packaging
is a separate delivery check, not permission to install over the user's release.

## 9. Actual local run and Playwright acceptance procedure

**Mandatory after implementation. Never run on installed release data.**

1. Create a unique directory under the approved temporary root, for example
   `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-060-<id>`, after
   verifying its parent. Use separate `data`, `projects`, `skills`, `bin`, and
   `evidence` children. Generate synthetic legacy fixtures there; do not copy live
   release credentials/history or registered project paths into the probe.
2. Build with `-tags dev`. Set process-scoped `KLM_DEV_DATA_DIR` to the absolute
   scratch `data` directory, `KLM_USER_SKILLS_DIR` to scratch `skills`, and internal
   skill/prompt overrides where needed to the intended source/fixture paths. Ensure
   queued outboxes/workspaces/native paths all resolve to scratch-owned locations.
3. Use dev API 17331 and Vite 5173 (`VITE_ENGINE_URL=http://127.0.0.1:17331`). Tauri
   dev Focus uses 17332 only if explicitly launched. If a dev port is occupied,
   identify ownership and coordinate; do not kill arbitrary processes or stop the
   installed release. Do not invoke an ambiguous `klm` from PATH.
4. For UI-only fixture runs, an explicit process-local
   `KLM_DEV_DISABLE_CLIENT_AUTH=1` is allowed only with the dev binary. Also perform
   a short separate auth-enabled dev pairing/restart check for the auth migration;
   do not claim auth was validated by a bypass run. Native model execution, if
   available/authorized, uses scratch project folders and explicit harmless prompts.
5. Start the exact scratch binary and frontend; record PID, executable, resolved
   data directory and URLs. Exercise migration, then restart **the same scratch
   store** and confirm SQL-only changes persisted and backup hashes did not change.
6. Use Playwright against the real running frontend/API (not intercepted fixture-only
   network responses). Capture requests, SSE packets, console errors and screenshots
   where useful. Test fixture/failure injection belongs behind test/dev-only paths
   or a separate probe executable, never an unauthenticated production API.
7. Stop only owned dev processes with their explicit data/control scope. Preserve
   evidence and backup hashes; clean only owned disposable artifacts when appropriate.

### Browser scenarios and evidence

* **Cold Home:** multiple projects with thousands of inactive sessions. Bootstrap
  contains compact projects/activity only; no project summaries or history requests
  until selection/open, including no hidden general-chat prefetch.
* **Select project A:** topic/session summary pages for A only, correct empty topics,
  archive section/order/recency and server-side search for a session beyond page one.
  Switching rapidly A→B cannot publish A results into B. B remains unloaded until selected.
* **Open session:** bounded latest page; scroll/load earlier, split long activity and
  consultation groups, huge event continuation, stable reading anchor. Verify exact
  expected IDs across traversed pages. Loading does not eagerly drain all pages.
* **Streaming and controls:** send a harmless real turn where harness availability
  permits; receive committed events, queue/steer/Send now/Stop, permissions/questions
  on their real owner. Open a second window/side chat and confirm shared state without
  cross-session drafts. Deterministic dev fixtures may cover presentation; label
  them separately from actual native execution evidence.
* **Unopened active session:** create/admit work without opening its history. Happening
  now/rail state changes; open later and see its durable content. Close all browser
  pages while work/release continues, reconnect and verify results. Also exercise
  a seeded restart recovery with zero clients before browser connection.
* **General/child/graph:** KLM Agent singleton and durable harness lock, source passage
  lookup outside current window, native-child view on explicit open, graph panel and
  node activation pages only while visible. Preserve graph pending-request routing.
* **Focus/navigation/mobile-responsive:** restore saved IDs without a global inventory;
  spawn/automation link into an unloaded project; multiple floating chats; side panel;
  mobile overlay/stack semantics where the host bridge is explicitly emulated.
  Browser viewport emulation alone does not prove Flutter/native host behavior.
* **Reconnect/races:** force transient disconnect, append/edit events meanwhile,
  reconnect with cursor and with expired cursor; correct scoped reset, no global
  reload or duplicate messages. Archive topic/remove project closes unavailable
  windows; unrelated unloaded histories remain absent.
* **Eviction:** visit more conversations/pages than cache budget, then reopen. Cache
  counters remain bounded; drafts/selected topic/layout survive; page retrieval is
  correct and no obsolete response overwrites a newer session.
* **Other persisted domains:** browse imported webhook/feed details and raw payload,
  Tasks cache/receipt, automation history/recovery, analytics and settings. Do not
  call real GitHub/Azure mutations merely to validate persistence.
* **Restart:** migrated fixtures retain histories/favorites/receipt behavior/auth;
  exact backup digests stable, no fresh legacy journal writes, no automatic replay
  of uncertain native operations. SQL-only writes remain after restart.

Record what passed, failed, was fixture-only, and remains human/native/installer
validation. The real local run and Playwright gate cannot be replaced by compilation.

## 10. Performance and boundedness success criteria

Use deterministic small/large generated databases with identical project/active
counts but 10× unrelated inactive sessions/events. Keep migration measurements
separate from normal warm-store startup. Record machine/build, DB size, row counts,
query plans/counters, startup time, request bytes, p50/p95 response/commit timings,
engine memory and browser retained-cache counters. These are acceptance checks,
not a general benchmark project.

Hard structural criteria:

* Post-migration startup reads no cold event text/data, webhook raw payloads, completed
  graph snapshots or historical analytics observation arrays into Go. SQL may scan
  indexed metadata for recovery/aggregates; that is not global object hydration.
* Resident engine state scales with live work, configured bounded caches and small
  configuration, not all historical records. Headless runtime memory is independent
  of which UI histories were opened.
* Initial UI data scales with project count and a bounded activity page, **not** total
  session/event count. Project lists are scoped and paginated; idle SSE sends no
  recurring all-session inventory.
* Normal history page ≤50 events and ≤2 MiB encoded payload (oversized records use
  lossless continuation). Project session pages ≤200; activity pages ≤100. Limits
  are enforced before bulk payload allocation. Open graph activity follows equivalent
  bounds; 50-event group expansion must not bypass the limit.
* Normal mutation work is O(changed/related records + indexed lookups), not a scan,
  clone, serialize or validate of every persisted session/run. Stream commit does
  not rebuild all event-ID maps. The large-history fixture must show the same logical
  rows touched for an unrelated small mutation as the small fixture.
* Backend delta/cache and frontend page-cache budgets are enforced globally as well
  as per session. Closing histories frees their caches/subscriptions subject to
  documented LRU policy. A long stream or long reading session cannot grow forever.
* SQLite WAL/checkpoint growth is bounded in steady state by configured thresholds
  and short reader lifetimes. Explicit exports have cancellation/time limits and
  documented snapshot costs; slow clients cannot retain unlimited output buffers.

Initial diagnostic targets on the same local machine: at fixed active/project set,
10× unrelated history should not increase engine steady-state retained memory by
more than 20 MiB or double warm-start/first-page/small-commit p95 latency. Measure
several samples to distinguish noise; investigate any breach with query plans and
allocation evidence. These targets supplement the hard structural checks; a fast
machine does not make an eager-load design acceptable. Import can take O(total
legacy data) time/disk, but memory must be bounded by streaming buffers, a bounded
record/frame, SQL caches and batched validation rather than total database size.

## 11. Version/build sources and release requirements

Current source already contains most requested version values. Verify consistency,
update only stale product-owned entries, and record the actual state rather than
claiming a version bump that was already present:

| File | Inspected value / action |
| --- | --- |
| `clients/desktop/package.json` | `0.6.0`; retain/verify. |
| `clients/desktop/package-lock.json` | Root and root-package `0.6.0`; keep synchronized, do not change dependency versions by global replacement. |
| `clients/desktop/src-tauri/Cargo.toml` | Package `0.6.0`; retain/verify. |
| `clients/desktop/src-tauri/Cargo.lock` | Verify the `klm-desktop` package entry against Cargo.toml, not unrelated dependencies also numbered 0.6.0. |
| `clients/desktop/src-tauri/tauri.conf.json` | `0.6.0`; source of Vite version fallback. |
| `clients/desktop/src-tauri/tauri.dev.conf.json` | Keep dev identifier/ports separate; no release lifecycle changes. |
| `clients/desktop/vite.config.ts` | Reads Tauri version unless `VITE_DESKTOP_VERSION` override. Verify displayed version for an actual release build. |
| `engine/version.go` | Default is intentionally `development`; release uses `-ldflags "-X main.engineVersion=0.6.0"`. Add explicit storage/protocol metadata independently of this label. |
| `engine/go.mod`, `engine/go.sum` | Pin SQLite driver and transitive sums, confirm Go/Windows support and pure-Go packaging. |
| `distribution/windows/build-engine.ps1`, `engine.nsi` | Both defaults already `0.6.0`; verify installer file/product version and embedded engine flag agree. |
| `distribution/windows/build-release.ps1` | Required `-Version 0.6.0`, propagates desktop/engine version and emits hashes. Preserve separate installers. |
| `clients/mobile/pubspec.yaml` | Already `0.6.0+3`; native launcher has no SQLite migration. Keep unless distributing a new APK, when its monotonically increasing build number must be chosen from actual release history. |

Never rename SQL schema version to “0.6.0” or confuse API health's current numeric
`version: 2` with application release version. Include release guidance: stop/upgrade
through the existing independent engine installer, first-start migration may take
time, exact backup location/retention, old engines are intentionally fenced, and
legacy backup restoration after SQL writes is a potentially lossy manual recovery,
not a supported automatic downgrade. Uninstall continues preserving user data and
the legacy backup. Do not delete backup files in installer cleanup.

## 12. Risks, decisions, and final handoff checklist

### Principal risks

1. **Breadth of global-state coupling:** session/history isolation touches execution,
   not only UI reads. A partial conversion can drop cold records on a later write.
2. **Legacy positional patches:** reorder/null/map/slice semantics can silently bind
   history to the wrong session unless differential fixtures cover them.
3. **Cross-file cutover on Windows:** prepared DB, sentinel and authority are not one
   filesystem transaction. Durable ordering and restart classification are mandatory.
4. **Hidden eager loads:** summaries, graph snapshot validation, auth-independent
   settings, analytics, feed grouping and private agent tools can reintroduce global
   memory/scan costs after the obvious history refactor.
5. **Large individual records:** row limits alone do not bound payload memory. Keep
   lossless payload storage and explicit continuation, including old large events.
6. **External effects during recovery:** outboxes, provider moves and native prompts
   must retain intent/uncertain semantics. Scratch data with real absolute paths is
   not isolated; fixtures must use scratch paths throughout.
7. **Sparse frontend state:** missing inventory no longer means deletion; navigation,
   archive handling and request controls require explicit availability/revisions.
8. **Independent clients/version labels:** 0.6.0 labels already exist and do not prove
   read-contract compatibility. Gate by protocol/storage contract, not label alone.
9. **Disk requirements:** retained originals + staging DB + final DB/WAL can require
   multiples of current data size. Estimate space, stream copies, surface actionable
   ENOSPC without discarding the backup, and never publish incomplete authority.

### Decisions selected by this plan

Normalized SQL rows, explicit transactional read/write sets, no global live
`diskState`, bounded active caches, read-only legacy import, immutable checkpoint+
journal backup, rejecting live `state.json` sentinel, exclusive SQLite authority,
scoped bootstrap/project/history APIs/SSE, and isolated real local/browser validation.
Filesystem authoring/native storage retains its current ownership.

### Decisions to settle before their implementation phase

* **“Before applying” resolved:** validate using isolated dev data before any real migration.
* Exact maintained SQLite driver pin/toolchain compatibility — Phase 0 gate.
* Final protocol version, cursor encoding and oversized event continuation schema —
  freeze jointly between engine/frontend in Phase 0; do not need a product redesign.
* Supported old release binaries available for a disposable downgrade test; if only
  loader fixtures are available, record that limitation and leave real-binary
  acceptance pending.
* Whether an updated native APK/installer package is part of the delivery; do not
  silently distribute/install anything from this implementation request.

### Completion / implementation handoff

The implementing coordinator should begin with Phase 0 and maintain an evidence
ledger per phase. Handoff must include changed files, current worktree ownership,
schema/driver/protocol versions, backup/cutover behavior, test commands and outputs,
actual scratch paths/PIDs/URLs, browser evidence and boundedness measurements,
review-round findings/fixes, and pending human checks.

Delivery is successful only when all of these are true:

- [ ] Every current persistence domain has a verified destination and restart path.
- [ ] Byte-exact original checkpoint and journals remain after success and retry;
      no 0.6.0 path deletes them.
- [ ] Migration/crash/downgrade fixtures pass, and post-cutover SQL is authoritative.
- [ ] No production whole-state/history eager-loading or global clone/diff workaround.
- [ ] Initial/project/open-history loading is demonstrably scoped and bounded.
- [ ] Headless queues/requests/graphs/background domains preserve existing semantics.
- [ ] Appropriate scoped tests and compilation completed with recorded evidence.
- [ ] Actual isolated dev run and Playwright scenarios completed, limitations explicit.
- [ ] Two independent implementation review/fix rounds completed and evidenced.
- [ ] Product-owned release sources agree on 0.6.0; schema/protocol versioning is separate.
- [ ] User/coordinator changes preserved; no unauthorized commits, pushes, release-data
      mutation, installation, or backup cleanup.
