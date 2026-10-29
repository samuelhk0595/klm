# Engine directory browser implementation

## Delivery state

Implemented the approved focused plan. The single independent static review
completed; the same implementer corrected its one required finding (R1). No second
or post-fix review occurred. Compilation is evidence only, not runtime acceptance.
No deployment, server/browser launch, project/session creation, provider/harness
work, suites, integration tests, commits or pushes were performed.

## Scoped changes

- `engine/directory_browser.go`: authenticated read handlers, path validation,
  lexical parent/alias handling, immediate directory-only scan, error mapping,
  cancellation checks and encoded-response/work limits.
- `engine/directory_browser_windows.go`: actual logical-drive bits and explicit
  native-picker implementation capability; no drive probing/share discovery.
- `engine/directory_browser_unix.go`: Unix `/` root and no native capability.
- `engine/api.go`: two route registrations under existing protection/no-store.
- `engine/sqlite_api.go`: both routes admitted by existing availability gate;
  storage-contract checks remain unchanged.
- `clients/desktop/src/engine.ts`: shared authenticated, abortable API helpers.
- `clients/desktop/src/features/projects/DirectoryBrowser.tsx`: feature-local
  browser and request generations/controllers, explicit selection/native action,
  retained manual draft, recovery, focus and keyboard behavior.
- `clients/desktop/src/features/projects/directoryBrowser.css`: scoped responsive,
  token-based list/navigation/focus styles; existing dialog/buttons retained.
- `clients/desktop/src/features/projects/ProjectDialog.tsx`: in-dialog browser on
  every client, manual path retained, unchanged POST and edit-directory contract,
  name/icon drafts and focus restoration.
- `product.md`, `engine/README.md`: narrow behavior/API documentation.
- This progress file and an implementer disposition appended to the review document.

## API and bounds

- `GET /api/filesystem/roots` returns roots, engine-user home when available, and
  engine-native picker implementation capability.
- `GET /api/filesystem/directories?path=<encoded absolute path>` returns cleaned
  lexical path, server-computed nullable parent, directory names/paths, incomplete
  and skipped count. Hidden directories and directory symlinks/junctions follow
  normal OS resolution; alias identity is not canonicalized.
- No ordinary-file inventory/content, writes, creation, uploads or recursion.
- Absolute UTF-8 NUL-free paths: 16 KiB maximum. Invalid/non-directory 400,
  missing 404, permission 403, other I/O errors recoverable 500.
- Batches of 128; at most 10,000 inspected entries, 1,000 returned directories,
  256 KiB encoded JSON including escaping/metadata. Caps/skipped entries mark
  incomplete; retained directories sort by name without claiming a global prefix.
- No filesystem work under `app.mu`. Handles close; context is checked between
  batches/stat calls. Blocking OS I/O cannot be forcibly interrupted; no timeout
  goroutines were introduced.

## Lightweight check evidence

Commands ran from the repository unless the working directory is specified.
All used already installed dependencies. Shell-local GOOS/GOARCH assignments did
not change the parent environment. Outputs stay outside the repository.

1. `gofmt -w engine/directory_browser.go engine/directory_browser_windows.go engine/directory_browser_unix.go`
2. From `engine`: `GOOS=windows GOARCH=amd64 go build -o C:/Users/Samuel/AppData/Local/Temp/klm-directory-browser-workflow-dNsCwb/klm-windows.exe .`
   — passed, release build, exit 0.
3. From `engine`: `GOOS=linux GOARCH=amd64 go build -o C:/Users/Samuel/AppData/Local/Temp/klm-directory-browser-workflow-dNsCwb/klm-linux .`
   — passed, release build, exit 0.
4. From `clients/desktop`: `./node_modules/.bin/tsc --noEmit -p tsconfig.json`
   — passed, exit 0.

Both affected Go builds and the affected TypeScript check were repeated once after
small implementation corrections (non-directory error classification, draft-edit
request invalidation and exact failed-target Retry); final runs passed. No test
suites or runtime probes were run.

After the review's UI-only R1 correction, only the affected TypeScript command was
rerun: `./node_modules/.bin/tsc --noEmit -p tsconfig.json` from `clients/desktop`
— passed, exit 0. The Go source did not change after the successful release builds.

## Preservation and blockers

Scoped pre-edit baselines (including absence of new files), hashes, immutable ready
snapshots and an implementation-only baseline delta are outside the repository at:
`C:/Users/Samuel/AppData/Local/Temp/klm-directory-browser-workflow-dNsCwb`.
Shared files were reread before precise edits; unrelated SQLite/Linux work was not
reset or reformatted. Only the three new Go files were formatted. Initial submission
snapshots/delta remain immutable. Separate pre-correction baselines capture the UI,
progress document and reviewer-authored review document before R1 changes; final
snapshots/correction delta are published separately with completion metadata.

The current concurrent SQLite API still requires read contract `3`. At submission,
`clientAuth.ts`/`engine.ts` contain no `X-KLM-Read-Contract` header or contract query.
Thus the existing shared client's SQLite requests, including these new routes, can
receive 426 until the storage work updates that client. This feature deliberately
does not bypass/change the contract or repair unrelated storage/frontend work.
No compiler blockers were observed in the submitted working tree.

## Independent reviewer disposition

- R1 (P2): corrected by this same implementer after receiving the event-based review
  result. Directory navigation now explicitly invalidates listing selection; only
  the current successful directory response restores it. Prior rows remain hidden
  while invalid, retaining old parent metadata solely for Up recovery.
- Native errors are separate from directory errors. Native invocation, null result
  and Cancel selection no longer clear failed-directory error/Retry or revive
  invalid selection. Existing generation/controller/success/error/finally guards
  remain intact.
- The affected TypeScript check passed. The reviewer did not review these fixes;
  this is the implementer's disposition, not a second review or runtime acceptance.
- No required findings remain unaddressed. The preexisting SQLite contract mismatch
  and all pending human runtime checks remain disclosed.

## Pending human validation

- Windows drives/home/up; Unix `/` and home under nonroot permissions; UNC share
  roots/manual absolute paths; symlinks/junction aliases and broken links.
- Missing/deleted/inaccessible/offline directories, Retry, complete-empty versus
  bounded-incomplete listings and skipped entries.
- Rapid navigation, typing during requests, cancellation and close/reopen without
  stale listing/native result writes; native success/cancel/failure on an engine
  with/without an interactive desktop.
- Desktop, remote browser and mobile WebView, narrow/landscape layouts, both themes,
  keyboard/touch, focus return and reachable actions. No Flutter/Tauri changes.
- Unchanged manual registration/edit behavior and authenticated/revoked/wrong-origin
  protection. Browsing/selecting alone must not register or execute anything.

## Parent deployment after the user's subsequent authorization

After the same implementer completed R1 and its TypeScript recheck, the user
instructed the parent to install the updated engine on the VPS. This section is
parent deployment evidence, not additional work/review performed by the subagents.

- Installed release Linux binary SHA-256:
  `05fe83a9174ffe24040a3cbab467572e5628ed12110b39790ff1b0f2c23caa3f`.
- Target: `187.77.43.129`, API `https://cubiqklm.codemob.com.br`, binary
  `/opt/klm/bin/klm`, existing nonroot `klm.service`. Uploaded and published hashes
  matched. No DNS/ingress/harness/provider changes were made.
- Before restart, authenticated health reported zero running sessions/graph runs.
  The previous installed hash matched the expected release. Stopped only this
  engine, took a root-protected binary/data backup, atomically replaced its binary,
  and started the existing unit. Persistent data and paired credentials were not
  reset. Backup: `/opt/klm/backups/directory-browser-dNsCwb`; its data archive is
  mode 0600 and backup directory mode 0700. Other services were not restarted.
- Socket inspection after start confirmed **127.0.0.1:7331 only**.
- A focused HTTPS API smoke reused the existing paired Chrome context's cookies,
  with no credential extraction/printing or new pairing: `/api/health`,
  `/api/filesystem/roots`, and the engine-home directory listing all returned 200.
  Roots included `/` and `/var/lib/klm`; native picker capability was false; the
  home listing was complete with no skipped entries. No project/session was
  created and no model prompt was sent.
- The disclosed SQLite contract mismatch is conditional on `app.store != nil`.
  The submitted/current startup path still uses the existing non-SQLite startup;
  the deployed release accepted the current paired client without contract-header
  changes or a 426 response. SQLite-backed client compatibility remains outstanding
  for that separate migration; no contract/auth bypass was introduced here.
- The Windows Tauri application/bundled frontend was **not rebuilt or installed**
  by this deployment. New UI source passed TypeScript, but using its new browser
  requires an updated local frontend/Tauri bundle. No frontend was moved to the VPS.
  Browser UI/native/mobile runtime acceptance remains pending; authenticated API
  success is not end-to-end UI acceptance. The earlier paused Chrome UI was not
  resumed to run additional scenarios.

Metadata-only deployment/smoke evidence and immutable source snapshots remain in
`C:/Users/Samuel/AppData/Local/Temp/klm-directory-browser-workflow-dNsCwb`.
No further review or test suite was performed.

## Reference-led picker usability refinement

The user subsequently requested the interaction/layout of their
`Downloads/pilepicker.png` reference. Updated only `DirectoryBrowser.tsx`, its scoped
CSS and the product behavior paragraph; no backend, VPS, installer or production
process changes were needed for this refinement.

- Native-style header with Cancel/title/Select, locations sidebar, compact current
  location/Up toolbar, name filter and sort, directory rows and manual path footer.
  No fabricated Recents/Downloads shortcuts or unsupported size/date columns.
- Single click selects a child and shows its absolute path; double click, Enter,
  ArrowRight or the row Open control navigates into it. Arrow keys/Home/End move
  selection. Explicit confirmation alone returns the path to the registration
  form; it never creates a project. With no row selected, current-directory
  selection remains available.
- Selecting a valid row explicitly replaces an unfinished manual-path draft.
  An unsubmitted path itself cannot be confirmed. Listing-validity and generation/
  cancellation guards remain; native cancellation cannot revive failed listings.
- Responsive location shortcuts and reachable actions on narrow screens; shared
  theme tokens and semantic buttons/focus styling retained.

Checks actually performed: frontend `npm run build` (TypeScript and Vite) passed;
rerun once after a focused selection/accessibility correction, also passed.
The skill's mechanical detector ran once on the two UI files and returned no
findings. Existing Vite large-chunk warnings remain; no suites or review agents.
A targeted Playwright batch using real authenticated VPS directory reads checked
filtering, click selection returned to the form, double-click/Up navigation,
keyboard selection, disabled unsubmitted-path confirmation and explicit row-draft
replacement. No project/session was created. Captures at 1440x900 and 390x844 plus
the dark theme were inspected; the narrow dialog had no horizontal overflow.
Screenshots/results are in the external workflow directory.

The updated Chrome preview remains open for the user's manual acceptance. It uses
local built dev assets at the existing test-only same-site frontend origin, with
normal real HTTPS API/cookies/CORS to the VPS. It does not host the frontend on the
VPS or replace the production Tauri installation. Native desktop/WebView2/mobile
acceptance and Windows/UNC/native-dialog runtime paths remain separate checks.

## New folder context menu and parent-row follow-up

Implemented the user's subsequent requests:
- Right-click the loaded directory's contents → **New folder**, also accessible
  through **Directory actions** and Shift+F10. Reuses the design-system Menu within
  the existing dialog. A focused inline name editor requires explicit Create;
  no mutation occurs merely by opening the menu/editor.
- Authenticated `POST /api/filesystem/directories` takes `{parent,name}`, creates
  one child with `os.Mkdir`, never overwrites or creates recursive parents, and
  returns 201 with the lexical path. Existing client authentication, CSRF,
  contract/identity and OS permission gates remain. Body, name and combined path
  are bounded; traversal components/control characters and unsupported Windows
  names are refused. Duplicate/permission/missing-parent responses are recoverable.
- The picker opens the newly created directory after confirmation. During the
  request, conflicting controls are disabled. Unknown transport outcomes show a
  refresh action and do not automatically repeat the mutation.
- Removed the toolbar Up arrow. The server-reported parent is the first **...**
  row, unaffected by name filtering/sorting, opened by double click or Enter.
  Filesystem roots have no parent row. Removed visible selected-folder text;
  the summary retains counts only.

Changed existing implementation files: `engine/directory_browser.go`,
`engine/api.go`, `engine/sqlite_api.go`, `clients/desktop/src/engine.ts`,
`DirectoryBrowser.tsx`, its CSS, `product.md`, and `engine/README.md`.
Scoped pre-edit copies: `C:/Users/Samuel/AppData/Local/Temp/klm-new-folder-Mfv7KJ`.
Only the changed directory-browser Go file was formatted. Release Linux/Windows
Go builds and TypeScript passed; frontend TypeScript/Vite build also passed.
The focused mechanical UI detector returned no findings. No suites, integration
or unit tests, review agents, commits or pushes were performed.

The checkout now activates SQLite contract 3, unlike the earlier installed VPS
release. The parent initially held deployment rather than silently migrate it.
The user then explicitly authorized clean reinstallation and loss of KLM data.
Stopped only `klm.service`, removed exactly the live
`/var/lib/klm/.config/klm/engine` directory after path/owner checks, recreated it
0700 for `klm`, and installed the checked Linux binary with SHA-256
`8fa390fddb90dcb5e9e850cc69bdef12de7f9a8ebb83988d1953f57224edaa34`.
A fresh `engine.sqlite3` initialized. The existing runtime user/service, HTTPS
route, DNS and TLS proxy were retained; unrelated project paths/sites, Windows
installation and local engine were not changed. API remains 127.0.0.1:7331 only.
Existing historical protected backups were not restored or automatically purged.

Playwright was re-paired normally over HTTPS using a fresh one-use code held only
in memory. Authenticated health returned 200 with **contract 3**, zero running
turns/graphs. A targeted real-UI smoke opened the right-click menu, submitted a
random temporary child under `/var/lib/klm`, observed **201**, and confirmed that
the picker opened it. Double-clicking the first `...` row returned to Home. That
exact temporary folder was removed using nonrecursive `rmdir` as `klm`, then Home
was refreshed. No project was registered and no agent turn was requested. No
unexpected post-pairing problem occurred. Permission/duplicate/Windows edge paths
remain manual checks; compilation is not their runtime acceptance.

Inspected captures of the real context menu and inline editor. The updated
Playwright window remains open at the new-folder editor for the user to validate;
its **Create** button has not been submitted for the default manual draft.
Artifacts and metadata-only result are in the same external temp directory
(`playwright.log`, `ui-result.json`, `new-folder-context-menu.png`,
`new-folder-editor.png`). Pairing values and cookie/credential values were not
saved in those artifacts. No Windows installer or production engine update was
performed as part of this follow-up.

## Rename and empty-folder deletion follow-up

The user subsequently requested renaming and deleting folders. Added **Rename**
and **Delete** to a child folder's context menu and the selected child's toolbar
menu. Right-click selects the actual target; blank space and the `...` row never
inherit a previously selected folder's destructive actions. Keyboard Context Menu /
Shift+F10 uses the focused row; touch can select a row and use Directory actions.

The inline name editor requires explicit Rename. The Delete confirmation names
the folder and its complete path, defaults focus to Cancel, and only submits after
Delete folder is pressed. Cancel/Escape performs no mutation. Both reuse the
picker's loading, stale-result, cancellation and unknown-delivery guards. Successful
rename reloads the parent and selects the renamed child; deletion reloads the parent
and clears selection. The status announces confirmed changes, without reintroducing
selected-folder labels. Singular count copy was corrected as part of that status.

New authenticated routes on `/api/filesystem/directories`:
- PATCH `{parent,name,newName}` -> 200 `{name,path}`. Atomic no-replace rename
  (`renameat2(RENAME_NOREPLACE)` on Linux, `MoveFile` on Windows). No unsafe
  check-then-overwrite fallback; unsupported filesystems/platforms return 501.
- DELETE `{parent,name,confirmation:name}` -> 200 `{path}`. Directory-only
  nonrecursive `rmdir`/`RemoveDirectory`, never a generic file-removal fallback or
  recursive deletion. Nonempty directories return 409 without deleting contents.

Same body/path/name limits and client-auth, CSRF, contract/engine-identity gates
apply. Directory links, filesystem roots, home/its ancestors, the entire engine
storage tree/its ancestors and registered project roots/ancestors (including
removed registrations) are refused. Destination protection prevents silently
retargeting a missing registered project. Existing aliases and missing leaves'
existing ancestors are resolved. Unverifiable protected paths fail closed.
Registration and these single-step mutations share app.mu; these integrity guards
are not containment against external filesystem/mount changes. Existing project
metadata is not rewritten. Ordinary project subfolders remain user-owned content.

Implementation added `engine/directory_mutations.go` and Linux/Windows/other OS
helpers; extended both route gates, frontend API wrappers, the picker/CSS and
product/API documentation. External pre-edit copies and artifacts are under
`C:/Users/Samuel/AppData/Local/Temp/klm-folder-actions-5Ce9eR`.
Linux and Windows amd64 Go builds passed. The first frontend build found a Button
ref typing error; it was corrected without changing the generic primitive, and the
TypeScript/Vite build passed. The focused UI detector returned no findings.

Updated only the VPS binary, preserving the exact existing SQLite file and all
live data/paired credentials. Verified no running sessions or graphs before the
restart. Installed SHA-256:
`b310d8bc4922ef9d9962ebd19a77706581d9cccdf698578c997fe0c30b44687b`.
Receipt `/etc/klm/folder-actions-update.json` records unchanged SQLite inode and
before/after counts (0 projects, 1 general conversation). Previous binary is in
`/opt/klm/backups/folder-actions-5Ce9eR`; there was no data wipe/restore, proxy/DNS/
TLS change, Windows installer update or local production-engine update.
API remains 127.0.0.1:7331 behind the existing HTTPS ingress.

The first preview attachment found its old CDP endpoint closed, before any test
mutation. Reopened the dedicated persistent Chrome profile on CDP **19234**; its
existing pairing remained valid, so no new code was issued. A targeted real-UI
smoke created one randomly named temporary folder (201), renamed it (200), opened
and cancelled deletion without sending DELETE, then explicitly confirmed deletion
(200). The temporary folder is gone. Only that test-owned folder was renamed/
deleted; other home entries were preserved. No project registration or agent turn
was requested. No unexpected engine/UI problem occurred in that smoke.

Inspected the real context menu and confirmation at desktop size and the
confirmation at 375px in dark theme. Original viewport/theme were restored. The
picker remains open at Home for manual validation. Test response/error listeners
were removed after completion, so recoverable manual errors do not freeze the
preview; only its local asset route stays active. `ui-result.json`,
`folder-actions-menu.png`, `delete-confirmation.png`, and `delete-mobile-dark.png`
contain the bounded evidence, without pairing secrets or cookies.
No suites, committed integration/unit tests, adversarial reviews, subagents,
sessions, graphs, commits or pushes were requested/performed. Windows runtime,
permissions/conflicts, protected-path errors and other edge cases remain human
acceptance checks; successful compilation/happy-path smoke is not their acceptance.
