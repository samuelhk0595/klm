---
reviewed: false
last_revision: "2026-11-19T21:26:35Z"
---

# Project Git identities: validation and VPS engine publication

Read when operating or rolling back the November 19, 2026 Git identity update.
The user authorized validation/publication in message
`6248f438-ec9c-4161-b216-7a5d742b7a2c`. The engine is published. The user then chose
local manual installation and requested Windows release installers for both the
desktop and engine; those artifacts are generated below. The separate web bundle
was not deployed to a new frontend host.

## Contract and validation

[Project Git identities](engine/README.md#project-git-identities) defines aliases,
default selection, metadata-only APIs, protected credentials, command syntax and
lifecycle. The UI belongs to KLM Project settings, not Werbert Settings. Shared
authentication remains the default without identities; configured managed work
never falls back to the shared account. This is not same-OS-user isolation.

The existing focused selection was run once on Windows and Linux:

```text
^Test(ProjectGit|RunnerGitIdentityDurableContext|RunnerColdReplyReconciliation)
```

- Windows: `go test -tags dev -run '<selection>' -count=1 -v -timeout=120s .`
  passed, **14 top-level tests, 14.414 seconds**.
- Actual Linux VPS, user `klm`: cross-compiled dev test executable passed the same
  selection, **19.055 seconds**, with no selected skips. It used an empty environment
  plus private temporary HOME/XDG/GH/Git/skills locations. No real shared token or
  provider credential was copied or used.
- Coverage includes HTTP/auth/storage/migration, candidate/replacement preservation,
  default/revocation/race semantics, synthetic concurrent aliases across projects,
  explicit failure/no fallback, reference-only runner/jobs, Codex shell policy and
  cold reply reconciliation. Actual Git and its private credential helper ran
  against a local repository with synthetic credentials, without GitHub traffic.
- Linux amd64 release build (`CGO_ENABLED=0`) succeeded. Client `npm run build`
  passed TypeScript/Vite. A subsequent Vite-only packaging build targeted
  `https://cubiqklm.codemob.com.br`; this was a different endpoint configuration,
  not another typecheck/test campaign.
- The coordinator's earlier settings-component/API browser checks used synthetic
  accounts and disposable data. Their receipt/screenshots remain in
  `temp_doc/project-git-identities-progress.md`. That fixture is not full-shell or
  real-token acceptance; it was not rerun during publication.
- No full suites, adversarial review, provider turn, real GitHub token verification
  or real authenticated Git HTTPS transfer was performed.

## Published engine

Existing VPS: `srv1763338`, `187.77.43.129`. Updated `/opt/klm/bin/klm`, prompts and
internal skills. Services restarted at **20:53:39 UTC**; verification completed at
**20:53:50 UTC**.

| Item | Result |
| --- | --- |
| Binary SHA-256 | `6eff385eab70294d0921e9a1aad28119d173fe5df7976b6f615d060f16268d51` |
| Engine source fingerprint | `ecefb6c0d5a86b51a5c739ac0bef0d2617c7513659963dfa016dfffff67d83be` |
| Database | Schema **6 → 7**; identity `a966288835000cc591796081871f8fd3` retained |
| Original records | All project/session/paired-client IDs and W Factory/Ticket/Assignment/operation IDs retained |
| Access | Existing policy hashes and connection retained; no new pair/connect mutation |
| Actual worker | PID `3249528`; executable hash matched the candidate |
| Services | KLM, W backend and W frontend active; W connected/available, protocol 1 |
| Health | Authenticated KLM `ok`, zero running sessions/graphs; W proxied health succeeded |
| New API | Authenticated lists succeeded for all three existing projects; zero identities, preserving shared mode |
| Integrity | Both active SQLite databases passed `PRAGMA quick_check` |
| Other components | W binary/frontend image and memory container hashes/identities retained |

The existing public API `https://cubiqklm.codemob.com.br/api/auth/session` returned
the expected unauthenticated 401 after publication. Listeners, TLS hop and routing
were preserved; no auth exemption was added.

### Maintenance recovery

Docker's frontend stop ended with status 137. The first strict stopped-service
assertion halted before any candidate installation/migration; both backends still
ran old bytes. Confirmed the stopped frontend, unchanged old KLM hash and empty
backup directory before continuing the same maintenance. All owned cgroups were
confirmed drained before consistent backups. The continued deployment succeeded;
the initial assertion failure is not counted as a passing check.

## Client publication boundary

No KLM frontend service/static mount exists on this VPS. `klm-api-hop` serves the
API; Werbert's frontend is a different application. Inspection of current systemd
units, listeners and KLM proxy routes confirmed this. `https://klm.codemob.com.br`
currently returns a frontend, but its separate publication destination/access was
not established.

Prepared bundle: `/opt/klm-git-identities-release-20261119-01/client.tar.gz`, SHA-256
`96146bd80e659c09d97edd0de9b2af4dab9882279b32455909048f8ac2377f38`.
It targets the VPS API and is recorded as `client_published:false`. The Git identities
screen needs an updated client. The user chose a local Windows installer rather
than another web publication. No new hosting topology or automatic local
installation was performed.

### Windows release installers for manual installation

Both **0.7.0 x64** installers were generated successfully:

```text
C:\Users\Samuel\Documents\Projects\Personal\klm\clients\desktop\src-tauri\target\release\bundle\nsis\KLM Desktop_0.7.0_x64-setup.exe
C:\Users\Samuel\Documents\Projects\Personal\klm\dist\KLM-Engine-0.7.0-x64-setup.exe
```

| Installer | SHA-256 |
| --- | --- |
| Desktop | `1050f7b6b7b73685096427e09da57a2a6c22c632bfd57b3933efb60604cf0fef` |
| Engine | `9dc16c499471b4c4cf8d5e07b592802bd8318ce49d33bb1742c4dd0f9f0648fc` |

Desktop packaging used the existing Tauri NSIS flow, a temporary version-0.7.0
config override, `VITE_DESKTOP_VERSION=0.7.0` and locked Rust dependencies. Its
TypeScript/Vite, Rust release and NSIS stages succeeded. Executable metadata
confirms ProductVersion/FileVersion 0.7.0. `VITE_ENGINE_URL` was unset for this
native build, retaining the standard saved-engine selection rather than forcing
the VPS as a new default. Source version manifests were not rewritten.

Engine packaging used `distribution/windows/build-engine.ps1 -Version 0.7.0`,
Windows/amd64/CGO-disabled release compilation with version injection, complete
prompt/skill assets and NSIS. Installer metadata confirms 0.7.0; Go build metadata
confirms Windows amd64. File presence and both hashes were checked. Neither
installer was executed; installation/upgrade and installed-client behavior remain
the user's manual validation.

## Evidence and rollback

Private release: `/opt/klm-git-identities-release-20261119-01` (0700), containing
source/artifact manifests, deployment receipt, Linux focused-test log/receipt,
candidate assets and prepared client. Local copies are in
`C:\Users\Samuel\AppData\Local\Temp\opencode\git-identities-vps-20261119`.
Source HEAD `cb9aa691e979eb42c36cb649d9b77d59480b7114` alone does not identify the
dirty/untracked implementation; the source manifest does. No commit/push occurred.

Stopped private backups under `backup/`:

- `installations.tar.gz`: prior KLM/W binaries/assets.
- `service-config.tar.gz`: units/drop-ins, environments and proxy configuration.
- `active-data-and-auth.tar.gz`: KLM configuration/data and matching private key,
  W Foreman store/binding, existing account authentication and skill trees.

Hashes are in `deployment-result.json`. Previous KLM binary:
`397913ba56520b08f8af73d409c44f37706a8eb250e402ab4b7eb87f452250e9`.

The old engine cannot simply replace the binary over schema 7. For an authorized
rollback, stop frontend/W/K, confirm drainage and preserve new post-update state.
Restore matching old binaries/config and the stopped schema-6 database with its
credential key/auth/artifacts as one coherent point in time, retaining owner/modes.
Old data restoration loses newer work and requires an explicit recovery decision.
Start K/W/frontend and verify integrity, identity, auth and actual readiness.
No rollback or historical work replay was performed.

Remaining acceptance: updated full client/Tauri navigation, live user-authorized
tokens/repository permissions, native Pi/OpenCode/Codex launches and graph/runner/
Team continuation in the installed environment. Synthetic checks and API readiness
do not establish those paths.
