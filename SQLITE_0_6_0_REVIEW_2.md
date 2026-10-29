# SQLite 0.6.0 — independent review 2 (final)

**Verdict: no material findings in this bounded review; no changes requested.**
The three round-1 findings are addressed. No additional material data-loss,
migration, execution or ordinary-workflow regression was established. This is
round **2 of exactly 2 total**, not universal verification or release acceptance.

Baseline: `64d45ef420ab1296c3f5b9b0dba5d79a27d3c36c`; current uncommitted worktree,
including new/untracked SQLite sources. Read current `AGENTS.md`, full `product.md`,
the implementation plan's storage/migration requirements, latest implementation
progress and `SQLITE_0_6_0_REVIEW_1.md`, including the implementer response.
Concurrent Linux, directory-browser, Pi and auth changes were preserved and were
not reported as SQLite defects.

## Prioritized findings

**None at the requested high/material threshold.** No cosmetic, minor or
speculative edge-case blockers are included.

## Round-1 disposition

### 1. P1 — HTTP/1.1 connection exhaustion: addressed

- **Code inspection:** `clients/desktop/src/features/chat/scopedConnection.ts:6–52`
  owns one physical source for the complete reference-counted scope set, closes it
  before replacement, batches scope changes and cleans up on HMR. Repeated `watch`
  scopes are validated by `engine/sqlite_updates.go:12–90`; the change bus routes
  bounded notices to matching scopes and resets slow subscribers.
- **Independent runtime evidence:** in a disposable browser context, subscribed
  through the actual Vite-served module to global + project + four existing
  conversations. All six scope listeners received `ready`. With them subscribed,
  parallel real health and session-detail requests returned **200 in 19/22 ms**;
  no stream needed to close to release those reads. Subscriptions/context were
  cleaned up. The focused multiplex HTTP test also passed, including committed
  invalidation routing and exclusion of an unrelated watched conversation.

### 2. P2 — unloaded automation navigation: addressed

- **Code inspection:** `clients/desktop/src/App.tsx:1196–1199` resolves a cache miss
  with exact session detail before ordinary navigation. `ProjectSettingsDialog.tsx`
  catches lookup failures and displays them. Destination and run-history actions
  share this callback.
- **Independent UI evidence:** a fresh disposable context opened Settings →
  Automations → the existing disabled **Review1 unloaded link** fixture. Before
  clicking **Open linked session**, there had been **zero** detail requests for
  destination `573d3002398f4b35d24f8944dc35fe09`. Clicking issued exact detail reads
  (**200**), closed Settings and rendered **Paged session 000** as the main heading.
  The automation was neither activated nor run. Run-history navigation was checked
  through the shared callback, not a separate runtime history fixture.

### 3. P2 — floating/mobile graph-request refresh: addressed

- **Code inspection:** `FocusSessionChat.tsx:38–56` subscribes to its own conversation,
  coalesces graph reads and schedules a follow-up if a notice arrives during a read.
  It is independent of main selection and is shared by floating/mobile consumers.
  Cards at lines 71–81 retain the graph node's owning session/request routes.
- **Independent browser evidence:** while main remained **Paged session 000**, a
  reversible scratch-title PATCH on floating **Paged session 124** returned **200**
  and generated a real SQL/SSE invalidation. A controlled graph HTTP response supplied
  one permission and one question; **both appeared without selecting/reopening the
  floating chat**, with one graph read observed. The route interception was removed,
  the original title restored successfully and the page reloaded.
- This establishes UI acquisition with real invalidation and controlled response
  data, **not live graph/provider execution**. Exact reply ownership was inspected;
  no permission/question reply was sent. Actual Flutter/WebView was not exercised.

## Other inspection and actual checks

- Inspected normalized root/nested-domain mappings, record/chunk writes and related
  transaction publication; production import/verification; exact backup capture;
  sealed-before-active read-only journal replay and positional history rebinding;
  prepared publication, version-600 fence and committed SQL authority/no fallback.
  Normal startup reserves the port under the data lock and uses SQL recovery before
  readiness. No automatic retained-backup deletion path was found in these paths.
- Sampled message admission, queue/steering/Stop/finalization and adapter persistence,
  graph execution/recovery/notifications, automation admission and independent
  webhook/board workers. Inspected compact bootstrap, bounded session/history windows,
  cache eviction, exact navigation and scoped client acquisition. This was targeted
  call-path inspection, not exhaustive execution of every domain or endpoint.
- Reverified **engine PID 536**, reported parent **26572**, exact executable
  `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-dev-r1.exe __worker`,
  listening on **127.0.0.1:17331**; **Vite PID 20148 / parent 70960**, repository Vite
  command and **127.0.0.1:5175**, before browser probes. Health reported zero running
  sessions/graphs. Used the supplied isolated dev fixture; no process was stopped.
- Ran only this focused command from `engine`, with `CGO_ENABLED=0`,
  `GOTOOLCHAIN=local`, and approved temporary `TEMP`/`TMP`:

  ```text
  go test -tags dev -run '^TestSQLite(MigrationReadOnlyReplayEquivalence|MigrationProductionImport|MultiplexScopedUpdates)$' -count=1 -v .
  PASS — ok klm/engine 5.844s
  ```

  All **three top-level tests** passed; replay covered v1/v2 subtests. The synthetic
  production importer reported **85 normalized domain tables**, committed/reopened,
  and checked retained backup fingerprints. These are disposable fixture results,
  not migration of installed or real user data.
- Two automation-probe attempts timed out on incorrect accessible-name assumptions;
  inspecting the actual accessible controls/heading produced the successful check
  above. Those attempts are not counted as passes. Disposable contexts were closed.
  Final current-page console query reported **0 errors / 0 warnings**.
- No full suite, review agents, implementation fixes, provider mutations, installation,
  commit or additional review round. Only this report was added to the repository.

Real installed-data migration, installer/LAN acceptance and live native/provider
execution remain separate human checks. The absence of material findings here does
not establish all crash/durability scenarios or all application workflows.
