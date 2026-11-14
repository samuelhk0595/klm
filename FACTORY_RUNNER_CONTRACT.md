# Factory / KLM Runner — Proposed First-Slice Contract

Status: **Design proposal for review, not an existing API or implementation.**

Companion to [the Foreman specification](FACTORY_PERSISTENT_AGENT_SPEC.md).
Scope: one local Werbert backend and one KLM engine, under the same OS user.
The public API names below are proposed; current ordinary-session HTTP routes
do not provide this complete contract.

The external runner is additive. Whether Werbert is disconnected or connected,
ordinary KLM interface and personal-session semantics remain unchanged. Apply
external tool manifests, input routing and shell/job policies only to externally
owned captures, never to global harness settings or existing personal sessions.

## 1. Ownership and execution identities

- Werbert owns the Factory, Chat, human-input ledger, Ticket, Assignment and
  objective assessment. It persists their links to runner IDs.
- KLM owns external runner conversations, native context generations, admitted
  turns/work executions, background jobs, native tool calls and pending requests.
- The authenticated local integration binds execution ownership outside model
  arguments. Display names, arbitrary owner fields and project files do not grant
  access or establish that a message was written by a human.
- Retain a server-owned integration ID separately from the current paired-client
  credential. Re-pairing requires an explicit authenticated reconnection to that
  binding, not a new owner or duplicate work. The first slice retains KLM's
  existing paired-client trust model; this ownership metadata is not a new
  isolation guarantee against other fully authorized KLM clients.
- Correlate engine, project, external conversation, work, native generation/turn,
  input, tool call and human-request IDs wherever applicable. Ticket/Assignment
  IDs are opaque client references to KLM, not KLM-owned business entities.
- External execution uses the existing registered project's directory and is
  excluded from ordinary KLM session/activity/request projections at their source.
  Foreman/worker role manifests do not inherit private general/linked-agent tools.

No SQL is shared between the two applications. Technical completion belongs to
KLM; completing the business objective belongs to Werbert. A completed turn does
not automatically complete an Assignment or Ticket.

## 2. Existing surfaces to reuse

Reuse KLM's client pairing, engine identity, project inventory, model catalogs,
saved Agent catalog, Library selection/snapshot machinery and native adapters.
Examples of current surfaces include `/api/state`, `/api/agents`,
`/api/projects/{id}/models/{harness}` and `/api/library`.

The local backend obtains and redeems a pairing code after the user's Connect,
keeps its own credential and sends the current storage contract/engine identity.
This proposal adds no unauthenticated loopback exception or new multi-engine
credential product. A changed engine identity is not the same runner association.

Werbert writes its own portable definition locally using the registered project
path; this first local slice needs no general remote-file-writing API.

## 3. Minimal new runner surface

All mutations take a bounded stable `operationId`; unchanged retries return the
original receipt, and changed reuse conflicts. Work is persisted before acceptance
is reported. Inspecting a receipt is separate from inspecting current work state.

| Proposed operation | Contract |
| --- | --- |
| `GET /api/runner/capabilities?projectId=...` | Engine identity, runner protocol version, per-harness availability and support for external tool manifests, next-step delivery, native continuation, shell jobs and selected capabilities. Unsupported means unavailable, not a fallback. |
| `POST /api/runner/connections` | Establish or explicitly reconnect the authenticated local Werbert binding, returning its stable integration ID. Credential replacement preserves existing execution ownership; a supplied display name alone never rebinds work. |
| `POST /api/runner/conversations` | Create/reuse one externally owned coordinator conversation from the client's durable conversation key, project and captured execution configuration/tool manifest. Admission alone starts no model turn. |
| `PATCH /api/runner/conversations/{id}/configuration` | Save a validated desired configuration against its expected revision. Apply it as a new generation only at the drained boundary defined by the Foreman specification; preserve the applied capture until then. |
| `POST /api/runner/conversations/{id}/inputs` | Retain attributed `human`, `application_event` or `tool_result` input with immutable source ID/payload and `steer`, `queue` or `passive` delivery. The trusted integration supplies original human provenance; a model cannot relabel an event. |
| `DELETE /api/runner/conversations/{id}/inputs/{inputId}` | Remove an exact unsent queued human input, with a durable operation receipt. Refuse a removal once native admission has claimed it; it is not a cancellation of running work. |
| `GET /api/runner/conversations/{id}` | Current generation/turn, input and interruption state, applied configuration, pending requests and history/result references. |
| `POST /api/runner/work` | Admit an externally owned Agent task from exact project, Assignment key, saved Agent source/revision, task and worker tool manifest. Capture effective settings/Library and return a work ID promptly; no all-child blocking join. |
| `GET /api/runner/work/{id}` | Observed execution status, captured configuration, cessation/uncertainty evidence and bounded result references. |
| `GET /api/runner/events?after=...&limit=...` | Ordered retained events for the integration with opaque exclusive cursor, bounded page and high-watermark. Include exact conversation/work/request IDs. |
| `GET /api/runner/content/{id}?offset=...&digest=...` | Bounded immutable full tool/result/history content behind compact events; EOF means content EOF, not success. |
| `POST /api/runner/tool-calls/{id}/reply` | Reply to one captured pending application-tool call, with exact identity, result/error and stable reply operation. Duplicate replies do not repeat the tool. |
| `POST /api/runner/requests/{id}/reply` | Commit a human question answer or exact offered permission/elicitation decision for its owning execution. Never interpret arbitrary prose as approval. |
| `POST /api/runner/conversations/{id}/stop` | Request cessation of the expected current coordinator turn and its owned jobs; independently admitted worker work is unaffected. |
| `POST /api/runner/work/{id}/cancel` | Cancel exact queued/running work. Receipt means requested, not confirmed cessation. |
| `GET /api/runner/operations/{operationId}` | Recover the original admission/control receipt after an uncertain HTTP response without creating another execution. |

Do not expose graph execution in this initial runner slice. Graphs remain KLM
capabilities, with a separately specified external admission contract later.

## 4. Captures and application tools

Coordinator capture includes harness/model/effort, assembled prompt digest,
selected skill/MCP tool references, role manifest revision and current checkout.
Worker capture additionally pins the saved Agent's full source identity/revision.
KLM materializes the actual execution resources using its existing mechanisms.
Missing capabilities fail admission explicitly; no stripping or silent substitution.

Foreman's execution capture always requests YOLO, following the user's fixed role
requirement. Werbert shows that state read-only in Chat. This does not grant new
tools, expand the captured task scope, auto-answer MCP elicitation or change a
worker's separate captured permission policy. The current frontend-mock phase
only displays/simulates this requirement and performs no real runner mutation.

Werbert provides a bounded JSON-schema operational-tool manifest at native context
creation. KLM makes only that external-role manifest and approved execution tools
available. It emits durable `tool_call` events for Werbert to handle. This is not
arbitrary access to KLM's internal role catalog or permission to select new Library
content. Each harness must demonstrate the actual native tool surface.

Minimal Foreman operations: discover/read saved Agents, open/read a Ticket, assign
its captured task, inspect owned work/results/history, relay a human action,
publish a purpose-bound Chat reply and assess the Ticket objective. These execute
in Werbert and validate the retained original human input and Ticket scope.
They are not new KLM Factory business endpoints.

Minimal worker operations: acknowledge the original request, report progress or a
structured human action, consult Foreman about intent/context, read owned evidence
and finish with a structured result. The saved Agent retains its expertise and
selected external tools; it acquires no Factory administration or nested
delegation authority from this wrapper.

Werbert journals the logical operation before a side effect and stores its result
before returning it to KLM. KLM records reply/native inclusion separately. If Stop
or disconnect occurs between those facts, reconcile the original operation rather
than execute it again. Admission, native acceptance, inclusion, human-facing
delivery and final settlement are distinct states, not one success boolean.

## 5. Responsiveness, jobs and delivery

KLM must provide the five-second shell handoff and bounded job waits required by
the reference for these external roles. Configure/wrap supported native tools
without changing ordinary personal-session defaults. A tracked command continues
once in the same owned process domain, retains output and returns a job reference.
Its completion is a durable attributed result; Stop/cancel follows actual job
ownership. A process restart does not rerun a lost command to recreate its job.

At model boundaries, admit only the input batch selected for that step. New
arrivals cannot mutate an in-flight request. Steering may enter within a turn;
queue waits for a new turn; passive context alone does not wake an idle agent.
Maintain one active coordinator turn. A long model response is not interrupted by
the shell wait limit. Do not infer five-second handling for arbitrary remote MCP
tools; unsupported long-tool behavior must be diagnosed and mapped explicitly.

Operational Factory handlers perform durable admission/read operations and return
promptly; delegation never waits for the worker result. Any necessary long
non-shell operation needs explicit job semantics rather than a misleading
five-second timeout that leaves unknown remote effects.

SSE, if used, is a wake hint. Werbert's persisted cursor and the event/content APIs
provide recovery. Commit a received event, its local state changes and notification
intent before advancing the cursor. Deduplicate by engine/event identity. Routine
progress is retained for inspection, not delivered as a coordinator wake. Relevant
results and pending human actions are actionable. No frontend subscription is
required for runner execution or durable event storage.

## 6. Reconnection, Stop and context changes

Disconnect does not imply worker cancellation. Reconnect checks the saved engine
identity and original work/operation IDs, reads current state and catches up from
the committed cursor. Do not create a replacement conversation, worker or human
request just because a response or event stream was lost.

Stop/cancel must confirm owned process/tool cessation before claiming cancellation
or admitting a conflicting replacement context. Incomplete native/tool/external
evidence remains explicitly unknown. A stopped coordinator is not a cancelled
Ticket, and cancellation is never an automatic execution retry.

Native contexts retain their frozen manifest. Changed Foreman configuration starts
a fresh generation only after a drained turn and relevant tool/effect boundaries;
it preserves the same Werbert conversation and independent workers. Compatible
idle continuation may reuse the verified native identity. Native compaction is
the initial policy; full transcript/work history remains available independently.

## 7. Implementation preflight, not delivered guarantees

Before coding, map each proposed operation to existing KLM services and identify
the smallest new external-execution records/hooks. Preserve the current dirty
checkout and unrelated personal-session behavior. No second harness adapter stack,
independent scheduler for the same work or automatic legacy-data migration is
introduced in Werbert.

For Pi, OpenCode and Codex, record actual support for the captured external tools,
serial input boundaries, five-second shell/job behavior, native requests,
Stop/cessation and continuation. These are native runtime checks, not inferred
from model names, prompt instructions or existing protocol signatures. Raise a
concrete incompatibility before reducing coverage or changing semantics.

The proposed Foreman journey must prove a real request -> Assignment -> runner
worker -> human intervention -> result -> independent objective assessment, with
the work absent from ordinary KLM views. Run only focused checks appropriate to
the approved implementation; this proposal does not authorize broad test suites,
deployments or production work.
