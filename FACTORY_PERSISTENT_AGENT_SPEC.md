# Persistent Factory Agent — Slice Specification

Status: **Draft for user review; implementation is not authorized by this document.**
Date: 2026-11-13.

## 1. Purpose and approval boundary

Introduce the persistent **Foreman** in Werbert, a separate Factory application
using KLM as its runner, with the public DeepSeek Superagent mechanism as the
behavioral baseline. This supersedes the earlier in-KLM Factory proposal.
The user selected that baseline
after reporting that Werbert left a necessary human permission request visible
only in Tickets while its coordinating Chat stayed silent.

The desired behavior is quiet ongoing work with prompt communication when a
person must act or can use a result. Frequent progress narration is not desired.

Each vertical slice requires a clear specification, human review and approval
before implementation. Approved scope must be followed exactly. Changes return
to review. This draft records the selected baseline and proposed design choices;
it is not permission to implement an entire Factory.

Werbert is implementation inspiration only. Its code, specifications and recorded
test/deployment claims do not establish that its behavior meets this requirement.

### 1.1 Current first-version direction

The user narrowed the initial deployment to Werbert and exactly one KLM engine
installed on the same computer. Remote engines, multiple connected engines and
distributed scheduling belong to later work. Ordinary KLM usage remains available
independently of Werbert.

The initial user journey is:

1. On first opening Werbert, discover the local release engine at the standard
   API port, `7331`, and display a Connect action. Development uses `17331`;
   developer endpoint selection is not an additional production engine slot.
2. After the user clicks Connect, establish and retain the engine connection with
   minimal friction. The user wants no manually copied pairing code for this
   local flow. The proposed automatic local handshake is specified in section 6.1.
3. List projects already registered in the connected KLM. The user selects a
   project to create its Factory. Keep this project-selection page accessible
   later so additional projects on the same engine can acquire Factories.
4. Create the project's base Factory definition using the proposed file contract
   in section 6.2, preserving existing definitions and local changes.
5. Open Factory configuration. Its initial agent configuration is Foreman's
   harness, model, effort, instructions, selected skills and selected MCP tools.
   Persist this configuration in versionable project files. Proposed validation
   and activation behavior are specified in sections 6.2 and 6.5.

Werbert delegates harness/provider/dependency installation and execution-environment
readiness to KLM rather than retaining its own environment setup workflow. KLM
diagnostics still need to reach Werbert when requested execution is unavailable.

**Selected Library scope:** reuse the connected KLM engine's existing shared
Library. The first delivery lists available skills and MCP connections/tools and
lets the user select Foreman's capabilities. Persist each Factory's selection in
its versionable definition files; KLM owns installed content and credentials.
Library installation, editing and removal through Werbert are deferred to a later
slice, superseding their earlier inclusion in the initial onboarding discussion.
When introduced, operations on this shared Library also affect its availability
in KLM. No separate Factory-owned Library or automatic inheritance to workers is
introduced by this decision. The proposed first slice discovers/selects an
existing saved worker Agent; agent authoring remains in KLM for this delivery.

Connection credentials and provider/MCP secrets stay in protected local storage;
portable files contain configuration and references rather than secret values.
This is distinct from the Factory's operational history and execution records.

**Connection evidence, not an approved implementation:** KLM authenticates ordinary
loopback API clients today. Its five-minute, one-use pairing code creates a separate
client credential. The local `klm pair` command obtains that code through OS-local
control. A Werbert backend running as the engine user could automate that handshake
after Connect, avoiding manual code entry without opening unauthenticated HTTP
access. Same-machine deployment alone does not implement this connection flow.

### 1.2 Approved execution visibility and workspace scope

The user approved keeping Factory-origin execution out of the ordinary KLM user
experience. Opening the same project in KLM shows personal sessions normally;
it does not expose Foreman, Ticket workers, their activity or their pending
questions/permissions in ordinary session lists, Happening now, Focus widgets or
other personal-conversation views. The execution origin must be represented in
the engine and respected by its user-facing projections, not hidden only by a
frontend title filter. The proposed public runner contract is linked in section 6.3.

Werbert owns the user-facing monitoring, histories and intervention flow for its
executions. KLM retains their technical execution records as runner evidence;
excluding them from personal views does not discard records. Factory definition
files alone do not start work or change a personal conversation's role or prompt.

For this version, Factory work and personal KLM sessions use the project's same
existing checkout. The user explicitly accepts concurrent changes in that shared
workspace. Do not add automatic worktree creation, branch switching, cleanup,
workspace isolation or a new project-wide exclusion policy to this slice.
Worktree/isolation strategies are future work. Shared machine resources and file
changes remain real effects of running the work; visibility separation is not
process or filesystem isolation.

**Compatibility invariant explicitly confirmed by the user:** ordinary KLM UI
and execution behavior must remain the same both with and without an attached
Factory. Connecting Werbert must not switch a project into a different personal
mode, change ordinary prompts/tool catalogs or session grouping, or alter personal
model/effort defaults, permission handling, queues, steering, Stop or native
compaction behavior. The five-second shell/job policy and external-role routing
apply only to captured external executions, not to shared harness-wide defaults.
Factory state or definition files do not modify personal-session admission.
Acceptance must include ordinary KLM journeys with the Factory disconnected and
connected/working. This invariant concerns UI/behavior, not guaranteed unchanged
performance or invisible file edits in the explicitly accepted shared checkout.

## 2. Exact reference

Repository: `deepseek-ai/dsh-superagent-prototype`.
Reference revision: `8c32a9dc5d72c033816c6a8110b37f40cf94fba4`.

Use this revision rather than a moving branch. The public prototype is the
reference, not an inferred internal version or unimplemented article designs.
Its mechanism combines prompts, structured tools, durable records and delivery;
copying prompt text alone does not reproduce it.

The user explicitly requires reproduction of the internal operating mechanism,
not only similar notifications: bounded foreground commands, background jobs,
serial next-step admission, asynchronous delegation and event-driven resumption.

Authoritative reference paths at that revision:

| Reference | Relevant behavior |
| --- | --- |
| [Philosophy](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/docs/000-philosophy.md) | Minimize human attention; act within authorization; deliver useful outcomes. |
| [Execution](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/docs/003-execution.md) | Persistent execution, human handoffs, recovery and delivery status. |
| [Prompts](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/docs/004-prompts.md) | Coordinator/worker roles and input timing. |
| [Communication](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/prompts/team/communication.md) | Silence during ordinary work; concise, source-directed notifications. |
| [Replies](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/prompts/replies.md) | Purpose-bound replies, acknowledgement ownership and delivery receipts. |
| [Worker handoff](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/prompts/handoff-worker.md) | Structured human-action requests rather than questions hidden in prose. |
| [Coordinator handoff](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/prompts/handoff-orchestrator.md) | Prompt delivery of the current question, even while work continues. |
| [Executor tools](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/src/app/tools/executor.ts) | Atomic report/relay recording and explicit notification classification. |
| [Notices](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/src/app/notices.ts) | Attention wakes coordination; routine progress stays internal. |
| [Reply enforcement](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/src/core/replies.ts) | Current-request matching, source routing and stable reply identities. |
| [Runtime presets](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/cordis.patch.yml) | Router/worker foreground shell and job-wait limits of 5,000 ms, enforced by runtime configuration. |
| [Delivery adapter](https://github.com/deepseek-ai/dsh-superagent-prototype/blob/8c32a9dc5d72c033816c6a8110b37f40cf94fba4/src/app/delivery.ts) | Separate next-step/next-turn/passive paths, producer attribution, persistence and delivery deduplication. |

## 3. Selected behavioral baseline

### 3.1 Roles and availability

- The coordinator routes requests and delivers human-facing outcomes. Workers
  perform technical investigation, planning, implementation and review.
- A worker owns the concrete acknowledgement before substantial work on a human
  request; delivery goes to the originating conversation and is deduplicated.
- Quick human questions receive direct answers without a separate acknowledgement.
- Coordination turns are serial. Delegated execution is asynchronous: when only
  delegated work remains, the coordinator yields instead of sleeping or polling.
- Actionable inputs can wake coordination. Informational context alone does not
  wake an idle coordinator. Busy delivery uses the next eligible step without
  interrupting the model/tool operation already in flight.
- Persisted conversation identity and work evidence are distinct from the lifetime
  of a resident harness process. Restart does not authorize replay of work.

### 3.2 Communication

| Situation | DeepSeek baseline |
| --- | --- |
| New human work request | One brief, concrete acknowledgement before substantial work. |
| Routine progress, a delegation or an unfinished milestone | Internal evidence; no unsolicited progress message. |
| Ready result the person can review or use | Concise outcome and actual result link in the originating conversation. |
| Necessary human decision, permission or resource | Prompt, concrete request in the appropriate source conversation. |
| Recoverable problem the agent can handle within scope | Continue recovery silently. |
| Definitive inability to finish | Explain the inability and any necessary human action. |
| Direct status question | Answer from actual evidence, even while work is unfinished. |

Creating a PR alone does not mean it is ready or that the objective is complete.
Do not announce an unfinished PR as a ready result. Do not forward or compress an
entire internal report into a notification. Detailed engineering evidence remains
available through the work record.

The reference defaults to one short sentence, adding another only for a material
limitation or necessary action. Its IM transport has a hard 100-word ceiling;
GitHub communication has separate source-specific rules. Applying this exact
transport constraint to the Werbert Chat is proposed in section 6.4.

### 3.3 Human-action lifecycle

1. The worker records one concrete question in the structured `human_action`
   field. A question written only in a summary is not this mechanism.
2. The host records the report and required coordinator relay atomically.
3. An active question produces an actionable relay, including when the worker
   reports `progress` and continues other authorized work. Do not wait for the
   whole work item to become blocked.
4. The coordinator sends an `attention` reply for the exact current question.
   The host checks that it matches the recorded request and routes it where the
   human message that raised the question was written, or to the work's original
   conversation when no such message exists.
5. Repeating the same question retains its request identity. A different question
   replaces it with a new request. Omission preserves a pending question;
   explicit clearing or a ready report clears it in the reference implementation.
6. Do not resend an unchanged request. Stored or queued delivery is not confirmed
   delivery. An undeliverable reply is recorded rather than claimed as sent or
   redirected to an invented destination.

Routine progress does not authorize an attention message. A `blocked` or `failed`
label alone does not establish that a person must act.

### 3.4 Internal execution mechanism

These are selected mechanism requirements, not merely prompt recommendations:

| Mechanism | Reference contract to reproduce |
| --- | --- |
| Foreground shell | Router and direct executor presets use both default and maximum foreground waits of 5,000 ms. A command still running becomes a tracked background job; it is not killed and relaunched or reported as failed just because five seconds elapsed. |
| Job waits | Default and maximum waits are also 5,000 ms. A model cannot regain a long blocking wait by choosing a larger timeout. |
| Background completion | Job completion returns an attributed result and can wake the owning agent. Model turns need not remain blocked awaiting completion. Job/process ownership, output and cancellation remain the runner's responsibility. |
| Delegation | Admission returns a work reference while execution continues separately. A later result/request wakes coordination; the coordinator does not hold a tool call for the delegated work's whole lifetime. |
| Serial coordination | The Foreman has one model/tool step in flight. Receiving another event does not start a second concurrent Foreman turn. |
| Next-step steering | An actionable arrival joins the next eligible model step, possibly within the same turn. The in-flight model response or tool operation is preserved. An input arriving after that step claimed its batch waits for another boundary. |
| Next-turn queue | Explicitly queued input waits for the ongoing turn to finish. This is distinct from next-step steering. |
| Passive context | Informational notices are retained without independently waking an idle Foreman. |
| Relevant event delivery | New work candidates, necessary human actions, completed jobs and relevant executor results use actionable delivery. Routine progress without a human-action request or answer remains internal. |
| Durable attribution | Producer/request/work identities survive delivery retries. Automated events remain automated input even when a harness represents them with a user-role message; they never become human authorization. |

Here, a **step** means one model request and its resulting tool execution; a
**turn** may contain several steps. The five-second shell handoff is not a promise
that every model request, remote MCP call or end-to-end reply finishes within five
seconds. The reference's shell policy must not be generalized into killing every
tool after five seconds. Long non-shell tools need an explicit mapping before the
slice claims the same availability.

The reference applies the five-second policy to its router and direct workers.
Its spawned subagents use `standard`, whose shell handoff is 60 seconds; do not
mistakenly claim that every nested agent has the router's five-second policy.

### 3.5 Factory/runner ownership and inspected gaps

Werbert owns durable work intake, Tickets, coordinator routing and notification
intent. KLM owns native execution, shell processes/background jobs, supported
model-step delivery boundaries and actual native request settlement. The public
runner contract must connect those responsibilities; adding instructions to
Foreman's prompt or restoring Werbert's private harness adapters is insufficient.

Static comparison of the inspected code, not runtime acceptance:

| Area | Evidence and consequence |
| --- | --- |
| Werbert input timing | `factory_inbox.go` distinguishes `steering`, `next_turn` and `passive`, with claimed input batches and step identities. `factory_dispatcher_boundary.go` admits pending input before an owned tool reply releases the model. These are useful reference mechanisms, not proof of the complete user journey. |
| Werbert native steering | `factory_native_protocol.go` has Pi `steer` and Codex `turn/steer`, including acknowledgement checks. In the new architecture the corresponding native adapter work belongs in KLM. |
| Werbert report classification | `ticket_assignment_store.go` coalesces undelivered routine progress but still enqueues it as a steering notice, whereas DeepSeek's `reportRelay` keeps ordinary progress internal. This is an observed mechanism difference to correct, not a diagnosed cause of the user's missing-permission incident. |
| KLM existing steering | `engine/codex_interactive.go` already sends `turn/steer`; `engine/opencode_interactive.go` uses a plugin-managed native model boundary. Existing chat steering is a building block, not a complete externally attributed Foreman inbox. |
| Five-second job handoff | The inspected Werbert/KLM paths do not establish a matching managed shell/job contract. A five-second shutdown deadline or version-probe timeout is not evidence of foreground-to-background conversion. Determine the actual implementation for each harness selected for the slice. |

Before implementation approval, document which existing mechanisms are reused,
which need a KLM runner extension and any unavoidable difference from the pinned
DeepSeek reference. Validation must observe actual delivery and process behavior,
not only matching setting names or prompt text.

## 4. Required KLM mapping: native questions and permissions

This is a platform adaptation to specify explicitly, not an existing DeepSeek
native-permission feature. DeepSeek's prototype runs unattended executors and
excludes interactive question tools from their tool set; its human-action route
is the structured report/attention mechanism above.

KLM already owns native questions and permission requests. To reproduce the
human-facing behavior for the user's reported failure, their exact pending
requests must be surfaced in the coordinating Chat. The mapping must preserve:

- The real execution, request identity, choices and permission/grant scope.
- A distinction between asking for task information and approving an operation.
- An explicit human response through the owning runtime's supported reply path.
- The pending request when a notification, model turn or transport fails.
- Work continuation only as permitted by the actual answer and runtime state.

Prose, a coordinator summary or a worker report cannot approve a native action.
The proposed Chat presentation and failure behavior are specified in section 6.4.
A Ticket-only attention badge does not satisfy the user's requirement.

## 5. Review scenarios

These describe observable behavior to validate after the scope and platform
mapping are approved. They are not claims of implementation or test results.

1. **Quiet work:** a delegated worker produces routine progress. The coordinator
   remains available for conversation without unsolicited progress narration.
2. **Necessary decision:** a worker records a concrete human-action request while
   other work continues. The question reaches the originating Chat without the
   user opening Tickets or requesting status.
3. **Native permission:** a worker requires permission. The coordinating Chat
   presents the real pending request and supported choices. Waiting alone grants
   nothing; the exact human reply reaches its owning execution.
4. **Deduplication:** the same pending question is reported/delivered again. It
   remains one logical request rather than repeated unsolicited questions.
5. **Resolution:** answering or clearing a request updates its pending state;
   unrelated progress is not mistaken for an answer or approval.
6. **Delivery failure:** a failed notification is not reported as delivered and
   the underlying request remains recoverable. The approved Chat failure display
   must make the situation visible without a Ticket-only discovery path.
7. **Ready outcome:** the coordinator delivers the usable result in the source
   Chat. An unfinished PR or successful process exit is not sufficient evidence.
8. **Recovery:** a restart preserves conversation and pending-request evidence,
   reconciles runtime ownership and does not repeat uncertain external actions.
9. **Long command:** start a command that runs beyond five seconds. It becomes
   one owned background job, with no duplicate execution, and releases its
   foreground tool wait. The Foreman can process newly admitted work before that
   command completes; its eventual result remains attributable and inspectable.
10. **Arrival during a turn:** while the Foreman has an in-flight step, deliver
    new Ticket work and another executor's relevant result. They enter the next
    eligible step without cancelling that step's predecessor or launching a
    concurrent Foreman turn. This scenario checks event timing, not the addition
    of a multi-worker scheduling product to the initial slice.
11. **Queue versus passive:** next-turn work waits for the turn boundary; passive
    context alone leaves an idle Foreman idle; an actionable result wakes it.
12. **Wait bound:** an attempted longer job wait cannot defeat the five-second
    foreground policy. Background results return by event rather than a Foreman
    sleep/status-poll loop.
13. **Ordinary KLM usage:** while Factory work is active or awaiting a human,
    opening its project in KLM retains the personal-session experience and does
    not list or surface Factory execution activity/requests. Those requests and
    results remain accessible through Werbert.
14. **Shared checkout:** both applications use the existing project checkout;
    Factory changes are visible there without a new worktree or automatic
    reset/branch change. This is the explicitly accepted first-version behavior.
15. **Stop Foreman:** stop a coordination turn while a Ticket worker is active.
    Only Foreman's owned turn is stopped; the worker continues. The conversation
    and Ticket evidence remain. A later actionable arrival can start a new serial
    turn after the stopped turn's cessation is confirmed; it does not replay the
    stopped input.
16. **Cancel Ticket:** request cancellation and observe a pending cancellation
    before confirmed worker cessation. Uncertain cessation must not be displayed
    as Cancelled or Completed. Other Tickets and personal KLM sessions continue.
17. **Configuration boundary:** save or externally edit the Foreman definition
    during a turn. That turn and all admitted workers retain their captures; a
    later drained turn uses the new validated configuration. Missing referenced
    resources block new admission visibly rather than being silently omitted.
18. **Continuity:** native compaction keeps the full user-visible history and
    authoritative work references. A harness switch preserves the Werbert Chat
    and creates a new native context, without claiming cross-harness native resume.
19. **Definition safety:** initial project selection creates no duplicate Factory
    on repetition, does not overwrite existing definition bytes and starts no
    model work. Stale settings saves conflict without replacing a newer file.

## 6. Proposed resolution of the remaining design points

The user approved Stop's separation from Ticket cancellation and delegated
preparation of the other design points. That authorizes this specification work,
not implementation or automatic approval of the proposals below.

### 6.1 Local connection and entry points

Run Werbert's backend locally as the KLM engine's OS user; its browser frontend
talks to that backend. The backend discovers release KLM at `127.0.0.1:7331` and
checks engine identity and runner compatibility. No LAN scan or remote-host UI is
introduced. A missing or incompatible engine is an actionable connection error,
not a request for Werbert to install harnesses or repair providers.

Connect performs the existing OS-local pairing handshake automatically, redeems
the one-use code as a native API client and protects the resulting credential in
local application storage. Codes and credentials never enter project definitions,
model prompts or logs. Reopening Werbert uses the retained engine binding; expired
or revoked credentials return to Connect rather than creating replacement work.

After connection, show KLM's registered projects. Selecting an already-connected
project opens its Factory; selecting a new project initializes its definition and
opens Settings. A permanent **Add Factory** entry returns to this project list.
Each Factory exposes **Chat**, **Tickets** and **Settings**. Chat is Foreman's
continuing conversation. Ticket details are optional inspection, not the only
place where interventions can be answered. No redesign of other Werbert modules
or new engine/environment setup wizard is included in this slice.

### 6.2 Versionable definition and prompt ownership

Use one portable definition at `<KLM project directory>/.werbert/factory.yaml`:

```yaml
schema_version: 1
name: Example project
foreman:
  harness: ""
  model: ""
  effort: ""
  instructions: ""
  skills: []
  mcp_connections: []
```

- Empty harness/model is the newly initialized, unconfigured state, not an
  instruction to silently select a provider. The user chooses from KLM catalogs.
- Effort is optional only where KLM supports the model default; KLM records the
  effective value. Instructions are the user's additional Foreman instructions.
- Skills are KLM skill IDs. Each MCP selection is `{connection_id, tools}`, with
  an explicit tool allowlist, translated to KLM's existing selection contract.
  No automatic selection of new tools is implied.
- KLM resource IDs are references to this engine, not a portable installation or
  embedded credential. Missing references produce a configuration diagnostic.

The local Werbert backend writes this file in the registered project's actual
directory. Saves validate the strict schema, use a content revision for optimistic
concurrency and atomically replace one file. Initialization creates only absent
files, preserves dirty content, and is idempotent per engine/project association.
An existing valid definition is read; an invalid one is preserved for correction.
Legacy `.werbert/factory-agent.yaml` files and operational records are not silently
converted, merged or executed by the new path. No commit is automatic.

Foreman's baseline role, event-handling and handoff instructions remain separate
application-owned Markdown under Werbert's `engine/prompts/`, including
`foreman.md`, `foreman-input.md` and `factory-worker.md`. They are implementation
assets versioned with Werbert; user instructions in YAML supplement them rather
than replacing them. KLM receives the captured assembled instructions for the
runner-owned conversation/work. KLM does not discover a Factory role merely by
opening a project containing `.werbert/` files.

Werbert's operational database owns Factory/conversation IDs, Tickets, Assignments,
work links, inbox/outbox entries and receipts. KLM owns native contexts, execution
records, captured capabilities and process evidence. None is exported as portable
Factory configuration. The first slice reads/selects existing saved Agents in KLM;
it does not duplicate agent authoring or automatically install a specialist.

### 6.3 Ticket delegation and runner boundary

Use this first complete path: **human Chat request -> Ticket -> one saved-Agent
Assignment -> KLM execution -> Foreman assessment -> Chat result**. A read-only
Code Reviewer task is the proposed acceptance workload, not a hard-coded product
specialist. The human request supplies the actual objective.

Operational conversation creates no Ticket. A substantive request creates one
Ticket bound to its retained original human input, objective and completion
criteria. Duplicate delivery of the same request resolves the same Ticket.

Foreman discovers and selects an available saved Agent. The Assignment records
the task, Agent identity/revision and its relationship to the Ticket. KLM captures
the actual harness/model/effort, instructions, selected capabilities and current
checkout before admission. Assignment admission returns promptly with durable
references. The worker's lifetime never holds Foreman's conversational lock.

Ticket views distinguish Queued, Working, Waiting for user, Completed, Failed,
Cancelling, Cancelled and Interrupted. Waiting for user means no executable work
on that Ticket can proceed without the answer; a pending request may coexist with
Working. Cancelling is intent, Cancelled requires cessation, and Interrupted
retains unknown outcomes. Foreman's visible assessment records a reason and the
actual result references rather than copying a worker's exit status.

The proposed first-slice boundary allows one Assignment per Ticket, with multiple
Tickets tracked independently. It adds no dependency graphs, automatic retry
policy, replacement-worker pipeline or Factory-specific capacity controls. A
failed/incomplete result remains honest and is reported; success still requires
a separate objective assessment. This intentionally bounded path is for review,
not an assertion that all wider Factory workflows have been implemented.

Workers use their saved Agent selections plus the application-owned worker role;
Foreman selections do not propagate. They report results and structured requests
to Foreman. Consultations concern original intent, scope or already-established
context. Foreman answers from retained evidence or asks the user; neither a
consultation nor an MCP result creates new authority.

The required generic KLM extension is detailed in
[the proposed runner contract](FACTORY_RUNNER_CONTRACT.md). Its essential missing
piece is externally owned conversation/work admission with attributed input,
durable events and application-tool request/reply handling. Ordinary
`POST /api/sessions/{id}/messages` and a blocking `subagent_run` are not substitutes.

Factory tools execute in Werbert; native harness adapters remain in KLM. Tool
manifests are captured at native-context creation, and KLM forwards operational
tool requests to Werbert through its owned runner channel. Werbert validates the
actual input/Ticket scope and persists each effect/receipt before replying. KLM's
private linked/general-agent tools are not exported as Factory tools.

### 6.4 Chat, human requests and controls

**New explicit layout requirement:** Foreman's message timeline and composer must
match the current KLM Chat exactly in appearance and interaction structure. Use
the actual KLM frontend source as the reference, not the older Werbert Chat or an
approximate redesign. Adapt/copy the required frontend structure into Werbert;
this does not authorize changes to KLM's interface.

The composer shows the effective model, effort and YOLO state as read-only
information. It offers no model/effort/harness or execution-mode editing there;
Foreman configuration belongs in Factory Settings. Foreman always runs in YOLO
mode; this is a fixed role requirement, not a composer toggle. Preserve the exact
pending human-question and elicitation/permission contracts that still apply to
delegated workers or external tools; Foreman's YOLO does not answer them.

**Approved:** Stop Foreman requests interruption of its current turn only. Existing
Ticket workers continue. Cancel Ticket is the operation that stops that Ticket's
delegated execution. Neither operation stops a personal KLM session.

Proposed precise control semantics:

- Foreman Stop targets the observed conversation/turn and its own runner-owned
  foreground/background activity, never independently admitted Ticket workers.
  A committed delegation remains valid even if Stop precedes its tool reply.
- Stop is not a permanent Factory pause. A later newly arriving actionable input
  can start a fresh serial turn after confirmed cessation. The stopped turn's
  input/effects are not automatically replayed. A batch whose delivery is unknown
  requires reconciliation rather than guessed re-delivery.
- Cancel Ticket records cancellation intent immediately and cancels queued or
  active worker execution through KLM. Show Cancelling until cessation is known;
  keep an explicit interrupted/unknown outcome if it cannot be established.
  Completed work is not rolled back and other Tickets continue.
- Normal human messages use next-step steering when Foreman is busy. Provide
  explicit next-turn queueing separately; queued messages have their own visible
  pending state and can be removed before admission. No additional Send now or
  global Factory pause mode is introduced by this proposal.

Native questions/permissions arrive as structured Chat requests with execution
identity, exact supported choices and pending status. Necessary worker questions
use the same Chat request presentation. The Foreman provides concise context;
the structured request is retained/renderable independently of that prose, so a
failed or stopped Foreman cannot make the request visible only in Tickets.
This is the explicit native-Chat adaptation of the DeepSeek attention mechanism.

Human replies address the exact pending request. Question answers and permission
decisions remain different operations. Acknowledging receipt is not approval or
proof of native resumption. Retries preserve the same request/reply identity and
conflicting second answers are refused. No permission is synthesized from a
normal chat sentence, elapsed time or a worker report.

Use the DeepSeek concise-message policy and a 100-word ceiling for Foreman's
notification prose; preserve complete questions, choices and evidence in their
structured request/result views rather than truncating them to fit. Direct
operational answers remain concise. Internal reports are inspectable through
linked Ticket/Assignment results, not broadcast as progress messages.

Pending delivery, successful publication to Chat and human response are separate
facts. On runner/Foreman unavailability, Chat retains the request or outgoing
delivery with a visible problem/retry state; it does not invent an assistant reply
or claim that the person saw it. Closing the UI preserves pending work. If the
Werbert backend is disconnected, KLM retains owned pending requests/events for
recovery; it does not redirect them into ordinary personal KLM chats.

### 6.5 Configuration changes

Distinguish **saved configuration** from the **configuration applied to a turn**.
Show a pending-change indicator when they differ.

1. Validate settings against the connected KLM's catalogs and Library; save the
   versionable file without interrupting a turn. A structurally valid external
   edit is also revalidated before execution.
2. Capture all Foreman settings and selected resource revisions at each complete
   turn boundary. New steering inside an admitted turn retains that capture.
3. Apply pending changes only after the prior native turn, tools and required
   delivery effects have settled. Do not wait for independent Ticket workers.
4. Instructions, model/effort, harness, selected tools or their native manifest
   changing creates a new native context generation under the same Werbert Chat.
   This simple first-slice rule avoids rewriting old native manifests or claiming
   cross-harness continuation. Rehydrate it from the evidence in section 6.6.
5. Already-admitted Assignments keep their captures through completion. An edit
   does not mutate a running worker, grant tools or trigger its retry. Later
   delegations capture the then-current saved Agent revision.

Missing harness/model, invalid effort or unavailable selected resources block the
next affected admission visibly. Do not silently choose a different model, remove
a selection or run the last valid configuration after an invalid desired edit.
Already-running work retains its independent execution contract. Readiness comes
from KLM; Werbert adds no dependency installation or mandatory memory bootstrap.

### 6.6 Context, compaction and recovery

Use native harness compaction for this slice. Add no AI Memory dependency,
pre-compaction consolidation, configurable context percentage or persistent
self-improvement loop. This is a proposed simplification for review, separate
from the approved DeepSeek input/background mechanism.

Werbert retains the complete Foreman transcript and work records; KLM retains the
native execution history it actually observed. Compaction reduces working context
without deleting either authoritative history. Native compaction is recorded when
observable; unsupported/missing telemetry remains unknown rather than fabricated.

Supply a bounded live-work overview at each turn, with exact IDs, objectives,
Assignment state/result references and pending human requests. Large results use
a preview and bounded full-result reads. Foreman's operational history/result
tools can retrieve older details without inserting complete logs into every turn.
No model-generated summary becomes proof of completion, authorization or delivery.

After an ordinary idle restart, KLM may continue a verified compatible native
context. After a deliberate configuration generation change, start a fresh
context with the role, selected configuration, recent conversation and live-work
overview; older history remains consultable. Preserve the user-visible Chat ID.

After a lost active turn, reconcile exact work/input/operation identities first.
Reconnect to known continuing work or record interruption without replay. Never
resend the old human input to manufacture a completed turn. Unknown historical
outcomes remain recorded and do not automatically block unrelated settled work.
Factory loading/history inspection creates no inference or new execution.

## 7. Scope and review status

| Area | Status |
| --- | --- |
| Separate Werbert Factory with KLM runner; one local engine | User-selected direction. |
| Shared Library list/select, versionable Foreman settings | User-selected direction. |
| Hidden external execution and shared current checkout | Explicitly approved. |
| Ordinary KLM compatibility | Explicitly required whether Werbert is disconnected or connected/working; both interface and personal-session behavior remain unchanged. |
| DeepSeek internal background/input/notification mechanism | Explicitly required; implementation evidence still needed per harness. |
| Stop only Foreman's current turn; separate Ticket cancellation | Explicitly approved; detailed event/queue races above are proposed for review. |
| File schema, local handshake, delegation protocol, configuration generations and native compaction | Proposed here under delegated design work; await document approval. |

The first slice covers local connection, KLM-project selection, Factory definition,
Foreman settings/Chat, existing-Agent discovery, Ticket/Assignment inspection,
one worker per Ticket, relevant human requests, result assessment, cancellation
and the runner mechanisms needed for that path. It does not include agent/Library
authoring, graph invocation/authoring integration, automation triggers, worktrees,
multiple engines, team/dependency scheduling or migration of old operational work.

Harness coverage is an implementation preflight item, not silent scope reduction:
the intended KLM integrations are Pi, OpenCode and Codex. Record actual support for
external tools, next-step delivery and five-second jobs for each target. If a
required mechanism cannot be supported, surface that concrete gap before proposing
any narrower first delivery. No matching prompt or API signature is runtime proof.

Werbert's existing data and user-authored definitions remain preserved reference
material. The implementation plan must identify a non-destructive new admission
path before replacing legacy environment/native-executor startup. This document
does not authorize deleting the old code, resetting databases or deploying.

Review sections 6 and 7 together with the runner appendix before implementation.
Any necessary deviation from the pinned reference must be documented rather than
silently adding behavior or claiming literal equivalence.

These documents specify intended behavior and proposed integration contracts.
They are not an implementation plan: sequencing, concrete code changes, schema
transitions, adapter preflight results and focused verification steps still need
to be written against the reviewed specification before implementation begins.

## 8. Frontend-first validation phase

The user explicitly requested two sequential Frontend Specialist subagents:
first write the mocked frontend implementation plan and return it; then a
different specialist implements that plan. The present authorization covers the
interactive mocked Werbert experience, not real Factory/runner backend work.

Cover connection/discovery, project selection/Add Factory, initial Foreman
configuration, shared Library selection, Chat, Tickets/Assignment inspection and
the relevant pending/working/result/error/Stop/cancellation states described above.
Make the journey interactive with mock state; no real engine connection,
credential provisioning, permission decisions, model execution or project writes
are needed to validate the frontend. Keep implementation limitations in the
documentation rather than filling the UI with mock/tutorial labels.

Use the existing Werbert frontend/design system and preserve its current dirty
checkout. The exact KLM Chat timeline/composer is the explicit exception to using
Werbert's older screen as the visual reference. Specialists must follow Werbert's
applicable documentation rules and use their selected writing-for-agents skill
for authored plan/change records.

Maintain a durable experience change log alongside the mock's maintained frontend
documentation. Record each difference from these specs, its reason and whether it
is a proposal, explicitly requested adjustment or confirmed user acceptance; map
it to the future real-backend behavior/API implications. Subsequent user feedback
must update the same record so the backend implementation plan can incorporate
the validated experience instead of reverting to an obsolete specification.

## 9. Current work status

Specification and frontend-mock delegation preparation. Real backend/runner
implementation, operational migration and provider-backed validation have not
been performed under this slice. Existing KLM and Werbert modifications remain
their owners' work. Mock delivery status belongs to the specialist handoff.
