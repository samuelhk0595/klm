# Saved engines: isolated native validation procedure

Prepared for the independent tester AFTER coordinator approval of CP3. These
commands have not been executed by the implementation session. Compilation is not
runtime acceptance. Use native Tauri/WebView2; a browser with mocked IPC does not
verify Credential Manager or DPAPI behavior.

## Boundaries

- Do not connect this probe to API 7331, installed engine data, release credentials
  or installed-client processes. Do not launch/stop/restart/reinstall them.
- Use a dev-tag engine on 17331, Vite 5173 and Tauri's dev Focus service 17332.
  If these or CDP 9222 are occupied, report the owning situation; do not kill/reuse
  an unknown listener or silently choose another app port.
- The VPS is the second engine only after the coordinator supplies its explicitly
  authorized HTTPS URL/access procedure. Do not deploy/restart/update it or edit
  its allowlists to force a test through without authorization.
- Use only synthetic test projects/data. Do not send real provider jobs, install
  skills/MCP servers or put real API keys/PATs into fixtures. Pairing codes/tokens
  stay out of reports, screenshots and commits.

## 1. Prepare isolated engine/profile files

Run the preparation in the tester's own environment, not in the implementation
turn. Example PowerShell (adjust only the checkout/probe paths if necessary):

```powershell
$repo = 'C:\Users\Samuel\Documents\Projects\Personal\klm-multi-engine'
$parent = 'C:\Users\Samuel\AppData\Local\Temp\opencode'
if (-not (Test-Path -LiteralPath $parent)) { throw 'Probe parent is unavailable.' }
$probe = Join-Path $parent 'klm-multi-engine-native-probe'
New-Item -ItemType Directory -Force -Path $probe
Get-NetTCPConnection -LocalPort 17331,17332,5173,9222 -ErrorAction SilentlyContinue
```

Stop if an expected port is already owned by another probe/app. Create an absolute
`engine-data`, `user-skills`, `webview-profile` and `project` directory below the
probe root using the tester's normal file tools. Keep the same absolute
`KLM_DEV_DATA_DIR` for every CLI command against this probe:

```powershell
$env:KLM_DEV_DATA_DIR = Join-Path $probe 'engine-data'
$env:KLM_USER_SKILLS_DIR = Join-Path $probe 'user-skills'
$env:KLM_SKILLS_DIR = Join-Path $repo 'engine\skills'
$env:KLM_PROMPTS_DIR = Join-Path $repo 'engine\prompts'
Remove-Item Env:KLM_DEV_DISABLE_CLIENT_AUTH -ErrorAction SilentlyContinue
go -C "$repo\engine" build -tags dev -o "$probe\klm-multi-engine-dev.exe" .
& "$probe\klm-multi-engine-dev.exe" start
& "$probe\klm-multi-engine-dev.exe" pair
```

Do not use VS Code's auth-bypass debug launch. Do not run ordinary `klm start`,
`klm stop` or `klm pair` from PATH: that could target the installed release.
Dev data/control overrides do not isolate global provider/harness accounts; no
provider execution/installation is needed for the draft/selection probe.

## 2. Native development client + CDP

Use a dedicated test app identifier too, so an existing ordinary DEV client's
Credential Manager entries and DPAPI documents are not overwritten. Create
`<probe>/tauri-probe.json` with the tester's file tools:

```json
{"productName":"KLM Multi-engine Native Probe","identifier":"dev.klm.desktop.multi-engine-probe"}
```

From `clients/desktop` in a separate PowerShell with the probe environment:

```powershell
$env:VITE_ENGINE_URL = 'http://localhost:17331'
$env:WEBVIEW2_USER_DATA_FOLDER = Join-Path $probe 'webview-profile'
$env:WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS = '--remote-debugging-port=9222 --remote-debugging-address=127.0.0.1'
npx tauri dev --config "$probe\tauri-probe.json"
```

The config changes only this launch's identifier/name; normal Tauri debug mode
uses the fixed dev Focus port. Its beforeDevCommand produces the embedded assets
and starts Vite. No installer is required. Do not reuse a nonempty unknown profile:
the native override is only a seed, so an old saved selection can take precedence.
Verify the active UI endpoint is exactly `http://localhost:17331` before pairing or
opening test data. If WebView2/environment/config profile behavior differs, report
the blocker instead of connecting to an unknown stored endpoint.

Use the tester's existing Playwright installation. Minimal connection example:

```js
import { chromium } from 'playwright';
const browser = await chromium.connectOverCDP('http://127.0.0.1:9222');
const pages = browser.contexts().flatMap(context => context.pages());
// Choose the KLM native app page by URL/title and confirm window.__TAURI_INTERNALS__.
// Do not choose a Focus/browser page or synthesize a Tauri bridge.
```

CDP availability, native dialogs/clipboard, WebView2 profile flags and physical
OS storage are unvalidated runtime paths. If CDP is unavailable, stop and report
that limit. Browser-only checks are separate evidence, never a substitute for
native pairing/DPAPI validation.

## 3. Focused acceptance matrix

1. Pair Local dev with the one-use CLI code; add authorized VPS with a distinct
   Engine name and Client name. Choose an icon, then initials; reopen desktop and
   confirm registrations/last selection persist. Domains receive no added port.
2. Local -> VPS -> Local: observe full document reload and destination-specific
   requests, projects, sessions, topics, KLM Agent, Analytics, Tasks and settings.
   Only the active native page maintains updates/renewal. Opening/searching the
   saved list must not probe inactive engines.
3. Preserve text/mentions/side passages and all representative form groups from
   `MULTI_ENGINE_DRAFTS_COVERAGE.md`, including invalid JSON and dummy secret inputs.
   Close dialogs with X to reach the selector, switch, return and reopen them.
   Verify no save/apply/answer/OAuth/install/graph execution happened on restore.
4. Delay a real save/submission/layout response using client-side network controls.
   Start switching: controls freeze and show waiting activity. Cancel switch
   restores the old UI and does not Stop/abort/undo the server mutation. Finish
   the request, retry switching, and check values/revisions/receipts remain coherent.
5. Message accepted with response lost: return to that engine; queue/history ACK
   must clear only matching composer/sent passages with its receipt, preserving
   newer text/sources. No automatic resend or new implicit message ID.
6. Audio recording, microphone prompt, icon picker and engine-native directory
   picker: switching requires finishing/cancelling the interaction. Transcription
   already in progress is awaited; its produced text remains an unsent draft.
7. Invalid/duplicate URL, invalid code, unknown/expired credential, candidate
   verification failure, repeated selection and current-entry no-op. Candidate
   failure must not invalidate active auth. Use saved access recovers after a later
   registry/verification failure without redeeming the old code again.
8. Destination offline/401: keep its own name/icon/URL with no previous-engine data;
   recover through Retry/Pair again/Switch/Add outside the gates. Test unselected
   and zero-entry state, including removal of the active entry without auto-picking
   another. Removal keeps server jobs/data and protected drafts.
9. Storage read/write/readback failure on the isolated profile only: block unsafe
   reload/selection commit, retain originals/dirty memory and do not claim rollback
   of a selection that might have persisted. Explicit registry recovery preserves
   its exact original in the protected profile document before replacement.
10. Save a large set of synthetic metadata entries in the isolated profile, without
    real pairing/network probes for inactive rows. Check search/scroll, long and
    Unicode names, focus ring, Arrow/Tab/Escape and focus return, both themes.

Use client network interruption for offline tests. Do not stop/restart the shared
VPS or the installed engine. Any database-identity replacement test uses intentional
separate disposable dev data, with no automatic replay and no production copy.

## 4. Evidence and cleanup

Report actual harness/client versions, native page/profile identity, endpoints,
reproducible steps, expected/actual behavior and token-redacted evidence. Distinguish
TypeScript/native compilation from OS storage, native UI and real HTTP/SSE results.
Leave test artifacts available to reproduce failures unless the coordinator asks
for removal. No cleanup touches installed/normal DEV data or the VPS.

After explicit permission to end the owned probe, stop only its absolute dev binary
with the same `KLM_DEV_DATA_DIR`; exit only the dedicated probe client. Do not use
release CLI stop or kill unknown PIDs.
