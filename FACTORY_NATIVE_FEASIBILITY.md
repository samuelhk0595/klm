# Factory native feasibility — step 0

Date: 2026-11-14. Branch: `release/0.7.0`. Platform: Windows amd64.

**Result: the exercised native mechanisms are feasible on all three pinned
harnesses.** Authenticated native model/tool calls demonstrated external
operational tools, bounded continuing shell jobs, actual next-step input, explicit
unanswered human requests and owned-process cleanup. No pinned-version
incompatibility or provider-readiness blocker was established.

This is step-0 mechanism evidence, not a delivered persistent runner, an exact
exclusive catalog attestation for every native/provider-side tool, or human
acceptance of the Factory journey. Scope limits are explicit below.

## Authorization and preserved work

The current user commissioned implementation and brief real native probes, with
step 0 first. That instruction authorizes this work independently of older
planning-only labels. It does not convert every proposal or mock experience in
the companion documents into human acceptance.

Only `FACTORY_NATIVE_FEASIBILITY.md` changed in the repository. Existing modified
specifications, the untracked backend plan and three local specialist TOMLs were
preserved. No production implementation, runner persistence/API, business code,
frontend, upgrade, installer, migration, deployment or application server was
introduced. Codex used its owned stdio app-server protocol; OpenCode used `run`,
without starting `serve` or a listening probe server. No delegation was executed.

## Native versions and isolated authentication

| Harness | Exact executable | Actual version stdout |
| --- | --- | --- |
| Pi | `C:\Program Files\nodejs\node.exe` with `C:\Users\Samuel\AppData\Roaming\npm\node_modules\@earendil-works\pi-coding-agent\dist\cli.js` | `0.85.1` |
| OpenCode | `C:\Users\Samuel\AppData\Roaming\npm\node_modules\opencode-ai\bin\opencode.exe` | `1.18.35` |
| Codex | `C:\Users\Samuel\AppData\Local\Programs\OpenAI\Codex\bin\codex.exe` | `codex-cli 0.153.4` |

Versions came from actual bounded `--version` executions in the earlier attempt;
no binary changed subsequently. Pi used the adapter-compatible unbundled CLI.

Relevant existing seams read before continuing: `engine/interactive.go`,
`platform_windows.go`, `codex_interactive.go`, `codex_library.go`,
`delegation_codex.go`, `pi_interactive.go`, `pi-permissions.ts`,
`opencode_interactive.go`, `opencode_library.go`, and
`opencode-permissions-plugin.mjs`.

The production Codex adapter preserves native auth/home; the probe retained its
normal native file-auth mechanism while copying credentials into disposable native
auth storage so history and auth refresh could not mutate the originals. Pi used
its existing `openai-codex` OAuth credentials, not the unconfigured `openai` API-key
provider. OpenCode projected only the existing `openai` OAuth entry, using the
managed private discovery/profile environment and dependency-lock layout. Provider
credentials were confined to native auth files, never tool manifests, TOMLs,
prompts or evidence. Each temporary auth file was removed in `finally` cleanup.
Every authenticated run verified the original auth file's SHA-256 was unchanged.

Each run used a disposable absolute cwd and redirected home, app-data, XDG,
native-profile and `KLM_USER_SKILLS_DIR` paths below:

`C:\Users\Samuel\AppData\Local\Temp\opencode`

Inherited harness/provider environment configuration was removed, including the
owning execution's `OPENCODE_CONFIG_CONTENT`. Models were explicit and supported:
Codex `gpt-6-astra` / low, Pi `openai-codex/gpt-6-astra` / low, and OpenCode
`openai/gpt-6-astra`. Codex and Pi confirmed that model through their actual native
catalogs; OpenCode executed real provider requests with it.

## Exact invocations and retained evidence

Run from `C:\Users\Samuel\AppData\Local\Temp\opencode`:

```powershell
& "C:\Program Files\nodejs\node.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode\factory-codex-authenticated.mjs"
& "C:\Program Files\nodejs\node.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode\factory-pi-authenticated.mjs"
& "C:\Program Files\nodejs\node.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode\factory-opencode-authenticated.mjs"
& "C:\Program Files\nodejs\node.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode\factory-opencode-authenticated.mjs" --handoff-only
& "C:\Program Files\nodejs\node.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode\factory-codex-authenticated.mjs" --surface-only
& "C:\Program Files\nodejs\node.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode\factory-codex-authenticated.mjs" --surface-control
& "C:\Program Files\nodejs\node.exe" "C:\Users\Samuel\AppData\Local\Temp\opencode\factory-codex-authenticated.mjs" --exec-bypass-only
```

The Codex surface comparison ran its two independent disposable native processes
concurrently. These scripts contain the exact native argv, captured controls,
schemas and native calls. `evidence.json` is retained in each directory below,
relative to the temporary root:

- `factory-codex-authenticated-20261114`
- `factory-codex-authenticated-20261114-surface-off`
- `factory-codex-authenticated-20261114-surface-control`
- `factory-codex-authenticated-20261114-exec-bypass`
- `factory-pi-authenticated-20261114`
- `factory-opencode-authenticated-20261114`
- `factory-opencode-authenticated-20261114-handoff`

Disposable supporting modules in that root: `factory-probe-jobs.mjs`,
`factory-pi-extension.mjs`, and `factory-opencode-plugin.mjs`. They are probe
support, not a new production adapter or durable job implementation.

The owned job workload was this single continuing PowerShell process, using 9 or
20 seconds for completion and 60 seconds for cancellation:

```powershell
powershell.exe -NoLogo -NoProfile -NonInteractive -Command "Write-Output ('START:'+$PID); Start-Sleep -Seconds <duration>; Write-Output ('END:'+$PID)"
```

Foreground tools and job waits returned after an elapsed-time budget of about
4,650 ms, below the required 5,000 ms ceiling even when arguments requested
60,000 ms. The timer released the wait, not the process. Completion used the
owned process's close event, not model polling. Stop used bounded
`taskkill.exe /PID <owned-pid> /T /F` and awaited process close. These disposable
measurements are not a production timing guarantee or kernel-domain stress test.

## Codex 0.153.4 — actual native evidence

`initialize` enabled `experimentalApi`. `thread/start` captured explicit model,
`approvalPolicy:"never"`, `sandbox:"danger-full-access"` and dynamic JSON-schema
tools `shell_start`, `job_read`, `job_wait`, `job_cancel`, `factory_report`.
Native replies and `thread/settings/updated` confirmed `never` /
`dangerFullAccess`. The managed Library controls disabled skill/plugin/app/hook
sources and both multi-agent features; additional native controls were:

```text
-c features.shell_tool=false -c features.unified_exec=false
-c features.multi_agent=false -c features.multi_agent_v2=false
-c features.default_mode_request_user_input=true
-c tools.experimental_request_user_input.enabled=true
-c cli_auth_credentials_store="file"
app-server --listen stdio://
```

- `shell_start(mode=complete,waitMs=60000)` arrived as a real
  `item/tool/call`, native request ID **0**. PID **14180** started once. Its native
  dynamic-tool duration was **4,663 ms**, returning `running:true` and START output.
- During that outstanding call, `turn/steer` sent client ID
  `factory-step-input` with nonce
  `step-7b8306c3-50b5-4585-85d0-00776f34db50`. Native user-message completion retained
  the client ID, and the next model step called `factory_report` with that exact
  nonce in the **same turn**. This proves more than a steer ACK.
- That turn completed before PID 14180 exited. A later close event retained
  `START:14180` / `END:14180`, exit 0, and triggered an attributed completion input
  on the same native thread.
- A necessary label question produced real `item/tool/requestUserInput`, with
  exact choices Alpha/Beta and `autoResolutionMs:null`. No answer was sent.
- A further turn started owned PID **28044**. Its shell handoff took **4,655 ms**;
  `job_wait(waitMs=60000)` returned in **4,805 ms**, still running. The native
  question then remained unanswered while that job continued.
- `turn/interrupt` returned, followed by native `turn/completed` with
  `status:"interrupted"` and `serverRequest/resolved`. Explicit owned-job Stop
  drained PID 28044 with exit 1 and no END marker. Stdio EOF exited Codex 0.

One initial authenticated probe exposed a **probe timing defect**: starting a
4,800 ms timer after Windows spawn overhead produced a 5,244 ms native tool
duration. The corrected measurement above budgets from tool entry. This was
fixed only in disposable support; no production timer was changed.

### Native shell controls, rather than prompt-only replacement

Two additional real authenticated comparisons used the same harmless request,
without a prompt instruction banning native shell use:

- With shell/unified-exec controls **on**, Codex emitted native
  `commandExecution`, source `unifiedExecStartup`, ran
  `echo FACTORY_NATIVE_SHELL_CONTROL`, returned its exact output and exit 0.
- With those controls **off**, the model reported no native shell and called only
  the external `factory_surface`; there was no native `commandExecution`.
  Both runs retained `multi_agent=false` / `multi_agent_v2=false` and created no
  native children.

The model's inventory also named `functions.exec` and other provider-side helpers.
A bounded follow-up reported `typeof process` and `typeof require` as undefined
there and executed no local bypass command. **Those helper names and JavaScript
results are model reports, not a captured native registry or host attestation.**
The native command event in the positive control is direct evidence; the disabled
surface comparison is behavioral evidence. Full provider-payload catalog
exclusivity and every hypothetical alternative path are not claimed here.

## Pi 0.85.1 — actual native evidence

Native arguments used `--mode rpc --offline --no-extensions --no-skills
--no-prompt-templates --no-themes --no-builtin-tools`, the single disposable
extension, an absolute session JSONL path, `--provider openai-codex
--model gpt-6-astra --thinking low`.

- `session_start` positively activated only the four owned job operations,
  `factory_report`, and `ask_user`. `before_provider_request` inspected the real
  outbound payload and recorded exactly those six names: no builtin shell or
  native delegation tool. This applies native selection controls, not prompt-only
  restrictions. Eligible owned operations ran without approval prompts.
- PID **35124** returned running in **4,654 ms** despite `waitMs=60000`.
  A native `steer` sent during the call was present in the **next actual provider
  payload**, and that step called `factory_report` with exact marker
  `factory-pi-next-step-20261114`.
- Native `agent_end`/`agent_settled` preceded the job's later close event.
  `START:35124` / `END:35124`, exit 0, woke the next turn with attributed evidence.
- Continuing PID **46700** had a **4,662 ms** handoff. Its one oversized
  `job_wait(60000)` returned in **4,654 ms**.
- `ask_user` used KLM's existing `ctx.ui.input("klm.question.v1:" + payload)` seam.
  A real correlated `extension_ui_request` retained the exact question/choices.
  No reply was fabricated. Native `abort` settled the turn; an extension control
  drained PID 46700 with exit 1. Native EOF shutdown exited 0.

## OpenCode 1.18.35 — actual native evidence

The native command was `run --format json --model openai/gpt-6-astra`, then
`run --format json --model openai/gpt-6-astra --session <same-id>` for event-driven
continuation. The owned plugin used the existing raw-schema tool registration,
pre-dispatch gate and `experimental.chat.messages.transform` seams. A minimal
disposable file request/reply bridge kept job ownership in the supervisor across
native CLI process exits; it launched no server.

- Private profile/discovery roots, empty external Library/MCP configuration and
  the existing managed dependency-lock layout prevented personal plugin/config
  imports or startup package installation. Native OpenAI authentication remained
  available. Captured YOLO allow policy was followed by explicit shell/delegation
  denies, disabled tool maps and a positive dispatch gate.
- Instrumentation immediately before actual provider `fetch` recorded native
  request tool names. They contained the owned operations and native file/question
  tools, but **no bash, shell, unified-exec, task or subagent tool**. The positive
  pre-dispatch gate permits only the captured probe operations and native question
  tool; out-of-manifest dispatch was not exercised.
- PID **46268** handed off in **4,651 ms**. At the next native model boundary,
  `session.prompt(noReply:true)` persisted the complete new input and the plugin
  added that saved message to the already-captured snapshot, matching KLM's seam.
  Actual outgoing payload inspection confirmed marker inclusion; the next model
  step called `factory_report` with `factory-opencode-next-step-20261114`.
- Continuing PID **2184** handed off in **4,652 ms**. Its one oversized job wait
  returned in **4,653 ms**. A real `question.asked` retained exact session,
  message/call, request and choice identities. The CLI-created question deny was
  replaced on this disposable session with the captured explicit policy; no
  personal policy changed. No question answer was sent. Stopping the native owned
  process and supervisor job drained both; PID 2184 exited 1 without END output.
- The initial completion job finished before that first native turn ended, so one
  targeted `--handoff-only` run closed the unresolved lifetime check. PID **37548**
  handed off in **4,655 ms**, remained running when the native CLI exited 0 at
  14.249 seconds, and completed once at 25.664 seconds with matching START/END PIDs.
  Its close event resumed the **same native session**; the model returned
  `START:37548` / `END:37548`. This was not a next-turn substitute for steering:
  same-turn steering had already been observed in the preceding model step.

The first OpenCode workload made three model-authored `factory_report` calls with
the same marker. The more explicit follow-up workload made one. Native inference
does not provide application-effect idempotency; durable calls/receipts remain
necessary in the commissioned runner.

## Support matrix and remaining implementation boundary

| Mechanism | Pi 0.85.1 | OpenCode 1.18.35 | Codex 0.153.4 |
| --- | --- | --- | --- |
| Authenticated native operational-tool call | Observed | Observed | Observed dynamic calls, including request ID 0 |
| Shell/delegation suppression | Actual outgoing tool list + positive selection | Actual outgoing list + config/input/dispatch gates | Explicit native switches + real shell on/off comparison; inventory qualification above |
| Fixed execution approval mode | Captured owned tools auto-execute | Captured allow policy, authoritative denies after it | Native `never` / `dangerFullAccess` |
| Same process foreground handoff under 5 s | 4,654 ms | 4,651–4,655 ms | 4,655–4,663 ms native durations |
| Oversized job wait remains bounded | 4,654 ms | 4,653 ms | 4,805 ms native duration |
| Event-driven completion after yield | Observed | Observed in targeted lifetime run | Observed |
| Actual next eligible step includes input | Payload + exact tool marker | Persisted snapshot + payload + exact marker | Native client ID + exact marker in same turn |
| Necessary explicit human request under YOLO | Native extension UI request | Native `question.asked` | Native `requestUserInput` |
| Stop and owned continuing process closure | Observed | Owned process shutdown observed | Native interrupt + owned job closure observed |

The same fixed external policy can be captured for coordinator and worker roles;
these probes exercise the native mechanism, not two integrated Factory roles or
an independent Assignment lifecycle. There is no reason here to narrow harness
coverage or upgrade a pin. Native MCP elicitation was not needed because the
necessary question path was exercised. Human answer delivery/resumption, durable
requests/cursors/receipts, restart recovery, native compaction, graph/session
isolation and full ordinary-UI regression acceptance remain implementation or
human-validation work, not inferred successes.

Minimal ownership contract for the next slice: KLM owns immutable external
origin/integration/role/capture, native generation/turn/step, input inclusion
states, operational call IDs, pending requests, jobs and cessation evidence.
Werbert owns Factory/Chat/Ticket/Assignment identities and objective assessment;
those IDs are opaque references to KLM. Browser DTOs should keep operation receipt,
current state, applied capture and input/request delivery states separate, with
decimal-string revisions/cursors and bounded evidence references. Keep these in
the plan's additive `runner_*` and Werbert `foreman_*` modules; this probe adds no
new public endpoint or file contract beyond the existing plan.

## Cleanup, checks and historical setup failures

Native calls had real response and process deadlines: Codex/Pi supervisors bounded
their native process to 60 seconds; each OpenCode native invocation had a 25-second
deadline. Owned Stop calls had 3-second kill-command deadlines. Processes were
awaited and all jobs reached observed close. Post-run process inspection found no
remaining authenticated-probe native process. Temporary native auth-file removal
was independently checked. Original native auth hashes remained unchanged.

The earlier empty-profile attempt produced Pi `No API key found for openai`, Codex
HTTP 401, and OpenCode `Provider not found: openai`. Those were **probe setup
errors**, corrected above, not external readiness blockers. Earlier disposable
version/readiness scripts and `factory-native-20261114/readiness-results.json`
remain historical evidence only. A quoted `node -e` metadata command also failed
PowerShell argument parsing before executing; its replacement file-based metadata
probe succeeded without printing credential values. The attempted public pinned
Codex source URL returned 404 and supplied no compatibility evidence.

Checks actually performed were the bounded native probes and their targeted
follow-ups, actual protocol/payload observations, original-auth hash comparisons,
cleanup checks and documentation readback. No full suite, build, review agent,
production service or preview server was run. **Human acceptance remains pending.**
