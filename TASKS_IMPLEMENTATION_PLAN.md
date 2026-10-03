# Tasks implementation plan

Updated: incorporating the user question-round decisions and subsequent local-layout
alignment confirmations (following the 2026-10-02 revision).
Status: planning; automation is not implemented. This update changes the plan only,
not the mock screens or engine. Remaining mock adjustments are recorded below.

This document records the approved Tasks decisions and proposes small implementation
slices, following `AGENTS.md` and `product.md`. **Approved** sections describe the
user's product contract. **Proposed** sections describe implementation options, not
additional product approval. Open decisions must be resolved before their dependent
slice ships. The Tasks/Telegram UI and its feature README are maintained separately;
local prototype behavior is not evidence of engine support or product approval.

## 1. Goal and approved contracts

Tasks are project-scoped automation launched manually, by schedule, or by webhook.
The first real scenario is daily Grafana log analysis: a graph reports errors and
warnings, then the main-chat orchestrator evaluates the report, invokes an authorized
correction graph, or asks the user for guidance/planning.

```text
manual / schedule / webhook trigger
  -> new normal main session (orchestrator)
  -> authorized graph -> report
  -> orchestrator assessment -> correction / user question / recorded outcome
```

The scheduler and webhook receiver start the orchestrator, never bypass it to invoke
a graph. Harnesses continue to execute agent work; Tasks does not replace them.
Graph completion or the end of a chat turn alone does not establish Task success.

### Approved operations UI

- Tasks remains below **New session**. Clicking a task entry opens its
  **details/operations modal**, not the editor. The list has a status indicator,
  not an enable toggle.
- Details opens on the **Details** tab; **Runs** contains the queue and history.
  **Edit** and **Run now** appear in Details. Enable/disable and delete live in the
  header overflow, accessible from either tab; delete retains confirmation.
  **Run now works for all trigger types, even disabled tasks**. Required input for
  a manual launch of a webhook task still needs a UI contract.
- New tasks are **always created enabled**, without an Enabled control in creation.
  Existing tasks retain enable/disable controls, including in editing. Once the
  relevant automation is implemented, saving a new task makes it eligible for new
  scheduled/webhook triggers; creation itself is not an immediate execution.
- Disable only prevents new automatic triggers. It does not stop active runs or
  discard already accepted queued runs. Manual admission remains available.
- Block deletion while any run is active or queued, including waiting for a user.
  Once deletion is allowed, preserve previous run results and their normal sessions.
- Details shows the trigger, orchestrator, run policy, session folder, question
  routing and Telegram link status; collapsible instructions and allowed graphs.
  Runs separates the **queued runs in FIFO order** from the **latest five
  non-queued runs**. Queue entries do not consume the five-entry history limit.
  Rows show accepted date/time and **actual trigger origin**, plus execution status,
  duration and expandable short result/error where applicable, and **Open session**.
  Captured run settings remain inspectable from expanded history and the run chat.
- Keep all runs awaiting answers accessible independently of the latest-five
  non-queued history limit. They may appear as additional rows in Runs; a separate
  callout is not required. Do not lose an older waiting run when newer parallel
  runs enter history, or duplicate a waiting run already among the latest five.
- List summaries expose active work or the last failure/interruption without loading
  every run; successful last-run status and Telegram status are omitted from the
  list. Enabled/Disabled remains muted status text, not a list toggle.
  Show the next firing in Details for enabled scheduled tasks, computed by the
  engine. Runtime summaries and next firing must not be inferred from a fixture
  or session idle state.
- Display execution duration excluding time blocked waiting for the user, and show
  user-wait duration explicitly as a separate metric. See the timing model below.
- Edit replaces details rather than stacking a second modal. Save and Cancel return
  to details. Deletion confirmation remains explicit. The editor groups fields into
  **Task**, **Orchestrator** and **Trigger**, preserving one draft across sections:
  Task contains instructions/folder/graphs, Orchestrator includes Telegram routing,
  and Trigger includes schedule and parallel-run policy. Hidden sections still
  participate in validation. Keep Create/Save and Cancel in the footer.
- **Open session** closes the modal and navigates to the normal chat. A task reference
  in the chat header returns to the task.
- Keep UI copy in English and reuse the existing design-system components/tokens.

### Approved session and history identity

- **Every Task Run creates a new, independent, normal top-level main session.**
  The normal sidebar and task run history reference the **same session ID**.
  There is no hidden task-only chat and no shared persistent conversation per task.
- The default visual session folder is `Tasks`. Reuse it when present; otherwise
  create it when creating the first run's session. Saving a task creates no folder.
  A task can choose another session folder. Preserve existing `Ungrouped` semantics.
- A session folder is visual grouping (`Session.Workspace`), not an execution
  directory or Git worktree. Task execution still belongs to the engine computer.
- The session retains orchestration, reports, questions, and answers, including
  Telegram answers. Users can continue the conversation after execution ends.
  Later chat neither rewrites the recorded run outcome nor automatically creates
  another Task Run.
- Graph node sessions remain private and accessible through the existing graph UI.

### Approved permission/question separation

Tasks default to YOLO so tool permission cards do not block execution. Model
questions remain interactive. Telegram connects to the engine to route **questions**,
not permission cards. YOLO does not waive graph authorization, lifecycle settlement,
objective assessment, or workspace/retry boundaries. YOLO **remains enabled** in
the normal session after a run ends; the user can change it with the normal chat
control. Keeping YOLO does not keep a finished run's automation authority active.
Preserve this existing default/persistence rule rather than adding a Task-specific
lock: being queued, executing or awaiting an answer does not by itself disable the
normal YOLO control. Normal chat's between-turn applicability still governs; this
is not approval for changing a running harness's permissions mid-turn.

### Approved configuration, authorization and admission

- Choose the orchestrator's harness, model and effort **per task** in its editor.
  Graph agents retain their independently configured harness/model/effort.
- Saving a task records recurring authorization to invoke its **selected graphs**
  only to fulfill the saved instructions, including correction graphs. Make that
  consequence explicit in the save UI. An empty selection authorizes no graphs;
  external payloads cannot expand authority. This is a Task grant, not a fabricated
  human message in each new session.
- Capture task configuration and the applicable grant **when the trigger is
  accepted**, not when the queued run starts. Edits affect later triggers only;
  active and queued runs retain their captured instructions/model/settings.
- **Allow parallel runs** is a per-task option, **off by default**. When off,
  accepted triggers enter FIFO and one run executes at a time. Waiting for a user
  retains that slot. Runs are independent: a failure does not block the next run.
  When on, runs may execute concurrently in distinct normal sessions. Cross-task
  resource limits and workspace collision policies remain open.
- Disabling leaves both active work and the accepted queue intact. The approval
  does not define special emergency revocation, changes to graph definitions, or
  the transition when parallelism is edited with runs already active/queued.

### Approved completion, timing and recovery

- The orchestrator explicitly records a final result after evaluating its graphs.
  A normal message/turn ending is not Task completion. The engine also records
  infrastructure failure/interruption. Exact tool schema/state names remain design
  work; the need for explicit completion is approved.
- User questions **do not expire automatically**. A run waits until answered or
  cancelled, or until its native execution fails/is interrupted. Persisting a
  question cannot make its in-memory callback survive a restart.
- Execution duration excludes time blocked waiting for user input. Expose that
  user-wait time separately so long unattended waits are visible as bottlenecks.
- On engine restart, mark already-started runs **Interrupted**, preserve history
  and artifacts, and do not automatically replay them. Known-not-started accepted
  queue entries may continue after reconciliation.

### Approved schedule and engine timezone

- The engine owns a **global persisted timezone**, selectable in Settings. All
  clients use this timezone for clock/date presentation and scheduling, rather
  than their device timezone. Tasks select days/time, not their own timezone.
- Cover the mock's chosen weekdays and one time of day; do not reduce that editor
  to daily-only execution. Arbitrary cron is not an approved requirement.
- **Recover missed runs** is per task and **off by default**. When off, skip missed
  scheduled firings and continue with the next future firing. When on, recover
  **only the most recent missed firing**, through normal admission/queue rules.
  This is missed-trigger recovery, not resumption of an interrupted run.
- Initial timezone selection, DST gaps/overlaps and behavior when the global
  timezone changes still require decisions. Persist instants independently of
  display timezone; changing presentation must not rewrite historical instants.

### Approved Telegram behavior and deferred security work

- Questions are **always available in KLM**, not only as a fallback when Telegram
  fails. **Route questions to Telegram** is an additional per-task routing toggle,
  **off for new tasks**. Off means KLM only; on means KLM plus Telegram, subject to
  delivery availability. This is not an exclusive question-channel selector.
- Capture the routing preference at trigger acceptance with the other task settings;
  editing it affects later accepted runs, not active/queued snapshots. Turning it
  off is not the same operation as unlinking the engine-global Telegram recipient.
- Both interfaces resolve the **same question**: only one answer may be accepted,
  with resolution/provenance reflected in KLM and stale Telegram replies rejected.
  This retains the shared resolver/race contract below, not two independent requests.
- One bot and one recipient account per engine, using outbound HTTPS **long
  polling**. The link is engine-wide, not browser-local. Verify pairing through
  an actual bot interaction; a typed username is not proof of identity.
- If unlinked/offline, keep the question available in KLM and expose the delivery
  problem. Retry Telegram delivery when possible only while that exact question
  remains valid. Do not block task admission merely because Telegram is unavailable.
- For this first version, the user explicitly permits **secret-marked questions
  and their answers through Telegram too**. Do not substitute a KLM-only policy.
  Bots have no end-to-end encryption; this temporary tradeoff is accepted.
- **Deferred security backlog:** revisit confidential-question transport, redaction,
  retention and a safer alternative to Telegram for secrets. This does not waive
  paired-user verification, bot-token protection, or the existing masking of secret
  answers in ordinary KLM history/logs. Do not claim end-to-end secret protection.

## 2. Current implementation and gaps

Source observations below describe the working tree inspected for this plan; it
contains concurrent prototype edits and unvalidated runtime capabilities.

| Area | Existing integration point | Missing for Tasks |
| --- | --- | --- |
| Prototype | `clients/desktop/src/features/tasks/prototype.ts`, `TasksPage.tsx`, `TaskDetailsDialog.tsx`, `TaskEditorDialog.tsx`, `TaskGraphAllowlist.tsx`, `TaskRunSnapshot.tsx`, `TaskSessionChat.tsx`; `App.tsx` owns local state. Sectioned editor, Details/Runs tabs, FIFO queue display, inspectable snapshots and all-types/disabled-task Run now are present; Telegram Settings simulates bot interaction | Durable CRUD, real runs/triggers, verified pairing and engine-backed navigation. Mock Run now only creates queued local records. Preserve older waiting rows beyond latest-five history and remove the extra Task-state YOLO lock. Replace the separate local chat adapter rather than persisting another chat type. |
| Sessions/folders | `engine/api.go` uses `newTopLevelSession` / `appendTopLevelSession`; `workspace()` accepts existing active folders or `Ungrouped` | Atomic Task Run + normal session + lazy folder creation; durable task backlink. Ordinary session creation does not create missing folders. |
| Persistence | `engine/store.go`, `transactions.go`, `stream_journal.go`; projects, sessions and graph records already persist | Task definitions/revisions, trigger receipts, run records, scoped authorization and delivery provenance. Extend existing persistence, not a separate chat store. |
| Graph execution | `graph_tools.go`, `graph_invocation_contract.go`, `graph_scheduler.go`, `graph_runtime.go`, `graph_records.go`; activities, assessments, result notifications, private node sessions | A deliberate standing Task authorization contract and Task Run association. Existing scheduler orders activities/dependencies; it is **not a time scheduler**. |
| YOLO | `Session.YOLO` in `store.go`, consumed by the adapter in `harness.go`; native adapters implement harness-specific policy | `graph_runtime.go` creates graph-node sessions without copying YOLO. Task execution policy must propagate explicitly, including continued nodes and corrections. |
| Questions | `questions.go` validates replies and serializes in-flight resolution; `graph_records.go` projects node questions into the owner chat | Shared KLM/Telegram resolver, durable exact-origin correlation, Telegram transport, answer provenance and recovery. Callback functions live in `turn.questions`; persisting a question does not persist its callback. |
| Recovery | `main.go` clears pending questions on restart; graph recovery records interruption without execution replay; message queue records uncertain delivery | Task-level recovery and revocation rules, durable trigger/delivery deduplication, explicit interruption/uncertainty in details. |
| Frontend updates | `clients/desktop/src/engine.ts`, engine `history.go` / `updates.go`, shared `platform.ts`, sidebar and chat routing in `App.tsx` | Task API types, bounded history plus independent active/waiting queries, backlinks, live list/details updates, run timing and header navigation. |
| Settings/time | `SettingsDialog.tsx`, `EngineTimeSettings.tsx`; shared App-local timezone selection used for prototype dates. Per-task orchestrator/parallelism/missed-run controls already exist; schedule uses weekday ToggleGroup and a 24-hour, minute-resolution TimePicker | Persist and expose the global engine timezone; unify relevant date/time formatting across clients. Replace local controls/fixture model catalog with real persisted settings and capabilities. UTC initial value and invalidated fixture next-firing timestamps do not establish production defaults or timezone-change rules. |

### Remaining mock alignment (not implemented by this plan update)

- `TaskDetailsDialog.tsx` currently uses only the latest five non-queued history
  entries. Separating queued entries does not protect an older waiting run from
  newer parallel runs. Retain access to every waiting run beyond that limit, using
  rows in Runs if desired; do not restore a mandatory separate callout.
- `TaskSessionChat.tsx` currently disables YOLO for every active/queued run through
  `runIsActive`. Remove this extra mock lock; preserve the approved default-on and
  post-run persistence contract, without overriding normal between-turn controls.
- The feature README describes the current mock, including these limitations; it
  is not authority to replace the newly confirmed contracts. Update it alongside
  the eventual mock corrections. No mock or README changes are made in this update.
- Local-only chat/queue state, frozen fixture metrics, simulated Telegram controls,
  dummy credentials, and absent post-edit next-firing calculation remain prototype
  boundaries, not accepted production behavior. Native subagent tracking, installer
  graph configuration and graph-canvas layout edits do not alter the Tasks contract.

The public API on 7331 is currently unauthenticated under the personal-use contract.
The private authenticated harness bridge is loopback-only. Do not expose the entire
7331 API to the internet to receive webhooks. Graph adapters still require human
runtime validation; Tasks must not imply stronger remote execution guarantees.

## 3. Proposed data and execution model

These are implementation shapes, not finalized wire schemas or lifecycle names.

```text
Project -> Task -> TaskRun -> one normal Session
                    |          -> GraphActivity -> GraphRun -> private node sessions
                    +-> trigger receipt, captured task settings, policy/authorization
                    +-> recorded outcome and question/answer provenance
```

| Record | Minimum information / constraint |
| --- | --- |
| Task | ID, project ID, revision, name/description/instructions, enabled, chosen weekdays/time or other trigger config, session folder, selected graph IDs, optional Telegram routing (default off; KLM always available), orchestrator harness/model/effort, YOLO policy, `allowParallelRuns=false`, `recoverMissedRuns=false`, recurring grant/revision reference and timestamps. No per-task timezone selector. |
| Engine settings | Persisted global timezone and Telegram bot/link configuration; expose nonsecret connection/pairing state. Clients use engine timezone, never silently override it from their device. |
| Trigger receipt | Task/revision reference, actual origin (`manual`, `schedule`, `webhook`), stable operation/delivery/occurrence key, received time, bounded external-data reference, admission result and run ID. A receipt is not itself authority. |
| TaskRun | ID, task/project IDs, immutable trigger-time settings/grant snapshot, exactly one new main session ID once admitted, actual trigger origin, received/scheduled/start/end times, execution policy, queue order, state, terminal summary/error and graph activity associations. Persist user-blocked intervals/timing separately. Keep Task Run IDs distinct from GraphRun IDs. |
| Session backlink | Task ID and Task Run ID in the normal session/history summary; normal top-level role, no graph-node role or hidden parent. This can be projected from the run relation rather than duplicating identity. |
| Question/delivery record | Task Run, owner session, exact target native session/turn/question, optional graph run/activation, paired channel/link identity, delivery state, accepted answer provenance and resolution state. Permit secret content on the approved Telegram delivery path, but never return bot tokens or unmask private replies in ordinary history/status projections. |

Use existing durable transactions for admission: deduplicate the trigger, validate
the current task/project, resolve or lazily create the destination folder, and
record the run/session/dispatch intent together **before** starting a harness.
Retries of the same admission return the same run/session; a new accepted Task Run
gets a new session. Rejected receipts need not create a run or folder. Dispatch
uses the captured revision even after task edits or disabling; do not reapply the
current enabled flag to an already accepted queue entry. Validate resource/native
availability and run authority at dispatch boundaries without replacing its snapshot.
Apply FIFO independently of predecessor success. Bounded queue/resource admission
and transitions between concurrency modes still need a contract.

Do not clone previous run history into the new session. Supply self-contained task
instructions, captured configuration and attributed trigger data. Scheduled/webhook
input must not be fabricated as a human chat message to satisfy graph authorization.
External payloads, including PR text, are untrusted data, never authority or trusted
instructions. Existing normal session creation helpers are reusable; the
agent-facing `session_spawn` authorization contract is not a timer launch API.

### Authorization and execution-policy boundary

- Current `validateGraphAuthorization` requires real user-event references in the
  invoking main conversation. Implement a distinguishable Task grant recorded by
  the approved save action, scoped to selected graphs and saved instructions.
  Capture it with each accepted trigger. Selection alone in an unsaved form, or
  graph selection in ordinary chat, remains non-authorizing. Empty means no graphs.
  Wire representation, audit provenance and explicit emergency revocation still
  need design; ordinary edits/disabling must not silently revoke accepted work.
- Proposed enforcement points: trigger admission, orchestrator graph invocation,
  actual graph start, and corrective retry. Restrict the grant to the active run,
  approved objective/graphs and workspace policy. Keep normal chat's existing
  user-event authorization path intact; never invent event IDs or weaken it globally.
- Capture an effective `TaskRun` execution policy (YOLO by default) and carry it
  through orchestrator turns, graph activity/run records, new/continued Agent and
  Join sessions, and corrective attempts. Check all three adapters. Do not rely
  solely on copying a flag to the first main session.
- Scope task tools, grant lookup and automatic continuations to the active Task Run.
  After termination, the same session is ordinary manual chat: historical task
  instructions or a backlink cannot reactivate the grant or mutate its outcome.
  Keep YOLO enabled for subsequent manual turns as approved; normal chat controls
  may change it. That flag does not restore the expired Task grant.
- Preserve existing one-active-graph-per-conversation, Choice settlement, report
  assessment and fresh/reuse requirements until a Tasks-specific change is approved.
  New sessions per run do **not** prevent concurrent writes to shared project files.

### Lifecycle and explicit finish

Explicit final-result recording is approved. Implement an idempotent, run-scoped
finish operation with report and assessment references; its exact schema/tool name
and settlement protocol remain proposed. Do not finalize success while required
work/questions remain unresolved or abandon pending effects by merely finishing a
turn. Native settlement must precede releasing the serial slot. Normal turn end
while waiting for a graph allows its result notification to resume this orchestrator.
A missing finish must become an observable recoverable condition, not false success;
its fallback policy is still open. Late results must not reopen terminal runs.

Proposed state-to-UI mapping (names are not a finalized wire schema):

| Execution phase/outcome | UI | Timing / behavior |
| --- | --- | --- |
| queued | Queued | Requested time recorded; not started, no execution duration. |
| starting/running | Starting / Running | Track actual start; task slot retained. |
| waiting_graph | Running | Graph work is part of execution, not automatically user wait. |
| waiting_question | Waiting for answer | Show pending request and Open session; no answer timeout. |
| succeeded | Succeeded | Explicit assessed result and frozen end time/metrics. |
| failed | Failed | Reason/result retained; next independent queued run may start. |
| interrupted | Interrupted | Restart/failure interruption retained; no automatic replay. |
| cancelling/cancelled (pending contract) | Cancelling / Cancelled | Do not imply current chat Stop already cancels the graph. |

Never infer Task state from normal session `idle/running/error`. A finished Task
session may be actively chatting without changing its recorded run status.

Proposed timing projection implementing the approved metrics:
- Persist requested, started and ended instants. Expose queue wait separately from
  elapsed time after start. A not-started run has no execution duration.
- Execution duration = elapsed time after start minus intervals blocked on user
  input. Expose user-wait duration alongside it; optionally expose total elapsed
  for reconciliation. These are wall-clock metrics, not CPU or token metrics.
- Persist blocked-interval transitions through question creation, reply delivery,
  dismissal, expiry from native termination, and run interruption. Do not double
  count overlapping questions. Live metrics use a defined observation time;
  terminal metrics stay frozen through later chat.
- Parallel graph branches can continue while another branch awaits a user. The
  exact definition of whole-run blockage versus partial question wait remains to
  be resolved; do not subtract each question's full lifetime independently.

## 4. Proposed delivery slices

Each slice ends with a focused human acceptance check. Resolve only its blocking
decisions first; do not build all trigger transports or a generic workflow framework
before a manual end-to-end run works.

### A. Persist task configuration and operations UI

1. Add Task records/validation and project-scoped CRUD using existing state commits.
   Candidate routes: `GET/POST /api/projects/{id}/tasks` and
   `GET/PATCH/DELETE /api/projects/{id}/tasks/{taskId}`. Use revision checks on edits.
   Enforce delete refusal while runs are active/queued atomically with admission;
   preserve historical run/session provenance when deletion is allowed.
2. Replace prototype state with typed engine requests in `engine.ts`; wire the
   approved Details/Runs and sectioned editor/delete flows using existing components.
   Keep graph catalog lookup, missing-graph errors and Telegram link state recoverable.
   Wire existing per-task harness/model/effort, Telegram routing, Allow parallel runs
   and Recover missed runs controls plus explicit save authorization copy to real
   state. Match frontend/backend validation across all editor sections; create new
   definitions enabled. Wire engine timezone Settings and shared date/time formatting;
   do not add a task timezone picker. Preserve drafts when Telegram Settings overlays
   the editor.
3. Saving/enabling a definition alone must not pretend a scheduler exists. Ship
   functional launch controls only as their execution slice becomes available.

Acceptance: reload to retain definitions; list click opens details; no list toggle;
Edit replaces details and Save/Cancel returns; delete cancellation preserves data;
saving with `Tasks` selected creates no sidebar folder. Empty history is real.
Verify Telegram routing, parallelism and missed-run recovery default off, new tasks
are created enabled without immediately launching a run, orchestrator selections
survive reload, and every client reflects the same engine timezone. Reject duplicate/invalid task input on
both UI and API paths. Deletion while queued/waiting/running is recoverably blocked.

### B. One manual Task Run, normal session, KLM questions

**Gate:** finalize run finish/settlement, Task-grant wire integration, cancellation
and workspace/retry contracts. Eligibility, per-task orchestrator selection,
trigger-time capture and recurring save authorization are now approved.
Support Run now for all task types regardless of enabled state. Before launching
webhook tasks manually, settle the input form/validation; never fabricate an event.

1. Candidate admission endpoint: `POST /api/projects/{id}/tasks/{taskId}/runs`,
   with an operation ID; return accepted run/session identity, not success. Add
   bounded latest-five **non-queued** history and single-run detail projections
   (candidate `GET .../runs?limit=5` / `GET .../runs/{runId}`; filtering contract is
   still implementation design), plus independently fetched/projected active/queued/
   waiting work. Queue entries are displayed FIFO without consuming history slots.
   Older pending questions remain reachable outside the five history entries; avoid
   duplicating waiting rows already present there.
2. Implement atomic admission and launch through normal main-session execution.
   Add run-scoped context to bridge/prompts and explicit execution-policy propagation
   in `graph_tools.go`, `graph_scheduler.go`, `graph_runtime.go` and adapters.
3. Link existing graph result notifications/assessments to the active Task Run.
   Record terminal outcome separately from future chat. Keep graph node activity
   in the existing graph UI.
4. Implement per-task FIFO/parallel dispatch. Use accepted snapshots, retain queued
   work across edits/disabling, and release FIFO successors after terminal settlement
   regardless of predecessor success. No automatic timeout for user waits.
5. Publish list summaries and live details through the current update mechanism:
   separate FIFO queue, latest five non-queued history entries, independent active/
   pending projections, expanded reports and captured settings, actual origin,
   execution/user-wait metrics and session links. Waiting rows remain accessible
   beyond the history limit; a dedicated callout is not required.
6. Replace `TaskSessionChat` and local-only sidebar/session branches in `App.tsx`
   with the normal engine-backed session route, composer, history, question cards,
   graph view, model controls and session actions. Keep only Task metadata/backlink
   as the additional integration. Never persist mock input as fabricated user
   consent; append post-run manual messages through ordinary session APIs.

Acceptance: manually run Grafana analysis; see a report and orchestrator assessment,
then a permitted correction or KLM question. YOLO avoids tool permission waits in
both orchestrator and graph nodes but never answers the question. Two runs produce
two normal sidebar sessions in one lazily created `Tasks` folder. Run history and
sidebar open identical sessions; header returns to task. Another selected folder
and `Ungrouped` behave normally. Later chat leaves the finished run unchanged and
YOLO enabled. Do not add a Task-state lock to normal between-turn YOLO controls.
Verify manual launch while disabled, each trigger type, FIFO after failure, parallel
independent runs, captured queued settings after edit/disable, and delete refusal
with pending work. Queued runs do not consume the five non-queued history slots;
an older waiting run remains accessible even after five newer parallel runs enter
history. User-wait time is separate from execution and queue time.

### C. Route questions through Telegram

1. Extract a transport-independent question resolver from `questions.go`; keep its
   validation/cancel/multi-select semantics for both HTTP/KLM and Telegram. Route
   node replies to the exact originating native request, not the owner session's
   unrelated question. Surface the transcript/provenance in the normal owner chat.
2. Use approved outbound HTTPS long polling and one bot/recipient per engine, with
   a durable offset/inbox, verified paired chat **and user**, protected bot token,
   and nonsecret Settings status. Exact credential storage/pairing protocol remain
   implementation design. Username entry/Confirm link cannot establish identity.
   Telegram polling is separate from incoming Task webhooks. Detect any existing
   Telegram webhook rather than silently replacing another consumer's setup.
3. Persist correlation and an answer claim before delivery. A KLM/Telegram race
   permits one accepted answer; reject stale/duplicate/wrong-run replies. Record
   accepted, delivered and uncertain separately. A lost native acknowledgment may
   leave delivery uncertain; never claim exactly-once remote execution or blindly
   retry the callback. Advance polling offset only after durable receipt handling.
4. Keep questions available through KLM **regardless of Telegram routing/link state**.
   Send additionally through Telegram only when the run's captured routing preference
   is on. If unlinked/offline, surface delivery failure and retry only while the
   native request is valid; routing off must not enqueue Telegram delivery.
   Send secret-marked questions/answers over Telegram too, as explicitly accepted
   for this version. Retain ordinary KLM private-answer masking and bot-token
   protection. Track confidential transport improvements in the deferred backlog,
   not as a KLM-only gate for this release. Unsupported native form behavior still
   needs an explicit contract.
5. A Telegram answer can be a user decision, but authorization use needs verified
   actor/channel, exact question scope and correct owner-conversation provenance.
   A node answer must not automatically authorize unrelated graphs. Current user
   response events alone do not provide this cross-channel contract.
6. Wire each Settings operation: validate/connect bot, read connection state, begin
   verified pairing, regenerate/cancel code, observe confirmed link, unlink account,
   and disconnect bot. Proposed pairing codes expire and are single-use; regeneration
   or Cancel invalidates previous codes. Bot identity comes from Telegram verification,
   not the entered label. Persist state before reporting Connected/Linked.
   Unlink/disconnect revokes delivery to the old identity; questions stay in KLM.
   Do not readdress queued private deliveries to a replacement account implicitly;
   replacement-account handling is an open decision. Disconnect also clears the link.

Acceptance: exercise all Settings actions, invalid tokens, pairing retries and
expiry, stale codes, and two clients observing one engine-owned connection state.
Confirm routing defaults off and produces KLM-only questions; enabling it retains
KLM questions while adding Telegram delivery. Unlinked/offline tasks still start
and retain KLM questions. Editing routing affects only future accepted snapshots;
unlink/disconnect retains its separate identity-revocation semantics. Deliver a
secret-marked test question via Telegram using dummy values only; ordinary history
retains its existing masking. Open Telegram Settings from the editor and preserve
its unsaved draft.

Acceptance: answer an orchestrator question and a graph-node question through Telegram;
inspect both in normal chat. Reply simultaneously in KLM and Telegram, retry an old
reply, disconnect the link and stop/restart while waiting. Show recoverable state;
do not deliver to a replacement question or recreate an expired native callback.

### D. Weekday/time scheduling and global timezone

**Gate:** DST gaps/overlaps, initial global timezone and changes to it, cross-task
resource policy and precise missed-occurrence reconciliation. Per-task FIFO/parallel
behavior, latest-only optional catch-up and interrupted-run non-replay are approved.

Implement selected weekdays plus one time of day in the persisted engine timezone.
Use durable occurrence keys and authoritative next-firing calculation through the
same admission/orchestrator path as manual launch, not `graph_invoke`. Persist the
scheduled instant separately from acceptance and actual start. Recalculate future
firings after configuration changes; do not rewrite already accepted snapshots.
Show the next firing only for enabled scheduled tasks. No browser timer is authority.

For an unavailable engine, Recover missed runs off skips old firings; on admits
only the latest missed one, respecting queue/parallel dispatch. Keep this separate
from restarting an already-started run, which stays Interrupted. Catch-up must not
invent eligibility for periods when a task was disabled or did not yet exist.
The snapshot rule for a recovered occurrence versus acceptance time needs explicit
reconciliation where configuration changed before recovery.

Acceptance: selected weekdays (including weekends only/all seven days) fire with
clients closed. Different browser/desktop/mobile timezones show the same engine
schedule and dates. Editing days/time refreshes next firing. Exercise missed-run
recovery on/off, several missed occurrences producing at most one recovery run,
restart with an active interrupted run and an untouched queue, and disable/edit
around a due instant. One occurrence creates at most one admitted run/session.

### E. Authenticated webhook admission

**Gate:** ingress deployment/authentication, event/filter scope, replay/dedup policy,
payload retention, manual webhook input, and cross-task resource limits. Reuse the
approved per-task queue/parallel and interrupted-run recovery contracts.

Propose an isolated authenticated ingress exposing only Task trigger ingestion,
not the full public engine API. Deployment may use a narrowly routed proxy or a
separate listener; no new port is approved. Validate provider signatures on bounded
raw payloads, verify configured event scope, and durably deduplicate delivery IDs
before acknowledging asynchronous acceptance. Persist dispatch intent first; reject
invalid signatures/oversized input without launching. No network wait on graph
execution. Reuse normal Task admission and keep external data attributed/untrusted.

Acceptance: a signed GitHub PR-merge delivery starts one new orchestrator session;
redelivery returns the existing receipt/run; invalid signature, wrong event and
oversized payload do not launch. Crash after persistence/before acknowledgment
does not duplicate a run. Details reports actual origin: webhook deliveries are
`webhook`, but Run now on a webhook-configured task is `manual`. Disabling refuses
new automatic admission while leaving accepted deliveries and manual launches intact.

## 5. Persistence, migration and recovery

- Extend `store.go` loading, state validation, transaction cloning/commits and journal
  replay together. Current state decoding rejects unknown fields and recognizes
  versions 1/2: define upgrade/older-binary compatibility explicitly, retain a recovery
  copy, and initialize existing installs with no tasks/runs. Do not import sample
  tasks or fake Telegram pairing as production state.
- Keep existing session IDs, folders, graph/native history and `Ungrouped` intact.
  Task configuration edits must not rewrite past captured settings or outcomes.
  Block deletion with active/queued runs and retain old sessions/results after
  allowed deletion. A tombstone or retained task-name snapshot can preserve context;
  exact deleted-task navigation and archive interaction remain open.
- Restart cannot restore a Go callback or prove remote tool completion. Mark active
  Task Runs Interrupted, invalidate outstanding question routing, and retain graph
  results/artifacts and uncertain deliveries. No automatic replay of possibly
  started work. Expire old Telegram reply controls; unanswered questions have no
  normal timeout but cannot outlive their native request.
- Reconcile accepted queue/dispatch intents separately from started runs. Continue
  known-not-started entries with captured settings; ambiguous native delivery must
  not be blindly replayed. Persist dedup keys and notification IDs across failures.
  Missed-schedule catch-up is a distinct latest-only opt-in admission path.
- Guard late graph notifications and resumed normal chat with active-run identity;
  preserve historical evidence without restoring expired task authority. Existing
  chat Stop does not stop an independently running graph; do not label it Task
  cancellation without defining the broader lifecycle.
- Validate missing/deleted graphs, removed projects and unavailable harness/models
  at launch. Proposed graph-reference handling must account for IDs changing on
  catalog rename: update live task references consistently without rewriting old
  captured run evidence or widening authority. Decide the mapping for queued grants.
  Missing/disabled entries stay visible and removable, never silently replaced.
- Keep the task's configured destination for future sessions separate from the
  current folder of a session moved by the user. Archived/renamed folders need an
  explicit policy; never create duplicate `Tasks` folders to evade conflicts.

## 6. Validation and delivery policy

This assignment changes documentation only; no build, runtime validation or tests
are required or claimed. During implementation, use the smallest relevant syntax,
type or build check, followed by the slice's manual acceptance path. Full suites,
integration tests and adversarial/review-agent work require an explicit request.
The acceptance paths above describe future checks, not completed verification.

## 7. UI-to-runtime coverage checklist

This matrix maps approved interactions and current prototype controls to real
integration work. Proposed implementation details are not additional product
approvals. Endpoint names beyond the candidate routes above remain flexible.

| UI element/action | Real data / operation | Failure and acceptance contract |
| --- | --- | --- |
| Task list, search/filter, status | Project-scoped persisted tasks, current work/last-failure summary and muted Enabled/Disabled status | Omit successful last-run and Telegram status from list. No sample rows imported; reload/other clients agree. Connection failure is not an empty list. |
| Create / Save | Validated fields, task revision and recurring grant saved together; new tasks always enabled | No Enabled control in creation and no immediate run on save. Preserve draft across Task/Orchestrator/Trigger sections and on validation/conflict/network error. Match trimmed name/instructions, optional description, case-insensitive duplicate-name check and limits (currently 80/240/12000 characters); confirm API validation units before implementation. |
| Orchestrator selectors | Per-task real harness/model/effort catalog | Missing models are recoverable errors, not silent substitutions; graphs retain their own agent settings. |
| Enable / Disable | Persist enabled flag for future automatic admission | No cancellation of accepted work; Run now still available. Reflect errors without false success. |
| Allow parallel runs | Per-task persisted flag, default off, dispatcher policy | FIFO after failed predecessor; user waits retain serial slot. Mode-switch edge policy still open. |
| Run now | Idempotent manual admission for every task type, enabled or disabled | Return admitted session/run, not success. Manual webhook input remains to be designed. |
| Details / Runs tabs | Default Details configuration/actions; Runs has separate FIFO queue, latest five non-queued entries, independent active/waiting projections, on-demand report and captured settings | Queue does not consume history slots. Every older waiting run stays reachable without duplicate rows; no mandatory callout. Graph success alone is not Task success. |
| Execution / user-wait timing | Durable timestamps/blocked intervals, live observation time | No negative/double-counted durations; no duration before start; post-run chat cannot change metrics. |
| Open session / task backlink | Normal session ID and preserved task/run relation | Remove special mock chat path; use normal history, composer, graph and question handling. Archived/deleted navigation must be explicit. |
| Session folder | Task destination vs session's actual current folder | Lazy default Tasks creation; moving one chat changes neither old run identity nor future task destination. |
| Allowed graphs | Live catalog plus saved IDs; captured grant on admission | Preserve missing/disabled selections for removal, show catalog errors, handle rename without widening grants. |
| Delete | Atomic check for no active/queued work; retained history | Refuse deletion with a reason while pending; Cancel is non-mutating; preserve previous sessions/results. |
| Days / time / next firing | Engine schedule and global timezone | Cover chosen weekdays; recalculate next firing after edits; no browser-zone authority. |
| Recover missed runs | Persisted per-task flag, default off | On recovers latest missed firing only; never replays an interrupted execution. |
| Settings timezone | Global engine preference read/write | All clients display same zone; change/DST semantics must be settled, not guessed. |
| Telegram routing / Link Telegram | Per-task opt-in routing flag, default off, captured on admission; engine-global link state | KLM always shows questions; on adds Telegram rather than replacing KLM. One accepted answer across both interfaces. Preserve draft while Settings opens. Unlinked/offline does not block task admission. |
| Connect bot | Protected token submission, verified bot identity | Invalid token/network failures retain recoverable form; never echo token. Connected only after verification. |
| Link account / New code / Cancel | Engine-backed pairing lifecycle | Proposed expiring single-use code; old codes invalidated on regeneration/cancel. Typed username is not identity proof. |
| Pairing confirmation | Verified bot interaction, paired user/chat IDs | Replace the explicit Simulate bot interaction mock action with verified engine status; no production user button or typed username establishes identity. Clients observe the same persisted link. |
| Unlink / Disconnect | Revoke paired destination; disconnect also clears bot/link | Retain task routing preference and KLM question; prevent stale replies/delivery to revoked identity. Pending delivery to a new recipient needs a decision. |

## 8. Remaining open decisions (not approved requirements)

Do not reopen decisions in section 1. The following are narrower unresolved edges
or design contracts, not reasons to keep the approved core behavior provisional.

| Decision/design contract | Resolve before |
| --- | --- |
| Manual input for webhook tasks; ingress deployment, endpoint/signatures, filters, replay window/dedup retention and payload storage | Webhook launch/admission. Run now eligibility is approved; a new ingress port is not. |
| Task-wide cancellation across orchestrator and graph, queued cancellation, late results, exact finish schema/settlement and missing-finish recovery | First run lifecycle. Explicit completion and indefinite question wait are approved; chat Stop is not yet whole-run cancellation. |
| Task grant wire/audit contract, explicit emergency revocation, numeric invocation limits, graph changes/rename/disable after capture | Graph-backed execution. Save authorization, empty=no graphs and trigger-time snapshots are approved; no numeric limit is approved. |
| Workspace and corrective retry policy, fresh/reuse decisions and cross-run artifacts | Corrections. Existing default remains original directory; mandatory new_worktree is not approved. |
| Concurrency-mode edits with queued/active runs, bounded queue/resource limits and concurrent writes across different tasks | Dispatch. Per-task optional parallelism/default-off/FIFO independent of failure are approved. |
| Initial global timezone, DST gaps/overlaps, timezone change effects on future/queued/catch-up work; captured revision for recovered missed occurrences | Scheduling/settings. Engine-global timezone and optional latest-only catch-up are approved. |
| Precise user-blocked time when parallel branches still work, multiple pending questions and interruption timing | Timing projection. Exclude actual user-blocked time and display it separately; no double counting. |
| Telegram token storage, pairing protocol details, unsupported forms, identity replacement while requests are pending and verified answer provenance for graph decisions | Telegram transport. One bot/recipient, long polling, KLM fallback and initial secret delivery are approved. |
| Archived/renamed destination folders, archived sessions, project removal and deleted-task backlink destination | Normal session/history integration. Preserve history and refuse task deletion while active/queued. |

### Deferred security backlog

Revisit confidential questions/answers sent via Telegram, including safer transport,
redaction and retention. The first version is explicitly permitted to deliver them;
this backlog must not be misrepresented as an implemented protection or a current
KLM-only restriction. Token protection, verified recipients and existing private
history masking remain baseline requirements.

**Superseded proposals:** shared per-task sessions; hidden task-run chats; manual
launch only when enabled/manual-type; automatic question timeout; per-task/browser
timezone; retrying interrupted runs automatically; KLM-only secret questions for
this first version; exclusive KLM-versus-Telegram question channel; queued entries
consuming the five-entry history limit; a Task-state lock on YOLO beyond normal
between-turn controls. Preserve one normal session per run and all approvals above.
