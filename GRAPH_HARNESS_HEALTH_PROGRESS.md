# Graph harness health — 2026-10-27

## Final runtime result

After the user restored the executable, **three consecutive graph runs completed
successfully across Pi, OpenCode and Codex**, including native continuation in
every branch and all five node types. The final development health check returned
HTTP 200 / `status: ok`, zero active graph runs, zero running sessions, all three
harnesses available, and no pending requests.

| Case | Run ID | UTC interval | Independent checks | Durable assessment |
| --- | --- | --- | ---: | --- |
| long-12 | `b9e68c3488edeaca044165f96b5255c3` | 15:58:16–15:59:40 | 3,601 | satisfied / succeeded |
| long-13 | `2460489fc11265bb588fd73c39fcde50` | 16:01:41–16:02:55 | 3,601 | satisfied / succeeded |
| long-14 | `50af00f3e471a5ff5600b3b682013cff` | 16:03:25–16:04:41 | 3,601 | satisfied / succeeded |

Total: **10,803 independent assertions**, **21 accepted Agent/Join activations**,
zero Choice rejections, zero reported harness failures, no hung graph, and empty
finality-error diagnostics in the inspected run results. Graph completion was
not accepted on its own: actual output, all node occurrences and durable positive
objective assessments were inspected through Playwright.

**Tool-event transparency:** `long-12` contains one expected silent negative
existence probe, recorded separately. In `long-14`, Pi submitted `ls runs/long-14
&&` with an incomplete shell suffix, received the syntax error and recovered.
The graph and all checks still completed. Therefore the extra stretch target of
three runs with **zero recoverable model/tool errors** was **not achieved** and is
not claimed. The evidenced result is graph/harness runtime stability, including
normal propagation and recovery of that model command error.

The disposable project's `evidence/final.json` retains the compact final result.
Detailed per-activation captures, acceptance records, objective-assessment receipts
and the final health snapshot are under
`C:\Users\Samuel\AppData\Local\Temp\opencode\health-*`. The engine's own development
history retains the actual runs. The runtime was left running for manual inspection.

## Requested acceptance

Exercise real graph execution through Playwright at `http://localhost:5173`.
Use all available node types and all three installed harnesses. Correct runtime
failures, stop/restart the development engine after source corrections, and obtain
three consecutive healthy graph runs, including larger native contexts.

The user explicitly authorized simulated human interaction in the development
application. This is a disposable runtime test, not production conversation input.

## Environment and workload

- Development API: `17331`; installed engine on `7331` is separate.
- Project: `C:\Users\Samuel\Documents\Personal\klm-graph-health`.
- Development project ID: `2600d3c57c7e6d38f2b97e065def1668`.
- Test conversation: `ed1e160816408baee332b8ccc0b16aee` (`Teste graph`).
- Graph: `harness-health` / **Harness Health**.
- Pi `0.85.1`: `openai-codex/gpt-6-luna`, low.
- OpenCode `1.18.35`: `openai/gpt-6-luna`, low.
- Codex `0.153.4`: `gpt-5.6-luna`, low, explicitly approved by the user after the
  native `model/list` catalog (including hidden models) did not offer GPT-6 Luna.
- Terminal preparation → Fork → three parallel Agent branches → Choice-based
  native continuation in each branch → Join agent → terminal success Choice.
- Branches implement separate numeric, text and ledger modules. The Join runs
  `node verify.mjs <caseId>` against independently generated expected results.
- Quick cases use 37 assertions. Long cases add many complete reference documents;
  actual context usage must be captured, not inferred from generated file size.
- Fork branches intentionally share the disposable directory and own disjoint files.
  This does not test isolated Git worktrees or commits.

## First observed run: failed

`quick-01`, run `4538072fd93e80422e0cb03102b172e8`, activity
`dd0d996890d185e2bc5bf62a8d3e58fa`.

- Run started `2026-10-27T06:45:58Z`, ended failed `06:46:27Z`.
- The calling Pi conversation received `fetch failed`, although the graph was
  accepted and all three native branches started.
- Pi submitted a valid `pi-verify` reservation after correcting one malformed
  Choice payload. Its expected post-Choice interrupt became `stopReason: error`,
  `errorMessage: This operation was aborted` in the actual native transcript.
- That expected shutdown was classified as a harness failure, interrupting the
  other two branches before Join. No success is claimed for this attempt.

## Implemented source corrections

1. `engine/linked_bridge.go`: graph invocation/retry previously inherited the
   bridge's 12-second HTTP write deadline despite synchronous multi-harness
   catalog discovery. Give these calls a bounded deadline covering three catalog
   discoveries and keep native client timeouts longer than that response budget.
2. `engine/pi-permissions.ts`, `engine/pi_interactive.go`, `engine/harness.go`:
   acknowledge the owned post-Choice abort before dispatching it. Normalize the
   observed exact transport-abort error only with that acknowledgment, a sealed
   Choice and a settled Choice tool. Genuine provider failures still fail.
3. Disposable agent prompts clarify run-relative paths, the nested Choice payload
   schema, and the absence of a Git repository.

Lightweight check: `go build -tags dev` succeeded. The development engine was
stopped with its dev CLI and restarted with the newly built binary, preserving
the existing explicit development-auth bypass. A corrective `quick-02` attempt
was requested with the explicit fresh/original test-workspace decision.

## Current acceptance

The three-run runtime result is verified in the final table above. The stricter
additional zero-recoverable-tool-error target was not achieved; its exact remaining
model command error is explicitly reported rather than hidden or relabeled.

## Subsequent observations

- `quick-02`, run `93320ed545d5855c2821b859abe6b539`, completed at
  `2026-10-27T06:59:59Z`. All three harnesses completed both activations in the
  same native session per branch, and Join passed all 37 assertions. The Pi
  expected-abort correction worked. Join corrected one malformed Choice payload,
  so this is useful correction validation but is not counted as a clean run.
- The low-effort orchestrator initially mistyped an activity ID, then mistook an
  asynchronous retry receipt (`started:false`) for a missing start and created
  a duplicate activity. The duplicate was explicitly abandoned before execution.
  `engine/prompts/orchestrator.md` now clarifies that retry already schedules the
  activity and must not be followed by a second invocation.
- `long-01`, run `0090542452a47fc4e471c4021b81bc3c`, ended **blocked**, not hung:
  OpenCode declined a workload requiring 44 context-file reads as infeasible.
  The other branches were correctly interrupted and the run reached a terminal
  outcome. No harness crash was observed in this attempt.
- The stress workload was adjusted to put a large deterministic reference
  corpus directly into each branch agent's native prompt, while retaining a few
  bounded on-disk reference reads and all 3,601 long-mode acceptance cases. This
  guarantees substantial initial context without depending on a model voluntarily
  reading 44 files. Actual native context observations are recorded below.

## Large-context continuation evidence and workload correction

`long-02`, run `498404cb88d7bbd404b4916de2132f7f`, ran from
`2026-10-27T07:10:35Z` through `07:13:05Z`. All three harness branches completed
their build and verification activations in the same native session per branch,
with zero branch Choice violations. The Join received all three deliveries and
terminated **blocked** because the Pi branch wrote its numeric module/report at
the project root instead of `runs/long-02`. The independent integration check
correctly failed with `ERR_MODULE_NOT_FOUND`. This is not counted as a clean run.

Actual retained native usage after this attempt:

| Harness | Context observation | Window | Notes |
| --- | ---: | ---: | --- |
| Pi | 126,162 tokens | 272,000 | Engine-retained native context; transcript peak model input 126,046 |
| Codex | 156,490 tokens | 258,400 | Engine-retained native context |
| OpenCode | Unavailable | Unavailable | 861,344 cumulative input tokens; not a current-context measurement |

The graph's three branch prompts contain a 123,708-character deterministic corpus
as actual initial input. Runtime evidence therefore includes substantial real
native context, not merely large files left unread on disk.

The workload correction prepared at that stopping point was:

- Critical run-relative path instructions are repeated after the large corpus.
- Build must check the exact target module path before reserving its Choice.
- `verify-module.mjs` validates that exact path and every assigned fixture,
  replacing fragile inline shell-JavaScript checks.
- Preparation writes the actual `task.json` and retains 3,601 acceptance cases.
- Those branch prompts were 126,695–127,286 bytes, within the 128 KiB limit.
- Initially pending corrective attempt: `long-03`, same activity
  `55df0653916e97afee0fd6460b62026c`, explicit fresh/original test-workspace decision.
  The browser send failed before submission because the engine was already down.

## External runtime blocker

Windows Defender Operational events **1116/1117** show:

- Detection/remediation: **2026-10-27 04:15:11–04:15:14 local** (`07:15Z`).
- Executable: `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-health-dev-2.exe`.
- Detection: `Trojan:Win32/Bearfoos.A!ml`.
- Action: quarantine; Defender terminated the running worker (PID 43788).
- The executable was absent after quarantine. No corresponding Go panic or crash record
  was found in `engine.log` or the recent Application error events.

A proposed independent-console relaunch failed because its executable had already
been quarantined. Its idle PowerShell process was closed. The cause was therefore
identified as Defender remediation, not established as a terminal-lifetime bug.
No Defender setting, exclusion or quarantine entry was changed, and no alternative
binary was launched to evade this detection. The development API was down at that
stopping point; the later user-authorized restoration is recorded below.

Both source builds (`go build -tags dev`) succeeded before quarantine. The
installed engine and unrelated pre-existing working-tree changes were preserved.
Execution was paused for review/restoration before the user authorized resumption,
as recorded below.

## Resumed validation after user restoration

The user explicitly reported restoring the executable and requested another
attempt. The same `klm-health-dev-2.exe` was present and its normal dev CLI start
succeeded. The development API returned on port 17331; no elevation, Defender
setting change or executable substitution was used.

- `long-03`, run `f1566deb5a5f2909d4735523f145dd02`, completed all 3,601 checks
  with accepted Choices in all seven agent activations. OpenCode attempted to
  read a missing `product.md`; that single fixture-read error excludes this run
  from the stricter clean sequence. Added the disposable project's product scope
  document rather than weakening the acceptance check.
- **Clean run 1:** `long-04`, run `75b1aff9236dca64b4c4a1ba7abdf2a4`,
  activity `e6547c4a8a0d4de6283149f18ad34538`. Started `15:27:35Z`, ended
  `15:29:14Z` on 2026-10-27. All three branches completed build and native
  continuation, all seven Agent/Join activations had accepted Choices with **zero
  error events and zero Choice rejections**, and Join passed **3,601 checks**.
  Detailed Playwright evidence is saved in the temporary test-artifact directory
  as `health-acceptance-75b1aff9236dca64b4c4a1ba7abdf2a4.json` and one full activity
  JSON per node occurrence. The live graph canvas and Pi activity panel were also
  exercised and captured through Playwright.
- `long-05`, run `ae9e7d86ebfb1a4254cbcc63211b7410`, also completed all 3,601
  checks and all native continuations, with zero Choice rejections. OpenCode made
  one malformed `apply_patch` call and recovered. This is normal tool feedback,
  not a graph/harness failure, but the additional zero-error-event acceptance
  check excludes it from the strict clean sequence. Added the native patch-format
  reminder to that disposable agent's prompt and retained the large corpus.
- **Clean attempt:** `long-06`, run `04a4b60c74ca25417a73240ba41629da`,
  activity `e0da1519bef73aa4ccb3b2e53002877c`. Started `15:35:12Z`, ended
  `15:37:56Z`. All seven agent activations completed with accepted Choices,
  **zero error events and zero Choice rejections**. Join passed all **3,601**
  acceptance cases without corrections. Playwright's per-activation evidence:
  `health-acceptance-04a4b60c74ca25417a73240ba41629da.json`.
- `long-07`, run `83d3a67b9ebcc4f330af5f1ad3e6ecea`, completed 3,601 checks
  with no error events or rejected Choices. `long-08` completed but included a
  recovered Pi call sending a shell command to `write`; the fixture now repeats
  native tool argument shapes after its large context corpus.
- `long-09` and `long-10` completed with no error events or rejected Choices.
  `long-11` completed but Codex guessed `context-1.txt` instead of `context-01.txt`.
  The fixture now enumerates exact filenames in `context-index.json.files` and
  repeats the zero-padding convention in the branch prompt.
- These recovered model/tool mistakes did not crash or hang the harnesses. They
  were nevertheless excluded from the stricter final sequence.

### Expected absence checks

The external acceptance collector initially treated every nonzero native command
exit as an incident. In `long-12`, Codex deliberately used silent `Get-Item` checks
for not-yet-created run-local artifacts: exit 1, no diagnostic/output, then normal
file creation and all acceptance checks passing. This is an expected negative
existence probe, not a harness or implementation failure. The collector preserves
it separately as `expectedProbes`, accepting only the exact silent run-local
artifact-check shape with empty output and exit 1. It still rejects actual command
diagnostics, tool validation errors, rejected Choices and infrastructure errors.
