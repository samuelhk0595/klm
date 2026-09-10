# KLM Engine

Small standalone Go HTTP engine for real projects, sessions, and delegated harness
execution. Standard library only; Go 1.24.2. Includes a private linked-agent MCP
bridge; no frontend or general-purpose engine MCP interface is bundled.

## Run

From this directory, in PowerShell:

```powershell
go run .
```

Build and run a standalone executable:

```powershell
go build -o klm-engine.exe .
.\klm-engine.exe
```

Optional flags:

```powershell
.\klm-engine.exe -addr 127.0.0.1:7331 -data-dir "C:\path\to\local-engine-data"
```

- `-addr` defaults to `127.0.0.1:7331`. Only literal loopback IPs are accepted,
  including `[::1]:7331`. No wildcard, LAN, or public binds.
- `-data-dir` defaults to `os.UserConfigDir()/klm/engine`, normally
  `%APPDATA%\klm\engine` on Windows. Relative overrides become absolute.
- Stop with Ctrl+C. The engine cancels its active harness processes and persists
  terminal session snapshots before releasing the data-directory lock.

## HTTP Contract

All mutation requests, including stop and directory picker, require
`Content-Type: application/json`. Stop and picker accept an empty body or `{}`.
Other mutation bodies are a single JSON object; unknown fields are rejected.
Errors are non-2xx JSON objects: `{"error":"message"}`.

| Method | Path | Body / response |
| --- | --- | --- |
| GET | `/api/health` | `{status:"ok",version:1,runningSessions:number}`; 503 on persistence failure |
| GET | `/api/state` | `{projects:Project[],sessions:Session[],harnesses:Harness[]}` |
| POST | `/api/projects` | `{name,folder,icon}` -> 201 Project |
| PATCH | `/api/projects/{id}` | `{name?,icon?}` -> 200 Project |
| DELETE | `/api/projects/{id}` | `{removed:true}`; hides the project, preserving its files and saved history; 409 while a session is running |
| POST | `/api/projects/{id}/folders` | `{name}` -> 201 Project |
| POST | `/api/sessions` | `{projectId,title,workspace,harness,model?}` -> 201 Session |
| POST | `/api/sessions/{id}/side` | `{harness?}` -> existing or newly created side Session; one per main, serialized atomically |
| PATCH | `/api/sessions/{id}/harness` | `{harness}` -> side Session; only before its first turn |
| POST | `/api/sessions/{id}/consultations/{requestID}/cancel` | `{}` -> Session; cancels the correlated request/continuation |
| PATCH | `/api/sessions/{id}` | `{title?,workspace?}` -> 200 Session |
| GET | `/api/sessions/{id}/models` | Harness model catalog; `?refresh=true` bypasses the two-minute cache |
| PATCH | `/api/sessions/{id}/settings` | `{model:string,effort:string}` -> Session; empty strings select harness defaults |
| GET | `/api/sessions/{id}/quota` | `{quota:QuotaSnapshot\|null}`; null for unsupported/unverified connections |
| POST | `/api/sessions/{id}/messages` | `{text,sources?:[{sessionId,messageId,passage}]}` -> 202 Session, already containing the durable user event |
| GET | `/api/sessions/{id}/events` | SSE, initial and subsequent full Session JSON snapshots |
| POST | `/api/sessions/{id}/stop` | 200 Session after cancellation completes; idle sessions are unchanged |
| POST | `/api/sessions/{id}/permissions/{permissionID}` | `{decision:"once"\|"session"\|"always"\|"reject"}` -> Session |
| POST | `/api/sessions/{id}/questions/{questionID}/reply` | `{answers:string[][]}` or `{cancelled:true}` -> Session |
| POST | `/api/dialogs/directory` | `{path:string\|null}` |

Adding the same directory again restores a removed project's identity, folders,
and session history with the name and icon chosen in the add dialog.

```typescript
type Project = {
  id: string
  name: string
  folder: string
  icon: string
  folders: string[]
}

type Session = {
  id: string
  parentId?: string
  sources?: { sessionId: string; messageId: string; passage: string }[]
  projectId: string
  title: string
  workspace: string
  harness: 'pi' | 'opencode' | 'codex'
  model?: string
  effort?: string
  resolvedModel?: string
  resolvedEffort?: string
  status: 'idle' | 'running' | 'error'
  events: Event[]
  permissions?: Permission[]
  questions?: QuestionRequest[]
  usage?: {
    inputTokens: number | null
    outputTokens: number | null
    context: { tokens: number | null; window: number | null } | null
  }
  createdAt: string
  updatedAt: string
}

type Event = {
  id: string
  consultationId?: string
  type: 'user' | 'assistant' | 'reasoning' | 'command' | 'mcp' | 'tool' | 'error' | 'status' | 'consultation'
  text: string
  title?: string
  status?: string
  createdAt: string
  data?: Record<string, unknown>
}

type Harness = {
  id: 'pi' | 'opencode' | 'codex'
  name: string
  available: boolean
  error?: string
}
```

### Model selection and account quotas

Catalog discovery uses isolated, prompt-free native control processes scoped to the
project directory: Pi RPC plus its installed model-capability library, OpenCode's
authenticated loopback provider/config API, and Codex app-server model pagination.
Catalog responses expose `connectedProviders` (`id`, `name`, `authType`) as an
explicit boundary. Pi derives this from its credential-filtered available snapshot;
OpenCode intersects native connected IDs with authenticated/environment or explicit
provider configuration, excluding automatic free catalogs. Codex's default OpenAI
catalog requires a native account connection. Both engine and picker filter models
against this list. Model IDs include provider identity
for Pi/OpenCode. Session settings reject unavailable models/efforts and active turns.
No global model or thinking defaults are written. Existing CLI model overrides remain
readable; new UI selections come from the actual catalog.

Quota snapshots contain `source`, `observedAt`, `windows` (`name`, `usedPercent`,
`resetsAt` in epoch seconds), and optional `stale`. They are in-memory only and
credential-fingerprint scoped. Codex reads `account/read` and `account/rateLimits/read`
without starting a thread; ChatGPT subscription auth is required. Cross-harness
OpenAI auth must identify the same Codex account.
OpenAI subscription detection recognizes native Codex protocol metadata as well as
the OpenAI/ChatGPT provider IDs; API-key billing and intermediary routes remain
excluded. Ordinary protocol/client headers and non-auth provider settings do not
disqualify a matching OAuth connection.
Claude Code must be installed with
readable OAuth credentials; direct Anthropic OAuth is matched by token or account
and organization from `/api/oauth/profile`, then reads `/api/oauth/usage`.

Anthropic's OAuth usage endpoint is a best-effort provider integration rather than
a documented standalone Claude CLI command. Missing/expired credentials or unsupported
responses yield no new quota. Credential rotation remains owned by the CLI; KLM never
rewrites auth files, logs tokens, sends a model prompt to obtain quota, or queries a
user-supplied quota host. Credential files currently cover Windows/Linux file-backed
CLI installations; OS keychain-only storage is not read. Pi credential helpers for
native Codex-protocol models can be resolved by `pi auth check --credentials --json
--no-refresh`; the returned credential stays inside the engine and must identify the
same Codex account. API keys, intermediary providers, custom auth/route overrides, and unknown
account relationships are hidden instead of showing unrelated subscription limits.
Quota responses may include a non-secret `reason` code when unavailable, such as
`account_mismatch` or `harness_oauth_unavailable`, for diagnostics.

### Token usage details

`usage` is persisted and published through the existing session snapshots. Input
totals include cached tokens; output totals include reported reasoning. Null means
unavailable, and absent usage on older sessions remains unknown until the next run.
Context is a latest available occupancy estimate, not the sum of session totals.
Counts update on harness usage reports, not every streamed text delta.

- Pi: `get_session_stats` provides authoritative session totals, including retained
  pre-compaction history, and `contextUsage`. Requested at startup, after assistant
  messages/compaction, and at settlement (with a bounded final wait). Missing optional
  telemetry does not fail execution.
- OpenCode: native history plus live message/step updates, deduplicated by IDs.
  Step-finish counts take precedence over message counts to avoid counting the same
  call twice. Input adds cache read/write; output adds the separate reasoning bucket.
  Context uses the latest exchange's input + output and the actual model's
  `/provider` context limit. Model lookup failure leaves capacity unknown.
- Codex: `thread/tokenUsage/updated.total` replaces cumulative counts; `last.totalTokens`
  and `modelContextWindow` supply context. Cached input and reasoning output are
  already included and must not be added again.

Compaction invalidates stale occupancy until a fresh usable context report arrives.
Context snapshots may lag new prompts, tool results, or streamed output. No client-side
tokenizer, hardcoded model capacity, or usage inferred from visible text is used.

IDs are opaque strings, and timestamps
are UTC RFC3339 with fractional seconds. Native harness identifiers and Pi's
session path live in a separate private section of `state.json`.

Validation and semantics:

- Project names and grouping folder names are trimmed, 1-60 Unicode characters,
  with no control characters. Duplicate grouping names are rejected
  case-insensitively; `Ungrouped` is reserved. These folders are labels, not
  filesystem directories. `Project.folders` initially equals `[]`.
- Project `folder` must be an existing absolute directory, checked again before
  every execution. Processes always use that directory as their working directory.
- `icon` may be empty. Otherwise it must be a `data:image/png;base64,` URL, at most
  128 KiB for the entire URL, and a fully decodable PNG no larger than 1024x1024.
  The client is responsible for resizing. The pixel bound prevents compressed
  image allocation bombs.
- Session titles are trimmed, 1-200 characters. Optional model IDs are trimmed,
  at most 200 characters, with no control characters or leading `-`.
- `workspace` names an existing project grouping folder or `Ungrouped`. An empty
  or omitted workspace becomes `Ungrouped`. It never changes the process cwd.
- New sessions are idle with `events: []`. The harness must be discoverable when
  creating a session, but availability does not imply authenticated or usable.
- Messages preserve whitespace, cannot be blank or contain NUL, and are capped
  at 128 KiB UTF-8. JSON request bodies are capped at 512 KiB.
- Only one turn per session can run; another message gets 409 until it exits.
  Different sessions may run independently. Linked-agent consultations queue and
  route through the scheduler described below; ordinary user messages are not queued.
- Event IDs remain stable when a harness item is updated. Array ordering is
  first-observed ordering. Item updates replace that item's data/text in place;
  Pi text deltas append. This is not a duplicate raw-event audit log.
- Event status is optional and describes the item, typically `running`,
  `completed`, `error`, `started`, or `cancelled`. Tool errors do not by themselves
  fail a turn, because a harness may recover. Explicit turn/assistant errors do.
- SSE uses default `data: <Session JSON>\n\n` events, not named events. It sends the
  current snapshot immediately, plus 15-second comment heartbeats. Slow clients
  may skip intermediate snapshots, but the latest snapshot includes every stored
  event. Reconnect simply receives the latest state; no Last-Event-ID is needed.

## Harness Adapters

Discovery runs at engine startup without sending any prompts or reading auth
files. Executable overrides are optional: `KLM_PI_BIN`, `KLM_OPENCODE_BIN`, and
`KLM_CODEX_BIN`. Each is a single executable path/name, not a shell command or
argument string. A Pi override can instead point to `cli.js`, requiring Node.js
on PATH. Invalid overrides make that adapter unavailable; they do not fall back.

On Windows, discovery checks `%APPDATA%\npm\node_modules`, npm package locations
adjacent to discovered PATH shims, native `.exe` files on PATH, and Codex's
`%LOCALAPPDATA%\Programs\OpenAI\Codex\bin\codex.exe`. Pi's installed package is
`@earendil-works/pi-coding-agent/dist/cli.js`; OpenCode's is
`opencode-ai/bin/opencode.exe`. No user-specific absolute installation paths are
hardcoded. `.cmd`, `.bat`, and `.ps1` shims are never executed or parsed.

All prompts are supplied verbatim through stdin and EOF, never shell-interpolated
or placed in command arguments. Model and native session identifiers are separate
argv entries. Children inherit the environment and use their harness's existing
authentication/configuration. The engine neither obtains nor returns credentials.

### Pi 0.80.6

```text
node <pi-cli.js> --mode rpc --no-extensions --extension <owned-extension.ts> --session <absolute-session.jsonl> [--model <model>]
```

One RPC process per turn; subsequent turns use the same session file at
`<data-dir>/sessions/<KLM-session-id>.jsonl`. Before prompting, the engine requires
both the approval extension's readiness notification and a validated get_state
response. The embedded extension gates every agent-dispatched tool_call using a
cancellable RPC UI selection. Unknown interactive dialogs fail closed. Responses
are sent over stdin, not inferred from stdout tool notifications.

`message_update.assistantMessageEvent` text/thinking deltas append to the correct
content-index item. `message_end` replaces each item with final text/thinking and
checks `stopReason`/`errorMessage`, even when the CLI exits zero.
`tool_execution_start`, `tool_execution_update`, and `tool_execution_end` share an
event; accumulated partial results replace rather than append. Bash is a command
event. `agent_settled` marks completion. Duplicate user messages, full `agent_end`
message copies, and turn boundary copies are not added to the UI history.

Pi is **not sandboxed** by this engine. The approval extension gates agent tool
dispatch, not arbitrary extension code or OS calls. Third-party extension discovery
remains disabled so later handlers cannot mutate arguments after approval. This
also disables extension-provided MCP tools. No ungated print-mode fallback exists.

### OpenCode 1.18.29

```text
opencode.exe serve --hostname 127.0.0.1 --port 0 --mdns=false
```

An owned loopback server is started per turn with random HTTP Basic credentials.
The engine uses directory-scoped HTTP/SSE directly, not `run --attach`, which
would still auto-reject prompts. Native session IDs are reused after directory
validation. It subscribes before posting prompt_async, tracks message/part IDs,
streams text/reasoning/tools, and reconciles final history without duplicating old
messages. Idle completes the turn; individual step-finish events do not. Stream
loss fails and aborts the turn rather than silently replaying actions.

Permission asked/replied events create/remove pending cards. Positive choices
always send native `once`; KLM owns remembered scopes. Native `always` is not used
because its in-memory cache can affect other sessions in the same instance.
Existing plugins/config and explicit denials are respected. Native question events
are routed to question cards. Tools already allowed by native rules do not require
a permission card.

### Codex 0.142.5

```text
codex.exe app-server --listen stdio://
```

The engine initializes JSONL RPC, starts/resumes the stored thread with read-only
sandboxing and on-request/user approvals, then starts the turn. Native IDs and
working directories are validated; no replacement thread is silently created.
Text/reasoning/output deltas update stable items and completed snapshots replace
them. Turn completion, not process exit, ends execution.

Command and file-change requests receive accept/decline only. Permission-profile
requests preserve all requested access and deny entries; native grants have a
minimum **turn** lifetime, labeled in the UI. Recognized empty MCP approval forms
support one-shot consent only when resource identity is ambiguous. Unsupported
forms, URLs, legacy/dynamic-tool requests are declined or failed, never
answered affirmatively. No native session or global policy amendments are used.

Native item/tool/requestUserInput requests are handled as questions, not approvals.
The engine enables `features.default_mode_request_user_input=true` and
`tools.experimental_request_user_input.enabled=true` through child-only CLI flags.
This Codex feature is under development in 0.142.5; no global config, collaboration
mode, or model is changed.

## Permission Decisions

Session snapshots include optional pending `permissions` entries with id, harness,
kind, title, description, patterns, details, decisions, allowLabel, createdAt and
resolving. IDs bind to the active engine turn and native request. Client payloads
contain a choice only, never raw native responses. Concurrent/stale decisions are
rejected; stop and restart invalidate pending requests.

- `once`: approve the native request without an engine rule. Codex permission
  profiles last for the current turn, not one operation.
- `session`: persist an exact grant for the logical KLM session.
- `always`: persist an exact grant for this project and harness, across sessions.
- `reject`: deny the native request. The harness may recover or stop.

Remembered grants use a hash of the full semantic action/resource scope plus
project, harness and permission kind. No wildcard expansion or parent-directory
inference is performed. All request patterns/arguments must match. Ambiguous
requests disable remembered choices. Grant persistence succeeds before a request
is released; delivery failures roll back the new grant. Pending requests are not
replayed after restart. A permission audit entry is included in session events.

Grants are stored in `state.json`'s permissionGrants field. They do not modify
global harness settings. There is no grants-management UI in this first slice;
session grants end when that logical session is no longer used, not at every turn.

## Question Tools

Session `questions` entries contain id, harness, items, createdAt, and resolving.
Each item contains id, header, text, options (label/description), multiple, custom,
and secret. Responses are positional `answers:string[][]`, one array per item;
the engine maps them back to each harness's native identifiers and labels.
Requests are tied to the active turn. Unknown choices, empty answers, duplicates,
unsupported multi-selection, stale IDs, and duplicate submissions are rejected.
Dismissal returns no default answer. Pending questions are cleared on stop/restart.
Answers do not affect permission grants. The answer audit is recorded in history;
private fields are masked there but sent to the receiving harness. Its own records
or model output may retain/echo values; this is not end-to-end secret redaction.

- OpenCode: handles question.asked/replied/rejected, GET /question reconciliation,
  and ordered native reply/reject endpoints. Custom entry defaults to allowed;
  multiple comes from the request. Native question registration is enabled on the
  owned server. Old CLI-run sessions may still deny question; use a new session
  rather than overriding intentional native permission rules.
- Pi: the owned extension registers sequential `ask_user` with structured questions,
  options, custom text, and multiple selection. Only this tool bypasses the redundant
  tool-permission gate. A versioned `klm.question.v1:` RPC input envelope carries
  the form, and JSON answer arrays return to the actual tool result. Other tools
  remain gated. Cancelled questions are explicitly reported as dismissed.
- Codex: native questions preserve IDs, offered choices, isOther/isSecret, native
  RPC IDs and resolution/expiry notifications. Replies map each question ID to
  `{answers:[...]}`. No native multiple flag is inferred from array-shaped answers.
  Malformed requests return empty answers and a warning, not fabricated consent.

## Linked-Agent Conversations

`Session.parentId` attaches exactly one side conversation to a main session in the
same project. Creation is idempotent under the engine mutex, including concurrent
clients. `/api/state` includes both; the client filters children out of its top-level
list. Main and side have independent native IDs/files and harness/model settings.
Switching a side harness is allowed only before any events or active execution.

Side messages accept up to eight source references from the parent conversation,
with at most 128 KiB of selected text total. Passages are never truncated. The engine
checks source IDs and attaches deterministic JSON reference context: project/main
identity, exact passages, each source and up to two neighboring events on each side
(2 KiB per surrounding event). A selection-free first turn includes the latest six
events with the same bound. The user question stays separate from reference material.
Original selected text and source references persist with the user event and session.

Shared tools, available to all three harnesses:

- `linked_discover`: caller/linked identity, role, project, harness and execution status.
- `linked_read`: only the linked conversation, latest ten events by default, maximum
  twenty per call; `cursor`/`nextCursor` page history, `messageId` targets a source,
  `offset`/`nextOffset` page long text in 4,096-character chunks. No native paths or
  arbitrary session targets are exposed.
- `linked_ask`: explicit short `topic` (120 characters) and `question` (32 KiB).
- `linked_answer`: `requestId` and `answer` (64 KiB), accepted only in the recipient's
  actual active consultation turn, then the agent is instructed to finish that turn.

Each execution owns a random-capability, authenticated loopback Streamable HTTP MCP
server with four tools. Access is bound to that turn and its linked counterpart;
capabilities are not placed in engine state and expire when the process finishes.
OpenCode dynamically registers `klm_linked` on its owned server and confirms connection
and tool-list readiness. Codex receives child-only `mcp_servers.klm_linked` overrides
and confirms the tool inventory through `mcpServerStatus/list` before starting a turn.
Pi's bundled temporary extension exposes the same operations and verifies the inventory
before its existing readiness notification. Linked communication tools bypass Pi's
redundant per-tool permission dialog. Codex automatically accepts native tool-call
approval forms for the ready, turn-owned `klm_linked` bridge; OpenCode automatically
replies `once` to permission requests matching its four exact internal tool names.
These internal operations never create permission cards or remembered grants. The
policy does not apply to other MCP servers or arbitrary elicitation forms, and
filesystem/shell tools remain gated. No user/global
installation or native transcript-format translation is required.

The installed versions inspected for this feature were OpenCode 1.18.30 and Codex
0.153.4; Codex's generated local app-server schema confirms the status-list contract.
Runtime readiness checks fail explicitly if an installed harness cannot expose the tools.

Requests persist in `state.json.consultations`, with question, answer, topic, endpoint
IDs, timestamps, status (`queued`, `answering`, `completed`, `failed`, `cancelled`,
`interrupted`), and delivery state. Both sessions receive atomically updated activity
events. Background output and permission/question answer audits carry `consultationId`
so the client groups the recipient's background answering work under the relevant
activity. The requester's resumed response is shown as ordinary chat messages,
including for previously saved consultations; its correlation ID is retained.
Native consultation prompts and outputs still enter the recipient's native history.

Idle recipients start immediately; busy recipients wait for the current turn to finish.
Tool requests wait at most five seconds, safely below HTTP timeouts. Reciprocal requests
yield immediately. If no terminal result arrives within that wait, the tool instructs
the requester to finish its turn; the engine starts a correlated continuation when
it is idle. No same-session parallel turn or polling loop is needed. At most four
questions may be outstanding per requester; identical pending questions deduplicate.
Tool results explicitly carry `action: "continue"` with terminal answers/failures
or `action: "yield"` while still pending. A reply present when the five-second wait
expires takes precedence over yielding and includes the full answer. Resumed prompts
explicitly replace the earlier yield instruction and ask the requester to use the
answer without another user follow-up.
Unanswered requests expire after ten minutes, including queue time. Stop cancels linked
requests involving that conversation; per-activity cancellation targets one request.
Failed/stopped requester turns suppress unattended continuations. Shutdown and restart
retain terminal interruption state rather than replaying background work; an answer
already persisted before restart remains readable even if its continuation was interrupted.

Manual validation: open from header and selection; reopen history; add later passages;
resize with pointer/keyboard and narrow layouts; test different-harness pairs in both
directions; queue against busy recipients; cancel, stop, and restart during consultation.
Only lightweight Go/desktop builds are part of implementation validation, not live model
execution, a full suite, or an adversarial review.

## Persistence And Recovery

- `state.json` holds projects, sessions, events, and the private native-session
  map. Every published mutation is serialized under a mutex and written to a
  same-directory temporary file, synced, closed, then atomically replaced.
- Windows replacement uses `MoveFileExW(REPLACE_EXISTING | WRITE_THROUGH)`.
  Unix uses rename plus parent-directory sync. Use a local filesystem with normal
  atomic rename/sync semantics, not a network share or cloud-synced data directory.
- In-memory state changes and SSE notifications happen only after persistence
  succeeds. A write failure preserves the last published snapshot, cancels active
  executions, blocks further mutations, and makes health return 503. Fix the
  underlying storage problem and restart; no automatic overwrite/retry is done.
- `engine.lock` is held exclusively by an OS file lock/handle for the engine's
  lifetime. It contains a diagnostic PID, but ownership is not inferred from the
  PID. Kernel lock release on process exit handles stale files safely. Do not
  delete the lock file while an engine is running.
- Startup refuses malformed/unsupported state instead of resetting it. Sessions
  left running become `error` with a durable interruption event. Native harness
  transcripts remain owned/formatted by the harness itself; they are not part of
  the engine JSON transaction, and a stopped turn may have partial native history.
- Stop/shutdown cancel the direct child and use Windows `taskkill /T /F` on that
  owned PID only. Unix children use a separate process group. The stop response
  waits for termination and the durable cancelled snapshot; there is no replay of
  the cancelled prompt. Force-killing the engine itself cannot run cleanup and
  may leave external children, which are not blindly killed by PID on restart.
- Output is consumed concurrently from stdout and stderr. JSON lines are capped
  at 2 MiB and total stdout at 32 MiB per turn; hitting a limit cancels the turn
  with an explicit error. Stderr is capped at 16 KiB and is never persisted,
  exposed, or logged. Exit failures report a bounded generic diagnostic instead.
- Session history is durable and unbounded across turns. This first implementation
  rewrites the full state per event, intentionally favoring simplicity over large
  histories/high token throughput. There is no retention, deletion, or compaction
  API yet. Back up the data directory with the engine stopped.

## Local Security Boundary

The listener, client address, Host header, and browser Origin must be loopback.
Allowed browser origins are HTTP(S) on `localhost` or loopback IPs with any port,
so a Vite client can use the API directly. Non-loopback and `null` origins are
rejected. Requests without Origin are accepted for local native clients. CORS
does not permit credentials. JSON-only mutations block ordinary cross-origin
HTML form submissions; there is no externally reachable listener.

This is a trusted-local-user API, not an authentication boundary between local
applications. Any local process or accepted loopback web application can use it.
Do not proxy it to a network or run untrusted web applications on loopback.

The engine never reads auth stores, prints environment values, or exposes binary
paths/command lines through discovery. Harness-generated tool/message content
can still contain sensitive project information and is shown/stored as content;
this is not a general-purpose content redaction system. Protect the data directory
with user-only OS permissions. State is plaintext, not encrypted.

The native directory picker is Windows-only. It runs a fixed PowerShell
`-NoProfile -NonInteractive -STA` WinForms `FolderBrowserDialog` script with UTF-8
stdout and no user interpolation. Dialogs are serialized, have a five-minute
timeout including queue time, and return only an existing absolute path or null
on dismissal. Request cancellation/shutdown kills the owned picker process tree.
Other platforms return 501 for this endpoint; Unix support has not been validated.

## Validation Scope

Installed Pi source and harness CLI help were inspected without sending prompts.
The implementation is intended for a lightweight Go build followed by manual
HTTP, picker, SSE, stop/restart, and authenticated chat validation. No tests or
test suites are included or run as part of this slice.
