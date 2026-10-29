# Harness large-event correction — implementation progress

Date: 2026-10-29.
Branch: `fix/harness-large-events`, based on `89d0cb4` (`release/0.6.0`).
Plan: `HARNESS_LARGE_EVENTS_CORRECTION_PLAN.md`.

## Current delivery

First A/B slice: discard-only correction for redundant OpenCode/Pi data.
Implementation and focused synthetic checks are recorded below. Human runtime
acceptance is pending. The full cross-harness large-content correction is not
complete; C/D/E and the remaining operational gates still apply.

## Contract and evidence (A)

Managed compatibility targets remain OpenCode **1.18.35**, Pi **0.85.1**, and
Codex **0.153.4**. This work does not change version admission.

- Installed `opencode-ai/package.json` reports 1.18.35. The observed summary
  events/byte counts are recorded in the plan; no real session payload was copied.
- Installed Pi `package.json` reports 0.85.1. Its `dist/modes/json-event.js`
  strips partial snapshots from `message_update`, while preserving usage and
  assistant delta metadata. `dist/modes/rpc/rpc-mode.js` serializes `toJsonEvent`.
- Installed `pi-agent-core/dist/agent-loop.js` emits `turn_end.message` and
  `turn_end.toolResults`, and `agent_end.messages`, after authoritative individual
  message/tool events. KLM's existing `adapter.pi` ignores these aggregate bodies.
- Pi settlement remains the existing `agent_settled`/managed child-state path in
  `pi_interactive.go`; neither omission nor native `agent_end` creates settlement.
- The actual offending large Pi event remains unidentified. Useful final messages,
  tool results, extension requests and errors are not classified as aggregates.

### Explicit field policies

| Transport/type | Exact decoded path | Treatment |
| --- | --- | --- |
| OpenCode SSE, `message.updated` | `properties.info.summary.diffs` | Validate/drain entire value, replace with `null` in the internal projection. Preserve sibling info/identity/control fields. |
| OpenCode HTTP message history | `[].info.summary.diffs` | Same omission, including user-message updates. |
| OpenCode HTTP message history | `[].parts[].state.metadata.diff`, `[].parts[].state.metadata.files[].patch` | Preserve established small previews; replace strings above 64 KiB with the existing explicit omission label. |
| Pi JSONL, `agent_end` | `messages` | Validate/drain redundant aggregate, replace with `null`. |
| Pi JSONL, `turn_end` | `message`, `toolResults` | Validate/drain redundant aggregates, replace with `null`. Preserve other fields. |
| Any other event/path | All fields | Preserve through the existing retained-content capacity, or fail explicitly. No generic `patch`/`diff` omission. |

OpenCode `session.diff` is deliberately **not** classified for omission in this
slice. Its supported-version presentation-only semantics still need confirmation.
Permission/question scopes, tool input/output and Codex file changes are not added
to the discard policy. Native databases, histories and model context are unchanged.

## Implementation (B, initial slice)

- Shared incremental JSON validator/projector in `protocol_json_projection.go`.
  No whole discarded string, array or native event map is decoded. Raw UTF-8,
  escapes, JSON grammar, nesting and decoded duplicate object keys are validated,
  including in discarded data. Duplicate keys are rejected rather than resolved
  differently during filtering and authorization.
- Rewindable staging in `protocol_staging.go`: retain up to 64 KiB of raw input,
  then use private temporary files below `<engine-data-dir>/ingestion-staging`.
  A first streaming pass validates/classifies the complete frame; a second
  applies its policy. This supports late type/identity keys without treating
  unknown content as disposable. It intentionally incurs disk I/O for large
  optional frames; single-pass routing is a possible later optimization.
- Staging accounts aggregate bytes against 2 GiB and checks a 512 MiB free-disk
  reserve before disk writes. Windows/Linux free-space helpers are implemented.
  Capacity/I/O failures remain owning-operation errors and do not set `storageErr`.
  Failed file removal remains charged to the staging budget until restart.
- Normal/error/cancellation paths close staging files. After taking the existing
  exclusive data-directory lock, startup clears abandoned unpublished staging
  before workers can use it. Nothing in this staging directory is canonical
  content, a durable event generation, or a native history archive.
  The OpenCode turn waits for its cancelled/closed SSE reader to release staging
  before exiting; Pi also releases staged input on process attachment failure.
- `protocol_sse.go` reads bounded byte blocks and recognizes LF, CRLF, CR,
  multiline/empty `data` fields, initial BOM, comments and ignored SSE fields.
  An incomplete frame is never published. Disconnection remains an explicit
  failure; no automatic reconnect/replay or assumed completion was added.
- Pi's runtime JSONL reader stages incrementally across `Write` boundaries;
  publication occurs only after complete validation. LF and escaped newlines
  retain their separate meanings. Existing valid final-line flushing is retained.
- The projected live paths use synchronous, unbuffered event handoff, replacing
  the previous queue of 64 decoded native maps. Cancellation is checked during
  reads/projection/staging and while waiting for handoff. No per-frame goroutine
  or writer is created.
- OpenCode HTTP history uses the same validator/policy primitives. Existing
  tool-preview behavior and history pagination are retained.
- Errors include the harness, an allowlisted event-type label and input byte
  count where classification is available. Unknown native labels are reported as
  `unclassified`; raw JSON, prompts, results and credentials are not diagnostics.
  Owning session/turn attribution remains the existing execution-error path.

### Explicit current capacities and pending B gates

- Retained useful/unknown event JSON still has the existing 2 MiB OpenCode live
  and 4 MiB Pi capacities. Oversized useful data fails explicitly; it is not
  replaced with an aggregate omission or silently clipped.
- OpenCode HTTP history still retains bounded whole pages (32 MiB projected,
  256 MiB wire); this does not solve enormous useful messages.
- Parser bookkeeping has a 256-depth limit, 4 KiB encoded keys, 1 KiB encoded
  top-level type and an estimated 1 MiB live duplicate-key budget. This is not
  yet the planned normalized-envelope/decoded-map budget.
- Codex, catalog probes and legacy nonprojected JSONL callers retain their current
  reader. C/D must migrate useful content, responses and native history before
  removing their fatal size branches.
- Unbuffered handoff removes queued maps but does **not** establish the planned
  engine-wide 64 MiB resident admission/fairness budget or bounded adapter caches.
- Per-incomplete-frame progress/idle deadlines, peak live-heap/concurrency evidence,
  and real Stop-under-backpressure acceptance are still pending. Existing process/
  request cancellation remains in force. HTTP history retains its existing request
  timeout; no healthy large useful-transfer guarantee is claimed.
- The native OpenCode/Pi policies are recorded above; the complete Codex policy
  and internal-consumer inventory remain part of C/D. This is not full A/B exit.

## Focused verification

Executed in the worktree's `engine` directory on Windows:

```powershell
go test -tags dev -run 'Test(Protocol|JSONLines|OpenCodeHistory)' -count=1 .
```

Initial run: passed (`ok klm/engine`, 8.268 s).
Final run after cancellation/accounting/reader-ownership adjustments: passed
(`ok klm/engine`, 9.838 s). `git diff --check` also reported no whitespace errors.

Synthetic/generated readers cover:

- OpenCode summaries of 25 MiB and 100 MiB, late type, escaped policy key,
  small retained identity/title, and the following native error event.
- Pi `agent_end`/`turn_end` aggregates of 25 MiB, late type, preserved correlation,
  and the following separate `agent_settled` event.
- OpenCode HTTP history with a 25 MiB summary on a user message, alongside the
  existing large/small tool-preview and useful-content-limit tests.
- SSE multiline data, LF/CRLF/CR, empty data, comments, Unicode and an incomplete
  final frame; JSON invalid escapes/UTF-8/numbers, trailing data and duplicate tags.
- Exact-path filtering; small assistant/tool/permission data preserved unchanged;
  useful 5 MiB Pi content rejected explicitly rather than discarded.
- Staging cleanup, cooperative cancellation and shared staging-capacity rejection.

No full suite, harness/model session, installer, production-data migration or
release rollout was run. Go tests compile the engine/test package; Linux helpers
have not been runtime-validated. Generated input is not a measurement of peak live
heap, UI memory or multi-session fairness.

## Next delivery (C)

1. Introduce KLM-owned staged useful-content generations and short atomic SQLite
   publication, preserving event identity/revision and existing chunks.
2. Carry compact handles/previews through adapters instead of strings/maps, first
   for OpenCode tool output and assistant text. Add resident-byte admission and
   incomplete-frame deadlines; retain already committed deltas on later failure.
3. Convert the reached `eventByID`/whole-event consumers: linked/general reads,
   source/final-answer/graph readers and export. Audit paths named in the plan;
   do not treat previews as answers or authorization scopes.
4. Reuse digest-bound content pages for details and implement authenticated,
   single-use ticketed streaming downloads across browser/Tauri/mobile WebView.
5. Then migrate Pi/Codex useful events, native children/history and human-decision
   details (D), measure the operational budgets and obtain human validation (E).

For human validation of this slice, use a separately built dev engine from this
worktree and exercise a real OpenCode summary/Pi aggregate-producing turn, including
Stop. Large useful results are still expected to hit the documented old capacities
until the next delivery.
