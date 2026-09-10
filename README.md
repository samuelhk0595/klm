# KLM

A desktop-style React client for a separate Go engine that delegates work to
Pi, OpenCode, and Codex. Product scope is defined in [product.md](product.md).

## Run Locally

Requires Go 1.24.2+ and Node.js 22.12+. Start the engine from the project root:

```powershell
go -C engine run .
```

In another terminal:

```powershell
npm --prefix clients/desktop install
npm --prefix clients/desktop run dev
```

Open http://127.0.0.1:5173. The engine listens on http://127.0.0.1:7331.
`VITE_ENGINE_URL` overrides the client endpoint. Do not expose the engine to a
network: it is a trusted-local-user API without remote authentication.

## First Slice

- Add a project using the rail's plus button. The engine opens a native Windows
  directory picker; choose its name and icon, then save. Directory is read-only
  with a Change directory button. Icons are optional and stored locally.
- Create a session under **Sessions** and choose Pi, OpenCode, or Codex. Select a
  model and supported effort/variant in the composer picker left of Send. Existing
  CLI credentials and model configuration are used; authenticate in the CLI first
  if necessary. Changes apply to the next turn.
- Send messages. The conversation displays actual text, exposed reasoning,
  commands, MCP calls, and generic tools emitted by the selected harness.
- Stop an active response, reload, or restart the engine. Projects, grouping
  folders, history, and native conversation identifiers are persisted. Unsent
  drafts last only until the browser reloads.
- Pending permissions appear above the composer for all three harnesses. Allow
  once, remember for the session, remember for this project/harness, or deny.
  Remembered grants match the exact displayed action/resources; they do not grant
  unrestricted tool access. Session grants survive restart with that logical session.
- Agent questions appear above the composer with option buttons and custom input
  when allowed. Select choices or type an answer, then **Send answer**, or **Dismiss**
  the request. The waiting tool receives your answer and the turn continues.
- Open **Side agent** beside Session log, or select a main-chat passage and choose
  **Ask side agent**. One persistent side conversation belongs to each main chat;
  it has independent harness/model controls, drafts, scrolling, permissions, questions,
  and Stop. Closing hides it. Drag the divider, or focus it and use arrow keys,
  Home, or End, to resize; narrow layouts show the side chat full-width.
- Both agents can retrieve or consult the other conversation across Pi, OpenCode,
  and Codex. Expand consultation activities for questions, answers, execution details,
  and cancellation. Busy recipients queue; replies arrive through the tool or a
  correlated continuation after the requesting turn finishes.

Old browser mock projects are left untouched in localStorage but are not imported:
their folder labels were not reliable absolute paths. The live app starts with
engine data only. Static fixtures remain isolated in the design-system gallery.

## Boundaries

The **Session Status Bar** (`clients/desktop/src/features/chat/SessionStatusBar.tsx`)
is the compact row below the message input. It shows live input/output token totals
and latest context usage; skills and MCP counts remain static reference values.
Quota/reset appears for verified direct subscription connections. Other routes
(including OpenRouter and API-key billing) have no quota slot.
Unavailable token/context metrics stay hidden and appear when reported by the harness.
Reported zero values remain visible. See [CONTEXT.md](CONTEXT.md) for shared terminology.
The side status bar shows only its own token counts and context usage.

No attachments, graph execution, model-account management, or general-purpose
engine MCP interface is included. The private linked-agent bridge is configured
only for KLM-owned harness executions. The engine normalizes only the events the harness
actually emits, so some text/reasoning arrives in completed blocks, not tokens.
Pi runs with only an engine-owned extension, providing `ask_user`, linked-agent tools, and gating other
agent-dispatched tool calls; third-party extensions and their MCP tools remain disabled. It is not
an OS sandbox. OpenCode uses its server API and preserves existing permission rules.
Codex uses app-server with read-only sandboxing and user/on-request approvals.
Some Codex permission-profile grants last for the whole turn; the card labels that
scope. Ambiguous requests cannot be remembered, and unsupported forms are declined.
No blanket native policy amendments or approval/sandbox bypass flags are passed.

OpenCode questions honor native session permissions. Older sessions created by
`opencode run` may retain a question-tool denial; use a new session in that case.
Codex's Default-mode question tool uses its under-development feature flag, set
only on the engine-owned child process. Global config and selected models are
unchanged. Pi questions use the engine-owned `ask_user` tool and RPC bridge.

Pending user questions and permissions cannot survive a stopped process; they are
cleared on stop/restart. Linked-agent requests retain terminal interruption/cancellation
records and are not automatically replayed after restart.
The engine stores remembered approvals in its local state, not in global harness
configuration. A revoke/manage-grants UI is not included yet.

Engine data defaults to `%APPDATA%\klm\engine` on Windows. It is plaintext local
data; protect it like the project itself. The native directory picker currently
supports Windows only, including empty directories and full absolute paths.

See [engine/README.md](engine/README.md) for the API, adapter details, executable
overrides, persistence, and security boundaries. See
[clients/desktop/README.md](clients/desktop/README.md) for frontend structure.

## Validation

Use `go -C engine build` and `npx tsc --noEmit` from `clients/desktop` for lightweight
compilation checks when needed. Full test suites and adversarial reviews require
an explicit request; human validation is the feature acceptance gate.
