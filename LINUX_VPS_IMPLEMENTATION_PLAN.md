# Linux engine control and Ubuntu VPS deployment plan

## Scope and execution order

This is the planning artifact, not evidence of implementation or deployment. Read `AGENTS.md` and all of `product.md`; inspected the current CLI/startup, Unix process/lock helpers, authentication/access settings, SQLite auth integration, desktop endpoint/auth transport, and Tauri bundled asset server. No software, remote infrastructure, application code, or test suites were changed/run by the planner.

Use exactly the user-requested three-agent workflow: **planner → implementer → third/final reviewer → implementer fixes → parent deployment/client tests**. The reviewer returns actionable findings to the implementer; do not create additional agents/review rounds. Review the actual final patch, not the entire dirty checkout. Parent must wait for resulting fixes before testing.

Deliver only:

1. Secure Linux local `start`, `stop`, readiness, `pair`, and `revoke-clients`, retaining Windows behavior.
2. A dedicated nonroot engine installation on the verified Ubuntu VPS and an HTTPS **API** at `https://cubiqklm.codemob.com.br`, using available Cloudflare/DNS/proxy access.
3. A paired local client and the initial Chrome/Playwright deployment check. **Stop at the first encountered post-pairing problem** and report it; do not fix onward, onboard providers, or install harnesses automatically.

Tauri on Windows retains its bundled frontend and local Focus server. It may select the remote engine. Do not serve frontend assets from the engine/VPS, copy the frontend to the VPS, alter default frontend ownership, or add a new hosting architecture. No Linux graph containment, ai-jail, harness installation, provider onboarding, speculative platform compatibility, commits, or pushes.

## Confirmed baseline and remaining discovery

- Correct SSH target: `root@187.77.43.129`, host `srv1763338`, key `C:/Users/Samuel/.ssh/codemob_vps`, existing verified `known_hosts`. Use strict host-key checking. **Never use `76.13.167.100`, accept its changed key, or bypass host verification.**
- Previously inspected: Ubuntu 26.04 x86_64; Docker installed; no node/npm/go/klm found; approximately 60 GB disk free / 7.7 GiB RAM. Public listeners already use 22, 80, 443, 8090, 8091. These facts must be refreshed narrowly before mutation, not assumed to describe the current ingress topology.
- A prior `GOOS=linux GOARCH=amd64 CGO_ENABLED=0` build compiled, but `engine/cli_unix.go` currently returns Windows-only errors. Compilation is not runtime support or deployment acceptance.
- `engine/cli.go` already serializes start/stop using `cli.lock`, checks local `ready`, launches `__worker`, and waits for local shutdown plus `engine.lock` release. Windows uses a user-only, network-denying named pipe in `cli_windows.go`.
- `main.go` currently reserves **`0.0.0.0:7331` before local control**. A deployment must prevent exposure from the first bind, not close the port after startup.
- Auth already supplies one-use five-minute pairing codes, persisted client credential hashes, full-access paired devices, exact origins, CSRF, Strict cookies, and immediate-proxy TLS trust. Reuse it unchanged; never mint pairing codes over public HTTP or enable the debug bypass.
- The checkout has extensive preexisting tracked/untracked SQLite work, including `main.go`, `api.go`, `client_auth.go`, `engine_access_settings.go`, `go.mod`, `go.sum`, and many `sqlite_*.go` files. Save a scoped baseline outside the repo, reread files immediately before editing, and use precise edits. Do not revert, stage wholesale, migrate, or fix unrelated work. Startup/auth may change while this workflow runs: inspect their **current** wiring rather than restoring a previously read version.
- Still unknown: owner/configuration of 80/443, origin TLS certificate availability, Cloudflare zone/DNS/proxy rights, existing domain records, available local browser/proxy tooling, and running local desktop/frontend version. Discover only necessary service/configuration metadata. Never dump environment files, tokens, private keys, Docker secrets, broad credential directories, or unrelated application configurations.

## 1. Implement the smallest secure Linux local control

### Expected source boundaries

- Add `engine/cli_linux.go` with `//go:build linux`; change the existing stub tag to `!windows && !linux`. Keep other non-Windows platforms explicitly unsupported rather than importing Linux-only credential APIs under `!windows`.
- Reuse `lockDataDir` in `engine/platform_unix.go` and existing common command dispatch, auth methods, storage and graceful shutdown. Do not duplicate startup/state recovery or introduce a second control HTTP server.
- Prefer no edits to `cli_windows.go`. If a shared lifecycle/readiness correction is necessary, make its Windows behavior impact explicit and compile the Windows target once.
- A narrow API-listen helper and a one-line replacement of the current `net.Listen` address are justified to avoid upstream exposure; leave unrelated shared startup/SQLite code untouched.
- Add Linux distribution artifacts under `distribution/linux/` plus focused setup/evidence documentation. Do not edit Windows installers or frontend code for deployment convenience.

### Unix socket security and protocol

1. Use a deterministic Unix-domain socket in the actual engine data directory (for example `control.sock`), keeping dev/release directories isolated. Validate socket path length and fail actionably if too long; do not silently fall back to a public/global `/tmp` socket.
2. Before use, verify the final data directory is a real directory, owned by the effective runtime UID, and inaccessible to group/other users (0700). Do not follow a substituted final symlink or silently accept an insecure existing directory. Installation must also put it below trusted parent directories.
3. Bind while holding the existing exclusive `engine.lock`; create socket permission 0600 inside that protected directory. Use `Lstat` and owner/type checks on existing paths. Never unlink a symlink, regular file, directory, foreign-owned socket, or live server.
4. Safe stale cleanup: only the engine-lock owner may remove an existing same-owner socket after a bounded local connection check demonstrates a stale/refused endpoint. Permission errors, timeouts and ambiguous live connections are failures, not permission to unlink. Cleanup on close must not remove a replacement endpoint (retain/check identity or use equivalent ownership-safe listener cleanup).
5. Verify Linux peer credentials (`SO_PEERCRED`, using existing `golang.org/x/sys/unix` dependency) on accepted connections and client-side connections where applicable; require the same effective UID. Administrator recovery runs commands **as the engine user**, not with root's different data directory. Do not persist/use a PID as authority for control or killing.
6. Implement the existing bounded newline protocol: `ready`, `stop`, `pair`, `revoke-clients`. Match Windows request/response meaning and CLI messages. Use bounded input/output, deadlines and per-connection close; reject unknown commands. `pair` calls `a.clientAuth.issuePair()`, revocation calls `revoke("")`, shutdown closes its channel once. Never log codes/credentials or include them in errors.
7. Preserve auth persistence errors and readiness failure for closing/storage-error states. Inspect startup ordering: currently control is created before some initialization and `server.Serve`. Do not let Linux `ready` announce success before required startup completes, or after a failed initializer; use the smallest explicit readiness gate if needed. Readiness remains protected local control, not an unauthenticated public health endpoint. No new public `ready` CLI command is required.

### Worker lifetime

- Linux background worker/restart helper must use a new session (`Setsid`) with no terminal/stdin dependency; redirect stdout/stderr to the existing protected engine log. Preserve start readiness timeout, repeated-start behavior, fixed port, stop's engine-lock wait, and no automatic interrupted-work replay.
- Keep response handling graceful: acknowledge stop, let shared shutdown drain resources, and let `stopEngine` wait for lock release. Never kill by a PID read from a stale file. For startup failure/timeouts, ensure the newly launched worker does not remain an untracked late-starting daemon; cleanup may target only the child actually launched by that invocation or its verified control peer, with bounded waiting.
- Keep `__restart` functional: helper must survive old-worker exit and use the same runtime user, data path, access settings and local control. Avoid a service configuration that kills the restart helper or starts an unmanaged competing worker.

### Close the upstream-exposure window

Add an explicitly configured API bind address/host setting (for example `KLM_API_BIND_HOST=127.0.0.1`) with existing `0.0.0.0` default preserved, especially on Windows. Accept only the documented literal loopback override(s), reject malformed/unsupported values, and apply it at the **first** public API `net.Listen`. No hostname resolution, implicit fallback after invalid configuration, dynamic port, or change to private bridge binding/`loopbackHost`.

The VPS unit must set this value before its first start. Confirm with socket inspection that only loopback:7331 is listening. Firewall isolation is defense in depth, not a substitute for the bind setting. If the discovered existing proxy is in Docker, its `127.0.0.1` is not the host: establish a deliberately private reachable hop (for example a narrowly exposed host-side TLS proxy behind its existing ingress). Do not change the engine back to public bind just to make a container reach it. If no safe minimal route fits the existing ingress, report the topology blocker instead of taking over ports.

## 2. Lightweight implementer checks and requested final review

- Format only changed Go files. Build the actual current working-tree engine once for release Linux amd64 with CGO disabled, outputting outside the repository, e.g. from `engine`: `$env:GOOS='linux'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'; go build -o <absolute-temp-path>/klm-linux-amd64 .` Restore prior environment afterwards. Do not use `-tags dev`, auth bypass, or a stale earlier artifact.
- Build release Windows once **if shared Go code changed** to catch cross-platform compilation/import regressions. No frontend check is needed without frontend edits. Do not run full suites, add integration/complex tests, or repeat overlapping checks by default.
- Record build results and artifact SHA-256, plus exactly which files were changed. A preexisting compilation failure belongs to its owner; do not overwrite concurrent SQLite work to make this patch build.
- Third/final reviewer checks only the delivered change and deployment assets for same-UID socket ownership, stale cleanup, bounds, startup readiness, clean daemon/restart/stop behavior, no public upstream window, actual immediate TLS proxy trust, no credential leakage, Windows preservation, and preservation of Linux execution gates. Return findings to implementer for fixes before parent deployment/client tests. This is the explicitly requested review, not permission for extra agents/suites.

## 3. Deploy without disrupting existing workloads

### Establish the private installation first

1. Through the verified SSH target, narrowly identify listeners/process/container ownership, existing virtual-host routing and certificate mechanism, firewall topology, and any existing KLM paths/user/unit. Confirm the intended hostname does not already serve another workload. Do not restart Docker, replace an ingress stack, flush firewall rules, or take ports 80/443.
2. Transfer the compiled binary to a temporary protected path and compare SHA-256 before publication. Build locally; no Go/Node/npm/harness install is necessary on this VPS.
3. Create a dedicated `klm` system user/group with nologin shell and persistent home `/var/lib/klm` (only if not already present and compatible). Use a root-owned immutable-to-runtime binary location such as `/opt/klm/bin/klm`, 0755, and a root-owned `/usr/local/bin/klm` link if desired. Make `/var/lib/klm` 0700 owned by `klm`. Do not copy Windows engine data or provider credentials.
4. Fix `HOME=/var/lib/klm` and `XDG_CONFIG_HOME=/var/lib/klm/.config` consistently in service and administrative invocations. Current release `engineDataDir()` then resolves to `/var/lib/klm/.config/klm/engine`; create protected directories accordingly. Document that `sudo klm pair` would target root's data, whereas `runuser -u klm -- env HOME=/var/lib/klm XDG_CONFIG_HOME=/var/lib/klm/.config /opt/klm/bin/klm pair` targets this engine. Apply the same environment for start/stop/recovery.
5. Root-owned service configuration sets `KLM_API_BIND_HOST=127.0.0.1`, the chosen exact frontend origin, and only the **actual immediate** TLS proxy address CIDR(s). Example loopback-host proxy: `KLM_TRUSTED_TLS_PROXIES=127.0.0.1/32`; add `::1/128` only if actually used. Never trust `0.0.0.0/0`, all Docker networks by default, or Cloudflare public ranges when the engine's TCP peer is a local proxy. No debug auth flag. Keep configuration/logs out of source control when operational/sensitive.
6. Access settings saved through UI/storage override bootstrap environment after restart. Respect the actual current JSON/SQLite storage implementation; do not edit/reset a live DB or delete paired-client storage to force configuration.

### Minimal systemd choice that preserves existing CLI/restart semantics

Use a small `Type=oneshot`, `RemainAfterExit=yes` unit with `User=klm`, `Group=klm`, the explicit environment above, `UMask=0077`, `WorkingDirectory=/opt/klm/bin`, `ExecStart=/opt/klm/bin/klm start`, `ExecStop=/opt/klm/bin/klm stop`, `KillMode=control-group`, and bounded generous timeouts covering CLI lock/readiness/shutdown waits (e.g. 150s start / 180s stop). `Restart=no`; enable at boot only after private startup succeeds. Basic compatible hardening such as `NoNewPrivileges=yes` is appropriate; do not advertise this as harness sandbox/containment or add filesystem restrictions speculatively.

This unit intentionally manages a background CLI worker and its restart helpers in one cgroup. It **does not** provide automatic crash recovery or reliable worker-liveness status: `active (exited)` alone is not proof the engine is ready. Use protected local readiness (`klm start` checks it) and document explicit recovery after a crash. A direct `klm stop` leaves the oneshot unit active; `systemctl stop klm` reconciles service state. Use systemctl for normal deployed lifecycle, run as-user CLI for pairing/revocation, and document `systemctl restart klm` for recovery. Do not substitute `Type=simple __worker` plus an unmanaged detached restart helper or unconditional restart policy that defeats `klm stop`.

Before public routing, verify one private startup, idempotent start/readiness, correct runtime UID/data/control-socket modes, loopback-only API binding, and graceful stop/start. Verify only this engine's processes and do not exercise harnesses. These are focused deployment smoke checks, not a test suite. If first startup/storage fails, keep the route unpublished and report that blocker.

### DNS, Cloudflare, TLS and ingress

1. Determine whether the task has an explicitly available scoped Cloudflare integration/token for the `codemob.com.br` zone. Use only that integration or an explicitly provided credential; never hunt broadly through user files or print its value. If absent, ask the user/zone owner to create/update the DNS record and provide only the necessary confirmation. Continue preparatory private installation where safe, but do not claim HTTPS completion.
2. Desired record: `cubiqklm.codemob.com.br` → `187.77.43.129`, proxied when supported by the existing topology. Do not publish a stale AAAA record to an unrelated/unprotected IPv6 endpoint. Preserve unrelated records and zone-wide settings.
3. Add only this hostname to the **existing** 443 virtual-host router, forwarding its API traffic to the private engine hop. Maintain valid browser-to-Cloudflare TLS and authenticated, validated TLS from Cloudflare to the origin (Full strict or equivalent per-host configuration). No Flexible mode or origin `-k` workaround. A Cloudflare origin certificate alone is not browser-trusted for direct bypass; that distinction must be explicit.
4. If certificate issuance/installation requires missing zone rights or a human DNS challenge, give the exact missing step. Avoid changing zone-wide SSL settings without knowing their effect on other workloads. Never print private keys or replace other sites' certificates.
5. The final TLS-aware immediate proxy must preserve the external `Host`, overwrite/remove client-supplied forwarding metadata, and send exactly one `X-Forwarded-Proto: https` based on its own verified TLS connection. Never blindly preserve incoming XFP, or claim verified HTTPS on an insecure arbitrary origin HTTP hop. Trust only its actual source IP at the engine.
6. Preserve methods, request body bytes and relevant headers. Disable API caching/challenge/interstitial behavior for this hostname's API; keep SSE unbuffered with suitable idle/read timeouts and bounded upload handling compatible with engine limits. Do not log Authorization, Cookie, pairing request bodies, or credential-bearing headers. Do not route private harness/Library/control sockets or nonpublic listener ports.
7. If available within this hostname's scope, restrict direct origin access to Cloudflare peers / authenticated origin pulls without changing unrelated sites. Regardless, upstream 7331 must never be publicly reachable on IPv4/IPv6; no security claim based solely on hiding the IP.
8. Before pairing, confirm DNS target, certificate validation, an unpaired `/api/auth/session` rejection (normally 401 through valid HTTPS, not a public success), and exact-origin OPTIONS/CSRF transport using no credential. Health/version are authenticated; do not weaken them for proxy health checks. Confirm no inadvertent UI hosting on the API hostname.

Rollback is scoped: remove/disable only the new virtual-host route/DNS entry and unit, stop this engine, and retain its protected data for recovery. Restore any backed-up hostname-specific ingress configuration. Never delete other users' data or restart unrelated services. Do not remove the nonroot runtime home as an automatic rollback step.

## 4. Chrome/Playwright pairing strategy — no cross-site cookie assumption

### Why ordinary localhost Chrome is not the remote browser test

Browser transport in `clients/desktop/src/clientAuth.ts` uses `credentials: include`; engine cookies are host-only, HttpOnly, Secure, **SameSite=Strict**. `http://localhost:7332` or `http://127.0.0.1:5173` → `https://cubiqklm.codemob.com.br` is cross-site. CORS allowlisting, a successful pairing POST, third-party-cookie toggles, or `ignoreHTTPSErrors` do not make Strict cookies work. The code may be consumed before the follow-up session check fails. Do not attempt this knowingly, relax SameSite, inject a bearer into browser storage, fake `IS_DESKTOP`, disable web security, or substitute an API-only bearer check for client validation.

### Compatible local-bundle browser route

Use an HTTPS frontend origin on the same schemeful registrable site, for example `https://klm-playwright.codemob.com.br`, with remote engine `https://cubiqklm.codemob.com.br`. These are **same-site but cross-origin**: the API cookie remains API-host-only; exact frontend Origin/CORS/CSRF still apply. No `Domain` cookie is required. Use the existing client Engine URL selection; its special default for `https://klm.codemob.com.br` points at a different engine, so explicitly select the requested endpoint. Inspect whether `VITE_ENGINE_URL` fixes the existing bundle before pairing.

For the initial controlled Playwright run, the least infrastructure-dependent option is a **test-only browser route** for that exact synthetic frontend hostname: fulfill only its UI document/assets from the already running local Tauri Focus server (`http://127.0.0.1:7332`, or the actual dev Focus port). Forward only asset requests to that local bundled-UI source, not cookies/Authorization or API traffic. Browser navigation retains the HTTPS same-site test origin; allowlist exactly that frontend origin in engine bootstrap settings. Leave every request to `https://cubiqklm.codemob.com.br` untouched so real DNS, TLS, proxy, cookie persistence, CORS, CSRF, API and SSE are exercised. Use a fresh browser context, no auth injection, and no global TLS-ignore setting. This requires no frontend deployment or DNS record for the synthetic frontend hostname.

State clearly that such routing is a **test fixture**, not evidence of real production frontend HTTPS delivery or a replacement for Tauri hosting. If routing is unavailable or browser origin/cookie behavior is incompatible, do not pretend it validates the client: use an already available local HTTPS reverse proxy with a trusted certificate for the exact same-site frontend origin, still proxying the Windows Tauri bundle locally, or report the missing local HTTPS/tooling prerequisite. Do not install proxy/certificate software or change DNS for another frontend host speculatively. Ordinary future Focus browser use needs real compatible same-site HTTPS frontend delivery; native Tauri remote access does not depend on it.

### Native Tauri is separate acceptance

Native Windows Tauri uses its bundled app origin, `X-KLM-Native: 1` on pairing, engine-scoped Windows Credential Manager storage, and in-memory Authorization bearer transport (cookies omitted), including fetch-based SSE. Built-in Tauri origins remain allowed. It does not depend on cross-site Strict cookies. Ordinary Chrome/Playwright cannot exercise WebView2/Tauri IPC or Credential Manager; **native Tauri pairing, persistence and restart behavior require human Windows testing** unless a separately supported native automation bridge is already available. Do not claim the Chrome test validates native credential storage.

## 5. Parent's short first deployment test and stop rule

After reviewer findings have been fixed and pre-pair prerequisites passed:

1. Start local Windows Tauri so the bundled Focus assets remain available. Open Chrome through available Playwright tooling in a fresh context using the compatible strategy above. Set/select `https://cubiqklm.codemob.com.br` before issuing a code. Do not expose a fixed local Vite override as a mysterious pairing failure.
2. Obtain one code through verified SSH using `runuser -u klm` and the service's exact HOME/XDG_CONFIG_HOME. Keep the value ephemeral; do not put it into documentation, screenshots, logs/traces, committed scripts, command-line arguments, or the final report. Fill the existing password field; use a recognizable test client name. Submit once.
3. The first post-pairing check is the real client's existing `/api/auth/session` confirmation and mounted workspace. Confirm a retained Secure/HttpOnly/Strict API-host cookie without recording its value. Then check initial inventory/settings and establishment of its normal updates SSE. Reload once to confirm the same paired client remains usable, only if no earlier failure occurred.
4. If an initial page/inventory response already shows no available harnesses, report it as the first surfaced execution blocker; do not install Node/harnesses or authenticate providers. If the client otherwise works, report connectivity/pairing success and note execution is not validated. Do not create projects/sessions or send prompts merely to seek another problem. Any later execution objective needs an explicit next-step decision.
5. **At the first post-pairing problem, stop further scenario execution and mutations.** Record the first failing operation, sanitized response/error, frontend origin, engine URL, whether pairing redemption succeeded, and whether cookie/session/SSE retention was reached. Do not repeatedly re-pair, repair cookie policy, change auth, install dependencies, or proceed to another problem. Safety cleanup may close the browser/test routes and protect the deployment; retain evidence without secret material. Do not revoke the working client merely as an additional test in this initial run.
6. Return a concise status: implemented/deployed items, lightweight build(s) actually run, HTTPS/DNS/private-bind result, pairing/check boundary reached, first problem or exact human blocker, and native Tauri manual acceptance outstanding. Never equate compilation, service `active` state, API-only bearer success, or a simulated frontend origin with complete product runtime acceptance.

## Current blockers to public/client completion

- Cloudflare zone/DNS/proxy and origin certificate rights are **not yet known**. If unavailable, the human must arrange the single API hostname and valid origin TLS; do not invent access.
- Existing ingress ownership/container reachability must be discovered before choosing its private engine hop; 80/443 are already occupied and must not be commandeered.
- Local Playwright/Chrome and bundled Focus source must actually be available. A real local HTTPS same-site proxy is optional if the scoped test-origin routing works; otherwise its absence is an explicit browser-test prerequisite.
- Missing harnesses/provider support and Windows-only managed adapters are known possible execution limits, **not authorization to implement or install them now**. Linux graph execution and protected MCP/board credential storage remain gated; preserve these failures.
