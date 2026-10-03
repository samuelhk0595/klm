# Tasks and Telegram UI prototype

Aligned with the approved decisions in `TASKS_IMPLEMENTATION_PLAN.md` (2026-10-02).
The root plan remains the authority for product contracts and unresolved decisions.
Tasks sits below New session in each project's sidebar. Telegram and the global
engine time zone are represented in Settings.

## Local-state boundary

- `useTaskPrototype` in `prototype.ts`, owned by `App`, holds project-scoped tasks,
  runs and normal `Session`-shaped local chats. Telegram connection/pairing state
  and the engine timezone selection are shared across projects in that same App.
  All survive section/project changes and closing Settings; reload resets them.
- This is **not engine persistence or synchronization across devices**. Settings
  represents the eventual engine-global scope, not a per-task/browser timezone
  contract. No engine timezone API exists yet. Nothing is written to local storage.
- No task execution, graph dispatch, FIFO worker, clock scheduler, webhook server,
  Telegram networking, native session creation or real session-policy changes are
  implemented. Saving/enabling definitions does not start anything.
- Graph choices still use the existing read-only project catalog. Orchestrator
  controls reuse the authoring prototype's `modelsForHarness` fixtures and the
  existing RadioGroup, HarnessIcon, SearchSelect and Slider patterns. Model names,
  compatibility and initial OpenCode/model/effort values are illustrative, not
  discovered capabilities or execution defaults. Changing them touches only a
  task draft; it never launches a harness or changes a real session's settings.

## Task configuration and operations

- Clicking a row opens details. Enabled/Disabled is muted text at the right of the
  list metadata row. Running and Waiting for answer use badges; successful last-run
  status is omitted from the list, while failed/interrupted results remain visible.
  Telegram status is not shown in the list. Details retains Edit and Run now, with
  Enable/Disable and Delete together in the overflow menu.
  The Details tab opens by default with task configuration, Instructions and Allowed
  graphs (initially collapsed). The Runs tab contains a dedicated Queued runs
  section and the latest-five non-queued history. Waiting for answer uses an amber
   badge on the run row, without a separate callout. Edit and Run now appear only
   inside Details, right-aligned below Allowed graphs. The overflow menu remains
   in the modal header for both tabs;
   changing tabs preserves expanded disclosures within the modal.
- Edit replaces details. Save, Cancel, close and Escape return to details; cancel
  discards the unsaved draft. Creating returns to the list. Opening Telegram
  Settings over the editor retains the complete draft, including model/effort,
  schedule, graph selections and policy toggles.
- The shared creation/editor modal uses a left-hand section menu: **Task** has
  name, description, session folder, instructions and Allowed graphs;
  **Orchestrator** has harness, model, effort and Telegram question routing;
  **Trigger** has the trigger, schedule controls and Allow parallel runs. Switching sections preserves
  one draft. The content panel scrolls independently; Cancel/Create/Save remain
  in the footer. At narrow widths the section menu becomes a scrollable top row.
- New tasks are always created enabled and expose no Enabled control. Editing an
  existing task retains its Enabled draft control. Validation covers the entire
  draft even when a section is hidden.
- Questions always remain available in KLM. **Route questions to Telegram** is an
  optional, right-aligned toggle in the **Telegram routing** section of Orchestrator,
  with the description "Send questions to Telegram while keeping them available in
  KLM." It is off for new tasks; enabling it adds Telegram
  routing rather than choosing an exclusive question channel. Details and captured
  settings show `KLM + Telegram` when on, otherwise `KLM`. The existing local
  `questions` value stores this preference (`telegram` means routing on).
  Link/Manage Telegram remains available when enabled, and opening Settings preserves
  the draft. This UI does not send questions or answers to a harness or Telegram.
- Name and instructions are required; description is optional. Existing limits
  (80/240/12000 characters) and case-insensitive duplicate-name checks remain.
- Choose the orchestrator's harness, model and supported effort per task. Changing
  harness requires selecting a compatible model again. Graph agents' independent
  settings are unaffected.
- The **Parallel runs** section has a right-aligned toggle and a brief description:
  "Allow multiple runs of this task at the same time." **Allow parallel runs**
  defaults off. Off represents FIFO admission; waiting for
  a user retains the serial slot, and failure does not block a later independent
  run. On captures eligibility for independent parallel dispatch. Neither option
  starts local workers. New runs stay Queued with no started time, execution/user-
  wait duration, generated assistant response or invented execution success.
- Schedule retains the Monday–Sunday toggle group and one 24-hour TimePicker;
  at least one day is required. **Recover missed runs** defaults off and describes
  recovery of only the latest missed firing. It is not interrupted-run resumption.
  The mock stores this choice but does not generate recovery runs.
- Saving concisely states that selected Allowed graphs are authorized to fulfill
  the saved instructions. Empty means no graphs. This represents the future
  recurring Task grant, not a real runtime grant or fabricated human consent.
  Disabled/missing saved graph IDs stay visible and removable; catalog errors are
  recoverable. New catalog entries are never selected automatically.
- **Run now works for manual, scheduled and webhook tasks, even when disabled.**
  Every click creates a new local queued run and a new local chat; actual origin
  is always Manual. A webhook task's manual action exercises admission and identity
  only. There is no fabricated webhook event or payload editor; the input contract
  remains unresolved.
- Disabling affects the automatic-trigger preference only. Existing running,
  waiting and queued fixtures remain intact, as do manual launch controls.
- Delete is refused while any run is queued, running or waiting, with the reason
  shown in the confirmation dialog only after selecting Delete. Task details has
  no preemptive disabled-task or deletion-restriction notices. The same guard is used for both menus,
  confirmation and the local state action. Allowed deletion retains all run/session
  records. A retained chat's backlink returns to the Tasks list after deletion;
  this remains a prototype navigation convention, not the final deleted-task route.

## Run snapshots, history and chat identity

- Admission copies instructions, allowed graph IDs, harness/model/effort, folder,
  Telegram-routing preference, policy flags and the global timezone observation. Subsequent
  task edits do not replace these snapshots, even for runs still queued. Captured
  settings are inspectable inside expanded history and in the run's chat.
- Queued runs are all visible in the Runs tab, oldest accepted first, with their
  accepted date/time as the primary row label, actual trigger origin beneath it,
  and Open session action. Latest-run rows use the same date/origin layout. The section
  heading conveys the queued state; rows do not repeat it.
  Expanding a queued row exposes captured settings without redundant admission or
  not-started summary text; queued chats omit that text too.
  They have no execution/user-wait duration and do not consume the latest-five
  history limit. This presentation does not dispatch or advance the local queue.
- The run and sidebar reference the same local session ID and events. `TaskSessionChat`
  remains a simulation adapter around the established ChatMessage/MessageComposer,
  not another persisted chat type. Replacing this adapter with actual normal chat
  routing is future backend work.
- Task instructions are attributed as task input (`agent_prompt`), not forged user
  authorization. Manual chat messages append only to that session, with local
  favorites. They neither resolve fixture questions nor start harness turns.
- Folder defaults to `Tasks`; it is always offered alongside active folders,
  the saved destination and `Ungrouped`, without duplicates. Saving creates no
  folder. Admitted local sessions appear in a local sidebar group, merged with an
  existing folder of the same name. Local-only groups cannot invoke engine folder
  actions or accept real session drops. Moving a chat updates that chat's visual
  folder only, not the captured destination or future task configuration.
- Open session closes details and navigates to the matching sidebar chat. Its
  header backlink returns to task details. Later chat, folder moves and local
  YOLO changes leave the recorded run outcome, timing and snapshot untouched.
- All local sessions start with YOLO on and retain it after terminal outcomes.
  The local composer uses the normal options-menu visual pattern; completed-run
  sessions can change YOLO locally. No real engine/session policy API is called.
  Questions have no automatic timeout, and YOLO never answers them.

### Fixtures and timing

- Morning project review: six completed weekday snapshots, each with 2m 15s
  execution and zero user wait. Details shows only the latest five.
- Prepare release notes: a question with Telegram routing enabled started eight
   minutes before the page-load observation, with 1m 15s execution and 6m 45s user wait. Five newer
   queued runs appear in their own section without displacing execution history.
   The waiting run has an amber badge and Open session action in history; Telegram
   delivery availability is shown in its chat, where the question stays available in KLM.
  Its older completed session is available from the sidebar: 3m 30s total elapsed,
  split into 2m 30s execution and 1m user wait.
- Triage incoming issues: disabled now, but contains previously accepted webhook
  fixtures: a failed run (18s), a later independent running successor (4m at the
  observation), an Interrupted run (42s), and a completed run (1m 36s). Automatic
  fixtures captured the enabled configuration before disabling. Interrupted never
  changes into a replayed run. Run now adds a separate Manual-origin queued run.
- Check release readiness: an enabled manual task whose last run failed after 42s,
  three hours before the page-load observation. Its sample error records an unreadable
  release checklist; details and its linked local chat retain the same outcome.
- Active fixture numbers are frozen at one page-load observation; expanded settings
  show Observed rather than implying a ticking runtime. Terminal metrics freeze
  at the recorded end. Execution subtracts recorded whole-run user-blocked time;
  queue time is never counted as execution. Queued runs have no started duration.
- Runs are independent records. No browser timer advances a queue, synthesizes a
  result or answers a question. Parallel-mode changes save future configuration
  only; redispatch/reconciliation of existing work is deliberately not simulated.

## Global timezone presentation

- Settings → Engine → Global time zone updates one shared local preference. Task
  screens omit repeated timezone labels; the editor offers weekdays/time, never
  a zone picker. Captured timezone data remains in each run's snapshot.
- Task history, expanded timestamps, local sidebar/chat titles and the fixture
  next firing use this preference. Stored ISO instants and run metrics do not
  change when presentation changes. Unrelated app date formatting is untouched.
- UTC is an explicit **fixture initial value**, not the approved production
  initialization rule. The unchanged enabled morning-review fixture has one next
  weekday firing at 09:00 UTC. Schedule edits or any timezone change invalidate
  that next-firing fixture; disabling hides it. New schedules have no computed
  next firing. There is no timezone conversion/scheduling or DST policy invented
  for future, queued or missed occurrences.

## Telegram simulated transitions

The screen models one bot and one recipient for the eventual engine-wide outbound
HTTPS long-polling connection. No requests are made to Telegram, no code/message is
sent, and no external identity is actually verified.

1. **Connect bot** uses a fixed, read-only dummy `fixture-token` field and supplies
   the fixture identity `@klm_fixture_bot`. There is no editable credential to
   accidentally retain or transmit. No token enters application state/persistence.
2. **Link account** creates the current local pairing code and waits for a bot
   interaction. **New code** replaces it; **Cancel** clears it. Pairing state is
   shared with the rest of the prototype and survives closing Settings.
3. **Simulate bot interaction** represents receiving a polling update correlated
   with the current code, with `@fixture_owner` and dummy user/chat identities.
   It consumes that code and shows the receipt/linked state. A typed username is
   never treated as proof of identity. This explicit simulation action is not an
   assertion that a real Telegram verification occurred.
4. **Simulate offline** retains the link but exposes delivery unavailability.
   **Restore connection** returns to the local online state. Pairing cannot receive
   a simulated interaction offline. No actual polling/retry loop is started.
5. **Unlink account** and **Disconnect bot** retain confirmation/cancel paths.
   Unlink clears the recipient; disconnect also clears bot and current pairing.
   Task routing choices and KLM questions remain available. Neither relinking nor
   restoring connection claims to have delivered outstanding questions.

Secret-marked questions/answers are permitted through Telegram by the initial
product contract. This prototype adds no KLM-only restriction. Existing private-
answer masking remains a requirement for backend integration. Confidential
transport, redaction and retention improvements are deferred; no end-to-end secret
protection, credential verification, code expiry or cross-device pairing is claimed.

## Targeted human validation (not yet performed)

1. Create/edit tasks through the Task, Orchestrator and Trigger sections. Confirm
   field values survive section changes. New-task creation has no Enabled control
   and saves an enabled task; editing retains the existing Enabled value/control.
   Select each harness, a model and effort. Confirm both policy toggles default
   off. Exercise weekday keyboard controls, weekends/all days/no
   days, 00:00 and 23:59. Check graph authorization copy appears only with selected graphs, disabled/
   missing graph handling and duplicate-name validation.
2. Open details and check the default Details tab: configuration, instructions and
   graphs, without run history. Switch to Runs for history and waiting-state badges;
   expanded rows/disclosures must remain expanded when switching back and forth.
   Confirm Edit and Run now appear only in Details and the overflow menu is available
   in both tabs, then Edit. Check Save,
   Cancel, close and Escape return to details. Open Telegram Settings from a filled
   unsaved draft, change sections/close, and verify every draft field is preserved.
3. In Settings → Engine, switch UTC to America/Sao_Paulo and Asia/Tokyo. Check all
   prototype dates and titles update across projects without changing instants,
   durations or captured settings. Verify there is no per-task timezone selector.
4. Run now on every trigger type while enabled and disabled. Confirm Manual origin,
   distinct session IDs, Queued/no duration, correct folder and captured settings.
   Edit instructions/model/graphs/folder/policies and run again; compare the old
   and new snapshots. Enabling parallelism must not start or reclassify old runs.
5. Inspect the waiting release-notes row and its 1m 15s / 6m 45s split. Confirm
   all five queued runs appear oldest accepted first, without durations, and can
   expand/open their corresponding sessions. New queued runs must remain visible
   beyond five entries without consuming execution history slots. Verify disabled
   triage retains running work and exposes failed and
   interrupted history. Disable again with queued entries; nothing is discarded.
6. Attempt deletion from both menus for queued/waiting/running tasks; confirm the
   reason and disabled confirmation. Delete the completed-only morning-review
   task (before manually launching it); retained sidebar sessions still open with
   snapshots/results. Cancel must preserve the definition.
7. Open the same run from history and sidebar, follow its backlink, move its chat
   to another existing folder or Ungrouped, and append a message. Confirm independent
   histories, unchanged snapshots/outcomes and YOLO still on after terminal runs;
   toggle it locally in a completed session. New-task saving alone adds no folder.
8. Confirm Route questions to Telegram defaults off in Orchestrator. Enable/save,
   reopen and cancel an unsaved toggle change; details and captured settings must
   show KLM + Telegram when routed and KLM otherwise. Questions remain in KLM in
   either case. Exercise every Telegram transition above, including New code, Cancel, offline,
   unlink/disconnect confirmation cancellation and reconnect. The waiting question
   stays visible in KLM throughout. Use only the provided dummy identities.
9. Check narrow/mobile widths and keyboard navigation. Confirm normal engine chats
   still work and task/Telegram/timezone actions issue no mutation, launch or external
   Telegram requests. Graph catalog reads and existing engine inventory/Settings
   reads remain expected. Reload clears all prototype changes.

## Remaining decisions

Plan section 8 still governs manual webhook input/ingress, whole-task and queued
cancellation, exact finish/settlement, Task grants and graph mutations, workspace/
retry behavior, concurrency-mode transitions/resource limits, initial timezone and
DST/change/catch-up rules, partial-branch wait timing, Telegram storage/protocol/
identity replacement, and archive/deleted-task navigation. Local fixture shapes
and controls do not settle these contracts or establish backend support.
