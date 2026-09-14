# Engine / Desktop implementation progress

Plan: `ENGINE_DESKTOP_IMPLEMENTATION_PLAN.md`. Implemented 2026-09-14.
Status: P1–P5 implemented and packaged; awaiting human runtime validation.

Existing local changes (graph engine, authoring, adapters, UI and documentation)
were present before this delivery and are being preserved.

| Stage | Status | Evidence / remaining work |
| --- | --- | --- |
| P1 — CLI | Implemented / built | Public start/stop, detached worker, per-user local named pipe and graceful shutdown. Stop waits on the actual pipe peer's process handle, without stored-PID killing. Final Go build succeeded. |
| P2 — Network | Implemented / built | Public API on 7331, distinct Host/CORS rules; private bridge unchanged. Shared frontend endpoint for fetch/SSE; TypeScript/Vite build succeeded. |
| P3 — Tauri | Implemented / built | Tauri 2, native window, tray and single instance; Rust release build and NSIS bundle succeeded. |
| P4 — Web / Focus | Implemented / built | Embedded Vite assets served by Rust on 7332; native Focus button and desktop/web presentation. HTTP LAN UUID compatibility included. |
| P5 — Installers | Implemented / built | Two independent NSIS installers generated. Engine includes adjacent prompts, user PATH, hidden HKCU Run launcher and data-preserving uninstall. Desktop uses Tauri's NSIS/WebView2 packaging. |

## Build evidence

- `go build -o <temporary path>/klm.exe .` succeeded for initial CLI/API integration.
- `npm run desktop:build` succeeded: TypeScript project build, Vite production build,
  Rust release compilation and NSIS bundle. Rust 1.93.1; Tauri 2.11.5.
- `distribution/windows/build-engine.ps1` succeeded, including the final
  `go build -trimpath` after the live-process shutdown wait was added, resource
  staging and NSIS 3.11 compilation. That initial installer was 6,704,483 bytes;
  subsequent installer fixes are recorded below.
- `environment.ps1` passed PowerShell's AST syntax parser. This was parsing only;
  install/uninstall/start actions were not executed.
- Vite emitted chunk-size notices for diagram dependencies; the build succeeded.

## Artifacts

- `dist/KLM-Engine-0.1.0-x64-setup.exe`
- `clients/desktop/src-tauri/target/release/bundle/nsis/KLM Desktop_0.1.0_x64-setup.exe`
- `dist/engine/klm.exe` and `dist/engine/prompts/`
- `clients/desktop/src-tauri/target/release/klm-desktop.exe`

Rebuild instructions: `distribution/windows/README.md`.

## Implementation notes

- Public API binds `0.0.0.0:7331`; `loopbackHost` and the private bridge keep their
  local protections. Public CORS is specific to Focus/Tauri/Vite clients.
- CLI operations serialize through `cli.lock`; the existing `engine.lock` still
  owns runtime exclusivity. Port binding precedes state recovery/work scheduling.
- The control pipe is scoped by user SID/data directory, permits that user and
  explicitly denies network logons. Shutdown uses the existing cancellation path.
- Focus embeds the same Vite output in Rust, serves only embedded assets on 7332,
  and requests graceful server shutdown with a bounded wait on desktop Exit.
- Desktop/native presentation uses Tauri context. Browser HTTP/SSE resolve the
  page hostname; UUID compatibility uses `crypto.getRandomValues` on HTTP LAN pages.
- Documentation updated in AGENTS, product, CONTEXT, root/engine/desktop READMEs
  and the original plan status. Existing graph/harness/UI work remains preserved.

## Pending human acceptance

1. If an older foreground engine is running, stop it first. Install engine, open a
   new terminal and run `klm start` twice: confirm one worker and survival after
   closing the terminal. Confirm existing projects/sessions remain available.
2. Install/open desktop; check native full-window layout, engine connection and
   **Session log → Focus → Side agent**. Focus opens the default browser with the
   gradient and draggable central shell; each client's drafts remain independent.
3. Close hides; tray Open and a second shortcut restore the same instance. While
   hidden, web reload works. Exit stops 7332 while the engine stays available.
4. On another device, open `http://IP:7332`, verify state/SSE and authoring/mentions.
   Confirm Windows Firewall authorization as needed. Native directory selection
   and execution still belong to the engine computer.
5. `klm stop` gracefully ends work and the worker; repeat stop is harmless. Start
   permits client reconnection and records interrupted runs without replay.
6. New Windows login starts only engine, without a permanent console. Confirm
   Git/harness discovery and graph prompt resources outside the checkout.
7. Uninstall desktop: only desktop/web stop. Uninstall engine: worker stops,
   engine Run/PATH registration is removed, and `%APPDATA%/klm/engine` remains.
8. Check fixed-port collisions: 7331 gives an engine startup error; occupied 7332
   reports Focus unavailable and never opens that unrelated service.

No full suites, integration tests, model/graph runs or review agents were used.
Neither installer was executed. Existing engine/client processes, login settings,
firewall rules and user data were not modified as part of build validation.

## Installer hang follow-up — 2026-09-14

Human installation stalled after `Created uninstaller`. Inspection found the
engine InstallDir/AddedPath and HKCU Run registration already written, but no
engine worker/listener or engine log. The installer subsequently exited before
its blocked call could be inspected directly.

The probable blocking step is the following synchronous `WM_SETTINGCHANGE`
broadcast. Its five-second timeout applies separately to recipient windows;
multiple slow windows can stall the installer before engine startup.

- Removed synchronous broadcasts from both install and uninstall sections.
- The environment helper now launches a self-contained hidden notification process
  through ShellExecute without waiting or inheriting nsExec output pipes. Uninstall
  can remove the helper file immediately because the child carries its command.
- Added explicit progress messages for stop, PATH/login setup, startup and readiness,
  plus timeouts for captured helper/CLI commands.
- PowerShell AST syntax check and NSIS recompilation succeeded. Updated artifact:
  `dist/KLM-Engine-0.1.0-x64-setup.exe` (6,703,437 bytes).
- Installation has not been rerun by the agent; human retry remains the acceptance
  check. Existing partial registration can be reused by rerunning the new installer.
