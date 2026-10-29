# Engine directory browser — single static review

## Scope and evidence

Third/final reviewer reviewed the implementation submitted in `implementation-ready.json` (published 2026-10-28T06:27:29.614Z), after the event-based ready wait completed. Review used the external scoped baseline-to-ready patch and immutable ready snapshots, not a HEAD diff. All submitted baseline/snapshot/patch SHA-256 hashes matched. Unrelated concurrent SQLite/Linux changes were not reviewed.

Read the approved plan, project/product context and relevant UI/React guidance. This was one focused static review; no source fixes, builds, test suites, servers, browsers, deployment or runtime creation were performed. The implementer's Windows/Linux Go builds and TypeScript check are reported as passed in the handoff; the reviewer did not rerun them.

## Required correction by file

### R1 — P2 / medium: native cancellation revives an invalidated directory selection

- **File:** `clients/desktop/src/features/projects/DirectoryBrowser.tsx`
- **Submitted lines:** 42–46, 77–99, 111–112 (principally line 79).
- **Reproduction:** On a Windows-capable engine, successfully load directory A. Reload A with Go after it becomes unavailable/unreadable (or a transient read failure occurs). The reload fails, so Select is correctly disabled. Click **Native picker on engine**, then cancel the native dialog or use **Cancel selection**. Directory navigation retained the previous `listing`; native invocation clears `error` and `failedTarget`, and cancellation clears `busy`. `draft === listing.path` is still true, so **Select this directory** becomes enabled for A without a successful reload. The old directory rows can also reappear as if the failed reload succeeded.
- **Impact:** Violates the approved rule that failed/unloaded/stale navigation must not be selectable. The registration validator remains a separate safeguard, but does not make the browser's current-directory state correct.
- **Concrete correction:** Track successful/current listing validity independently of the generic error message. Invalidate selection when a directory navigation begins, and make it valid again only when that generation's directory request succeeds. Native start/cancel/null results must not restore validity or erase the failed-directory state/Retry. Clearing the selectable listing at navigation start is another small option; if retaining the last listing for Up recovery, keep it explicitly nonselectable. Preserve the existing generation/controller guards, including `finally` and native results.
- **Disposition:** Required before feature acceptance; awaiting the same implementer's correction. No post-fix review is authorized or claimed.

## Blockers versus optional polish

- No P0/P1 security or compilation blocker identified in this scoped static review.
- R1 is the one required functional correction (P2).
- No additional optional-polish findings are requested; preserve scope.

## Preservation observations

- Both read routes are registered beneath the existing paired-client protection/no-store wrapper. Existing Origin, TLS, cookie/CSRF and private-loopback boundaries are unchanged; helpers use `request` → `clientFetch`, encoded paths and abort signals.
- SQLite availability admits the two new routes while retaining the existing storage-contract gate. The handoff's known missing shared-client contract-header/query work is preexisting and **not** a finding against this feature; do not bypass the gate to address it.
- Windows roots enumerate actual logical-drive bits without drive probes; Unix exposes `/`; home is engine-derived. Lexical path cleaning, server parents (including UNC share roots), directory aliases and OS-permission failures are handled without recursive/content reads or filesystem writes.
- Listing code uses bounded batches, raw-entry/item/encoded-JSON budgets, cancellation checks and closed handles outside `app.mu`. Partial read errors are errors; skipped/capped results disclose incompleteness rather than claiming complete emptiness.
- Request generations cover success/error/finally, navigation, draft edits, native results and component cleanup. R1 concerns retained selection validity, not a missing stale-response guard.
- Browser view replaces the existing form rather than stacking dialogs/forms. Draft name/icon/manual path and edit-directory immutability remain in the parent; selection itself does not register anything. Escape cancellation, focus-return paths, semantic controls and scoped token-based responsive styles are present in source.

## Limits and pending human validation

Static inspection is not runtime acceptance. Windows drives/UNC/junctions/native picking, Unix nonroot permissions/symlinks, offline or deleted paths, large/incomplete listings, rapid navigation/cancel/reopen, keyboard/focus/Escape, mobile long paths/landscape/actions and both themes still require human validation. Denied/revoked/wrong-origin requests must also be checked in the real authenticated client environment. Blocking filesystem syscalls are not forcibly interruptible, as disclosed.

The same still-active implementer will consume the atomic result, apply necessary focused fixes and record dispositions/evidence. This document reviews only the immutable initial submission; later dispositions are implementer-reported, not a second reviewer pass.

## Same implementer's correction disposition (not post-fix reviewed)

- **R1: corrected.** Added explicit current-listing validity, invalidated at every navigation start and restored only by its current successful directory response. Invalid retained rows are hidden; previous parent metadata remains usable solely for Up recovery. Select requires valid listing state independently of error text.
- Native errors now have separate state. Starting native selection, cancelling explicitly or receiving a null native result preserves failed-directory error/Retry and cannot restore invalid listing validity. Existing generation/controller guards are preserved.
- **Affected check:** from `clients/desktop`, `./node_modules/.bin/tsc --noEmit -p tsconfig.json` passed (exit 0). No Go source changed after the reported successful release builds, so those were not repeated.
- Original ready snapshots and reviewer result remain immutable outside the repository; pre-correction baselines, final snapshots and correction delta accompany `implementation-complete.json`.
- This section is authored by the same implementer. No second review or runtime validation occurred. R1's real native failure/cancel recovery flow remains part of human acceptance; the preexisting SQLite contract mismatch remains unchanged.
