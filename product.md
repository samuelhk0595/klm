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
For now, "interface for coding-agent harnesses" describes the idea without
assuming automatic coordination across multiple agents.

In this document, "harness" means the external software that receives delegated
work and manages its execution with an agent. It is not just the language model.

## Current Definition Boundaries

The first real slice is a separate Go engine with persistent projects and sessions,
and a React client. A project has a name, directory, and icon. Each session belongs
to one project and uses Pi, OpenCode, or Codex headlessly. The client presents the
messages, exposed reasoning, commands, and MCP/tool events those harnesses emit.

The engine remains independent of the client. A private, KLM-managed MCP bridge
connects linked conversations for owned OpenCode and Codex processes; Pi uses an
automatically loaded extension for the same operations. A general-purpose public
engine MCP interface remains outside this slice.

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

The header button beside Session log opens the side chat. Selecting a passage in
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

## Graph Selection in Chat

The main chat composer has a graph selector at the bottom left, opposite model
and effort controls. It shows graphs belonging to the selected project and the
selected graph's name. Lists longer than five graphs enable search and scrolling.
This slice is selection UI only: graph invocation, execution and its relationship
to sending a chat message still need definition.

The next UI slice will prototype graph-run tracking in the chat timeline and a
read-only graph canvas using fictional run data. It does not yet define or
implement graph execution, activation controls, or the detailed tracking layout.

Further decisions will be made as features are discussed and validated by the
user. They must not be treated as already approved requirements.
