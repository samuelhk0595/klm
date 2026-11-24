# First Werbert Factory / KLM runner delivery

Date: 2026-11-14 local / 2026-11-15 UTC. **Implemented, with a real representative
end-to-end journey established.** Human document/mock/visual acceptance remains
pending. This report is the entry point for the current uncommitted implementation.

**Publication update, November 15:** KLM 0.7.0 and Werbert Foreman are installed on
the existing VPS and connected. The [publication receipt](../../Hermes/werbert/docs/testing/klm-foreman-vps-deployment-20261115.md)
owns exact source/artifact hashes, preserved data, schema-compatible rollback and
operational checks. Deployment uncovered and corrected a Linux control-discovery
recursive lock; its real Unix discovery/Ready/Stop regression passed as `klm`.
Earlier statements below about no deployment describe their original local cohort.

## Authorization and source scope

The current user explicitly commissioned implementation and brief testing using
the four saved specialists. This authorizes the work independently of historical
planning-only labels in the plan/specification/runner proposal. Those input
documents and the original three local specialist TOMLs were preserved. Their
proposals are not retroactively marked human-reviewed or whole-mock accepted.

Both shared checkouts retain pre-existing and concurrent dirty work. Backend
attribution is listed in [backend evidence §3](FACTORY_BACKEND_EVIDENCE.md#3-specialist-attributable-file-inventory);
frontend ownership is Werbert's new `frontend/src/features/foreman/` plus its entry,
small shared mock-control seams and maintained experience documents. Concurrent
KLM personal-history/frontend changes are separate work.

## Implemented path and actual contracts

- KLM's additive `engine/runner_*` module owns immutable external origin,
  integration/captured roles, durable receipts/input inclusion, selected manifests,
  native history/requests and continuing owned shell jobs. External Foreman and
  workers always capture YOLO; personal settings and default execution stay separate.
- Ordinary projections, exact-ID routes, subscriptions, native grants and analytics
  exclude external executions. Parent integration also corrected the ordinary
  health `runningSessions` count to exclude external turns; its focused boundary
  test passes on the final source.
- Werbert's namespaced `foreman_*` module owns protected local Connect, stable
  project/Factory/Chat association, strict CAS `.werbert/factory.yaml`, captured saved
  Agent Assignment, Ticket/publications and independent objective assessment.
  `WERBERT_MODE=foreman` bypasses incumbent runtime/environment/memory startup.
- Real `/foreman` uses the existing source-shaped views with actual backend data.
  `/dev/foreman` remains fixture-only. Settings retains the requested runtime
  controls, document editor and Library modal. Chat shows applied read-only runtime;
  exact request cards remain independent of Foreman prose. Worker tabs read actual
  native history without creating or controlling executions.

The actual protocol uses snake_case new DTO fields and string revisions/cursors:

| Contract | Authoritative implementation reference |
| --- | --- |
| KLM runner protocol 1 | [engine/RUNNER_API.md](engine/RUNNER_API.md), `engine/runner_contract.go` |
| Werbert browser API | `C:\Users\Samuel\Documents\Projects\Hermes\werbert\engine\docs\foreman-api.md`, `engine/internal/engine/foreman_contract.go` in that checkout |
| Local discovery | Protected user pipe/socket `discover`; explicit Connect uses existing pair redemption, with no HTTP authentication exemption. |
| Runner ownership | Authenticated `X-KLM-Runner-ID`, existing contract 3 and engine identity; stable connection/key, application effect admission and original-operation lookup. |
| Browser mutation | `/api/v1/foreman`, `X-Foreman-Intent: 1`, stable `operation_id`, exact Host/origin checks; no bearer/code in browser. |
| Config boundary | Desired file digest and applied capture are distinct; fixed `capture_request` revalidates before complete turns. |
| Execution tools | `shell_start`, `job_read`, `job_wait`, `job_cancel`; maximum 5,000 ms foreground/wait, same continuing process. MCP-to-job conversion remains deferred. |

## Verification and corrections

1. **Step 0 first:** real authenticated Pi 0.85.1, OpenCode 1.18.35 and Codex
   0.153.4 probes established operational calls, actual next-step inclusion,
   continuing shell handoff/waits below five seconds, unanswered necessary human
   questions and owned process drainage. Codex shell-on/off comparison was actual
   native behavior. Full provider-side catalog attestation is not claimed.
   [Native feasibility](FACTORY_NATIVE_FEASIBILITY.md) retains exact evidence.
2. **Compilation/focused checks:** both Go builds, targeted KLM visibility/receipt
   boundaries and Werbert startup/CAS/business checks passed. All three actual
   native runner checks passed. Frontend `npm run build` passed (2,114 modules);
   small subsequent source changes passed `npx tsc -b`. Exact commands/results and
   source-cohort limits are in [backend evidence](FACTORY_BACKEND_EVIDENCE.md) and
   Werbert's maintained frontend experience document.
3. **One scoped review:** four material findings were fixed and the same reviewer
   closed all four through narrow follow-up: job completion wake/finality race,
   restart cancellation uncertainty, recent conversation on generation change and
   ordinary invalidation-only updates. [Scoped review](FACTORY_SCOPED_REVIEW.md)
   records its limits. No extra reviewer hierarchy or broad suite was used.
4. **Real browser/native path:** protected discovery/Connect → existing registered
   project → Factory/configuration → persistent Chat → one Ticket/Assignment →
   actual Pi worker → exact native question → human's Alpha answer through its UI
   card → actual file/result → separate satisfied Foreman assessment → final Chat
   publication and visible Ticket result. Worker read-only history, editor Escape
   retention and cancellation dismissal with zero cancel requests were observed.
   [Browser evidence](FACTORY_BROWSER_VALIDATION.md) includes all earlier blocked
   attempts, final identities, screenshots and cleanup; it does not conceal driver
   failures or claim the final selector command exited successfully.
5. **Concrete integration fixes:** launcher LF framing/disposal, bodyless mutation
   media type and project icon DTO were corrected. Runtime-discovered fresh
   `ticket_read` stale-error handling was fixed, targeted regression passed, and
   the corrected real path succeeded. Parent then fixed health-count visibility;
   `go test -tags dev -run '^TestRunnerDurableBoundary$' -count=1 -timeout=45s .`
   passed (`klm/engine`, 0.964 s), including one personal plus one hidden turn.
   The final health-only correction is compilation/focused-test evidence, not
   another provider/browser cohort. Targeted `git diff --check` was clean.

Final representative artifact is `delivery-label.txt`, exactly hex `416c706861`,
SHA-256 `b1a96dd646bccaa24cef7a3db22a6f995f05658f4f1c3272913e258c03e6fb24`.
Ticket `1534328993cab48945f83b200c23549f` became Completed only after retained result,
`native_completed` worker revision `112`, and a separate satisfied assessment
against that revision. Actual human answer provenance is the parent's question
tool, not a generated user message.

## Run and validate

Both source backends must include this implementation; the existing installed
engine was not upgraded. The smallest isolated reproduction is the maintained
Windows launcher and three terminals:

1. Build KLM from `KLM/engine` with `go build -tags dev -o
   "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-factory-backend-check.exe" .`;
   build Werbert from its `engine` with `go build -o
   "C:\Users\Samuel\AppData\Local\Temp\opencode\werbert-foreman-backend-check.exe"
   ./cmd/engine`. These are manual reproduction commands, not extra checks claimed
   here. The launcher requires the existing pinned Pi/native authentication.
2. Run `powershell.exe -NoLogo -NoProfile -NonInteractive -File
   "C:\Users\Samuel\Documents\Projects\Personal\klm\engine\validation\Start-ForemanJourney.ps1"
   -RunRoot "C:\Users\Samuel\AppData\Local\Temp\opencode\factory-manual-validation"
   -BrowserOrigin "http://127.0.0.1:5174" -Minutes 15`. The run root must be new;
   ports 17331 and 18090 must be free. Keep the terminal owned until cleanup.
3. From Werbert `frontend`, set `$env:WERBERT_ENGINE_URL =
   'http://127.0.0.1:18090'` and run `npm run dev -- --host 127.0.0.1 --port 5174
   --strictPort`. Open `http://127.0.0.1:5174/foreman`, Connect, choose the disposable
   project, explicitly configure Pi/model/effort, and save. Submit one bounded task;
   answer any actual native question. Inspect result and independent assessment.
4. Stop the frontend with Ctrl+C. The backend window expires automatically; it can
   end early when `<RunRoot>/stop.request` exists. Read `cleanup.json` and confirm
   exits/auth-copy removal. Retained databases are private evidence, not portable
   definitions. No restart or replay of an interrupted task is automatic.

For ordinary local use, Werbert's explicit mode uses existing `WERBERT_LISTEN`,
`WERBERT_DATA_ROOT`, exact `WERBERT_ORIGINS`/`WERBERT_HOSTS`; release KLM defaults to
7331, development uses `WERBERT_KLM_PORT=17331` and matching absolute
`WERBERT_KLM_DATA_DIR`. See the actual API document for protected binding/storage.

## Remaining acceptance

No known material blocker remains for the demonstrated first vertical slice.
Human visual/product acceptance, broad ordinary KLM connected/disconnected UI
comparison, nonempty selected Library/MCP/provider variants, audio, independent
paused-history permutations, native compaction/restart/Linux and integrated
OpenCode/Codex full journeys remain manual checks. Their implementation is not
claimed universally validated by the representative Pi journey.

All validation services were stopped, auth copies removed, and original auth and
saved Agent hashes preserved; final validation ports were free. Existing installed
7331/7332 listeners were untouched. No commit/push, deployment/installer, operational
data migration or personal specialist/Library selection change was performed.
