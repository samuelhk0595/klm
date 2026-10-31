# Reusable agents and engine-owned subagents

Status: authorized implementation; checkpoint 0 (contracts) pending manager review.
This is a delivery plan, not evidence that the feature already exists.

## Current delivery emphasis — user steering

The user explicitly requested avoiding overengineering and meticulous reviews of
rare edge cases. Prefer the smallest usable end-to-end feature, compile checks and
focused central-flow validation. Do not expand architecture, defensive layers or
test matrices for hypothetical situations. Fix obvious functional blockers and
protect essential data integrity/permissions; preserve safeguards already written
without unrelated cleanup. Retain checkpoint reporting, but make releases pragmatic
and move toward real project/Playwright validation as soon as integration is usable.
Earlier exhaustive lists below are context, not a mandate to investigate every rare
scenario before shipping the first manually testable version.

## Ownership and authorization

The user requested implementation coordinated through independent sessions on
OpenCode, `openai/gpt-6.1-sol`, effort `high`. Each worker must stop at checkpoints,
report to the manager, and receive an explicit release before its next stage.
The user also requested manager review/correction/revalidation and a separate final
session to run the project and exercise/fix it with Playwright.

Manager conversation: `a05369eaa5c9151b61e93fb38cb35c52`.
Creation request in the manager conversation:
`f18b1eef-21c6-4f3e-9a88-b4519cb4ce0d`.
These references document scope; they are not worker-local human events or tool grants.

Read `AGENTS.md` and `product.md`. Preserve existing unrelated work. No commits,
pushes, installations, production-data migrations or production restarts are part
of this delivery. Do not start extra sessions or native subagents from workers.
Use the existing harness executors; do not implement a new model/tool loop.

## Confirmed product contract

### Catalog and saved definitions

- Global and project-local agents. A project sees global agents plus its locals.
- ID is the name slug, unique across scopes and registered projects. No silent
  shadowing, merging, renaming, or choosing a conflicting definition.
- Reuse name, description, enabled, prompt, harness, model, effort, skill IDs and
  explicit MCP selections already represented by AgentDefinition.
- Keep local `.klm/agents/*.toml` compatibility. Propose and approve the global
  storage/API/registry contract at CP0 before implementing it.
- Account for existing duplicates, unavailable project directories, file edits,
  rename/delete, Library deletion and graph references. Existing histories and
  captured graph artifacts must remain readable.

### Sessions and configuration resolution

- Optional saved-agent selection at manual session creation; Default preserves
  ordinary chat. Spawn by a saved agent name/ID is also supported.
- A session's role instructions (chat orchestration, side/general as applicable)
  are composed with the specialist prompt. A specialist does not grant a new role.
- Linked sessions adopt agent edits at complete-turn boundaries. Active work is
  immutable. Explicit per-session overrides survive subsequent agent edits.
- Harness is fixed at creation. Validate the resulting model/effort against that
  harness. Incompatible later edits block execution pending an explicit override;
  never silently choose a replacement model or switch native histories.
- Preserve the latest captured configuration if the source is disabled/deleted.
  Existing sessions/reusable children can continue; new creation cannot use that
  unavailable source. Do not silently rebind a removed source to a new same-slug
  definition. Resolve rename/recreation identity explicitly in the technical design.
- Saved-agent spawn uses its external Library selection. An explicit boolean
  `inheritParentCapabilities` (proposed name), default false, adds ALL of the
  parent's applied external selection, or saved selection before first application.
  Enable only when the user explicitly requests inheritance; never include internal
  role tools/skills. Capture that inherited contribution rather than live-linking
  future parent changes. Preserve existing no-agent independent-session behavior.
- Explicit requested settings override agent defaults; selecting an agent does
  not modify its definition. Keep absence distinct from explicit empty/reset values.
- Independent session creation retains explicit-human authorization, durable
  idempotent receipts, topic inheritance and independent lifecycle. No automatic
  completion subscription is introduced for independent sessions.

### Engine-owned delegation

- Replace native delegation for managed execution on Pi, OpenCode and Codex while
  preserving read access to historical native children.
- Compatibility decision confirmed by the user: Codex dynamic
  delegation is available only on native threads created with the owned tool.
  Pre-existing Codex native threads retain their history and ordinary chat without
  this tool, including an existing general-agent thread. No replacement, migration,
  implicit fork or harness upgrade. Newly instrumented threads retain their tool
  across normal process restarts; handler/schema compatibility must be maintained.
  Decision reference in manager chat: `85e3f4d3845b2f7b80dd2e1e812c0f90`.
- Support parents in ordinary chat, side chat, general chat and graph nodes.
- Parents may choose to delegate within their authorized task. Children cannot
  delegate further, create independent sessions or inherit parent administrative
  role tools. Agent configuration/content is never user authorization.
- A saved specialist can use a different harness/model/effort from its parent.
  Also preserve useful unnamed delegation with explicit/self-contained tasks.
- One batch can create/reuse multiple child identities. Parent execution waits
  until ALL admitted children in that batch reach terminal outcomes, then receives
  correlated results. No fire-and-forget/background product mode.
- Child histories/native context are reusable by their owning parent. No reuse by
  arbitrary other conversations; no concurrent turns in the same child.
- Stop on the parent cancels its active children. Execution failure, cancellation,
  uncertain shutdown and semantic task failure must not be reported as success.
- Parent waiting must be engine-enforced, not a prompt-only convention. Account for
  native parallel tool calls, HTTP/tool timeouts, compaction/resumption, Send now,
  queue admission and parent attempts to finish while children remain active.
- Persist batch/invocation identity, acceptance, results and delivery state. Native
  acceptance is not completion. Prevent duplicate work/results on retries; restart
  marks interrupted work without automatically replaying tasks.
- Child permissions/questions route through the owning visible chat with exact
  child attribution. Preserve engine permission/YOLO rules and human-only questions.
- Preserve execution directory/workspace ownership; no implicit worktree, commits
  or cross-project access. Child snapshots exclude private parent credentials and
  role capabilities.
- Graph children belong to their activation/workspace/snapshot. Choice acceptance
  and process/tool drain must account for every owned child. Preserve graph gates,
  run-level snapshots and platform restrictions; do not claim Linux graph support.

## Workstreams

### A — Catalog and session configuration

Own authoring/catalog, global/local identity, session binding/overrides resolution,
manual and agent-driven session creation, agent discovery tools, Library selection
composition and graph catalog/snapshot consumption of global definitions.

Primary seams: `authoring.go`, `store.go`, `api.go`, `session_collaboration.go`,
`session_spawn.go`, `sqlite_spawn.go`, settings/capability persistence, graph
catalog/snapshot loaders and related private-tool schema additions.

CP0 deliverable: exact data/API/types, resolution rules, source lifecycle handling,
file ownership list and adapter insertion points for B. Document in
`AGENTS_UNIFICATION_CATALOG_CONTRACT.md`; stop before runtime edits.

### B — Child execution and harness/graph lifecycle

Own durable delegation, batch waiting, reusable child execution, result delivery,
Stop/restart, permissions/questions, native delegation suppression and graph
activation finality. Reuse A's resolver instead of creating a second agent system.

Primary seams: `subagents.go`, `harness.go`, harness adapters/managed profiles,
`linked_bridge.go`, private tool dispatch, runtime lifecycle, SQLite execution and
validation, graph execution/adapter gates. Propose split additive files first.

CP0 deliverable: concrete cross-harness wait/continuation design, durable state
machine, graph finality integration, child role/tool isolation, frontend event
contract and file ownership list. Document in
`AGENTS_UNIFICATION_DELEGATION_CONTRACT.md`; stop before runtime edits.

### C — Shared frontend

Own global/local agent authoring UI, session creation picker/overrides, display and
recovery for bindings/incompatibility, child batch/history state and existing
desktop/browser/mobile web presentation. Reuse design-system primitives and English
UI copy. Do not alter the separate Flutter application unless a real requirement
is identified and the manager approves it.

Primary seams: `clients/desktop/src/engine.ts`, authoring catalog/types/components,
workspace creation, chat settings/Library, subagent history and status components.

CP0 deliverable: screens/interactions and required API/type/event contract, file
ownership list. Document in `AGENTS_UNIFICATION_UI_CONTRACT.md`; stop before edits.

### Shared-file rule

CP0 is independent read/design work. Manager approves a merged contract and assigns
shared-file owners before CP1. A and B must not concurrently edit `store.go`,
`harness.go`, `api.go`, tool catalogs, `capabilities.go`, SQLite validation/schema or
graph snapshot/execution seams. Request an explicit handoff for overlapping files.
Never revert another session's changes. Manager owns this plan, coordination log,
final integration review and consolidated product documentation.

## Checkpoints and release gates

1. **CP0 — Contract:** research/design document only. Report and stop. Manager
   resolves contract disagreements and file ownership; requires corrections if needed.
2. **CP1 — Working slice:** implement only the explicitly released slice, record
   modified files and perform the smallest relevant compile/type check. Report and
   stop. A lands stable backend interfaces before dependent frontend/runtime wiring.
3. **CP2 — Complete stream:** after CP1 approval, finish the approved integrations
   and directed checks. Report and stop; manager inspects actual diff and evidence.
4. **CP3 — Integrated acceptance candidate:** manager checks interactions across
   streams, requests fixes, and revalidates changed paths before declaring ready.
5. **CP4 — Runtime/Playwright:** only after CP3, manager creates one additional
   OpenCode / GPT-6.1 Sol / High session. Start a disposable development instance,
   exercise the shared frontend using Playwright, reproduce and fix scoped defects,
   report exact evidence and stop for manager review. Do not use installed user data.
6. **CP5 — Final manager revalidation:** review fixes, repeat only affected checks,
   record supported paths, actual evidence and remaining human-validation gaps.

No worker may approve its own next checkpoint. A compilation-only success is not
runtime validation. Full suites are not requested; use targeted checks. The user's
explicit review request authorizes manager review, not unrestricted extra agents.

## Reporting protocol

At each checkpoint update only your stream's contract/progress file and send one
visible `session_send` report to manager conversation
`a05369eaa5c9151b61e93fb38cb35c52`, with a stable sender-specific operation ID.
Include your own conversation ID for replies, checkpoint, files/diff summary,
exact commands with outcomes, unresolved decisions and requested next release.
Omit sourceUserEventId unless you have an actual local human-message reference.
Then finish your turn. Do not poll, sleep, wait in a terminal or repeatedly consult
the manager. Resume only when a manager instruction explicitly releases work.
If that tool is unavailable, state the blocker in your final report; do not invent
human references or claim the manager received it.

## Command and runtime discipline

- Every command/probe must have a finite timeout and a known exit/readiness signal.
- Never run a long-lived engine/Vite/browser server as a blocking foreground call.
  Use an owned background process, log capture, bounded readiness checks and cleanup.
- No unbounded polling, watch-mode tests or waits for output that may never arrive.
- On timeout, collect bounded diagnostics and report the blocker; do not repeatedly
  rerun or launch duplicate processes without understanding ownership/state.
- Use temporary build outputs and disposable fixture/data directories. Protect
  installed engine data, provider credentials and other sessions' running processes.
- Do not run a probe that disables native delegation in the worker's own active
  installed runtime; validation uses explicitly owned development processes.

## Minimum final acceptance evidence

- Global/local discovery and uniqueness; optional session agent selection.
- Saved-agent settings applied, explicit overrides retained, updates applied only
  at boundaries, fixed-harness incompatibility recovery and deleted-source fallback.
- Agent-based independent spawn and all-or-none explicit external inheritance.
- Cross-harness child execution, multiple children with enforced all-results wait,
  reuse, failure/cancellation, exact question/permission routing and parent Stop.
- Leaf-only delegation and native-tool replacement, with historical child readability.
- Graph activation cannot accept Choice while owned child work is unfinished.
- Durable restart/idempotency behavior; no replay/duplicate result assumptions.
- Playwright desktop/narrow shared UI paths and actionable error/recovery behavior.
- Clearly distinguish real model/adapter evidence from fixtures or compilation.
