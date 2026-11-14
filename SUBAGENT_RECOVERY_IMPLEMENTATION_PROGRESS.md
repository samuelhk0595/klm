# Subagent recovery — 2026-11-12

## Observed incident

Read-only inspection of the installed engine's SQLite records located project
**Werbert**, conversation **One Agent** (`875dc804a3a6e599630e03fe455da633`).
Batch `dac2c236ae5cdcb832258c7ae9c11af4`, operation
`factory-implementation-batch-07`, settled at 2026-11-12 08:29 UTC with three
completed children and one `finality_unconfirmed` child (`graph-suspension`,
invocation `6e50b5325f390224524c3e26d03460e2`). Its error is
`Could not read the OpenCode event stream.` The exact underlying transport cause
was not retained by the existing parser and is not established by these records.

For that child, `nativeSettled=false`, `processesDrained=true`, and
`toolCallsSettled=true`, with no uncertain tool calls. The batch delivery was
cancelled. Its immutable unconfirmed outcome still matches `activeDelegation`, so
every subsequent parent startup fails. Stop/restart alone cannot resolve it.

## Implemented recovery contract

- A stopped blocked conversation presents **Recover subagents**, followed by
  **Continue without replay** confirmation. Existing child histories remain available.
- The authenticated public API accepts
  `POST /api/sessions/{id}/delegation/batches/{batchId}/recover` with
  `{action:"continue_without_replay", operationId, revision}`. No model-facing
  recovery tool is exposed.
- Exact ownership, batch CAS revision, restoration and absence of live parent/child
  turns/runtimes are checked under the engine lock. All children must be terminal;
  process and tool drainage (including absence of uncertain calls) are required.
  Missing native completion alone can be acknowledged without claiming success.
- An immutable recovery decision changes only delivery to `abandoned`. Original
  outcomes/digests/results, native aliases and quarantined child leases remain intact.
  Identical operation retries return the existing decision; other decisions conflict.
- Recovery records an engine-attributed status event and pauses the queue. It does
  not submit a prompt, replay children or dispatch existing queued messages. The user
  sends a new message or explicitly uses Send now afterward.
- Subsequent parent turns receive one bounded summary of the most recently recovered
  batch, including all children and full-output references. It is reference context,
  not a fabricated native tool result, human message or authorization. The same
  recovered operation ID cannot admit/replay another batch.
- SQLite publication schema **5** refines the active-parent unique index to exclude
  explicitly abandoned batches. Upgrade is transactional; no existing record is
  reclassified. Older engines reject schema 5 rather than silently misinterpreting it.
- Graph-node recovery remains the existing explicit graph corrective-attempt flow.
  Unconfirmed processes or remote tool calls stay blocked rather than being cleared
  by a user assertion. Startup never auto-acknowledges the recovery decision.

## Validation and rollout

Checks completed:

- `go test -tags dev -run '^(TestDelegationContracts|TestPublicationSchemaChain)$' .`
  passed. These focused checks cover recovery eligibility, preserved uncertain
  outcomes/quarantine, explicit-decision transitions, and schema upgrade/reopen
  from versions 1–5. This also compiled the engine and its test package.
- `npx tsc --noEmit` in `clients/desktop` passed.

Production data was
inspected read-only; the installed engine and the affected conversation have not
been modified by this implementation. Applying it requires updating the engine
and shared frontend. In One Agent, choose **Recover subagents → Continue without
replay**, then send the continuation. Review existing changes and the saved child
results before repeating any unfinished task; the uncertain child cannot be reused.

Pending human acceptance: the recovery card and confirmation on desktop/Focus/side
chat/mobile WebView, lost HTTP acknowledgement, a real subsequent native turn,
quarantine retention, and recovery persistence after restart.
