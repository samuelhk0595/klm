# Project automations implementation

## Delivered surfaces

- Project settings → Automations: create/edit, pause/reactivate, delete, schedule-only
  Run now, relative next occurrence, pending/running counts and session-link history.
- Saved task/harness/model/effort/YOLO, graph allowlist, Library positive selections,
  existing topic or a new name/color created atomically with the automation.
- Three session modes: new per run (legacy default), create once and reuse, or an
  existing normal same-project session. Initial input is creation-only; subsequent
  iterations and existing-session runs use a separate follow-up prompt. A linked
  reusable destination stays fixed. The shared session waits for existing work;
  multiple automations targeting it serialize without erasing history or preferences.
- One-time schedules use an explicit date/time in the configured engine zone.
  A durable consumed marker prevents reactivation/time-zone edits from repeating
  an accepted one-time occurrence. Missed one-time execution still requires the
  normal per-occurrence human recovery decision.
- Project-local released GitHub PR-opened triggers with only the three documented
  reference values, durable deduplication and FIFO independent of the message queue.
- Alarm recurrence and preview share one calendar calculation. Global IANA time zone
  starts as America/Sao_Paulo and is also the engine's Analytics base.
- Normal sessions and execution pipeline, captured capabilities across continuations,
  outcome tools bound to the root turn/run, one classification reminder, compact
  HTTP/SSE status, steering-only composers and confirmed Stop on shared chat surfaces.
- Startup/shutdown interruption records and missed schedule intervals. The general
  singleton's narrow recovery tools read the affected entries and apply a real
  later human recover/discard decision atomically with idempotent trigger admission.
  An unconfigured general harness is not silently locked; its notice remains pending.
- Library deletion previews include future automation references and captured runs;
  deleting an item removes future selection references and preserves captured runs.

## Explicit implementation assumptions

- Weekly recurrence means the next selected weekday (e.g. next Tuesday), including
  today if its time is still future. No start-date field is requested.
- Every N days uses an internal local-date anchor selected on save. Monthly 31 skips
  months without a 31st. Invalid DST wall times skip; a repeated wall time occurs once.
- Pausing blocks new automatic triggers, not already accepted FIFO entries. Resuming
  does not recover the paused period. FIFO entries use current configuration when
  admitted; active runs keep their snapshot. Recovery entries keep saved configuration.
- Missed downtime is stored as an interval with the saved schedule and IANA zone;
  stable individual occurrences are enumerated lazily in pages of at most 50.
  Recovery/discard decisions select explicit occurrence IDs (up to 50 per call),
  with durable exact-set idempotency receipts. Recovering three selected missed
  times queues three runs, serialized by the automation and shared destination. No sessions
  are created merely by detecting or reading downtime. Interrupted runs retain the
  original session link; recovery follows the saved session mode, reusing its
  linked destination with follow-up input or creating a new per-run session.
- Selecting the harness default resolves the known model/effort at save so the new
  session is born with concrete configuration. A missing default requires choosing
  an explicit model. Version, projection and Library gates still apply at execution.
- The public personal-use API and existing private loopback authentication boundaries
  are unchanged. Automatic prompt events use `agent_prompt`, never fabricated humans.

## Lightweight checks

Completed on 2026-10-23 after implementation:

- `gofmt -w` on the changed Go implementation files.
- From `engine`: `go build -tags dev -o "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-automations-dev.exe" .` — successful, no compiler errors.
- From `clients/desktop`: `npx tsc --noEmit` — successful, no type errors.

An intermediate Go build and TypeScript check also succeeded; the final pair above
includes the subsequent UI, lifecycle and recovery changes. No full suite,
integration tests, installed harness runtime or human acceptance is claimed here.

## Originating review corrections

- R1: The initial harness lock ignores attributed internal recovery notices,
  including earlier unmarked notices referenced by recovery records. It retains
  lock evidence from real legacy turns, accepted-message receipts and queues,
  including removed queued inputs. Loading repairs the earlier notice-only lock
  migration while preserving opaque legacy locks without history.
- R2: Removed the single catch-up-per-interval assumption. Recovery now reads actual
  missed scheduled occurrences and resolves chosen IDs individually or in selected
  batches. Saved time zone, pending boundary, decisions and receipts survive restart.
  A legacy pending interval without a saved zone asks for its original IANA zone;
  it does not silently adopt the current engine setting. The explicit resolution
  saves that zone with its receipt. Already-resolved legacy intervals are preserved
  without admitting additional work.
- R3: History renders only current session-title links. Refresh reloads the loaded
  page window instead of collapsing to the first page. Keyed state and aborted
  requests keep another automation's links from appearing during navigation.
- R4: Backend Save validation rejects webhook references in a scheduled prompt.
  Switching modes retains the user's prompt and shows an English inline error;
  the user must remove the references before saving. Webhook substitution remains
  literal and single-pass.

Correction checks completed on 2026-10-23:

- `gofmt -w` on the eight Go files changed in this correction pass.
- From `engine`: `go test -tags dev -run '^TestAutomationReview' -count=1 .` —
  successful (`ok klm/engine`, final run 1.367s). Only the three review-regression tests ran.
  They cover notice-only versus accepted-input harness lock evidence, in-memory
  restart serialization, paginated occurrence identity/selection and exact-set
  receipts, preserving a saved zone, recovering all three selected missed times,
  rejecting unknown legacy time zones until an explicit original zone is supplied,
  and rejecting scheduled webhook references while keeping one-pass substitution.
- From `clients/desktop`: `npx tsc --noEmit` — successful, no type errors.

The pure helper regressions in `engine/automation_review_test.go` do not execute
the engine, HTTP endpoints, an automation or a native harness, or use actual user
data. An earlier targeted run also passed; the final rerun includes the legacy-zone
guard. No full suite or additional runtime/build workflow was run in this pass.

## Manual acceptance after originating review

Create an automation with a new topic and verify it is reused. Exercise settings and
Library on Pi/OpenCode/Codex; use a real project webhook and a different source
repository. Confirm FIFO versus webhook parallelism, paused Run now and unchanged
recurrence, daily/every-two-days/weekly/monthly calculations and engine time-zone
changes. Verify steering, native requests, cancellation versus confirmation of Stop,
the two explicit objective outcomes and the one-reminder failure path. Restart during
a run and after a missed occurrence, then recover/discard through KLM Agent and confirm
new sessions are only created after the human decision. Confirm completed automation
chats work as ordinary sessions, graph authorization remains mandatory and outcome
tools never appear in children or other conversations.

After review corrections, manually cover first-ever KLM Agent use with a recovery
notice, changing its harness before input and restarting before input, then verify
the lock after a real accepted message. Recover/discard separate missed times and
confirm normal serial execution of all chosen occurrences. Load more than 50
session links, leave the history visible through refresh, and switch automations
during a slow request. Switch a webhook template to Schedule and remove references
to clear the validation error. These interactive paths still require human acceptance.

## Reusable sessions and Once extension

A user-requested native planning subagent returned a read-only plan before this
extension was implemented. No additional session or graph was created for planning.
The implementation adds optional mode/session/follow-up fields with legacy defaults,
multi-run session associations, atomic first-session linking and durable ownership
checks. Effective run settings are projected during execution/policy checks; normal
session preferences remain intact, including when a saved recovery snapshot is used.
Runtime catalog keys include run identity so outcome tools cannot persist into a
later run or ordinary chat. Stale Stop confirmations reject an ended/different run.

Human acceptance remains required for all three modes, multiple follow-ups in one
native conversation, changes to session options between runs, busy/archived targets,
two automations sharing a target, recovery of several occurrences in that target,
and Stop followed by ordinary chat. Validate Once before/after its selected local
date/time, after pause/reactivate and time-zone changes, and after restart/downtime.
Extension checks completed on 2026-10-23: `gofmt -w` on the changed Go files;
`go build -tags dev -o "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-automations-dev.exe" .`
from `engine`, and `npx tsc --noEmit` from `clients/desktop`, both successful.
An intermediate pair also compiled; the final pair includes the one-time consumed
marker and the recovery/ownership changes. No full suite, native harness execution
or manual UI/runtime acceptance was performed for this extension.
