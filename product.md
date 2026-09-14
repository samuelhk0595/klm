# Product

## Definition

The product is an interface for users to interact with coding-agent harnesses.
It provides its own interface features and delegates work execution to connected
harnesses.

The user interacts with this software, while harnesses such as Pi, OpenCode,
CoPilot, Codex, and CloudCode carry out the delegated work. These names represent
integrations mentioned in the product vision, not a list of integrations already
implemented. Official names and integration methods will be confirmed as each
integration is defined.

## Product Role

- Be the interface the user interacts with.
- Provide features that support that interaction.
- Connect to external harnesses and delegate the requested work to them.
- Present the progress and results made available by the harnesses to the user.

The product is not intended to replace harnesses or reimplement their execution
mechanisms. The interface is the interaction layer; harnesses are responsible for
executing the delegated work.

## Provisional Terminology

"Orchestrator" is a provisional description, not a definitive classification.
"Interface for coding-agent harnesses" remains the product description. Within
graph execution, orchestrator specifically means the main-chat agent invoking
user-authorized activities; it does not imply unrestricted autonomous coordination.

In this document, "harness" means the external software that receives delegated
work and manages its execution with an agent. It is not just the language model.

## Current Definition Boundaries

The first real slice is a separate Go engine with persistent projects and sessions,
and a React client. A project has a name, directory, and icon. Each session belongs
to one project and uses Pi, OpenCode, or Codex headlessly. The client presents the
messages, exposed reasoning, commands, and MCP/tool events those harnesses emit.

### Windows Delivery and Focus

The engine and Tauri 2 desktop client are installed separately. `klm start` starts
the engine in the background and returns after readiness; `klm stop` requests
graceful shutdown through a local, per-user Windows channel. The engine installer
adds user PATH and starts the engine at login via HKCU Run. Data stays under
`%APPDATA%/klm/engine`; uninstall preserves it. Stopping manually does not disable
the next login start. There is no pre-login Windows service.

The public HTTP API listens on `0.0.0.0:7331`. Network authentication and TLS are
deferred by the approved personal-use contract; grants, sandbox and the private
authenticated loopback bridge keep their existing boundaries. Commands, project
files and native directory dialogs always belong to the engine computer.

The desktop fills its native window. Closing it hides to the tray; **Open** restores
it and **Exit** ends only desktop/web. The client neither starts nor stops the
engine. Opening another desktop shortcut restores the same instance.

While Tauri is active, it serves the same built application on `0.0.0.0:7332`.
The **Focus** header button, between **Session log** and **Side agent**, opens
`http://localhost:7332` in the default browser. The browser retains the gradient
and draggable central window and hides the redundant Focus button. Another device
uses `http://IP-OF-ENGINE-COMPUTER:7332`; API and SSE use that same host on 7331.
Ports remain fixed; collisions are reported instead of selecting another port.

Local development uses API **17331** and Tauri Focus/web **17332**, with Vite HMR
on **5173**. Engine builds with the `dev` tag use a separate `engine-dev` data
directory and control channel; the dev desktop has a distinct app identifier so
development and the installed release can run concurrently.

Persisted data comes from the shared engine. Drafts and local browser/WebView state
remain independent. The desktop must stay active, including in the tray, for page
reloads/assets in Focus. Client autostart is outside this delivery. Packaging builds
and human runtime validation are tracked in `ENGINE_DESKTOP_IMPLEMENTATION_PROGRESS.md`.

The engine remains independent of the client. A private, KLM-managed MCP bridge
connects linked conversations for owned OpenCode and Codex processes; Pi uses an
automatically loaded extension for the same operations. A general-purpose public
engine MCP interface remains outside this slice.

The private bridge also exposes graph operations according to the conversation's
role. Graph-node sessions receive their own Choice capability, without access to
main/side conversation history or tools to invoke other graphs.

Permission prompts are supported for OpenCode, Pi, and Codex. Users can allow a
request, remember the exact scope for the session, or remember it for the project
and harness. Denial remains available. Native grant lifetimes or unsupported
interactive forms must not be represented as narrower or broader consent than the
harness actually supports. Permission UI never implies blanket approval or a
bypass of harness protections.

Agents can ask users questions through their harness tools. The interface shows
clickable choices and an optional custom answer, including multi-select when the
tool supports it. Answers are returned to the waiting tool, not sent as a separate
chat turn, and never become remembered permission grants. Users can dismiss a
question or stop the turn.

## Mobile Client

**KLM Harness** in `clients/mobile` is a Flutter mobile client with two native
screens: a list of saved hosts and an Add host form. Each host has a user-provided
name and an IP address or domain. Save persists it locally in the app; hosts remain
listed across app restarts. The list does not imply an online-status check.

Selecting a host opens its web frontend inside a WebView. An IP without an explicit
port uses the frontend default **7332**. A domain without a port gets no added port,
allowing tunnel/proxy URLs. Explicit ports and HTTP(S) schemes are preserved;
without a scheme, IPs use HTTP and domains use HTTPS.

The KLM logo at the top of the web project rail returns to the native host list.
While loaded, the WebView has no additional native toolbar. Failed loads offer
Retry and a native KLM home action. Mobile does not start servers or relocate
execution/files to the device. Native screens reuse KLM's visual language through
Flutter equivalents of its colors, typography, inputs, buttons and cards.

## File and Folder References

In both main and side chats, typing `@` opens project file/folder suggestions.
Confirming a suggestion attaches that reference to the draft. Plain paths and
unconfirmed `@` text remain ordinary message text. Selected references remain
visible as inline badges in the draft and persisted chat history. Badges show only
the file/folder name; full project-relative paths remain in the message metadata.
They are atomic selections in the editor: deleting a badge removes its reference.
Failed submission preserves the draft.

UTF-8 text files contribute bounded content to the model's initial message.
Folders contribute their immediate children, including subfolder names, without
recursively attaching file contents. Empty files and folders are valid. References
must resolve inside the conversation's project; binary files, images, and PDFs are
not supported in this slice. Discovery respects nested `.gitignore` rules and
excludes `.git` metadata. Explicit folder attachments list real children, including
ignored entries.

KLM accepts up to eight unique paths per message and deduplicates their content.
OpenCode materializes native file parts with its own Read limits. For Pi and Codex,
the engine prepares user-message context: up to 2,000 lines and 50 KiB per file,
2,000 characters per line, or 2,000 entries and 50 KiB per directory. Truncation is
recorded and included in the attached context. The prepared JSON context is capped
at 200 KiB; oversized submissions require fewer/smaller attachments.

The engine persists references and preparation metadata, while the harness owns
the native message content, history, compaction, and subsequent tool execution.
Selecting context does not change tool permissions or sandbox settings.

## Linked Side Agent Conversation

Each main chat can have one persistent side conversation, attached to that chat
and excluded from the top-level session list. The side agent has its own native
harness session, model, draft, execution state, and stop control. Main and side
may independently use Pi, OpenCode, or Codex. A side harness can be changed before
its first turn; model settings apply to subsequent turns.

The Side agent header button opens the side chat (after Focus on desktop). Selecting a passage in
a main-chat message exposes **Ask side agent**, attaching that exact passage and
its source reference to the side composer without sending a prompt. Later
selections join the same conversation. Closing the panel hides it; reopening
restores its history. The panel is resizable with pointer or keyboard, remembers
its width, and occupies the full panel area when two usable chats do not fit.

Context starts with main-conversation identity, project, selected passages and
bounded nearby messages, plus the user's question. Selected passages remain intact;
either agent can retrieve more messages or consult the other. No generated summary
or native fork is required. Starting without a selection provides the same tools.

KLM's internal linked-conversation tools are automatically allowed on its owned
bridge. They do not require user permission cards or remembered grants. Other
MCP servers, filesystem actions, and command execution retain their permission flow.

Consultations run in the recipient's actual native conversation and enter its
native history. The engine distinguishes them from user turns and groups their
output under expandable activity entries with the question, answer, and status.
Idle recipients start answering; busy recipients queue until their turn ends.
Each conversation has only one active turn. Bounded waits and reciprocal-wait
deferral avoid deadlocks. Replies return through the waiting tool or a correlated
continuation after the requester yields. Requests persist cancellation, failure,
and restart-interruption states.

The side Session Status Bar shows its own context usage and input/output token
counts. Quota, skills, and MCP indicators belong only to the main status bar.

This feature does not imply simultaneous turns inside one native session or
automatic switching between harnesses during a task.

## Agent and Graph Authoring

Project agents and graphs can be authored independently of graph execution. Each
agent is a TOML in `.klm/agents`; each graph is a YAML in the sibling `.klm/graphs`.
The project files are the source of truth. Agents use real harness model identifiers
and require effort only when supported by the model. Their prompts live in TOML.

Graph creation starts with a name and optional description. Save persists the graph
definition; node positions autosave on drag release to a separate layout file, even
before the draft's first Save. A layout alone is not a catalog graph. Saved graphs
can be reopened, edited, enabled/disabled and deleted. Agent references are preserved
when agents are renamed; referenced agents cannot be deleted or disabled.

Authoring includes terminal Choice output, explicit Terminal output mapping and
per-Fork-branch worktree isolation, enabled by default. Saving a draft is separate
from the engine's execution preflight. `.klm/harness.toml` remains deferred.
`GRAPH_AUTHORING_REFINEMENT.md` defines the approved contracts; its latest execution
and minimal-monitor decisions take precedence over earlier visual experiments.

## Graph Execution

The Go engine implements asynchronous graph invocation, persistent activities and
runs, and Agent, Choice, Terminal, Fork and Join execution. The main-chat agent
invokes work with references to the user's express authorization and a self-contained
task, available as `run.input.task`. Selection alone is not authorization. One run
may be active per conversation, independently of the invoking chat turn. The
orchestrator receives results and assesses whether the objective was satisfied
before success-dependent activities proceed. Corrective attempts retain the approved
authorization, concrete-correction and explicit fresh/reuse workspace rules.

Invocation default approved on 2026-09-13: omitting `workspace` uses `original`,
the conversation's registered project directory, without asking or requiring a
reference to a workspace-choice message. The product default is the basis for
that directory; omission is not express consent. Explicit `original` and
`new_worktree` remain available; `original` never accepts an arbitrary path.
`workspace.authorizationEventId` is not mandatory. Activity `authorization`
retains `{eventId,text}` references to existing `user` events in the invoking
conversation, providing traceability rather than semantic proof or a permission
grant. Tool schemas must describe the nested authorization/workspace contracts
completely. Orchestrator instructions apply defined defaults and ask only when
missing information or ambiguity blocks execution. This supersedes the earlier
mandatory workspace question, while retaining the user's fresh/reuse decision for
corrective attempts involving earlier artifacts; see refinement section 13.

Each run captures its graph and agent definitions at actual start. Nodes use private
native sessions and explicit input/output mappings. `continue_target` reuses the
node's latest session in that run only with the same working directory/harness and
adapter-confirmed continuity; otherwise it starts a new session. A Choice is reserved,
then accepted only after the adapter confirms node shutdown and completion of
actually started tool calls, with new calls prevented by the seal; reservation
does not dispatch the destination. Invalid submissions and normal endings without
a Choice share a three-violation limit per activation. The engine's built-in blocked
escape remains distinct from an author-defined Choice with the same name.

Terminal executes one noninteractive PowerShell script in its inherited directory,
with `$payload` supplied as data. Output mapping can use `command.result`; a nonzero
exit still follows that mapping, while infrastructure failure fails the run. Fork
starts parallel branches with explicit payloads and either separate worktrees or
the inherited workspace. Join waits for all incoming deliveries and delegates
integration to its agent. The engine records workspace provenance and explicit reuse;
it does not implicitly reset the original directory, commit work or clean worktrees.

Node questions and permissions use the existing cards in the owning main chat and
route answers to the originating native session. Session grants remain local to
that node session. Shutdown/restart records interrupted runs without automatic
resumption, while retaining scheduled activities and pending result notifications.

Windows mechanisms for Pi, OpenCode and Codex are implemented and admitted by the
capability preflight, but `RuntimeValidated=false`: real model tasks and native
session continuation still require human validation. OpenCode requires version
1.18.30 and its owned plugin handshake. Unix graph execution is not enabled.

External-MCP boundary approved on 2026-09-13 (refinement section 14): permit
configured MCPs in OpenCode/Codex without a server-name allowlist or disabling
them. This replaces the preventive configuration ban. The engine guarantees the
node and tool-call lifecycle, not shutdown of shared MCP servers or detached
remote tasks continuing beyond a tool response. A "task started" response finishes
that call's lifecycle responsibility without proving task success. Error/cancellation
or a lost callback without evidence of call completion retains uncertain finality;
local shutdown is not proof of remote cancellation or a normal result. Grants,
approval flows and sandbox settings remain unchanged. Implementation and runtime
validation of this revised boundary are tracked separately.

Nested Forks, worktree cleanup, user interruption/resumption controls, concurrent
runs within one conversation, rich graph inputs/files, activation limits and task
timeouts remain outside this delivery.

## Graph Selection in Chat

The main chat composer has a graph selector at the bottom left, opposite model
and effort controls. It shows graphs belonging to the selected project and the
selected graph's name. Lists longer than five graphs enable search and scrolling.
Selecting a graph adds a **Graph** tab beside **Chat**, below the session title.
The tab displays the selected graph in a read-only canvas with pan, zoom and reset
view controls. Nodes, connections and settings cannot be edited there. The composer
and its status bar are visible only in Chat; Graph uses that space for its canvas.
Selecting **None** hides Graph and returns to Chat if
needed; selecting a graph makes the tab available without opening it automatically.

The selector uses the project's file-backed catalog and persists selection per
conversation in the engine, independently of Send. Starting an authorized run selects
its graph. Selecting or opening Graph never starts execution; selecting None or
switching views does not stop a run or remove its pending questions/permissions.
Catalog errors remain recoverable errors rather than an empty catalog or silent None.

When the selected graph matches the conversation's active run, Graph displays the
captured definition with real active/completed-node progress from the engine.
Parallel nodes can be active together; a new activation takes precedence over an
earlier completion. Join collecting inputs is distinct from its agent running.
The composer graph icon becomes a blinking LED for that matching active run,
including startup, shutdown and human waits. The name remains plain text. On run
completion, activity and LED clear and the view returns to idle configuration;
unexecuted nodes are never marked completed.

The active cards retain the pink-to-blue body shimmer and Running indicator.
Earlier CSV timers and Run/Input/Output panels are historical visual experiments,
replaced in the production path by the minimal monitor approved in section 12 of
`GRAPH_AUTHORING_REFINEMENT.md`. Separate run tabs, execution history/detail panels
and a new sidebar request component remain deferred. Final results still reach
the main-chat orchestrator. Integrated runtime/UI behavior awaits human validation.

The initial agent or terminal node carries a small **Start** badge inside its header.
All node cards have the same moderate resting shadow, independent of the active
card's shimmer. Reduced-motion preferences disable the shimmer.

When no graph run is active, clicking a card opens a right-side configuration panel
in the same visual style as the authoring panels. Values appear as text, not disabled
form controls. It includes agent execution settings and their inherited/override
source, Choice contracts and output templates, Fork branch/worktree settings, Join
settings and terminal commands as applicable. There are no editing actions.
Required/Optional is shown as text in this first version; its final presentation
still needs user validation. Values come from the real project catalog, including
terminal Choice output, Terminal output and each Fork branch's isolation setting.

Further decisions will be made as features are discussed and validated by the
user. They must not be treated as already approved requirements.
