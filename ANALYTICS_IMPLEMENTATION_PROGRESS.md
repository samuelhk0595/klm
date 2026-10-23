# Engine-backed Analytics — implementation handoff

## Status

Implemented in the shared registered project directory on 2026-10-23 using
OpenCode / GPT-6.1 Sol. The user-approved visual layout is retained. This is an
implementation handoff whose **originating-agent source review and R1–R3 correction
re-review are complete**. Ready for human validation using updated development
engine/frontend builds; runtime acceptance is still pending. Compilation does not
establish runtime accounting accuracy or feature acceptance.

Originating source-review corrections R1–R3 are implemented below, with targeted
in-memory evidence and a new dev build. Re-review is still pending before manual
validation.

Read the entire `ANALYTICS_IMPLEMENTATION_PLAN.md`, `AGENTS.md`, `product.md` and
`ANALYTICS_VISUAL_PROTOTYPE.md` before editing, and inspected the relevant source
and initial working-tree status. Preserved unrelated existing changes. No commits,
pushes, new agents/sessions/subagents/graphs, full suites, complex/integration tests,
live harness probes or installed engine/installer lifecycle operations were run.

## Implemented paths

### Durable collection and native boundaries

- Versioned `diskState.analytics` initializes after journal replay and is saved
  before serving. Collection starts prospectively; old `Session.Usage` is never
  converted into spend, and old event histories are never mined for time/tokens.
- Shared `execute` reserves one durable execution before preparation/runtime side
  effects. Main, side, general and graph-node agents use that path. Ordinary owner
  links are explicit; node records retain graph run, activation and graph turn IDs.
- Persisted dispatch-attempt evidence is separate from confirmed acceptance.
  Preparation failures do not become used sessions or executed agents. Unconfirmed
  submissions expose incomplete coverage rather than measured work.
  Checkpoints never read a clock for unaccepted reservations. A failure before
  dispatch settles without an execution/usage uncertainty tail; unconfirmed
  dispatch uncertainty begins at dispatch, not at preparation/reservation.
- Pi starts on observed `agent_start` and ends on `agent_settled`. OpenCode starts
  from the successful correlated `prompt_async` submission's observed dispatch;
  confirmed idle caps accounting at the idle signal, excluding final reconciliation.
  Codex starts on successful correlated `turn/start` and ends on matching
  `turn/completed`. Steering remains in the same execution.
- Owned error/cancellation termination settles accounting before later stream/MCP/
  graph cleanup. Failed/unconfirmed termination closes at the last observation with
  an unknown tail. Retained runtime idle is not agent effort.
- Fifteen-second liveness checkpoints clock accepted open work and owned runs only.
  Startup closes accepted observations at their durable checkpoint, clears waits,
  marks the tail unknown, and never charges the offline interval. Unaccepted records
  retain the preparation-versus-dispatch distinction. Unknown-tail upper bounds carry
  coverage uncertainty across dates without adding duration. Live projections are
  provisional and never replayed. Monotonic anchors guard elapsed observations;
  a discontinuity freezes the affected turn's timing rather than inflating it.

### Waits and native children

- Requests retain an internal hashed native origin, distinct from their projected
  UI conversation. Known child requests never pause the root just because the card
  appears there. Unresolved origins produce partial coverage.
- Proven serial Pi permission/question waits and Pi gateway elicitation use a
  union of wait keys. Callback success, dismissal, denial, timeout, cancellation and
  settlement close keys; failed callbacks keep the wait open. Automatic approvals
  do not create waits. YOLO resolution reuses the wrapped callbacks.
- **OpenCode/Codex root call ownership is not proof that the entire native agent
  is exclusively blocked.** Their pending root-request intervals are classified as
  unknown, rather than fabricated as working or fully blocked time. Overlapping
  waits are unioned; resulting working/blocked sums are partial. This deliberately
  withholds time where the adapters lack an exclusive-blocked signal.
- Linked consultations account only for the actual bounded tool wait (at most five
  seconds), not the remote consultation's later lifetime. Reciprocal immediate
  yield does not create a wait. Nonserial paths use the same unknown-span policy.
- OpenCode root children get durable, replay-safe invocation evidence only when a
  live child assistant message can be associated with its native task key. Counts
  use stable hidden-session identities across invocations. Task pending/running
  wall-clock alone does not prove child execution duration.
- **Native child time and token accounting remains unavailable.** Live child usage
  maps are not an authoritative complete history and are not added to the ledger.
  Missing pre-metadata messages and side/node/nested paths remain uncovered.
  Pi arbitrary native children are unsupported; managed Codex delegation remains
  disabled. No capability, delegation or security setting was expanded for stats.

### Atomic usage

- Root cumulative samples use the existing cache/reasoning normalization. Input and
  output maintain independent baselines and completeness, with nullable components.
- The native epoch is a hash of engine session, harness and native identity (plus
  Pi's normalized session file). No native path or raw native frame is stored in
  the accounting ledger or returned by the API.
- Pi has a bounded optional two-second pre-prompt stats reconciliation and a bounded
  final read. OpenCode uses the complete pre-prompt root-history walk as baseline,
  retains existing message/step precedence, and uses final reconciliation.
- A new Codex thread can begin at known zero. **Resumed Codex's first valid total
  establishes a baseline; its prior lifetime total is not charged.** Subsequent
  observed deltas are partial. Codex final token completeness is conservatively
  withheld: the existing stream does not prove a final cumulative sample for the
  whole engine turn. Missing first/final samples remain incomplete.
- Baseline/counter changes, execution sequence/delta and `Session.Usage` share one
  small fsynced stream-journal frame. Replay uses ordered revisions and checked
  sequences, not another increment. Equal snapshots add no spend; decreases are
  diagnosed and disabled until authoritative boundary reconciliation, never reset
  into another lifetime spend. Nil does not clear a trusted high-water count.
- Accepted deltas belong to engine receipt time; they are not spread proportionally
  across execution duration. Gaps between turns/resumes stay unallocated. Arithmetic
  exceeding the existing JSON-safe integer limit becomes unavailable.
- The hot usage path does not call a full-state transaction/checkpoint or clone
  conversation histories. Boundaries/liveness use the existing transaction journal.
  Analytics-only stream frames are replayable; current liveness uses transactions.
- Lifecycle transactions share the immutable published Analytics ledger and stage
  only touched execution/run records through dedicated accessors. Mutable timing,
  reasons and wait keys are private to the touched row; retained usage arrays stay
  shared/read-only. Existing-row WAL patches address changed fields (including a
  span's end), never the entire ledger or unchanged usage payloads. Staged changes
  publish only after the same lifecycle frame is fsynced. Counter/delta/Session.Usage
  continues to use its existing atomic stream mutation. Legacy whole-object frames
  and one-time nil initialization remain replayable.
- Evidence is retained without rollups, purge or downsampling. A single execution
  is checked at 100,000 spans, 100,000 usage deltas and 1,024 open wait keys; exhaustion
  follows the storage-failure path rather than silently dropping evidence. General
  long-term retention and ledger growth policy remain deferred.

### Compact API and aggregation

- `GET /api/analytics` supports `period=week|all`, optional Monday `weekStart`,
  `projectId=all|<id>`, IANA `timeZone`, `compare=previous|none`. Invalid/duplicate/
  unknown query parameters and future weeks return 400; explicit missing/removed
  projects return 404; unavailable accounting storage returns 503. Reads invoke no
  harness, graph, selection or mutation. Go embeds `time/tzdata` for Windows.
- One locked compact snapshot supplies one `asOf`/revision for all surfaces;
  aggregation runs after unlocking. It copies accounting/inventory metadata, not
  native history, conversation events or graph source text. Responses are no-store.
- All projects includes nonremoved registered projects and archived history. Removal
  hides retained identity/evidence; restoration reveals it. General-agent work is
  collected with its own owner but excluded from this project dashboard.
- Created sessions come from retained ordinary-session inventory, including unused
  sessions. Used sessions, executed agents, observed native children, graph IDs and
  run IDs are distinct period-wide identity sets, not sums of daily cardinalities.
- Topics use current ordinary-owner membership, project-scoped collision-safe keys
  and a real per-project Ungrouped bucket. Equal names stay separate. Side/child/node
  work follows the owner; moving/renaming a topic does not rewrite the ledger.
- Retained valid graph run facts supply historical counts/durations. New runs get
  lifecycle-atomic timing plus checkpoints through the transaction hook. Interrupted
  recovery timestamps are uncertainty bounds, not measured duration. Graph groups
  use project + captured GraphID and deterministic latest captured names; slug
  delete/recreate is not a stronger definition identity. Graph wall-clock is never
  added to node/agent effort. Average run is in-period observed wall-clock per timed
  contributing run, withheld when relevant timing is incomplete.
- Monday local calendar boundaries and half-open intervals use AddDate, including
  DST days. Current comparison ends at the same local weekday/time one week earlier;
  Go timezone rules resolve ambiguous/nonexistent local times, not a fixed UTC offset.
  The calendar/selector always covers 13 weeks. All time starts at retained evidence
  and coalesces consecutive week-aligned ranges to at most 13 chart buckets.
- Coverage distinguishes complete observed scope, partial lower-bound sums,
  unavailable and not applicable. Independent token components propagate to totals;
  no missing component is substituted with zero. Empty covered buckets can be zero.
- All-role totals consistently mark working time, identified waits and input/output/
  total tokens partial when native-child coverage is unsupported (including Pi,
  OpenCode side/node/nested and missed pre-metadata evidence). Known observed root
  values are preserved; absent child evidence never implies complete coverage.
  Known disabled managed Codex delegation remains a separate not-applicable case.

### Frontend and Home

- All Overview KPIs, activity chart/metric switch, topic/harness rankings, weekdays,
  time breakdown, calendar, Topics table and Graphs table use the real snapshot.
- The existing shared `request`/endpoint resolution is reused unchanged. Requests
  cancel on scope replacement/hide; obsolete generations cannot publish into a new
  scope. Refresh is approximately 15 seconds only while the Analytics panel/document
  is visible, plus focus/visibility return. Same-query errors visibly mark retained
  data as the last snapshot; new-query errors never show another scope's values.
- Missing values show `—`; Live cues remain accessible. Per the user's later
  presentation decision, available values have no Partial badge or coverage-only
  explanation in visible, hover or screen-reader text. Technical coverage remains
  internal to accounting/API and comparison decisions.
  Comparisons require compatible complete metrics and nonzero prior values;
  misleading percentages are suppressed. Chart lines break through missing buckets;
  zero/one-point charts are safe. Calendar unknown/zero/future appearances differ.
- Removed Analytics fixtures and temporary development Home sample inclusion,
  local sample bookmarks and now-unused local Details props. Real feed fetching,
  categories/search/preferences, grouping, Details/bookmarks, reading-position
  behavior and Happening now remain on their existing live paths.
- No Home layout stylesheet, chat/composer ownership, draft/queue/Stop routing,
  overlays, independent Focus widgets/windows or mobile bridges were redesigned.
  Menu 240px preference / 160–360 bounds, wide launcher right88/bottom0, expanded
  dock right14 and identical Close/Minimize bar remain as approved source behavior;
  visual/runtime regression acceptance is still pending.

## Adaptations and limitations for review

1. Existing Usage is left as a status-bar snapshot rather than seeded into an
   unallocated counter map. Every resumed adapter must establish its current native
   baseline; this is stricter than trusting a migration checkpoint and cannot create
   an installation spend spike.
2. Graph lifecycle timing uses one narrow transaction hook rather than separate
   edits at every scheduler/runtime terminal branch. Graph reservation/finalization
   and timing remain atomic.
3. Native child duration/usage is deliberately not implemented from incomplete task
   maps. Identity-only OpenCode evidence is partial. The approved surfaces are wired
   and expose those limitations; universal child measurement is not delivered.
4. Nonserial root waits without exclusive-blocked evidence are unclassified spans.
   Gateway elicitation with unknown native origin marks partial coverage without
   blindly subtracting root time. More precise native signals would be additional
   adapter work, not a reason to invent complete time.
5. Fresh/resumed Codex token totals remain partial unless stronger starting/final
   evidence becomes available. No managed Codex delegation was enabled.
6. No historical execution or token backfill is supplied. Compilation alone does
   not validate adapter event ordering, callback races, journal recovery under a
   real crash, actual native totals, DST UI behavior or frontend visual fidelity.

## Exact implementation file list (relative to the initial shared tree)

Engine:

```text
engine/analytics_records.go                 added
engine/analytics_transactions.go            added (R2 correction)
engine/analytics_corrections_test.go         added (bounded in-memory correction checks)
engine/analytics.go                         added
engine/analytics_api.go                     added
engine/api.go                              route only; preexisting edits preserved
engine/store.go
engine/main.go
engine/transactions.go
engine/stream_journal.go
engine/harness.go
engine/usage.go
engine/pi_interactive.go
engine/opencode_interactive.go
engine/opencode_permission_gate.go
engine/codex_interactive.go                 preexisting runtime edits preserved
engine/permissions.go
engine/questions.go
engine/mcp_elicitation.go
engine/linked_bridge.go
```

Frontend (AnalyticsPage/analytics.css already existed untracked at handoff):

```text
clients/desktop/src/features/analytics/AnalyticsPage.tsx
clients/desktop/src/features/analytics/analytics.css
clients/desktop/src/features/analytics/analyticsApi.ts       added
clients/desktop/src/features/analytics/useAnalytics.ts       added
clients/desktop/src/features/analytics/format.ts             added
clients/desktop/src/features/analytics/fixtures.ts           deleted (preexisting untracked)
clients/desktop/src/features/klm-agent/KlmAgentWorkspace.tsx
clients/desktop/src/features/klm-agent/GithubActivityCard.tsx localDetails removed; now matches HEAD
clients/desktop/src/features/klm-agent/sampleFeedPosts.ts    deleted (preexisting untracked)
```

Documentation:

```text
ANALYTICS_IMPLEMENTATION_PROGRESS.md          added
ANALYTICS_VISUAL_PROTOTYPE.md                 status/accounting updates
product.md                                  targeted Analytics/sample status updates
```

`clients/desktop/src/engine.ts`, `agent-home.css`, runtime/platform/release/settings
work and the existing untracked debug executable were not edited by this task.
`ANALYTICS_IMPLEMENTATION_PLAN.md` remains unchanged. Git status alone does not list
the deleted untracked fixture files or the Details file restored to its HEAD content;
the list above records changes relative to the actual initial shared working tree.

## Commands and results

- Initial and final `git status --short`; scoped diffs inspected to preserve shared
  work. Applied `gofmt -w` to the changed Go files; its initial syntax finding in a
  newly written weekday loop was corrected, then formatting succeeded.
- Scoped `git diff --check -- <changed tracked paths>` passed (exit 0); Git emitted
  only the repository's LF-to-CRLF working-copy notices.
- From `engine`:
  `go build -tags dev -o "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-analytics-dev.exe" .`
  Initial attempt found a duplicate existing `minTime` helper and the graph name's
  actual `GraphRecord.Definition.Name` location. Fixed both; the same dev build
  passed. Repeated only that build after subsequent scoped unknown-tail and
  wait-classification fixes; each passed. **Final result: exit 0.** Output stayed
  outside the repo. The built executable was not run.
- From `clients/desktop`: `npx tsc --noEmit`.
  Initial attempt found nullable `reduce` inference in chart selection; added the
  explicit numeric accumulator. Retry passed with **exit 0**. Subsequent frontend
  edits only added English reason-code strings to the formatting map.
- No Go/frontend full suite, runtime probe, browser automation, integration test,
  engine restart, installer operation, commit or push was performed.

## Originating-review corrections R1–R3

Corrections were requested through the existing authorized review/correction
workflow. Re-read the current source and shared `git status --short` before edits.
Only these six paths changed during this correction pass:

```text
engine/analytics_records.go
engine/analytics_transactions.go            new
engine/transactions.go
engine/analytics.go
engine/analytics_corrections_test.go         new
ANALYTICS_IMPLEMENTATION_PROGRESS.md
```

### R1 — unaccepted reservation lifecycle

- Extracted the production checkpoint body into `checkpointAnalyticsLocked` and
  explicitly skip missing, ended or unaccepted executions before `analyticsInstant`.
  Slow preparation cannot create a clock-discontinuity diagnostic or poison later
  working spans. Acceptance establishes its first native anchor and clears stale
  preparation-era timing flags only while the record is still unaccepted.
- Unaccepted settlement/recovery takes a clock-free path. No dispatch means no
  unknown tail or execution/usage uncertainty. Dispatch without acknowledgement
  retains `dispatch_acceptance_unknown`, bounded from Dispatched to settlement/
  recovery, with no measured span. Genuine post-acceptance drift, interrupted tails
  and offline exclusion retain their prior behavior.
- The bounded check exercises preparation -> checkpoint -> acceptance -> exactly
  10 seconds observed work -> settlement; preparation failure with covered empty
  use/time/token totals; unconfirmed dispatch; and both unaccepted restart cases.
  No sleeping or native engine/harness is involved.

### R2 — incremental accounting in lifecycle transactions

- Fixed the pointer-level whole-ledger serialization issue rather than only adding
  pointer descent. Generic transaction cloning/diffing now skips the Analytics
  field. Dedicated transaction-local overlays stage only touched executions/runs;
  unrelated transactions neither clone nor compare the collector's historical maps
  or usage arrays. No full-map shallow copy is required either.
- Timing/reason/wait edits clone only the touched record's mutable fields. Usage is
  immutable to lifecycle callbacks and shared; attempting to change an existing
  row's usage there is rejected in favor of the atomic usage stream mutation.
- The existing format-1 lifecycle frame contains sparse execution field/span
  patches and changed small run rows together with ordinary session transitions.
  Publication of staged rows follows the frame's fsync. Replay adds descent only
  for the Analytics pointer and its execution/run maps; unrelated patch semantics
  are unchanged. Old full-object Analytics patches still replay. One-time nil
  collector initialization is the only allowed full-object Analytics patch.
- The actual generated checkpoint probe contains 256 closed executions × 256
  retained usage samples, plus one live execution and one open run. It confirms
  pointer sharing on an unrelated transaction, no usage-array copy on the active
  checkpoint, unchanged published state before publication, and these four patches:

  ```text
  [sessions 0 status]
  [analytics executions live last]
  [analytics executions live spans 0 end]
  [analytics graphs run]
  ```

  The encoded patch array was **326 bytes**. No closed record, whole collector,
  counter or retained usage payload appeared in the patches. In-memory replay of
  a legacy whole-object base followed by these sparse patches matched publication;
  nil legacy initialization was also checked. This is bounded synthetic evidence,
  not real crash/fsync/runtime acceptance or a performance benchmark.

### R3 — consistent all-role coverage

- Unsupported child coverage now propagates to `BlockedMs`, `InputTokens` and
  `OutputTokens` as well as `WorkingMs`; `TotalTokens` inherits the component
  coverage. The policy does not depend on observing child metadata, so missed early
  messages and unsupported role paths cannot falsely produce complete all-role
  token/wait totals. Delayed receipt-date usage retains that same coverage policy.
- Existing measured root values remain unchanged lower-bound sums. Missing metrics
  remain unavailable. Managed Codex's known-disabled native-delegation case stays
  separate. Existing frontend comparison gates already require complete coverage,
  so the corrected API suppresses those misleading comparisons without UI edits.
- The bounded check uses a complete Pi root and a complete OpenCode side execution
  with no child instrumentation: work, blocked waits, input, output and total tokens
  are all partial, retaining the observed 100 input / 20 output tokens. The disabled
  managed Codex case keeps complete known components and not-applicable child count.

### Correction verification

From `engine`:

```text
gofmt -w analytics_records.go analytics_transactions.go analytics.go transactions.go analytics_corrections_test.go
go test -tags dev -run "^TestAnalyticsCorrection" -count=1 -v .
go build -tags dev -o "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-analytics-corrections-dev.exe" .
```

- Only the three named bounded in-memory correction checks ran; all passed
  (`TestAnalyticsCorrectionReservationLifecycle`, `TestAnalyticsCorrectionSparsePatches`,
  `TestAnalyticsCorrectionAllRoleCoverage`). No integration/complex tests or full
  suite was run.
- The one correction-pass Go dev build passed (exit 0). Its absolute temporary
  output directory was verified first; the executable was not run.
- No frontend file changed, so no frontend typecheck was repeated.
- Final `git status --short` was inspected; scoped
  `git diff --check -- engine/transactions.go` passed (only the LF/CRLF notice).
- No live engine/harness probe, installed-engine lifecycle, session/graph/agent
  creation, commit or push occurred. Review of these corrections precedes manual
  validation. Native adapter completeness and real persistence/runtime acceptance
  remain the previously documented limitations, not solved by these pure checks.

## Review hotspots and later manual route

Originating review should concentrate on atomic usage replay/checkpoint equivalence;
native acceptance/idle/termination timing; pending callback success/failure versus
wait classification; child-origin routing before task metadata; missing components,
counter resets and native resume gaps; and half-open local-calendar aggregation.
Also inspect current membership/owner links, interrupted tails crossing midnight,
captured graph grouping, aggregate distinct-set totals and stale frontend requests.

After that review and delegated corrections, use the complete acceptance route in
plan §9. Key manual steps are:

1. Launch an explicitly chosen development environment with fresh and preserved
   older state. Check honest empty/unavailable history and live-only Home content.
2. Run authorized ordinary/side work plus an unused created session in two projects.
   Check created versus used and period-wide identity counts; leave a completed
   runtime idle and overlap work to check accounting stops/adds correctly.
3. Exercise real permissions/questions/elicitation, failed replies, concurrent waits,
   Stop/resume and optional Pi stats failures. Compare available native totals over
   repeated turns; verify no lifetime spike, replay spend or invented child duration.
4. Only with express authorization, inspect real graph runs/continued nodes/retries.
   Check separate graph wall-clock and agent effort. In a disposable environment,
   interrupt/restart across an offline period and a date boundary; verify unknown
   tails remain uncharged with partial coverage.
5. Move/rename equal-name topics; archive/remove/restore project history. Check the
   general agent never enters selected-project totals. Exercise local Monday/DST,
   current-week elapsed comparison, the 13-week calendar and All time beyond it.
6. Switch filters rapidly, disconnect/reconnect, hide/return and cross midnight.
   Then validate the approved Home/Analytics/chat visuals on desktop, browser Focus,
   narrow/mobile and both themes, including overlays/drafts/queue/Stop/permissions,
   real feed grouping/Details, independent widgets/windows and Back to hosts.

## Originating-agent re-review outcome

Source re-review inspected the R1 reservation/acceptance/finalization distinctions,
R2 staged mutations and sparse patch generation/publication/replay, and R3 aggregate
coverage propagation, plus the bounded correction-check source. No remaining blocker
was identified in those corrections. They are cleared for manual validation, not
certified as runtime-correct.

The implementing session reported the successful correction checks and Go dev build
above, and the earlier frontend TypeScript check. The originating agent did not
repeat those checks or run an engine/browser probe. No installed engine or desktop
was updated/restarted by this workflow. Use current development builds for testing;
missing historical usage and incomplete native-child coverage remain explicit.

## User presentation decision — remove Partial indications

The user assumes measurements may be incomplete and requested no repeated Partial
indications. Removed the shared badge and coverage-only descriptions from available
values, including totals, table cells, rankings, chart captions/tooltips, calendar
titles and screen-reader labels. The helper no longer has partial-reason copy or an
incomplete-coverage fallback. No replacement warning or global disclaimer was added.

Measured values, honest unavailable `—`, useful unavailable/not-applicable details,
Live/provisional indicators and loading/error/Retry behavior are preserved. Backend/
API coverage flags, accounting and complete-only percentage/comparison safeguards
are unchanged; the technical coverage documentation above remains authoritative.

Changed paths in this focused presentation pass:

```text
clients/desktop/src/features/analytics/AnalyticsPage.tsx
clients/desktop/src/features/analytics/format.ts
ANALYTICS_IMPLEMENTATION_PROGRESS.md
ANALYTICS_VISUAL_PROTOTYPE.md
```

Verification: `npx tsc --noEmit` from `clients/desktop` passed (exit 0). Source search
of Analytics TS/TSX/CSS found no remaining user-facing Partial text; the API's technical
`partial` enum remains. No UI/runtime probe or full suite was run. No engine lifecycle,
session/agent/graph creation, commit or push occurred. Visual acceptance of this
presentation change remains a manual check.
