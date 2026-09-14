# Windows releases

The workflow in .github/workflows/release-windows.yml builds the independent
Windows x64 engine and desktop installers on every push to main, including merges
and direct pushes. It can also be run manually on main from GitHub Actions.

The workflow installs Node 22, Go from engine/go.mod, and stable Rust on Windows
2022, runs npm ci, and calls distribution/windows/build-release.ps1.
Tauri downloads its NSIS packaging tools; the engine reuses NSIS.
The runner provides Visual C++ build tools. No installed KLM engine is required.

Both installers use the same numeric version: the major/minor from
clients/desktop/src-tauri/tauri.conf.json, with its patch plus the workflow run
number. For example, base 0.1.0 and run 12 produce 0.1.12. Re-running a job retains
that version. Keep this workflow's identity and increase the base major/minor
when starting a new version series. Do not lower the base version.

Only after both builds succeed does the publish job create a draft release at
the triggering commit, upload both installers plus SHA256SUMS.txt, and publish
it. Failed uploads leave a draft; a retry completes it. Already published releases
are preserved on retries. Runs are not canceled by later pushes.
Publication uses the built-in GITHUB_TOKEN with contents: write only in the
publish job; no personal token is needed. Repository policies must permit this.
Tag rules must allow the workflow to create v* tags.

Engine upgrades retain the install location, registry identity and user data.
The existing installer asks before interrupting work, calls klm stop, replaces
the binaries and starts the installed version. Silent engine installation
declines the interruption confirmation. Desktop keeps its Tauri identifier and
per-user NSIS installation; the client and engine have independent lifecycles.
Installers are currently unsigned.

## Local build

From the repository root, with Go, Rust/MSVC, Node and frontend dependencies
installed, run:

    .\distribution\windows\build-release.ps1 -Version 0.1.1

Outputs are in dist/release. The command builds packages only; it never runs an
installer or stops the installed engine. Use a fresh output directory when
collecting artifacts from repeated local builds with different versions.
CI always builds in a fresh checkout.

The workflow and its source prerequisites must be committed together: desktop
Tauri sources/icons/Cargo.lock, frontend package lock, engine Windows CLI/runtime
and prompts, and distribution scripts. Local untracked files are not available
on GitHub runners. Do not include build outputs, credentials or user data.

Validation still required in GitHub: the first hosted run, release downloads,
and an upgrade over an installed desktop and engine with data preserved.

References: [Tauri Windows packaging](https://v2.tauri.app/distribute/windows-installer/)
and [GitHub release CLI](https://cli.github.com/manual/gh_release_create).
