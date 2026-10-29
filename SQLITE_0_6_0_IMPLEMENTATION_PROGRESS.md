# SQLite / on-demand 0.6.0 — implementation progress

## Status

**Implementation and exactly two independent review rounds complete; release/human
acceptance remains partial.** Round 1's three findings were fixed. Final round 2
reported **no material findings or changes requested**. No further review round is
planned. The isolated SQL engine/shared frontend remain available for manual use.
This ledger distinguishes delivered/reviewed code, measured evidence and remaining
runtime/release/performance acceptance; it does not certify every original plan
acceptance criterion as fulfilled.

## Boundary and baseline (2026-10-27)

- Read `AGENTS.md`, full `product.md`, and full 919-line implementation plan.
- User resolved “Before applying”: validate before real migration. Code changes and
  isolated dev probes are authorized; installed engine and real release/dev stores
  remain untouched. No commit, push, reset, installation, or unrelated process stop.
- HEAD: `64d45ef420ab1296c3f5b9b0dba5d79a27d3c36c`.
- Starting worktree: modified `CLIENT_AUTH_IMPLEMENTATION_PROGRESS.md`; untracked
  `PI_SUBAGENTS_IMPLEMENTATION_PLAN.md`, `PI_SUBAGENTS_IMPLEMENTATION_PROGRESS.md`,
  and planner-owned `SQLITE_0_6_0_IMPLEMENTATION_PLAN.md`. Preserve these changes.
- Confirmed `engine/cli.go:12–25`: **KLM_DEV_DATA_DIR**, absolute, dev-tag only.
- Actual toolchain: `go version go1.24.2 windows/amd64`. Keep Go version unchanged.
- Loaded React and code-review skills; no delegation or extra reviews.

## Phase ledger

| Phase | State | Evidence / outstanding work |
| --- | --- | --- |
| 0: contracts/inventory | Implemented; two reviews complete | All diskState roots have checked table mappings; generated all-field roundtrip passes. Normal CLI uses SQL exclusively; contract-3 compatibility is documented. Reviews sampled production coverage; not every domain has live-provider acceptance. |
| 1: SQL/import/cutover | Integrated/reviewed; synthetic migration evidence | Exclusive startup importer, exact-source retention, committed authority/no fallback and version-600 fence. Final reviewer reran v1/v2 replay and production import/backup/reopen fixtures successfully. No real-data upgrade performed. |
| 2: session runtime/repositories | Integrated/reviewed; native acceptance pending | SQL execution/controls/requests/native-child paths connected; targeted lifecycle/append/precision evidence recorded. Imported events without stream checkpoints and very large non-text metadata retain documented exact-record fallback. |
| 3: background domains/recovery/auth | Integrated/reviewed; provider acceptance pending | SQL auth/outboxes, graph/automation scheduling/execution, webhook/feed, board/recovery and analytics connected to normal startup. Scoped fixtures/CLI evidence does not establish real provider success. |
| 4: scoped API/SSE/React | Integrated/reviewed; core browser evidence | Bounded on-demand reads/caches, exact navigation/search, identity-bound state and one physical multiplexed SSE connection. Round-1 ordinary-use regressions fixed and independently checked in round 2. |
| 5: local/browser validation | Core isolated flows evidenced; acceptance incomplete | Real CLI + Vite/Playwright checks recorded; native/mobile/installer/LAN acceptance and the original plan's full 10× memory/latency measurements remain outstanding. |
| 6: two reviews/fixes | Complete — exactly 2 rounds | Round 1 requested 3 fixes, all addressed; final round 2 reported no material findings/changes requested. Reports retained; no further round started. |
| 7: release/docs | Documentation/source alignment complete; distribution pending | Product sources confirmed at 0.6.0 (mobile 0.6.0+3); migration/backup/fence/contract/limitations documented. No final release packages/install or real migration performed. |

## Storage decisions

- Exact originals, when the sources existed, remain at
  `<data>/legacy-backups/<migration-id>/state.json`, plus `stream.sealed`,
  `stream.journal`, `client-auth.json`, and `engine-access.json` when present.
- Root `state.json` becomes an explicit rejecting downgrade fence only after a
  verified prepared SQL database is published. It is **not** the original backup.
- No backup deletion/retention timer in 0.6.0. No automatic legacy fallback after
  SQL authority. No production global-state-in-SQL compatibility workaround.

## First-invocation implementation inventory (historical; superseded by continuation sections)

- `engine/go.mod`, `engine/go.sum`: pinned pure-Go `modernc.org/sqlite v1.44.3`.
  Registry module declares Go 1.24.0; installed successfully using `GOTOOLCHAIN=local`
  on Go 1.24.2. Newer inspected v1.46.2/v1.48.2/v1.49.1 require Go 1.25;
  v1.61.0 requires Go 1.26. No Go upgrade. Driver build used `CGO_ENABLED=0`.
- `engine/sqlite_store.go`: exclusive staging creation; open-existing-only; identity/
  schema verification; one writer + up to four read-only readers; every-connection
  FULL sync/foreign keys/busy timeout/disk temp/cache limit; SQL atomic callback,
  integrity/foreign-key checks, WAL authority gate/checkpoint/close; decimal revisions.
- `engine/sqlite_schema.go`: base metadata/auth/access tables. Canonical domain rows
  and their indexes/chunks are now created by `sqlite_domains.go`; the unused initial
  project/session/event projection tables were removed before integration.
- `engine/sqlite_backup.go`: streamed exact copies of all five legacy source files,
  explicit absence/hash/length manifest, exclusive generations, fsync and atomic
  manifest publication, source revalidation, injectable capture boundaries. Does not
  delete successful or incomplete backups. Cutover is now in `sqlite_cutover.go`.
- `engine/sqlite_legacy_journal.go`: strict read-only framed reader, bounded legacy
  frame, CRC/JSON/format/revision validation, sealed-before-active, covered-prefix
  validation, incomplete-tail reporting without truncation, retired Tasks filtering.
  Old runtime reader remains in place until startup conversion.
- `engine/sqlite_legacy_stage.go`: import-only SQL positional tree; token-streamed
  typed checkpoint decoding (including nested growing collections); unknown-field,
  duplicate-key and numeric-range rejection; retired Tasks filtering; streamed export.
  This tree MUST be converted to domain tables and dropped before authority. It is
  not an alternative runtime object store. Compatibility transforms are now in
  `sqlite_legacy_normalize.go`; full semantic/ownership validation remains pending.
- `engine/sqlite_legacy_patches.go`: atomic staging replay of typed replacements,
  resize/delete, SQL-based session-ID/history rebinding, format-0 event merges,
  independent graph/view revisions, scoped analytics usage mutation. Must complete
  differential and bounds checks before using outside fixtures.
- `engine/sqlite_migration_test.go`: six focused top-level tests for replay oracle,
  exact backup/retry, corruption/tails, rejected input/frame rollback, retired fields,
  connection policy/transaction rollback/reopen. More domain/cutover tests required.
- Plan clarification updated and this progress ledger created. No frontend edits yet.

## First invocation: historical commands/results (disk-full episode)

1. `go version`: `go version go1.24.2 windows/amd64`.
2. `GOTOOLCHAIN=local go get modernc.org/sqlite@v1.44.3`: succeeded, Go unchanged.
3. `gofmt -w sqlite_store.go sqlite_schema.go sqlite_backup.go sqlite_legacy_journal.go sqlite_legacy_stage.go sqlite_legacy_patches.go sqlite_migration_test.go`: succeeded.
4. With `CGO_ENABLED=0`, `GOTOOLCHAIN=local`:
   `go test -tags dev -run '^(TestSQLiteMigration|TestSQLiteAtomicity)' -count=1 -v .`
   initially failed **before compilation**: `write ...go-build...: There is not
   enough space on the disk.` C: reported **0 bytes free**.
5. Removed only sqlite/libc download ZIPs newly fetched by this task, recovering
   171,630,592 bytes temporarily. Extracted dependencies retained; no broad cache
   cleanup or unrelated files deleted. Retry exposed package-name collision with
   existing `binary` type; fixed imports to `encodingbinary`.
6. Same focused command then compiled and ran **6 top-level tests**:
   **4 passed, 2 failed** (3 failing subtests/locations), overall FAIL in 3.199s.
   Passed: ImmutableBackup (2 boundary subtests), DamagedJournal (8 cases),
   RetiredFields, AtomicityConfigurationAndReopen (all four reader connections).
   Failed: ReadOnlyReplayEquivalence (both v1/v2),
   StageRejectsUnknownAndRollsBack (valid-fixture setup). Root cause: existing
   activity-board uint64 fields use `json:",string"`; typed token importer initially
   treated their encoded strings as ordinary uint64 scalars.
7. Added quoted-number field handling to decoder/staging/patch replacement and ran
   gofmt. **The fix is NOT verified**: rerunning the same command failed compiling
   the engine package: `compile: writing output: write $WORK\b119\_pkg_.a: There is
   not enough space on the disk.` C: again reports **0 bytes free**.
8. Removed only additional task-downloaded ZIPs for x/exp (20251023), modernc memory
   1.11.0 and mathutil 1.7.1. C: still reports 0 free. No further build attempts.

All tests used Go-owned `t.TempDir` synthetic fixtures and cleanup; no real engine
data directory was opened. No application/native harness/frontend server started,
no listening ports acquired, no unrelated process stopped. No running test child
process remains. No installed engine migration or real release/dev user-data changes.

**No app-run/Playwright evidence yet.** Production SQL restart, headless recovery,
API/UI on-demand loading, acceptance measurements and independent reviews remain
pending. Reopen and paging evidence below is isolated repository/cutover evidence.

## Same-implementer resumption: disk resolved

**Resolved on same-implementer resumption:** user freed C:. Confirmed
15,162,490,880 bytes free (~14.1 GiB). User declined D: scratch; all external scratch
remains under the approved C: temporary directory. Worktree baseline unchanged.
The exact focused command above now passes **6/6 top-level tests**, including
both v1/v2 replay cases, in **3.059s**. Quoted-number fix is verified for those fixtures.
Latest disk check: **23,466,688,512 bytes free on C:**. Disk is not a current blocker.
No D: use, new cache cleanup, real data access or application processes this invocation.

### Additional code delivered this invocation

- `sqlite_domains.go`: explicit mapping for **every diskState root field**, plus
  separate tables for nested histories, topic/order/folder rows, queues/requests/
  sources/usage/capabilities, graph assessments/dependencies/receipts/activation
  events/joins/outbox writes, analytics spans/deltas/waits/reasons, automation
  recovery decisions/resolutions/occurrences, Library outbox projects/writes, and
  transcription vocabulary. Roots added later fail schema setup without a mapping.
  Child tables have owner foreign keys; content chunks reference their record.
  Small per-record configuration remains JSON; no whole-state/project/history blob.
- `sqlite_import_domains.go`: streams positional staging into `data_<domain>` rows
  and 64 KiB chunks; records null/array/map/singleton shape and counts separately;
  SHA-256 per-record content digest. Independent uint64 values remain exact.
  Native event IDs remain composite, without a new global uniqueness constraint.
  Imported acceptance IDs get an indexed sender/identity projection, including
  legacy session IDs containing `/`.
- `sqlite_content.go`: bounded digest-checked base64 byte chunks, exact offset/
  continuation and streamed record reconstruction for verification/export. No
  runtime global reconstruction. Domain-table allowlist cached once, not per chunk.
- `sqlite_history.go`: indexed bidirectional ordinal paging, max 50 rows / 2 MiB
  page budget, oversized content references, persisted event-count reads, explicit
  missing-session errors. No work/consultation group expansion or cursor scans.
- `sqlite_cutover.go`: immutable backup/source identity checks, prepared DELETE-mode
  DB close/flush, durable publish intent, live version-600 rejecting fence,
  fenced-before-authority ordering, SQL committed authority, conservative missing/
  corrupt/new-schema failure, stale prepared-generation replacement after old-engine
  writes, retained post-authority checkpoint conflicts. Does not start workers.
  **Requires a complete importer callback; only a fixture callback exists today.**
  It rejects publishing a DB that still contains `legacy_nodes`.
- `sqlite_auth.go`: strict bounded legacy auth/access import, hash/timestamp retention,
  explicit initialized policy flag, normalized origin/proxy rows, changed-client-only
  transactional persistence. Saved empty lists do not fall back to environment.
- `client_auth.go`, `engine_access_settings.go`: SQL-backed construction/save paths
  via `newClientAuthStorage(dir, store)`, commit-before-memory publication and saved
  versus active policy preserved. Existing `newClientAuth(dir)` remains the legacy
  entry point until main startup is converted; no production dual-write.
- `sqlite_session_repository.go`: related-record transaction/change sets, bounded
  metadata-only reads, exact event ID reads, append/update events, indexed collision
  checks, atomic user event + accepted fingerprint + general harness lock. Idempotent
  retries do not advance revision. Failed transactions return no publishable changes.
  These primitives do not yet replace the existing application command handlers.
- `sqlite_legacy_normalize.go`: nonmutating v1/v2 load compatibility transforms,
  graph/view revision distinction, desired Library normalization/preparing clearing,
  legacy side-role defaults, general lock inference without history hydration,
  transcription/board/webhook defaults, graph workspace defaults; separate post-replay
  folder/topic migration with SQL name identity map and preserved archives.
- Extended importer for promoted anonymous Go fields in snapshot agents. Checkpoint
  JSON flattens `AgentDefinition`, while old generated patches address it by Go name.
  Both leaf patches and whole embedded replacements now preserve surrounding fields.

### Fresh evidence

All Go commands from `engine`, with `CGO_ENABLED=0`, `GOTOOLCHAIN=local`, `-tags dev`.
All fixtures use `t.TempDir`, synthetic credentials and scratch paths; no native
process/provider effects. Tests close stores/readers and clean their fixture dirs.

1. Initial 6-test rerun: **PASS**, `ok klm/engine 3.059s`.
2. All-field domain test initially found promoted `AgentDefinition` fields missing
   from the reflection decoder. Fixed and verified both field promotion and generated
   patch grammar; no skipped domain workaround.
3. Latest combined storage command:
   ```text
   go test -tags dev -run '^(TestSQLiteMigration|TestSQLiteRecovery|TestSQLiteAtomicity|TestSQLitePaging|TestSQLiteBoundedReads)' -count=1 -v .
   PASS
   ok klm/engine 10.529s
   ```
   **17 top-level tests passed** at that point. Includes:
   - 11 named crash-boundary subtests, immutable checkpoint/journal hashes after retry
     and reopen, legacy loader rejection, SQL-only writes retained after reopen.
   - Source advancement before fence plus another interruption during replacement.
   - Post-authority stale checkpoint conflict preservation and missing-DB refusal.
   - Every current serialized field nonzero roundtrip (storage fixture, deliberately
     NOT a semantically valid complete engine startup fixture); null/empty roundtrip.
   - Embedded generated patches and max-uint64 quoted revision fields.
   - Saved credential admission/last-use/reopen/revocation; failed revocation rollback;
     no legacy auth file writes; absent versus saved-empty policy.
   - Atomic user acceptance/idempotency/harness lock, late injected SQL failure with
     no event/count/receipt/identity publication; removed-queue ID reservation.
   - 125-event history with a split consultation group, 50-row pages both directions,
     a 4,000,087-byte Unicode event recovered losslessly via chunks and checked digest.
     Encoded latest page: **7,379 bytes**. SQL plan:
     `SEARCH data_session_events USING INDEX sqlite_autoindex_data_session_events_2 (owner_id=? AND position<?)`.
   - 50 versus 500 unrelated cold events: same one-event mutation changed **6 SQL
     rows** in each fixture. This is write-set evidence, not full engine RSS evidence.
4. Added compatibility-transform test afterwards:
   `go test -tags dev -run '^TestSQLiteMigrationCompatibilityTransforms$' -count=1 -v .`
   **PASS**, both v1/v2 subtests, `ok klm/engine 2.851s`. Compared normalization to
   the old loader on disposable copies; checked topic/archive/side/general semantics
   and retry-stable topic IDs. Now **18 SQLite top-level tests exist**; the last one
   was run separately rather than redundantly rerunning unchanged tests.
5. Existing auth/access regression scope after adapting their storage seams:
   `go test -tags dev -run '^(TestClientAuth|TestEngineAccess)' -count=1 -v .`
   **8/8 PASS**, `ok klm/engine 2.408s` (pairing, sliding expiry, stream revocation,
   ingress exceptions, saved-vs-active policy and restart confirmation).
6. `gofmt` performed on all edited Go files. Successful scoped tests compile the
   Windows dev package with CGO disabled. No separate app/release binary build yet.

### Previous continuation (superseded by the latest continuation below)

This is a **context continuation, not completion or review handoff**. There is no
external blocker now. Do not ask the coordinator to start either review yet.

1. Finish the production import gate. Current order to compose is streaming
   `stageLegacyCheckpoint` -> `normalizeLegacyCheckpoint` -> `replayLegacyStage` ->
   post-replay validation/transforms including `migrateLegacyStageTopics` ->
   `importStagedDomains` + `importSQLiteAuth` -> relational/semantic validation and
   count/digest comparison -> drop import-only tables. Only then pass it to
   `openOrMigrateSQLite`. **Do not substitute `sqliteFixturePrepare` from tests.**
   Complete semantic checks for session/parent/project/topic ownership, receipt
   authorization, graph snapshots/relations, queue/capabilities, webhook/board,
   automation and analytics. Compare normalizations at correct legacy replay order.
   The current every-field test proves storage retention, not those semantics.
2. Finish specialized query/index columns before domain runtime conversion. Generic
   routing columns currently include owner/project/session/status/role/time, event
   kind/correlation/revision. Webhook expiry/dedupe/sequences, automation due/destination,
   topic ordering/archive/recency, analytics time and graph work admission need their
   actual indexed projections; do not decode chunks for broad eligibility scans.
3. Convert actual runtime and command paths. `main.go`, `store.go`, `transactions.go`,
   `stream.go`, `usage.go`, all adapters/tools and background workers STILL use global
   `diskState`; JSON journaling remains production authority. No partial SQL startup
   bridge/global hydration was introduced. Replace explicit operation read/write sets,
   exact authorization/source event lookups, queue/grant/request/receipt transactions,
   graph/outbox/automation/webhook/board/analytics recovery. Wire SQL auth construction.
   Treat `acceptUserEvent` as a storage primitive, not complete HTTP authorization/
   queue/native dispatch. Keep external effects out of SQL transactions.
4. Complete per-event mutation/content handling for oversized history (the initial
   exact-event mutation reader still has the 256 MiB legacy-record cap). Enforce
   active-cache/global memory limits, use small indexed event metadata for favorites/
   authorization, and avoid giant per-event runtime hydration. Add bounded consistent
   export/cancellation and global delta-cache eviction.
5. Implement protocol-gated bootstrap/activity/project/detail/lookup/history APIs,
   scoped SSE/change sets, frontend query/cache ownership and pagination/eviction.
   **No React files changed yet.** Initial Home must load no unopened histories;
   sparse Focus/navigation availability cannot mean deletion. Current content proposal
   uses digest-bound base64 chunks of encoded event JSON and decimal ordinal cursors;
   finalize version/representation jointly with frontend, including metadata revisions.
6. Add indexed, idempotent startup recovery, long-migration CLI readiness and orderly
   SQL shutdown/checkpoint ownership. Production outboxes must not write state.json.
7. Actual isolated build/run + Playwright remains mandatory before initial reviews.
   Use fresh approved C: temp root, absolute `KLM_DEV_DATA_DIR`/`KLM_USER_SKILLS_DIR`,
   free ports and owned processes only. No app processes/PIDs/ports or Playwright
   evidence exist yet. Verify integrated migration, zero-client workers, switching,
   paging, SQL restart/backup hashes and request scope. Then coordinator owns two
   reviews/fix rounds and relevant reruns; release/docs/version checks follow.

Precise current code edits beyond new SQL files: `engine/go.mod`, `engine/go.sum`,
`engine/client_auth.go`, `engine/engine_access_settings.go`. New SQL files are all
`engine/sqlite_*.go` listed by git status. Plan clarification and this progress file
are the only SQLite documentation edits. Preserve baseline CLIENT_AUTH progress and
PI docs. No commits, pushes, resets, installations or real migrations.

### Exact new SQL file list

```text
engine/sqlite_auth.go
engine/sqlite_auth_test.go
engine/sqlite_backup.go
engine/sqlite_content.go
engine/sqlite_cutover.go
engine/sqlite_cutover_test.go
engine/sqlite_domains.go
engine/sqlite_domains_test.go
engine/sqlite_history.go
engine/sqlite_import_domains.go
engine/sqlite_legacy_journal.go
engine/sqlite_legacy_normalize.go
engine/sqlite_legacy_normalize_test.go
engine/sqlite_legacy_patches.go
engine/sqlite_legacy_stage.go
engine/sqlite_migration_test.go
engine/sqlite_schema.go
engine/sqlite_session_repository.go
engine/sqlite_session_repository_test.go
engine/sqlite_store.go
```

## Third invocation: production import gate and headless runtime/recovery

**Still incomplete; not review-ready or release-ready.** No external blocker.
Latest disk check: C: free **19,785,195,520 bytes** (~18.4 GiB). No D: scratch used.
Baseline user/coordinator modifications remain preserved. No commit/push/reset.

### Implemented

- `sqlite_indexes.go`: explicit indexed routing/eligibility columns for sessions,
  projects/topics, graph records, webhook expiry/dedupe/sequences, automations,
  analytics, board and receipts. Only approved metadata fields are projected;
  secrets, raw payloads, prompts and captured definitions stay out of these indexes.
  Time indexes use fixed UTC nanosecond keys, avoiding RFC3339Nano lexical-order
  errors at fractional-second boundaries. Serialized record timestamps stay intact.
  Fixed nested-field extraction to allocate a fresh nested map rather than merge
  child fields into parent metadata. Non-string open event `data.childSessionId`
  values are ignored by that projection, not rejected or coerced into references.
- `sqlite_prepare.go`: **production `prepareSQLiteImport` now exists**. Pipeline:
  read-only source -> streamed stage -> checkpoint compatibility -> journal replay
  -> targeted post-replay compatibility -> topics -> domain/auth import -> semantic
  validation -> independent source/chunk/header/index/collection verification ->
  durable verification report and validation marker -> drop import-only tables.
  Journal-without-checkpoint fails closed. Fresh stores use a tiny empty seed, never
  a hydrated historical `diskState`. `GraphViewRevision` remains distinct on replay.
- `sqlite_validate*.go`: bounded record/relationship validation for projects,
  sessions/hierarchy/topics/grants, collaboration receipts and human references,
  webhooks, board, accounting, automations/recoveries and graphs. One captured graph
  is compiled/validated at a time; its activations/deliveries/joins are streamed.
  Graph dependency-cycle checking is SQL-backed. Analytics observations are checked
  in ordinal order without assembling span/usage histories. Archived-folder
  case-folded duplicates use a temporary SQL table, not an unbounded Go map.
- `sqlite_import_domains.go`: records source references while staging, then validates
  re-read actual chunks against source hashes; verifies positions/map keys, routing
  headers, split-field presence, owner association and collection coverage/counts.
  Imported record revisions start at the imported global revision, not zero.
  Auth/access originals remain byte-exact backups. Cutover now requires both no
  positional/import-reference tables and `domain_validation=complete`.
- Split desired/applied/preparing conversation capabilities into separate record
  tables, so startup/missed-schedule recovery can read desired selection without
  loading the previous captured capabilities. **80 domain tables now**, up from 77.
- `sqlite_record_writes.go`: related-record metadata/leaf writes, explicit child
  writes/deletes, indexed ordinal allocation, collection counts, quota/sequences,
  ownership guards and capped change sets. Split children cannot be silently
  replaced by a nonempty metadata payload. Existing map/array identities are retained.
- `sqlite_webhooks.go`, `sqlite_webhook_worker.go`: real SQL admission with current
  HMAC secret/config/project recheck, immutable payload/pending/quota/sequence atomicity,
  deduplication, bounded release/expiry batches and transactional PR-open automation
  trigger admission. Removed projects pause release; TTL still applies. No UI is
  needed. Native/model/provider work is never executed inside SQL transactions.
- `sqlite_change_bus.go`: scoped post-commit notices with decimal revisions/ordinals,
  no global session-ID inventory, a 64-consumer cap, one pending notice per consumer
  and 64 KiB frame cap (<=4 MiB pending). Slow consumers get scoped reset instead
  of blocking durable writes or retaining unbounded history. **HTTP/SSE transport
  and frontend consumption are not wired yet; these are notices, not final payload
  delta handling.**
- `sqlite_recover_sessions.go`: batched queue pause/uncertainty, dead permission/
  question removal, preparing-capability clearing, interrupted chat events, exact
  parent native-child invocation updates and paired consultation-card recovery.
  A durable correlation table bridges a crash between consultation and chat recovery.
- `sqlite_recover_background.go`: indexed graph interruption/notifications, automation
  run failure and missed calendar intervals, uncertain recovery notices, and open
  accounting settlement preserving unknown tails. Captured graph definitions and
  historical usage remain cold/unchanged. `recoverSQLiteRuntime` orders accounting,
  automations, graphs, consultations, then sessions before readiness; it does not
  execute work or automatically replay recovery choices.
- `main.go` adds the SQL store/change-bus ownership fields; `webhooks_worker.go`
  selects the SQL worker when that store is supplied. **`runEngine` still does not
  open/migrate SQL**: normal commands/adapters and read surfaces remain JSON/global
  state until the rest of the conversion is completed. No partial eager SQL startup
  adapter was introduced.

### Fresh focused evidence

All commands ran from `engine` with `CGO_ENABLED=0`, `GOTOOLCHAIN=local`, dev tags.
Fixtures are synthetic and disposable; no native model/provider execution or real
data directories. The final verification command explicitly set process-scoped
`TEMP` and `TMP` to `C:\Users\Samuel\AppData\Local\Temp\opencode` after checking it
exists; keep doing this for subsequent checks/builds.

- Production-import fixture initially failed the **old validation oracle** because
  its graph lacked a node name and used an invalid Choice label. Corrected the
  fixture; no production validator was weakened to make it pass.
- `go test -tags dev -run '^TestSQLiteMigrationProduction' -count=1 -v .`:
  **3 PASS**, `ok klm/engine 8.122s` at that point. Actual importer cut over and
  reopened the synthetic multi-domain store. Six invalid ownership/state subcases
  were refused without publishing SQL or changing root source bytes.
- `go test -tags dev -run '^TestSQLiteRuntime' -count=1 -v .` (before adding recovery
  cases): **3 PASS**, `ok klm/engine 3.047s`. Actual SQL worker resumed pending work
  after reopen with zero UI subscriptions; verified HMAC/dedupe/raw bytes, release
  sequence and one durable automation trigger, rollback, removal/expiry/quota, and
  scoped/slow-consumer notices. This is worker evidence, not an HTTP app run.
- Session recovery focused test: **PASS**, `ok klm/engine 2.561s`. Injected late SQL
  failure rolled back the batch; reopening between consultation/chat phases retained
  correlation. Grants, native IDs and accepted fingerprints survived; queued input
  became paused/uncertain, child parent invocation settled, and retry wrote nothing.
- Background recovery barrier focused test: **PASS**, `ok klm/engine 2.557s`.
  Reopened between phases, preserved graph snapshot digest and usage observations,
  retained pending graph results, failed interrupted automation, captured missed
  interval/time zone, marked uncertain notice, and proved no missed work replay or
  repeated recovery writes.
- After schema/metadata/capability/recovery changes, ran the relevant combined scope:
  ```text
  go test -tags dev -run '^TestSQLite' -count=1 -v .
  PASS
  ok klm/engine 37.926s
  ```
  **26 top-level tests passed**, including all previous migration/crash/paging/auth
  tests and the new production/runtime/recovery cases. All 11 cutover boundaries
  still pass. Page remains 7,379 bytes for 50 records; 4,000,087-byte content remains
  lossless. One-event mutation now changes **7 SQL rows** (new metadata index), equal
  with 50 versus 500 unrelated events; prior 6-row evidence predates that index.
- Subsequently strengthened the independent verification gate and added four
  corruption cases (chunk body, routing owner, event ownership, omitted collection
  reference), and initialized imported row revisions from the imported global floor:
  ```text
  go test -tags dev -run '^(TestSQLiteMigrationProduction|TestSQLiteMigrationVerification)' -count=1 -v .
  PASS
  ok klm/engine 11.264s
  ```
  **4 top-level tests passed** in this last affected-scope run. There are now **27
  SQLite top-level tests**; no claim that all 27 ran together after the final change.

No test worker/goroutine remains. No app/frontend server, native harness or listening
port was started. No actual app/Playwright evidence, packaged binary, full test suite
or independent review occurred. No real release/dev data or installed engine touched.

### Previous continuation — superseded by the fourth-invocation continuation below

Context continuation only; **do not start coordinator reviews yet**. No disk blocker.
The production importer and SQL recovery foundations are now usable; do not repeat
earlier scaffolding or unchanged verification. Continue the actual application path:

1. Finish filesystem authoring/Library outbox SQL roll-forward and board reconciliation.
   Existing `recoverGraphCatalogChanges` / `recoverLibraryDeletions` still call
   `saveState`. Reuse their filesystem safeguards OUTSIDE SQL transactions, then
   settle intent in SQL. Graph outbox changes must update only indexed matching
   session selections, run catalog aliases and activity graph IDs, preserving
   captured definitions. `catalogGraphId=nil` differs from explicit empty alias;
   decode candidates or add a presence projection rather than conflate them.
   Library recovery must preserve captured/active capabilities and credential refs.
   Add indexed capability-selection/credential references instead of scanning every
   historical snapshot. `LibraryAuthoringChange.Project` currently retains its nested
   Project snapshot in the per-change payload: split its topic/folder/order collections
   if needed so large project metadata is not a nested eager-load escape hatch.
2. Convert core runtime/commands and adapters from `a.state` to related-record SQL
   operations. Implement project/topic/session creation, operational-detail/queue/
   request/native mapping/grant repositories and current-turn contexts first, then
   convert stream/usage writes, authorization/source lookups, linked/general/child
   tools and controls. Keep only active handles/bounded caches, no global diskState
   hydration or clone/diff. New `putRoot`, `putExistingRecord`, `putSingleChild`,
   `sessionTransaction`, `acceptUserEvent` and recovery methods are actual primitives;
   they are not complete HTTP/runtime orchestration or authorization.
3. Normal graph scheduling/execution, automation admission/destination serialization,
   analytics recording/reporting, board operations, webhook HTTP management/ingress/
   feed projection, and Library/MCP reads/writes must use SQL. The webhook worker is
   already SQL-backed when `app.store` is nonnil, but existing HTTP handlers still
   use `a.state`. Do not initialize `app.store` in the running app until its used
   commands/read surfaces are genuinely converted. No fallback/global loader.
4. Complete event payload separation and bounded mutation for huge events. Current
   generic event JSON chunks support lossless bounded page/content reads, but exact
   mutation still decodes a whole event up to the 256 MiB legacy cap. Split text/data
   from small mutable metadata and support efficient text append and bounded source
   previews/authorization/favorites without giant runtime event hydration. Capture
   atomic per-turn stream/usage/accounting changes and publish only after commit.
5. Implement protocol-gated bootstrap, activity, paged project topics/summaries,
   exact-ID detail/lookup, bidirectional history/content, exports and scoped SSE.
   Use the index columns and `sqliteChangeBus` but finish actual delta payload/cursor
   semantics. Project-list cursor revisions should reflect membership/order/search
   changes, not every streamed text commit. No recurring all-session snapshots.
6. Refactor React acquisition/cache/navigation against that contract. **Still no
   React edits**. Initial Home must not open general history; selected projects,
   opened conversations, Focus/overlay/side/mobile surfaces own bounded shared scopes.
   Preserve drafts, geometry, explicit missing-vs-unloaded state and out-of-order
   response protection. Finish huge-content presentation and page-window eviction.
7. Replace `runEngine` startup after lock/port reservation with
   `openOrMigrateSQLite(...,prepareSQLiteImport,...)`, outbox handling and
   `recoverSQLiteRuntime`; add board reconciliation, SQL auth construction, workers,
   readiness and orderly shutdown/checkpoint. Extend CLI readiness for long migration
   via protected local status. Remove production JSON journal/checkpoint writes,
   global message indexes and historical snapshot validation.
8. Build/run the actual isolated app and use Playwright before initial reviews.
   Fresh absolute approved C: `KLM_DEV_DATA_DIR` and `KLM_USER_SKILLS_DIR`; free ports,
   exact owned binary/PIDs only. Verify integrated bootstrap/switching/paging/scopes,
   SQL restart and unchanged backup hashes, zero-client operations and auth-enabled
   pairing. Then coordinator runs the two independent review/fix rounds. Version /
   release docs/manual installer/native/LAN evidence remain pending.

### Additional files touched in this invocation

New:
```text
engine/sqlite_indexes.go
engine/sqlite_validate.go
engine/sqlite_validate_background.go
engine/sqlite_validate_automations.go
engine/sqlite_validate_graphs.go
engine/sqlite_prepare.go
engine/sqlite_prepare_test.go
engine/sqlite_record_writes.go
engine/sqlite_webhooks.go
engine/sqlite_webhook_worker.go
engine/sqlite_webhooks_test.go
engine/sqlite_change_bus.go
engine/sqlite_recover_sessions.go
engine/sqlite_recover_sessions_test.go
engine/sqlite_recover_background.go
engine/sqlite_recover_background_test.go
```
Also modified existing implementation files `sqlite_domains.go`,
`sqlite_import_domains.go`, `sqlite_session_repository.go`,
`sqlite_legacy_normalize.go`, `sqlite_cutover.go`, `sqlite_cutover_test.go`, plus
`engine/main.go`, `engine/webhooks_worker.go` and this progress ledger. Earlier
implementation files and baseline user/coordinator work remain in the worktree.

## Fourth invocation: existing command/read/adapter integration

**Incomplete; not review-ready.** This invocation prioritized edits to existing
handlers and adapter entry points instead of another independent storage test matrix.
Latest C: free-space observation: **19,049,840,640 bytes**. No external blocker.

### Existing application paths now converted when SQL owns the store

- `api.go`: health/bootstrap, project create/update/order/remove/restore, session
  create/rename/archive/restore/topic move, session visibility checks, and registration
  of scoped project/session/content/pending routes. SQL paths never hydrate `a.state`.
- `project_registration.go`: manual restoration and general-agent registration share
  SQL matching/mutation. General active matches remain unchanged and duplicate matches
  conflict with IDs; restoration preserves grouping/history. Manual restore chooses
  the earliest persisted project position when legacy removed duplicates exist.
- `topics.go`: create, create-and-move atomically, rename/color/archive and ordering.
  Ordering uses disk-backed SQL positions and persisted members, including archived
  members, instead of a project/session object graph. Existing topic identities survive.
- `general_agent.go`, `linked.go`: general singleton, side-conversation singleton,
  selection inheritance and initial harness changes use scoped SQL. General input
  checks preserve recovery-only notices and accepted-message locking. Side inheritance
  reads the compact applied selection view, not the captured tools/skill payloads.
- `session_controls.go`: shared rename/archive control entry points have SQL paths.
  Stop, Send now, model/YOLO changes and permission/question replies remain pending.
- `history.go`: actual HTTP history route uses bounded bidirectional SQL pages in SQL
  mode. `sqlite_history.go` now returns session ID and decimal snapshot/event revisions.
  History/content still preserve full oversized data behind digest-bound continuation.
- `models.go`, `metadata.go`, `project_paths.go`, `authoring.go`: existing read helpers
  use exact SQL session/project metadata before catalog/Git/path filesystem operations.
  Those external operations occur outside SQL read/write transactions.
- `stream.go`, `harness.go`, `usage.go`: actual adapter stream flush, native identity
  and usage entry points select SQL. No global event/message index is rebuilt.
  Streaming retains the existing 40 ms batching and no-op detection. Usage and sparse
  accounting delta/baseline/reason updates share a commit. The numeric accounting
  calculation is now a shared pure helper used by both paths.
- `sqlite_app.go`: application write gateway publishes only after commit and sends SQL
  failures through existing fail-closed cancellation. Owned native finalizers may
  drain produced output after command admission closes, before DB shutdown.
- `updates.go`: real `/api/updates` route now supports SQL global/project/conversation
  scopes. It registers before reading the initial revision, sends `ready`, then
  bounded scope notices; it never sends a global session inventory. Existing client
  authentication/revocation and write deadlines remain in the request pipeline.

### Outboxes and captured references

- Existing `commitGraphAuthoring` selects SQL intent -> filesystem commit -> SQL
  settlement. Graph selections/aliases/activities are queried by affected project
  and graph ID; captured graph payloads remain unchanged. Nil catalog aliases remain
  distinct from explicit deleted aliases.
- `recoverSQLGraphCatalog` and `recoverSQLLibrary` replace direct state.json writes
  for SQL recovery. Existing Library recovery HTTP handler selects the SQL path and
  publishes settled changes after commit.
- Library file preconditions and existing skill-root/revision/trash safeguards are
  retained. Project/file intents are processed incrementally, outside SQL transactions;
  partial filesystem progress is retryable using expected-or-intended hashes.
- Added indexed current selection and credential references maintained at import and
  record write/delete. Library removal updates only matching desired selections and
  automation definitions, preserves applied/captured selections and credentials needed
  by active sessions/runs, and prunes unreferenced credentials without reading histories.
- Split the nested Library authoring Project snapshot's topics/order/folders/archive
  collections into rows. Compact capability views expose selection/diagnostics without
  loading captured catalogs/skill payloads. The view preserves the full valid positive
  selection; diagnostic previews are bounded while originals remain canonical.

### Current SQL read contract (not yet consumed by App)

- Header **`X-KLM-Read-Contract: 3`**, or SSE query **`contract=3`**. Unversioned SQL
  inventory calls return 426 instead of pretending a sparse list is complete.
- Bootstrap returns `{contract, engineId, revision, projects, harnesses, general?}`;
  no project topics, session inventory, queue bodies or histories.
- `GET /api/projects/{id}/sessions`: max 200; keyset cursor, title/topic search,
  archive scope, optional topic scope. Explicit `topicId=` means Ungrouped.
- `GET /api/projects/{id}/topics`: max 200, server-side group order and cursor.
  Project/topic cursors are bound to the scope and **project list revision**. Membership,
  title/search, archive and order changes invalidate them; streamed text does not.
- Session responses are `{session, revision, eventCount, more}`. Operational children
  are bounded; applied capability views omit captured catalogs. More pending rows use
  `GET /api/sessions/{id}/pending?kind=queue|permissions|questions&afterId=...`.
  An individually oversized pending item fails explicitly rather than returning an
  endlessly repeating empty page. Pending-control UI pagination is still needed.
- History uses `before` / `after` decimal ordinals, with per-page/per-event decimal
  revisions. Content is scoped to the session and ordinal. Project list recency now
  comes from an indexed event metadata projection, not a history text scan.
- `/api/updates?scope=global|project|conversation&id=...&contract=3` emits compact
  `ready` / `invalidate` / `reset` notices. Final event delta payload consumption and
  reconnect/cache windows remain frontend/integration work.
- **Temporary safety gate `sqlRouteAvailable`** refuses unconverted routes in SQL
  mode instead of allowing access to the empty legacy state. This must not be treated
  as the finished product: remove it only after all required endpoints use SQL.
  `runEngine` still starts the legacy path; SQL startup is not enabled prematurely.

### React change

`clients/desktop/src/features/chat/HistoryLoader.tsx` now loads earlier pages through
the existing explicit button, only while visible, and aborts/discards requests on
session change/unmount. Removed the observer that could recursively drain a collapsed
history. Preserved the prepend anchor and existing design-system/copy. This component
still uses the old client page shape until App/engine.ts are migrated; it is not a
claim of completed contract-3 frontend acquisition or bounded global caching.

### Focused checks and actual evidence

All Go checks used dev tags, CGO disabled and the existing Go 1.24.2 toolchain, with
process-scoped `TEMP`/`TMP` under the approved C: opencode directory. No full suite.

- Added **three integration-oriented tests**, using existing application methods:
  `TestSQLiteHTTPProjectTopicSessionVertical`, `TestSQLiteExistingAdapterWrites`,
  `TestSQLiteOutboxIntegration`.
- The HTTP test drives `app.routes()` with an intentionally empty legacy `diskState`:
  project/topic/session CRUD, archive/restore constraints, general/side singleton,
  harness choice, project ordering/removal/restoration, metadata-only bootstrap,
  project-scoped pages, explicit Ungrouped and invalidated cursors, and bounded history.
  It also starts an owned loopback `httptest.Server` on an OS-selected free port and
  reads actual SSE `ready` and post-commit project invalidation over HTTP. The server
  and stream are closed by test cleanup. This is **not** the normal CLI app run or
  Playwright evidence. Test uses the explicit dev auth bypass, not pairing evidence.
- Adapter test invokes actual `p.put`, `p.nativeID`, `p.setUsageBaseline` and
  `p.setUsageFinal` with SQL and no legacy state. It checks append/replace/no-op,
  native path preservation, usage/accounting atomicity and shutdown drain. No native
  harness executable/model was launched; the configured fixture binary is never run.
- Outbox test executes real scratch authoring filesystem operations, external-conflict
  refusal/retry, selected-connection removal, active capture credential retention,
  unused credential pruning, graph rename settlement and unchanged snapshot digest.
- One fixture compile error (`turn` variable shadowed the type) was corrected. No
  production rule was weakened. Later focused integration runs passed:
  ```text
  go test -tags dev -run '^(TestSQLiteHTTPProjectTopicSessionVertical|TestSQLiteExistingAdapterWrites|TestSQLiteOutboxIntegration)$' -count=1 -v .
  PASS
  ok klm/engine 4.545s
  ```
  After history revisions and outbox notification wiring, the affected HTTP/outbox
  pair passed in **4.058s**. After the final capability-view size bound adjustment,
  the HTTP vertical test passed in **2.750s**. No all-SQLite rerun this invocation.
- Frontend check initially used a nonexistent `tsconfig.app.json` and failed with
  TS5058. Correct command:
  ```text
  .\node_modules\.bin\tsc.cmd --noEmit -p tsconfig.json
  ```
  **Exit 0, no diagnostics.** No Vite/package/Rust/Flutter build or browser run.
- Fixed a scalar compatibility issue observed during integration: legacy JSON `null`
  numeric revision/sequence fields are decoded as zero by Go. Normalize those typed
  values before writing SQL decimal metadata; never store the literal `null` as a
  live revision/sequence. Originals remain byte-exact in backups.

### Previous continuation — superseded by the fifth-invocation continuation below

Do not begin reviews. No external blocker. Do not spend another invocation rebuilding
the schema/importer or rerunning unchanged matrices. Continue application integration:

1. Convert **message admission + `scheduleMessagesLocked` + `execute` + finalization**
   together. These still use `a.state` and are blocked by the SQL route safety gate.
   `execute` currently reads global state for analytics reservation, topic/general
   context, graph hooks, capability preparation, automation settings and finalization.
   Convert each dependency with scoped records/current-turn handles; never give it a
   hydrated global/session history compatibility object. Existing stream/native-ID/
   usage persistence is already converted and should be reused.
2. Convert queue controls, Stop/Send now and pending permissions/questions, live YOLO,
   grants and linked/general/child private tools. Preserve cancellation-before-write
   for Stop, confirmed owned shutdown before replacements, exact native request ownership,
   uncertainty and human authorization references. `stopSessionLocked`,
   `finishTurnStateLocked`, analytics lifecycle, `prepareCapabilities`, and adapters'
   history reconciliation still need SQL work. Their public routes must remain blocked
   until they actually operate on SQL, not merely return empty legacy state.
3. Complete normal graph/automation scheduling/execution, board reconciliation/runtime,
   analytics reports/recording, Library/MCP management and webhook HTTP/feed paths.
   Outbox recovery and the SQL webhook worker/recovery barrier are already implemented.
   Whitelist only genuinely converted endpoints during integration. Signed webhook/
   OAuth ingress must be exempted from the client read-contract header when converted.
4. Finish event text/data separation and efficient append/favorite/source previews:
   current exact-event mutation still reads a whole event (256 MiB legacy cap).
   Keep the already-integrated adapter entry point but replace its large-event internals.
   Enforce active/global cache budgets. `p.keys` and native reconciliation maps also
   need bounded current-turn identities/indexed lookup rather than lifetime growth.
5. Migrate `engine.ts`, `App.tsx`, `sessionState.ts` and view consumers to contract 3.
   **App still polls/reconciles the old global inventory and uses numeric history
   revisions.** Add health/contract preflight; do not download old global inventories
   as compatibility fallback. Separate project summary pages, exact detail lookup and
   bounded bidirectional history windows. Share subscription ownership across main,
   general dock, overlay, side, Focus and covered mobile sheets. Do not auto-subscribe
   to unopened native-child history. Preserve unloaded/missing distinction and saved
   Focus/navigation/drafts. General Home must not create/open history while dock is shut.
6. Add compact global activity/rail aggregates and runtime-only activity publication;
   project session views currently include indexed recency but no complete activity
   aggregation. Connect scoped SSE notices to cache refresh/deltas without losing
   mutations or advancing contiguous cursors incorrectly. Finish group continuation,
   oversized content UI, pending-control pagination, eviction and scroll anchors.
7. Only once the used paths are converted, wire normal `runEngine` to SQL importer,
   outboxes, recovery barrier, SQL auth, workers and shutdown; replace long-migration
   CLI readiness and retire production JSON checkpoints/journals/global indexes.
   `sqlRouteAvailable` is a temporary fail-closed guard, not a deliverable capability cut.
8. Run the normal isolated CLI app and frontend, then Playwright, before requesting
   the coordinator's two reviews. Still no normal app PID/URL, Playwright screenshot,
   full native turn, LAN/installer acceptance, or review evidence. Use fresh absolute
   C: scratch data/skills, free ports and exact owned processes; never real dev/release data.

### Precise files changed this invocation

New:
```text
engine/sqlite_library_references.go
engine/sqlite_app.go
engine/sqlite_outboxes.go
engine/sqlite_commands.go
engine/sqlite_api.go
engine/sqlite_http_test.go
engine/sqlite_adapter.go
engine/sqlite_integration_test.go
engine/sqlite_updates.go
```
Existing application files:
```text
engine/api.go
engine/project_registration.go
engine/session_controls.go
engine/topics.go
engine/general_agent.go
engine/history.go
engine/graph_snapshot.go
engine/library_api.go
engine/stream.go
engine/usage.go
engine/harness.go
engine/updates.go
engine/metadata.go
engine/authoring.go
engine/models.go
engine/project_paths.go
engine/linked.go
clients/desktop/src/features/chat/HistoryLoader.tsx
```
Existing SQL implementation files: `sqlite_domains.go`, `sqlite_import_domains.go`,
`sqlite_indexes.go`, `sqlite_record_writes.go`, `sqlite_session_repository.go`,
`sqlite_legacy_normalize.go`, `sqlite_history.go`; plus this progress document.
Baseline CLIENT_AUTH progress and PI documents preserved. No commit/push/reset,
installation, real migration or unrelated process stop. All owned test streams,
listeners, DB handles and fixture directories were closed/cleaned by their tests.

## Fifth invocation: message-to-execution SQL integration

**Incomplete; not review-ready.** Normal `runEngine` remains on the legacy startup
path until remaining domains and client acquisition are converted. No real migration
or installed-engine operation was performed. No commit/push/reset or delegation.

### Application entry points converted

- `api.message` now dispatches SQL admission: existing input validation, real filesystem
  mention preparation outside SQL/app mutex, rechecked session/project/harness/Stop
  generation, fingerprint/collision checks, queue capacity, captured prepared payload,
  durable general harness lock and atomic queue/receipt acceptance. Agent instructions
  remain `agent_prompt`, not human authorization.
- `scheduleMessagesLocked` selects indexed executable queue owners in SQL, excludes
  archived/removed/automation-busy destinations before limiting, preserves steering /
  queue-head / exclusive-interrupt priority, commits user-event incorporation and running
  state, then invokes the **existing `execute` function**. It reads no cold history.
- Native steering claim/ack/incorporation, Pi echo correlation, OpenCode native-session
  boundary checks, explicit queue removal and Send now are SQL-backed. Prepared payloads
  survive uncertain delivery, old accepted identities remain reserved and uncertain
  retries retain the existing replacement-ID behavior.
- `execute` now reads scoped turn metadata/native identity/event count, uses SQL
  accounting reservation, topic/general context, capability selection and automation
  settings, and calls SQL finalization. Pi/Codex/OpenCode startup reads, resolved model
  persistence, native child creation/status/capabilities and invocation projections
  have SQL branches. Existing adapter streaming/native-ID/usage paths are reused.
- Capability preparation preserves the existing version/isolation/skill validation and
  MCP preparation flow. Desired/applied/preparing updates stay at complete turn
  boundaries. MCP credential lookup/refresh uses SQL; upstream secrets remain protected.
  The capability HTTP endpoint now uses decimal selection revisions in SQL mode.
- SQL accounting covers reservation, dispatch, acceptance, bounded tail-span updates,
  attributable/unknown waits, confirmed-end correction, termination and child-partial
  observations. Checkpointing visits active handles only; no span/usage histories are
  assembled. Confirmed end preserves clock-discontinuity unknown tails.
- `finishTurnStateLocked` settles request cleanup, queue uncertainty, in-flight event
  statuses, consultations, automation outcome/recovery notices and credential pruning
  in SQL. Replacement release still requires confirmed shutdown. Owned finalizers can
  persist already-produced output after admission closes.
- Stop now cancels before its SQL write, preserves unrelated graph execution, cancels
  exact consultations, stops retained owned runtimes and settles an automation only
  after confirmed shutdown. Existing `waitSessionStop` remains the confirmation gate.
- Native permission/question creation/dismissal and exact reply controls now use SQL.
  Reply callbacks run outside SQL/app locks. Saved-rule lookup is indexed and deny-first;
  live YOLO preserves the active Codex sandbox downgrade restriction. Automatic replies
  retain automatic attribution. Elicitation still needs the separate explicit human
  endpoint and never becomes a general-agent or YOLO answer.
- Added `permission_reply_intents` for new SQL replies. Newly staged grants are excluded
  from matching until delivery settlement; failure rolls them back. Startup removes
  unresolved new intents/grants after a crash, without changing imported legacy grants.
  Normal turn finalization does not race away an in-flight reply's intent.
- Added owned-runtime activity metadata with durable revision notifications and current
  reasoning reset/update. This supports upcoming activity/client projection; global
  activity/rail endpoints are still pending.
- Basic linked discover/read/search/ask/answer and consultation scheduling/expiry now
  operate on scoped SQL records. Retained the five-second wait, reciprocal deferral,
  four-outstanding-question bound, duplicate-question check and correlated continuation.
  The private read output is byte-bounded, but its whole-event input decoding still
  shares the pending large-event limitation noted below.
- General-agent SQL dispatch now covers registration, project/session inventory,
  session detail, attributed `session_send`, Stop/queue/archive/rename, request replies,
  models/settings and graph selection. Other private operations still fail explicitly
  and are not complete (see continuation). Initial normal-agent topic assignment is SQL.
- Both legacy commit entry points now **fail closed if called with `app.store != nil`**.
  An overlooked writer cannot create a competing JSON journal. This is a development
  safety guard, not permission to leave any supported feature unconverted.

### Checks and process/data ownership

Only affected scoped tests were run; no full SQLite or repository suite this invocation.
All checks used `-tags dev`, `CGO_ENABLED=0`, `GOTOOLCHAIN=local` (Go 1.24.2).

New focused tests in `engine/sqlite_execution_test.go`:

1. `TestSQLiteMessageExecutionLifecycle`: actual HTTP message route -> durable queue
   admission -> scheduler -> existing `execute` -> capability preparation failure ->
   SQL finalization/accounting. Uses an intentionally nonexistent **absolute scratch
   harness path**, so no installed harness/model is invoked. Verifies one user event,
   one error event, consumed queue, closed accounting, idempotent retry and no legacy
   state/index hydration. This is failure-lifecycle evidence, not a successful model run.
2. `TestSQLiteOwnedRequestsAndGrantRollback`: permission delivery failure removes its
   provisional grant/intent; successful session-scoped reply persists the exact scope;
   unoffered question answers fail and offered answers remove the pending question.
3. `TestSQLiteSteeringInterruptAndStop`: actual admission/steering helpers claim and
   incorporate a two-input batch once; Send now reserves/cancels, confirmed finalization
   releases only the selected replacement, and explicit Stop pauses it/cancels restart.
   Uses owned in-memory turn handles, not a native process fixture.
4. `TestSQLiteGeneralRegistrationAndNativeChild`: real private general-tool dispatcher
   registers/reuses a project and searches it; native child callbacks create an
   engine-owned child, persist output and settle the exact parent invocation. Confirms
   no synthetic user project/root role and no legacy state fallback.

Latest combined affected scope:
```text
go test -tags dev -run '^(TestSQLiteMessageExecutionLifecycle|TestSQLiteOwnedRequestsAndGrantRollback|TestSQLiteSteeringInterruptAndStop|TestSQLiteGeneralRegistrationAndNativeChild|TestSQLiteHTTPProjectTopicSessionVertical)$' -count=1 -v .
PASS
ok klm/engine 5.534s
```
**5 tests passed**. After adding the general inventory's explicit missing-project
check, the affected general-registration/native-child test passed again in **2.900s**.
Earlier execution/request runs passed in 3.468s, 4.223s, 4.802s and 4.835s as their
dependent paths were integrated; no unchanged broad matrix was rerun.

One check could not create Go's work directory because the previously approved
`C:\Users\Samuel\AppData\Local\Temp\opencode` directory had disappeared externally.
Confirmed its parent existed and restored only that approved scratch root using the
marker `.klm-sqlite-scratch`. C: had **16,705,302,528 bytes free**; this was not a disk-full
blocker. Subsequent checks explicitly used process-local TEMP/TMP there and passed.
Fixture `KLM_USER_SKILLS_DIR` values were absolute scratch paths. No installed engine,
real release/dev store, real skill library, provider/model execution, or unrelated
process was modified/stopped. No normal CLI/frontend process or Playwright was run.

### Latest exact continuation — SAME implementer

Context continuation, **not completion/review handoff**. Do not rebuild the existing
storage foundation or rerun unchanged matrices. Main priorities now:

1. Complete the remaining SQL execution/tool dependencies, then enable normal startup.
   Admission, scheduler, execute, finalization, controls and request persistence now
   have real SQL paths; build on them. Remaining normal graph/automation scheduling
   is still legacy: the SQL `scheduleLinkedLocked` branch currently starts messages
   and consultations only. Do not mistake that for completed graph/automation delivery.
2. Finish private general/history/feed/create/spawn operations and graph tools.
   `sqlite_general_execution.go` deliberately returns an unavailable error for
   unconverted operations; creation/spawn options, general history/search/events,
   feed reads and project graph listing still need complete SQL dispatch. Graph
   conversation context/catalog reads are SQL, but mutation/inspection/tool handlers
   and normal graph execution/scheduling remain to convert. Remove all temporary
   unavailable behavior before release/review. Maintain explicit-project creation,
   real human references, receipts and agent-authored provenance.
3. Complete Library/MCP management/OAuth HTTP paths, automation APIs/admission/worker,
   board reconciliation/provider operations and analytics reports. Core capability
   capture, credentials, request/elicitation flow and accounting lifecycle are already
   SQL-backed. Preserve captured data and do not read historical snapshots to evaluate
   live eligibility. Signed webhook/OAuth ingress needs its correct contract-header
   exemption when public handlers are converted.
4. Finish bounded event text/data handling and native reconciliation/current-turn key
   caches. Current exact-event reads, private nearby-message/history reads and terminal
   updates may still decode one whole event up to the old 256 MiB cap. That remains a
   real limitation, not the final bounded-runtime architecture. Public history pages
   are bounded and lossless continuation is implemented, but runtime hot paths need
   efficient field-level append/preview/favorite handling.
5. Convert `runEngine` to exclusive SQL startup/outbox/recovery/auth/worker/shutdown,
   migration readiness/status and journal retirement only after remaining used paths
   are converted. `sqlRouteAvailable` and the legacy-write guard are temporary safety
   gates, not a supported partial-feature release. Main startup still uses JSON today.
6. Finish the contract-3 React acquisition/cache/navigation/pagination work. App still
   uses global inventory and numeric revisions; only HistoryLoader cancellation/manual
   paging was changed previously. Implement preflight, compact global activity, exact
   lookups, scoped/ref-counted subscriptions, bounded page windows/LRU, draft/Focus
   preservation, no hidden general/native-child history loads, continuation/scroll
   behavior and pending-request pagination. Runtime activity metadata now exists.
7. Actual isolated normal CLI + frontend run and Playwright remain mandatory, followed
   by coordinator's two fresh independent review/fix rounds. There is still no normal
   app PID/URL, successful native model turn, Playwright screenshot, installer/LAN
   acceptance, or independent review evidence. Use only approved C: scratch data and
   skill roots, free ports and owned processes. No real migration/install/commit/push.

### Exact new files this invocation

```text
engine/sqlite_execution_analytics.go
engine/sqlite_execution_queue.go
engine/sqlite_execution_context.go
engine/sqlite_execution_finish.go
engine/sqlite_execution_requests.go
engine/sqlite_execution_stop.go
engine/sqlite_execution_settings.go
engine/sqlite_execution_test.go
engine/sqlite_linked_execution.go
engine/sqlite_topic_tool.go
engine/sqlite_general_execution.go
engine/sqlite_capabilities_api.go
```
Existing files edited: `analytics_records.go`, `message_queue.go`, `api.go`,
`capabilities.go`, `mcp_connections.go`, `mcp_oauth.go`, `mcp_elicitation.go`,
`harness.go`, `graph_tools.go`, `pi_interactive.go`, `codex_interactive.go`,
`opencode_interactive.go`, `runtime.go`, `models.go`, `store.go`, `transactions.go`,
`subagents.go`, `pi_subagents.go`, `linked.go`, `linked_bridge.go`, `permissions.go`,
`permission_policy.go`, `questions.go`, `live_yolo.go`, `session_controls.go`,
`general_agent_tools.go`, `session_commands.go`, `topics.go`, and existing SQL files
`sqlite_adapter.go`, `sqlite_schema.go`, `sqlite_recover_background.go`,
`sqlite_outboxes.go`, `sqlite_indexes.go`, `sqlite_commands.go`, `sqlite_api.go`.
All paths above are under `engine/`; this progress file was updated too. No frontend
source changes or frontend checks in this invocation. Prior unrelated work preserved.

## Sixth continuation — SQL spawn and automation execution

This is implementation progress, **not release completion or a review handoff**.
Normal startup and the React client still need their SQL/contract-3 conversion.

### Integrated

- Formatted and compiled the previously added `sqlite_spawn.go` paths and existing
  `session_spawn.go`, `linked_bridge.go`, and `sqlite_general_execution.go` dispatch.
  Creation validates a real human message, checks sender-scoped durable receipts,
  validates the model catalog outside the write transaction, rechecks ownership and
  settings, then atomically creates the conversation, attributed queue, receipt,
  and source card. General creation remains explicit-project; removed projects
  cannot return a general creation receipt as though they were available.
- Fixed source-preference corruption during automation-owned spawning: the frozen
  execution settings are inherited by the child but are never written back into
  the source conversation's saved preferences.
- Added `sqlite_automations_worker.go` and routed the existing
  `scheduleAutomationsLocked` entry point to it under SQL authority. Candidate
  discovery uses bounded 100-ID pages of due schedules, pending FIFO owners, and
  runs needing continuation. Historical run payloads are not hydrated to decide
  eligibility. Parallel webhook admission is bounded to 100 heads per automation
  per pass; a blocked automation does not hide another automation's pending work.
- Schedule occurrence/dedupe/once-consumption updates are atomic. Admission freezes
  current or recovery settings, serializes reusable destinations against chat,
  queued messages, graph and consultation work, and atomically writes the consumed
  trigger, session link, attributed prompt, and active run. Existing destinations
  use follow-up prompts and preserve saved preferences/native continuation.
- The worker starts the existing `execute` pipeline only after COMMIT and directory
  checks. Continuations wait for correlated graph/consultation work and preserve
  the one-reminder rule. `automation_succeed`/`automation_fail` now write SQL through
  the shared dispatcher, enforcing exact live root-turn/run ownership and rejecting
  conflicting or late classification. Existing SQL settlement remains responsible
  for the final outcome after process settlement.
- Added indexed `needsContinuation`, snapshot mode, and snapshot session-mode
  projections for live scheduling. These changes are still pre-release schema work;
  only newly created isolated fixture databases were used.

### Narrow verification

All Go commands used `CGO_ENABLED=0`, `GOTOOLCHAIN=local`, and the approved absolute
scratch `TEMP`/`TMP`. Each fixture used fresh temporary engine/project/skill roots.

- `go test -tags dev -run '^TestSQLiteGeneralRegistrationAndNativeChild$' -count=1 .`
  passed (`4.671s`) after formatting the spawn dispatch changes.
- The focused two-case command ran `TestSQLiteSpawnFrozenSettingsAndReceipt` and
  `TestSQLiteAutomationSchedulingAndOutcome`. Spawn passed (`0.75s`); automation
  initially exposed a metadata-only write receiving creation/capability collections.
  Fixed admission to reload the exact persisted metadata before writing its run link.
- Reran only the failed case:
  `go test -tags dev -run '^TestSQLiteAutomationSchedulingAndOutcome$' -count=1 -v .`
  passed (`ok klm/engine 3.013s`). No unchanged broad matrix or full suite was run.
- `git diff --check` completed without whitespace errors (existing LF/CRLF warnings).
- Spawn evidence covers rejecting agent-authored authorization, available options,
  identical retry/conflict behavior, inherited topic/frozen settings, preservation
  of source preferences, attributed child input, and no legacy-state hydration.
- Automation evidence covers a paused queue blocking reuse while another schedule
  proceeds, durable one-time consumption, real `execute` failure settlement with a
  deliberately missing scratch harness, follow-up/current-setting capture, no
  duplicate admission, outcome conflict/late-reply rejection, and successful
  classification settlement. **No installed harness/model success is claimed.**

### Next exact work

1. SQL automation HTTP management, recovery notification delivery/read/resolve, and
   graph scheduling/tools remain unfinished. The regular automation worker now has
   its SQL scheduling path; the general recovery worker still needs conversion.
   `scheduleLinkedLocked` still starts only SQL messages and consultations. Build on
   these integrations rather than creating another scheduler or global state adapter.
2. General creation/options and project graph listing now have SQL dispatch; finish
   general history/search/events/feed. Other graph/Library/MCP/board/analytics HTTP
   operations and bounded event mutation remain as recorded in the fifth continuation.
3. Finish exclusive SQL startup and the contract-3 frontend, then perform the actual
   isolated CLI/frontend run and Playwright checks before coordinator review rounds.
   There is still no complete app PID/URL, native model success, browser acceptance,
   or independent review evidence for this refactor.

New files in this continuation: `engine/sqlite_spawn.go` (started before this
context continuation), `engine/sqlite_automations_worker.go`, and
`engine/sqlite_automations_execution_test.go`. Existing edits: the three spawn/tool
dispatch files above, `engine/sqlite_indexes.go`, `engine/automations_worker.go`,
`engine/automation_tools.go`, and this ledger. The newly observed Linux delivery
files and changes, alongside the previously recorded unrelated files, were preserved.

## Seventh continuation — automation APIs, graph execution, MCP management

**Context boundary, not completion or an approval gate.** The user explicitly
instructed this same implementer to continue the full scope across resumptions,
without awaiting approval between subphases. No additional agents, commits,
installation, real data migration, or unrelated process shutdown are authorized.
The installed engine and real dev/release data remain untouched. Normal startup
still uses legacy persistence; enabling SQL startup before the remaining production
dependencies are converted would be an incomplete release.

### Integrated production paths

**Automation management/recovery**

- Existing automation routes select SQL handlers under SQL authority: list, exact
  detail, create/edit, toggle, delete, run history, Run now, schedule preview, and
  engine time-zone settings. Added exact GET
  `/api/projects/{id}/automations/{automationId}`. Revisions and received sequences
  on these SQL wire objects are decimal strings. Creation receipts, fixed reusable
  links, current-session settings, new-topic creation, Library recovery gates and
  known MCP tool validation are rechecked within the final related-record write.
  Catalog/filesystem validation and credential protection stay outside SQL writes.
- Automation lists use `offset`/`limit` (maximum 50), `hasMore` and `nextOffset`.
  Topics are acquired through the existing project-topic page endpoint, rather than
  embedded as an unbounded collection in automation listings. Run history joins
  indexed run/session metadata without reading old run captures.
- General-agent recovery read/resolve now uses SQL. Groups and occurrences are
  paged; only the requested receipt/selected decisions are read. Real later human
  input is checked by event position and type. Each selected occurrence has its
  own durable decision and recover-trigger; exact operation retries return the
  saved result, conflicts fail, and unresolved interval boundaries are advanced.
  Decisions/receipts/selected occurrence IDs remain separate child rows.
- Regular recovery notification scheduling now uses SQL: bounded pending batches,
  no implicit general harness lock, normal execution only once configured/free,
  and uncertain-delivery handling through the existing SQL finalizer. No replay
  of missed work occurs without the documented later human decision.

**Graph tools, scheduling, workspaces, runtime and HTTP**

- Private graph activity invoke/update/assess and bounded activity/run/event reads
  now dispatch to SQL. Authorization uses real human records. Dependency changes
  are atomic and reject cross-owner/self/duplicate/cyclic dependencies through
  indexed/recursive queries. Operation receipts and assessments are independent
  rows; activity version/order and run/occurrence wire values use decimal strings.
- SQL graph tools accept bounded `offset`/`limit` pages and named collections.
  `graph_activities`: optional `activityId`, with authorization/dependencies/
  assessments/operations collections. `graph_get_run`: activations/deliveries/
  joinRounds/workspaces/workspaceUses, or joinExpected/joinArrivals plus `roundId`.
  Oversized records return ownership-scoped digest/content descriptors.
  `graph_record_content` reads 16 KiB base64 UTF-8 JSON chunks with decimal offsets.
  Run summary reads do not hydrate retained activation/delivery/workspace history.
- `scheduleLinkedLocked` now invokes SQL message, graph-activity, graph-result and
  consultation scheduling. Activity selection uses indexed ready status,
  dependencies, priority/correction/order and the active automation allowlist.
  Capture reservations are held in owned handles; snapshots/catalog/model/Library
  validation and filesystem preflight occur before atomic run reservation.
- Snapshot capture/revalidation reads SQL project and MCP configuration. Run
  snapshots remain separate immutable payloads. Reservation atomically writes the
  activity/run/capture, selected graph and analytics timing. Runtime handles and
  native execution start only after COMMIT.
- The existing `runGraph`, node launch, Terminal, Fork, Join and workspace entry
  points have SQL paths. Initial/new/reused workspace identity, source-run ownership,
  Git provenance and association rules are retained. External Git/worktree work is
  between durable reservation and settlement, never inside SQL transactions.
  Join inputs load only the captured frontier of that active round.
- SQL transitions cover Choice ownership/reservation, idempotent violation records,
  seal → drain → acceptance ordering, missing-Choice correction, native completion,
  per-edge delivery, causal Join rounds, terminal-result checks, cancellation and
  final settlement. Only live owned worker handles are held by the graph loop.
  The loop and finalizers do not reconstruct global `diskState` or retained histories.
- Result notifications use the existing conversation execution pipeline with SQL
  ownership, archive deferral, attempt/backoff state and normal completion callbacks.
  Their prompt contains scoped activity/run metadata; additional history is paged.
- Graph HTTP GET/PATCH conversation state, run summary, and node activity are now
  SQL-backed. Projections aggregate indexed node/Choice/Join metadata and page
  projected pending requests. Node activation/event windows are bounded. Added
  `/api/sessions/{id}/graph/content` for ownership-checked digest continuations.
  Graph HTTP revisions/occurrences are strings. The frontend still needs conversion.

**Library/MCP/OAuth**

- MCP create/edit, exact detail, connection test, OAuth begin/callback/disconnect
  now read/write SQL while preserving existing validation, locks, encryption,
  PKCE/state/issuer/redirect checks and native credential refresh paths.
  OAuth network exchanges and credential protection are outside SQL writes.
  The callback is explicitly exempt from the client read-contract header, retaining
  its independent state validation and public ingress authentication boundary.
- Added GET `/api/library/mcp/{connectionID}` for full selected connection/catalog
  detail. SQL Library listing has separate bounded connection and skill windows.
  Connection list catalogs contain revision/resources/toolCount, with
  `detailRequired=true`; fetch the exact endpoint for tool schemas. Connection pages
  use `offset`, `limit`, `hasMore`, `nextOffset`; skill pages use `skillsOffset`,
  `skillsHasMore`, `skillsNextOffset`. Harness status is still version-gated.
- Added maintained `credential_connections` and `capture_references` projections.
  Disconnect can revoke current and captured credentials by connection ID without
  opening historical captures. Capture references also provide the foundation for
  the remaining Library deletion preview conversion. They contain only IDs, never
  upstream secrets, and are updated/imported/deleted with their owning records.
- **Library deletion preview/admission is still legacy and must be converted next.**
  SQL recovery/settlement outboxes already exist in `sqlite_outboxes.go`; reuse them.
  Do not claim complete Library management merely because MCP CRUD is connected.

### Verification actually performed

All commands used Go dev tags, `CGO_ENABLED=0`, `GOTOOLCHAIN=local`, and the approved
absolute scratch `TEMP`/`TMP`; application fixtures used fresh temp engine/project/
skill roots. Existing installed harnesses and real engine data were not used.

1. `go test -tags dev -run '^(TestSQLiteAutomationManagementHTTP|TestSQLiteAutomationRecoveryDecisions)$' -count=1 -v .`
   passed, `ok klm/engine 5.180s`. Covers real HTTP management/receipt/toggle/delete/
   timezone routes and paged recovery with later-human checks, exact-set receipts,
   per-occurrence admission, conflict rejection and split persistence.
2. Graph activity tool check initially rejected the **test's** explicitly empty
   workspace mode, correctly. Fixed the fixture to request `original`.
   Its isolated rerun passed (`3.426s`). After connecting runtime scheduling, the
   tool-only fixture reserves its capture slot to keep that test scoped to tools.
3. `go test -tags dev -run '^(TestSQLiteGraphActivityTools|TestSQLiteGraphChoiceSettlement|TestSQLiteGraphRuntimeFailureSettlement)$' -count=1 -v .`
   passed, `5.018s`. Covers real tool transactions and cycle rollback, native turn
   Choice ownership/seal/drain/acceptance, terminal notification, and the real SQL
   graph loop launching normal execution against an intentionally missing harness
   and settling its failure. The runtime fixture seeds an already captured run;
   it does **not** prove end-to-end catalog capture or a successful model turn.
4. `go test -tags dev -run '^TestSQLiteGraphHTTPPages$' -count=1 -v .`
   passed, `3.296s`. Covers real graph HTTP projection, projected permission,
   decimal run revision and bounded node activity/native event read.
5. `go test -tags dev -run '^TestSQLiteMCPManagementAndDisconnect$' -count=1 -v .`
   passed, `3.724s`. Uses an owned loopback HTTP fixture for a deliberately failed
   MCP probe. Covers real CRUD/detail/list/test/disconnect routes, encrypted
   synthetic credentials, no secret in public responses, revocation of old
   captured credentials, and callback access without the client contract header.
   No real OAuth provider/token exchange or installed MCP server acceptance claimed.
6. Targeted `git diff --check` on touched tracked entry points completed without
   whitespace errors; only repository LF/CRLF warnings. No full suite, repeated
   unchanged matrix, frontend build, Playwright or independent review was run.

### Exact next work — continue without approval

1. Convert `libraryDeletionPreview` and `deleteLibraryItem` in `library_api.go` to
   bounded/indexed SQL references plus existing authoring/filesystem outboxes.
   `selection_references`, new `capture_references`, and `credential_connections`
   are already maintained. Preserve current selections versus captured history,
   expected-file hashes, skill trash protection and credential pinning semantics.
2. Finish remaining authoring/API gates and scoped production dependencies. Graph
   authoring CRUD largely shares existing SQL project/outbox paths, but inspect
   its Library recovery check before enabling those routes. General history/search/
   events/feed, board/provider reconciliation, analytics reports, webhook HTTP/feed,
   transcription, event/favorite/export/native reconciliation and any remaining
   `a.state` accesses used under SQL authority still need conversion. Favor direct
   integrations and bounded SQL reads, never a hydrated compatibility `diskState`.
3. SQL graph scheduling/runtime is now implemented; do **not** restart that work.
   Remaining graph validation is actual capture/model/Terminal/Fork/Join/reuse,
   callback and UI/manual acceptance. Add narrow checks only for paths changed
   or failures found. Keep scheduler work independent of any UI subscriptions.
4. Complete exclusive SQL `runEngine` startup, recovery/outboxes/auth/workers,
   migration readiness and shutdown/journal retirement. Retire the temporary
   `sqlRouteAvailable` safety gate only after every used production path is converted;
   it is not a permanent unsupported-feature policy. No real migration is authorized.
5. Complete frontend contract-3 acquisition, exact-ID lookup, scoped subscriptions,
   bounded cache/page windows, decimal revisions, new automation/Library/graph pages,
   and visibility rules. Preserve concurrent frontend work. Then run the actual
   isolated normal CLI/frontend and Playwright, followed by the coordinator's two
   independent review/fix rounds. No normal app PID/URL or browser/model success
   exists yet for this refactor. No approval is needed to continue these phases.

### Files in this continuation

New files under `engine/`:
`sqlite_automations_api.go`, `sqlite_automation_recovery.go`,
`sqlite_automation_api_test.go`, `sqlite_graph_tools.go`, `sqlite_graph_reads.go`,
`sqlite_graph_tools_test.go`, `sqlite_graph_transitions.go`,
`sqlite_graph_workspaces.go`, `sqlite_graph_scheduler.go`,
`sqlite_graph_execution.go`, `sqlite_graph_notifications.go`,
`sqlite_graph_execution_test.go`, `sqlite_graph_api.go`,
`sqlite_mcp_management.go`, `sqlite_library_api.go`,
`sqlite_mcp_management_test.go`.

Existing entry points updated: `automations_api.go`, `automation_recovery.go`,
`graph_tools.go`, `graph_runtime.go`, `graph_scheduler.go`, `graph_snapshot.go`,
`graph_workspaces.go`, `graph_terminal.go`, `graph_activity_api.go`, `linked.go`,
`api.go`, `library_api.go`, `mcp_connections.go`, `mcp_oauth.go`, and prior SQL
files `sqlite_api.go`, `sqlite_indexes.go`, `sqlite_domains.go`,
`sqlite_library_references.go`, `sqlite_record_writes.go`. This ledger updated too.

Concurrent unrelated changes were observed during implementation in `product.md`,
`engine/README.md`, `clients/desktop/src/engine.ts`, ProjectDialog and filesystem/
Linux delivery files, in addition to the earlier recorded unrelated work. Preserved
all of them; read current files before overlapping edits in the next continuation.

## Eighth continuation — Library deletion, history/feed, board and remaining HTTP

**Context boundary, same-agent continuation, not review-ready.** Continue the full
implementation without approval between phases. The user requested no routine
memory handoff; this document is the continuation source. No subagents, commits,
installation, real data migration, real provider mutation or unrelated process
shutdown occurred. Main engine startup is still legacy; the frontend and normal
CLI + Vite + Playwright validation remain unfinished.

### Production integration completed here

**Library deletion and authoring**

- SQL deletion preview streams desired and captured reference indexes into a hash,
  retains only the requested 50-reference page, and fingerprints project metadata.
  Agent files are inspected outside SQL transactions. The temporary preparation
  holds only the changed authoring-file write set, not histories or captured runs.
- Deletion rechecks the SQL reference/project fingerprint under the writer lock and
  atomically admits the existing split Library outbox. Project snapshots, expected
  file hashes and intended writes use separate child rows. Filesystem roll-forward
  and skill trash protection reuse `applySQLLibraryFiles`; settlement reuses
  `finishLibraryDeletion`, preserving captured records and pinned active credentials.
- Preview returns `references`, `total`, `hasMore`, `nextOffset`, `revision`, and
  `itemRevision`; `offset`/`limit` do not change the full confirmation fingerprint.
- Agent-save Library recovery gating now reads SQL. Existing project-scoped
  agent/graph authoring handlers were enabled through the SQL route guard after
  confirming their project lookup and graph authoring outbox paths were already
  converted. The existing concurrent filesystem-browser route support was preserved.

**Webhook management and signed ingress**

- SQL create/patch/delete, connection/event pages, exact raw event payload and
  released-event reads now serve the existing HTTP routes. Cursor scope, explicit
  TTL filtering, removal of pending/released rows, quotas and decimal sequences
  remain intact. Exact payload responses stream stored JSON chunks rather than
  decoding all retained payloads.
- Shared ingress validation still enforces raw-body HMAC, single headers, encoding,
  content type, body limits and two ingress slots. SQL admission rechecks current
  connection/project/secret under the normal lock, then uses the existing atomic
  accept/dedupe/quota/pending implementation. Notifications occur after COMMIT.
  Only the signed ingress route is exempted from the client read-contract header;
  it retains its independent authentication/validation boundary.

**General history, search, export and feed**

- General `session_messages_list`, `session_messages_search`, `session_events_read`
  and `feed_read` now have SQL dispatch. Unknown names return an unknown-tool error,
  rather than a temporary unavailable message for implemented capabilities.
- History cursors bind destination/filters and recipient revision, use absolute
  positions, and scan at most 100 candidate metadata records per request. A search
  can return an empty page with `nextCursor`; consumers must continue that cursor.
  Each response contains at most 20 entries and remains within the tool byte budget.
- Oversized messages/events use `encoding=base64-json`, byte offsets and opaque
  continuation cursors. General tool descriptions explain this SQL contract;
  `sinceRevision` is a decimal-string tool argument. Pre-import delta revisions
  request a full reset rather than silently omitting imported events.
- Added a fixed-chunk record reader and streaming JSON-string decoder for literal,
  case-insensitive search over message text/consultation answers. Unicode/escaped
  strings are decoded rune-by-rune, with a bounded KMP matcher/excerpt. No complete
  conversation or giant event text is hydrated to perform this search.
- Session export streams the existing normalized domain representation, including
  split history/operational collections, from one consistent read transaction.
- GitHub feed now derives compact `github_activity_summaries` once during ingress
  or legacy import. Original payloads stay lossless and cold. SQL CTEs evaluate
  explicit run/job/check/suite/deployment links at the requested release watermark,
  retaining ambiguity and TTL/project visibility before paging. SQL selects card
  states/child winners/counts; only selected page records reach Go. A selected
  card's composition hash streams member IDs/relations, not their payload history.
- Feed list, filters/search, grouped members/children, exact details, bookmark
  anchors and private `feed_read` use that same SQL projection. Details decode
  only the selected retained event. Project-name search remains Unicode-aware.
  Original v2 card IDs/cursors and large decimal provider IDs are preserved.

**Activity board**

- Existing config/phase, provider discovery, connection save/delete, item pages,
  sprint choices, refresh, move/cancel and receipt routes now select SQL handlers.
  Public provider I/O, token protection and current-sprint lookup remain outside
  write transactions. Current configuration metadata is loaded as needed; retained
  item/move histories are not hydrated into a `diskState` adapter.
- Maintained `board_item_views` provides phase/sprint filtering before paging;
  import and current-connection refresh/mapping changes update it. Item reads use
  project/snapshot/assignment gates, revision-bound cursors and indexed pending
  move lookup. Board config returns at most 100 unresolved moves with
  `movesHasMore`/`movesNextOffset` (`offset` request parameter).
- Refresh retains the original two-provider-operation slots and owned jobs. A
  complete provider scan updates only that connection atomically, preserves newer
  Azure revisions, rejects cross-connection identity conflicts, retires missing
  items, and reconciles unresolved moves. Partial/failed scans preserve old items
  and mark sync state stale.
- Startup marks interrupted provider writes unconfirmed and old snapshots stale;
  it never replays uncertain writes. Azure assignment migration, revision fences,
  mapping validation, cancellation receipts and post-write authoritative reads are
  retained. Old terminal receipts with unparseable dates are not pruned.

**Additional HTTP/runtime dependencies**

- Transcription settings and upload configuration use SQL metadata plus the bounded
  split vocabulary. The public view never returns its API key. Actual transcription
  network execution still uses the existing implementation and was not exercised.
- Favorite mutation targets an exact indexed completed assistant event and updates
  it transactionally. **It still decodes/rewrites that one event**, like the current
  SQL stream writer; efficient large-field mutation remains a real outstanding task.
- `/api/sessions/{id}/events` now aliases the bounded conversation-scoped SQL
  invalidation stream. Old clients remain rejected by contract negotiation; this
  is not a reintroduction of full-history SSE payloads.

### Scoped checks and evidence

All Go commands used fresh temporary data/project/skill roots, dev tags,
`CGO_ENABLED=0`, `GOTOOLCHAIN=local`, and approved C: scratch TEMP/TMP. No installed
engine data or upstream provider credentials were used.

1. `TestSQLiteLibraryDeletionAndWebhookHTTP` passed, `ok klm/engine 3.479s`.
   Covers real deletion preview/admission/settlement for an MCP session selection,
   signed ingress without client-contract header, dedupe, release, exact payload,
   patch and connection deletion. Initial compile caught duplicate filesystem
   route cases already present from concurrent work; removed only my duplicate.
   New agent-file/skill-trash admission behavior is implemented but not newly probed;
   earlier outbox filesystem recovery evidence remains in prior continuations.
2. `TestSQLiteGeneralHistoryAndStreamingExport` passed initially (`3.484s`), then
   reran only this changed case after adding pre-migration delta reset handling:
   final `ok klm/engine 2.530s`. Covers large Unicode/escaped message continuation,
   literal streamed search, lossless streaming export, and imported revision reset.
3. `TestSQLiteGithubActivityGrouping` passed, `3.365s`. Compares SQL against the
   existing grouping oracle for explicitly linked run/job/check/deployment events,
   provider ID `9007199254740993`, card identity/state/composition/child count,
   member paging, bookmark anchors and grouped details. No whole feed matrix rerun.
4. `TestSQLiteBoardHTTPAndRecovery` passed, `2.917s`. Covers real phase/config and
   indexed item pages, pre-dispatch cancellation receipt, startup conversion of a
   pending write to unconfirmed, and blocked configuration changes. Its intentionally
   invalid protected credential fails before provider HTTP; real GitHub/Azure board
   discovery, refresh, writes and sprints are **not** claimed verified.
5. `TestSQLiteTranscriptionSettingsAndFavorite` passed, `2.852s`. Covers saved SQL
   vocabulary/key configuration without secret disclosure and exact favorite mutation.
6. All new/changed Go files were formatted. Targeted `git diff --check` completed
   without whitespace errors, with existing LF/CRLF warnings only. No full suite,
   unchanged test matrix, frontend build, browser run or review agents were used.

### Precise next work — same implementer, continue without approval

1. **Analytics reporting remains legacy.** `analytics_api.go`/`analytics.go` still
   construct `analyticsSnapshot` from global maps. Implement SQL time/scope queries
   and streaming/aggregate accounting over split spans/usage/reasons, preserving
   coverage/provisional/unknown-time semantics, current-topic membership, counts,
   graph wall time and the existing response. Do not solve it by loading all SQL
   executions/spans/owners into that old snapshot. Projects/topics/distinct graph
   catalog metadata can be acquired separately on demand.
2. Finish bounded large-event stream/favorite mutation and native reconciliation/
   current-turn caches. `sqlite_adapter.go:commitSQLStreamLocked`,
   `sqlite_session_repository.go:eventByID/writeEvent`, terminal normalization and
   several exact-event paths still decode a single record up to the old 256 MiB
   cap. General search/export no longer need full event/history hydration, but this
   does not finish hot-path write amplification. Check any remaining used `a.state`
   paths before enabling normal SQL startup; the temporary route guard is not a
   finished disabled-feature policy. Folder routes are obsolete topic predecessors,
   not a reason to reintroduce old folder behavior.
3. Wire exclusive SQL `runEngine` initialization and shutdown with the current
   concurrent Linux/public-listener changes preserved. Replace legacy load/replay/
   topic writes, roll forward SQL outboxes, run scoped recovery barriers, initialize
   analytics/board/general metadata, start all headless workers, and checkpoint/close
   only after owned writers/workers drain. Update CLI migration readiness beyond its
   old timeout. Keep original backups, version-600 downgrade fence and no fallback.
4. Finish frontend contract-3 acquisition and bounded caches/subscriptions. Include
   new Library deletion/reference pages, exact MCP catalog detail, board move pages,
   decimal revisions, graph windows/content, automation pages and stream invalidations.
   Preserve drafts/Focus/native/mobile behavior and concurrent directory-browser work.
5. Then actual isolated normal CLI + Vite + Playwright, fixes, and coordinator-owned
   review rounds. No app PID/URL, successful model/provider run, or integrated browser
   validation exists yet for the SQLite release. Do not report review-ready before it.

### File ownership in this continuation

New `engine/` files: `sqlite_library_deletion.go`, `sqlite_webhooks_api.go`,
`sqlite_library_webhook_http_test.go`, `sqlite_json_stream.go`,
`sqlite_general_history.go`, `sqlite_general_history_test.go`,
`sqlite_github_activity.go`, `sqlite_github_activity_test.go`, `sqlite_board.go`,
`sqlite_board_api.go`, `sqlite_board_worker.go`, `sqlite_board_sprints.go`,
`sqlite_board_test.go`, `sqlite_transcription.go`, `sqlite_favorites.go`,
`sqlite_settings_test.go`.

Existing entry points edited: `library_api.go`, `authoring.go`, `api.go`,
`webhooks_api.go`, `webhooks.go`, `general_agent_tools.go`, `graph_adapter.go`,
`history.go`, `github_activity_api.go`, `activity_board_api.go`,
`activity_board_worker.go`, `transcription.go`. Existing SQL files edited:
`sqlite_api.go`, `sqlite_indexes.go`, `sqlite_domains.go`,
`sqlite_import_domains.go`, `sqlite_webhooks.go`, `sqlite_general_execution.go`.
No frontend source changes in this continuation. The new derived tables/projections
remain pre-release schema work; all checks create fresh isolated SQL databases.

## Ninth continuation — recovered after user cancellation (in progress)

The previous invocation was cancelled after a server-launch tool response appeared
stuck. **Work was preserved.** Recovery inspected actual OS processes and redirected
logs before assuming command completion. No unrelated process was stopped.

### Recovered implementation and evidence

- `sqlite_analytics.go` now supplies scoped streaming iterators to the existing
  analytics aggregate/response semantics. Distinct counts remain SQL aggregates;
  spans/usage are read in bounded batches, while owned-live clock projections stay
  limited to active handles. Current-topic attribution, coverage reasons, provisional
  timing, graph timing and weekday/DST splitting remain in the shared calculation.
  `TestSQLiteAnalyticsReportSemantics` passed (`2.750s`), comparing SQL and legacy
  report totals/topics/harnesses/graphs/weekdays on the focused fixture.
- `sqlite_event_stream.go` stores text-last JSON with persisted SHA-256 checkpoints
  and small event metadata. Appends rewrite tail chunks only; status/favorite
  metadata changes stream old encoded text into transaction-local staging chunks.
  Stream/finalization/favorite entry points use these paths. Existing imported
  records without checkpoints still take a one-time exact-record fallback; unusually
  large non-text metadata is not cached. `TestSQLiteBoundedEventAppend` passed
  (`2.957s`): a >2 MiB Unicode event append touched fewer than 30 SQL rows, content
  and final status survived metadata rewrite, and staging chunks were retired.
- Normal `runEngine` now exclusively calls `serveSQLiteEngine` after the existing
  data-directory lock and port reservation. It uses the verified importer/cutover,
  SQL outboxes, recovery barriers, auth, general singleton, board/headless workers,
  passive runtime checkpoints and drained shutdown checkpoint/close. Original
  legacy cleanup/load/replay/write startup code was removed. The public route refusal
  gate was removed. No fallback or dual-writing was added.
- Local control is opened during preparation. Windows readiness now checks
  `controlReady`; pairing/revocation are gated until ready. CLI start observes its
  owned child until readiness/exit and prints periodic preparation status instead
  of killing valid migrations after 45 seconds. Current Linux readiness behavior
  was preserved. Engine CORS now explicitly permits `X-KLM-Read-Contract`.
- Frontend work exists in `App.tsx`, `engine.ts`, `clientAuth.ts`, `sessionState.ts`,
  `HistoryLoader.tsx`, new `scopedConnection.ts` and `EventContent.tsx`, plus workspace,
  general dock, automation and Library adapters. Requests negotiate contract 3;
  metadata envelopes/history pages are adapted; revisions compare as BigInts;
  project/topic pages load on selection; visible conversation scopes share streams;
  histories are bounded to 200 events with inactive-history eviction. Unopened native
  children and the minimized general dock no longer request histories. Large event
  content is explicitly paged/downloaded. Topic mutation envelopes were adapted.
- Added `/api/activity/sessions` for bounded live/attention metadata and project
  session `archived=all` cursor scope. These are the latest engine changes in the
  second successful dev build. The frontend conversion is **not yet complete**:
  pagination/cache retention across refresh, remaining surface contracts, large-event
  UI details and actual browser behavior require finishing/validation.
- Recorded commands: the first normal isolated `klm start` returned success;
  a later owned `klm stop` returned success; a subsequent dev build completed with
  no compiler diagnostics; the second `klm start` returned success. The final chained
  TypeScript check produced no diagnostics and proceeded to launch Vite PID `37128`.
  Earlier TypeScript errors (a shadowed metadata helper and a missing `useRef` initial
  argument) were fixed before that final command.

### Actual recovered processes/logs

Verified with bounded OS inspection after cancellation:

- Engine PID **49040**, executable
  `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-dev.exe`, command `__worker`,
  listener `127.0.0.1:17331`. Its scratch log records SQL contract-3 readiness at
  **2026-10-28 12:29:35**. The preceding first startup log records every cutover
  stage, ending in committed authority and readiness at 11:48:06.
- Data root: `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-ui-20261028-c9`.
  Skill root: same Temp parent, `klm-sqlite-ui-20261028-c9-skills`.
  Disposable project: same Temp parent, `klm-sqlite-ui-project-c9`.
  Engine was started with `KLM_DEV_DISABLE_CLIENT_AUTH=1` and loopback public bind.
- Vite PID **37128**, creation **2026-10-28 12:29:46**, command matches our prior
  `vite --host 127.0.0.1 --port 5174 --strictPort` launch. Its redirected log
  `klm-sqlite-vite-c9.log` says ready on 5174; `klm-sqlite-vite-c9-error.log` is empty.
  The user/coordinator flagged 5174 as shared with other work: **leave it untouched**
  and use a fresh explicitly owned frontend port for resumed validation.
- Port 5173 belongs to unrelated node PID **52108**; do not use or stop it.
- No `go.exe`, `compile.exe`, `link.exe`, or `*.test.exe` process was found.
- No Playwright/browser acceptance has yet been recorded. Next: launch a fresh
  redirected background Vite on free **5175**, record exact ownership, probe bounded
  HTTP readiness, and test the actual app with a newly created browser tab.

Continue without approval. Use bounded commands and readiness probes, never an
unbounded SSE terminal request. No real data/install changes, commits or subagents.

### Ninth continuation — resumed integration and actual browser evidence

The cancelled launch was recovered successfully. All subsequent shell commands
used explicit 15/20/60-second bounds. The replacement Vite was launched through
`Win32_Process.Create` with redirected stdout/stderr and stdin from NUL, avoiding
an inherited foreground-server/tool pipe. Its creation returned immediately;
bounded HTTP probes returned engine health contract 3 and frontend HTTP 200.
No Go compiler/linker/test process was left running at the final OS inspection.

**Current owned running application (recheck identity before any control action):**

- Engine **PID 3108**, executable
  `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-dev-c9-bound.exe`,
  `__worker`, loopback **17331**. This is the latest compiled engine, including
  graph activity paging and client database-identity binding. Earlier owned worker
  PIDs 49040, 72092 and 71092 were stopped through their scratch CLI control channel.
- Frontend **http://127.0.0.1:5175/**, node **PID 20148**, launcher cmd **PID 70960**.
  Command contains the absolute repository Vite script and `--port 5175 --strictPort`.
  Logs: approved Temp parent `klm-sqlite-vite-c9-resume.log` and
  `klm-sqlite-vite-c9-resume-error.log`; initial startup logged ready in 545 ms.
- Engine data/skills/project roots remain the isolated paths above. Engine health
  at the final probe: `status=ok`, `contract=3`, zero running sessions/active graphs.
  `state.json` is the version-600 fence, database ID
  `9f7ba91f7d22ba27542b98de5658eb63`, original-backup reference
  `legacy-backups/79792b3f8fe435a1eb5439e099530162/state.json`.
- **Do not touch ports 5173 or 5174** (observed PIDs 52108 and 37128). They were
  neither used nor stopped in the resumed validation. Release/Pi/Tauri/other-task
  processes were not controlled.
- Playwright used a newly created tab (index 1); the preexisting blank tab was not
  used. No unbounded terminal SSE request was made.

**Additional production fixes in the resume:**

- Project session windows retain the number of loaded pages across invalidation,
  cap at four 100-row pages, rotate on further explicit paging, and offer Newest
  sessions. Metadata from unrelated cold projects is evicted except explicit
  visible/recent pins. Exact saved/navigation IDs are fetched independently of
  list membership, rather than opening a different fallback conversation while
  the requested ID is unresolved. Spawn links also resolve missing metadata on click.
- History remains capped at 200 events with six recent inactive history caches.
  Fixed Latest messages to apply its explicit replacement window, cancel stale
  navigation requests and surface recoverable errors.
- Fixed large-content display across UTF-8 boundaries using bounded lookbehind/
  lookahead. Download remains exact raw JSON bytes. Conversation export now saves
  the response Blob directly instead of JSON parsing/re-encoding it in JavaScript.
- Persisted event/header reads use `UseNumber`, preserving opaque large numeric
  values through metadata rewrites. The append test now includes
  `9007199254740993` and checks exact preservation.
- Added manual automation-list, skill-page and deletion-reference navigation;
  unloaded selected skills are no longer falsely labeled missing. Added bounded
  additional permission/question pages. Graph activation/event paging controls and
  graph-content descriptors are connected; node activity defaults to the latest
  50 events with earlier/later navigation. Global activity has explicit page controls
  and does not reinterpret absence from a partial activity page as an idle session.
- Added database-identity-bound navigation/Focus/pending-delivery storage, adopting
  pre-contract preferences only once. Client requests carry `X-KLM-Engine-ID`;
  the server rejects stale identities on application operations while allowing
  bootstrap/auth/independent ingress to recover normally. Bootstrap binds identity
  before dependent activity reads; old caches/UI state and in-flight/queued sends
  are invalidated on a database change. CORS allows the identity header.
- Tightened conversation refresh coalescing: a pending invalidation is not lost
  behind a busy read, fired timers are cleared, and scope reset semantics discard
  potentially stale older cached pages. Ready notices already covered by the loaded
  snapshot do not trigger a redundant read.

**Actual CLI/browser checks (not model-generated work):**

1. Through the live UI, registered disposable project **SQLite Validation** at
   `...\Temp\opencode\klm-sqlite-ui-project-c9` and created **Scoped conversation**.
   Project/session mutation, metadata, project/topic pages, graph projection and
   selected-history HTTP requests returned 200/201. Initial browser console had no
   errors. No user message was sent to a harness.
2. Prepared a fixture only after stopping the verified scratch worker and acquiring
   `engine.lock`. New opt-in `TestPrepareSQLiteBrowserFixture` requires explicit
   scratch directory/project/session environment variables, validates that they
   lie under the explicit TEMP root, and refuses a running engine. It added 260
   small history events, one large Unicode event, 125 extra sessions, and a cold
   project with 20 histories. Preparation completed (`2.530s`), then the normal CLI
   reopened the SQL database successfully. This is fixture setup, not model evidence.
   IDs: project `0ed566547125b43e5721110f80b658d4`, selected conversation
   `a7a7e11203c2ae83bcfac218bbcf5353`, cold project
   `3453534b26074d54c3b15b2051de944d`.
3. Reload restored the exact selected conversation although it was outside the
   first 100-session page. Only that project's pages and the selected conversation's
   history were requested. A filtered network query for the cold project ID returned
   no requests. The oversized event appeared as an on-demand content entry.
4. Four explicit Load earlier actions produced visible message counts
   **99 → 149 → 199 → 200**. Latest messages exposed the explicit return action.
   Its first browser check failed, revealing the replacement-window bug; after the
   fix it returned to 49 small messages plus the large-content entry.
5. Open content / Next part fetched bounded chunks. Initial inspection found two
   UTF-8 replacement characters at a chunk boundary; corrected the decoding window.
   The follow-up check found zero replacement characters and working Previous part.
6. Favorited a real fixture row through the UI, observing Remove from favorites.
   Navigated to the minimized general workspace and captured **no history requests**
   there. Draft text survived navigation back to the project conversation.
7. SQL Analytics rendered in the actual workspace with HTTP 200 and no console
   errors (146 fixture-created sessions, no asserted model activity).
8. Loaded project sessions beyond page one: visible ungrouped rows increased
   **100 → 125**. Created **Browser-created topic** through the UI; after SSE refresh,
   the oldest loaded session and all 125 loaded rows remained, along with the draft.
9. Created a disposable unconnected MCP configuration in Library, then used its
   Usage/Delete preview and removal action. The UI returned to No MCP connections
   with no alert. The dummy endpoint was not connected/tested; no upstream provider
   or real secret was used.
10. A delayed-response check held history responses for 350 ms while three favorite
    writes arrived. An HMR-retained pass failed; after the scheduling fixes and a
    clean page load, the check observed **three delayed history reads and the final
    invalidation applied**. Do not treat the earlier failed pass as success.
11. Latest engine/browser reload returned successful contract-3 endpoints, restored
    History validation, showed 49 messages plus the content entry, and had no alerts
    or console errors. Scoped `:db:` navigation/Focus preference keys were present.

Screenshot: **`.playwright-mcp/klm-sqlite-c9-scoped-history.png`**. The screenshot
tool refused the external Temp destination, so evidence was saved in its allowed
workspace output directory. The screenshot shows the actual SQL fixture app, not
a mock UI. One click after shrinking the viewport timed out because the existing
oversized floating shell put the rail off-screen; using an 1800×1100 viewport made
the normal controls reachable. A contenteditable check was corrected to use
`innerText` instead of `inputValue`. These were bounded tool responses, not another
stuck foreground command.

**Latest lightweight checks:**

- `TestSQLiteBoundedEventAppend` and `TestSQLiteGraphHTTPPages`: passed together
  (`5.221s`) after numeric preservation and graph paging changes.
- `TestSQLiteClientIdentityBinding`: passed (`4.704s`), including stale-identity
  mutation rejection, bootstrap recovery, and matching-identity session access.
- Latest dev binary build completed successfully and was started with the normal
  CLI. Multiple owned stop/start cycles returned explicit success.
- Final `tsc --noEmit -p tsconfig.json` returned **TypeScript check completed
  successfully** after the refresh/identity fixes. No full suite or unchanged
  matrix was run. `git diff --check` found no whitespace errors (LF/CRLF warnings).
- No installed harness turn, real OAuth/provider operation, installer/LAN acceptance
  or independent review is claimed by this browser evidence.

### Ninth-continuation boundary — historical; superseded by tenth continuation

The normal app is now integrated and core browser validation exists. **Do not
restart the storage/domain refactor or revert to the eighth-continuation status.**
Still finish these specific client edges before declaring the full scope complete:

1. Topic headers/options: project session pages are bounded/preserved, but topic
   Load more still appends while refresh replaces the first 200 topics. Make topic
   windows/refresh consistent and ensure loaded/exact sessions whose topic lies
   outside the current topic page retain their correct topic label/archive state.
   `topics.ts:topicGroups/sessionTopic` currently relies on loaded project topics.
   Automation/topic pickers also need access beyond their initial 200-topic page.
2. Library search currently filters the loaded skill window. Manual skill paging
   now exposes all pages, but finish scope-wide search rather than claiming a
   missing match outside the loaded page. Verify page transitions/deletion reference
   controls without rerunning whole domain matrices. Project/global metadata and
   other small UI caches should retain only their intended bounded scopes.
3. Validate the final graph/large pending-request pagination adapters on suitable
   isolated fixtures. General permission/question extras are implemented; projected
   graph-request windows and any legacy over-capacity queue recovery deserve the
   remaining integration check. Keep exact owning-chat reply scope and no implicit
   privilege expansion.
4. Complete release/storage-contract documentation (version sources already show
   desktop/Tauri 0.6.0 and mobile 0.6.0+3; dev engine version is intentionally
   `development`). Then the coordinator can request its two independent review/fix
   rounds. No extra agents or approval gate should be introduced by this implementer.
5. Preserve the running owned app for continuation/manual checks, or stop **only**
   the verified owned worker/Vite when validation is complete. Use the exact scratch
   `KLM_DEV_DATA_DIR` and recheck PIDs/command lines before stopping. Never stop/use
   5173/5174 or unrelated release/Pi/Tauri/browser processes.

No routine memory handoff was written. This ledger records the recovered and newly
verified state for the same implementer at the context boundary.

## Tenth continuation — initial review handoff (2026-10-28)

### Practical completion

- Topic/session metadata now carries exact topic identity, ordering/archive state
  and compact runtime/request counts independently of loaded topic pages. Topic
  windows rotate with First/Next controls; scoped Unicode search applies before
  pagination. Topic dialogs and automation topic/session pickers use remote pages
  and exact saved-ID reads rather than interpreting unloaded options as missing.
- Library search covers configured skill metadata and MCP catalog descriptions
  before paging. Mounted Library selectors use debounced remote search, preserve
  selected off-page entries, and keep explicit skill navigation.
- Additional graph requests preserve the node's owning conversation. Pending
  request/queue/reference controls expose continuation and return navigation.
  Uncertain queued Send now retains its explicit user retry decision.
- Fixed a concrete integration omission found during completion: consultation
  cancellation still read legacy `a.state`. The SQL handler now updates the exact
  consultation and related cards atomically, then cancels matching owned execution
  and returns the normal contract-3 session envelope.
- Updated `engine/README.md` and `product.md`: sole SQLite authority, immutable exact
  originals/manifest, root version-600 fence, no automatic fallback/rollback,
  readiness/recovery, contract-3 pages/scoped SSE and fragment encoding. Existing
  desktop/Tauri/Windows version sources are **0.6.0**, mobile **0.6.0+3**; no bump
  was necessary. Local unversioned engine intentionally reports `development`.

### Recovered command status and affected verification

The cancelled parallel Go/TypeScript invocation returned **user aborted**, with
no completion output. No surviving `go.exe`/`gofmt.exe`/TypeScript process from it
was found. Those checks were run to completion once after recovery:

```text
gofmt -w sqlite_topic_views.go sqlite_api.go sqlite_activity.go sqlite_text.go sqlite_library_api.go sqlite_paging_completion_test.go sqlite_paging_browser_fixture_test.go sqlite_consultation_control.go linked.go
go test -tags dev -run '^(TestSQLiteTopicAndLibraryPaging|TestSQLiteLargeRequestAndGraphPages)$' -count=1 -v .
TestSQLiteTopicAndLibraryPaging PASS (11.06s)
TestSQLiteLargeRequestAndGraphPages PASS (1.15s)
ok klm/engine 14.470s
```

These checks cover off-page topic/archive identity, Unicode topic/session search,
cursor scope, whole-scope Library matching, 205 root/205 graph-node permissions,
35 queued items, owning-node projection, latest 50/125 node events, queue removal
and exact SQL consultation cancellation. They do not execute a native harness.

- `clients/desktop`: `node_modules/.bin/tsc.cmd -b --pretty false` completed with
  no diagnostics.
- `go build -tags dev -o C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-dev-c10.exe .`
  completed successfully (`CGO_ENABLED=0`, `GOTOOLCHAIN=local`, unchanged Go 1.24.2).
- `git -c core.safecrlf=false diff --check`: completed with no output/errors.
- No full test suite, speculative edge-case matrix, review agent, install, commit
  or real user-data migration was performed.

### Actual updated CLI + Playwright evidence

Verified old engine PID **3108** executable/command line before stopping it via
its CLI with the exact scratch `KLM_DEV_DATA_DIR`. With the engine stopped and its
data lock acquired, ran the opt-in `TestPrepareSQLitePagingBrowserFixture`:
**PASS, 10.110s** (fixture body 7.23s). It adds 215 disposable topics, 55 scratch
skills and local graph authoring sources. This is synthetic fixture preparation,
not provider activity. New CLI `start` succeeded and reopened the existing SQL
identity; health reports `status:ok`, contract `3`, zero running sessions/graphs.

On `http://127.0.0.1:5175/` at 1800×1100, actual UI actions verified:

1. Restored **History validation**, still displaying its bounded history and the
   correct off-page **Hidden Καλημέρα target** topic. Topic search `ΚΑΛΗΜΈΡΑ`
   found that current topic beyond the initial page.
2. Next topics / First topics successfully rotated topic windows; all topic
   continuation/search requests returned HTTP 200.
3. Library search **needle** found **review-skill-054**, whose description is
   **Needle skill beyond the first page**; actual SQL API returned HTTP 200.
4. Unsaved automation editor remotely found and selected **History validation**
   as an existing destination, then found and selected the off-page Unicode topic
   in new-session mode. Preview/list/search/detail requests returned HTTP 200.
   Discarded the editor; no automation was saved or executed.
5. Final browser console: **0 errors, 0 warnings**. Returned to the conversation.

Screenshots: `.playwright-mcp/klm-sqlite-c10-library-search.png` and
`.playwright-mcp/klm-sqlite-c10-topic-history.png`; prior core browser evidence and
history/content/cache checks remain in the ninth-continuation section above.
Large request/graph continuation has real SQL HTTP fixture evidence; no new
browser permission-grant/model execution is claimed.

### Owned running fixture / next reviewer

- Baseline remains `64d45ef420ab1296c3f5b9b0dba5d79a27d3c36c`; changes are uncommitted.
  Preserve concurrent client-auth/Pi/Linux/directory-browser work.
- Engine **PID 23284**, executable
  `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-dev-c10.exe __worker`,
  API **http://127.0.0.1:17331**. Old PID3108 was stopped; do not reuse that PID.
- Vite **PID 20148**, parent **70960**, repository
  `clients/desktop/node_modules/vite/bin/vite.js --host 127.0.0.1 --port 5175 --strictPort`,
  frontend **http://127.0.0.1:5175/**; Playwright tab **1**, tab0 left untouched.
- Data: `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-ui-20261028-c9`.
  Skills: same Temp parent, `klm-sqlite-ui-20261028-c9-skills`.
  Project: same Temp parent, `klm-sqlite-ui-project-c9`.
- Scratch auth arrangement: dev-tag build with `KLM_DEV_DISABLE_CLIENT_AUTH=1`,
  `KLM_API_BIND_HOST=127.0.0.1`, absolute scratch data/skills overrides. No real
  credential or external provider configuration was needed.
- Database ID `9f7ba91f7d22ba27542b98de5658eb63`; original backup remains under
  `legacy-backups/79792b3f8fe435a1eb5439e099530162/` in this scratch data directory.
  Fixture project/session IDs remain those recorded above; off-page topic ID is
  `2b4b3c2c8a057e45d8cd3019673aedf6`.
- **Do not touch 5173/5174, installed/release engine or unrelated processes.**
  Reverify executable/PID before stopping the fixture; use the CLI with its exact
  scratch data override. Engine log is `<scratch-data>/engine.log`; Vite logs are
  `...\Temp\opencode\klm-sqlite-vite-c9-resume{,-error}.log`.

**No observed main-flow blocker remains for initial review.** This is not release
acceptance. Real installed-data migration, installer/LAN and actual provider/native
turns (including external OAuth/board and graph execution effects) remain explicit
human checks. Imported events without stream checkpoints/very large non-text
metadata retain the documented bounded exact-record fallback; catalog metadata
costs can grow with catalog size. The preexisting oversized Focus geometry issue
after viewport shrink remains recorded rather than expanded into this task.

Parent can now dispatch **reviewer 1 of exactly 2 total rounds**, returning concrete
fixes to this implementer before reviewer 2. Neither round has started here.

## Round 1 fixes — ready for second/final review (2026-10-28)

Technically confirmed and addressed all three findings in
`SQLITE_0_6_0_REVIEW_1.md`; its **Implementer response** section contains the exact
fixes, checks and browser evidence. No extra review/agent was started.

- **P1:** global/project/open-conversation scopes now multiplex over one physical
  HTTP/1.1 update connection; server still validates scope ownership and returns
  scoped bounded invalidations. Actual UI main + 3 floating chats: **one stream
  for six scopes**; parallel health/detail/idle Stop **HTTP 200 in 4/5/5 ms**.
- **P2 navigation:** automation links resolve unloaded IDs through exact detail
  reads and display lookup errors in Settings. Actual saved **disabled** automation
  linked to unloaded Paged session 000; clicking fetched HTTP200 detail, closed the
  dialog and navigated correctly. Fixture automation
  `19104fd16bb88ccd0b5aa692554db7a9` remains disabled, never executed.
- **P2 graph requests:** floating/mobile `FocusSessionChat` subscribes to its own
  conversation invalidations and coalesces graph reads independently of main
  selection. Real SQL SSE + controlled graph-response browser probe showed new
  permission/question cards in floating A while main B remained selected, and
  preserved the exact node-owned reply URL. No native graph/provider execution or
  Flutter runtime acceptance is claimed; intercepts removed and page reloaded.
- Targeted Go checks: `TestSQLiteMultiplexScopedUpdates` and
  `TestSQLiteHTTPProjectTopicSessionVertical` **PASS, 5.073s** combined. TypeScript
  `tsc.cmd -b --pretty false`, fresh dev build, owned CLI restart and whitespace
  check succeeded. No full suite or unrelated test matrix was run.
- README documents repeated watched scopes/one connection; release stays **0.6.0**.

**Current ownership supersedes the tenth-continuation PIDs:**

- Engine **PID 536 / parent 26572**,
  `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-dev-r1.exe __worker`,
  API **17331**. Identity-verified PID23284 was stopped before replacement.
- Vite unchanged **PID 20148 / parent 70960**, frontend **5175**, Playwright tab1.
  Data/skills/project paths, database ID and immutable-backup paths remain unchanged
  from the initial handoff. Auth remains dev-only bypass + loopback binding.
- Do not touch 5173/5174 or installed/unrelated processes. Reverify current ownership
  before any stop/replacement. Current fixture health: ok, zero native/graph runs.
- Final browser console: **0 errors / 0 warnings**. New screenshot paths and the
  corrected failed setup attempts are recorded in the review-response section.

No known unresolved round-1 finding remains. Parent may dispatch **reviewer 2 of
exactly 2 total rounds**, returning any concrete fixes to this same implementer.
Real migration, installer/LAN and live-provider acceptance remain separate checks.

## Final implementation handoff — both reviews complete (2026-10-28)

### Final review disposition and fresh evidence already obtained

`SQLITE_0_6_0_REVIEW_2.md` is the final independent report: **no material findings,
no changes requested**. It independently confirmed all round-1 fixes:

- Actual Vite module: global + project + four conversation listeners all received
  ready over the multiplexed source; parallel real health/detail returned **200 in
  19/22 ms**, without closing a stream.
- Fresh browser context had **zero** destination detail requests before the disabled
  automation's unloaded-session link; click fetched exact detail (200), closed
  Settings and rendered Paged session 000.
- Real SQL/SSE invalidation refreshed floating Paged session 124 while main stayed
  on Paged session 000. Controlled graph response supplied permission/question;
  both appeared with one graph read. Interception removed/title restored/reloaded;
  no native graph or reply was executed by reviewer 2.
- Reviewer 2 ran:
  ```text
  go test -tags dev -run '^TestSQLite(MigrationReadOnlyReplayEquivalence|MigrationProductionImport|MultiplexScopedUpdates)$' -count=1 -v .
  PASS — ok klm/engine 5.844s
  ```
  Three top-level tests passed, replay includes v1/v2. Synthetic production import
  covered **85 normalized domain tables**, retained-backup fingerprints and SQL
  commit/reopen. This is not real user-data migration.
- Implementer's last affected checks remain valid: two HTTP/scope tests **5.073s**,
  TypeScript build check without diagnostics, fresh dev binary build, owned CLI
  restart and whitespace check. No source changes followed except documentation.
  Earlier domain tests/browser evidence are recorded above, not rerun or summed
  into an invented full-suite result.

Finalization made **documentation-only edits**. No further tests/builds/browser
probes or reviews were run. Process/listener ownership and the scratch marker/
manifest were read to make this handoff accurate.

### Backup/fence precision — correction to earlier fixture shorthand

Production behavior preserves each **existing** source byte-for-byte under
`<data-dir>/legacy-backups/<migration-id>/{state.json,stream.sealed,stream.journal,client-auth.json,engine-access.json}`,
plus `manifest.json` recording presence, byte counts and SHA-256. Captures are
retained throughout 0.6.0, including incomplete generations; no automatic cleanup.
Root `state.json` becomes the rejecting fence, not a live checkpoint or original.
Committed SQL never automatically falls back/reverses authority.

The actual running browser fixture was created from an **empty data directory**.
Its retained manifest is at:

```text
C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-ui-20261028-c9\legacy-backups\79792b3f8fe435a1eb5439e099530162\manifest.json
```

It records **all five original sources as absent**. There is consequently no
original `state.json` backup file in this generation. Earlier statements that an
“original backup remains” refer to the retained generation/manifest; they must not
be read as evidence that this fresh fixture migrated a populated checkpoint.
Byte-exact populated-source evidence comes from the disposable migration tests,
including reviewer 2's production-import rerun.

The fixture's root `state.json` was read during finalization and is exactly:

```json
{"version":600,"storageFormat":"klm-sqlite","databaseId":"9f7ba91f7d22ba27542b98de5658eb63","originalBackup":"legacy-backups/79792b3f8fe435a1eb5439e099530162/state.json"}
```

The canonical originalBackup filename may refer to an absent source; the manifest
is authoritative about which original files actually existed. Restoring old JSON
after SQL changes is not an automatic/lossless downgrade and loses those changes.

### Version/source confirmation

Desktop npm/package-lock root entries, Tauri config/Cargo package and matching
`klm-desktop` Cargo.lock entry, Windows engine build/installer defaults are **0.6.0**;
mobile is **0.6.0+3**. These values were already present, not a newly performed bump.
Release linker sets `main.engineVersion`; current local binary stays `development`.
Go remains 1.24.2, pure-Go SQLite driver **modernc.org/sqlite v1.44.3**, SQL schema
**1**, HTTP read contract **3**, downgrade fence **600** (separate namespaces).

### Honest acceptance boundary / practical limitations

No material unresolved implementation defect was established by the two bounded
reviews. This does **not** fulfill every acceptance measurement in the original
plan. In particular, its 10× unrelated-history steady-state memory / warm-start /
p95 latency diagnostic targets have **not been fully measured**. Existing evidence
is narrower: same **6 SQL rows changed** with 50 vs 500 cold events, indexed/bounded
history pages, actual no-cold-history requests, capped browser windows and transport
response timings. Do not claim the complete performance criteria passed.

Remaining acceptance: real populated installed-data migration, packaged release/
installer/login/LAN behavior, actual Flutter/WebView and live Pi/OpenCode/Codex
turns, external MCP/OAuth/board operations and graph/automation provider effects.
Synthetic recovery/domain checks and UI response probes are not substitutes.
Imported events without stream checkpoints or large non-text metadata may use a
bounded exact-record fallback (existing 256 MiB event ceiling); catalog/report
metadata costs can grow with catalog size. Earlier Focus geometry after viewport
shrink remains a recorded limitation, not expanded scope.

### App left available / manual follow-up

Final read-only ownership check confirmed:

- Engine **PID 536**, parent **26572**,
  `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-dev-r1.exe __worker`,
  listening **127.0.0.1:17331**.
- Vite **PID 20148**, parent **70960**, repository Vite script with
  `--host 127.0.0.1 --port 5175 --strictPort`, listening **127.0.0.1:5175**.
- Open frontend **http://127.0.0.1:5175/**. Same isolated data/skills/project paths,
  dev-only auth bypass/loopback binding and Playwright tab1; no fixture process was
  stopped during finalization. Do not use/stop unrelated 5173/5174 or release tasks.

Manual validation can use History validation for paging/content/export, topic/
Library searches, the disabled Review1 unloaded link and multiple floating chats.
The disabled automation must remain unrun unless the user intentionally chooses
real execution. Native/provider acceptance should use a separately chosen disposable
project/conversation. Before any future real upgrade, obtain explicit authorization
and a complete stopped-engine data/artifact backup; preserve the original upgrade
generation and verify counts/history/native identities after first start.

Baseline remains `64d45ef420ab1296c3f5b9b0dba5d79a27d3c36c`; changes are uncommitted.
Concurrent user/coordinator work preserved. **No real-data migration, installation,
commit/push/reset, backup deletion, subagent or third review was performed.**

## Authorized actual-history copy validation — completed read-only inspection (2026-10-28)

User explicitly authorized briefly stopping/restarting the installed release to
capture consistent actual history and migrating **only its copy**. This supersedes
the earlier no-real-data-copy boundary; the installed source is still never opened
by the new importer or manually edited. Exactly two reviews remain complete.

- Source: `C:\Users\Samuel\AppData\Roaming\klm\engine`; actual installed executable
  `C:\Users\Samuel\AppData\Local\Programs\KLM Engine\klm.exe`. Original PID37084
  was verified before graceful stop. Source was held with a read-only exclusive
  engine-lock handle during capture; all copied file hashes compared against the
  stopped source. Installed CLI restarted in `finally`; current release **PID71624**
  listens on7331. Its API requires pairing (401 without credentials); no credential
  was disclosed. A summary-total calculation failed after all comparisons, so the
  private manifest was reconstructed from the stable copy after source restart.
  Stop/copy/restart duration was not reliably retained and is not claimed.
- Copy: `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-actual-history-20261028-copy1`.
  **323 files / 1,187,670,728 bytes**: checkpoint plus present auth file, native
  sessions, capability references and workspaces. Sealed/active journal and access
  settings were absent after graceful stop and recorded as absent. Excluded logs,
  locks, ephemeral `.opencode-*`/`.pi-*`, empty worktrees and external project/native
  home trees; no user project files were copied or written.
- Captured checkpoint **1,113,468,793 bytes**, SHA-256
  `314cf8d41741b547e67cc914f5599c38354e90337e4a4f27dfcb97230041ddd8`.
  Private per-file evidence: `<copy>/copy-capture-manifest.json`.
- Preflight measured about9.04GB free and3.13million lexical JSON nodes, with an
  estimated7.6GB peak extra working set on disk. Free space independently increased
  to13.78GB by launch; no unrelated cleanup performed. Temporary Windows/dev-only
  `sqlite_actual_copy_probe_test.go` compiles a standalone opt-in inspection helper;
  **no product feature/transport/review was added**. It calls the production importer
  offline, starts no recovery/outboxes/workers/harnesses and cancels below1.5GiB
  free. Its later serve mode is GET/HEAD/OPTIONS-only, loopback17331, provider/model/
  quota discovery disabled. No copied automation/provider action can execute there.
- Existing importer **PID30420**, parent54880,
  `...\Temp\opencode\klm-actual-copy-inspector.exe -test.run=^TestInspectActualHistoryCopy$ -test.v -test.timeout=0`.
  Logs `...\Temp\opencode\klm-actual-copy-import{,-error}.log`. User interruption
  did not kill it; resumed this same process, never recopied/restarted the import.
  At ~574s, process remained active, ~48MB resident, small reported Go heap,
  ~9.59GB free; staging SQL read showed15 projects/576 sessions/139,940 events.
  These are progress counts, **not a completed migration result**.
- Production backup generation in this actual copy:
  `<copy>/legacy-backups/687efb281dd3ced5377026a24bc8263c/`. Its retained state SHA256
  independently matched the captured checkpoint. Source credentials remain private.
- Synthetic fixture remains intact and serving: enginePID536/17331, VitePID20148/5175.
  Do not replace that listener until offline import succeeds. Other5173/5174 and
  installed release remain untouched after the authorized capture/restart.

### Completion evidence (supersedes the progress snapshots above)

- Existing import completed successfully in **874.60 seconds** (14 min 35 s).
  Production verification covered 85 domains: **15 projects, 576 sessions, 139,940
  events, 545 native mappings, 22 graph runs, 2 webhook connections, 211 webhook
  events**, and zero automation definitions/runs. Actual results are saved privately
  in `<copy>/copy-import-result.json`; import stderr is empty.
- Independently rehashed the retained backup checkpoint after completion. SHA-256
  exactly matches the capture/migration manifests above. Original installed data
  remains under its original executable, PID71624/7331, independently reverified.
- SQL file measured **2,774,167,552 bytes**. Actual migration has substantial one-time
  time/storage cost; do not infer fast migration from the small synthetic fixtures.
  C: had about10.7GB free after completion.
- Synthetic enginePID536 was stopped only after successful offline import. Its
  fixture directory remains intact. Copy inspector **PID50076**, same opt-in helper
  executable, now owns loopback17331. VitePID20148 still serves **5175**. Verify
  ownership again before touching processes. Normal source and unrelated ports
  remain unchanged.
- Inspector reopened committed SQL in serve mode in **0.065s** and confirmed the
  same database ID/counts. This is a restricted read-only server, not a timing claim
  for full normal startup/recovery. No workers/harnesses/provider actions run there;
  imported running statuses can reflect capture-time work rather than live work.
- Parent independently opened **http://127.0.0.1:5175/** using Playwright. KLM /
  Publicar KLM rendered; Load earlier messages issued `history?before=885` (200).
  Switching to Cubiq rendered its Playwright CLI conversation. Project/topic/session
  and both conversations' history requests returned200. Screenshot:
  `.playwright-mcp/klm-sqlite-real-history-copy.png`.
- Do not claim a clean console for this special inspector: its wrapper deliberately
  blocks model/quota discovery, currently surfacing as CORS failures. One stale
  identity SSE request returned409 before successful reconnect. History browsing
  worked; writes are blocked. Live native/provider execution remains untested.
- No third review, installed-source migration, installation or commit occurred.

### Full normal CLI startup measurement on the migrated copy

At the user's request, replaced the read-only inspector temporarily with the
normal dev engine executable, using this same 2,774,167,552-byte migrated database.
Read-only SQL preflight confirmed no authoring/Library outboxes, automation work,
board connections, pending webhook records, eligible graph runs/notifications or
consultation deliveries. Both queued messages were already paused. No scheduler
or discovery bypass was enabled for the normal engine measurement.

- First normal start after migration: **478.4 ms**, measured from process launch
  through CLI completion; CLI reported ready and HTTP health returned200.
- Subsequent normal **`klm stop`: 53.5 ms**, exit0.
- Normal **`klm start` after that stop: 189.5 ms**, exit0, measured through the
  public CLI's ready confirmation. Immediate HTTP health returned200 with zero
  running sessions/graphs.
- First normal worker used about22MB resident at the sampled point. No whole
  history decode was required. These are local development-build measurements
  with warm Windows/filesystem caches, not cold-boot benchmarks or an old-vs-new
  controlled comparison.
- Gracefully stopped the measured worker and restored the read-only inspector
  (PID25000) on17331, Vite unchanged on5175. Installed release PID71624/7331 was
  never stopped or reconfigured during this measurement. No extra review/tests.
