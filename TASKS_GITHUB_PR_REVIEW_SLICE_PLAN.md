# Tasks vertical slice: GitHub PR-opened reviews

> Superseded implementation boundary (2026-10-05): the user clarified that Task
> webhooks must be provider-independent. GitHub PR review is task/graph configuration,
> not engine webhook logic. See [Generic Task webhooks](TASKS_WEBHOOK_HANDOFF.md).
> The provider binding, PR-specific admission and hard-coded review policy below are
> historical design mistakes, not current requirements.

Date: 2026-10-05. Status: focused source implementation added; human runtime and
deployment acceptance pending. See [setup/manual handoff](TASKS_GITHUB_PR_REVIEW_HANDOFF.md).
Derived from [TASKS_IMPLEMENTATION_PLAN.md](TASKS_IMPLEMENTATION_PLAN.md), including
its newer approved decisions and Slice A implementation notes. That broader plan
remains the source for Tasks contracts; this document selects one end-to-end delivery.

## 1. The deliverable

**Register a repository webhook on each of five GitHub repositories. When a new PR
is opened, GitHub calls a public HTTPS URL, KLM starts the repository's review Task,
and a new normal session contains the assessed review report.**

```text
GitHub pull_request / opened
  -> public HTTPS, webhook-only route
  -> engine signature/filter/dedup + durable admission
  -> new normal main session / orchestrator
  -> saved-authorized review graph
  -> review report -> orchestrator assessment -> explicit Task result
```

This combines only the necessary parts of broader slices B and E. Telegram and
scheduled execution are not prerequisites. Build the complete path for one real
repository first, then configure the same path for the other four.

### Scope recommendations requiring confirmation

- Use **one project-scoped review Task and one webhook binding per repository**:
  five repository hooks, five bindings, five Tasks, normally five existing KLM
  projects. Use one public hostname with different binding paths. Repository names,
  project mapping and the actual review graph(s) still need to be supplied.
- Review `pull_request.opened`, including draft and fork PRs targeting the configured
  base repository. Ignore later pushes (`synchronize`), reopening, review requests,
  ready-for-review transitions and merges. Draft inclusion follows "every new PR";
  confirm if the desired policy is instead to wait until ready for review.
- Deliver the report **in KLM first**. Automatically posting a GitHub comment/review,
  approval, requested-changes verdict or check status is a separate product decision;
  it is not assumed by "run a review against that PR".
- Configure these five Tasks with default **FIFO**, Telegram routing off and only
  the selected review graph authorized. Keep review work observational: inspect
  changes and report findings; do not repair, merge, push or run PR-provided scripts.
  YOLO is a permission mode, not a read-only security sandbox.

**Deferred from this slice:** Grafana, schedules/catch-up, Telegram pairing/delivery,
generic webhook providers/filter builders, automatic GitHub hook provisioning,
GitHub App/OAuth onboarding, auto-fixes and automatic re-review on every push.
Existing saved settings and approved Tasks behavior are preserved, not redefined.

## 2. Starting point: reuse what exists

Inspected current working-tree source, not just the original prototype description:

| Already present | Work needed for this slice |
| --- | --- |
| `engine/tasks.go`: durable Task definitions, revision checks, save authorization identity, enabled state, tombstoned deletion | GitHub binding/configuration, durable receipts/runs and actual consumption of the captured Task grant |
| `engine/api.go`: Task list/create/edit/enable/delete routes | Ingestion route, run admission/read/cancel and binding setup operations |
| `TaskEditorDialog.tsx`, `TaskDetailsDialog.tsx`, `TasksPage.tsx`, `types.ts`: engine-backed configuration and Details/Runs presentation | GitHub setup fields, real run projections/session navigation; Run now currently has no execution callback |
| `store.go`, `transactions.go`, `stream_journal.go`, `updates.go`; Tasks schema migration already exists | Additive validated run/receipt/binding persistence, replay/recovery and live invalidation |
| Normal sessions, graph execution/notifications/assessment, KLM question cards, YOLO adapters | Task-run lifecycle, scoped graph authority, explicit YOLO propagation to graph nodes and completion |
| `engine_time.go` and frontend engine-time formatting | Reuse existing date presentation; no scheduler or timezone feature work |

Configuration alone is not working automation. The first useful delivery must
include the real webhook, real normal session, actual review and durable result.

## 3. GitHub form: exact setup

Checked the latest screenshot in `C:\Users\Samuel\Pictures\Screenshots`:
`Captura de tela 2026-10-05 004246.png`. It shows GitHub's **Add webhook** form
with form-encoded content and **Just the push event** selected. Change those values:

| GitHub field | Value for this slice |
| --- | --- |
| Payload URL | Copy the Task binding's public URL, e.g. `https://hooks.example.com/hooks/github/<binding-id>`; illustrative route, not a deployed endpoint |
| Content type | **application/json** |
| Secret | The high-entropy secret configured for this binding in KLM; use a different secret per repository |
| SSL verification | **Enable SSL verification**; valid public certificate |
| Which events | **Let me select individual events** -> **Pull requests** only |
| Active | Checked |

Repeat in **Repository -> Settings -> Webhooks -> Add webhook** for all five repos;
GitHub requires repository ownership/admin access. Selecting Pull requests subscribes
to the event family. **KLM filters `action == "opened"`**; the form does not establish
that action filter. `pull_request_review` is the wrong event for PR creation.

GitHub sends a `ping` when the hook is created. Verify its signature and acknowledge
it without creating a Task Run. Record the successful ping in binding status so the
user can distinguish connectivity from a completed review.

The webhook secret verifies incoming deliveries. It does **not** grant access to
private repository content. Proposed smallest setup: reuse authenticated `gh` on
the engine computer, under the engine's user, and verify read access to all five
repositories. Keep GitHub credentials out of payload URLs, task text and logs.
No KLM GitHub-login flow or write-capable bot setup is required for KLM-only reports.

## 4. Minimal runtime contract

### A. Binding and public ingress

Proposed `GitHubWebhookBinding`: binding ID, task/project IDs, expected numeric
repository ID, display `owner/repo`, protected secret reference and last receipt/ping
status. One binding selects exactly one task; payload fields cannot choose another
project, graph, instructions, working directory or secret. Resolve repository ID
from trusted setup, not from an arbitrary first delivery. Match the PR's **base**
repository, so fork PRs are not mistaken for unrelated repository events.

Expose only `POST /hooks/github/{bindingId}` through a narrowly configured HTTPS
reverse proxy/tunnel to the engine. Deny every other public route/method; specifically
do not proxy `/api/*`, session tools or the private bridge. The existing API on 7331
has no authentication, so publishing its whole origin is not an acceptable shortcut.
Reuse its listener behind the route restriction; no new listener port is required
by this proposal. The actual hostname/proxy provider is a deployment choice.

Receiver sequence:

1. Locate the configured binding; require JSON and apply a bounded raw-body limit
   before allocation/decoding (proposed initial limit: 1 MiB; explicit 413 on excess).
2. Verify `X-Hub-Signature-256` using HMAC-SHA256 over the **unaltered raw body** and
   constant-time comparison (`hmac.Equal`). Reject missing/invalid signatures.
3. Validate repository identity, event header, required payload fields and action.
   `ping` succeeds without admission; other valid but irrelevant actions are recorded
   as ignored. Only `pull_request` + `opened` enters automatic admission.
4. In one durable transaction, deduplicate and record the receipt, capture Task
   revision/grant and PR identity, create the Task Run and its normal session, resolve
   the lazy folder, and append dispatch intent. Only then return 202. A duplicate
   returns a successful acknowledgment linked to the existing receipt/run.
5. Dispatch asynchronously. No GitHub fetch, model/catalog discovery or graph run
   in the delivery request. GitHub expects a 2xx response **within 10 seconds**.
   Storage/admission failure must not be acknowledged as accepted work.

Proposed responses: 200 ping/duplicate, 202 accepted, 204 ignored/disabled, 400 malformed
payload, 401 invalid signature, 404 unknown binding, 413 oversized, 415 wrong content
type, 503 storage/capacity failure. Keep public responses minimal; show diagnostics
in KLM. Persist authenticated ignored/disabled dispositions so redelivery does not
unexpectedly launch previously ignored work after enabling.

Use `(binding ID, X-GitHub-Delivery)` as the delivery key and retain a body digest to
reject reuse with conflicting content. Add `(binding ID, repository ID, PR number,
opened)` as the logical automatic-admission key: one review of the opening event,
even if duplicate hook configuration or a changed delivery header produces another
receipt. Headers are not covered by the body HMAC; do not treat a fresh delivery ID
as fresh authority. Retain compact dedup records for the binding's lifetime in this
first slice. Explicit manual reruns are distinct operations, not webhook redeliveries.

### B. Task Run, authority and normal session

Reuse the broader plan's approved contracts:

- One new **normal top-level session per accepted run**, including queued runs;
  sidebar and history use the same session ID. Default `Tasks` folder is created
  lazily with session creation; saving a task/binding creates no folder.
- Capture instructions, model/effort, graph IDs, grant and other Task settings at
  acceptance. Saved authorization permits selected graphs for those instructions;
  webhook content is untrusted input, never a forged human authorization event.
- Wire a distinguishable Task-grant reference into graph invocation/start validation,
  bound to the active run/session. Keep normal chat's user-event path unchanged.
  The main orchestrator invokes the review graph; the receiver never does.
- Persist a TaskRun execution policy and propagate default YOLO to main and graph
  sessions, including continuation. KLM questions remain interactive, with no
  automatic answer/timeout. YOLO stays enabled after the run but Task authority ends.
- Use existing per-task FIFO/default-off parallel policy. Waiting questions and
  Needs attention retain the serial slot; failure releases the next independent run
  after settlement. Configure this rollout with parallelism off; do not silently
  serialize a saved parallel configuration as though its policy were implemented.
- Disable blocks new automatic admissions but leaves accepted work intact. Delete
  refuses queued/active/waiting work; allowed deletion preserves runs and sessions.
- Require explicit assessed finish. Missing finish with no graph/question becomes
  Needs attention. **Cancel run** handles queued work or the orchestrator plus owned
  graphs; uncertain finality is Interrupted, not asserted cancellation. Ordinary
  chat Stop keeps its current meaning. Finished outcomes cannot be rewritten by
  later conversation or late notifications.
- Restart interrupts started/uncertain runs without replay. Known-not-started queue
  entries can continue with captured settings. Persist the dispatch boundary so a
  crash after native launch but before acknowledgment cannot cause blind relaunch.

Minimum new durable data: binding; authenticated delivery disposition/digest;
TaskRun with task snapshot/grant/session, origin and PR input; dispatch state; graph
activity associations; outcome/report; timestamps and user-wait intervals. Use the
existing store/journal/migration pattern, preserve schema-1 definitions, and validate
cross-record references. No separate task chat database or external queue service.

### C. Review the correct PR revision

Capture repository ID/name, PR number/URL, base and head repository IDs, **base/head
commit SHAs**, draft flag and delivery ID. Treat titles, bodies, diffs and repository
content as data; none may expand saved instructions or request credential disclosure.

Proposed first review graph: one reviewer agent returning a terminal Choice with
a report payload. The orchestrator passes a self-contained request appropriate
to its entry role, not an instruction
to implement a fix. Use a configured graph per project, reusing an existing suitable
graph where available. No new graph editor or automatic cross-project copying.

The reviewer uses `gh`/GitHub APIs with repository arguments from the verified
binding and reads diff/file context for the **captured SHAs**, including merge-base
semantics for the PR diff. Do not blindly call a current-PR diff endpoint after
queue delay and label it as the opening revision. Do not switch/reset the user's
working checkout to review a branch. Prefer API/object reads; unavailable commits,
truncated patches, pagination limits, large/binary files or missing credentials must
produce explicit coverage gaps or Needs attention, not a claim of a complete review.
Build URLs from the configured GitHub host/repository; do not follow arbitrary
payload URLs or interpolate PR text into commands.

Report: repository/PR, reviewed head/base SHAs, summary, actionable findings with
severity and file/line references, and coverage limitations. Finding bugs is still
a **successful review execution**; an API/model failure is not "no issues found".
The orchestrator assesses completeness, retains the report in the session and
records the final result. Requests for clarification stay in KLM. Corrective graph
retry retains the existing explicit fresh/reuse decision; no automatic retry loop
or repair graph is needed for this demand.

### D. Focused UI wiring

- Extend the existing **Webhook** trigger section with a GitHub repository binding
  (verified `owner/repo`), fixed **Pull request opened** event and secret setup.
  Offer Copy URL and a one-time setup secret; never include secrets in ordinary
  Task/list/run projections. Retain only protected engine-side secret storage.
- Details shows repository, endpoint, last ping/delivery disposition and any setup
  failure. Reuse existing Details/Runs, overflow, Edit and delete confirmation flows.
- Runs shows the FIFO queue, latest five non-queued entries and all older waiting
  runs without duplication. Include PR link/number, actual origin, short report,
  captured settings and execution/user-wait timing. Open session and the chat header
  backlink use the normal session route. Publish changes through existing updates.
- Proposed **Run now** input for these webhook Tasks: a PR URL/number within the
  bound repository. Resolve and capture its current SHAs before durable acceptance;
  use an operation ID and `origin=manual`, even when the Task is disabled. This is
  the recovery/rerun path, not fabricated webhook input. Generic inputs for other
  webhook providers and unrelated trigger execution remain deferred.

## 5. Implementation sequence: one vertical path, then rollout

| Step | Focused change and integration points | Demonstrable exit |
| --- | --- | --- |
| 1. One-repo walking slice | Add binding UI/config and signed receiver; TaskRun admission/session creation in engine state; Task-grant graph invocation, YOLO propagation and explicit finish through existing graph/bridge paths; wire report and session navigation. Establish the restricted HTTPS route and engine-user GitHub read access for the pilot | Open a real PR in one selected repo -> 202 delivery -> normal session -> real review graph -> assessed report visible in Runs/chat. Ping and duplicate delivery do not start reviews. |
| 2. Make that path operable | Complete FIFO, KLM waits, cancel/settlement, disabled/manual behavior, delete refusal, restart reconciliation and concise delivery diagnostics. Add PR-input Run now through the same admission service | A second PR queues, a waiting run is reachable, cancel/failure allows the next run, and restart never repeats started work. Manual rerun uses a fresh run/session. |
| 3. Five-repo rollout | Retain and verify the exact-path HTTPS deployment; configure the other four trusted repo-to-project/task mappings and GitHub hooks; verify engine-user read access and model availability for all five | One newly opened PR in each repo yields the matching independent review/session; public requests cannot reach general engine APIs. |

Relevant code areas: `engine/tasks.go` plus focused webhook/run modules;
`api.go`, `store.go`, `transactions.go`, `stream_journal.go`, `main.go`, `updates.go`;
graph invocation contract/tools/runtime/scheduler and linked bridge/prompts;
frontend `engine.ts`, Tasks components, `App.tsx` and existing chat/sidebar navigation.
Implement necessary lifecycle behavior alongside the first runner, not a second
temporary execution path. The table is delivery order, not permission to deploy
unsettled cancellation/recovery or bypass existing graph lifecycle checks.

## 6. Manual acceptance and availability

1. Add each hook using the form above; signed ping succeeds with **zero runs**.
2. Open a PR in each repository. Verify routing, captured SHAs, new normal sessions,
   lazy folder creation, real reports, graph assessment and frozen outcomes.
3. Redeliver an accepted event in GitHub: no second run/session. Send an unrelated
   PR action, wrong-repo payload, invalid signature or oversized request: no review.
4. Open two PRs quickly in one repo; confirm FIFO. Push commits while one waits;
   confirm its report still identifies the captured revision, not a newer diff.
5. Disable during a run: queued work continues, new automatic input is ignored;
   manual Run now still works. Delete is refused while work is pending.
6. Ask a KLM question, answer/cancel it, and exercise Cancel run and restart. Waiting
   rows remain reachable; no false success or duplicate execution after recovery.
7. Continue chatting after completion: original review outcome is unchanged and
   the Task grant cannot be reused. Use the task backlink to return to history.
8. From the public hostname, `/api/state`, session endpoints and non-webhook paths
   must be inaccessible. Verify this before registering all five production hooks.

**Availability matters:** the engine and HTTPS forwarding process must be running;
the desktop does not need to be open. GitHub **does not automatically retry failed
webhook deliveries**. For the smallest release, use GitHub's Recent Deliveries and
manual Redeliver after an outage. A previously accepted delivery still deduplicates.
If "every PR" must include unattended outage recovery, add that explicitly to this
slice: a GitHub delivery-reconciliation job or durable external ingress is needed;
it is not solved by KLM's local queue. Do not promise lossless delivery while offline.

Implementation checks follow `AGENTS.md`: one relevant lightweight build/type check
for changed layers, then these human acceptance paths. No full suite or review-agent
work is implied. This planning turn runs no application or runtime tests.

## 7. Choices to confirm before implementation

1. **Result destination:** accept KLM-only reports initially, or must the review also
   be posted to GitHub? If posting is required, include a narrowly defined comment
   or review operation, write credential scope and duplicate/uncertain-post handling.
2. **Five mappings:** repository names, KLM projects, selected review graph(s), and
   engine-user GitHub read access. The five-task arrangement is a recommendation.
3. **Public endpoint:** hostname and the available HTTPS proxy/tunnel. Confirm the
   route-only boundary and whether it must persist across engine-computer restarts.
4. **Delivery expectation:** online engine plus manual redelivery after outages,
   or unattended recovery as part of acceptance? Include drafts at creation as
   proposed, or trigger only when ready (which changes the requested event policy)?

Narrow runtime design choices can be settled with implementation: finish tool wire
schema, protected secret mechanism, explicit queue/storage bounds, graph catalog
changes after capture, and archived/missing resource errors. Recommend failing or
requesting attention rather than silently substituting a graph/folder/model. These
edges must not reopen the broader plan's already approved core Tasks behavior.

## Sources checked

- [Creating webhooks](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks)
- [Events and payloads: pull_request](https://docs.github.com/en/webhooks/webhook-events-and-payloads#pull_request)
- [Validating webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)
- [Webhook best practices](https://docs.github.com/en/webhooks/using-webhooks/best-practices-for-using-webhooks)
- [Handling failed deliveries](https://docs.github.com/en/webhooks/using-webhooks/handling-failed-webhook-deliveries)
- User's latest GitHub form screenshot and the current local Tasks implementation.
