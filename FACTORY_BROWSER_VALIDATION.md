# Factory browser validation — integrated delivery

**Latest outcome:** the parent completed the real Pi-backed human-answer → file
result → separate satisfied assessment journey on November 15, 2026 UTC (November
14 local). Final Chat publication and actual read-only worker history were visible.
The last driver exited 1 on a hidden-option selector after those behaviors had
passed; that driver is not claimed as an overall passing command. See the final
section for exact evidence. Earlier blocked attempts below remain historical.

2026-11-14, Windows. One owned disposable startup attempt; no real browser pass
was reached. `/dev/foreman` was not used. Human acceptance remains pending.

## Blocker

With validation ports 17331, 18090 and 5174 free, the Node supervisor at
`%LOCALAPPDATA%/Temp/opencode/foreman-browser-owned.cjs` launched
`engine/validation/Start-ForemanJourney.ps1` with a new disposable root
`%LOCALAPPDATA%/Temp/opencode/factory-browser-20261114-a`. KLM logged that it was
listening on `127.0.0.1:17331`, but `ready.json` was never produced. The
supervisor's 80-second readiness deadline expired; the launcher later reported
`Disposable KLM protected discovery did not become compatible/ready.` at line
131. Werbert and Vite were never started. No Connect, project registration,
native inference, browser request or human question occurred.

Expected: protected discovery returns a compatible ready identity, allowing the
launcher to register its disposable project and start Werbert. Actual: discovery
did not complete; its per-attempt errors are swallowed by the launcher, so the
received pipe response and exact runtime cause were not captured.

**Source-based lead, not runtime-confirmed:** `Control` in the launcher uses
`StreamWriter.WriteLine`, which on Windows sends CRLF; `engine/cli_windows.go`
matches `discover\n` and `stop\n` exactly, without stripping CR. This is a
plausible cause of both the failed discovery and failed graceful Stop. The
implementation owner should verify and correct the launcher/protocol as needed;
do not regard this hypothesis as a captured pipe response.

## Cleanup and remaining validation

The launcher exited with code 1 after its `finally` completed. Its retained
`cleanup.json` records `klm_exit_observed: true`, `forced_klm_termination: true`,
`control_stop_error: Owned control stop unavailable.`, `native_auth_removed: true`,
`original_auth_unchanged: true`, and `saved_agent_unchanged: true`.
`werbert_exit_observed: false` reflects that Werbert was never started. The
launcher and owned KLM process were observed exited; subsequent process and
listener inspection found no validation backend and no listeners on the three
validation ports. The pre-existing 7331/7332 listeners were left untouched.
Forced termination is not evidence of graceful native drainage. No worker had
been admitted. Private logs and cleanup evidence remain in the disposable root;
do not attach its database or credential storage.

All requested real-browser scenarios remain untested: connection and identity,
settings/editor/Library, responsive and theme visuals, Chat, Ticket and saved
Agent, native question and actual human reply, Stop/cancel ownership, worker
history, retained result and separate assessment. No application HTTP failure,
DTO, console error or screenshot exists from this attempt. After the owner fixes
the blocker, perform the authorized bounded journey with a fresh disposable root
and obtain the actual human answer only if the real native question appears.

## Corrected-launcher retry — blocked during setup client revocation

The explicitly authorized retry used the rebuilt binaries whose SHA-256 values
match `FACTORY_BACKEND_EVIDENCE.md` §7, and the fresh disposable root
`%LOCALAPPDATA%/Temp/opencode/factory-browser-20261114-b`. Ports 17331, 18090
and 5174 were free. Protected discovery now succeeded: the sanitized
`discovery-diagnostics.json` records 14 attempts, LF framing, `ready=True`, API
port 17331 and protocol 1. This is a successful recheck of the earlier discovery
blocker, not evidence of a working browser journey.

The launcher then exited before `ready.json`: its recorded error was
`Disposable KLM DELETE /api/auth/devices/67ea6ccfe97f3bd5a374514566a8258c failed.`
This was its short-lived setup client's self-revocation, in the `finally` at
`Start-ForemanJourney.ps1:167`. Expected: setup succeeds, the temporary client
is revoked and Werbert starts. Actual: the DELETE raised an exception; the
launcher suppresses the HTTP status/body, and a failure inside that `finally`
can also mask a preceding setup error. Neither the actual response nor success
of the preceding Pi-readiness/project-registration steps was established. No
claim is made that the revoke itself failed to persist. Pair redemption had
returned a client ID. Werbert, Vite and the browser were never started, and no
model task or human question was submitted.

The bounded Node supervisor's creation command printed PID 38472 but its shell
tool timed out; the supervisor itself completed. Its `final.json` records
launcher PID 13552 exited 1, no supervisor cleanup errors, and the launcher's
`cleanup.json` records KLM exit observed, auth copy removed, original auth hash
unchanged and saved Agent unchanged. Werbert was never launched. Subsequent
process inspection found neither PID nor the validation executables; all three
validation ports were free and the existing 7331/7332 listeners remained. Unlike
the first attempt, no forced KLM termination or control-stop error was recorded.
Sanitized supervisor artifacts are under `%LOCALAPPDATA%/Temp/opencode/
foreman-browser-control-b`; backend logs and data remain private in the run root.

The owner should expose sanitized status/error diagnostics for the setup sequence,
including any original error masked by revocation, and correct the actual blocker
before another authorized fresh-root attempt. All real-browser scenarios and
human acceptance listed above remain untested.

## Third startup attempt — invalid project icon

The parent explicitly authorized a fresh affected-path retry using the corrected
launcher and `%LOCALAPPDATA%/Temp/opencode/factory-browser-20261114-c`. The
binary hashes matched the two rebuilt artifacts in `FACTORY_BACKEND_EVIDENCE.md`
§7. Discovery again returned ready on port 17331, protocol 1; authentication,
pinned Pi readiness and the setup request advanced to project registration.

**Concrete blocker:** `POST /api/projects` returned **HTTP 400**. The launcher's
original error was `Disposable KLM POST /api/projects failed (status 400).`
The request was `{name: "Disposable Foreman journey", folder: "<run root>\\project",
icon: "folder"}`. Source inspection establishes that `prepareProject` calls
`validIcon`, which accepts an empty string or a valid PNG data URL, and rejects
the literal `folder` (`engine/project_registration.go:13–25`,
`engine/api.go:335–352`). The launcher intentionally suppresses response bodies;
the exact response message was not captured. Correct the disposable setup DTO
before another explicitly authorized attempt. No `ready.json`, Werbert, Vite,
browser or inference was reached. No earlier setup error or secondary revocation
error was reported by the corrected launcher.

The redirected supervisor start command printed PID 45844, but the tool still
timed out despite isolated stdio; this is an additional tooling limitation to
resolve for a live interactive pass. The supervisor subsequently completed: its
`foreman-browser-control-c/final.json` records launcher PID 39732 exit 1 and no
supervisor cleanup errors. `cleanup.json` records observed KLM exit, removed
temporary auth, unchanged original auth and unchanged saved Agent; Werbert never
started. No forced KLM stop was recorded. Process and listener checks found no
validation executable or supervisor and no listeners on 17331/18090/5174;
the unrelated 7331/7332 listeners remained. Sanitized diagnostics and supervisor
output are retained in `foreman-browser-control-c`; backend data remain private.

The requested browser controls, real native question and actual human response,
worker result and separate assessment are all still untested; human acceptance
remains pending.

## Fourth attempt — real browser reached Settings, automation stopped

The explicitly authorized fresh root `factory-browser-20261114-d` used the
corrected empty project icon and one foreground Node lifecycle
`%LOCALAPPDATA%/Temp/opencode/foreman-foreground-d.cjs`. Normal authenticated
startup succeeded; `ready.json` records an unauthenticated KLM state request as
401, Pi 0.85.1, engine `2dbc8257f09fae0ebe5f0d7af7f0bf16` and disposable
project `c718f8d25f87b04c6552b42ac96d42ae`. Browser `/foreman` displayed that
same engine identity and protocol 1; clicking **Connect** displayed the actual
registered project. Clicking **Create Factory** opened Settings at Factory
`f7236f39de10a4cf183c779a96b2ee13`. Only Connect and Factory initialization
mutations were observed; no model input was sent.

The pass then stopped due to a **validation-driver mistake**, not an established
application defect: it looked for `Switch to dark theme` while the browser had
opened in dark mode and the actual control offered the opposite action. No
Settings selection, save, instruction editor or Library interaction was
performed. A screenshot at `<run root>/settings-light.png` is actually the
initial **dark**, 1440×900 Settings view; the filename does not describe its
theme. It shows the empty model selection and `0 skills · 0 MCPs`. One generic
browser console resource-404 message was observed without an identified URL; no
Foreman API response of 400 or greater was captured. Its cause is inconclusive.

The foreground supervisor returned and awaited its launcher. `cleanup.json`
records both backend exits, auth-copy removal, unchanged original auth and
unchanged saved Agent, with no forced KLM termination recorded. Vite's owned
process tree was terminated and awaited; subsequent PID and listener checks
found no owned processes or listeners on 17331/18090/5174. Existing 7331/7332
listeners remained untouched. Werbert's Windows process termination is not a
graceful application-close acceptance check.

Repeated selection / stable Chat, settings and Library controls, light and
narrow layouts, saved applied capture, Ticket/worker, Stop/cancel, native
question and human answer, result and assessment remain untested. No question
arose and no Alpha/Beta answer was provided. A subsequent authorized run should
use theme-independent selectors and continue the blocked journey; no native
task was admitted in this attempt.

## Fifth attempt — real native question observed, unanswered

The parent authorized fresh disposable root `factory-browser-20261114-e`; a
single foreground Node supervisor launched and awaited the corrected launcher,
actual Vite and headless Edge. Protected discovery, explicit Connect, project
selection and Factory creation succeeded. The actual Settings controls selected
Pi, `openai-codex/gpt-6-astra`, and the slider's **Low** mark. Empty Library
showed zero skills, connections and tools; both categories were opened without
selection. `PUT /api/v1/foreman/factories/69ddbb1d5ef1b8f26ad92d5d2e9dfec8/definition`
returned 200 with `saved:true`, `applied:true`, `readiness:"ready"`; Chat showed
GPT-6 Astra, Low and read-only YOLO. Initial and opposite-theme 390px screenshots
are in the run root. Full visual acceptance is not established. The editor
Escape test was **inconclusive** because the driver used a non-exact label that
matched both a tablist and textarea; it did not make an edit.

One actual substantive Chat input was accepted (202), input ID
`35e6ddf3db38a5d38d9cd0ccb36fd88f`. The Ticket list showed exactly one Ticket
`460a49cb9f19df3df6deee038b796ecc`, state `Waiting for user`, with saved
`journey-worker` Assignment `9d45f07e8c463b9ab7e2e1625e5c92c7` and work
`b4ef2af40aabc3da2e698e0b5400a8f1`. It showed native state `running`, separate
from Foreman's running turn. The `GET /chat` native request was `pending_questions`,
state `pending`, request ID
`810df95af9dd2de4229ad77b227143ded45b44c1675c49d3c6d65e1b850929d0`, native ID
`ee45e43ebedfabb2bfda80a4e3f02d3f`, owned by that work. Foreman Chat visibly
rendered the exact native question: **“Which release label should
delivery-label.txt contain?”** Choices: **Alpha**, **Beta**; single choice,
no custom answer. Screenshot: `<run root>/native-question.png`.

No answer was selected or submitted. The synchronous foreground tool could not
invoke the human question tool while its owning lifecycle was running. Its
launcher was then stopped and both backends exited; the native callback's
post-shutdown availability and resumability are **unknown**, not an active live
request or permission to replay the task. The parent must obtain any actual
human decision and determine safe recovery of this exact work/request before
attempting a reply; do not automatically create replacement work. The label file
was absent on post-run inspection. No worker result, answer delivery/resumption
or separate objective assessment was observed.

The only browser console error was the identified `GET /favicon.ico` 404; no
Foreman API 4xx/5xx was captured. The lifecycle logged only Connect, Factory
creation, Settings save and Chat input mutations. Cleanup records KLM and
Werbert exits, temporary auth removal, unchanged original auth and saved Agent;
no forced KLM stop was recorded. Vite and browser were closed, and subsequent
process/listener checks found no owned processes or validation port listeners;
7331/7332 remained untouched. Werbert termination is not graceful-close proof.
Repeated selection, editor retention, nonempty Library allowlists, Stop with
independent worker, cancel-dismiss and worker native tab remain untested. Human
acceptance is pending.

## Sixth attempt — same native question, validation-driver refusal

The parent reports that the actual human used the question tool to choose
**Alpha** for the question “Which release label should delivery-label.txt
contain?”, explicitly for a fresh disposable validation run. No human event ID
was supplied. The parent authorized fresh root `factory-browser-20261114-f`;
the interrupted `-e` request above was neither replied to nor replayed.

The foreground run connected, created Factory `ad855c9d5f65e8afdfb7294bc8bccc34`,
saved and applied Pi `openai-codex/gpt-6-astra` / Low / empty Library, and
rendered read-only YOLO in Chat. The corrected editor selector verified that a
draft survives Escape; it was then cleared before saving. One input
`cf4d869107b62551b50f6f66f52048f7` was accepted. One Ticket
`97b82cff2a82295df5427b0dac3899c6` had one Journey Worker Assignment
`8df44ecd89c2a5552ad1438b2785a301`, work `e66a8d7b7e49cb5d7575c9af3ac7f75d`.
The Ticket was `Waiting for user`; the work's native state was `running`.

The actual native request `5ca94b429564880925ee4aa9f22e4984a753305aeb3cde4fb4fbe6b6149757e6`
(native ID `ae331df6cd8013968c42e284298276a5`) displayed the **same exact**
question and Alpha/Beta choices in Foreman Chat. Its payload had `multiple:false`
and `custom:true`. The driver erroneously treated `custom:true` as making this a
different question and stopped before forwarding the authorized Alpha. This is a
validation-driver error: the extra custom option did not alter the offered Alpha
answer. No request-reply mutation occurred, no result or assessment was observed,
and the label file was absent. Do not infer that the stopped native callback can
be resumed or replay this work without an explicit recovery decision.

The only browser console error was `/favicon.ico` 404; no Foreman API error was
captured. The lifecycle returned with KLM and Werbert exits observed, auth copy
removed, original auth and saved Agent unchanged, and no forced KLM stop
recorded. Browser/Vite were closed and their owned processes awaited. Subsequent
process and listener checks found no validation processes or listeners on
17331/18090/5174; unrelated 7331/7332 listeners remained. The earlier `-e`
request and this `-f` request are distinct and both have unknown callback
availability after shutdown. Human-answer delivery, worker result and separate
assessment remain unvalidated.

## Parent integration follow-through — real answer, result and assessment

After the Browser Tester reached the real question, the parent used the human
question tool. The human explicitly selected **Alpha** for forwarding to the same
label question in a fresh disposable run. No human message ID is invented here.
The parent corrected the driver's erroneous rejection of `custom:true`: offering a
custom answer does not invalidate the human's explicit offered Alpha choice.
Earlier stopped requests were preserved, not resumed or automatically replayed.

### First completion and concrete runtime correction

Root `factory-browser-20261114-g` completed the real path using the actual browser
and authenticated backends. Ticket `da8d723dd6c00a9a57e373fce417a450`, work
`c4d1844725a350ac275e47b00fd64a81`, and request
`6d51c08b964472e7346b01656d7d9cd0ae44854eaa2032f42e26ebce4545ba91`
retained one Assignment and the actual UI-submitted Alpha answer. The file contained
five bytes `41 6c 70 68 61`; native work settled at revision `112`, and the separate
assessment was satisfied. The foreground driver returned successfully with cleanup.

That run exposed a fresh `ticket_read` returning stale `sql.ErrNoRows`. Foreman
recovered through actual native-history reads; this did not make the defect
acceptable. The backend owner cleared the consumed lookup error, added the focused
`TestForemanFreshTicketRead` regression, and rebuilt Werbert. See
[backend evidence §8](FACTORY_BACKEND_EVIDENCE.md#8-runtime-discovered-fresh-ticket_read-correction).

### Corrected-path follow-through

One affected-path run used root `factory-browser-20261114-h`, final Werbert artifact
SHA-256 `6D1977A051BACC434B98B4B98BEFF05F49668FA307454D0DAA7C23082C236F13`
and the KLM artifact from backend evidence §7. Exact invocation, foreground and
bounded to ten minutes with finally cleanup:

```powershell
# Working directory: C:\Users\Samuel\AppData\Local\Temp\opencode
node "C:\Users\Samuel\AppData\Local\Temp\opencode\foreman-foreground-d.cjs"
```

| Evidence | Actual observation |
| --- | --- |
| Normal local authentication | Unauthenticated KLM state returned 401; protected discovery, pairing, explicit browser Connect and Factory creation succeeded. Engine `b239ca24a874f339a855ccba873597b3`. |
| Factory/configuration | Factory `2dcd4d700abbbb03c02507832732f1fe`; explicit Pi 0.85.1 / `openai-codex/gpt-6-astra` / Low; Settings returned saved/applied/ready with digest `1e1ee43cf395b3ff7b7ab12c6c890306dab9adc8832d682a646f6f2dd0a568dc`. Editor draft survived Escape. |
| Real asynchronous work | Input `55f01a9512593219d13b351f41a7f8ff`; one Ticket `1534328993cab48945f83b200c23549f`; Assignment `5879f155c4c2d44eae494e0f04206e02`; saved source UID `fc36188decf7449f974825931353ddcc`; work `d1ee60333b0040f8bf52472de615d4a3`. |
| Exact human intervention | Request `0905f6bb48e6ebeff9be0cd5688d787f296f03925ec33b262938f91155ed666c` visibly asked the label question in Foreman Chat. The actual human's Alpha was selected and submitted through that exact card. Reply operation `746b5591-32f8-4aea-8aaa-dc9aade9b224` returned 202; native execution actually continued. |
| Worker tab | Waited for real native history rows: `runner_input`, `agent_start`, completed `worker_acknowledge`, running `ask_user`. Own Pi/model/Low/YOLO capture rendered. No input/composer was present; closing the tab created no execution mutation. |
| Cancel dismissal | Opened the exact Ticket confirmation, selected Keep working; zero `/cancel` requests. No worker cancellation was claimed by this browser check. |
| Real artifact | `<run root>/project/delivery-label.txt` is exactly five bytes, hex `416c706861`, SHA-256 `b1a96dd646bccaa24cef7a3db22a6f995f05658f4f1c3272913e258c03e6fb24`. |
| Result versus assessment | Retained structured worker result arrived separately; native state `native_completed`, work revision `112`. Foreman independently recorded `outcome:satisfied`, `basis_work_revision:112`, reason and result references. Ticket became Completed, revision `9`. |
| Final Chat | Purpose-bound result publication `publication-ffff27e47aacfbe9fed2b5301aee591180cc1b8118d7ce57bc1ba469a52dfb6e`, cursor `158`. After Jump to latest, its actual text/path/hash were visible; Foreman state was idle. |
| Ticket UI | Visible body contained Completed, native_completed, retained worker result and Objective satisfied with the independent assessment reason. The driver's exact-text selector incorrectly chose a hidden filter `<option>Completed</option>` and timed out after this state was visible. This is a driver failure, not a passing final assertion or an application defect. No further rerun was performed for that selector. |

The saved screenshots `worker-native-readonly.png` and `completed-chat.png` in the
`-h` root were read and inspected by the parent. They establish their displayed
dark desktop views, not pixel parity or all responsive interactions. Earlier `-e`
captures cover opposite-theme/narrow Settings only. Empty Library categories were
inspected in that earlier run; nonempty MCP/skill selection was not exercised.

No Foreman API 4xx/5xx or page exception was observed during the completed paths.
The only console error was the identified `/favicon.ico` 404. Native request state
remained `delivered` after resumption, truthfully distinct from native settlement;
its historical card was locked. Native success was not inferred from HTTP 202.

### Cleanup and limits

Both foreground runs awaited Vite/browser shutdown and launcher exit 0.
`cleanup.json` records both backend exits, temporary native-auth removal, unchanged
original auth and unchanged saved Agent. No forced KLM termination was recorded.
Final `Get-NetTCPConnection` inspection found no listeners on 17331/18090/5174.
The existing 7331/7332 services were untouched. Werbert Windows process termination
is not graceful-close acceptance. Private databases/credentials remain private
validation artifacts; they should not be attached to reports.

The representative real journey is established; broad human acceptance is pending.
Remaining manual checks include nonempty Library/MCP variants, real audio, repeated
Factory selection, independent paused-history readers, broad ordinary KLM UI
comparison, native compaction/restart/Linux, and integrated OpenCode/Codex journeys.
The separately recorded native/boundary checks cover all three pins and independent
Stop/cancel ownership; those are not browser proof of every control permutation.
