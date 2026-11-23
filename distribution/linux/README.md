# Linux engine deployment

Release Linux amd64 supports the existing `klm start`, `stop`, `pair`, and
`revoke-clients` commands. The daemon and restart helper detach from the terminal;
control uses a same-UID Unix socket, not HTTP or a PID file. Other non-Windows
platforms remain unsupported. This is engine/API delivery only: **no frontend
server, Linux desktop/installer, automatic provider login or ai-jail**.
Managed chats and graphs now share the nonroot subreaper/pidfd lifetime mechanism.
Graph admission probes the supervisor; each node/Terminal/Git command has an owned
domain and confirmed termination remains required. Protected MCP secret storage
uses an engine-specific AES-256-GCM key. Focused native/runtime and real messaging
acceptance remains pending in `../../V0_6_2_IMPLEMENTATION_PROGRESS.md`. This adds
no cgroup delegation, syscall filter, namespace or host privilege. See
`../../LINUX_ORDINARY_TURN_SUPPORT_HANDOFF.md` before publishing the candidate.

## Install privately first

The following is a **new installation template**, not an automated installer.
Inspect existing users, paths, units, listeners and ingress before applying it.
Do not replace another workload or copy Windows engine data. Do not publish a
route until protected local startup succeeds and API binding has been inspected.

1. Verify the uploaded binary's SHA-256 against the implementer's final artifact.
   Install that release (not a dev-tag build) as root-owned `/opt/klm/bin/klm`,
   mode 0755. Internal `engine/prompts/*.md` are embedded during `go build`
   (including `-trimpath`), so the binary needs no source checkout or prompts
   directory. Optional editable prompts go in `/opt/klm/bin/prompts`; an explicit
   `KLM_PROMPTS_DIR` is authoritative and must contain every required prompt.
   Empty/unreadable prompt files fail preparation, never omit the guidance.
   Embedding covers prompts only: the general agent's internal skill still needs
   `engine/skills` packaged as `/opt/klm/bin/skills`, or an absolute `KLM_SKILLS_DIR`.
   No Go/Node installation is needed for the engine itself on the VPS.
2. If absent, create a dedicated system account:
   `useradd --system --user-group --home-dir /var/lib/klm --shell /usr/sbin/nologin klm`.
   Inspect an existing `klm` account for compatibility instead of changing it blindly.
3. Create `/var/lib/klm`, `/var/lib/klm/.config`, `/var/lib/klm/.config/klm`, and
   `/var/lib/klm/.config/klm/engine`, owned by `klm:klm`, mode 0700. Ensure parent
   paths are trusted, root/runtime-owned and not writable by unrelated users; do
   not install through symlinked or shared writable parents. Existing insecure
   data directories are refused, not silently chmodded. The final data directory
   must not be a symlink. `control.sock` must fit the Linux 107-byte path limit.
4. Install `klm.service` to `/etc/systemd/system/klm.service` (root:root, 0644).
   Create `/etc/klm` (root:root, 0700), copy and adapt `klm.env.example` to
   `/etc/klm/engine.env` (root:root, 0600). Set only the needed exact frontend
   origins and the **actual immediate TLS-terminating proxy** CIDRs. Never trust
   all networks. Keep this operational file out of source control. The unit fixes
   `HOME`, `XDG_CONFIG_HOME` and `KLM_API_BIND_HOST=127.0.0.1` before first startup.
5. Run `systemctl daemon-reload`, then `systemctl start klm`. Inspect the CLI
   startup result via `journalctl -u klm`; worker diagnostics are in the protected
   `/var/lib/klm/.config/klm/engine/engine.log`. Stop on a startup/storage error;
   do not reset/migrate a live DB by hand to work around it.
6. Before ingress, inspect `ss -ltnp '( sport = :7331 )'` as administrator: expect
   **only 127.0.0.1:7331**, not wildcard IPv4/IPv6. Inspect the actual engine UID,
   data-dir mode 0700 and socket/log modes 0600. Validate repeated CLI start with
   the same environment, then `systemctl stop klm` / `systemctl start klm` for a
   graceful lifecycle. Enable with `systemctl enable klm` only after success.

`KLM_API_BIND_HOST` accepts only unset (legacy `0.0.0.0` default), `127.0.0.1`, or
`::1`; invalid values fail before listening. The default Windows binding is
unchanged. This setting does not affect private authenticated loopback bridges.
An existing saved access policy overrides origin/proxy environment bootstrap;
use authenticated Engine settings and restart rather than deleting auth storage.

## Local administration and lifecycle

Always use systemctl for normal deployed start/stop/recovery. Run pairing and
revocation as the service user, with the service's data environment, for example:

```sh
runuser -u klm -- env HOME=/var/lib/klm XDG_CONFIG_HOME=/var/lib/klm/.config /opt/klm/bin/klm pair
runuser -u klm -- env HOME=/var/lib/klm XDG_CONFIG_HOME=/var/lib/klm/.config /opt/klm/bin/klm revoke-clients
```

`sudo klm pair` targets root's different data directory and is not recovery for
this engine. Keep one-use pairing codes ephemeral: no command arguments, logs,
traces, screenshots or documentation. Pairing does not bypass HTTPS, exact-origin,
CSRF or existing credential storage rules. Revocation reports failure if durable
auth persistence fails. Never change socket permissions to permit other users.

For a direct `start` readiness check or CLI recovery, supply **all** the unit's
HOME/XDG/bind and bootstrap origin/proxy values, not just HOME; `start` may launch
an engine if it is down. Prefer `systemctl restart klm` for recovery. `start`
waits for protected local readiness; a timed-out Linux startup terminates only
the child that invocation launched. `stop` acknowledges locally, then waits for
the data lock to release after shutdown. Only the engine-lock owner reclaims a
same-owner 0600 stale socket after connection refusal; files, symlinks, live or
ambiguous endpoints are never unlinked. Socket replacement is identity-checked
on close. Do not manually unlink a live socket or use the lock file's diagnostic
PID as authority to kill a process.

The intentionally small systemd unit is `Type=oneshot` / `RemainAfterExit=yes`,
with `Restart=no` and `KillMode=control-group`. It permits the existing detached
HTTP Restart helper to stop/start within the same cgroup. It provides boot start
and bounded group cleanup on service stop, **not automatic crash supervision**.
`active (exited)` is not proof of readiness or liveness. After a crash, use
`systemctl restart klm` and inspect startup/log output. Direct `klm stop` leaves
the oneshot unit active; reconcile it with `systemctl stop klm`. No new execution
is automatically replayed after restart. `NoNewPrivileges` is basic hardening,
not a harness sandbox or secure containment implementation.

## HTTPS ingress prerequisites

The API hostname is `https://cubiqklm.codemob.com.br`. Reuse the existing 443 router;
do not take over occupied ports, open 7331, change global firewall/TLS policy, or
restart unrelated workloads. DNS/Cloudflare credentials, certificate ownership
and actual ingress topology must be established separately by the deployer.

- TLS must be valid from browser to edge and authenticated/validated from edge to
  origin (Full strict or equivalent). No Flexible TLS or certificate-check bypass.
- The immediate trusted proxy must itself verify/terminate TLS, preserve external
  Host, remove incoming `Forwarded`/`X-Forwarded-*`, and set exactly one
  `X-Forwarded-Proto: https` based on that verified connection. Trust only its
  actual TCP peer CIDR at the engine. A generic cleartext forwarding hop must not
  assert HTTPS merely because a request supplied that header.
- A Docker proxy cannot reach host loopback as its own `127.0.0.1`. Establish a
  deliberately private TLS-aware host hop, or report a topology blocker. Do not
  switch the engine back to wildcard binding to solve container reachability.
- Preserve body bytes, methods and relevant headers; disable API caching and
  challenge pages, stream SSE without buffering with suitable idle timeouts, and
  keep upload limits compatible with engine limits. Do not log credentials,
  Cookies, Authorization or pairing bodies. Route only public API traffic, never
  private capability listeners or the control socket. Serve no UI assets here.
- Before pairing, verify DNS, TLS, the unauthenticated `/api/auth/session`
  rejection and exact-origin transport. Browser cookies remain host-only,
  Secure/HttpOnly/SameSite=Strict: localhost Chrome is cross-site with this API.
  Native Windows Tauri uses its existing native transport. Compatible same-site
  browser validation and the first-post-pairing-error stop rule are in
  `LINUX_VPS_IMPLEMENTATION_PLAN.md`; no browser fixture/frontend changes ship here.

Rollback only this hostname's route/configuration and this unit. Stop the engine,
retain protected user data and do not remove the runtime home or other sites.
Build evidence and outstanding acceptance are in
`LINUX_VPS_IMPLEMENTATION_PROGRESS.md`. No deployment is implied by these files.

## 0.6.2 runtime configuration and recovery

- Supported target remains **Linux amd64 glibc**. No ARM64 or Alpine/musl install
  support, Linux Activity Board credential backend or engine-hosted frontend.
- Graphs require kernel/user-policy support for subreaper, pidfd_open and
  pidfd_send_signal. A failed helper handshake reports unavailable capability;
  do not bypass the preflight. Fork/Join and worktrees retain existing policies.
  Pi, OpenCode and Codex still require separate real-model lifecycle validation.
- Terminal YAML uses `shell: powershell`, `bash` or `sh`. Older missing fields
  remain PowerShell; select sh/Bash explicitly before using Linux shell syntax.
  Payload data is available at `$KLM_GRAPH_PAYLOAD_FILE`; no jq/Python is required
  for commands that do not parse JSON. There is no executable fallback.
- Install/login to `gh` as the engine user only when requested. Harnesses resolve
  absolute GH_CONFIG_DIR from the engine's original environment: explicit
  GH_CONFIG_DIR, then XDG_CONFIG_HOME/gh, then ~/.config/gh. Tokens and GH_HOST keep
  gh's native precedence. Do not copy credentials to temporary OpenCode profiles.
- OpenCode inspection reports module/open/schema/query/timeout categories and a
  diagnostic ID with the effective Node path/version. Sanitized engine logs record
  duration/category/exit and stderr length, not native stderr, auth data or rows.
- Native Pi delegation still needs compatible pi-subagents 0.70.0. Foreground SDK
  children and local background runners stay descendants of the parent's supervisor;
  the parent waits for native settlement and closes the domain at the boundary.
  This does not enable both native and engine-owned delegation paths together.
- Native picking is optional: a graphical host environment and Zenity or KDialog
  must already exist. It opens on the engine host, never a remote browser. No
  display/backend on a VPS means use the normal remote directory browser.

### MCP key and backups

MCP manual secrets, stdio secret environment and OAuth tokens/refresh share an
authenticated AES-256-GCM envelope. The random 32-byte key is stored separately at
`<engine-data-dir>/mcp-secrets/credentials.key`. The key directory must be owned by
the engine user with mode 0700; the regular, single-link key must be mode 0600.
Creation is exclusive/atomic, and existing links, unsafe ownership/modes and invalid
key length are refused rather than repaired or overwritten. Dev/release directories
have independent keys. No graphical keyring/D-Bus session is required.

Back up the consistent SQLite database **and this matching key** using equivalent
private access controls, then restore them together. Ordinary configuration exports
and run snapshots contain only credential IDs, not the key/plaintext. A missing key
with dependent credentials is never regenerated. Missing/incorrect key, unsupported
envelope and integrity failures leave ciphertext intact; restore the matching key
or explicitly disconnect/re-authenticate as appropriate. DPAPI blobs from Windows
are not portable and need new authentication on Linux. Do not include the key in
logs, harness environments or release artifacts. This protects storage at rest;
it does not isolate secrets from a process already running as the engine user.
