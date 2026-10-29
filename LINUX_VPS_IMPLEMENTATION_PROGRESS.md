# Linux VPS implementation progress

## Delivered code / distribution slice

- Linux `start`/`stop`, local readiness, `pair` and `revoke-clients` reuse existing
  dispatch, startup, auth and durable storage. The deterministic per-data-dir Unix
  socket requires a real same-effective-UID 0700 data directory, a same-owner
  0600 socket and kernel `SO_PEERCRED` validation on both ends. Requests/responses
  are bounded newline messages with deadlines. Pairing material is not logged.
- Only the engine-lock owner reclaims a same-owner stale socket after bounded
  connection refusal; other errors, symlinks, files and live endpoints fail
  closed. Close checks socket identity before unlinking, without Go's implicit
  pathname unlink. Long paths fail actionably. Logs reject symlinks/nonregular,
  foreign-owned or non-0600 files.
- Linux worker/restart helpers use `Setsid` and detached stdin/log output. Readiness
  waits for completed initialization and HTTP startup, and rejects closing,
  stopped or storage-error states. Startup timeout cleans up only its newly
  launched child. Stop acknowledges then shared lifecycle waits for lock release.
- `KLM_API_BIND_HOST=127.0.0.1` (or literal `::1`) applies to the first public API
  listen. Invalid overrides fail; unset retains `0.0.0.0`, including Windows.
  Private loopback capabilities and auth/origin/CSRF/TLS policy are unchanged.
- Added a dedicated-user, loopback-first Linux systemd oneshot template, bootstrap
  environment example and installation/ingress/recovery documentation. The unit
  deliberately has **no automatic crash supervision**; `active (exited)` is not
  readiness. Existing HTTP Restart stays in the service cgroup.
- Windows pipe/control implementation is unchanged. Shared hooks preserve its
  existing log/timeout behavior. Other non-Windows/non-Linux control remains an
  unsupported stub. Frontend/Tauri, Windows installers, Linux graph containment,
  ai-jail, harness installers/provider login and unrelated SQLite work were not
  changed by this slice.

## Scoped baseline / exact implementer files

External baseline (captured before edits):
`C:\Users\Samuel\AppData\Local\Temp\klm-linux-implementation-UcdYP2\baseline`
with `baseline-manifest.json` and original dirty `git-status.txt` alongside it.
The main.go baseline includes the preexisting SQLite edits; do not compare the
whole dirty checkout against HEAD as this implementation's delta.

Changed existing files:
- `engine/cli.go`
- `engine/cli_unix.go`
- `engine/main.go` (directory validation, first-bind selection, Linux readiness
  field/gate only; preserved preexisting startup/storage implementation)

New files:
- `engine/cli_linux.go`
- `engine/cli_support_other.go`
- `engine/api_listen.go`
- `distribution/linux/klm.service`
- `distribution/linux/klm.env.example`
- `distribution/linux/README.md`
- `LINUX_VPS_IMPLEMENTATION_PROGRESS.md`

## Lightweight checks actually performed

Formatted only the six touched Go files. From `engine`, both release builds
completed successfully (no dev tag, no tests):

```sh
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o C:/Users/Samuel/AppData/Local/Temp/klm-linux-implementation-UcdYP2/klm-linux-amd64 .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o C:/Users/Samuel/AppData/Local/Temp/klm-linux-implementation-UcdYP2/klm-windows-amd64.exe .
```

Environment overrides were command-scoped, not persisted. Linux artifact:
`C:\Users\Samuel\AppData\Local\Temp\klm-linux-implementation-UcdYP2\klm-linux-amd64`

SHA-256:
Original candidate: `710148404cb7cc17f944271028cab5f9b074b61f7ea03b7e0ebcfcc084e8caf4`.

After the review correction, the parent repeated the two release builds successfully.
Final Linux artifact SHA-256:
`14da93252f053366909be93d83072b802793ca5c3e37f365a909a60e9748c415`.
No full suites or integration tests ran. See deployment evidence below; builds alone
are not complete runtime acceptance.

## Single final native review and correction

The implementer could not delegate because its child tool catalog lacked the
subagent tool. The parent invoked the authorized third/final reviewer instead;
findings are in `LINUX_VPS_REVIEW.md`. Exactly three native subagent roles were
used (planner, implementer, reviewer), with no independent sessions or extra review.

One P2 predeployment finding: startup could republish readiness after local stop
acceptance. The parent applied the focused fix because the finished implementer
cannot be resumed through the available tool: close the monotonic shutdown latch
under `a.mu` before acknowledging stop, and check that same latch under `a.mu`
before startup publishes readiness. Both release targets compiled after the fix.
No second review or claim of adversarial/runtime concurrency validation is made.

## Private VPS deployment evidence

The parent installed the final hash-verified binary at `/opt/klm/bin/klm` on
`187.77.43.129` (`srv1763338`), using a new nonroot `klm` user and the supplied
systemd unit, enabled for boot. Persistent data is under
`/var/lib/klm/.config/klm/engine`. The actual API socket is **127.0.0.1:7331 only**.
Directory mode 0700 and socket/log modes 0600 were inspected, owned by `klm`.
Repeated start reported already running; systemd stop removed the control socket,
and subsequent start succeeded. Unpaired loopback `/api/auth/session` returned 401.

Existing ingress: Traefik v3.3 in Docker, edge network IP `172.18.0.2`, host gateway
`172.18.0.1`; file-provider config at `/opt/codemob-infra/traefik/dynamic`, resolver
`letsencrypt` already uses Cloudflare DNS challenge. The existing configured token
queried the exact `codemob.com.br` zone successfully without being printed/copied.
The requested `cubiqklm.codemob.com.br` record was absent at discovery.
## HTTPS ingress deployed

Created only the `A` record `cubiqklm.codemob.com.br -> 187.77.43.129`, TTL 300,
**DNS only / not Cloudflare-proxied**. No zone-wide SSL/firewall setting was changed.
DNS-only avoids putting browser challenge pages in front of the API. The API is
`https://cubiqklm.codemob.com.br`; no frontend was installed on the VPS.

Added `/opt/codemob-infra/traefik/dynamic/klm-api.yml` and a dedicated public CA
certificate file `klm-origin-ca.pem`, without restarting existing ingress/sites.
Traefik's existing `letsencrypt` resolver issued the publicly trusted certificate
through its configured Cloudflare DNS challenge. Both public DNS HTTPS and direct
origin HTTPS returned 401 for the unpaired auth-session request, with normal TLS
verification (no certificate-ignore flags). Exact test-origin preflight returned
204 with credentials and the expected Access-Control-Allow-Origin.

Because Docker Traefik cannot use host loopback, a dedicated `klm-api-hop` Caddy
container terminates a second validated TLS connection on **172.18.0.1:17443**,
then proxies to 127.0.0.1:7331. Traefik pins that private CA and server name; Caddy
accepts only Traefik's 172.18.0.2 source and overwrites forwarding metadata from
its own TLS connection. Engine trust is restricted to the actual immediate
loopback TLS proxy, 127.0.0.1/32. A single UFW rule permits only TCP from
172.18.0.2 to 172.18.0.1:17443 on bridge br-04ae1ae1c87c. No public 7331/17443
allow rule was added; public port 443 was already the existing ingress.

The new proxy uses `caddy:2.10.2-alpine`, digest
`sha256:4c6e91c6ed0e2fa03efd5b44747b625fec79bc9cd06ac5235a779726618e530d`,
nonroot UID/GID 65534, readonly root filesystem/config/cert mounts, no-new-privileges,
128 MiB/128-process limits and restart-unless-stopped. The image's file capability
requires NET_BIND_SERVICE in its bounding set; all other capabilities are dropped.
This is ingress hardening, **not agent sandboxing or ai-jail**.

Private ingress config/certificates are protected under `/etc/klm/ingress`;
DNS creation receipt is `/etc/klm/dns-receipt.json`. Public certificates renew via
existing Traefik ACME. The private hop leaf was issued for **365 days** and needs
explicit renewal/redeployment before expiry; it has no automatic renewal job.
The CA lifetime is 3650 days. Caddy's readonly default-storage housekeeping emits
startup log warnings/errors; proxy configuration validation and actual traffic
worked. No post-pairing repair of that nonfunctional housekeeping was attempted.

## Initial Chrome pairing — completed; stopped on first blocker

Used installed Chrome through Playwright, fresh dedicated profile. Only UI assets
were fulfilled from the local Windows Tauri server (127.0.0.1:7332) under the
**test-only** same-site origin `https://klm-playwright.codemob.com.br`. That hostname
has no newly registered DNS/frontend server. All requests to the API used real
public DNS, normal TLS, actual CORS/CSRF and the ordinary browser cookie flow.
This is not native Tauri/WebView2 or production frontend-hosting acceptance.

Generated one code through SSH via the same-user protected Linux `klm pair`, kept
it in process memory, and submitted through the existing UI with client name
`VPS Chrome validation`. No pairing values or cookie values were saved in logs,
documentation, command arguments, traces or screenshots.

Observed, in order:
- Pair redemption: **200**.
- Existing client's authenticated `/api/auth/session`: **200**.
- Initial `/api/state`: **200**, zero projects/sessions and all three harnesses
  unavailable, each reporting executable not found.
- The normal updates SSE connection had also returned **200** by this boundary.
- Retained API-host cookie metadata: Secure, HttpOnly, SameSite=Strict; cookie
  value was never recorded.

**First blocker:** Pi, OpenCode and Codex are not installed/available to the engine
user on the VPS; Pi also requires Node.js. Stopped at the initial inventory, made
Chrome offline to prevent further automatic requests, and did not reload, create
projects/sessions, send agent prompts, install harnesses, log into providers,
introduce ai-jail, or make further infrastructure repairs. A metadata-only result
is in the external build folder's `pairing-result.json`.

## Remaining human acceptance / next requested work

- Decide the next harness installation/onboarding slice. No provider connection,
  actual model execution, Linux managed-adapter support or sandbox acceptance is
  implied by successful connectivity/pairing.
- Native Windows Tauri remote pairing and Credential Manager persistence remain
  separate human checks. Use Engine URL `https://cubiqklm.codemob.com.br` and a new
  code from the Linux engine user; do not reuse the consumed Chrome code.
- The single review finding is corrected and rebuilt; detached HTTP restart,
  crash/stale recovery and startup/stop concurrency remain runtime-unvalidated.
  Reload retention was intentionally not tested after the first blocker.
- Operational follow-up: private hop certificate renewal, readonly Caddy
  housekeeping and fixed Docker peer-IP coupling are documented limitations,
  not claims of automatic lifecycle/certificate management.

Scoped rollback: remove only the new hostname's DNS record/Traefik route/CA file,
stop/remove `klm-api-hop`, delete only the exact KLM UFW rule, and stop/disable
`klm.service`. Preserve `/var/lib/klm` data and all existing workloads/configs.
Do not delete the runtime home or modify unrelated DNS/firewall rules.
