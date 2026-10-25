# Client authentication review

Scope: the approved plan and exactly two reviewer/implementer correction rounds.
This is a code review, not acceptance of the installed/browser/mobile deployment.
Implementation attribution follows `CLIENT_AUTH_IMPLEMENTATION_PROGRESS.md`;
the full dirty repository diff is not attributed to this feature.

## Round 1 — corrections requested

### R1.1 — stale authentication responses can undo newer auth state (P2)

Files: `clients/desktop/src/clientAuth.ts:19-35,38-53` and
`clients/desktop/src/features/settings/ClientConnection.tsx:18-49`.

`clientFetch` invalidates the current credential for every arriving 401, even when
that request was dispatched before a successful new pairing. Multiple concurrent
requests can be in flight when the first 401 opens the pairing form. If a second
old 401 arrives after `pairClient` sets the new credential and clears `blocked`, it
deletes the new native credential and blocks the newly paired client again.

The inverse race also exists: `check()` unconditionally calls `setConnected(true)`
for a successful response, even if another request has since invalidated auth.
Overlapping renewal/focus checks can therefore remount the protected app after
denial; subsequent synthetic 401s do not emit another denial event because the
transport is already blocked. The UI can get stuck mounted but unauthenticated.

Correction: a small auth generation/invalidation guard (or equivalent cancellation)
must make old responses incapable of clearing a new credential or reconnecting
after a newer denial. Add focused deferred-response regressions for both orders;
do not introduce a generalized state-machine framework.

### R1.2 — disconnected clients cannot select an already-paired engine (P2)

File: `clients/desktop/src/features/settings/ClientConnection.tsx:12,18-21,52-62,76`.

The editable Engine URL is used only by Pair client. Retry connection always
checks the module's original ENGINE_URL. If engine A is unavailable and the user
enters already-paired engine B, Retry still calls A. The only path that persists B
requires consuming a new local pairing code, even though B already has a usable
browser/native credential. This is especially blocking for native secure-store
credentials when the former engine is offline and Settings is inaccessible.

Correction: allow selecting/normalizing the entered engine and checking its
existing authentication without requiring a fresh code. A persist-and-reload path
is sufficient and already aligns with transport isolation. Keep engine credentials
separate and handle a fixed VITE_ENGINE_URL honestly rather than pairing B and
silently continuing to use A after reload. Verify switching from unavailable A
to paired B without calling the pairing endpoint.

### R1.3 — native development CSP forbids the required remote HTTPS transport (P2)

File: `clients/desktop/src-tauri/tauri.conf.json:26`.

The native development CSP allows `http:` and `ws:` connections but not `https:`.
The new connection flow requires HTTPS for every remote engine. A desktop:dev
client therefore cannot pair with or use the production engine: WebView blocks
the HTTPS fetch before authentication/CORS can run. Release CSP already permits
HTTPS. This is an integration fix needed by the new transport policy.

Correction: narrowly permit HTTPS in the development connect-src policy while
preserving the rest of the existing CSP. Do not relax remote transport security.

## Evidence and next handoff

- Read the auth middleware/storage/local-control implementation, client transport,
  pairing/revocation components, native credential adapter, API integration and
  focused tests. Checked actual mobile audio flow: upload uses the shared frontend
  transport rather than a separate unauthenticated native upload.
- Findings above are code-path findings; no new test pass is claimed by the reviewer.
  A scoped git-diff command timed out; it did not supply review evidence. Existing
  implementer test/build results remain attributed to the implementer's report.
- Implementer should correct R1.1–R1.3, run only targeted regressions/checks relevant
  to those corrections, and send a single round-2 handoff with actual results.
- Round 1 review complete; implementer corrections delivered as detailed below.
  The completed round-2 result and final correction evidence follow below.

## Round 1 correction evidence — implementer

- R1.1: numeric generation guards in `clientAuth.ts` reject stale responses before
  invalidation/data delivery. Pairing transitions and explicit retry advance the
  generation; `ClientConnection.check()` guards connection/error updates as well.
  Deferred-response regressions cover an old native 401 after successful re-pair
  (the new credential remains installed) and late renewal success after denial
  (protected content remains unmounted).
- R1.2: Connect uses the edited URL, normalizes/persists it, then reloads to select
  that engine's existing credentials. No pairing request is needed to switch.
  A shared selection guard rejects a conflicting `VITE_ENGINE_URL` before either
  redemption or storage mutation. Regression checks cover offline A → paired B,
  native B credential lookup without forwarding A's bearer, and the override case.
- R1.3: only `https:` was added to native development connect-src. Other CSP and
  remote transport restrictions are unchanged. Native runtime acceptance is pending.
- Files changed: `clients/desktop/src/clientAuth.ts`, `clientAuth.test.ts`,
  `platform.ts`, `features/settings/ClientConnection.tsx`, new
  `features/settings/ClientConnection.test.tsx`, `clients/desktop/src-tauri/tauri.conf.json`,
  and scoped README/progress/review documentation.
- Implementer checks: targeted Vitest command for auth transport, connection UI and
  existing update reconnect tests **passed (9 tests / 3 files)**; final
  `npx tsc --noEmit` **passed** after correcting two new test-query option types;
  targeted tracked-file `git diff --check` **passed**. Earlier Go/cargo evidence
  was not rerun or reattributed. No full suites or live engine/data changes.

This section records implementer corrections rather than a review result. The
reviewer's completed second-round result follows.

## Round 2 — reviewer result

The reviewer inspected the round-1 corrections and their focused regression
tests. The auth-generation guards address the two reported stale-response orders;
Connect now selects an existing engine without redeeming a code; development CSP
now permits HTTPS without broadening its other directives. No new suite was run
by the reviewer; the nine passing tests and TypeScript result above remain
attributed to the implementer.

One concrete correction remains in URL identity handling:

### R2.1 — normalized selection and raw configured engine use different native credential keys (P2)

Files: `clients/desktop/src/platform.ts:15-20,34-40,45-66`,
`clients/desktop/src/clientAuth.ts:4,14,60-61`, and
`clients/desktop/src/features/settings/ClientConnection.tsx:59-64,73-76`.

`selectableEngineURL()` compares normalized URLs, but `resolveEngineURL()` returns
an override (or saved URL) with only trailing slashes removed. Native storage and
equality checks use that unresolved spelling. A valid override such as
`https://klmengine.codemob.com.br:443` resolves with `:443`, while selection/pairing
normalizes to `https://klmengine.codemob.com.br`. The override guard correctly
considers these the same engine, but pairing stores the credential under the
normalized string and reloads because the strings differ. After reload the native
client looks up the original `:443` key, finds no credential and asks to pair again.
Connect likewise repeatedly reloads instead of probing that same engine.

Correction: use one canonical engine identity consistently for resolved config,
selection, transport and native credential keys. Cover explicit default ports and
hostname case. Keep legitimate base-path prefixes and dev ports intact. Retain
useful recovery/error behavior for invalid legacy saved addresses; a normalization
failure should not crash the application during module import. No generalized URL
registry or credential migration framework is needed.

Focused acceptance: a native client with a valid fixed override containing `:443`
can pair once and recover the same stored credential on reload, without repeated
reload/redemption. Equivalent spellings must compare equal; genuinely different
engine overrides must still be rejected before pairing.

Round 2 review is complete. It is the final requested review round; the final
implementation correction and its evidence follow, without another broad review.

## Final correction evidence — implementer

R2.1 is fixed. `resolveEngineURL()` canonicalizes override/saved URLs through the
same helper used by selection, keeping fetch bases, equality checks and native
credential keys identical. Invalid legacy saved URLs are caught and fall back
without breaking module initialization. The default engine-port heuristic applies
only to bare IP input; absolute URLs retain standard default-port semantics, making
normalization idempotent for IPs too. Base prefixes and explicit dev ports survive.

- Changed implementation paths: `clients/desktop/src/platform.ts` and new
  `clients/desktop/src/platform.test.ts`. Scoped README/progress/review documentation
  was updated. No backend/native source was changed in the final correction.
- Actual checks: targeted Vitest run across platform, auth transport, connection UI
  and update reconnect **passed (18 tests / 4 files)**; `npx tsc --noEmit` **passed**;
  tracked platform `git diff --check` **passed** (LF-to-CRLF notice only).
- New tests use real URL resolution/normalization and auth modules with mocked OS
  credential storage/network: mixed-case explicit `:443` override pairs once and
  retrieves the identical native key after reload, equivalent spellings compare
  equal, different overrides reject, and paths/dev ports/legacy fallback work.
- The earlier Go/native build results remain attributed to the initial implementer
  checks, not rerun here. No full suite, live engine/data changes, installers,
  commits or pushes. Actual production/native/mobile manual acceptance is pending
  as specified in `CLIENT_AUTH_IMPLEMENTATION_PROGRESS.md`.

Both requested reviews are complete and their correction lists are implemented
with focused evidence. This final correction report is not an additional review.
