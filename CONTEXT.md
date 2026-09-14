# KLM

Shared terminology for the interface to coding-agent harnesses.

## Language

### Engine, Desktop and Focus

**Engine**:
The independently installed Go runtime for persistence and harness execution.
`klm start` launches its background worker; `klm stop` requests local graceful
shutdown. Its public API listens on port 7331; user data is separate from binaries.
The Windows installer registers per-user login startup, not a pre-login service.

**Desktop client**:
The separately installed Tauri 2 application in `clients/desktop`. It contains the
native window, tray and static frontend server on port 7332. Window close hides;
tray Exit ends desktop/web without stopping the engine. It has no login autostart.

**Focus**:
The browser presentation of the same application, opened by the desktop header
button. It retains the gradient and draggable central shell. It is available at
`http://localhost:7332` locally or `http://IP-OF-ENGINE-COMPUTER:7332` on another device
while Tauri runs. API/SSE use the page's host on 7331. It does not transport drafts
between clients, execute on the viewing device, or start a conversation/run.

**Listen address**:
`0.0.0.0` means all IPv4 interfaces; it is not an address to open in a browser.
Public API/web network access is distinct from the private loopback harness bridge
and per-user local engine-control pipe.

**Session Status Bar**:
The compact row below the message input showing session-related indicators: skills,
MCPs, quota, context-window usage, and input/output token counts.
_Avoid_: Information labels below input, composer status bar, session stats footer.

### Graph authoring

**Graph**:
A reusable definition of activities, available outcomes, and transitions that guide
work through agent, terminal command, Fork, and Join nodes.

**Graph run**:
One execution of a graph for a submitted task, with captured graph/agent definitions,
node activations and private native-session history. Its records persist in the
engine; this does not imply a browsable execution-history UI.
_Avoid_: Graph (when referring to one execution).

**Graph activity**:
An authorized objective scheduled in a conversation. It may await dependencies or
the user and may require multiple corrective runs. A scheduled activity is not yet
a run; completion is assessed by the orchestrator before it counts as success.

**Selected graph**:
The graph chosen for a chat from its project's catalog. It is not an execution
authorization or the only graph the user can authorize the orchestrator to invoke.
Selection is persisted per conversation. Selecting/opening it does not start work;
selecting None hides Graph without cancelling an active run.

**Orchestrator**:
The agent in the user's chat that invokes authorized graph runs and acts as the
communication bridge between those runs and the user.
_Avoid_: Initial graph node, Fork/Join coordinator (when referring to the chat agent).

**Run authorization**:
The user's express permission to invoke a graph for a requested activity, given
either in the original request or in response to the orchestrator's proposal; it
also covers further attempts toward that same objective, including when the
orchestrator interprets a completed result as insufficient.

**Activity authorization reference**:
An entry `{eventId,text}` in invocation `authorization`, grounding the requested
activity in the conversation. The engine validates an existing `user` event in
that conversation; the reference is traceability, not semantic proof of consent
or a native permission grant. It is separate from choosing the run workspace.

**Run workspace**:
The directory context for an invocation, distinct from the session's visual
grouping folder. Decision approved 2026-09-13: omitted `workspace` defaults to
`original`, the registered project directory, without a question or mandatory
workspace-choice message reference. The product rule supplies that default;
omission is not express consent. Explicit `original` and `new_worktree` modes
remain available. `original` is not an arbitrary path, and
`workspace.authorizationEventId` is not mandatory. This supersedes the earlier
requirement to ask when workspace selection was omitted.

**Fresh/reuse decision**:
The user's choice to start a new attempt fresh or build on earlier artifacts and
workspaces. It remains required when applicable to retries; the original-directory
default does not decide it. Reuse requires explicit associations; fresh in the
original directory uses existing files without an engine reset (W1/W2).

**Run task**:
The self-contained text prompt prepared by the orchestrator for a graph execution,
available as `run.input.task`; it is not the conversation's history.

**Agent node**:
A graph activity that references a reusable agent and may override its harness,
model, and effort. Multiple nodes can reference the same agent while remaining
distinct activities. Each agent node has its own human-readable name, independent
of the reusable agent's name and identity.
_Avoid_: Agent definition (when referring to an occurrence in a graph).

**Node activation**:
One occurrence of executing a node within a graph run, distinct from that node's
reusable identity and from other visits to the same node.
_Avoid_: Previous run (when referring to an earlier visit within the same graph run).

**Choice**:
A named outcome available to linked AI agent and Join nodes, whose valid accepted
submission communicates the end of that node's current work. It can end the run or
pass a structured output to one destination; it does not necessarily mean success.
_Avoid_: Executable node, parallel branch.

**Terminal command node**:
A named graph activity that runs a user-configured terminal command without
selecting an AI agent, and passes explicitly configured output along its structural
connection. It replaces the previous Code node concept and can be the initial activity.
The current executor runs one noninteractive PowerShell script with `$payload` as
data; `command.result` is available to its output mapping. Absent output means `{}`.

**Node/tool-call finality**:
The confirmed end of a node's work: no new tool calls after sealing, and calls
actually started have completed before Choice acceptance. Decision approved
2026-09-13 (refinement section 14): this guarantee excludes shutdown of shared MCP
servers and detached remote tasks continuing beyond a returned call. Configured
external MCPs in OpenCode/Codex are allowed without a server-name allowlist or
disabling them; their presence alone is not evidence of unresolved work.

**Detached external task**:
Work that continues outside the node after a tool returns, for example a
"task started" response. That response completes the call's lifecycle responsibility,
not the remote task or the user's objective. The engine does not guarantee its
termination or cancellation, or shutdown of the shared server that hosts it.

**Uncertain finality**:
Missing confirmation that the node or an actually started call finished, including
error/cancellation or a lost callback without completion evidence. Local shutdown
or a cancellation request does not prove remote cancellation and must not become
an accepted Choice or normal result. This state remains distinct from a detached
task whose tool call already returned. Grants and sandbox rules are unaffected.

**Choice input payload**:
The arguments supplied when a node selects a choice, governed by that choice's
required and optional fields.
_Avoid_: Choice output payload.

**Choice output payload**:
The structured data assembled from configured values and input references for the
destination of a choice.
_Avoid_: Transition prompt (as a standalone transfer configuration), choice input payload.

**Payload field name**:
A field identifier normalized like a Choice name: lowercase a-z, digits and
underscores; whitespace becomes underscores, accents are folded and other
symbols removed. Leading/trailing underscores are stripped on completion.
This applies to Choice input/output and Fork output field names, not their values.

**Choice session policy**:
Controls starting or continuing the destination agent's session. Applies only
when the destination is AI agent or Join, not Fork or Terminal command.

**Blocked**:
An escape outcome for an agent that cannot supply the data needed to complete its
activity through the intended choice. It is distinct from a run failing because
an invalid result was accepted or invalid submissions exhausted their limit.
The engine-owned blocked identity is separate from any author-defined Choice named
blocked; an authored name alone does not determine the run's status.

**Fork**:
A control node that starts parallel branches through structural connections.
Each branch has its own manually configured output JSON; incoming data is not
automatically duplicated across branches. Branch outputs are configured on the
Fork and can reference `payload.<field>` and `run.input.task`. Each branch chooses
separate-worktree isolation (default true) or the inherited workspace. Nested Forks
are outside the current execution scope.

**Work branch (ramificação)**:
A path of work started in parallel by a Fork, which may contain multiple nodes
and choices. Its human-readable name is free text, such as "Security review".
_Avoid_: Git branch (when referring to the path of work).

**Reconvergence**:
The meeting of parallel work paths at a Join, after which the graph continues
along a single sequential path.
_Avoid_: Monolithism (when referring to the return from parallel to sequential flow).

**Join**:
A synchronization and integration node that delegates integration work to a
selected reusable agent. The engine supplies the participating worktrees and
integration instructions; the author may add an optional prompt and an optional
output Git branch. Its agent reports its outcome through a Choice, like other
AI agents. Synchronization and flow remain engine responsibilities.

**Join output Git branch**:
An optional name base for a newly created integration branch/worktree that becomes
the continuation workspace. When omitted, integration returns to the workspace
and branch that originated the parallel work.

**Join round**:
One collection of fresh incoming deliveries and the resulting Join-agent
activation, distinct from earlier or later rounds at that Join within the run.

**Git branch**:
A named Git reference tracking a line of commits. Its name follows Git naming
rules. A Git branch alone does not provide a separate working directory.
_Avoid_: Ramificação (without qualification when referring to Git).

**Fork worktree**:
A separate Git working directory for isolated work on a Fork branch, with its
own HEAD, index and Git branch. Fork branches may instead share the inherited
workspace; parallel execution alone does not imply a new worktree.

**Incoming revision**:
The Git revision delivered to a Fork by the execution, resolved to one commit
when that Fork activates. All worktrees and new Git branches created by that
activation start from that same commit. This is execution context, separate
from the manually configured output payload.
_Avoid_: Current HEAD (without identifying and fixing the incoming revision).

### Run monitoring

**Graph preview tab**:
The read-only view of the graph selected in the conversation's composer, distinct
from graph authoring. The same Graph tab shows active/completed-node progress when
that selected graph has an active run in the conversation.
It uses the run's captured graph while active and real catalog configuration while
idle. The initial monitor has no Run/Input/Output panels or historical run browser.

**Active run**:
A run occupying the conversation's execution slot, including starting, running,
human/input waits and ending. It does not require an agent to be executing a tool
at that instant. Its LED appears in the composer when the selected graph matches.

**Collecting Join**:
A Join waiting for its incoming deliveries. Collection is shown separately from
the Join agent's Running state; an empty payload still counts as a delivery.
