# SQLite 0.6.0 — independent review 1

**Verdict: changes requested.** One high-priority browser workflow blocker and two
ordinary navigation/request regressions. This is round **1 of exactly 2**; no
delegation or additional review round was performed.

Baseline: `64d45ef420ab1296c3f5b9b0dba5d79a27d3c36c`. Reviewed the current uncommitted
SQLite implementation, including new/untracked SQL files. Concurrent Linux,
directory-browser, Pi and auth work was treated as unrelated unless directly
involved in SQL persistence. No implementation files were changed.

## 1. P1 — per-scope streams exhaust browser connections and block API controls

**Location:** `clients/desktop/src/features/chat/scopedConnection.ts:10–23`,
with callers in `clients/desktop/src/App.tsx:482,520,607–650`.

Every distinct global, project or conversation scope opens a separate long-lived
HTTP stream. The standard local engine serves HTTP/1.1. A project view with its
main conversation and three other floating conversations therefore holds six
connections to the same engine: global + project + four conversations. Chromium's
per-origin connection pool is exhausted; history/detail reads and mutations,
including Send/Stop, queue behind streams that do not finish. Reference counting
only deduplicates identical scopes and does not prevent this.

**Actual targeted reproduction:** after verifying fixture PID/executable and ports,
used a disposable Playwright browser context against `127.0.0.1:17331`. Opened the
actual global, project and four existing conversation SSE scopes, with contract 3.
All **six received `ready`**. A subsequent `/api/health` fetch remained pending for
the entire 1,000 ms observation; closing one stream released it immediately
(HTTP 200 at **1,017 ms** total). All probe streams/context were closed. This was a
transport reproduction with real endpoints, not a UI-click or native-agent test.

**Minimal suggested fix:** multiplex the bounded set of requested scopes over one
physical update connection per client, preserving scope ownership/reference
counting and scoped invalidations. Do not require HTTP/2 or hide the failure by
limiting users to fewer floating chats.

## 2. P2 — automation session/history links silently fail for unloaded sessions

**Location:** `clients/desktop/src/App.tsx:1196`.

The Project Settings `onOpenSession` callback still only searches
`engine.sessions`, and does nothing when the ID is absent. That collection is now
a sparse, bounded cache. Both the linked-session action and run-history buttons
call this callback (`features/automations/AutomationsSettings.tsx:104,144`). The
new exact-ID resolver used by spawn cards does not cover this path.

**Normal reproduction / inspection evidence:** in a project with more than the
first loaded session page, open Settings → Automations → an automation with an
older, unloaded destination/run. Click its linked-session or history button. The
callback's `find` returns undefined, so no detail request or navigation occurs and
the settings dialog remains open. This follows directly from the callback; it was
not exercised against a saved automation in the running fixture.

**Minimal suggested fix:** resolve a cache miss through
`GET /api/sessions/{id}`, merge the result and invoke the existing project
navigation. Surface a real lookup failure instead of silently ignoring the click.

## 3. P2 — floating/mobile chats do not refresh graph permission/question state

**Location:** `clients/desktop/src/features/chat/scopedConnection.ts:29–33` and
`clients/desktop/src/App.tsx:637–649`;
`features/workspace/FocusSessionChat.tsx:37–45` is the affected consumer.

Conversation invalidations now refresh only operational detail and history.
`respondSQLSession` does not populate the graph projection
(`engine/sqlite_api.go:363–437`). The separate live graph refresh in
`App.tsx:574–593` only targets `activeId`. A floating/mobile `FocusSessionChat`
fetches its graph once on mount/session change; its stable `onGraph` callback does
not cause another fetch on subsequent conversation updates.

**Normal reproduction / inspection evidence:** open conversation A in a floating
window, then select B in the main window. If A's graph subsequently starts or
raises a node permission/question, its floating chat receives detail/history
updates but retains the old graph projection. The new request card never appears,
leaving graph work waiting until A is reopened/selected in the main view. A mobile
sheet whose session differs from `activeId` has the same acquisition gap. This is
a traced production call-path finding, not a live provider/graph reproduction.

**Minimal suggested fix:** refresh graph projection on conversation invalidation
for each open top-level chat consumer, using the shared scope subscription. Keep
the exact node-owned reply routes and bounded request paging.

## Review and verification performed

- Read `AGENTS.md`, `product.md`, the full implementation plan and latest ninth/
  tenth continuation evidence; inspected the diff against the specified baseline.
- Inspected migration capture/replay/normalization, prepared publication/fence/
  authority decisions, normalized record transactions, SQL startup/recovery,
  session execution/queue/permissions, background graph/automation/webhook paths,
  scoped reads/updates and client cache/navigation integration. Read selected
  focused tests as supporting code, **without treating inspection as test execution**.
- Reverified engine **PID 23284**, exact `klm-sqlite-dev-c10.exe __worker` path,
  API **17331**, Vite **PID 20148 / parent 70960**, and UI **5175** before probing.
- Performed only the targeted read-only SSE connection-pool reproduction above.
  An earlier Vite-module-based probe timed out without a usable result; its
  disposable context was gone before completion, and final inspection showed only
  the original two browser tabs/context. No process was stopped or restarted.
- No Go/TypeScript checks or full suites were rerun. No installed/real dev data,
  backup contents, credentials or provider operations were touched.

No material migration data-loss or security regression was established by this
inspection. That is not proof of real-data migration, crash durability or native
harness correctness. Existing documented focused-test/Playwright evidence remains
implementer evidence; real migration, installed/LAN and live-provider acceptance
remain separate human checks.

## Implementer response — round 1 fixes (2026-10-28)

All three findings were technically confirmed against the current production
paths and addressed. This response is implementation evidence, **not a second
review or a change to the independent review's original verdict**.

### 1. P1 connection exhaustion — addressed

- `scopedConnection.ts` retains scope/listener reference counting but now opens
  **one physical connection** for all mounted scopes. Scope changes batch into a
  reconnect; ready notices resynchronize reads. Retry/backoff and HMR cleanup
  belong to this one source.
- `sqlite_updates.go` validates up to 64 repeated `watch=kind/id` scopes before
  subscribing. `sqlite_change_bus.go` keeps one bounded pending frame per physical
  subscriber and routes invalidations only to affected watched scopes. Slow-reader
  reset covers the watched set. Existing single-scope routes remain compatible.
- New `TestSQLiteMultiplexScopedUpdates` passed; it checks real HTTP stream
  readiness and committed global/project/conversation routing while excluding an
  unrelated watched conversation. Existing `TestSQLiteHTTPProjectTopicSessionVertical`
  also passed, including the original single-scope path. Combined: **5.073s**.
- Actual UI clicks opened **three floating chats plus a different main chat**.
  The resulting stream watched global + project + four conversations on **one
  physical request**. While it stayed open, parallel real requests completed:
  health **200/4 ms**, session detail **200/5 ms**, idle Stop **200/5 ms**. No stream
  had to close to release them. No native/model message was sent.

### 2. P2 unloaded automation link — addressed

- The Project Settings callback in `App.tsx` now resolves a cache miss through
  exact `GET /api/sessions/{id}`, then uses normal merge/project navigation.
  `ProjectSettingsDialog.tsx` catches lookup errors and displays them in the dialog.
  Linked-session and history buttons share this callback.
- Created a **disabled** scratch automation `Review1 unloaded link` pointing at
  **Paged session 000**, outside the initial session page. Its detail action read
  **Open linked session**, demonstrating the cache miss. Clicking it issued exact
  detail reads (HTTP 200), closed Settings and selected **Paged session 000**.
  No automation was activated/run. This saved disabled fixture remains available
  for reviewer 2: automation `19104fd16bb88ccd0b5aa692554db7a9`, destination
  `573d3002398f4b35d24f8944dc35fe09`.

### 3. P2 floating/mobile graph refresh — addressed

- Shared `FocusSessionChat.tsx` now subscribes to its own conversation's notices
  and coalesces graph refreshes, including one follow-up when a notice arrives
  during a read. This is independent of `App.activeId`, uses the multiplexed
  connection, and applies equally to floating chats and the mobile-sheet consumer.
  Permission/question cards retain their exact graph node's reply routes.
- Browser probe: main selection remained **Paged session 121** while a different
  floating **Paged session 124** received a **real SQL mutation/SSE invalidation**.
  A controlled graph HTTP response introduced both a permission and a question;
  both appeared without selecting/reopening that floating session. Deny used
  `/api/sessions/ui-r1-node-owner/permissions/ui-r1-permission` (intercepted, not a
  real permission), then refresh removed the cards. **Two graph reads** observed.
  All response interceptions were removed, the temporary title restored, and the
  page reloaded. This proves UI acquisition/owner routing; it is **not** live
  provider/graph execution or actual Flutter/WebView acceptance.

### Checks, browser evidence and ownership

- `node_modules/.bin/tsc.cmd -b --pretty false`: completed without diagnostics.
- Fresh `go build -tags dev` to `...\Temp\opencode\klm-sqlite-dev-r1.exe`: succeeded;
  owned CLI stop/start and health succeeded. Go 1.24.2, `CGO_ENABLED=0`, unchanged
  dependency/toolchain versions. `git diff --check`: no whitespace errors.
- Final browser console **0 errors / 0 warnings**. Screenshots:
  `.playwright-mcp/klm-sqlite-r1-multiplex-four-chats.png`,
  `.playwright-mcp/klm-sqlite-r1-floating-graph-refresh.png`,
  `.playwright-mcp/klm-sqlite-r1-unloaded-automation-link.png`.
- Some browser setup attempts failed: saved-layout injection did not open the
  windows, overlapped/background windows blocked clicks, and one automation name
  selector matched multiple buttons. Normal UI opening/minimize/restore and an
  exact selector produced the successful evidence above. Failed attempts are not
  counted as passes; no application fix outside these findings was added.
- Current engine **PID 536**, parent **26572**,
  `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-sqlite-dev-r1.exe __worker`,
  **http://127.0.0.1:17331**. Prior PID23284 was identity-verified and stopped.
- Vite unchanged: **PID 20148 / parent 70960**, repository Vite script,
  **http://127.0.0.1:5175/**. Playwright tab1; tab0 untouched.
- Same absolute scratch data/skills/project roots as the initial handoff:
  `...\Temp\opencode\klm-sqlite-ui-20261028-c9`,
  `...\Temp\opencode\klm-sqlite-ui-20261028-c9-skills`,
  `...\Temp\opencode\klm-sqlite-ui-project-c9`.
  Dev-only open auth and loopback binding remain; original retained backups unchanged.

Ready for parent-owned **review 2 of exactly 2 total rounds**. No additional
review/agent, full suite, real-data migration, installation or commit was performed.
