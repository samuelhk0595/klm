# 0.6.2 — VPS operational deployment

## Latest update — automation webhooks and JSON diagnostics (2026-11-11)

The user requested updating the running VPS engine in this conversation, human
message `3ee37c1f-71dd-4b24-a310-61dc78311075`.

The release Linux amd64 engine was rebuilt from the current shared checkout and
published to the existing `/opt/klm/bin/klm` installation on `srv1763338`.
This includes `issue_comment` automation routing and the specific JSON rejection
diagnostics. The desktop-side save serializer/revision updates require a separate
updated desktop build; this deployment updates the remote engine.

| Item | Value |
| --- | --- |
| Version / target | 0.6.2 / linux-amd64, no dev tags, CGO disabled |
| Source HEAD | `1766fb18fe2f111cbd320a61c99787f7f28154a6` plus current uncommitted changes |
| Engine source fingerprint | `9afa367e35ca6d4e1e6b1f8a62c2d7c2727bf75eff2dc306cb22ad3bcb6fd73b` |
| Binary SHA-256 | `47ca16f457bc18b4edb8e6ec5ebcfda47995a9021e799f103c1cf1352b5b4321` |
| Archive SHA-256 | `a14d59955a2113bdc0e7d828f0046e0a60638d5240d81e465e82ec5375753289` |
| Maintenance window, UTC | 2026-11-11 15:56:09–15:56:24 |
| Effective worker checked | PID 2090186, UID 999, matching `/proc/<pid>/exe` hash |
| Staging and evidence | `/opt/klm/releases/0.6.2-20261111-automation-json-01` |
| Private backup | `/opt/klm/backups/before-0.6.2-20261111-automation-json-01` |

There were zero running sessions, graphs and automations immediately before
maintenance and after restart. Existing distribution assets were checked against
the prior manifest before replacement. The unit was stopped and its cgroup drained
before backing up the installation, consistent engine data/auth and service config.
The backup includes any engine-local MCP key when present; none existed at this
backup point. No live database copy, manual SQL edits or data restore was performed.

Verification completed:

- Existing `klm.service` active (exited), with its actual worker independently checked.
- Listener remains `127.0.0.1:7331`; no proxy or service configuration changes.
- SQLite `quick_check` returned `ok` before replacement and after restart; schema 4
  retained, and the existing paired-client identity/credential fingerprint matched.
- Data/database ownership and modes remain `klm:klm` 0700/0600. Root-private backup
  directory/archive modes are 0700/0600.
- Local and public HTTPS `/api/auth/session` returned the expected unauthenticated
  401. No service-journal error lines appeared in the deployment window.
- Installed manifests and deployment/verification results remain in private staging.

Creation/editing through the real desktop, GitHub delivery and harness behavior
remain human acceptance checks. No new pairing credential, test automation,
session, commit, push or PR was created.

## Result

**Linux engine published at 2026-11-01 20:24:08 UTC; frontend publication is blocked.**
This establishes operational availability, not acceptance of graphs, automations,
native children, provider behavior or protected MCP credentials.

The user requested publication for manual tests in **Graph on Linux**, human message
`6d7cca02-d7fc-4bd4-a0a1-be39c22ba997`. Its agent forwarded execution to this
implementation conversation; the actual request was checked in bounded history.
No new sessions, subagents, graphs, commits, pushes or GitHub releases were created.

## Existing installation confirmed

- SSH host identity verification retained; hostname `srv1763338`, Linux x86_64.
- Existing unit: `klm.service`, user `klm` (UID 999), working directory `/opt/klm/bin`.
- Binary: `/opt/klm/bin/klm`. Prompts/skills use its existing sibling directories.
- Data: `/var/lib/klm/.config/klm/engine`, directory 0700, database 0600, owned by klm.
- HOME `/var/lib/klm`; XDG_CONFIG_HOME `/var/lib/klm/.config`.
- Existing gh PATH/drop-in preserved, including `/var/lib/klm/.local/bin`.
- `/etc/klm/engine.env`, unit and drop-in content hashes were unchanged after deploy.
- Public API remained loopback `127.0.0.1:7331` behind the existing Caddy/Traefik TLS
  route. No proxy, port, firewall, certificate or unrelated service changes.

## Artifact actually installed

Release build from the shared uncommitted implementation checkout:

```text
GOOS=linux GOARCH=amd64 CGO_ENABLED=0
go build -trimpath -ldflags="-X main.engineVersion=0.6.2" -o <private-temp>/engine/klm .
```

No dev tag. Binary, complete prompts and internal skills were packaged together.

| Item | Value |
| --- | --- |
| Version / target | 0.6.2 / linux-amd64 |
| Source HEAD | `d1857b06df6c58284b8304f22bf05103fb47b392` plus the uncommitted 0.6.2 implementation |
| Engine source fingerprint | `217c12d89d312d78da1cbf515814c1ba43be0fe18e4b3e93b5298a2a47b26be8` |
| Engine package | `KLM-Engine-0.6.2-linux-amd64.tar.gz` |
| Package SHA-256 | `76f3cce7237b45f84eccf4f26bdaf5e16464629ec67706e6a3759ab581a762cd` |
| Binary SHA-256 | `1982e62c960efb7181b946e58dfc76e02f029494d6f4068af1e1adb88f2147b6` |
| VPS staging / evidence | `/opt/klm/releases/0.6.2-20261101-manual-01` |
| Effective worker observed | PID 444654, UID 999, `/opt/klm/bin/klm __worker` |

Transferred archive hash, each installed binary/asset hash, and the effective
`/proc/<worker>/exe` hash all matched. The source manifest stays in the approved
local temporary deployment directory. `installed-manifest.json` and
`deployment-result.json` are retained in the private VPS staging directory.

## Controlled maintenance and data preservation

Before maintenance there were **zero active chat turns and zero active graphs**,
but three old automation rows still marked running and two queued steering inputs.
The service cgroup contained only the engine worker, with no native workloads.

The existing unit was stopped and complete cgroup drainage confirmed **before**
backup. No active-SQLite file copy was used. The stopped database passed
`PRAGMA quick_check`. The original installation, user data/auth and service
configuration were preserved in a root-private backup before replacing assets.

After restart, normal engine recovery marked the three old automation rows
`execution_failed` (four failures total including the earlier one) and paused the
two pending messages. These were ordinary startup transitions, not manual SQL
edits, replay or deletion. The initial restart snapshot briefly showed one recovery
turn; the final smoke showed **no active turns**. No graph/automation/provider test
task was dispatched; normal configured restart recovery was allowed to run.

Preservation checks confirmed:

- All five paired-client identities/credential hashes preserved; no new pairing
  code or verification credential was issued.
- Database identity and schema 4 unchanged; all prior project/session IDs retained.
- Access policy, environment, unit/drop-in and gh PATH preserved.
- Database integrity check still `ok`; original private owner/modes retained.

## Operational smoke only

- `klm.service` is **active (exited)**, the expected oneshot unit state; the real
  worker and its actual binary hash were checked separately.
- Local and public HTTPS `/api/auth/session` returned **401 without credentials**,
  as expected. This checks availability/auth enforcement, not an authenticated login.
- API listener remained **127.0.0.1:7331**.
- New log section had one listening record and no panic, fatal, storage-failure or
  generic-error categories. Native stderr/auth contents and raw user logs were not
  printed. No provider/graph functionality was tested.
- Current process HOME/XDG/PATH matched the existing installation.

API for the existing client connection: **https://cubiqklm.codemob.com.br**.
This is an API URL, not a newly hosted frontend.

## Frontend: prepared, not published

Inspection found no existing KLM frontend service/static mount on this VPS; the
only KLM container was the API TLS hop. The old test hostname
`klm-playwright.codemob.com.br` no longer resolved. `klm.codemob.com.br` responded
403 through Cloudflare and no accessible existing publication destination for it
was established. The previous publisher's harness was unavailable for consultation.

A production Vite build succeeded with
`VITE_ENGINE_URL=https://cubiqklm.codemob.com.br`. Assets are prepared, with integrity
verified, at:

```text
/opt/klm/releases/0.6.2-20261101-manual-01/KLM-Web-0.6.2-vps-assets.tar.gz
SHA-256: 05bd4c148c234d00159e3ce2a7cdbc6de4273176ad809ff10521dd9c3908156e
```

`frontend-artifact.json` explicitly records `published:false`. No new web server,
hostname or topology was invented, and the local Windows installation was not
updated. **Composer/settings/editor changes require the updated frontend before
their manual acceptance.** Completion needs the actual frontend destination and
access/authorization to update it, or a separately requested client distribution.

## Backup and rollback available

Private root directory (0700):

```text
/opt/klm/backups/before-0.6.2-20261101-manual-01
  installation/             # entire previous binary/assets tree
  data-and-auth.tar.gz       # stopped consistent data + user configuration/auth
  service-config.tar.gz      # existing unit/drop-in/environment
  backup-manifest.json       # hashes, identity/pairing checks and captured paths
```

Archives are 0600. Data/auth includes `.config` (the complete engine data tree),
`.local/share/opencode`, `.pi`, `.codex` and `.klm` as present. No credentials or
private archive contents were copied into this repository/report.

Previous binary (0.6.1) SHA-256:
`78ad765d6aa5e55268ad3e4c3956cf73f5545a401176b4267e6b5c91d8b13bdd`.

Rollback was **not executed**. For intentional recovery: stop the existing unit,
preserve a fresh consistent copy of post-deployment data/auth, verify the protected
backup, restore the prior binary/assets with their recorded permissions, and decide
explicitly whether a point-in-time data restore is needed for compatibility. Do not
blindly overwrite current SQLite/user auth or discard changes made during testing.
Start only the existing unit, then repeat process/hash/HTTPS smoke checks.

There was **no MCP key at this backup point**, and the deploy smoke did not create
one. Once Linux MCP tests create ciphertext, preserve its matching
`mcp-secrets/credentials.key` with that database in every new backup. Never mix
post-deployment ciphertext with a missing/older key, and never regenerate a key to
make a rollback open the database. Windows DPAPI remains a different backend.

## Pending

1. Supply/authorize the accessible existing frontend publication destination.
2. Perform the manual Linux/Windows feature scenarios in
   `V0_6_2_IMPLEMENTATION_PROGRESS.md`. Operational deployment is not functional
   acceptance; `runtimeValidated` remains false for graph adapters.
