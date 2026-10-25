# Secure client pairing and engine authentication

## Goal and approved scope

Protect engine data and execution even when its URL is public. Current deployment:
frontend `https://klm.codemob.com.br`, engine `https://klmengine.codemob.com.br`.
The frontend may stay public. A client must pair before accessing the engine.
All paired clients have full engine access in this first version.

User-approved lifecycle: expire credentials after **7 consecutive days without
authenticated use only**. There is **no 30-day or other absolute lifetime**.
Pairing codes expire after 5 minutes and can be consumed once.

Deliver the smallest coherent implementation. No accounts/passwords, JWT framework,
roles, per-project permissions, cloud identity service, or unrelated refactors.
Read `AGENTS.md` and `product.md`; preserve existing dirty work. The repository has
many pre-existing modifications, including API, store, app, and client transport.
Record the relevant starting state before edits so reviews can isolate this work.

## 1. Local bootstrap and durable credentials

- Add `klm pair` using the engine's protected local control channel (Windows
  per-user named pipe), not an unauthenticated HTTP issuance endpoint. Keep
  `klm start` / `klm stop` and dev/release data and control separation intact.
- Generate a cryptographically random 256-bit pairing secret; display it only to
  the local command caller. Accept redemption over the authenticated transport
  described below, with a bounded request body and bounded rate limiting.
- Consume the unexpired code atomically with durable creation of one client
  credential. Concurrent redemptions cannot create multiple clients; persistence
  failure cannot report successful pairing. Use existing locking/persistence patterns.
- Credentials are opaque, random 256-bit values. Persist hashes only, together
  with a client ID/name, creation time and last authenticated use. Never put raw
  credentials/codes in logs, URLs, exports, normal state responses or localStorage.
- Enforce inactivity server-side, including after engine restart. Renew on valid
  authenticated use, never revive an expired credential. Persist activity without
  rewriting full state for every SSE event. Document the chosen bounded strategy.
- Provide protected local recovery/revocation if all client access is lost.

## 2. Public API authentication and origin policy

- Centralize default-deny authentication around all public application routes:
  reads, mutations, histories, files/downloads, settings, uploads and SSE.
  Neither localhost nor LAN addresses bypass authentication.
- Keep narrowly enumerated exceptions only: pairing redemption and existing
  externally authenticated ingress/callbacks. GitHub ingress retains signature
  validation; OAuth callbacks retain their own state/transaction checks. Inspect
  actual routing rather than exempting whole prefixes such as `/api/library`.
  Do not expose existing detailed health/state information without authentication.
- Replace wildcard CORS with exact configured origins. Include the production
  frontend via configuration and legitimate desktop/dev origins as appropriate;
  never reflect arbitrary origins or accept arbitrary subdomains/`null` origins.
  Credentialed CORS uses `Vary: Origin` and explicit methods/headers. Preflight
  must work without opening the underlying operation.
- Cookie-authenticated mutations require exact Origin validation and a simple
  CSRF defense, including multipart uploads. CORS is not authentication. Native
  bearer clients still require their credential when no browser Origin is present.
- Remote credentials require HTTPS. Account explicitly for TLS-terminating
  proxies; do not trust arbitrary forwarded headers. Keep direct loopback native
  development/desktop working with authentication; do not silently permit plaintext
  LAN credentials. Document any HTTPS requirement for LAN browser use.
- Preserve the private authenticated harness/Library bridges and `loopbackHost`.
  Client credentials must not replace or broaden their separate capabilities.

## 3. Client connection and credential transport

- Add a small connection/pairing screen using existing design-system primitives:
  engine URL, client name, pairing code, submit/error state. All UI copy is English.
- Browser: use a host-only HttpOnly Secure cookie (prefer `__Host-` prefix), Path=/,
  SameSite=Strict for the supplied same-site HTTPS deployment. Refresh browser
  cookie lifetime along with the 7-day inactivity window; do not introduce an
  absolute expiration accidentally. Fetch and EventSource must send credentials.
- Desktop/mobile: inspect their actual transports and origin contexts. Use OS
  secure storage for native persistent credentials if cookie transport is not
  applicable; do not place a long-lived bearer token in frontend storage. A WebView
  loading the supplied HTTPS frontend can follow the browser cookie flow. Do not
  assume Strict cookies work for unrelated sites or native Tauri origins.
- Keep transport changes centralized (`platform.ts`, `engine.ts`, SSE connection
  helper and native adapters where needed). Cover uploads and direct download paths,
  not just the main JSON helper. Engine switching must not leak one engine's
  credential or displayed state to another engine.
- On 401/expiration/revocation, stop reconnect loops, clear protected displayed
  state and return to pairing. Ordinary network failures remain reconnectable.
- An authenticated open client counts as use. Ensure long-running SSE clients
  maintain authentication using a minimal authenticated renewal mechanism; server
  output alone must not keep an abandoned credential alive indefinitely.

## 4. Connected devices and revocation

- Add a compact `Connected devices` settings section: name, last use, current
  client indication, revoke one and revoke all. Do not expose credential hashes.
- Revocation is durable and immediately rejects further requests and terminates
  that client's existing SSE streams. It does not stop unrelated engine work.
- Keep local recovery available independently of browser credentials.

## 5. Verification and delivery

- Use narrowly targeted authentication/pairing checks and relevant compilation or
  type checks. No full suites or elaborate integration infrastructure. Use isolated
  temporary data for probes; never pair/revoke real devices or restart the live
  installed engine as a development check.
- Verify concrete boundaries: unauthenticated reads/writes/SSE denied; allowed and
  rejected origins; one-use/expired pairing; valid auth; 7-day sliding expiry and
  restart persistence; revocation/stream closure; signed webhook compatibility.
  Add small focused regression tests when they directly protect these boundaries.
- Document local pairing/recovery, allowed origins, proxy HTTPS configuration,
  first connection after upgrade and manual browser/desktop/mobile checks.
- Report actual checks and unvalidated paths in
  `CLIENT_AUTH_IMPLEMENTATION_PROGRESS.md`. No commit/push or installer generation.

## Exactly two review/correction rounds

The main conversation agent reviews; the implementing Sol 6.1 agent corrects.
After initial implementation, hand off a concise changed-file list, checks,
limitations and baseline-aware change summary. Round 1 covers actual correctness,
authentication boundaries and usable client connection. Send only concrete,
in-scope defects to the implementer. After its correction handoff, round 2 examines
the corrections and remaining in-scope defects. The implementer fixes that list
and reports targeted verification. Do not invent defects to fill a round or expand
into speculative hardening. If a round finds no defects, record that fact.

Review status: implementation delivered; both requested review rounds complete;
all requested corrections delivered, including R2.1 and its focused evidence.
Production/browser/native/mobile manual acceptance remains pending.
See `CLIENT_AUTH_REVIEW.md` and `CLIENT_AUTH_IMPLEMENTATION_PROGRESS.md`.
