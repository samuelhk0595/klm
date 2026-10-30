# Multi-engine: independent native Playwright validation

Date: 2026-10-30. Tester session: `2f206511d102aa6338998a076e337d96`.
Coordinator: `bd8d7990868ce6bf1b3d22bab9187734`.

## Result

**Directed validation completed, with one UI locator defect and explicit coverage
limits. This is not full feature acceptance.** Real Windows/Tauri/WebView2 pairing,
Local DEV → VPS → Local DEV switching, native credential retention, representative
protected drafts, process reopen, transport teardown and several failure/recovery
paths were exercised. No production code was fixed by the tester.

Tested source: `72a959f9f31b87ceac6134e7eadc4d0999dad4ca`, branch
`feat/desktop-multi-engine`, worktree
`C:\Users\Samuel\Documents\Projects\Personal\klm-multi-engine`.
Draft PR: <https://github.com/samuelhk0595/klm/pull/40>.
HEAD was still that commit at completion. No merge, commit, deploy, installer,
full test suite, adversarial review, spawned session, native subagent or graph
execution was performed. The original dirty `klm` worktree was not used for builds
or modified. This report is the tester's only new tracked-content deliverable.

## Environment and evidence

All tester scripts, redacted request metadata, screenshots and encrypted backups:

`C:\Users\Samuel\AppData\Local\Temp\opencode\multi-engine-cp4`

- Go 1.24.2 Windows/amd64; engine compiled from this worktree with `-tags dev`.
- Node 22.22.3, npm 11.16.0, Cargo 1.93.1, Playwright 1.58.2.
- Frontend Vite 7.3.6 build and native Rust debug build completed. No full suite ran.
- Real native app name `KLM Multi-engine Native Probe`, version `0.6.0`.
- Native app identifier/app-data directory: `dev.klm.desktop.multi-engine-probe` /
  `C:\Users\Samuel\AppData\Roaming\dev.klm.desktop.multi-engine-probe`.
- New WebView2 profile: `<probe>\webview-profile`. Native user-agent reported
  Edge/Chrome 156.0.0.0. `window.__TAURI_INTERNALS__` was present; native app/path
  calls and DPAPI IPC executed successfully. No bridge was simulated.
- Native page `http://127.0.0.1:5173/`; Vite 5173; native Focus listener 17332.
- CDP was attached with `chromium.connectOverCDP` to
  `http://127.0.0.1:9222`. Listener verification showed loopback only.
- Local authenticated API: `http://localhost:17331`, database identity
  `3f9136a9da08becfde15669bf096d948`. Engine data/control and user skills were below
  the probe root. `KLM_DEV_DATA_DIR` and `KLM_USER_SKILLS_DIR` were absolute;
  `KLM_SKILLS_DIR`/`KLM_PROMPTS_DIR` explicitly used this worktree's internal sources.
  `KLM_DEV_DISABLE_CLIENT_AUTH` was removed from the process environment.
- VPS: `https://cubiqklm.codemob.com.br`, database identity
  `a966288835000cc591796081871f8fd3`; SSH hostname `srv1763338` confirmed using the
  supplied existing key and strict known-host validation. No key contents were read.
  Final read-only binary SHA-256:
  `1bc277aae4eb1703154b6ccd4a6f0bb9a0602c3cd37458daec361bb8949684bf`.
  The public CLI has no version/help subcommand beyond its usage output; no remote
  source revision is inferred from the desktop's version.
- Local engine PID 19708; initial native PID 4928, reopened native PID 41260;
  Vite PID 6696. Reopened native WebView2 CDP owner was 14368.

The first `Start-Process` launch was held open by terminal/process containment;
the user interrupted it. Its processes ended. The subsequent launcher used WMI
process creation, detached child processes and file-backed stdout/stderr. The
already-built native executable and Vite were launched independently, with the
same dedicated identity/profile and DEV ports. This was an environment-launch
issue, not a product pairing failure. The final owned engine ran its absolute
DEV executable's worker entry point with the isolated environment.

No request headers, bearer values, cookies, pairing codes, response bodies or
HAR/traces were recorded during pairing. Local and SSH-issued codes were captured
by automation in memory and entered into password inputs. Password inputs were
masked in screenshots. Form secrets used for coverage were synthetic and were
never submitted to a provider, MCP server, GitHub or transcription service.

## Directed results

| Scenario | Outcome | Evidence / qualification |
| --- | --- | --- |
| Authenticated native Local pairing | PASS | Actual CLI code, native keyring, authenticated workspace. `native-identity.json`, `local-after-pair.png`. |
| Authenticated native VPS pairing | PASS | Real pair POST 200, authenticated session/state 200, updates 200. `vps-pair-network.json`. Current native/Vite origin was accepted without changing VPS allowlists. |
| Local → VPS → Local | PASS | Full document navigations, different endpoint/DB identities, local identity restored. `basic-roundtrip.json`. |
| Stream teardown / one active endpoint | PASS | CDP Network tracking showed the live old Local updates stream end on switch; the remaining live updates stream belonged only to the VPS. Return also terminated VPS streams. No page errors in this capture. `sse-lifecycle.json`. |
| Active-entry no-op | PASS | Selecting the current Local entry closed the list without another document navigation. `basic-roundtrip.json`. |
| Local name/icon/initials and domain URL | PASS | Distinct local names, cloud/server icons, restored initials, no added domain port. Local icon edits did not reconnect/reload. `selector-recovery.json`, native registry metadata. |
| Canonical duplicate / invalid URL | PASS | Uppercase VPS host plus trailing slash rejected as duplicate before candidate traffic; non-loopback HTTP rejected before network. `connection-errors.json`. |
| Invalid pairing code | PASS | Real local alias API rejected invalid code with 400; current Local document remained active. `connection-errors.json`. |
| Candidate verification fails, A remains usable | PASS | Actual pairing of `http://127.0.0.1:17331` succeeded; client interception then failed only candidate verification. Active `localhost` authenticated session still returned 200. `Use saved access` recovered without another pair POST. `connection-errors.json`. The alias was subsequently removed. |
| Remove active registration | OBSERVED PASS | Removing the active local alias reloaded into explicit no selection even with Local/VPS still saved. The automatically opened manager required adjusting the automation's next locator. No server work/data was deleted. |
| Destination offline | PASS, client fault injection | Requests to VPS aborted on the client. VPS name/icon/URL and recovery actions remained, without Local workspace content. Returned through selector. `selector-recovery.json`, `offline-vps-recovery.png`. VPS was not stopped. |
| Destination 401 | PASS, client fault injection | Intercepted session 401 exposed Pair again and selector. This correctly cleared that probe profile's VPS credential. A later **real new pairing** repaired it. `selector-recovery.json`, `vps-auth-repair.json`. No server credential revocation was performed. |
| Full native process exit/reopen | PASS | Killed only owned native PID 4928 and its children, then launched the same binary with the same profile. Native credential authentication and DPAPI drafts returned. `native-reopen.json`. |
| Native DPAPI persistence | PASS for exercised data | Four `.dpapi` documents present; native decrypt/read worked. Ciphertext did not contain tested chat/dummy-secret plaintext sentinels. Encrypted backups retained. `native-vault-preflight.json`; this is not a separate cryptographic audit. |
| Cancel switch during real finite write | PASS | Real isolated graph layout POST reached the engine and returned 200; client held the response. Switch displayed `Saving graph layout` and Cancel switch. Cancel kept Local and did not cancel/undo the accepted layout save. `graph-guards.json`, `switch-waiting-layout.png`. |
| Native draft write failure | PASS for retention/blocking | Backed up/renamed only this app's drafts directory and temporarily replaced its path with an owned file. Native write failed with `Cannot create private draft storage.`; no switch reload or VPS request occurred. Dirty composer remained in store memory. After filesystem restoration the same composer returned. `storage-write-recovery.json`, `native-write-failure.png`; see runner caveat below. |
| Native decrypt/read failure | PASS | Backed up the exact Local document, temporarily corrupted only that file, reloaded: explicit decrypt failure, exact corrupt bytes untouched, selector available. Could leave to VPS; restoring original bytes and returning restored the draft. `dpapi-read-failure.json`, `dpapi-read-failure.png`. |
| Registry write failure | PASS, Storage fault injection | No destructive reload; current runtime stayed Local. `registry-faults.json`. |
| Registry readback failure after successful write | PASS, Storage fault injection | Durable selection actually became VPS while current page remained Local; neutral error did not claim rollback. Fixture was restored afterward. `registry-faults.json`, `registry-readback-failure.png`. |
| Corrupt registry / preserved-original recovery | PASS | Malformed record failed closed and stayed exact before explicit recovery. Restoring valid JSON preserved the original inside native pre-auth DPAPI. `registry-recovery.json`. |
| Zero-entry cold start | PASS | Explicit preserve-and-start-new-list produced zero entries/null selection; cold reload did not re-import VITE seed; Add engine stayed available. `registry-recovery.json`, `zero-entry-recovery.png`. Valid fixture registry restored afterward. |
| Large list / inactive probes | PASS, metadata fixture | 152 local metadata rows; 360px scroll viewport with 7770px content. Search with Unicode/long names, Home/End, Escape/focus return passed. No requests to 150 synthetic inactive domains. `selector-scale.json`. |
| Light/dark / avatar placement | PASS for directed desktop check | Avatar y positions were Theme 780, Engine 816, Settings 852. Inspected 1440×900 native screenshots in both themes; long text truncated within the selector. `selector-many-light.png`, `selector-many-dark.png`, `endpoint-draft-isolation-final.json`. |

## Protected draft coverage

Edits used normal UI controls and Close/X rather than Save/Apply/Discard. Restoration
was checked after actual endpoint switches. Representative values were additionally
checked after native process exit/reopen. No remote provider jobs were used.

| Group | Runtime coverage |
| --- | --- |
| Normal main chat | Text and real file mention from a synthetic local project file survived reload/switch. Main text also checked after process reopen. |
| Side chat | Independent text survived roundtrip and reopen hydration. No selected-history passage fixture was generated. |
| General chat | Local unsent text survived Local/VPS/Local and was present in native protected document after process reopen. |
| Remote main chat | VPS-only unsent text returned on a second real switch and did not appear in Local. |
| New session | Unsent Unicode title survived roundtrip before creating the empty local fixture session. No initial prompt was sent. |
| Topic | Unsent name and Teal color survived roundtrip. No topic creation/move submitted. |
| New project / directory browser | Name, typed path and directory filter survived roundtrip and reopening the in-app browser. No folder mutation submitted. |
| Agent editor | Name and prompt survived Close/X and roundtrip. No agent saved/executed. |
| Graph details / subeditor | New-graph name/description retained. A local authoring-only layout fixture then exercised an **unapplied terminal command/name** surviving switch and full native reopen. No graph ran. Choice/Fork/Join RAW previews were not individually exercised. |
| Transcription settings | Model/language/dummy API key retained; key also checked in UI after native reopen. No transcription/save. |
| Engine access settings | Unsaved origins/proxy CIDR retained. No access save or restart. |
| Library | Invalid RAW JSON arguments, environment and headers plus dummy bearer retained. No connection save, OAuth, installation or external MCP request. |
| Tasks connection | Name/repository/dummy PAT retained. No discovery, Save, external project/mapping or task movement. |
| Webhook | Name/retention/dummy secret retained. No webhook registration or delivery. |
| Automation | Name/time/initial prompt/new-topic name and current editor step retained. No automation save/activation/run. |
| Project Settings section locator | **FAIL CP4-01**, below. Field values remained recoverable by explicitly opening their sections. |

Evidence: `drafts-roundtrip.json` (successful field assertions before its section
locator failure), `drafts-project-verified.json`, `graph-guards.json`,
`native-reopen.json`, `endpoint-draft-isolation-final.json`.

Observed non-GET traffic during final draft restoration was only three Local
idempotent `/api/sessions/<local-main>/side` initialization requests for the already
opened side chat. There were no message/answer/Save/Apply/OAuth/execution calls and
no VPS mutations during that capture. These side initialization requests are
reported explicitly rather than claiming that restoration made literally no POSTs.

## CP4-01 — Project Settings reopens in General

Already sent to the coordinator using operation
`cp4-01-project-settings-section-reset`.

1. In the local fixture, open Project settings → Tasks and type a new unsaved
   connection, or choose another non-General section.
2. Close Project settings with X, switch Local → VPS → Local.
3. Open Project settings using its sidebar button.

Expected: retain the project's previous section, as promised by the coverage
matrix's same-project/section locator. Actual: General opens. The values themselves
still exist when Tasks/Webhooks/Automations are explicitly opened.

Concrete source linkage: `App.tsx:1030–1033`, `openProjectSettings` defaults its
section argument to `general` and explicitly overwrites
`ui.projectSettingsSection:<projectId>` before reopening. This is a navigation
continuity defect, not an observed draft-content loss. No tester workaround or
product fix was applied. No changed product commit was retested.

Standalone evidence: `<probe>\issue-project-settings-locator.md`.

## Limits — not passed by inference

- No real accepted-message/lost-ACK/provider execution fixture was used. Therefore
  queue/history ACK reconciliation, uncertain-delivery receipts and no-replay after
  an actual lost message ACK remain **unverified**, despite no unsolicited sends
  in the tested draft restoration paths. No simulated native bridge/ACK was used.
- No real running remote job/graph/automation was started to prove work continuity.
  No Stop/Restart calls were observed during switching, but continued remote work
  itself is **unverified**.
- Microphone enumeration reported one audio input with permission `prompt`.
  Permission was not auto-granted. Recording/microphone-prompt guards, actual native
  OS icon/directory picker open/cancel guards and transcription-in-flight were
  **not exercised**. In-app directory browsing is separate covered behavior.
- No pending native question/elicitation, side passage/history fixture, task phase
  mappings/move, uncertain automation Run now, revision-conflict/missing-entity,
  database replacement at the same endpoint, old-client migration or permission
  answer was fabricated. These rows remain **unverified**.
- Representative groups are covered, not every field/variant in the inventory.
  Rename-session/project/topic variants, Session Library selections and graph
  Choice/Fork/Join subeditors need their own acceptance route.
- No full browser/mobile/installer/LAN regression or accessibility audit. Theme,
  keyboard and scroll evidence is directed native desktop validation only.
- SSE teardown was observed with CDP. A 60-second inactive-renewal soak and every
  possible background poll were not separately measured.

## Runner caveats and reproducibility

Scripts in the evidence directory are **stateful directed probes**, not a suite to
batch-run from an arbitrary profile. They retain original evidence, including
runner failures: strict/substring locators, hidden textarea label text, a menu used
before bootstrap replaced the recovery rail, fixture metadata rejected for
length/trailing whitespace, and an overbroad assumption that every POST meant a
draft submission. Those were narrowed in subsequent probe scripts; they are not
additional product defects.

In particular `storage-write-failure.json` records a runner timeout looking for a
composer while the correct draft-error gate was visible. It had already confirmed
no reload/current endpoint preservation. The follow-up
`storage-write-recovery.json` confirms retained dirty memory and composer recovery;
it does not claim a deterministic explicit Retry click. Similarly
`endpoint-draft-isolation-final.json` interprets the captured non-GET traffic
explicitly rather than deleting the original overstrict assertion artifact.

For reproduction use the dedicated profile/environment and existing scripts;
recover encrypted/registry fixture originals before changing scenarios. The
coordinator must release any corrected commit before the tester retests it.

## Cleanup / remaining artifacts

- Owned engine stopped using the absolute DEV binary and the same isolated data,
  skills and prompts environment; CLI reported `KLM engine stopped.`
- Owned reopened native PID 41260 and its WebView2 descendants were terminated.
  Initial native PID 4928 had already been terminated for the reopen test.
- Owned Vite PID 6696 and its child processes were terminated.
- No installed/principal local engine process, API 7331, normal profile/data/control,
  provider auth, VPS service/configuration or worktree was stopped/removed/updated.
- Safe evidence, disposable project/engine data, new WebView2 profile, dedicated
  native app-data/DPAPI files and encrypted backups were retained for reproduction.
  Dedicated probe credentials stay in that app identifier's native keyring namespace;
  test pairing records remain on the VPS. No global client revocation was attempted.
- Final registry fixture again contains only Local DEV and the authorized VPS,
  with Local active. Synthetic metadata rows, alias entry, fault hooks and temporary
  corrupt files were removed/restored. Product-source `git diff --exit-code` was
  clean before this report was added; build-generated/ignored artifacts remain.
- Final process/port query found no remaining owned PIDs 19708/41260/6696 or
  listeners on 17331/17332/5173/9222.
- `git status` still marks `clients/desktop/src-tauri/Cargo.toml` modified after
  the native CLI/build's line-ending handling, despite an empty content diff and
  a HEAD restore. `git ls-files --eol` reports index LF/worktree CRLF; both the
  filtered worktree hash and HEAD object are
  `1ed087fc858962dbda36d85ce2c3c98509cabe99`. No Cargo dependency/feature/content
  change or staged change was introduced. This status artifact is reported rather
  than hidden by staging or changing Git configuration.
