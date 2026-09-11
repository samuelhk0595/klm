# KLM

Shared terminology for the interface to coding-agent harnesses.

## Language

**Session Status Bar**:
The compact row below the message input showing session-related indicators: skills,
MCPs, quota, context-window usage, and input/output token counts.
_Avoid_: Information labels below input, composer status bar, session stats footer.

### Graph authoring

**Graph**:
A reusable definition of activities, available outcomes, and transitions that guide
work through agent, terminal command, Fork, and Join nodes.

**Graph run**:
One execution of a graph for a submitted input, with its own activity and session
history.
_Avoid_: Graph (when referring to one execution).

**Selected graph**:
The graph chosen for a chat from its project's catalog, intended as the target
of a future graph invocation. Selection alone does not start a graph run.

**Agent node**:
A graph activity that references a reusable agent and may override its harness,
model, and effort. Multiple nodes can reference the same agent while remaining
distinct activities. Each agent node has its own human-readable name, independent
of the reusable agent's name and identity.
_Avoid_: Agent definition (when referring to an occurrence in a graph).

**Choice**:
A named outcome declared in a graph and available to AI agent and Join nodes
linked to it.
It can end the run or pass a structured output to one destination.
_Avoid_: Executable node, parallel branch.

**Terminal command node**:
A named graph activity that runs a user-configured terminal command without
selecting an AI agent. It can be the initial activity of a graph, for example
to run git pull before agent work. It replaces the previous Code node concept.
Shell, execution environment, result transport and failure behavior still need
refinement.

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

**Fork**:
A control node that starts parallel branches through structural connections.
Each branch has its own manually configured output JSON; incoming data is not
automatically duplicated across branches. Branch outputs are configured on the
Fork. Data references and the execution contract still need refinement.

**Work branch (ramificação)**:
A path of work started in parallel by a Fork, which may contain multiple nodes
and choices. Its human-readable name is free text, such as "Security review".
_Avoid_: Git branch (when referring to the path of work).

**Join**:
A synchronization and integration node that delegates integration work to a
selected reusable agent. The engine supplies the participating worktrees and
integration instructions; the author may add an optional prompt and an optional
output Git branch. Its agent reports its outcome through a Choice, like other
AI agents. Synchronization and flow remain engine responsibilities.

**Join output Git branch**:
An optional target branch for integrated work in the shared repository. It does
not mean switching branches in the main working directory. The destination when
omitted and the handling of existing branches still need refinement.

**Git branch**:
A named Git reference tracking a line of commits. Its name follows Git naming
rules. A Git branch alone does not provide a separate working directory.
_Avoid_: Ramificação (without qualification when referring to Git).

**Fork worktree**:
A separate Git working directory assigned to one work branch, with its own
HEAD and index and its own Git branch. Each work branch uses its worktree through
its activities up to integration. Worktrees share the repository's objects and
branch references, but do not share checked-out files.

**Incoming revision**:
The Git revision delivered to a Fork by the execution, resolved to one commit
when that Fork activates. All worktrees and new Git branches created by that
activation start from that same commit. This is execution context, separate
from the manually configured output payload.
_Avoid_: Current HEAD (without identifying and fixing the incoming revision).
