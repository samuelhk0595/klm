# Secure client authentication — implementation evidence

## Status

Initial implementation delivered for review round 1 (2026-10-24).
Both requested review rounds are complete. R1.1–R1.3 and the final R2.1 correction
are implemented, with targeted checks passed. This document records
implementation and targeted checks, not human acceptance of browser/native/mobile behavior.
Implementer session: `ca4b148d5ab96b27cea345eb8f1f180f`.
Reviewer session: `189adaf6e05fd2c93755ea67e756071a`.
Round-1 handoff was accepted into the reviewer's history using operation
`client-auth-implementation-ready-v1` (message `48d1683a07396dc895298a96075d7f8c`).
The reviewer subsequently delivered R1.1–R1.3 in `CLIENT_AUTH_REVIEW.md`.
No polling or monitoring is active.
Round-2 correction handoff was accepted into the reviewer's history using operation
`client-auth-round1-corrections-ready-v1` (message `2dab4bdeec91dda1dfa09b1df443b01d`).
The reviewer subsequently completed round 2 and requested the final R2.1 correction.
The final correction/evidence handoff was accepted into the reviewer's history as
`client-auth-final-evidence-v1` (message `dde47dcc0503774b34c6dc821736f596`).
No further review round or monitoring was started.

## Release 0.6.0 preparation (explicit user request)

- Created `release/0.6.0` from the existing local `release/0.3.0` HEAD (`fe5fe30`).
  The commit is scoped to this authentication implementation and release-version
  metadata; pre-existing unrelated changes remain in the shared working directory.
  Overlapping files were staged by the saved starting baseline rather than adding
  their entire current contents. No original working-tree content was reverted.
- Desktop npm/package lock, Cargo/package lock and Tauri metadata are `0.6.0`.
  Mobile is `0.6.0+3`; engine NSIS/build defaults and VS Code's release input are
  `0.6.0`. No installers or runtime data were generated for this preparation.
- Exported the actual staged tree to an isolated temporary snapshot (without the
  other uncommitted functionality), then verified that exact source combination:
  `go test -tags dev -run ^TestClientAuth -count=1 .` **passed** (5 focused checks),
  `npx tsc --noEmit` **passed**, and the four-file targeted frontend Vitest command
  above **passed** (18 tests). These are staged-release checks, not live acceptance.
- A staged-metadata consistency probe **passed** for npm/Cargo/Tauri/mobile/NSIS/
  engine-build/default-release-input versions. `git diff --cached --check` passed.
- Manual browser dev feedback identified a `localhost` page with a fixed
  `127.0.0.1` API override. The documented dev path now uses `127.0.0.1` for both.
  Successful real-client pairing has not yet been reported, so the remaining
  production/native/mobile acceptance items below are still pending.
- Commit and branch push are expressly requested in this release step. Earlier
  statements about no commits/pushes describe the implementation/check phases.

## Baseline and attribution

Read `AGENTS.md`, `product.md`, and `CLIENT_AUTH_IMPLEMENTATION_PLAN.md` before
editing. The project was already extensively dirty. A complete tracked starting
diff was captured before these edits at:

`C:\Users\Samuel\AppData\Local\Temp\opencode\client-auth-starting.diff`

The initial status included unrelated automation, graph, analytics, runtime,
Library, installer, project and chat work, plus untracked files. The plan itself
was already present. No existing untracked file was overwritten. **Do not attribute
the full current git diff to this implementation.** In particular, `App.tsx`,
`store.go`, `platform_windows.go`, `platform_unix.go`, and the distribution scripts
were already modified and were not edited by this implementation.

### New files owned by this implementation

- `engine/client_auth.go`: credential persistence, pairing/revocation, default-deny
  middleware, origin/TLS/CSRF policy, protected device endpoints.
- `engine/client_auth_test.go`: five focused authentication-boundary checks.
- `clients/desktop/src/clientAuth.ts`: centralized cookie/native bearer fetch,
  native credential lifecycle and authenticated fetch-streaming SSE adapter.
- `clients/desktop/src/clientAuth.test.ts`: three focused transport checks.
- `clients/desktop/src/features/settings/ClientConnection.tsx`: authentication gate
  and pairing form; unmounts the protected application on 401.
- `clients/desktop/src/features/settings/ConnectedDevices.tsx`: safe device list and
  revoke-one/all actions.
- `clients/desktop/src/features/settings/client-connection.css`: existing visual
  tokens used for the pairing/device surfaces.
- `clients/desktop/src-tauri/src/client_auth.rs`: Windows Credential Manager adapter.
- `CLIENT_AUTH_IMPLEMENTATION_PROGRESS.md`: this evidence/status document.

### Narrow edits to existing files

- `engine/api.go` (already dirty): add five auth routes and wrap all public routes
  in the new middleware; remove wildcard CORS. Existing handlers/media checks kept.
- `engine/main.go` (already dirty): one `clientAuth` field and startup initialization.
- `engine/cli.go`: dispatch `pair` and `revoke-clients` through existing local control.
- `engine/cli_windows.go`: named-pipe requests and CLI-only code display.
- `clients/desktop/src/engine.ts` (already dirty): JSON and multipart requests use
  the centralized authenticated transport.
- `clients/desktop/src/features/chat/updatesConnection.ts`: default stream factory
  uses authenticated fetch streaming; existing cursor/retry logic remains.
- `clients/desktop/src/platform.ts`: production endpoint default, normalized URL
  helper, remote-HTTPS validation, and configure-through-helper behavior.
- `clients/desktop/src/main.tsx`: wrap `App` in the connection gate. `App.tsx` itself
  was not changed; unmounting clears all mounted workspace/chat/graph/device state.
- `clients/desktop/src/features/settings/SettingsDialog.tsx` (already dirty): add
  the Connected devices section only.
- `clients/desktop/src-tauri/src/main.rs`, `Cargo.toml`, `Cargo.lock`: register three
  credential commands and add `keyring` (Windows native) / direct `url` dependency.
  The lockfile gained two packages (`keyring`, `zeroize`) and dependency references.
- `engine/README.md`, `product.md` (both already dirty): describe the new approved
  auth contract and replace conflicting insecure-network descriptions narrowly.

## Implemented contract and choices

- `klm pair` generates 256 random bits over the existing per-user Windows pipe,
  displaying the code only to its local caller. One code is outstanding at a time,
  expires in five minutes, and is consumed under the same mutex as durable client
  creation. It stores only an in-memory hash; restart invalidates outstanding codes.
- Redemption is the sole unauthenticated pairing API. JSON is bounded to 2 KiB,
  with 20 attempts per engine per minute (including malformed JSON attempts).
- Each device gets an opaque 256-bit credential. Server storage contains only
  SHA-256 hashes, IDs/names, creation and authenticated-use times. It is separate
  from state/history exports and is not exposed by application APIs.
- `client-auth.json` is independently synchronized, atomically replaced, and capped
  at 128 active clients. Every admitted request persists last use before serving
  protected content. This deliberately simple bounded-file strategy costs a small
  sync/replacement per request, without rewriting full application state or writing
  per SSE event. Failed persistence denies access/pairing/revocation.
- Expiration is **only seven consecutive days without authenticated use**. There
  is no absolute lifetime. Renewed use survives restart and never revives expiry.
  Cookie age renews on admitted browser requests; the connection gate also sends
  an authenticated one-minute renewal while open. SSE output itself does not renew.
- All public routes are protected by default, including health/version, state,
  settings, history/downloads, uploads and SSE. No localhost or LAN auth bypass.
  Exact exceptions preserve existing GitHub HMAC ingress and OAuth state validation;
  no Library or webhook-management prefix is exempt.
- Exact configured-origin credentialed CORS, explicit headers/methods, `Vary:
  Origin`; cookie mutations (including multipart) require an allowed Origin and
  `X-KLM-CSRF: 1`. Preflight does not authenticate or execute an operation.
- Remote credentials/codes require HTTPS. Forwarded TLS is trusted only for explicit
  immediate-peer CIDRs and one exact `X-Forwarded-Proto: https` value. Direct HTTP
  loopback is permitted only with loopback peer AND Host, no forwarding headers,
  and a valid credential. Private harness/Library bridge capabilities are untouched.
- Browser and supplied same-site HTTPS mobile WebView use a host-only Secure
  HttpOnly `__Host-` Strict cookie. Cookie names include the fixed API port so
  concurrent dev/release clients on the same hostname do not overwrite each other.
  Direct loopback HTTP has a separate HttpOnly Strict cookie. Native Tauri uses
  engine- and app-identifier-scoped Windows
  Credential Manager storage plus an in-memory bearer; no frontend secret storage.
- JSON, multipart and SSE share authentication. Existing session exports already
  pass through the JSON request helper before producing a local Blob download.
  Engine switching reloads, selects separate native storage and clears mounted data.
- A 401 blocks further transport attempts, clears the native credential and
  unmounts the protected app. Ordinary connectivity errors retain reconnectability.
- Revocation persists first, cancels active requests/streams, and interrupts blocked
  SSE writes via their response deadline. Streams also have inactivity-expiry timers.
  Engine work is unaffected. `klm revoke-clients` is independent protected recovery
  and clears outstanding pairing codes, followed by `klm pair` for new access.

## Checks actually performed

- `go test -tags dev -run ^TestClientAuth -count=1 .` in `engine`: **passed** (five
  focused checks). Covers unauthenticated read/write/SSE/download/health rejection;
  exact origins, CSRF, preflight and trusted/untrusted forwarding; concurrent
  single-use and expired codes; failed persistence and HTTP browser redemption;
  rate limit; hash-only storage; cookie attributes/renewal; no absolute creation-age
  expiry; seven-day inactivity and restart persistence; durable revocation and live
  stream closure; actual signed ping acceptance/invalid signature rejection and
  invalid OAuth state rejection through the protected router.
- The immediate blocked-write cancellation change initially made the stream test
  receive `unexpected EOF` instead of a graceful final chunk. The assertion was
  corrected to accept either EOF form while still failing a timeout; the final
  targeted Go run passed. This is intentional transport interruption on revocation.
- `npx tsc --noEmit` in `clients/desktop`: **passed**, including final new test files.
- `npx vitest run src/clientAuth.test.ts src/features/chat/updatesConnection.test.ts`:
  **passed** (four checks in two files). Covers browser credentialed multipart/CSRF,
  401 request blocking, native secure-storage/bearer adapter behavior with mocked OS
  calls, fragmented SSE parsing/cancellation and existing cursor/retry cleanup.
- `cargo check` in `clients/desktop/src-tauri`: **passed**. One unused import warning
  from the new adapter was removed afterward. This checks compilation, not actual
  Credential Manager use, WebView network policy or runtime persistence.
- Go formatting applied to changed Go files.
- `git diff --check --` for the changed tracked implementation paths: **passed**;
  Git printed only the repository's LF-to-CRLF normalization notices.

All Go authentication fixtures use temporary directories and isolated handlers/
an ephemeral test server. Frontend checks mock credentials/transport. No real codes,
devices or engine data were used. No full suites, installers, live engine restart,
commits or pushes were performed.

## Review corrections and remaining manual acceptance

### Round 1 corrections (reviewed in round 2)

- **R1.1 corrected:** added a small numeric authentication generation. Invalidation,
  explicit retry and pairing transitions advance it. Responses dispatched under an
  older generation abort without invalidating the replacement credential or
  delivering stale protected data. The connection gate also checks its captured
  generation before applying a result/error, preventing late successful renewal
  from remounting after denial. Pairing blocks old requests during native credential
  replacement and advances the generation on completion.
- **R1.2 corrected:** the connection screen's **Connect** action normalizes/persists
  the entered engine and reloads when it changes. This selects the new engine's
  existing cookie/native credential without redeeming a new code; the old engine's
  in-memory bearer is not forwarded. Same-engine Connect explicitly retries.
  A fixed `VITE_ENGINE_URL` is checked before pairing or persisting selection, with
  a clear error when the user selects another engine. Settings uses the same guard.
- **R1.3 corrected:** added only `https:` to native `devCsp` connect-src. All other
  CSP restrictions and remote HTTPS enforcement remain intact.

Correction files: `clients/desktop/src/clientAuth.ts`, `platform.ts`,
`features/settings/ClientConnection.tsx`, `src-tauri/tauri.conf.json`,
`src/clientAuth.test.ts`, new `src/features/settings/ClientConnection.test.tsx`,
plus scoped documentation edits in `engine/README.md`, this document and
`CLIENT_AUTH_REVIEW.md`. Paths under `src/` above are relative to `clients/desktop`.

Actual correction checks:

- `npx vitest run src/clientAuth.test.ts src/features/settings/ClientConnection.test.tsx src/features/chat/updatesConnection.test.ts`
  **passed, 9 tests in 3 files**. New regressions defer an old 401 until after a
  successful native re-pair and defer a successful renewal until after denial.
  They also exercise unavailable engine A → existing paired engine B selection
  without redemption, engine-B native secure-store lookup/bearer isolation after
  module reload, and fixed-override rejection before pairing or selection writes.
- `npx tsc --noEmit` **passed**. Its first correction run found unsupported `exact`
  options on two new testing-library role queries; those options were removed and
  the final type check passed. Production code had no reported type errors.
- Targeted `git diff --check` for the modified tracked platform/config files
  **passed**, with only LF-to-CRLF normalization notices.
- The native CSP change was checked as a narrow configuration edit; actual native
  development HTTPS behavior still requires the manual check below. Go/native
  source was not changed for these corrections, so their earlier check results
  remain historical; no backend or native rebuild was rerun for round 1.

The reviewer completed the second and final review, accepted the reported round-1
cases as addressed, and identified the one URL identity correction below.

### Round 2 final correction (R2.1 complete)

- `resolveEngineURL()` now uses the same normalization as selection for fixed
  overrides and saved configuration. Hostname case, trailing slashes and explicit
  default ports resolve to one identity for `ENGINE_URL`, native secure-store keys,
  fetch paths and engine comparison. Base-path case/prefixes and explicit dev ports
  remain intact.
- Saved configuration normalization remains inside the existing guarded storage
  read. Invalid legacy addresses fall back to the normal derived engine instead
  of throwing during module initialization.
- Bare IP input still defaults to port 7331. An absolute HTTP(S) URL now honors its
  standard default port when no port is specified; the IP port heuristic applies
  only to bare input. This makes canonicalization idempotent for explicit `:80`/
  `:443` IP URLs as well as domains, without changing explicit dev ports.
- Final correction paths: `clients/desktop/src/platform.ts`, new
  `clients/desktop/src/platform.test.ts`, and scoped documentation in
  `engine/README.md`, this document and `CLIENT_AUTH_REVIEW.md`. No backend/native
  source changes or credential migration framework were added in this round.

Final correction checks actually performed:

- `npx vitest run src/platform.test.ts src/clientAuth.test.ts src/features/settings/ClientConnection.test.tsx src/features/chat/updatesConnection.test.ts`
  **passed, 18 tests in 4 files**. The nine new platform cases use the actual
  normalization/resolution and transport modules (only the OS store and network are
  mocked). They verify a mixed-case `:443` override pairs once, uses the same native
  key and fetch base immediately and after module reload, and still rejects a truly
  different override. They also cover preserved base paths/dev ports, repeated
  normalization/default-port identity and invalid saved configuration at module init.
  The previous nine auth/connection/reconnect regressions also pass in this run.
- `npx tsc --noEmit` **passed**.
- `git diff --check -- clients/desktop/src/platform.ts` **passed**, with only the
  repository's LF-to-CRLF normalization notice.

Both reviews and all their requested corrections are now complete. No additional
review round was initiated. This final evidence is implementation/fixture evidence;
production/browser/native/mobile human acceptance remains pending below. Earlier
Go/native compile evidence is unchanged and was not rerun or reattributed.

### Manual acceptance

1. Configure the engine process with
   `KLM_ALLOWED_ORIGINS=https://klm.codemob.com.br` and actual immediate TLS-proxy
   CIDRs in `KLM_TRUSTED_TLS_PROXIES`. Proxy preserves Host and overwrites the TLS
   header; do not trust forwarded values from arbitrary peers. See `engine/README.md`.
2. After the user upgrades/restarts the engine, verify URL-only access to reads,
   mutations, files and streams is denied, including localhost/LAN. Pair separately
   in each browser/WebView/native client with a local `klm pair` code.
3. In the production browser, verify host-only Secure HttpOnly Strict cookies,
   credentialed CORS, refresh/reopen, uploads, export downloads and a long-running
   conversation/renewal. Verify rejection of another origin and missing CSRF header.
4. In the native Windows client, verify the actual Tauri Origin, authenticated
   loopback transport, Credential Manager persistence after reopening, remote HTTPS,
   engine switching and dev/release storage separation. Also check native development
   HTTPS under the revised CSP. With engine A unavailable, use Connect to select an
   already-paired B without a new code; confirm a fixed VITE override rejects an
   incompatible selection before any redemption. Confirm equivalent hostname case/
   explicit default-port spellings pair once and reopen using the same native entry.
5. In the Flutter app, load the supplied HTTPS frontend/engine and verify actual
   WebView cookie persistence, pairing/reopen, SSE, native audio upload and return to
   hosts. No mobile bearer bridge is needed for this same-site browser transport.
   Plaintext LAN uploads/access are now rejected; configure HTTPS LAN hosts if used.
6. Revoke another device with an open stream and then the current/all devices;
   verify immediate return to pairing, cleared displayed data, rejected reconnect,
   and unaffected engine work. Verify local recovery with `klm revoke-clients`.
7. Confirm real GitHub signed ingress and a real provider OAuth round trip through
   the production TLS proxy. The focused checks do not establish provider acceptance.

There is no cookie workaround for unrelated frontend/engine sites, no plaintext
LAN fallback, and no non-Windows native credential implementation in this Windows
desktop slice. These are explicit transport boundaries, not silent protection
reductions. Browser policies, native storage and installed CLI behavior still need
the human checks above. No implementation blocker was found by compilation or the
targeted fixtures; runtime paths remain unvalidated until those checks occur.
