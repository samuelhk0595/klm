# Factory backend evidence and integrated-journey handoff

**Latest integration status:** the real Pi browser/human-answer/result/assessment
journey subsequently completed; see [completion report](FACTORY_IMPLEMENTATION_COMPLETION.md)
and [browser evidence](FACTORY_BROWSER_VALIDATION.md). The original evidence and
preparation snapshots below retain their historical scope. Parent changes and
final source-cohort limits are recorded at the end of this file.

Date: 2026-11-14. Windows amd64. Owner: the backend specialist for KLM and Werbert.

Read this file to reproduce the backend checks, identify this specialist's changes,
or prepare the browser specialist's first real integrated journey. The current
user commissioned implementation and bounded native testing; historical
planning-only labels do not revoke that instruction. Implementation authorization,
technical evidence, and human acceptance of documents/mock/UI remain separate.

**Current result:** both backend implementations compiled, focused boundary checks
passed, and actual runner/provider checks passed on all three pinned harnesses.
The complete Werbert → authenticated KLM → Foreman → worker → human answer → result
→ assessment journey has **not** been run. The launcher below is prepared support,
not evidence that this remaining journey works. No server was launched in this
evidence/preparation batch, and passing checks were not rerun to manufacture logs.

## 1. Source and artifact identity

Repository roots used throughout:

- **K:** `C:\Users\Samuel\Documents\Projects\Personal\klm`
- **W:** `C:\Users\Samuel\Documents\Projects\Hermes\werbert`
- **T:** `C:\Users\Samuel\AppData\Local\Temp\opencode`

HEADs read during preparation:

- K: `2fd3020301ab3708399a005fd1aad46db7c53bb2` (`release/0.7.0`).
- W: `d8379e8d8d079918088794d18824d6bc8b994d96`.

Both working trees are dirty. These HEADs identify the base, **not a committed
snapshot of the tested changes**. No clean pre-task patch or complete build-time
source archive was retained. Concurrent changes must not be attributed to this
specialist merely because they appear in `git diff`.

The following artifacts were found during the initial preparation batch. These
historical hashes were superseded by the rebuilt browser artifacts in §7:

| Artifact under T | SHA-256 |
| --- | --- |
| `klm-factory-backend-check.exe` | `34D230D22ED137B3705FF224B20ED7ABC3E88C8CD6ED733F260B84F95C64385F` |
| `werbert-foreman-backend-check.exe` | `EA22D128851E53F185FACEEBFE9F60DF472F82CA624FEBA119C84FBD8E733FC3` |

These identify those earlier executables, not the current concurrent working
tree. Use §7's current hashes for the browser retry. Rebuild an affected backend
if its source changes again before the browser run.

## 2. Exact completed commands and what they establish

Commands in each block were run with that repository subdirectory as the working
directory. They are historical completed commands, not instructions to rerun
everything. Normal proportional-validation policy still applies.

### Compilation

Working directory: **K/engine**:

```powershell
go build -tags dev -o "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-factory-backend-check.exe" .
```

Outcome: exit 0. Dev build, fixed API port 17331; no installed binary changed.

Working directory: **W/engine**:

```powershell
go build -o "C:\Users\Samuel\AppData\Local\Temp\opencode\werbert-foreman-backend-check.exe" ./cmd/engine
```

Outcome: exit 0. This command also passed after the last Ticket list paging / list
projection change. **That final paging change is compile-only evidence**: the
Werbert focused behavior checks below preceded it and were not repeated afterward.

### Focused KLM boundary check

Working directory: **K/engine**:

```powershell
go test -tags dev -run '^TestRunnerDurableBoundary$' -count=1 -timeout=45s .
```

Outcome: PASS, about 1.1 s in the recorded terminal result. Source:
`engine/runner_boundary_test.go`. It uses disposable SQLite and in-process
`httptest` requests with explicit `KLM_DEV_DISABLE_CLIENT_AUTH=1`. It checks changed
operation-body refusal, immutable external origin, unique work key, 404 on the
listed ordinary exact-ID routes, ordinary project inventory/global subscription
exclusion, and original retained native history. It does not exercise real
transport pairing, every ordinary projection, or a live model.

### Focused Werbert checks

Working directory: **W/engine**:

```powershell
go test ./internal/engine -run '^TestForeman(StartupAndDefinitionBoundary|SourceAssignmentAndAssessment)$' -count=1 -timeout=60s
```

Outcome: PASS, about 0.27 s in the recorded terminal result. Sources:
`foreman_boundary_test.go`, `foreman_business_test.go` in `internal/engine`.

- Startup check: Foreman branch excludes legacy runtime construction/recovery,
  preserves the seeded legacy record bytes, refuses stale definition writes, and
  refuses a Connect mutation without the intent header.
- Business check: original human source deduplicates the Ticket; one Assignment
  uses the original work receipt; acknowledgement is deduplicated; structured result
  is retained; assessment fails before native settlement and succeeds against the
  matching settled revision.
- The business check supplies a disposable **controlled KLM HTTP transport**. It
  is not real Werbert-to-KLM/provider end-to-end evidence.

### Actual native runner checks

Working directory: **K/engine**, separately for each indicated harness:

```powershell
$env:TEMP = "C:\Users\Samuel\AppData\Local\Temp\opencode"
$env:TMP = $env:TEMP
$env:KLM_RUN_FACTORY_NATIVE_PROBE = "1"
$env:KLM_FACTORY_PROBE_HARNESS = "pi" # separate successful runs: opencode, codex
go test -tags dev -run '^TestRunnerNativeCodexOwnedBoundary$' -count=1 -timeout=100s -v .
```

`TestRunnerNativeCodexOwnedBoundary` is the historical name for all three cases.
Its opt-in environment variables are read in `runner_native_probe_test.go`.

| Harness | Pinned version | Explicit model / effort | Last recorded successful test duration |
| --- | --- | --- | --- |
| Pi | 0.85.1 | `openai-codex/gpt-6-astra` / low | PASS, 19.48 s |
| OpenCode | 1.18.35 | `openai/gpt-6-astra` / low | PASS, 18.82 s |
| Codex | 0.153.4 | `gpt-6-astra` / low | PASS, 20.24 s |

These checks use actual native provider calls through KLM adapters, while their
runner HTTP requests are still in-process with the explicit dev auth bypass.
They observe an operational marker, a continuing owned shell command, two exact
unanswered native questions, original duplicate worker receipt recovery,
Foreman-only Stop, and independent worker cancellation. The owned-job ledger has
no starting/running jobs at the end. Original saved-Agent bytes and original
native auth SHA-256 remain unchanged. The test has a 75-second internal context,
50-second condition deadlines and the 100-second outer Go timeout.

The native command was:

```powershell
Write-Output ('START:'+$PID); Start-Sleep -Seconds 60; Write-Output ('END:'+$PID)
```

No human answer was synthesized. These checks stop/cancel the unanswered requests;
they do not establish answer delivery or resumption. An early integrated runner
attempt exposed ordinary-only analytics lookup (`sql: no rows in result set`);
runner analytics branching was corrected. Codex also needed its external shell /
question controls in per-thread configuration as well as process launch flags.
The successful durations above are from the corrected paths.

### Evidence retention and earlier native feasibility

The Go-check stdout/results above were recorded in the owning execution's tool
transcript. **There is no separate retained Go-test log file to cite.** Their
`t.TempDir()` stores/native auth copies were removed by test cleanup. Do not cite
an invented test-output directory or mistake unrelated `foreman-preview-*.log`
files for these checks. The build executables listed in §1 remain available.

Step-0 native protocol/payload evidence is retained separately. Read
[FACTORY_NATIVE_FEASIBILITY.md](FACTORY_NATIVE_FEASIBILITY.md) for exact Node
commands, controls, timing corrections and limitations. Existing `evidence.json`
artifacts under **T**:

```text
factory-codex-authenticated-20261114/evidence.json
factory-codex-authenticated-20261114-surface-off/evidence.json
factory-codex-authenticated-20261114-surface-control/evidence.json
factory-codex-authenticated-20261114-exec-bypass/evidence.json
factory-pi-authenticated-20261114/evidence.json
factory-opencode-authenticated-20261114/evidence.json
factory-opencode-authenticated-20261114-handoff/evidence.json
```

Their sibling scripts are `factory-{codex,pi,opencode}-authenticated.mjs` with
`factory-probe-jobs.mjs`, `factory-pi-extension.mjs` and
`factory-opencode-plugin.mjs`. Step 0 observed actual next-step input and sub-five-
second shell handoff/waits, including completion after native yield. It is distinct
from the production runner checks above. Earlier empty-profile 401/provider setup
failures were corrected using disposable native auth copies. Codex model-reported
helper inventories remain behavioral evidence, not complete catalog attestation.

## 3. Specialist-attributable file inventory

This is an ownership inventory, not a claim that every hunk in a shared modified
file belongs to this specialist. Review these files together with their relevant
contract. Preserve concurrent work in the same files.

### KLM additions

Under **K/engine**:

```text
RUNNER_API.md
runner_api.go                    runner_attention.go
runner_boundary_test.go          runner_capabilities.go
runner_capture_gate.go           runner_contract.go
runner_context.go                runner_control_windows_test.go
runner_controls.go               runner_discovery.go
runner_effects.go                runner_inputs.go
runner_jobs.go                   runner_native_probe_test.go
runner_output.go                 runner_recovery.go
runner_reply_storage.go          runner_request_validation.go
runner_runtime.go                runner_store.go
runner_regression_test.go
```

Earlier step-0 addition: **K/FACTORY_NATIVE_FEASIBILITY.md**. This preparation adds
**K/FACTORY_BACKEND_EVIDENCE.md** and:

```text
engine/validation/Start-ForemanJourney.ps1
engine/validation/foreman-journey-agent.md
```

### KLM modified seams

Under **K/engine**:

```text
README.md                       agent_catalog_api.go
agent_resolution.go             analytics_records.go
api.go                          capabilities.go
cli_linux.go                    cli_windows.go
codex_interactive.go             codex_library.go
graph_adapter.go                harness.go
linked_bridge.go                main.go
mcp_client.go                   mcp_connections.go
message_queue.go                opencode-permissions-plugin.mjs
opencode_interactive.go         permission_policy.go
pi-permissions.ts               pi_interactive.go
sqlite_api.go                   sqlite_app.go
sqlite_delegation_upgrade.go    sqlite_execution_queue.go
sqlite_execution_requests.go    sqlite_library_deletion.go
sqlite_runtime.go               sqlite_schema.go
sqlite_session_repository.go    sqlite_store.go
sqlite_updates.go               store.go
```

Changes are external-origin/role capture, positive manifest/dispatch selection,
selected Library capture, steering/inclusion attribution, schema-6 runner ledgers,
ordinary-session guards/projections, protected local discovery, exact native
requests/replies and owned jobs/effect drainage. Saved-Agent source resolution
adds `PreserveSource`; native settings never rewrite original Agent TOMLs.

### Werbert additions

Under **W/engine/internal/engine**:

```text
foreman_admissions.go            foreman_api.go
foreman_boundary_test.go         foreman_business_test.go
foreman_connection.go            foreman_content.go
foreman_contract.go              foreman_controls.go
foreman_definition.go            foreman_dispatch.go
foreman_local_linux.go           foreman_local_windows.go
foreman_operations.go            foreman_outgoing.go
foreman_overview.go              foreman_reply_storage.go
foreman_runtime.go               foreman_store.go
foreman_tools.go                 foreman_updates.go
foreman_views.go
foreman_overview_test.go
```

Other additions under **W/engine**: `docs/foreman-api.md`, `prompts/foreman.md`,
`prompts/foreman-input.md`, `prompts/factory-worker.md`.

### Werbert shared-file changes

- `internal/engine/config.go`: explicit Foreman/legacy mode and literal-loopback
  requirement.
- `internal/engine/server.go`: Foreman startup/handler/close branches.
- `internal/engine/store.go`: `openStoreMode(root, foreman)` and namespaced schema
  branch before legacy migrations. Existing/concurrent transactional DDL helpers,
  connection PRAGMAs, work-change signals and legacy Assignment migrations are
  **not** owned by this Factory specialist.
- `go.mod`, `go.sum`: `github.com/Microsoft/go-winio v0.6.2` for protected Windows
  local control.
- `README.md`: Foreman API link/startup paragraph and revision metadata. Other
  incumbent harness/version documentation edits are unrelated work.

### Explicitly excluded baseline/concurrent work

K root specs/plan, `AGENTS.md`, `product.md`, local `.klm/agents/*.toml`, all
`clients/**`, `AUTOMATIC_PAGINATION_FIX_PROGRESS.md`, `engine/sqlite_history.go`
and `engine/sqlite_message_history_test.go` are not this backend specialist's
implementation. In W, incumbent `factory_*`, `assignment_*`, `ticket_*`, graph,
automation, memory and their docs/tests, root/context docs, specialist TOMLs and
all `frontend/**` are baseline/concurrent work. The inventory above, not the whole
dirty `git status`, defines this specialist's review scope. No commit was made.

## 4. Prepared real backend/browser journey — not executed yet

The remaining narrow check uses **one representative harness, Pi 0.85.1**. It
uses actual production handlers and Foreman mode, with **normal KLM client auth
enabled**. It does not reuse the in-process fake peer or bypass from prior tests.

### Ownership, credentials and startup

Wait for the frontend specialist to finish and the parent to authorize the launch
window. Invoke the prepared script in its own foreground PowerShell process:

```powershell
powershell.exe -NoLogo -NoProfile -NonInteractive -File "C:\Users\Samuel\Documents\Projects\Personal\klm\engine\validation\Start-ForemanJourney.ps1" -RunRoot "C:\Users\Samuel\AppData\Local\Temp\opencode\factory-integrated-20261114" -BrowserOrigin "http://127.0.0.1:5174" -Minutes 15
```

The run root must not exist and must be a direct child of **T**. Use a new suffix
for an explicitly authorized retry; do not reuse interrupted source identities or
automatically repeat uncertain native work. The script refuses occupied backend
ports. It does not build, install or upgrade anything. Optional `-KlmBinary` and
`-WerbertBinary` accept separately rebuilt owned artifacts.

The parent/browser specialist must keep the launcher invocation owned and awaited;
allow at least the 15-minute window plus 100 seconds for bounded shutdown (tool
timeout 1,020,000 ms or longer). It returns when the deadline is reached or
`<RunRoot>/stop.request` exists. Keep that owner alive during browser work. Killing
the launcher skips PowerShell `finally`; that is not successful cleanup.

Mechanism:

1. Create a private current-user-only scratch root; no production project/data is
   copied. Create a disposable project and a complete saved `journey-worker.toml`
   **before** execution. Its prompt comes from the repository fixture Markdown.
   It has a real source UID, Pi/model/effort and empty skill/MCP selection. This is
   a saved Agent available through the actual catalog, not a synthesized worker
   peer or a modification of a personal specialist. It does not validate the
   user's existing specialist configurations or nonempty Library selections.
2. Copy `%USERPROFILE%/.pi/agent/auth.json` into `<RunRoot>/pi/auth.json`; never log
   bytes. Redirect Pi's native home with `PI_CODING_AGENT_DIR`. Original auth is
   hash-checked at cleanup and the disposable auth file is removed. Fresh HOME,
   USERPROFILE, APPDATA, LOCALAPPDATA and skills paths are child-only. Inherited
   KLM/Werbert/Pi/OpenCode/Codex/provider controls are cleared for this window.
3. Start the dev executable as owned `__worker`, using `KLM_DEV_DATA_DIR=<RunRoot>/klm`,
   `KLM_API_BIND_HOST=127.0.0.1`, fixed port 17331 and explicit pinned Pi binary.
   `KLM_DEV_DISABLE_CLIENT_AUTH` remains absent. Verify unauthenticated `/api/state`
   returns 401.
4. Discover/ready/pair use the same user-protected named pipe as normal KLM. Its
   name includes the user's SID and first eight SHA-256 bytes of the lowercase
   absolute data directory. The setup client redeems a code only in memory,
   registers the disposable project through authenticated `POST /api/projects`,
   then revokes itself. No code/credential is printed, persisted in browser data
   or embedded in project files.
5. Start actual Werbert with `WERBERT_MODE=foreman`,
   `WERBERT_LISTEN=127.0.0.1:18090`, `WERBERT_DATA_ROOT=<RunRoot>/werbert`,
   `WERBERT_KLM_PORT=17331`, `WERBERT_KLM_DATA_DIR=<RunRoot>/klm`, exact
   `WERBERT_ORIGINS=<BrowserOrigin>` and `WERBERT_HOSTS=127.0.0.1:18090`.
   Its explicit browser Connect performs its **own** protected discovery/pair/
   redemption/integration binding. Windows DPAPI protects
   `<RunRoot>/werbert/foreman-connection.protected`.
6. Emit credential-free `ready.json` only after both backends are ready. It names
   the actual project/engine IDs, saved-Agent hash/source, owned PIDs and deadline.
   The script submits no Chat input and performs no inference on its own.

The existing frontend Vite proxy reads `WERBERT_ENGINE_URL`. Frontend owner starts
its own preview, when ready, from **W/frontend** with:

```powershell
$env:WERBERT_ENGINE_URL = 'http://127.0.0.1:18090'
npm run dev -- --host 127.0.0.1 --port 5174
```

Visit `http://127.0.0.1:5174/foreman`. Vite's existing `changeOrigin: true` proxy
uses the allowed backend Host. If another browser port/origin is required, change
the launcher parameter to the exact matching origin. No frontend file is edited
by this preparation. The browser calls Werbert only.

### One bounded substantive task and actual human intervention

Browser specialist:

1. Read `ready.json`. In the UI, discover and explicitly Connect; confirm actual
   `connected=true`, matching engine identity and a retained protected connection.
   Select the named disposable project. Reopening selects the same Factory/Chat.
2. Choose Pi, `openai-codex/gpt-6-astra`, low effort in Settings, empty Library,
   then Save. Confirm saved/applied capture, read-only YOLO and no inference from
   Connect/selection/Settings reads. If the real model/catalog is unavailable,
   report it; do not choose a fallback or install content.
3. Submit this substantive Chat input once with one stable operation identity:

   > Create delivery-label.txt in this disposable project containing exactly my
   > chosen release label, Alpha or Beta, in UTF-8 without BOM or trailing newline.
   > I have not chosen the label yet. Open one Ticket and assign the existing
   > Journey Worker saved Agent. The worker must ask me the necessary label question
   > using native ask_user and wait for my actual answer. Neither Foreman nor the
   > worker may choose for me or infer a default. After my answer, write the file,
   > retain its exact path and SHA-256 as the result, then separately assess the
   > original objective. Do not change any other project file, delegate further,
   > use the network beyond the configured model provider, or commit/push.

4. Observe one source-bound Ticket, one saved-Agent Assignment and its one
   acknowledgement, actual native worker history and the exact **native** pending
   request. A `worker_report` prose question alone does not close native request
   validation. Capture request/worker/Assignment identities before answering.
5. The supervising agent must use the normal **question tool** to ask the actual
   human the offered Alpha/Beta choice, including enough context to identify this
   disposable task. Submit only that human's response to this exact native request
   through the UI/request endpoint. Browser automation must not choose an option
   merely to unblock the run. If no human answer is available, leave it unanswered,
   capture that fact, cancel/stop through the owned flow and report this path
   incomplete; do not claim result/assessment success.
6. After an actual answer, verify request delivery/resumption separately from its
   HTTP receipt. Inspect the file bytes and actual digest, retained result endpoint
   and read-only worker history. Confirm Completed only after a separate satisfied
   assessment with the settled current work revision and actual result references.
   One native turn ending or EOF from a result download is not success.

Keep this first journey narrow. Existing adapter/Stop/cancel checks do not need
repetition unless this run exposes a changed or failing path. The necessary human
choice is unresolved in the task; configuration and Library descriptions do not
supply that answer or authorize additional effects.

### Shutdown and retained evidence

Before ending the window, stop a live Foreman turn and explicitly confirm Cancel
for unfinished work, then observe cessation. Create `<RunRoot>/stop.request` to end
the launcher early. It terminates only its retained Werbert process, sends Stop
over its disposable KLM control pipe, waits up to 75 seconds for KLM exit, then
uses owned-process termination only if that wait fails. Forced termination is
recorded and must not be described as graceful native drainage. Werbert termination
on Windows is **not** a graceful `App.Close` acceptance check.

`cleanup.json` records observed backend exit, auth-copy removal, original auth hash
comparison, saved-Agent byte comparison and exceptional/forced shutdown. Inspect
it and owned native-process cleanup before claiming the window clean. Retained
SQLite/DPAPI files/logs stay private under the run root; never attach credential
storage, database/key pairs or raw auth to a report. Capture sanitized browser/API
evidence with actual IDs, request status, result digest, assessment basis and
outcome. `ready.json` alone means startup, not a completed journey.

## 5. Material requirements still unvalidated

- Full browser/real-backend journey and actual human answer delivery/resumption.
  The full launcher lifecycle is unvalidated; its protected LF control function
  has now passed the bounded standalone smoke in §7.
- Installed-engine protected discovery/pairing and install/uninstall/login/LAN.
  The prepared path exercises equivalent handlers on disposable dev storage,
  not the installed engine's lifecycle.
- Nonempty selected skill/MCP snapshots, upstream MCP timeout/elicitation variants
  and real catalog readiness permutations. MCP remains synchronous by design;
  five-second handoff applies to shell/job waits, not every tool/model request.
- Native compaction, including the requested initial policy, and verified idle
  continuation/configuration-generation history across restart lack dedicated new
  integrated evidence. Adapter reuse is not runtime acceptance.
- Crash/restart permutations for incoming/outgoing delivery, native input
  inclusion, pending callbacks, effects, protected replies and jobs; unknown
  outcomes must remain interrupted/unknown rather than replayed.
- All ordinary KLM projections/aggregates and full personal-path regression
  behavior. Focused guards/checks cover specific paths, not exhaustive acceptance.
- Linux runtime/containment and secret-storage acceptance; Windows compilation
  and local probes establish neither Linux execution nor same-user isolation.
- Final Ticket paging/list projection behavior is compile-only, as noted above.
- Human acceptance of specifications, mock deltas, interaction and visual fidelity.

No full suite, review agent, installer, upgrade, deployment, commit or push was
performed by this specialist. The API handoff remains **W/engine/docs/foreman-api.md**
and **W/engine/internal/engine/foreman_contract.go**, with KLM's technical contract
in **K/engine/RUNNER_API.md**. Code review after frontend completion should use the
attributable inventory in §3 and these explicit evidence limits.

## 6. Checks actually performed in this preparation batch

Working directory: **K**. The only executable validation of the new launcher was
this syntax-only parse; no launcher statement or server was executed:

```powershell
$tokens = $null; $errors = $null; $null = [System.Management.Automation.Language.Parser]::ParseFile("C:\Users\Samuel\Documents\Projects\Personal\klm\engine\validation\Start-ForemanJourney.ps1", [ref]$tokens, [ref]$errors); if ($errors.Count -gt 0) { $errors | Format-List Message, Extent; exit 1 }; "PASS: Start-ForemanJourney.ps1 parsed; no launcher statements executed."
```

Outcome: exit 0; stdout:
`PASS: Start-ForemanJourney.ps1 parsed; no launcher statements executed.`

Artifact identity lookup:

```powershell
Get-FileHash -LiteralPath "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-factory-backend-check.exe", "C:\Users\Samuel\AppData\Local\Temp\opencode\werbert-foreman-backend-check.exe" -Algorithm SHA256 | Format-List Path, Hash
```

Outcome: exit 0, the two hashes recorded in §1. Dedicated filesystem reads/globs
confirmed all seven step-0 `evidence.json` paths in §2. Git HEAD/status and focused
diff inspection supplied the attribution limits in §3. Documentation was read
back. No previously passing Go build/test/provider probe was repeated.

## 7. Reported-findings closure and browser retry artifacts

This follow-up was explicitly commissioned after the original reviewer and browser
specialist returned. Its scope is the four reported defects plus startup framing.
It is not a new review round or human acceptance. Read
`FACTORY_BROWSER_VALIDATION.md` for the first attempt: KLM listened, protected
discovery failed, cleanup required forced KLM termination, and **no Werbert/Vite,
native work or browser journey started**. That historical evidence is preserved.

### Finding-to-fix mapping

| Finding | Change / source location |
| --- | --- |
| P1: job completion stranded during an active turn or at worker settlement | `engine/runner_jobs.go:263,315–386`: terminal job state and retained attributed completion input now commit in one transaction before the live-job done barrier; wake the active steering consumer after commit. A delayed completion can reopen terminal native state only while actionable input is retained. `engine/runner_runtime.go:212–274`: worker settlement checks retained actionable inputs in the same transaction and remains schedulable rather than publishing `native_completed` first. |
| P1: cancellation after restart falsely claimed cessation without handles | `engine/runner_controls.go:298–379`: retain an existing uncertainty error when no live native turn exists; account for all admitted execution effects and check persisted `starting`/`running`/`unknown` job rows at final settlement. Missing handles do not erase uncertainty; unresolved work remains `interrupted`. Positively settled jobs can still reach `cancelled`. |
| P2: fresh native generation lost recent operational conversation | `engine/runner_context.go:10–62`, `runner_contract.go:21`, `runner_runtime.go:143–149,324–327`: persist at most 24 conversation previews within 32 KiB when rotating generation, preserving original session/event/ordinal and input source kind/ID. Read stored previews rather than full histories and supply them on native initialization as retained reference, without inserting new human inputs. Werbert `engine/internal/engine/foreman_overview.go:89–153` now supplies at most 24 / 32 KiB recent human and assistant/worker publication previews alongside work/request overview, preserving producer and source IDs. Its `engine/prompts/foreman.md` clarifies reference-versus-authorization semantics through the existing prompt loader. |
| P2: invalidation-only ordinary changes disappeared | `engine/runner_store.go:166–209`: keep ordinary invalidation-only transactions when no external session changed; in an external-only session write, preserve unrelated invalidation domains such as `harnesses` while stripping runner/session scopes. Regression includes disconnected/connected integration and mixed external-session plus harness invalidation. |
| Browser startup: control framing and swallowed diagnostics | `engine/validation/Start-ForemanJourney.ps1:63,71,122–146`: explicitly write LF; dispose reader/writer before closing the underlying pipe; retain sanitized readiness/protocol or pipe-failure diagnostics in `discovery-diagnostics.json`. No pairing code is included. |

The disposal fix was demonstrated by the new smoke, not guessed: after LF was
corrected, the first smoke failed on `StreamWriter.Dispose` because the underlying
pipe had already been disposed. Only disposal order changed after that failed
check. Werbert's existing `engine/internal/engine/foreman_connection.go:42` uses
Go `fmt.Fprintln`, which already writes LF, so its client needed no framing change.
KLM's LF-only protected protocol was retained.

New focused regression sources: K `engine/runner_regression_test.go`,
`engine/runner_control_windows_test.go`; W
`engine/internal/engine/foreman_overview_test.go`. No frontend, root plan/spec/TOML,
concurrent history-pagination implementation or incumbent Werbert executor file
was changed in this closure.

### Exact affected checks and outcomes

Working directory **K/engine**:

```powershell
go test -tags dev -run '^TestRunner(JobCompletionWakeAndSettlement|CancelRetainsPersistedUncertainty|GenerationConversation|OrdinaryInvalidationOnly|LauncherProtectedControlLF|DurableBoundary)$' -count=1 -timeout=60s -v .
```

This first combined invocation **exited 1**, total 10.958 s, solely because the
protected-control smoke exposed the writer-disposal exception. Its other selected
tests actually passed:

```text
--- PASS: TestRunnerDurableBoundary (0.90s)
--- PASS: TestRunnerJobCompletionWakeAndSettlement (1.73s)
--- PASS: TestRunnerCancelRetainsPersistedUncertainty (4.29s)
--- PASS: TestRunnerGenerationConversation (0.88s)
--- PASS: TestRunnerOrdinaryInvalidationOnly (1.71s)
```

Cases include completion during active work / after native settlement, persisted
unsettled job states without handles / retained native uncertainty / a positively
settled job, bounded generation context with original human provenance and no new
input authorization, and ordinary invalidations with/without integration binding
and with/without external session changes. These are disposable focused fixtures,
not real provider/restart acceptance. The existing visibility boundary check also
remained passing.

After correcting disposal, reran **only the failing smoke**, still in **K/engine**:

```powershell
go test -tags dev -run '^TestRunnerLauncherProtectedControlLF$' -count=1 -timeout=30s -v .
```

Outcome: exit 0, exact relevant output:

```text
PASS: protected discover, pair (redacted), stop using launcher LF; no application listener.
--- PASS: TestRunnerLauncherProtectedControlLF (1.23s)
PASS
ok   klm/engine  1.338s
```

The smoke parses the actual launcher and evaluates **only its Control function**,
against a disposable real `serveEngineControl` Windows named pipe. It observes
ready discovery identity, a nonprinted one-use pair response, and the owning Stop
signal. The pipe is closed and the disposable store is cleaned by the test. It
does not launch the application's HTTP server, redeem pairing over HTTP, access
native provider credentials, run a model or establish full startup success.

Working directory **W/engine**:

```powershell
go test ./internal/engine -run '^TestForemanRecentConversationProvenance$' -count=1 -timeout=30s -v
```

Outcome: exit 0:

```text
--- PASS: TestForemanRecentConversationProvenance (0.04s)
PASS
ok   werbert/engine/internal/engine  0.175s
```

This verifies bounded recent operational human/assistant exchange and original
producer/source attribution in the actual overview builder. The earlier business
test and final Ticket paging behavior were not rerun; paging remains compile-only.

Formatting used `gofmt -w` on the changed runner Go files/tests and the two Werbert
overview Go files, with exit 0/no stdout. No pinned native provider probe was
repeated: adapter transport/configuration mechanisms were unchanged in this fix.
The code checks and the one necessary protected-control smoke above were the
entire affected test scope. Results remain in the execution transcript and this
sanitized record; no separate raw Go-test log artifact was created.

### Both browser binaries rebuilt after fixes

Working directory **K/engine**, actual shell command:

```powershell
if (-not (Test-Path -LiteralPath "C:\Users\Samuel\AppData\Local\Temp\opencode" -PathType Container)) { throw 'Build output parent is unavailable.' }; go build -tags dev -o "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-factory-backend-check.exe" .; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; "PASS: KLM dev backend artifact rebuilt."
```

Outcome: exit 0; `PASS: KLM dev backend artifact rebuilt.`

Working directory **W/engine**, actual shell command:

```powershell
if (-not (Test-Path -LiteralPath "C:\Users\Samuel\AppData\Local\Temp\opencode" -PathType Container)) { throw 'Build output parent is unavailable.' }; go build -o "C:\Users\Samuel\AppData\Local\Temp\opencode\werbert-foreman-backend-check.exe" ./cmd/engine; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; "PASS: Werbert Foreman backend artifact rebuilt."
```

Outcome: exit 0; `PASS: Werbert Foreman backend artifact rebuilt.`

The same `Get-FileHash ... -Algorithm SHA256 | Format-List Path, Hash` command in
§6 then exited 0 and returned these artifacts under **T**. The Werbert artifact
was subsequently superseded by §8's targeted `ticket_read` correction:

| Browser artifact | SHA-256 after fixes |
| --- | --- |
| `klm-factory-backend-check.exe` | `EFC45463E73C785AE732222C0A7C1E5A70B49D0E3BB38267AD8DC873D98C9B6F` |
| `werbert-foreman-backend-check.exe` | `3198BCC4BD659B560AFDB473D3EBFA4779385E651DD6CBF8F3EACF8EA8815CFB` |

### Startup status and browser handoff

The corrected protected-control path is runtime-verified by the bounded smoke.
**Full launcher startup and the browser journey have not been retried here.** The
original browser evidence remains a failed startup, not a browser pass. Use the
§4 lifecycle command with a **new** direct-child run root, for example
`C:\Users\Samuel\AppData\Local\Temp\opencode\factory-integrated-20261114-b`.
The default binary paths now refer to the rebuilt artifacts above. Read
`discovery-diagnostics.json`, `ready.json` and `cleanup.json` for actual observed
startup/cleanup facts, and retain any new concrete failure rather than swallowing
it. The frontend specialist's proxy/origin and actual-human-question requirements
are unchanged. All §5 acceptance limits remain, apart from the explicitly verified
standalone protected-control exchange. No long-lived application server, full
suite, new reviewer, installer, migration/deployment, commit or push was launched
in this closure.

## 8. Runtime-discovered fresh `ticket_read` correction

The parent reported a successful later owned integrated run at
`C:\Users\Samuel\AppData\Local\Temp\opencode\factory-browser-20261114-g`:
actual native question → actual human Alpha response through the question tool
and UI → file containing exactly Alpha → retained result → separate satisfied
assessment, Completed at basis work revision `112`. Ticket:
`da8d723dd6c00a9a57e373fce417a450`; work:
`c4d1844725a350ac275e47b00fd64a81`; native request:
`6d51c08b964472e7346b01656d7d9cd0ae44854eaa2032f42e26ebce4545ba91`.
The parent also reported both backend exits, unchanged original auth/Agent bytes,
and original-reviewer closure of the four findings. These are parent-supplied run
facts, not a journey rerun by this specialist. Final browser/history/publication
evidence and acceptance remain with the parent.

That run retained an assessment reason explaining a workaround: `ticket_read`
returned `sql: no rows in result set`, so Foreman inspected the native history
directly. Source inspection confirmed the concrete cause in Werbert's
`engine/internal/engine/foreman_tools.go`: the initial absent operation receipt
was consumed as a fresh-call condition, but its `sql.ErrNoRows` remained in the
outer `err`. Successful `ticket_read` set `result` without overwriting that error,
then failed the common dispatch error check.

### Small correction and scope

- `engine/internal/engine/foreman_tools.go:59`: clear the consumed lookup sentinel
  immediately after confirming `sql.ErrNoRows`, before intent lookup/dispatch.
  Unexpected receipt lookup failures, Factory ownership checks, changed-body
  conflicts and immutable receipt replay remain intact.
- `engine/internal/engine/foreman_business_test.go:19`:
  `TestForemanFreshTicketRead` verifies the fresh owned Ticket/result, unchanged
  receipt on a repeated call after the stored Ticket changes, refusal of another
  Factory's Ticket and conflict on changed-target reuse of the same call identity.
- Inspection of the other dispatch branches was limited to this same stale-error
  mechanism: successful alternatives assign their own outer `err` before the
  common check. No additional branch fix was needed.

### Exact checks and new Werbert artifact

Working directory: **W/engine**. Formatting command exited 0 with no stdout:

```powershell
gofmt -w "internal/engine/foreman_tools.go" "internal/engine/foreman_business_test.go"
```

The single targeted behavior check:

```powershell
$env:TEMP = "C:\Users\Samuel\AppData\Local\Temp\opencode"; $env:TMP = $env:TEMP; go test ./internal/engine -run '^TestForemanFreshTicketRead$' -count=1 -timeout=30s -v
```

Outcome: exit 0, first and only invocation:

```text
=== RUN   TestForemanFreshTicketRead
--- PASS: TestForemanFreshTicketRead (0.05s)
PASS
ok   werbert/engine/internal/engine  0.171s
```

Rebuilt **Werbert only**, actual command:

```powershell
if (-not (Test-Path -LiteralPath "C:\Users\Samuel\AppData\Local\Temp\opencode" -PathType Container)) { throw 'Build output parent is unavailable.' }; go build -o "C:\Users\Samuel\AppData\Local\Temp\opencode\werbert-foreman-backend-check.exe" ./cmd/engine; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; "PASS: Werbert Foreman backend artifact rebuilt after ticket_read fix."
```

Outcome: exit 0; stdout:
`PASS: Werbert Foreman backend artifact rebuilt after ticket_read fix.`

```powershell
Get-FileHash -LiteralPath "C:\Users\Samuel\AppData\Local\Temp\opencode\werbert-foreman-backend-check.exe" -Algorithm SHA256 | Format-List Path, Hash
```

Outcome: exit 0. Current Werbert artifact SHA-256:
`6D1977A051BACC434B98B4B98BEFF05F49668FA307454D0DAA7C23082C236F13`.
This supersedes the Werbert hash in §7. KLM's source/artifact was not changed or
rebuilt by this correction. The test output is retained in this sanitized record
and the execution transcript, with no separate raw log file. No server, provider
probe, review, full suite, frontend edit, commit, push or deployment occurred.
The corrected fresh tool read has focused regression evidence; it was not rerun
through the real UI/provider in this correction batch.

## 9. Parent integration closure and final ordinary-count guard

The parent subsequently exercised the corrected Werbert artifact through the real
browser/Pi journey in disposable root `factory-browser-20261114-h`, including
actual human Alpha reply, worker resumption/file result, separate satisfied
assessment, visible final Chat publication and native read-only worker history.
[Browser evidence](FACTORY_BROWSER_VALIDATION.md#parent-integration-follow-through--real-answer-result-and-assessment)
owns identities/screenshots, successful behaviors, driver selector failure and
cleanup. That is actual corrected-path evidence, distinct from §8's focused test.

Parent-owned launcher corrections after §7 were: JSON media type on bodyless DELETE
mutations, sanitized HTTP status diagnostics with preservation of the original
setup exception, and empty valid project icon instead of the invalid literal
`folder`. These changes are attributable to the parent, not the backend specialist.

During final integration the parent observed that ordinary `/api/health` still
counted every active turn using `len(a.runs)`. `engine/api.go:health` now counts only
turns without a runner capture. Both SQLite and legacy response branches use that
ordinary count; personal turns retain their existing contribution. A small addition
to `engine/runner_boundary_test.go` checks a live-map fixture containing one
personal and one external turn returns `runningSessions:1`.

Actual command, working directory **K/engine**:

```powershell
go test -tags dev -run '^TestRunnerDurableBoundary$' -count=1 -timeout=45s .
```

Outcome: exit 0, `ok klm/engine 0.964s`. This compiles the final KLM source and checks
the count plus the existing visibility/receipt/history boundary. No provider run
was repeated for the health-only correction. The KLM browser binary hash in §7
therefore identifies the preceding browser cohort, not the final health-count
source. Rebuild before manual use of the final source; do not infer that the old
artifact contains the count correction. The original scoped review's four findings
are closed; these later direct integration corrections were not a second review.

Targeted `git diff --check` on the parent's changed KLM code/launcher returned 0.
Final listener inspection returned no validation listeners on 17331/18090/5174;
the `-h` cleanup JSON confirmed both backend exits, native auth-copy removal and
unchanged original auth/saved Agent. No production migration, installer/deployment,
commit or push occurred. Human acceptance and the completion report's remaining
manual checks stay separate.
