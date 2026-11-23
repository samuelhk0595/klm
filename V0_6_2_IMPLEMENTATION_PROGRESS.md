# 0.6.2 — Implementation progress

Date: 2026-11-01. Reference: `V0_6_2_IMPLEMENTATION_PLAN.md` (preserved unchanged).
Source implementation A–H delivered in the shared checkout; **functional human
acceptance remains pending**. Implementation was performed on Windows; those
cross-compilation checks are not Linux runtime evidence. The initial implementation
did not deploy or modify real data. A separately requested operational deployment
subsequently published the Linux engine at 20:24 UTC; the frontend remains pending.
See `V0_6_2_VPS_DEPLOYMENT.md` for artifact hashes, smoke evidence and rollback.
No graph execution, independent session/delegation, commit, push or GitHub release.

## A — OpenCode diagnostics and gh

- Inspection uses the harness-resolved Node (common resolver when absent), a
  read-only SQLite connection and structured module/open/schema/query categories.
  An absent database remains its own pre-inspection case; unsupported organization,
  well-known remote configuration and administrator policy remain refused.
- Timeout and caller cancellation are distinct. Unexpected process failures report
  exit status and diagnostic ID, without claiming outdated Node. Logs contain
  executable/version when available, duration/category/exit and bounded stderr byte
  count, never native stderr, environment, database rows or auth contents.
- Conversation, project/global model catalogs and specialist preparation retain
  safe inspection details through a typed error; unrelated native errors remain
  behind the normal catalog message.
- Harness and catalog environments pin absolute `GH_CONFIG_DIR` from the original
  engine environment: explicit GH, then XDG/gh, then the native per-user default.
  Tokens/hosts retain native precedence. No credential copies, HOME changes or
  modifications to independently configured MCP server environments.
- Main files: `opencode_library.go`, `models.go`, `harness_discovery.go`,
  `gh_environment.go`, `opencode_interactive.go`, agent catalog/resolution handlers.

## B — Ordinary automation chats and ownership

- Removed composer steering coercion, automation-only Send now hiding, graph,
  Library, agent/settings and live-YOLO edit locks from the shared client/SQL APIs.
- Queue dispatch uses actual in-memory turn ownership, not stale automation/session
  `running` metadata. Normal FIFO, queue limits, message identities, uncertainty,
  Stop and replacement-shutdown barriers remain in place. Agent instructions keep
  ordinary FIFO rather than being converted to steering for automation targets.
- New input dispatch transactionally finalizes a still-owning automation as
  `interrupted_by_user`, preserving history. Enqueueing/steering alone does not.
  Completed runs keep their result. Normal completion no longer pauses untouched
  queued input just because it belonged to an automation conversation.
- Unstarted reminders and graph-result continuations defer to ready queue input.
  Automatic turns carry an explicit immutable run ID. Notifications retain the
  originating activity/run and only recover automation ownership if that exact
  original run is still active; unrelated/late turns receive neither its outcome
  tools nor its captured selection/grant.
- OpenCode/Codex automation settlement retires the owned runtime even after a
  successful native turn. Pi automation settlement does the same. Drainage is
  derived from actual shutdown, never inferred from an idle retained server.
- Model/Library captures stay immutable. Explicit human live YOLO changes are
  recorded separately from the admitted capture and keep Codex's downgrade gate.
- Main files: `sqlite_execution_{queue,finish,stop,settings,context}.go`,
  `sqlite_automations_worker.go`, `sqlite_graph_notifications.go`, harness adapters,
  `App.tsx`, Focus chat, queue and selection/settings dialogs.

## C — Automation configuration graph authorization

- New runs capture `graphGrant` with project, automation, run, configuration
  revision/digest and selected graphs. No retrospective grants for incomplete or
  older runs. Explicit recoveries capture their saved configuration at admission.
- An exact owning turn may omit message references for a selected graph. Only the
  engine binds the `automation_configuration` source; strict tool arguments reject
  model-supplied automation/source fields. Normal human references remain valid.
- SQL activity authorization records retain the discriminated provenance. Scheduler
  and startup validation verify it against the original persisted run, independently
  of later chat ownership or live configuration. Legacy `{eventId,text}` remains
  readable. Activity notification origin is not itself an authorization grant.
- Retry correction/assessment and explicit fresh/reuse checks remain unchanged.
  Authorization never broadens sessions, Library, skills, tool grants or questions.
- Persistence uses the existing extensible SQL record/authorization-child payloads;
  no schema/transport-contract renumbering or invented human events is required.
- Main files: `automation_graph_authorization.go`, `graph_records.go`,
  `graph_invocation_contract.go`, SQL graph tools/scheduler/validators, automation,
  orchestration and collaboration prompts.

## D–E — Linux graph admission and Terminal shells

- Linux amd64 capability discovery performs a real trusted-helper handshake,
  including subreaper and pidfd availability; activation admission still creates
  a separate process domain. `runtimeValidated` remains false.
- Existing agent/Join, owned children, Terminal and Git/workspace executors use
  supervised ownership. Fork/Join and native continuation retain their existing
  routing and finality checks. Pi/OpenCode keep pre-tool gates; Codex acceptance
  requires confirmed complete domain termination. Supervisor loss, timeout and
  uncertain MCP responses remain failure/uncertainty, not acceptance.
- Terminal `shell` persists through YAML/JSON/canvas/snapshots. Absent legacy fields
  always resolve to PowerShell. New nodes explicitly save Windows PowerShell or
  Linux sh based on an engine-reported default. Bash is selectable. Older-editor
  saves preserve an existing shell instead of silently dropping it.
- Missing executables fail preflight with node/shell/recovery. Scripts execute as
  private files with inherited cwd. PowerShell retains `$payload` and encoded
  constant-launcher exit semantics; Bash/sh receive the JSON path without eval,
  interpolation, mandatory jq/Python or added `set -e`. Startup ENV/BASH_ENV hooks
  are removed. Output/exit/truncation handling stays on the existing owned executor.
- Main files: `graph_platform_{linux,other}.go`, `graph_adapter.go`,
  `platform_linux.go`, `graph_terminal.go`, `terminal_shell.go`, authoring/schedulers,
  frontend graph file mapping, canvas and Terminal/configuration panels.

## F — Headless Linux MCP credentials

- AES-256-GCM versioned envelope with fresh nonce and MCP-domain associated data.
  The 32-byte random key lives outside SQLite in
  `<engine-data-dir>/mcp-secrets/credentials.key` (directory 0700, file 0600).
- Descriptor-relative, no-follow opens validate owner/modes/regular file/link count.
  Key publication uses exclusive atomic rename and directory sync. Existing keys
  are not overwritten or silently permission-repaired. Missing dependent keys,
  incompatible format and integrity failures remain explicit recoverable errors.
- Manual/header/bearer secrets, stdio secret environment and OAuth initial/refresh
  use the same storage-context-aware path. Windows DPAPI and existing blobs remain
  unchanged. Key material never enters harness environment/arguments or snapshots.
- Backup/restore requires the consistent database plus matching private key. Ordinary
  exports exclude the key. Windows blobs are not portable to Linux. Protection is
  at rest, not isolation from another process running as the engine user.
- Main files: `mcp_secrets_linux.go`, platform secret backends, connection/OAuth paths.

## G — Native Pi and optional graphical picker

- Kept native `pi-subagents` 0.70.0 distinct from engine-owned `subagent_run`.
  Removed Linux refusal only with supervisor admission; SDK foreground children and
  local background runners stay in the parent's owned domain. Parent settlement
  retires that domain after native child liveness/result handling, including
  cancellation and detached local descendants. No simultaneous activation of both
  delegation paths or changes to inherited model/effort/policy.
- Factory path comparisons respect Linux case sensitivity. Leaf capability and
  administrative-tool exclusions remain. Missing optional package does not block Pi.
- Linux picker availability is dynamic from graphical environment plus installed
  Zenity/KDialog. Directory-only argv, cancellation, timeout/process cleanup and
  absolute-existing-path validation reuse the current endpoint. VPS headless hosts
  retain remote browsing and advertise no native picker. Nothing is installed.
- Main files: `pi_subagents.go`, `pi_interactive.go`, `pi-subagent-factory.mjs`,
  platform picker and filesystem capability handlers.

## H — Documentation and version preparation

- Updated `product.md`, root/engine/Linux READMEs, project guidance and prompts.
- Desktop npm/lockfile, Tauri/Cargo/lockfile and Windows engine build/NSIS defaults
  now identify 0.6.2. No installers/releases were generated. Native mobile code was
  not changed or distributed; its own release version/build remains a release-time
  decision. Storage/transport versions were not renamed after the release number.
- Excluded Linux Activity Board credentials, desktop, installers, engine UI hosting,
  ARM64 and Alpine/musl. Existing Windows packaging and data remain intact.

## Lightweight checks actually performed

| Check | Result / limit |
| --- | --- |
| Windows `go build -o <approved-temp>/klm-062-windows.exe .` | Passed; initial attempt found one unused scheduler import, fixed before successful build. |
| Linux `GOOS=linux GOARCH=amd64 go build -o <approved-temp>/klm-062-linux .` | Passed cross-compilation, including Linux crypto/supervisor/picker code. No Linux execution. |
| `go test -run '^TestV062' -count=1 .` | Passed on Windows; small pure contract tests cover cross-scope/config grant rejection, forged wire provenance, legacy human references, legacy shell/YAML round trip and gh directory/token/host precedence. No graph/harness dispatch. |
| Desktop `npx tsc --noEmit` | Passed after frontend changes. |
| `node --check pi-subagent-factory.mjs` | Passed syntax check; no package/model execution. |
| `gofmt` / `git diff --check` | Formatting applied; diff whitespace check passed. |

During implementation, no full suites, integration tests, review agents, real
providers, native Tauri or Flutter builds, installers, Linux graphical probes or
VPS operations were run. The later requested deployment built release Linux
amd64 with `-trimpath -ldflags=-X main.engineVersion=0.6.2`, built Vite assets for
the VPS API, published binary/prompts/skills through the existing unit, and checked
only operational startup/integrity/availability. It did not dispatch graph/model
test tasks. Normal engine restart recovery ran; no manual state reset was applied.

## Pending human acceptance

1. Reproduce the actual VPS OpenCode issue; exercise module/schema/query/timeout
   failures and cancellation in both model catalogs and chat startup.
2. Validate automation idle/active/failed/successful input, steering, FIFO, Send now,
   Stop, stale-run recovery and classification/reminder races with every harness.
   Confirm no lost/duplicate input, concurrent turns or inherited grant/classifier.
3. Compare manual/automation gh account discovery with explicit GH/XDG settings and
   after restart; do not expose tokens in diagnostics.
4. Execute selected automation graphs without another confirmation; reject other
   graphs/projects/turns. Check saved provenance, restart/idempotency, late result
   attribution and explicit retry fresh/reuse behavior.
5. On Linux amd64, validate Pi/OpenCode/Codex agent chains, Terminal, terminal Choice,
   Fork/Join, original/worktree workspaces and native continuation. Check supervisor
   loss, shutdown uncertainty and descendant cleanup without touching other chats.
6. Exercise Bash/sh with spaces/Unicode/metacharacter payloads, output mapping,
   nonzero exit and cancellation; verify existing PowerShell graphs on Windows.
7. Linux manual and OAuth MCP save/test/use/refresh after restart; inspect synthetic
   storage for plaintext absence, missing/wrong-key and owner/mode/link failures.
   Confirm Windows continues reading existing DPAPI blobs.
8. In the native Pi path, complete one child and cancel another (foreground/background)
   without orphans/admin tools. Check absent package and independent managed path.
9. Graphical Linux Zenity/KDialog select/cancel with Unicode/spaces; headless remote
   browser flow. Windows native picker, composer and affected session controls.

Version acceptance requires the focused real Linux and affected Windows scenarios;
successful builds and pure contract tests are not a substitute for those checks.
