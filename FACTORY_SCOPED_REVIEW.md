# Factory runner scoped review and closure

Date: 2026-11-14. **Outcome: all four original material findings are closed.**
Read this record when assessing the commissioned first Factory/KLM runner slice.
Closure is source-review evidence, separate from browser/native and human acceptance.

## Scope

- **K:** `C:\Users\Samuel\Documents\Projects\Personal\klm`.
- **W:** `C:\Users\Samuel\Documents\Projects\Hermes\werbert`.
- Original review: attributable backend files/seams in
  [backend evidence §3](FACTORY_BACKEND_EVIDENCE.md#3-specialist-attributable-file-inventory),
  plus W's production `frontend/src/features/foreman/**`, entry and documented
  shared-control integration through FX-019. Baseline identities and dirty-work
  attribution are recorded in that evidence. Concurrent K client/history work and
  incumbent W Factory/Ticket/assignment changes were excluded.
- Acceptance criteria: the commissioned backend plan, persistent-agent and runner
  specifications, actual runner/Foreman API contracts and maintained experience
  ledger. Current implementation authorization supersedes historical planning-only
  labels; whole-document/mock acceptance remains pending.
- This follow-up inspected only the four fixes below, their directly affected
  paths, regression sources and recorded check evidence. It was not another broad
  review. Production sources were read-only.

## Original findings and closure

| Finding | Original failure | Inspected correction and outcome |
| --- | --- | --- |
| **1 — P1: stranded job completion** | Completion during an active Codex turn retained input without waking steering; worker settlement could then publish `native_completed` and refuse that continuation. | **Closed.** K `engine/runner_jobs.go:315–386` commits terminal job state and attributed input together, wakes steering after commit, and reopens terminal native state only with retained actionable input. Its live-job done barrier follows that commit (`263–264`). `runner_runtime.go:212–274` checks retained actionable input before worker finality and schedules after successful settlement. `TestRunnerJobCompletionWakeAndSettlement` covers active-turn wake and schedulable settlement without launching a harness. |
| **2 — P1: false cancellation after restart** | Missing live handles allowed persisted unknown jobs/native uncertainty to become `cancelled`. | **Closed.** K `engine/runner_controls.go:298–379` retains the prior uncertainty error without a live turn, checks all admitted execution effects, and checks persisted `starting`/`running`/`unknown` jobs in the final transaction. Unconfirmed cessation remains Interrupted. `TestRunnerCancelRetainsPersistedUncertainty` covers missing handles, each unsettled job state, retained native uncertainty and a positively settled control. |
| **3 — P2: lost conversation on configuration rotation** | A new generation received only pending inputs and work/request summaries, losing recent operational exchanges. | **Closed.** K `engine/runner_context.go:10–59` retains bounded previews with original session/event/ordinal and source identity. `runner_runtime.go:143–149,324–328` persists them at rotation and supplies them when initializing native context. W `engine/internal/engine/foreman_overview.go:89–143` adds recent Chat publications with original producer/source identity; `engine/prompts/foreman.md:30–33` distinguishes reference context from new authorization. Both windows cap at 24 entries / 32 KiB. The generation and Foreman provenance tests cover retained operational constraints, bounds and attribution; the K test confirms no additional input authorization was inserted. |
| **4 — P2: lost ordinary invalidations** | The runner filter removed invalidation-only harness updates even without a connected Factory. | **Closed.** K `engine/runner_store.go:166–209` preserves ordinary invalidation-only changes and unrelated invalidations accompanying external-session changes. `TestRunnerOrdinaryInvalidationOnly` checks connected/disconnected bindings and mixed external-session writes, including absence of hidden session IDs in ordinary notices. |

No material regression was established in these directly affected paths. No
additional correction is requested for the four reviewed findings.

## Verification limits

The reviewer inspected the implementation and regression sources; no executable
check was rerun. [Backend evidence §7](FACTORY_BACKEND_EVIDENCE.md#7-reported-findings-closure-and-browser-retry-artifacts)
owns the exact commands, outcomes and rebuilt executable hashes. Its initial
combined K command failed only on launcher disposal; all four regression tests
and the existing durable-boundary test passed in that invocation. The corrected
protected-control smoke then passed independently. W's focused recent-conversation
provenance test and both backend rebuilds are recorded as passing.

The control smoke exercises the launcher's extracted function against a real
disposable protected Windows pipe; it is not full application startup or a browser
journey. The regressions are disposable fixtures, not provider execution or crash/
restart acceptance. This review launched no tests, providers, browsers, servers or
additional reviewers and made no production edits, commits or pushes. Concurrent
browser retry results and human acceptance remain separately owned evidence.
