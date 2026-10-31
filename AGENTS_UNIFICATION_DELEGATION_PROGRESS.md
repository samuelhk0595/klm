# Stream B — delegation CP1a feasibility evidence

Status: **CP1b candidate corrected for manager finding B-R1 and hermetically
revalidated; awaiting manager review, not approved/activated.** The last recorded
package compilation was blocked during concurrent A work; no new compilation is
attempted until the manager releases the stable shared seam.
User approved new-native-thread-only Codex compatibility after CP1a.
Feature is not activated or wired into product execution. Worker:
`e80963a1b2bdb9c249b765921feb7d28`; manager:
`a05369eaa5c9151b61e93fb38cb35c52`.

| Axis | Current evidence |
| --- | --- |
| Codex thread born with dynamic tool | Local fixture passed: 155s barrier, single result/continuation, persisted schema after process restart, pending cancellation/no automatic replay. |
| Codex thread born without tool | Still blocked: no supported adoption/config update seam identified on exact-pin source; no history replacement attempted. |
| Pi 0.85.1 | Local sequential tool fixture passed pending/readers/continuation/cancellation; 2s pending observation. |
| OpenCode 1.18.35 | Local plugin fixture passed after per-user native-task suppression correction; 2s pending observation. |
| Product integration / real model | Not implemented or validated in CP1a. |

## B-R1 — persistent fatal-stop latch correction

Manager found a local fail-closed defect in the additive transport: a fatal wait/
uncertain post-admission failure rejected its Promise, then `finally` cleared
`active`. While native OpenCode abort HTTP was still pending, `isWaiting()` became
false and continuation/common tool gates could reopen. No product wiring was
active, but this was a concrete defect in the proposed local safety mechanism.

Reproduced the manager's original hermetic regression BEFORE changes:

```powershell
node "engine/probes/agent-delegation/manager-stop-latch.mjs"
```

10-second shell bound, exit1:
`Missing expected rejection: Fatal delegation must remain locally gated until owning runtime stops.`

Correction, limited to `delegation-protocol.mjs`, `delegation-pi.mjs`,
`delegation-opencode.mjs` and the explicitly released manager probe:

- Transport now has an execution-lifetime **fatalStop latch**, independent of the
  active Promise. It is armed synchronously before awaiting best-effort cancel;
  normal requests are aborted through an internal AbortController. `finally`
  releases call bookkeeping, never the fatal latch. There is no reset API: the
  owner must actually end/dispose this execution and its transport.
- Initial admission with uncertain acceptance/delivery and post-admission failures
  latch closed even when no batch ID was returned. Unknown IDs are not invented;
  owner Stop/drain remains authoritative. Known-batch cancellation stays best-effort,
  <=1s and shared once across the original call and its aliases.
- `gate`/`execute` check the latch before work and after awaited RPC boundaries.
  Late valid-looking results cannot release a call after another alias has failed.
- New semantic APIs: `isStopRequired()`, `canContinue()` and
  `assertCanContinue()`. `isWaiting()` remains bookkeeping, not a sufficient
  continuation predicate. Alias reference counting keeps the local barrier closed
  even if the original result settles while an alias response remains in flight.
- Pi checks `assertCanContinue()` before agent start AND at the pinned
  `before_provider_request` hook; failure invokes native abort. OpenCode's
  messages-transform hook checks the same semantic guard. Native abort HTTP stays
  non-awaited to avoid the circular tool/abort wait; its completion is not trusted
  as the local latch's lifetime.
- Invalid input/local busy and proven `admit` rejection with `accepted=false` stay
  recoverable and do not cancel/poison another valid batch. The accepted=false marker
  is recognized only on the admission RPC, not on a later wait/other RPC. An alias
  whose response/cancellation is uncertain follows the fatal path without creating
  another execution or returning a late success.

Preserved the manager's original first regression/expectation in
`manager-stop-latch.mjs`. Extended that SAME focal fixture with:

1. Fatal stop keeps tool, new-execute and semantic continuation guards closed.
2. Invalid input and a proven pre-admission rejection recover normally.
3. Invalid/independent requests and rejected aliases do not poison a valid pending
   original batch or send inappropriate cancellation.
4. Lost/cancelled alias replies close the latch, cancel once, refuse late original
   success and keep only one idempotent execution in the mock backend.
5. Lost initial admission response stops locally without invented cancel identity
   or automatic re-admission.
6. Wait transport errors and wait errors carrying a misleading accepted=false
   marker remain fatal after admission.
7. Pending alias continuation remains blocked even after the original result returns.
8. Mock Pi before-model hooks and mock OpenCode continuation/common-tool hooks stay
   closed while native abort is deliberately still unresolved.

The same command then exited0:

```text
Fatal delegation retains its stop gate.
B-R1 focal cases passed: persistent fatal/continuation latch, recoverable rejection, alias cancellation/uncertainty and both mock harness hooks.
```

Each additional asynchronous scenario has a 2-second bound; whole-command shell
bound is10s. Fetch is entirely mocked, including the fixture's accepted-work map;
no sockets, process, harness, model or engine starts. Hook modules are loaded
in-memory with explicitly synthetic metadata only to test state, not to claim a
managed product handshake. Native timing/process drain remains unvalidated here.
Final `node --check` on each of the three changed assets passed; the progress
document's Git whitespace check reported no errors (only LF/CRLF normalization).

No schema/data type/tool-spec digest or persisted native contract changed. No Go
compilation, native probe, 155s wait, runtime registration, SQLite or frontend edit
was performed for B-R1. Manager will provide the stable A seam before a single
combined compilation. This correction does not approve CP1b or release CP2.

## CP1b additive implementation (current checkpoint)

User decision relayed by manager: **Codex new native threads only** (reference
`85e3f4d3845b2f7b80dd2e1e812c0f90` in manager history, not a local user event).
Legacy adoption research is closed. Old native threads, including old general,
retain normal chat and capability-unavailable detail; no fork, migration or pin update.

New unregistered source files:

| Files | Delivered contract |
| --- | --- |
| `engine/delegation_records.go` | Concrete three-domain records, enum states, scope/native/finality identity, structural validators, pure state/delivery/quarantine helpers. |
| `engine/delegation_tools.go` | Final v1 1–8-child request, max concurrency4/one batch guards, strict decoding/hash conflict checks, private bounded RPC DTOs, all-terminal results and immutable result digests. |
| `engine/delegation_codex.go` | V1 dynamic spec/golden digest/manifest, exact-turn call parser, safe native result encoder and legacy-vs-new compatibility detail helper. |
| `engine/delegation_prompts.go` | Dedicated leaf-role prompt composition with specialist + self-contained task; no parent role/history injection or subagent_report. |
| `engine/delegation_assets.go` and `delegation-{protocol,pi,opencode}.mjs` | Embedded inert assets, pure private-profile renderer, short authenticated RPC, all-result native Promise, Pi sequential mode and OpenCode per-input task suppression + unconditional gate. |
| `engine/delegation_contract_test.go` | One focused pure contract test, no SQLite/harness/model. Optional temp-only asset materialization for probes. |

No existing runtime/adapter/catalog, A-owned shared persistence/harness/capability
files or frontend file was edited. No init/route/tool registration was added.
`AGENTS_UNIFICATION_DELEGATION_PERSISTENCE_HANDOFF.md` is the concrete root/index/
validation/recovery/native-manifest/hook handoff; it supersedes CP0 Go sketches.

The invocation uses **A's `AgentExecutionCapture`** directly; child binding is not
duplicated. V1 complete dynamic spec digest is golden-guarded at
`2b9a6f688babd8f22a33a631ab0a707c77c050a9ad6a3214c9d390b4a83902c0`.
Changing name/schema/description requires explicit version compatibility handling.

### CP1b checks and actual assets

- `gofmt -w` was limited to the six new B Go files.
- Targeted `go test -tags dev -run '^TestDelegationContracts$' -count=1 -v .`
  succeeded while implementing the additive contract (0.137s / 0.174s / 0.142s).
  Rechecks followed changes to result/finality/digest guards or fixture asset
  materialization; no full suite or unrelated tests ran.
- The **last** same targeted check after the final additive DTO change was blocked:
  `sqlite_agents.go:17: undefined: json`. That is an A-owned concurrently edited
  file; B did not alter it. Do not describe the final combined package as passing.
  Manager/A must repair and repeat the one target after shared integration settles.
- `node --check` passed for the three new wrapper assets. No 155s wait was repeated.

Actual compiled-schema assets from the successful test were rendered under:

`C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-assets-2097304649`

The optional fourth argument to `local-tool-fixture.mjs` loads those actual B
assets in an owned fresh native process, with a **simulated private delegation
backend** and deterministic local provider. Provider parameters are asserted
deep-equal to the complete Go-generated schema. No real engine service/children run.

| Pin / fixture root under approved temp parent | Outcome / duration | Cleanup |
| --- | --- | --- |
| Pi0.85.1 / `delegation-pi-wRLGqI` | passed, **5,228ms**; actual versioned asset handshake, pending2s/readers/unique continuation, cancel/no replay, explicit follow-up;4 local provider calls. | PID4256 EOF code0, absent. |
| OpenCode1.18.35 / `delegation-opencode-eV53e9` | passed, **6,084ms**; same checks, actual complete schema, native task absent from model catalog. | Root44576 + observed descendant52900 tree termination confirmed, both absent. |

OpenCode1.18.35's exact `tool/registry.ts::fromPlugin` has an explicit raw-JSON
compatibility branch: non-Zod property definitions use `Schema.Unknown`, preserving
arguments. The new plugin uses that branch and `tool.definition` to publish the
full schema. Thus no SDK/zod installation, ambient resolver or unimplemented schema
factory is needed. The actual pinned provider request attested the full v1 schema.

The final output-budget and pre-admission rejection refinements preserve that
schema/handler shape. The current JS source was additionally checked by the
hermetic `protocol-check.mjs` using the frozen Go-rendered metadata (source override
avoids another package build while A's unrelated import is broken). It passed:
one active batch, exact retry without duplicate work, denied gate honored, two
members consolidated, and unsafe finality triggers cancellation rather than a
native result. It wrote `protocol-check.json` in the asset fixture root. No harness
or provider is started by that check.

Important refinements: proven atomic input/busy/conflict rejection does not abort
an existing admitted batch; ambiguous/post-admission failures do stop the owner.
The native consolidated budget is32KiB and rejects rather than silently truncates;
CP2 must prepare bounded all-member summaries/content references before activation.
Actual record digest/finality guards and quarantine are pure helpers, not durable
admission/restart implementation.

Commands for these short actual-asset checks (parent existence verified first):

```powershell
node "engine/probes/agent-delegation/local-tool-fixture.mjs" pi "C:\Users\Samuel\AppData\Roaming\npm\node_modules\@earendil-works\pi-coding-agent" "C:\Users\Samuel\AppData\Local\Temp\opencode" "C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-assets-2097304649"
node "engine/probes/agent-delegation/local-tool-fixture.mjs" opencode "C:\Users\Samuel\AppData\Roaming\npm\node_modules\opencode-ai\bin\opencode.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode" "C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-assets-2097304649"
node "engine/probes/agent-delegation/protocol-check.mjs" "C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-assets-2097304649" "engine/delegation-protocol.mjs"
```

Native probe bounds stayed90s scenario/115s shell; actual durations above. Hermetic
protocol check had10s abort/15s shell bounds, and its loopback server was closed.
Targeted Go commands had120s shell bounds. Final report stops this checkpoint;
CP2 persistence/runtime activation remains unreleased.

## Released scope and changed files

Manager released only bounded, disposable pinned-harness protocol probes, a local
deterministic provider when needed, and documentation/handoff. No real providers,
user credentials, user/native histories, installed data/process mutation, graph
execution, user-session creation, delegation of work, product enablement or durable
foundation implementation is included.

This checkpoint changed only:

- `engine/probes/agent-delegation/codex-dynamic.mjs` — isolated Codex protocol driver,
  local HTTP Responses fixture, request/result logging, deadlines and owned native
  process cleanup. New-thread/long-wait/resume scenario was exercised during the
  continuation below; the initial report stopped before executing it.
- `engine/probes/agent-delegation/codex-source.mjs` — public-source inspection at the
  exact 0.153.4 release commit, downloaded only to disposable temporary directories.
- `engine/probes/agent-delegation/codex-artifacts.mjs` — offline persisted-spec/result
  attestation of the already-created Codex fixture artifacts, no new native process.
- `engine/probes/agent-delegation/local-tool-fixture.mjs` — short independent Pi and
  OpenCode native-local-tool fixtures, with deterministic loopback Chat Completions.
- `AGENTS_UNIFICATION_DELEGATION_CONTRACT.md` — incorporated manager-approved wait/
  sibling/capacity/unnamed fallback semantics and added minimal Go persistence shapes
  for A in section 10.2. A owns all schema/domain/store/API/persistence integration.
- This progress file.

Frontend and concurrent A/C changes are untouched. No build/full suite/integration
test suite was run; this is a standalone disposable native protocol probe.

## Actual timeout calculation

Inspected constants:

- `engine/models.go`: `modelCatalogTimeout = 45s`.
- `engine/linked_bridge.go`: linked server = `45s + 10s = 55s`;
  graph tool server = `3*45s + 10s = 145s`;
  `linkedToolClientTimeout = 145s + 5s = 150s`.
- `engine/codex_library.go` projects that **150-second** timeout on the owned
  `klm_linked` MCP server. External Library has a distinct 330-second timeout.

The scripted new-thread scenario holds the dynamic request for **155,000ms**,
one wait beyond the real ordinary linked MCP client timeout, under a 230-second
scenario deadline. Initially it was not run because the adoption blocker was reached
first. Under the manager's subsequent explicit CP1a continuation release it ran
once past the long-wait boundary, measuring **155,010ms** (details below).

## Initial checkpoint commands (retained evidence)

All shell calls had finite tool timeouts; no server/watch process was left running.

1. `git status --short` — baseline recorded; unrelated work preserved.
2. `Get-Command node,pi,opencode,codex -ErrorAction SilentlyContinue | Format-List Name,Source,CommandType`
   — found Node and pinned Codex binary; Pi/OpenCode command shims were only located,
   not executed and no wrapper source was imported from those shims.
3. `node --check "engine/probes/agent-delegation/codex-dynamic.mjs"`
   — exit 0; syntax check only (10-second shell deadline).
4. `& "C:\Users\Samuel\AppData\Local\Programs\OpenAI\Codex\bin\codex.exe" --version`
   — exit 0, `codex-cli 0.153.4` (10-second deadline).
5. After `Test-Path -LiteralPath "C:\Users\Samuel\AppData\Local\Temp\opencode"`:

   ```powershell
   node "engine/probes/agent-delegation/codex-dynamic.mjs" "C:\Users\Samuel\AppData\Local\Programs\OpenAI\Codex\bin\codex.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode" adoption
   ```

   Exit 0 with structured `outcome="blocker"`; scenario deadline 60 seconds,
   shell deadline 90 seconds. Exit 0 means the fixture established its asserted
   blocker conditions, **not** that delegation feasibility passed.
6. `git diff --no-index --check -- NUL` against the progress/contract documents and
   probe script — no whitespace errors (normal LF/CRLF normalization notices).

## Executed adoption fixture and artifacts

Artifact root:

`C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-codex-adoption-2ruv3b`

Files: `report.json`, `events.jsonl`, `provider-1.json`, `provider-2.json` and disposable
HOME/config/native history below that root. These are fixture-only artifacts,
retained for manager inspection. No transcript was edited by the probe.

- Local provider URL: `http://127.0.0.1:54966/v1` (closed at cleanup).
- Actual native thread ID: `01a18f29-1289-79d2-adfe-09ee56f3777b`.
- Native app-server PIDs: **8848**, **32792**, both recorded closed and absent after
  cleanup. First process closed at 1,646ms; second at 2,153ms.
- Total probe duration: **2,154ms**.
- At 2,120ms the resumed thread provider surface was observed; at 2,130ms the
  supported-protocol blocker assertion was recorded.

Isolation: child environment is constructed from a minimal Windows runtime/PATH
allowlist, with HOME/USERPROFILE/CODEX_HOME/APPDATA/LOCALAPPDATA/TEMP/TMP and user
skills redirected below the fixture root. No auth/API-key/proxy variables are
forwarded. The local fixture rejects provider Authorization headers. Provider
selection is `klm_fixture`, `requires_openai_auth=false`, Responses HTTP loopback;
remote model discovery/external migration and native skill/plugin/app/multiagent
sources are disabled. Native tool instructions/history are freshly generated in
the disposable cwd, not imported from this project's or the user's native history.

Sequence:

1. Start installed Codex app-server; initialize with experimental protocol enabled.
2. `thread/start` **without** `dynamicTools`, local model/provider fixture only.
3. `turn/start`; local deterministic final response; native `turn/completed` is
   `completed`. Actual provider tool catalog does not contain `subagent_run`.
4. Close stdin and wait for native process close, with a five-second bounded EOF
   path and owned-PID-tree fallback if needed. Assert PID no longer exists.
5. Start a fresh app-server with the **same** disposable CODEX_HOME and resume the
   **same thread ID** through supported `thread/resume` (no fork/new thread).
6. `turn/start`; local final response; `completed`; provider request still lacks
   `subagent_run`. There were two provider requests, both served locally.
7. Inspect the schema exported from this exact installed binary at CP0:
   `ThreadResumeParams.properties` has no `dynamicTools`; scanning request
   definitions consuming `DynamicToolSpec` yields **only `ThreadStartParams`**.
8. Close the second app-server, assert native PID absent, close all local fixture
   connections and HTTP server. Both native processes exited through EOF cleanup;
   no built-in shell/tool execution or independently detached process was requested.

Observed provider tool catalog after resume:

```text
exec_command, write_stdin, request_user_input, view_image,
get_goal, create_goal, update_goal, web_search
```

These were catalog observations only; the deterministic provider emitted no tool
calls and the probe did not invoke them. Native EOF/close + PID absence is the
measured cleanup evidence here; it is not validation of graph Job containment,
arbitrary tool descendants, remote work cancellation or production Stop.

Raw native notices include deprecated `connectors` flag, under-development
`skip_host_skill_discovery` and missing metadata for fixture model. The latter is
fallback **metadata**, not a real-model/provider substitution: both captured HTTP
requests have `model="fixture"` at the loopback provider. No fixture provider error
or credential-bearing Authorization header was observed.

## Initial blocker interpretation and technical options

Retained from the first checkpoint; the directed source investigation requested in
option 1 was subsequently completed below without finding a supported adoption seam.

**Confirmed blocker for the proposed exported API design:** on installed 0.153.4,
dynamic-tool registration is exposed at thread creation, but the exported supported
resume API supplies no registration field/method. A thread born without the tool
resumes without it on the real provider surface. Therefore new-thread support cannot
be treated as adoption support for existing native histories.

This does not claim no future/internal upstream mechanism can exist. It states that
the proposed supported app-server path is missing an adoption seam, and no compatible
mechanism was established. No undocumented resume field, transcript rewrite,
`history` override, silent fork/replacement or giant MCP timeout was attempted.

Options for manager direction before further implementation:

1. Research/verify a **supported native same-thread tool-registration or configuration
   update seam** on the installed pin, if one exists beyond the exported API. Require
   actual provider-surface attestation, preserved native ID/history, pending-wait and
   cancellation proof. Absence from this schema must not be bypassed by assumptions.
2. If the installed pin lacks that seam, a compatible upstream/native API change or
   different approved pin would be needed to retain the dynamic-tool design across
   existing histories. This checkpoint does not authorize installation or changing pins.
3. Investigate a different Codex delegation transport only under a new scoped release,
   proving an actual engine-enforced wait/continuation boundary. Ordinary bounded MCP
   calls, keepalive/progress, or interrupt-after-receipt alone do not establish it.

Not proposed as silent fallbacks: enabling new-thread-only Codex coverage, replacing
existing threads, increasing timeouts to an arbitrary huge value, or declaring old
histories migrated just because the ordinary chat still resumes. Those change the
approved acceptance contract and need a manager/user decision.

## Continuation release and exact-pin source investigation

Manager accepted only the limited original adoption observation and released a
directed same-thread source investigation plus independent new-thread Codex and
Pi/OpenCode probes. The original adoption scenario was **not repeated**.

GitHub annotated tag `rust-v0.153.4` resolves through tag object
`042fb41b7c813ac7999105e886b2b7aa715b5081` to commit
**`3d2ee51ca2d5db578f328aa75e20aa22c0197c9a`**. Every inspected source file below was
downloaded using that exact commit, not `main`. Artifact roots:

- `C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-codex-source-pmWtaI`
- `C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-codex-source-5qeoaF`

Each contains a `manifest.json` with original public URL, local filename, SHA-256
and matching symbol/line numbers. No build/installation of source occurred.

Directed question: can `thread/resume.config` or a supported update/runtime API
provide the startup dynamic tool vector without replacing the native history?

| Exact-pin source | Finding |
| --- | --- |
| `codex-rs/app-server/src/request_processors/thread_processor.rs:1407-1453` | New-thread path unwraps/validates caller `dynamic_tools` and passes them in start options. |
| Same file `:3761-3837` | Resume overrides load through `config_manager.load_for_cwd` into typed `Config`, then `thread_manager.resume_thread_with_history`. Open JSON transport does not itself provide an arbitrary runtime vector. |
| `codex-rs/config/src/config_toml.rs`, `ConfigToml` and `ToolsToml` | No dynamic tool registration field. `ToolsToml` exposes web search, experimental request-user-input and update-plan settings, not an arbitrary function catalog. Random unknown keys could be ignored; they are not evidence of registration. |
| `codex-rs/core/src/config/mod.rs`, `Config` | No dynamic-tools field; `extra_config` is separate persistence metadata, not a function registration map. |
| `codex-rs/thread-store/src/types.rs:67-93` | `ExtraConfig` is an empty struct; `CreateThreadParams.dynamic_tools` is a distinct startup field. It does not supply a supported `resume.config` escape hatch. |
| `codex-rs/core/src/thread_manager.rs:229-261,1084-1117` | `StartThreadOptions.dynamic_tools` defaults to an empty vector. Resume constructs options from history/config and the defaults; no caller dynamic vector is injected there. |
| `codex-rs/core/src/session/session.rs:135,854-864` | Session dynamic tools are an internal session-configuration vector, persisted as startup metadata. No supported live config setter was found in this path. |
| `codex-rs/app-server/src/dynamic_tools.rs`, `on_call_response` | Waits for the client response and submits `Op::DynamicToolResponse`; this is result delivery, not adding/changing tool specs. |
| `codex-rs/core/src/tools/handlers/dynamic.rs` | `DynamicToolHandler` has no parallel-dispatch opt-in on this pin, and `request_dynamic_tool` waits on the response oneshot. This differs from assumptions based on newer `main` source. |

Conclusion: **no supported same-thread config/update seam was identified on the
exact pin**. Existing-thread adoption remains a compatibility blocker for this
design. No speculative field variant was run. A future upstream API/pin decision
or independently proven transport is still needed; no product/version choice is
made by the worker. `resume.config` being open is correctly distinguished from
typed configuration actually defining the required registration behavior.

## Codex new-thread/resume/cancellation fixture

### Corrected assumption before the long wait

First `new-resume` attempt:
`C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-codex-new-resume-u0hvA8`.
It failed at **16,043ms**, waiting for a second dynamic call to start concurrently
with the first. The first call was observed, the second had not begun. PID **13844**
closed and was absent; final outcome was `failed`. **No 155-second wait occurred**
in this attempt. Inspection of the exact-pin handler justified correcting the
assumption rather than retrying blindly: dynamic siblings are serialized here.

### Successful measured run

Root:
`C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-codex-new-resume-v06w3A`.
Native thread: **`01a18f2f-7257-7503-bd7e-2d153894a655`**.
Total duration **158,018ms** under 230,000ms scenario / 260,000ms shell deadlines.

| At elapsed ms | Evidence |
| --- | --- |
| 533 | Actual `item/tool/call` for registered `subagent_run`, native JSON-RPC response ID 0; one local provider request. |
| 155,543 | After one **155,010ms** wait: still exactly one provider request, no native turn completion; `thread/read` responded while the request was pending. |
| 155,608 | After releasing the main result, the **second dynamic tool** `fixture_sibling` was dispatched from the same native response (response ID 1). It was not a built-in and did not start concurrently in this fixture. No next model request before this sibling's result. |
| 155,698 | Both tool responses supplied; exactly one provider continuation, consuming the consolidated main result; native completed. |
| 156,137 | Fresh app-server resumed the same native thread, with no `dynamicTools` in the resume request. Persisted `subagent_run` was present in the actual provider surface and invoked as `call_3_main`; the fresh client handled it. |
| 157,982 | `turn/interrupt` during that pending request produced native `interrupted`. A third app-server resumed the same ID; 1.5s bounded idle observation saw no tool call or additional provider request/re-execution. |

Exactly **three** local provider requests occurred. No actual child executor was
run: the harness tool requests/results and parent scheduling are what this proves.

Offline `codex-artifacts.mjs` checked first-versus-resumed full dynamic function
spec equality and the expected task JSON schema, and found exactly **one** main
consolidated tool output in the continuation. `spec-attestation.json` records:

```text
specSHA256 = 506680bc147d2e7f4e99519e5b12c28afb1455e71cd9fce4ad59a76923a02eea
identicalSpecAfterProcessRestart = true
requiredSchemaMatched = true
consolidatedOutputCount = 1
```

Wrappers must continue supporting this persisted schema/name across process
restarts. Changing the registered schema in engine code does not rewrite existing
native specs; require compatible versioned handling or fail preparation explicitly.
This persistence evidence applies only to a thread **born with** the tool, not to
adoption of legacy histories.

PIDs **9960**, **51164**, **13296** all exited code 0 via EOF+close and were checked
absent (155,738ms / 156,206ms / 158,018ms). Local provider was closed. `report.json`,
`events.jsonl`, three provider requests and spec attestation are retained.

## Pi 0.85.1 local-tool fixture

Root: `C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-pi-xJQ24Y`.
Installed package metadata asserted version **0.85.1**. Actual installed CLI ran in
RPC mode under a disposable agent dir/HOME/session with a local `models.json`.

The owned extension registers a local `subagent_run` with `executionMode=sequential`
and a plain object schema. No native pi-subagents extension/package is loaded.
`session_start` sets the active inventory to this tool; all available tools were
the known native built-ins plus this local tool, with no native `subagent`.

- At 4,144ms registration/active inventory was confirmed.
- Tool Promise stayed pending for a 2,000ms observation with one provider request;
  `get_state` responded and reported streaming; no `agent_settled` or next model call.
- Releasing the fixture gate produced one continuation and `agent_settled`.
- Second pending tool was cancelled via `clear_queue` + `abort`; the extension's
  abort signal cancelled its short RPC wait and the fixture observed cancellation.
  A 300ms bounded observation saw no replay or next provider call.
- Explicit later follow-up completed normally without repeating the cancelled tool.
- Four total local provider requests; total duration **7,036ms**; native PID **28344**
  exited code 0 by EOF and was absent. Fixture server/short pending requests closed.

This establishes native Promise/RPC/abort feasibility for the local wrapper, not
product permission routing, engine child containment or graph seals.

## OpenCode 1.18.35 local-plugin fixture

Actual executable `--version` returned **1.18.35**, and fixture `/global/health`
also asserted that version. Fresh private config/data/cache/native DB/HOME and an
owned local plugin were used. `args:{}` needs no SDK/zod installation: the pin's
public plugin `tool()` helper is an identity wrapper and the native registry handles
the raw empty argument shape. No npm package was installed.

First attempt:
`C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-opencode-oJfOTd`,
duration **3,053ms**, outcome `failed`. Plugin tool registration succeeded, but
config-level `tools.task=false`/permission denies did **not** remove `task` from the
actual model catalog in this fixture. Fail-closed provider assertion stopped before
emitting any model tool call. This is retained evidence, not hidden as a pass.

Concrete correction came from exact-pin source:
`packages/opencode/src/session/llm/request.ts::resolveTools` filters on
`input.user.tools[k] !== false`, and `session/prompt.ts::createUserMessage` provides
that per-user tools map. The public `chat.message` hook can set it before persistence.
The owned plugin now sets `output.message.tools.task=false` on each accepted user
input, keeps config denies and an unconditional pre-execution `task` rejection.
This requirement belongs in future product wiring, including under YOLO.

Corrected successful run:
`C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-opencode-rJ3QYX`.
Native fixture session: **`ses_e70cacad0ffeJmMw5bZHnHJc5y`**, as recorded in
`report.json.nativeSessionId`.

- At 1,343ms local plugin registration was confirmed. The native registry endpoint
  still lists `task`, as expected for an installed built-in; that is not the model's
  permitted inventory.
- Actual provider surface included `subagent_run` and **excluded `task`**.
- Tool Promise stayed pending for 2,000ms with one provider call; native SSE events
  remained readable, `/session/status` responded `busy`, and no early idle/model step.
- Gate release produced a unique continuation and idle.
- A second pending local-tool request was cancelled through native `/abort`; its
  AbortSignal cancelled the short gateway wait, and gate cancellation was observed.
  No automatic next model call/replay in the 300ms observation.
- Explicit follow-up completed, for four total local provider requests.
- Total duration **5,699ms**. Owned serve PID **30328** and observed descendant
  **29848** were terminated by `taskkill /PID 30328 /T /F`; native close confirmed.
  First attempt's PID **4868** and observed descendant **52296** were likewise
  terminated. A separate one-shot read verified all four recorded PIDs absent.

Source references for the local tool API and suppression path are tied to
`https://github.com/anomalyco/opencode/tree/v1.18.35`, including `packages/plugin/src/tool.ts`
(`ToolContext.abort` and raw tool definitions) and `packages/plugin/src/index.ts`
(`chat.message` and tool-before hook). The unconditional before gate was installed;
this probe did not ask the model to bypass a hidden task and does not prove every
other native delegation entry point impossible. No native task executor was invoked.

## Continuation commands, limits and cleanup corrections

All probe launch commands below first checked their existing approved temporary
parent. Paths are literal arguments; no user process/server was reused.

```powershell
node "engine/probes/agent-delegation/codex-source.mjs" "C:\Users\Samuel\AppData\Local\Temp\opencode"
node "engine/probes/agent-delegation/codex-source.mjs" "C:\Users\Samuel\AppData\Local\Temp\opencode" session
node "engine/probes/agent-delegation/codex-dynamic.mjs" "C:\Users\Samuel\AppData\Local\Programs\OpenAI\Codex\bin\codex.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode" new-resume
node "engine/probes/agent-delegation/codex-artifacts.mjs" "C:\Users\Samuel\AppData\Local\Temp\opencode\delegation-codex-new-resume-v06w3A"
node "engine/probes/agent-delegation/local-tool-fixture.mjs" pi "C:\Users\Samuel\AppData\Roaming\npm\node_modules\@earendil-works\pi-coding-agent" "C:\Users\Samuel\AppData\Local\Temp\opencode"
node "engine/probes/agent-delegation/local-tool-fixture.mjs" opencode "C:\Users\Samuel\AppData\Roaming\npm\node_modules\opencode-ai\bin\opencode.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode"
```

- Source calls: per fetch 15s, max 1MiB/file, shell 90s and 25s respectively.
- Codex new-resume command ran twice: failed sibling assumption at 16,043ms, then
  corrected scenario at 158,018ms; **only the latter performed the long wait**.
- Pi shell limit 115s / scenario 90s; OpenCode same, two executions with the concrete
  correction above. Local wrapper admission uses short RPC and read waits <=800ms
  with <=2s fetch deadlines; there is no long blocking MCP request.
- Codex specs/result offline attestation: 10s shell bound, no native processes.
- One-shot `Get-Process -Id @(29848,30328,52296,4868)` equivalent in the recorded
  PowerShell check returned no live PID. No polling loop or unrelated process kill.
- `node --check` syntax checks and Git whitespace checks cover the edited scripts/
  documents. No model-backed test, product build or full suite is claimed.

The reviewed deadline/cleanup issues were corrected before continuation runs:
scenario deadlines abort a shared lifetime controller; pending RPC waits/timers are
abort-aware, all future sends/process creation check that controller, and cleanup
then closes existing owned processes. A timeout does not just abandon a Promise
while another branch keeps launching native processes. Logs cap at 8MiB, provider
payloads at 4MiB and frame/event counts are bounded. Cleanup failures set final
`outcome=cleanup_unconfirmed` and nonzero exit; hard cleanup watchdog writes that
failure report before exiting. No stale `passed` outcome is retained on teardown
failure. Future OpenCode cleanup also records/verifies observed tree termination PIDs.

## Remaining acceptance boundaries

Confirmed only by these deterministic **local-provider protocol fixtures**:
Codex new-thread pending barrier/unique result/same-ID tool persistence/cancellation,
Pi sequential local wrapper, OpenCode local plugin wrapper with corrected native
task surface suppression, short transport waits and native readers/abort activity.

At the end of CP1a, legacy adoption was blocked. The subsequent user decision
described in the CP1b section accepted new-native-thread-only compatibility and
closed that investigation. New-thread pass remains explicitly not legacy coverage.

Still unvalidated: arbitrary built-in sibling ordering/effects, all native bypass
paths, production role/permission/credential isolation, engine child execution and
all-results batching, Stop/process domain aggregation, SQLite admission/restart/
delivery, graph Choice/finality/workspace snapshots, UI and real model/provider
behavior. The sibling evidence here is specifically **serialized dynamic tools**,
not universal tool parallelism or built-in containment.

This CP1a report enabled no feature or durable foundation. Its compatibility options
were presented to the user; the resulting new-native-thread-only decision is
implemented only as the additive eligibility helper in CP1b. No upstream/pin,
transport or native-history change was selected by the worker.

## Persistence handoff

Contract sections 5, 10.1 and **10.2** provide A with the three delegation domains,
standalone Go record-shape proposal (standard strings/JSON only), indexed keys,
atomic transaction boundaries, unique ownership/receipt constraints, quarantine of
uncertain children, bounded no-replay recovery and graph/parent/role validation.
No dependency on not-yet-compiled A types was added to Go source, and no storage/
schema/API/index/migration file was edited. Durable implementation remains CP1b.
