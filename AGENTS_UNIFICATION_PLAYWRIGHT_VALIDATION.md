# Agents unification — practical validation CP-V1

Publication note: this browser/native-adapter evidence predates stacked conflict
integration. The publication uses storage schema4, preserving predecessor
content2 → catalog3 → delegation4; integration checks and limits are separately
recorded in `AGENTS_UNIFICATION_PUBLICATION.md`. No post-integration browser rerun
or upstream-model/native-graph acceptance is implied.

Status: **core practical validation complete and manager-reviewed; four reproduced
fixes rechecked; CP-V2 passed graph selection, narrow layout and Focus toggles**.
The unexercised paths below remain explicit; this is not universal feature/provider/
platform acceptance.

The CP-V1 environment-loss details below are historical. The short recovery and
remaining checks are recorded in the CP-V2 addendum at the end of this report.

Validator: `fbdec52fedc91ea40576bb6b5c1f7a13` (OpenCode / GPT 6.1 Sol).
Manager: `a05369eaa5c9151b61e93fb38cb35c52`. Date: 2026-10-31.

## Environment and execution

Read AGENTS.md, product context and CP-R2; used CP-R1 only for local-provider setup.
Preserved the extensive pre-existing working-tree changes.

Initial preflight found `127.0.0.1:5173` occupied by unrelated PID `8580`, parent
`38792`, command `node node_modules/vite/bin/vite.js --host 127.0.0.1`.
The manager expressly released fixture frontend port **15173**, with **15174** as
one fallback. A bounded check found 15173/15174/API17331 free. Only 15173 was used;
no product port/config/script was changed and PID8580 was never stopped/reused.

Absolute disposable root used throughout:

```text
C:\Users\Samuel\AppData\Local\Temp\opencode\agents-cpv1-20261031
```

- Binary: `<root>\bin\klm-dev.exe`, built with `-tags dev`.
- `KLM_DEV_DATA_DIR=<root>\data`; confirmed `engine/cli.go:12-25` honors the
  absolute override for development builds. SQLite initialization and engine logs
  were actually produced under that root.
- `KLM_USER_SKILLS_DIR=<root>\skills`; one disposable `cpv1-helper` skill.
- Own `home`, `tmp`, `config`, `xdgdata`, `cache`, `pi`, `codex`, `project`,
  `logs`, `screenshots`; no real HOME/config/history copied.
- Filtered native environment retained only OS executable-discovery variables,
  then supplied disposable HOME/USERPROFILE/APPDATA/LOCALAPPDATA/XDG/TEMP values.
  No upstream credential/proxy variables were forwarded.
- `KLM_DEV_DISABLE_CLIENT_AUTH=1`, `KLM_API_BIND_HOST=127.0.0.1`, disposable
  `KLM_ALLOWED_ORIGINS=http://127.0.0.1:15173`.
- `VITE_ENGINE_URL=http://127.0.0.1:17331`; Vite CLI:
  `--host 127.0.0.1 --port 15173 --strictPort`.
- Real installed versions observed through the engine inventory: Pi **0.85.1**,
  OpenCode **1.18.35**, Codex **0.153.4**.
- Existing Playwright imported from
  `C:\Users\Samuel\AppData\Local\Temp\opencode\multi-engine-cp4\node_modules\playwright\index.mjs`;
  existing Chromium executable from
  `C:\Users\Samuel\AppData\Local\ms-playwright\chromium-1248\chrome-win64\chrome.exe`.
  No dependency/browser installation.

The provider was a real local Node HTTP server on an OS-assigned loopback port,
implementing deterministic OpenAI-compatible SSE completions/Responses. Pi and
OpenCode used fixture models; Codex used a real catalog identifier (`gpt-5.2`)
against the **local simulated provider**, not an upstream model. Initially using
`fixture` on Codex correctly rejected child preparation because it was absent from
that harness's model catalog; the fixture was corrected, not catalog validation.

**Evidence categories:** frontend + HTTP/SQLite backend were real; installed native
parent/child adapters were real; model decisions/content were deterministic local
simulation. No full-frontend endpoint mocks or inserted transcript events.

### Bounded command shape

All shell commands supplied finite tool timeouts. Preflight/process/path commands
used 10s; one direct Pi version diagnostic used 15s (observed 4.783s). Builds and
TypeScript used 120s.

```powershell
# engine working directory
go build -tags dev -o "<root>\bin\klm-dev.exe" .

# desktop working directory
.\node_modules\.bin\tsc.cmd --noEmit --pretty false --project tsconfig.json

# each browser cycle, tool timeout 115000ms
node "<root>\runtime.mjs" <scenario>-browser.mjs
```

The first two detached launch approaches (PowerShell Start-Process, then Node
detached/unref) printed startup/PID but left the tool awaiting the process tree;
the user interrupted both. Read-only checks confirmed their listeners/PIDs were
gone before another launch. They were replaced with **one bounded lifecycle per
command**: start local provider, CLI `start` (75s child budget), Vite owned PID/logs,
readiness (20s), Playwright scenario (75s child budget), browser close, **same
binary/environment/dataRoot CLI `stop`**, own Vite kill, provider close. A 100s
runtime watchdog supplemented the 115s tool timeout. Subsequent commands returned
without the launcher hang.

Scenario script names: `browser`, `catalog`, `project`, `specialist`, `overrides`,
`delegation`, `reuse-stop`, `question`, `opencode`, `codex`, `source`, `graph`,
`layout`; a bounded inventory-only diagnostic also ran. Fixture selector mistakes
(ambiguous buttons, wrong request field/method) were corrected in temporary
scripts. They are not product bugs. Some cold starts hit navigation timeouts or
the installed Pi's 5s version-probe limit; no installation was attempted.

## Results actually observed

| Scenario | Result / limit |
| --- | --- |
| Global agent without project | Created through Settings → Agents using real Pi model discovery and selected temporary Library skill. Edited description; disable/enable retained the skill. |
| Temporary project/local agent | Registered an existing temporary directory through the UI; created a project-local specialist through project Agents. |
| New session Default | Created through UI; no specialist or inherited Library selection. |
| New session specialist | Global specialist selected through picker; response explicitly had `overriddenFields: []`. |
| Independent overrides | After fix below: UI PATCH200 saved only effort/prompt/skills; reset effort retained prompt/skills; Reset overrides restored empty override set. |
| Source next-turn refresh | Edited global prompt through UI, sent next native Pi turn and asserted its actual provider request contained `CPV1 updated source marker.` |
| Fixed harness recovery | Changed source to Codex/GPT-5.2; existing Pi session stayed Pi and showed incompatible-model recovery. Chose compatible Pi model/effort overrides, saved, and verified send re-enabled with exact composer draft retained. |
| Pi two children | Real root wrapper/private bridge/two real child adapters, completed2/2 and delivered result; provider saw exactly two leaf requests and two parent requests. |
| History/Output | Opened actual child history read-only (no textbox), actual digest-bound Output and Escape. UI showed children inside batch, not as top-level session rows. |
| Reuse | Reused an exact returned child ID through native tool invocation; card showed Reused #2 and child inventory count did not grow. |
| Stop | UI Stop cancelled a two-child batch while preparation was still pending; backend returned terminal2/2, cancelled delivery/outcome. This is **not** proof of cancelling model/tool work already running in both children. |
| Child question/YOLO | Native Pi `ask_user` projected to owner, remained pending for an explicit answer under YOLO. UI answered Blue to the exact `request.id` via the owner's route (200); child and parent then finished. |
| OpenCode two children | Initially reproduced errors below; after both Go fixes, actual native parent + two native children + continuation passed with terminal2/2/completed. |
| New Codex two children | Actual instrumented new thread/dynamic tool + two real native child adapters + consolidated continuation passed with terminal2/2/completed. No legacy-thread behavior relabelled or changed. |
| Graph Save/global agent | Real graph editor created a disposable single-agent graph referencing global `cpv1-global`; real POST Save200 returned the persisted definition/revision. No graph invoked. |
| Stream stability | Graph-view ready/catalog refresh repeatedly reopened identical SSE scope sets before fix. After fix, 2s stable-window assertion allowed at most two new streams and passed. |
| Graph selection / narrow / Focus toggles | Attempted next, but did **not** complete. Last browser cycle hit `ERR_INSUFFICIENT_RESOURCES`/connection failures before graph selection, then the temporary root disappeared. Wide floating browser screenshots were observed; maximize/restore and narrow interactions were not reached. |

No actual permission/MCP elicitation scenario ran; the completed owner-reply
scenario was the supported native Pi question path.

## Four reproduced fixes (only validator code changes)

1. **Session settings Save navigated/reloaded and discarded edits.**
   The model/effort entry rendered SessionAgentDialog's form inside the composer
   form. Browser warnings showed invalid nested forms; clicking Save issued a
   navigation to `/?` and no PATCH. In
   `clients/desktop/src/features/agents/SessionAgentDialog.tsx`, render the dialog
   through `createPortal(..., document.body)` and stop submit propagation.
   Recheck: actual PATCH200 for independent save/reset; no page navigation,
   dialog stays mounted; later fixed-harness recovery preserved composer draft.

2. **Windows process finality falsely failed during concurrent shutdown.**
   OpenCode children produced `TerminateProcess: Access is denied`; an earlier
   fixture teardown during a pending Pi question exposed the same condition.
   In `engine/platform_windows.go`, after ACCESS_DENIED, wait up to5s on the
   **already retained owned process handle** for confirmed exit rather than
   immediately rechecking with zero timeout. Failure remains failure if not
   signaled; no PID inference or containment weakening. The next actual OpenCode
   recheck no longer returned ACCESS_DENIED, exposing the separate bug below.

3. **Successful managed OpenCode children always retained false drain state.**
   Adapter settlement intentionally leaves its server alive (`processesDrained`
   false), but the mandatory child shutdown combined successful drainage with
   that earlier false value. In `engine/harness.go`, the actual mandatory
   `runtime.shutdown()` result establishes the child drain barrier, while
   preserving `processDrainFailed` and all existing error/finality checks.
   Combined recheck: native OpenCode batch2/2 completed/delivered and parent
   continued; subsequent new Codex batch also passed.

4. **Same-scope SSE reopen/catalog-ready feedback loop in graph view.**
   The browser logged over100 repeated aborted identical `/api/updates` streams
   during a short graph interaction. In
   `clients/desktop/src/features/chat/scopedConnection.ts`, retain the existing
   stream when the final sorted scope set is unchanged after batched effect
   cleanup/resubscription. A real error still clears the source and retries.
   Recheck: real graph POST Save200 and the bounded stream-stability assertion
   passed. The later layout resource/network failure remains unresolved and is
   **not** claimed fixed by this change.

Checks: initial Go dev build; rebuild after each of the two relevant Go fixes;
TypeScript after portal fix, then after scoped-stream fix. All returned exit0.
No full suite, integration suite, adversarial review or new test suite was run.

## Evidence paths and disappearance blocker

Scripts originally wrote `<root>\logs\checks.jsonl`, `provider.jsonl`, per-request
JSON, `runtime.json`, `cleanup.json`, `data\engine.log`, Vite logs, scenario
`*-trace.zip`, API result JSON (`pi-batch`, `reuse-stop-batches`, `question-pending`,
`opencode-batches`, `codex-batches`, `source-turn-request`, `graph-saved`), and
screenshots01–19. Traces were snapshots/screenshots of the real served frontend.
Screenshots15 (completed OpenCode batch) and13 (Pi owner question under YOLO) were
also opened with the image tool and are visible in this conversation's tool
history. Their wide browser layouts showed readable batch/actions/question UI.

**Those filesystem paths are no longer available.** After the last layout attempt,
Read of the approved temp parent returned only `finance-agent-restart.log`.
Explicit 10s Test-Path checks returned **False** for both:

```text
<root>
C:\Users\Samuel\AppData\Local\Temp\opencode\multi-engine-cp4\node_modules\playwright\index.mjs
```

The validator did not issue any directory/file deletion of the temporary root or
existing Playwright dependency. Cause is unknown; do not assume the manager or
another process caused it. Do not treat missing traces as retained evidence.
Observed runtime/tool outputs and two image attachments remain in the conversation;
this repository report records them with that retention limitation.

## Cleanup / remaining gate

Normal bounded cycles closed the browser, used the same isolated CLI/env for
`stop`, terminated only their Vite PID and closed their local provider. Example
early owned PIDs were runtime46956/Vite17224 and runtime51772/Vite20352 (the
interrupted launch attempts); each cycle later recorded fresh owned IDs.
Final bounded listener check returned **no listener on17331 or15173** and still
returned unrelated5173/PID8580. No cleanup command targets the installed engine.
The now-absent binary/dataRoot cannot be used for another CLI stop.

Remaining practical gate:

- Restore an approved disposable environment and existing Playwright dependency
  availability, then finish graph selection, narrow viewport and Focus toggles;
  investigate the final resource/network failure if reproducible.
- Delete/duplicate-error recovery UI was planned for that last narrow scenario
  but was **not reached**. Global create/read/update/toggle and local create did run.
- Native Tauri desktop, Flutter/mobile host, side/general combinations, saved-agent
  top-level spawn, actual MCP permission/elicitation, long-output pagination and
  already-running-child Stop were not exercised here.
- Graph Activity child read and native graph Choice/seal/drain remain unexercised.
  No graph invocation or fabricated human authorization; CP-R2 synthetic fixture
  evidence was not repeated or relabelled as native finality.
- No upstream live model/provider validation or production installation.

No independent sessions/subagents were created in this real project by the
validator; UI sessions/owned children existed only in the disposable engine.
No commit/push. Stop after manager report pending explicit review/revalidation.

## CP-V2 — short clean-browser follow-up

Manager released only graph selection, narrow viewport/Focus toggles, and checking
whether the last resource/network failure recurred. **All four checks passed in
one bounded start/readiness/browser/stop command.** No new product code changes,
CRUD/binding/delegation repeats, model turns, graph execution or extra suites.

### Restored dependency and disposable environment

Workspace search found no installed Playwright package. Ports17331/15173 were
free; approved temporary parent and existing Chromium1248 executable existed.
Used a new absolute root:

```text
C:\Users\Samuel\AppData\Local\Temp\opencode\agents-cpv2-20261031
```

Under the manager's explicit dependency exception:

```powershell
# 20s tool timeout / 15s fetch timeout / no retries
npm.cmd view playwright version --fetch-retries=0 --fetch-timeout=15000 --cache "<cpv2-root>\npm-cache"
# Returned 1.64.1.

# 60s tool timeout; PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
npm.cmd install --prefix "<cpv2-root>" --ignore-scripts --no-audit --no-fund --save-exact playwright@1.64.1 --fetch-retries=0 --fetch-timeout=20000 --cache "<cpv2-root>\npm-cache"
# Exit0: added2packages in5s (Playwright and its core dependency).
```

Only temporary package.json/lock/node_modules/cache were written. No project
dependency files or npm global installation changed, and no browser was downloaded.
Used the existing Chromium via explicit executablePath:
`C:\Users\Samuel\AppData\Local\ms-playwright\chromium-1248\chrome-win64\chrome.exe`.

A Go dev build to the new root's `bin\klm-dev.exe` completed exit0 (120s bound),
needed because the previous binary no longer existed. The lifecycle retained the
same isolation rules: own absolute data/skills/HOME/config/native-history/temp
directories, filtered credential-free environment, API17331/dev-only auth bypass
bound to loopback, origin15173 and Vite strictPort15173. No local model provider
or model task was needed for these checks.

Normal API calls created one temporary project, one empty OpenCode session and
one saved terminal-node graph used **only for selection**; its `echo CPV2` command
was never executed. Real fixture IDs:

- Project `d36281dbbc498f70d4c866a2d78c0f31`.
- Session `244b019e52abc02d0ef801152dc7cddf`.
- Graph `cpv2-select-only`.

`node "<cpv2-root>\runtime.mjs"` used a115s tool bound,100s watchdog,
60s CLI child budget,20s frontend readiness and75s browser child budget.
It returned normally, running cleanup in finally. Recorded runtime PID41384 and
owned Vite PID31828; the engine was stopped through the same binary/env/dataRoot.

### Actual browser results

1. **Graph selection PASS.** Clicked Choose graph → CPV2 Select Only in the real
   composer. The UI displayed the selected graph. A real GET returned
   `selectedGraphId: "cpv2-select-only"`, `run: null`, revision9.
2. **Focus toggles PASS.** Clicked Maximize window then Restore window; asserted
   the actual shell's `is-maximized` class was added then removed, capturing both.
3. **Narrow viewport PASS.** At390×844 the real composer remained visible and
   `document.documentElement.scrollWidth <= innerWidth`. Opened Session options,
   observed the YOLO switch, then Escape closed it. No YOLO setting was changed.
4. **Clean browser network/resource observation PASS.** No pageerror, HTTP>=400
   response or non-cancellation request failure was observed. In particular,
   `ERR_INSUFFICIENT_RESOURCES` did not recur. Three `/api/updates` requests were
   recorded over the whole page/scope transition sequence, with at most two new
   streams allowed in the final2s stable window. Ordinary cancellation of removed
   scopes was excluded from network-error assertions.

Results were emitted directly in the command output and written immediately to
this repository report after the scenario returned. Temporary files at observation
time: `<cpv2-root>\logs\results.json`, `lifecycle.jsonl`, Vite logs, fixture/runtime
JSON and screenshots01–06. No trace was required for this short follow-up.
The image tool opened `04-narrow-chat.png` and `03-focus-restored.png`, attaching
both real screenshots to this conversation. The narrow composer/controls remained
within the viewport; the wide restored shell retained its floating gradient layout.
The model picker was still loading in this deliberately provider-free fixture;
these checks do not assert model-discovery readiness. Temporary paths are not a
promise of indefinite retention.

Cleanup evidence: lifecycle log records CLI start at05:09:51Z, same-env CLI stop
(`KLM engine stopped.`) at05:10:13Z and owned-Vite termination. A final10s bounded
check found no listeners on17331/15173 (also none on5173 at that observation) and
no remaining PID41384/31828. No command stopped PID8580 or targeted an unrelated
runtime. No production engine/data changes, commits or pushes.

The earlier resource failure remains historically unexplained; a successful short
clean-browser follow-up establishes only that it did not recur in these checks.
All other CP-V1 limits remain explicit: native Tauri/Flutter, graph Activity/native
Choice finality, remote MCP paths and unexercised combinations are not acceptance
claims. The narrowly released CP-V2 work is finished; stop after one manager report.

## Manager closure

Reviewed the four fixes in source and ran only affected checks:
SessionAgentDialog.test.tsx (5 passed, 10.62s) and TestDelegationContracts (exit0,
0.198s). CP-V2 introduced no new product changes and its four practical checks
were accepted without redundant test runs. All implementation/validation sessions
are stopped. The code is ready for the user's manual validation; no production
engine update, commit or push was performed. Earlier missing temporary artifacts
remain missing evidence, not recovered by the later successful browser checks.
