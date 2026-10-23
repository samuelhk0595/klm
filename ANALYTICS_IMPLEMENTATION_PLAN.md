# Analytics: engine-backed implementation plan

## 1. Status, scope and handoff

**Planning only. No implementation is delivered by this document.** The user approved the frontend layout; statements in `product.md` and `ANALYTICS_VISUAL_PROTOTYPE.md` that visual approval is pending are older than this request. Backend accounting and real behavior are not yet approved by manual testing.

Goal: replace deterministic Analytics fixtures with the smallest sound engine-owned accounting, durable storage, read API and frontend wiring for **Overview / Topics / Graphs**, without redesigning Home, Analytics or chat. Do not introduce an external analytics service, generic telemetry platform, billing/cost estimates, productivity scores or inferred historical usage.

This document is self-contained for the next implementation agent. The requested workflow is planning, a separately authorized implementation session (OpenCode, Sol 6.1, High), originating-agent review/corrections, then human testing. This plan/report does **not** authorize creating sessions, running graphs, changing permissions or launching reviews. Follow the real user's authorization and `AGENTS.md`. No commits or pushes unless explicitly requested.

Only this document is changed by the planning task. The shared tree already contains unrelated modifications, including `engine/api.go`, `engine/runtime.go`, `engine/codex_interactive.go`, `clients/desktop/src/engine.ts`, settings, release files, `product.md`, and the prototype frontend. Re-read current files and `git status --short` before implementing; never reset, replace or stage the shared tree wholesale. An untracked debug executable is not an analytics artifact and must be left alone.

## 2. Evidence from the current code

References below identify current functions/types rather than relying on line numbers in a changing shared tree. **Observed** means inspected source behavior, not validated real-harness behavior.

### Frontend

- `clients/desktop/src/features/analytics/AnalyticsPage.tsx`: all values and scope choices come from `fixtures.ts`. `summarize`, `WorkChart`, rankings, calendar, tables and comparisons recompute from fixture execution rows. Week index `12`, Friday index `4`, the anchor and project names are hard-coded. `Delta` treats a zero baseline specially but has no missing-data semantics. `Ranking` keys rows by display name, unsafe for real duplicate topic names.
- `fixtures.ts`: 13 Monday–Sunday weeks ending at the illustrative 2026-10-23 anchor. Execution IDs, token amounts, graph runs and project/topic names are synthetic. Tokens are calculated from minutes, **not measured**. Its created count only considers sessions also used in the period, which must be corrected for real inventory.
- `analytics.css`: existing full-width responsive cards, Overview-only KPIs, chart, calendar and tables can remain. No page title/date subtitle should be introduced.
- `KlmAgentWorkspace.tsx`: Home and Analytics stay mounted behind `hidden`; Analytics currently has no visibility prop. `hideChat` makes both Close and Minimize the same minimized state. The existing children contain the real conversation/composer. Preserve these ownership boundaries.
- `agent-home.css`: preferred navigation width 240px, 160–360px bounds, centered Home composition, timeline capped at 840px around max-800px posts, side-only shadows above 800px, bottom-touching chat. Wide Home **launcher `right: 88px`**, expanded dock **`right: 14px`**. Do not restore the old launcher offset.
- `KlmAgentWorkspace.tsx` / `sampleFeedPosts.ts` / `GithubActivityCard.tsx`: development Home mixes nine sample posts and local bookmarks/details with real feed data; standalone Focus feeds do not. Remove that temporary Home inclusion for delivery. Do not alter live feed fetching, pipeline grouping, bookmarks, Details, Happening now or independent widgets.
- `clients/desktop/src/engine.ts::request` is the shared fetch/error/AbortSignal path and uses `ENGINE_URL` from `platform.ts`. Reuse it rather than hard-coding localhost or ports.

### Persistence and inventory

- `engine/store.go::Session` retains `ID`, `ProjectID`, `ParentID`, `Role`, `GraphRunID`, `GraphNodeID`, `TopicID`, `CreatedAt`, current status and the latest `Usage`. It has no durable ordinary-turn ID, timing ledger or usage-by-date ledger. `nativeSession` retains native ID/path.
- `diskState` retains all sessions/projects, graph records and native mappings. `now()` emits UTC RFC3339Nano. `Project.Removed` hides rather than deletes retained identity/history.
- `api.go::getState` filters removed projects and graph-node sessions; the general agent requires its dedicated path. This UI inventory is **not** a complete analytics source. Do not aggregate browser session summaries or use Happening now as an accounting feed.
- `topics.go::sessionTopic`, topic assignment/moves and `TopicSchema`: topic identity is project-scoped; only ordinary top-level sessions persist their topic. Resolve side/native children through their owner, graph nodes through `GraphRun.ConversationID`. Current membership is available; historical topic-movement attribution is not.
- `store.go::commitLocked` delegates to transactional journal persistence. `transactions.go::commitTransactionLocked` and `stream_journal.go::{appendStreamRecord,replayStreamJournal,checkpointJournal}` provide ordered durable revisions, fsync-before-publication, replay and checkpointing. `loadState` rejects unsupported/corrupt state. Preserve these guarantees.
- `usage.go::adapter.setUsage` deliberately uses a small `streamRecord` on the hot path rather than checkpointing/cloning all histories. Analytics cannot become a second non-atomic write after that snapshot.

### Executions, waits and recovery

- `main.go::turn` is an in-memory handle, not a persistent identity. `harness.go::app.execute` is shared by main, side, general and graph-node agents; it includes preparation, bridge startup, native work, native/tool drain and shutdown. Timing the entire function would include overhead that is not native working time.
- `harness.go::finishTurnStateLocked` settles cancellation/errors/queue state; `runtime.go::finishTurn` retains runtimes. Runtime retention, subprocess/server survival and UI status are **not** continuing agent effort.
- Dispatch occurs in `pi_interactive.go::runPi` (`prompt`), `opencode_interactive.go::runOpenCode` (`prompt_async`), and `codex_interactive.go::runCodex` (`turn/start`). Native completion is Pi `agent_settled`, OpenCode confirmed idle/final reconciliation, Codex matching `turn/completed`. Pi `agent_end` is explicitly not the final settlement event; `turn_start`/`turn_end` are internal cycles, not new KLM sessions.
- `message_queue.go::scheduleMessagesLocked`, `linked.go::scheduleLinkedLocked`, ordinary message dispatch and `graph_runtime.go::launchGraphAgent` all reach `execute`. Steering stays in a turn; Send now starts a replacement after shutdown. A turn ID cannot simply be the first user message ID: consultations, notifications and graph corrections also execute.
- `permissions.go::{requestPermission,dismissPermission,replyConversationPermission}` and `questions.go::{requestQuestion,dismissQuestion,replyConversationQuestion}` retain pending cards but discard them on resolution. Automatic approvals are not human blocked waits. Reply callbacks execute outside the state lock and may fail; `Resolving` alone does not mean unblocked.
- `mcp_elicitation.go::mcpElicit` uses the same pending-question flow, with a bounded actual wait. `linked_bridge.go`'s ask dispatch waits at most five seconds before yielding. The entire lifetime of a consultation or its eventual notification must not be charged as requester blocked time.
- OpenCode child requests are presented through the parent adapter and may arrive **before child task metadata**. UI card location does not prove which execution is blocked. Native origin is available in request/session/thread metadata but needs explicit accounting propagation.
- `main.go::runEngine` recovers running sessions as interrupted. `graph_records.go::interruptGraphState` currently writes `EndedAt = restart time` for active runs/activations. That timestamp is recovery bookkeeping, not proof of execution during engine downtime.

### Tokens and native children

| Source | Observed semantics | Implication |
| --- | --- | --- |
| `usage.go::piSessionUsage` | `get_session_stats` cumulative native-session tokens; input = input + cacheRead + cacheWrite; output is reported output | Never sum snapshots. Initial stats and stats after assistant messages/compaction are optional; final stats can be missing. |
| `pi_interactive.go::runPi` | Correlates stats request ID and session file; an initial request is issued during setup but the prompt need not wait for its result | Establish analytics baseline before dispatch if possible; a late initial snapshot cannot safely be treated as this turn's usage. |
| `harness.go::adapter.pi` | Stores assistant `message_end.usage` on **each content part**; `agent_end` includes repeated messages | Do not sum event Data usage: one message can appear multiple times and local message indices reset each adapter. |
| `usage.go::openCodeUsage` | Replaces messages by message ID and `step-finish` parts by part ID; steps take precedence over message fallback; sums the native session | Replay/reconciliation is idempotent in memory, but the analytics dedupe/baseline must survive process/engine restart. Input includes both cache fields; output includes reasoning separately. |
| `opencode_interactive.go::runOpenCode` | Walks preexisting root messages before prompting, then SSE plus final reconciliation | Strong baseline opportunity for root/main/side/node tokens. Child usage maps start when task metadata is found and are not a guaranteed complete child history. |
| `usage.go::codexSessionUsage` | Uses thread `total`, not `last`; input already includes cached input, output includes reasoning | Do not add cached/reasoning fields again or use context usage as spend. |
| `codex_interactive.go::frame` | Filters root usage to owned thread/turn; `childFrame` handles known children | Existing stored snapshot is not sufficient proof of a fresh pre-turn baseline after interruption or native history changes. |
| `subagents.go::ensureSubagent` | Persists native child as a hidden session, deduped by parent + native ID | Count hidden child identity once across repeated invocations; do not count each parent card or tool update as a new child. |
| Same function | Refuses children when parent already has ParentID or GraphRunID | Current native-child visibility does not cover side-agent children, graph-node children or nested native descendants. Do not claim universal child coverage. |
| `openCodeSubagent.updateTask` | Task invocation key distinguishes reopening a child from terminal replay; terminal states can be replayed | Reuse these keys for observed invocation identity, not display title; task pending duration is not automatically actual child execution duration. |
| `codex_library.go` managed profile | Disables `multi_agent` and `multi_agent_v2` | Native children are disabled on the managed Codex path. Do not re-enable native delegation for Analytics. Older child decoding is not proof current execution supports it. |
| Pi owned adapter | No `ensureSubagent` integration in Pi event normalization | No generic Pi native-child metrics can be promised; a shell-launched agent is not automatically a measured child. |

Graph facts: `graph_scheduler.go` creates one durable `GraphRun` at actual run reservation, separately from a scheduled activity. `graph_runtime.go::launchGraphAgent` already creates a unique `turnID` per node attempt, including corrections, and can reuse a node session with `continue_target`. `GraphActivation` includes occurrence and current/finished turn IDs. `GraphRun` has captured `GraphID`, current `CatalogGraphID` alias, snapshot, CreatedAt and EndedAt. Run IDs, not activity IDs or notifications, identify attempts. Terminal/Fork bookkeeping is not agent execution; a Join's actual agent is.

## 3. Conservative v1 accounting policy

These are proposed implementation choices, not claims that the prototype already implements them. They avoid a blocking product question for a useful first delivery.

### 3.1 Scope and identity

1. **All projects = nonremoved registered projects.** Include archived sessions/topics in historical analytics; archive is not deletion or Stop. Exclude removed projects from the selector and results, without deleting their ledger. Restoration reveals their retained figures with the same IDs. Revalidate this on every request.
2. Collect the general agent's own executions with `projectId = null`, role `general`, but **exclude them from this project-filtered v1 UI**. Do not invent a general-agent project or assign its work to the selected project. Its spawned ordinary sessions count under their real project. A future general/engine scope is an explicit product extension, not a hidden change to “All projects.”
3. `ownerSessionId`: main session for main/side/native work; invoking ordinary conversation for graph work; general session for general-owned work. Follow explicit relations, never cwd, name, prompt text or sender attribution. A spawned independent session owns its own subsequent work.
4. **Sessions used**: distinct ordinary owner-session IDs with a confirmed execution or real graph-run activity intersecting the period. Do not count hidden side/node/child sessions as additional user sessions. Mere creation, queued input, retained runtime and preparation failure are not use.
5. **Sessions created**: distinct retained ordinary top-level sessions with CreatedAt inside the period, including unused and independently spawned sessions. This can exceed Sessions used. Do not infer creation from the first execution. General/side/node/native sessions do not inflate it.
6. **Agents executed**: distinct executed KLM session IDs for main, side and graph-node roles. Repeated turns/corrections/continued node sessions are one agent identity but add their distinct working intervals. Native children are a separate distinct count by parent + harness + native session/thread ID (use the stable hidden Session.ID where available).
7. Topics use **current owner-session membership**, matching the approved Topics detail. Project + topic ID is the key. Side/node/child work follows its owner. Include a real per-project **Ungrouped** bucket with null topic ID. Equal names never merge. Topic moves regroup old observed work; no invented historical attribution and no retroactive ledger rewrite. Topic renames/colors use current metadata. Empty retained topics may appear with zero recorded activity when the period is covered.
8. Graph grouping uses project + **captured GraphID**, not display name or current catalog alias. Snapshot names label historical rows (latest captured name within the group, deterministic tie break). Deletion does not erase observed runs; rename may produce a different captured-ID row. Preserve aliases as metadata only. Do not promise a durable definition identity across delete/recreate of the same file slug; document that v1 groups by captured ID.

### 3.2 Working time and identified waits

- Internal units: integer milliseconds, UTC instants. Round only for display. Each agent execution has a single nonoverlapping timeline of **working**, **identified blocked**, or **unknown** intervals. Sum timelines across identities; working parent and child intervals may overlap and legitimately contribute two agent-hours.
- Begin at confirmed native work (native acceptance/start signal, or a correlated successful prompt submission with observed dispatch time), not session creation, queue insertion, runtime startup or capability preparation. A reserved dispatch with no confirmed acceptance is uncertain, not measured work.
- End normal work at native settlement, not after final token reads, process flush, graph seal/drain or runtime cleanup. On cancellation/error without a native end, use confirmed owned termination as the observed end; flag uncertain shutdown rather than pretending the stop-click time was final. Keep waiting for permission entirely separate from any idle time after a turn ends.
- Measure **observed active execution less explicitly identified blocking intervals**. It is not CPU/GPU or pure inference time. Normal tool execution belongs to its executing agent; detached servers after settlement do not. Explicit retry/backoff only becomes blocked time when the harness supplies an actual bounded waiting state; do not parse prose to infer it.
- Maintain a set of wait keys on each execution. Start on an actual blocking permission/question/elicitation/consultation-wait transition; end on confirmed callback completion, native dismissal, timeout or cancellation. Multiple waits overlap as a union, not a sum. Failed replies keep the wait open. Automatic permissions do not create user-wait spans.
- Carry an internal native origin on requests before projection to the owning UI. A child's request does not stop the parent clock just because its card is shown there. Hold unresolved-origin requests until identity resolution; if attribution is impossible, mark coverage partial, do not silently subtract from the root. Preserve public permissions/security routing.
- Do not assume any pending tool means the entire agent is blocked when other native work can proceed. Classify a parent waiting on a native child only when the adapter proves it is exclusively blocked; otherwise retain an unclassified span/partial timing coverage. Never subtract child duration from the parent merely because a relationship exists. Unsupported native-child timing stays unknown; task queued-to-finished wall-clock is not fabricated as child working time.
- Graph-run wall-clock remains an independent measure including startup, tool waits, joins and shutdown until terminal settlement. Nodes already contribute agent-hours. **Never add graph wall-clock to agent-hours.** Concurrent runs each have their own wall-clock; this sum can exceed elapsed human wall time.
- Persist a modest liveness checkpoint (e.g. every 15s only while executions/runs are open), plus every boundary and terminal transition. On an abrupt restart, close observed timing at the last durable checkpoint, mark the tail unknown/interrupted, and never charge offline time. A quiet long tool call needs these checkpoints; last message time alone is insufficient. Live GETs may project to request asOf only for currently owned, healthy open executions, with provisional coverage.
- Use monotonic elapsed time inside a process and UTC anchors for buckets. On a wall-clock discontinuity, do not produce negative/huge intervals: checkpoint and mark the affected mapping unknown. Durations are observations, not estimates of how much unobserved work occurred.

### 3.3 Token accounting

- Tokens are **reported input including cache + reported output including reasoning** according to the existing per-harness normalization. No dollar cost, no conversion from time, no context-window-as-usage. Input and output each remain nullable independently.
- Namespace counter state by KLM session + harness + native session identity (Pi identity plus normalized session file). Native replacement creates a new counter epoch only when identity change is proven. Ordinary managed-process replacement does not.
- Establish a trusted pre-dispatch baseline per component. Known newly created empty native history may start at zero; an existing history needs an authoritative baseline. Seed migration from current Usage as a checkpoint only, never a historical event. On the first subsequent resume, reconcile before prompting; any gap belongs to unknown/unattributed prior usage, not the new turn.
- For an authoritative monotonic cumulative counter with baseline B and current C, persist only `C - B`, then atomically replace B with C. Equal snapshots/replayed events add nothing. Nil is not zero and never clears a trusted high-water mark. A decrease or correction is **not** permission to count C from zero: diagnose/rebaseline at a native boundary, mark the interval incomplete and do not mint a second copy of old usage. Missing-final-snapshot coverage stays incomplete even if the next turn can recover a native total.
- Partial input/output updates need separate baselines and coverage. Never add root totals and child totals unless those sources are confirmed disjoint. OpenCode root history is session-scoped; verify child collection is equally scoped. Codex totals are thread-scoped in the adapter, but managed child execution is disabled. Do not use a combined provider/global total as another source.
- OpenCode: use the pre-prompt root history walk as baseline and the existing per-message/per-step precedence. Child live maps alone are not authoritative cumulative snapshots: reconcile the owned child's history before accepting a full-snapshot delta; otherwise record only deduplicated completed step/message units with proven identity and mark the uncovered range partial. Choose one accounting mode per native epoch, not snapshot deltas plus step deltas together.
- Pi: do not derive spend by adding persisted assistant content parts. Add a bounded pre-prompt stats reconciliation where possible; absence must not block ordinary execution, but makes token accounting unavailable/partial for that turn. Existing optional stats and the final-read path need an explicit coverage outcome, not a promise all turns return final totals.
- Codex: use persisted baseline only when its continuity is verified. A fresh native thread can begin at zero. On resumed/interrupted history without a trustworthy starting total, use the first valid total as baseline and mark the first interval unknown; do not allocate the thread's lifetime total to the current prompt. Do not assume `last` is the whole engine turn.
- **Date attribution:** v1 tokens belong to the engine receipt timestamp of the accepted usage delta. Do not divide them across time in proportion to duration. When a delayed sample can be correlated to its still-known completed execution, keep that execution identity, but keep the observed timestamp. If it spans an unobserved/native gap, retain it as unattributed diagnostic coverage rather than inventing a dated usage row. This is observed usage by date, not a provider invoice.
- Summed token values must remain exact within the existing JSON-safe integer limit (`maxTokenCount`). Overflow yields unavailable + diagnostic, never a wrapped or rounded integer.

### 3.4 Coverage and history

- Install a durable `collectionStartedAt` and schema version once. **No synthetic migration of execution time or daily usage.** Do not infer it from CreatedAt/UpdatedAt, text lengths, the full stream journal, fixture values or graphs' elapsed duration.
- Existing Session.CreatedAt is factual inventory evidence. Existing GraphRun IDs and ordinary completed/blocked/failed CreatedAt–EndedAt intervals can supply retained historical graph metrics when valid. Interrupted legacy runs have unknown end timing because recovery overwrites EndedAt; do not treat them as exact durations. A scheduler failure before creating a run is not a run.
- Historical session-used/agent/child/time/token metrics before collection are unavailable, even when creation and graph facts are available. For historical graph contribution to Sessions used, use its explicit owner; otherwise pre-collection ordinary chat usage remains unknown. Do not present the known subset as a complete historical count.
- Unsupported child coverage is a capability limitation, not an observed zero. Managed Codex native delegation disabled can be reported as not applicable for that path; Pi arbitrary children and OpenCode side/node/nested children remain unsupported. V1 need not redesign native child history to expand this coverage.
- A metric carries `value` (sum/count of observed evidence or null), `coverage` (`complete`, `partial`, `unavailable`, `not_applicable`), and reason codes. Complete means complete **for the declared observed accounting scope**, never universal model/remote-work visibility. Preserve provisional active values separately. Zero is valid only in a covered empty bucket or known disabled capability.
- Display missing as `—`; display partial observed sums with a small accessible `Partial` cue/tooltip in the existing surface, not as exact totals. Calendar distinguishes no data, zero work, and future. Break chart lines through missing buckets. No fixture fallback on error, no zero-filling pre-collection history. Avoid repeated limitation paragraphs in the UI; keep detailed explanations here/docs.
- Suppress percentage comparisons when either metric is unavailable/partial or the prior value is zero. A newly introduced collector does not justify an apparent infinite productivity gain.

## 4. Proposed minimal persistence contract

Add focused `engine/analytics_records.go`, `analytics.go` (normalization/aggregation) and `analytics_api.go`. Do not expose ledger contents or native paths in `/api/state`.

Suggested durable state, with explicit versioned initialization after journal replay:

```text
AnalyticsState
  schemaVersion: 1
  collectionStartedAt: UTC instant
  executions: map<executionID, AnalyticsExecution>
  counters: map<nativeEpochKey, AnalyticsCounter>
  graphTiming: map<runID, AnalyticsGraphTiming>  // new runs/checkpoints only

AnalyticsExecution
  id; sessionId; ownerSessionId; projectId?; role; harness
  nativeEpochKey?; nativeInvocationKey?; parentExecutionId?
  graphRunId?; activationId?; graphTurnId?
  reservedAt; acceptedAt?; nativeEndedAt?; lastObservedAt; terminalStatus?
  spans: [{sequence, start, end?, kind: working|blocked|unknown}]
  openWaitKeys: bounded set of internal identifiers
  usage: [{sequence, observedAt, inputDelta?, outputDelta?, sourceKey}]
  coverage/reasons for identity, timing, wait classification, input, output, children

AnalyticsCounter
  nativeEpochKey; lastInput?; lastOutput?; observedAt
  owningExecutionId?; authoritativeBaseline; per-component coverage
  mode: cumulative | deduplicated_units
  unit checkpoints only when a harness cannot provide full cumulative snapshots

AnalyticsGraphTiming
  runId; lastObservedAt; observedEndAt?; interruptedTailUnknown
```

- Generate an ordinary execution ID once per `execute` invocation before dispatch; persist reservation before side effects. Store it on `turn`/adapter. Reuse the existing graph turn ID for node-attempt correlation (or retain it as a unique alias); corrections get another execution, same agent identity. All dispatch entry points must reach this shared reservation.
- Native child invocations need their own stable key, not the parent's engine-turn ID alone: OpenCode native task call/part key plus child identity, Codex native thread/turn identity when available. Replayed terminal parts cannot open another execution. Multiple references to the same child invocation do not create duplicate intervals. Unsupported origins are explicitly unknown.
- Keep one open segment per execution; close/open on state transitions. Union simultaneous child invocation spans when they describe the same native execution. Do not dedupe different agents just because their intervals overlap.
- Store normalized IDs, times, counts and small enumerated diagnostics only. No prompts, file contents, tool outputs, request bodies, credentials, URLs or full native frames in this ledger.
- Use existing `commitLocked` for low-frequency reservation/boundary/finalization and startup recovery so lifecycle and analytics transitions commit together. For usage ticks preserve the existing small-journal path: extend `streamRecord` with a bounded analytics mutation that updates the execution's usage record **and counter baseline in the same fsynced frame as Session.Usage**. Replay/checkpoint must apply all parts exactly once by revision. Never “save usage, then save analytics” in separate transactions.
- The new stream extension must allow analytics-only checkpoint records when needed, validate permitted combinations of transaction/stream records, and retain backwards replay of records without analytics. Reject malformed references/sequences before publication. Checkpoint includes the new state fields. Missing fields on old state initialize safely; a failed persistence follows `failStorageLocked`, not a silently successful counter increment.
- If a simpler single `commitTransactionLocked` implementation is selected for low-rate usage, it must preserve atomic snapshot+baseline+delta and must not regress the deliberately optimized streaming path into full-history writes. Do not build another WAL/database or reconstruct accounting from rotated journal segments.
- No persisted daily aggregate in v1. Derive all totals/rankings from the same normalized rows and identity sets at query time. This avoids rollup invalidation for topic moves, timezone changes and period boundaries. Build a compact read snapshot under `app.mu`, aggregate after unlocking; don't clone native event histories/graph source text on every GET.
- Retain ledger with the existing engine data lifecycle. No automatic purge/downsampling in this slice. Document growth; add a simple checked record-size policy rather than silently dropping old rows if a limit is reached. Long-term retention policy is deferred, not permission to delete data.

### Exactly-once invariants

1. One accepted dispatch/invocation key maps to at most one ledger execution; duplicate finish/start/native history events cannot reopen terminal executions.
2. Snapshot and baseline advance atomically with its usage delta. Replay repeats state, not spend.
3. A native epoch survives process restart; explicit retry/Send now is a new execution with the same counter continuity unless native identity changes.
4. A repeated graph notification/assessment is not a new run or duplicate node execution. A concrete graph retry creates a new run ID and counts independently.
5. All output metrics are derived from normalized evidence once. Session/agent/native/graph counts use sets over the entire requested period, not a sum of daily cardinalities.
6. In-process active projection and persisted closed spans cannot both contribute the same tail. After restart no in-memory projection is replayed and no interval extends to startup time.
7. “Exactly once” applies to captured, durable observations, not guaranteed delivery of every upstream token event. Missing signals remain missing.

## 5. Periods, timezone and graph date rules

- Request an IANA timezone using `Intl.DateTimeFormat().resolvedOptions().timeZone` (UTC fallback only when unavailable). Engine validates with `time.LoadLocation`; embed Go `time/tzdata` for installed Windows portability. Do not use the engine host's local zone or a fixed UTC offset silently. Return the effective timezone and boundaries; show it in existing period/calendar accessible detail.
- All local weeks begin Monday at 00:00 and end next Monday, **half-open `[start,end)`**. Calculate boundaries with calendar dates/AddDate in the chosen location, not multiples of 24h. Split time spans at local midnight for chart/calendar/weekday buckets; DST days can be 23/25 hours.
- Current week ends at server `asOf`; future days are upcoming regardless of whether the project has any data. Historical weeks are complete calendar weeks. Week-to-date comparison is the previous Monday through the same local weekday/time-of-day, using calendar arithmetic (not subtracting 168 elapsed hours). This is a more precise live counterpart of the fixture's equal-elapsed-weekday rule. Return both cutoffs; document DST behavior. All time has no comparison.
- Keep the rolling **13-week calendar and selector** ending in the current week, independent of selected week. A click selects that date's Monday. Earlier/uncovered dates remain explicitly unavailable, not fixture heat. All time covers all retained evidence, not just those 13 weeks; earliest retained creation/run time and collection start determine its bounds.
- All-time chart uses consecutive week-aligned buckets. If more than 13 weeks exist, coalesce consecutive weeks into at most 13 contiguous ranges, labelled by actual start with full range in hover/accessible text. Do not truncate All time to the calendar window. One/zero-point charts must not divide by `groups.length - 1` or index `-1`.
- Durations are clipped to the selected interval and split into days. Graph runs count distinct IDs whose known run interval overlaps the period; active runs overlap through asOf. Run hours are the sum of their clipped wall-clock intervals, **including waits but excluding unknown offline tails**. For an interrupted legacy run with no trustworthy end, count its known start if in range and mark run-count/duration coverage partial elsewhere; do not manufacture overlap after its last evidence.
- `Average run` should mean the average **observed in-period wall-clock per contributing timed run**, not full lifetime duration divided by a clipped period's start count. Keep the column in place and explain “In-period average” in its accessible detail; nullable/partial follows contributing timing coverage. Return `timedRuns` so the denominator is explicit. Do not show an exact average when any relevant duration is unknown.
- Current/previous summaries, rankings, weekdays and table totals use the same resolved project/time scope. Calendar active-days count is for its displayed 13-week scope (days with positive known working time), not all lifetime dates. Table total rows come from aggregate sets, never adding distinct counts in child rows blindly.

## 6. Proposed read API and frontend contract

One endpoint is sufficient for the existing surfaces:

```http
GET /api/analytics?period=week&weekStart=2026-10-19&projectId=all&timeZone=America%2FSao_Paulo&compare=previous
GET /api/analytics?period=all&projectId=<real-project-id>&timeZone=UTC&compare=none
```

The date above illustrates syntax only, not a seeded observation. Omitted period/weekStart selects the current week. Accept `weekStart` only as a valid Monday date, and only one value per scalar query parameter. Validate timezone, enums and project identity. Missing/removed explicit projects return 404; invalid dates/unsupported future ranges return 400; unavailable persistence returns the existing 503 convention. Reads never invoke harnesses or alter Library/session/graph selections.

Suggested typed response (normal JSON numbers limited to safe integers):

```ts
type Metric = {
  value: number | null;
  coverage: 'complete' | 'partial' | 'unavailable' | 'not_applicable';
  provisional?: boolean;
  reasons?: string[]; // stable codes; English UI maps only relevant messages
};
type Totals = {
  sessionsUsed: Metric; sessionsCreated: Metric;
  agentsExecuted: Metric; nativeSubagents: Metric;
  workingMs: Metric; blockedMs: Metric;
  inputTokens: Metric; outputTokens: Metric; totalTokens: Metric;
  graphsUsed: Metric; graphRuns: Metric; graphWallMs: Metric;
};
type AnalyticsResponse = {
  version: 1; revision: string; asOf: string; timeZone: string;
  collectionStartedAt: string;
  scope: { projectId: string | null; includesArchived: true; includesGeneral: false };
  projects: { id: string; name: string }[];
  period: { kind: 'week' | 'all'; start: string; endExclusive: string;
    weekStart?: string; current: boolean };
  availableWeeks: { startDate: string; endDate: string; current: boolean }[];
  summary: Totals;
  comparison: null | { start: string; endExclusive: string; summary: Totals };
  series: { key: string; start: string; endExclusive: string; label: string;
    future: boolean; workingMs: Metric; tokens: Metric;
    previousWorkingMs?: Metric; previousTokens?: Metric }[];
  weekdays: { weekday: number; future: boolean; workingMs: Metric }[];
  topics: { key: string; projectId: string; topicId: string | null;
    name: string; projectName: string; color: string; archived: boolean;
    totals: Totals }[];
  harnesses: { id: string; name: string; workingMs: Metric }[];
  graphs: { key: string; projectId: string; graphId: string;
    name: string; projectName: string; runs: Metric; wallMs: Metric;
    timedRuns: number; averageMs: Metric }[];
  calendar: { startDate: string; endDate: string; activeDays: Metric;
    days: { date: string; weekStart: string; future: boolean; workingMs: Metric }[] };
  coverage: { reasons: string[]; childSupport: {
    harness: string; role: string; status: string
  }[] };
};
```

Implementation details:

- Engine resolves one `asOf` and immutable compact snapshot for **all** fields of a response. Do not fetch each table separately and race totals. Revision is a freshness hint; live elapsed projections can change without a ledger revision, so revision alone is not an HTTP cache validator.
- `scope.projectId = null` represents All projects, not a synthetic database ID. Topic/graph `key` is a collision-safe encoding of scoped identity (e.g. JSON tuple), not ambiguous colon concatenation. Stable keys also fix duplicate-name React rows.
- Totals for unavailable components must propagate coverage; `totalTokens` combines input/output only under the declared observed-sum semantics, never substitutes null with zero. If only one component is known, expose it independently and mark total unavailable/partial, not exact.
- Add `features/analytics/analyticsApi.ts` for response types and `getAnalytics` via `request`; `useAnalytics.ts` for loading, query state and visibility. Keep pure formatting/date labels out of `fixtures.ts` in a small local helper. No fixture execution type should remain the production view model.
- Pass `visible={visible && page === 'analytics'}` from `KlmAgentWorkspace`. Fetch on first visible entry, scope changes, focus and visibility return; refresh about every 15s only while visible. Abort obsolete requests; guard against old responses replacing a new project/week. Hidden Home/Analytics panels stay mounted to preserve selection; hidden polling stops. No new analytics SSE stream or polling every session history.
- Existing engine inventory changes can invalidate the chosen project; handle removed/restored project results and keep previous data only if visibly stale **for the same query**. A failing new filter must not show the prior project's values as its own. Initial loading/error/empty/partial states use current card slots and normal Retry, not sample values.
- Replace `Execution[]` aggregation in `AnalyticsPage`/`WorkChart` with server summaries/buckets. Preserve the tabs, controls, layout and chart interaction. Percent rankings use only compatible observed working values and carry partial coverage; don't imply a complete distribution when time is missing. All comparison availability is determined per metric by coverage, not by array length.

## 7. Ordered implementation steps and exit conditions

### Step 1 — Lock the semantics and remove hidden assumptions

Read this plan, current `AGENTS.md`, product/prototype and shared diff. Implement the conservative policies above unless the user explicitly chooses otherwise. Define the metric/coverage types and identity helpers first. In particular, separate sessions created/used, current topic grouping, general exclusion, observed timestamp tokens and graph wall time. No fixture imports in engine code.

**Exit:** field names and semantics can represent empty, partial and unavailable results without turning missing observations into zeros.

### Step 2 — Add durable normalized accounting, initialization and recovery

Files: new `analytics_records.go`, `store.go::diskState`, `stream_journal.go`, narrowly scoped `transactions.go` only if needed, `main.go::runEngine` recovery.

Initialize schema/collection timestamp after replay, seed existing usage as unallocated baselines, persist before serving. Add atomic lifecycle updates, small usage journal mutation and checkpoint replay. Add open-execution/run liveness checkpointing under existing app lifetime. Recovery closes at the recorded last observation before existing recovery rewrites statuses. Keep original native history and queues untouched.

**Exit:** an interrupted record cannot silently continue accumulating after restart; persisted baseline/delta cannot diverge.

### Step 3 — Wire root/main/side/general/graph-node timing and usage

Files: `main.go::turn`, `harness.go::{execute,finishTurnStateLocked}`, `pi_interactive.go`, `opencode_interactive.go`, `codex_interactive.go`, `usage.go`, graph launch/finalization.

Reserve once in shared execution; bind native identity, capture dispatch confirmation/native settlement and failure/cancel exits. Keep preparation and retained-runtime idle out of working spans. Use graph attempt IDs and owning conversation. Record steering inside the same execution, not a new agent. Token source handlers must supply native epoch, source observation identity, baseline vs live phase and completeness; `setUsage` alone lacks that context. Update Session.Usage exactly as today for the status bar while adding normalized accounting atomically.

**Exit:** one real ordinary turn can drive all non-child execution metrics; repeated snapshots/turn notifications have no additional spend or count. Side and graph node reuse the same path rather than separate timing formulas.

### Step 4 — Wire identified waits and supported native children

Files: `permissions.go`, `questions.go`, relevant callback removals in `live_yolo.go`, `mcp_elicitation.go`, bounded wait in `linked_bridge.go`, `subagents.go`, OpenCode task/child handlers and existing Codex child handlers where reachable.

Carry request origin internally; cover resolve/dismiss/deny/timeout/Stop/error paths. Close waits only once; do not change permission or question behavior. Add observed child-invocation keys and reconcile child usage when authoritative evidence exists. Unknown ownership/timing becomes partial coverage. Do **not** expand native delegation, allowlists or graph authorization to improve statistics. Do not create fake hidden sessions for unsupported children.

**Exit:** a known permission wait is excluded exactly once; a child request projected to a parent cannot blindly pause the parent; unsupported native-child rows are visibly not universal measurements.

### Step 5 — Reuse graph facts and implement one aggregation path

Files: `graph_scheduler.go` run creation, `graph_runtime.go` settlement, startup interruption handling, new `analytics.go`.

Read retained GraphRuns for graph identities/counts and valid existing durations. Add last-observed timing for new runs to avoid restart-time inflation. Implement UTC interval intersections, timezone/day/week splitting, current-topic ownership, per-scope distinct sets, project/archive rules and per-metric coverage. No graph duration enters agent working totals. Terminal-only runs still contribute graph facts, not an executed agent.

**Exit:** summary, tables, chart, weekday bars and calendar reconcile from one source under the same query, including cross-midnight and overlapping executions.

### Step 6 — Expose the read endpoint and wire Analytics

Files: new `analytics_api.go`, narrow route addition in `api.go`, new frontend `analyticsApi.ts`/`useAnalytics.ts`/formatting helper, `AnalyticsPage.tsx`, one visibility prop in `KlmAgentWorkspace.tsx`; minimal CSS only for loading/unknown/stale cues.

Validate query input, return compact summaries with coverage, handle cancellation and stale queries. Replace all hard-coded 2026 dates, fixture project/topic/graph IDs and minute/token conversions. Keep the fixed 13-week calendar and shared filters; All time must not accidentally mean last 13 weeks. Do not change the shared `request` or endpoint resolution unless a specific need is demonstrated.

**Exit:** opening Analytics reads actual engine facts and never displays fixture values if no facts are available.

### Step 7 — Remove temporary Home samples and update delivered-status docs

Files: `KlmAgentWorkspace.tsx`, `sampleFeedPosts.ts`, `GithubActivityCard.tsx` only for now-unused `localDetails` plumbing after checking callers, `fixtures.ts` after extracting real formatting helpers, `ANALYTICS_VISUAL_PROTOTYPE.md`, narrowly targeted `product.md`, optional engine README API section.

Remove sample import/filter/render/local-bookmark state and sample-based project counts/empty checks. Preserve real posts and all recent shared feed changes. Delete unused fixture modules only after checking imports. No stylesheet cleanup/redesign. Update docs to distinguish approved layout, actual implementation delivered, collection start/coverage, and manual behavior still pending. Planning alone does not justify those production-status updates.

**Exit:** Home and Focus widgets are live-only in development and production; docs never describe compilation as real runtime acceptance.

## 8. Essential decisions and blockers

### No product decision blocks the conservative v1 above

Use these explicit defaults: project-only UI, archived history included, removed projects hidden/restorable, current topic membership, future-only execution accounting, historical creation/valid graph facts only, observer-receipt token dates, local Monday weeks, equal local elapsed-time comparison, and partial native-child coverage. Layout approval is not evidence that missing metrics exist.

### Ask the user only if these requirements are imposed

- **Complete historical agent-hours/tokens before collection:** unavailable evidence is a hard blocker to claiming this. Offer future collection and independently factual old metrics, not fabricated backfill or an estimation feature disguised as facts.
- **Universal native-child accounting across every harness/role:** current adapters do not provide it. Full support would require a separately scoped adapter capability expansion and real runtime evidence. Until then ship explicit partial/unavailable coverage, not zeros. Do not enable Codex delegation or unsupported extensions.
- **General-agent activity included in the project dashboard:** requires an explicit scope choice/label (e.g. engine-wide or separate KLM Agent option). It cannot be silently attributed to a project.
- **Historical topic attribution instead of current membership:** needs a new policy and forward topic-assignment snapshots/events; old movements cannot be reconstructed reliably. Current-membership regrouping matches the prototype and is the default.
- **Cost, billable tokens, unlimited retention guarantees, external/native work outside engine ownership, or full graph lineage through slug reuse:** additional product scope, not implied by this delivery.

Implementation blockers are technical, not requests to lower accuracy: if a baseline cannot be established or a wait cannot be attributed, represent the metric as partial/unavailable. If the existing persistence path cannot atomically record a delta and its baseline, correct that before showing any token total as durable. Keep the rest of the approved feature working; do not spend the first delivery building a speculative telemetry framework.

## 9. Lightweight verification and manual acceptance

### Checks for the implementation agent

Do not run full suites or add integration/complex unit tests. After the scoped edits, run **one Go dev build** (e.g. `cd engine && go build -tags dev -o <absolute temporary output outside the shared tree> .`) and **one frontend type check** (`cd clients/desktop && npx tsc --noEmit`). These cover different languages; avoid overlapping builds/tests. Inspect the scoped diff for unrelated changes and compile fixes. Report actual commands/outcomes only. No graphs, live model probes or review agents are implicitly authorized by this plan.

A tiny pure accounting example may be used during implementation to reason about arithmetic, explicitly synthetic and never stored in the user engine: two agents working simultaneously for ten minutes produce twenty agent-minutes; a third five-minute graph node adds five, not its run's wall-clock again. One session used on two days remains one weekly session. A wait interval overlapping another wait is subtracted once. This is a calculation aid, not real telemetry or a test-suite request.

### Human acceptance route (execute later with explicit user participation)

1. **Fresh/old engine:** inspect a fresh development data directory and a preserved older state. Fresh data shows honest empty/unknown metrics, no Atlas/KLM fixture entities, 2026 anchor or sample posts. Old Session.Usage does not become a spike on installation; old graph/inventory evidence has distinct coverage.
2. **Main + side + independent session:** run small ordinary work in two projects, including an unused created session, and a side turn. Confirm created vs used, per-project filtering, agent identity counts and harness rankings. Side work belongs to its main owner/topic; independent spawned-session work belongs to its own real owner. Creating sessions requires separate real authorization.
3. **Idle and concurrency:** let a completed runtime sit idle; hours must stop. Overlap two confirmed executions: both contribute time. A live snapshot is provisional; native completion stops it even if a process/server remains retained. Closing/minimizing chat never changes accounting or stops execution.
4. **Waits:** trigger a real non-YOLO permission/question, leave it pending, resolve/deny/dismiss it; try concurrent known waits. Confirm union subtraction, correct owner and resume. Do not toggle YOLO or answer an elicitation just to simplify a check. A waiting consultation after yielding is not continuously charged to its requester.
5. **Tokens:** inspect reported native totals around two turns for each available harness. Confirm delta normalization, cache/reasoning semantics, duplicate snapshot stability and no lifetime-total spike on resume. Stop an active turn and resume manually; missing final samples remain partial rather than silently complete.
6. **Native children:** only where user-authorized and supported, exercise an OpenCode child including a repeated invocation and a request before task metadata. Child identity counts once; invocation time/usage does not replay. Check Pi unsupported and managed Codex disabled states remain honest. Do not create native children in this planning task.
7. **Graphs:** only under explicit human graph authorization, inspect a run with parallel/continued nodes, a correction/retry and a terminal-only run. Runs count by run ID, executed agents by session ID, graph wall-clock separately. Waiting/sealing has no duplicate node effort; notifications/assessments are not executions.
8. **Persistence:** restart after a completed turn: totals unchanged. In a disposable environment, interrupt an open run/turn, remain offline, restart: offline time is not counted, unknown tail is flagged, queued work is not replayed by Analytics. Do not crash the user's production engine for validation.
9. **Identity lifecycle:** rename/move a topic and verify current-membership historical regrouping; two equal names in different projects remain separate. Archive/restore retains historical figures. Remove/restore a disposable project hides/reveals the same retained records. General-agent work is not assigned to a project; empty registered project list stays usable.
10. **Dates:** check Monday transition, Sunday end, a turn spanning midnight, DST transition in a chosen IANA zone, current-week cutoff and previous comparison, All time beyond 13 weeks, first-covered partial day, zero vs unavailable days. Weekday bars, calendar and summaries use identical time boundaries.
11. **Read lifecycle:** switch projects/weeks rapidly, disconnect/reconnect engine, hide Analytics, return/focus and cross midnight. Obsolete responses cannot overwrite the current filter; hidden polling stops; Retry does not replace errors with fixtures.
12. **Visual regression by user:** native desktop, browser Focus, narrow browser/mobile host, light/dark; Home centered with 240px preferred menu/160–360 bounds, posts max800px and side gutters/shadows only above800px, right Happening now, search/refresh/settings without Activity heading. Analytics retains Overview-only KPIs, full-width cards/tabs and no redundant title. Verify wide-Home launcher right88px/dock right14px, identical Close/Minimize bottom bar, drafts/queue/Stop/permissions/overlay routing, real feed, independent Focus widgets/windows and Back to hosts bridge unchanged.

## 10. Planning evidence and completion criteria

Planning inspected `AGENTS.md`, all of `product.md`, the prototype document, the Analytics frontend and Home integration, plus the execution/usage, requests, graph, inventory and persistence code referenced above. No harness tasks, graphs, tests, builds, sessions, subagents, commits or pushes were run by the planner. Source reading and working-tree inspection are evidence for this plan, not runtime validation.

Implementation is ready for originating-agent review/manual testing when all production fixture imports and sample Home inclusion are gone, normalized evidence persists atomically, every approved Analytics surface reads the single real API with honest coverage, checks for the changed languages compile, and the approved layout/chat/feed boundaries remain intact. Review/corrections are a later authorized workflow; successful compilation alone is not feature acceptance.
