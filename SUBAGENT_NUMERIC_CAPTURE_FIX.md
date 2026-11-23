# Subagent numeric capture persistence fix

## Scope and change

The reported OpenCode parent/Plan Writer child failure is consistent with the
numeric decoding mismatch described below. The exact incident capture/database
was not inspected; this is not forensic confirmation of that execution.

`decodeSQLiteRecord` uses `json.Decoder.UseNumber`, including for generic MCP SDK
tool schemas in `Capture.ExternalCapabilities.Connections[].Catalog.Tools`.
`validateDelegationWrite` previously decoded the incoming invocation with
`json.Unmarshal`, producing `float64` values in the same generic fields. Thus
`reflect.DeepEqual` could reject an unchanged capture containing a number such as
`inputSchema.properties.limit.maximum = 50`.

The invocation validation branch now also uses `UseNumber`. Both sides retain
the JSON number text, including integers beyond float64's exact range, precise
decimals and small exponents. The complete capture, task, ownership, terminal
immutability and monotonic dispatch-evidence comparisons remain in place. There
is no schema migration, record deletion or new lease/finality exemption.

Files changed for this fix:

- `engine/sqlite_delegation.go`: incoming invocation decoding only.
- `engine/sqlite_delegation_numbers_test.go`: focused repository unit regression
  using the existing temporary SQLite fixture, with no harness or MCP execution.
- `SUBAGENT_NUMERIC_CAPTURE_FIX.md`: implementation, recovery and validation notes.

## Related paths inspected

- `delegationExecution.markDispatch` reads an invocation, changes only state and
  dispatch intent, and writes through the corrected gate. Its transaction commits
  before the runner is called. A validation rejection rolls back that intent and
  the associated batch update; it does not dispatch that child.
- `finishItem` / `finishDelegationItem` and startup recovery write invocations
  through the same gate. The correction removes the numeric false positive there
  too. Child lease release still requires a terminal invocation with confirmed
  native/process/tool finality.
- `runBatch` propagates persistence failures through `fail`, cancels the parent,
  and waits for its workers. `close` returns the recorded fatal error; `harness.go`
  can attach it to `stopErr`, and `sqlite_execution_finish.go` emits the reported
  process-stop wording. These error/finality rules were not weakened.
- `initializeDelegation` consults `activeDelegation`, retaining the continuation
  block for unsettled work, uncertain delivery or unconfirmed finality.

## Existing recovery after using the corrected binary

Recovery runs at startup, before ordinary/graph recovery and engine readiness.
This source change does not repair a running engine or clear its pending batches
when a user simply sends `continue`.

For a pending, nonterminal invocation whose persisted `DispatchAttempted` is
false, `recoverDelegationRuntime` already treats the absence of dispatch as
confirmed native/process/tool finality. It records an `interrupted` outcome
without replay, releases that child's lease through the normal finalization
checks, and settles the batch after all members are terminal. Pending result
deliveries/aliases become `cancelled`; deliveries already `sending` become
`uncertain`. The parent's queue is recovered without replaying its input.

Consequently, a batch affected solely by the rejected pre-dispatch update can
be finalized by existing startup recovery with this decoding fix. Once its
pending deliveries are cancelled and there is no other uncertainty, it no longer
matches `activeDelegation`. Records and historical errors remain available.

For a nonterminal invocation with `DispatchAttempted = true`, restart alone is
not proof of native or remote completion, even if `NativeAccepted` is false.
Recovery records interruption with unconfirmed finality, retains the active
lease/quarantine, and settles the batch as `finality_unconfirmed`. Continuation
remains blocked. Already-terminal invocations retain their existing evidence;
confirmed terminal members are not automatically reclassified as uncertain.
Existing `uncertain` result deliveries also remain blocking. This fix adds no
automatic reconciliation or forced unlock for these cases.

## Lightweight verification

Run from `engine`:

```powershell
go test -run '^TestDelegationInvocationNumericCaptureValidation$' -count=1 .
```

Final result: `ok klm/engine` (1.177 s). An initial run exposed a fixture-only
snapshot seal mismatch from raw schema key ordering; the fixture was corrected
to use decoded schema maps before the passing run.

Coverage includes state/dispatch-only persistence, normal item finalization,
terminal no-op writes, exact input/output schema numbers (`50`,
`9007199254740993`, `0.123456789012345678901`, `1e-400`), and rejection of real
task/owner/model/schema-number changes. The adjacent large-integer mutation is
rejected even after resealing the modified snapshot. This check compiles the Go
package and exercises temporary repository records; it is not runtime acceptance
of a real harness, MCP, result-delivery flow or startup recovery.

No full suite was run. The real engine database was not accessed or modified,
and the running engine was not restarted. No commit/push was performed.

## Pending human validation

1. In an isolated runtime/data directory, use an OpenCode parent and the saved
   Plan Writer specialist with its ai-memory `memory_*` selection. Confirm the
   actual captured catalog contains numeric schema constraints. Run
   `subagent_run`, confirm child execution and all-child result delivery, then
   continue the parent without either reported persistence/continuation error.
2. Validate interruption/Stop and normal finalization with the same specialist;
   confirm finality evidence, lease lifecycle and correlated result delivery.
3. On an intentional separate copy of affected data, validate startup recovery
   with the corrected binary. Inspect persisted dispatch evidence first: a fully
   pre-dispatch batch should become interrupted/cancelled without replay and allow
   later continuation when no other uncertainty exists. Nonterminal dispatched
   members and uncertain deliveries must retain their protective block. Do not
   delete rows or clear leases to manufacture this outcome.

The startup recovery description above is based on code inspection. Neither a
recovery runtime probe nor the real Plan Writer/ai-memory flow was executed for
this change. Any deployment/restart or work on actual affected data is a later,
explicit human validation step.
