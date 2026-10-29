# Linux VPS final static review

## Required correction before deployment

### [P2] Latch stop atomically with readiness publication

**Paths:** `engine/cli_linux.go:227-232`; `engine/main.go:277-279`.

Local control starts accepting connections before initialization completes. A `stop` received during that interval clears `controlReady` under `a.mu`, but releases the mutex before acknowledging and closing `shutdown`. Startup then unconditionally sets `controlReady = true`. If the stop handler is descheduled between its unlock/acknowledgement and channel close, another local `ready`, `pair`, or `revoke-clients` request can observe readiness and succeed after stop was accepted (including after its acknowledgement). The channel check at lines 239-245 does not fence this interval because the channel is not closed yet. Normal `start`/`stop` CLI serialization reduces exposure but does not make the control protocol's startup/stop state transition safe.

**Fix:** Record a monotonic stop-requested state under `a.mu` before releasing that mutex or writing the acknowledgement. For example, close `shutdown` through `stopOnce` while holding `a.mu`, then acknowledge after unlocking; startup must check that same latched state/channel while holding `a.mu` and must not publish readiness after stop. Keep stop acknowledgement and existing graceful-shutdown/lock-release waiting intact. Do not set `a.closing` early without checking its existing shutdown/admission semantics.

This is a lifecycle/readiness correctness blocker for this delivery, not a demonstrated cross-UID or public-auth bypass.

## Optional polish

None raised.

## Scope and limitations

- Reviewed only the supplied baseline-relative implementation patch and its ten submitted/current touched files. SHA-256 comparisons showed all ten current files matched their submitted copies at review time. Read `AGENTS.md`, all of `product.md`, and `LINUX_VPS_IMPLEMENTATION_PLAN.md`; existing CLI, Unix lock, Windows control and auth functions were reference context only.
- Source-checked data-directory/socket ownership and modes, both peer-credential checks, bounded protocol, stale/replacement cleanup, startup/stop/restart handling, first-listen bind selection, and distribution/documentation boundaries. No additional actionable findings were identified in the scoped delta. Unrelated SQLite changes were not reviewed or modified.
- Existing Traefik/Docker/private-hop deployment is not implemented by this patch; no finding assumes an unshipped topology. No infrastructure or secret values were inspected.
- Static review only: no builds, test suites, runtime probes, deployments, client tests or additional agents. Linux/Windows compilation is implementer-reported evidence, not independently rerun here. Actual Linux lifecycle/socket behavior, private proxy/TLS configuration, pairing/SSE and native-client acceptance remain pending after the correction. Only this review document was written.

## Parent disposition (not another review)

The parent corrected P2 before deployment: stop closes the shutdown latch under
`a.mu` before acknowledgement, and startup checks that latch under the same mutex
before publishing readiness. Linux and Windows release builds succeeded after the
fix. The finished implementer could not be resumed through the available tool,
so the parent made this focused correction rather than create a fourth subagent.
Subsequent deployment/client evidence is recorded separately in
`LINUX_VPS_IMPLEMENTATION_PROGRESS.md`; no second review is claimed.
