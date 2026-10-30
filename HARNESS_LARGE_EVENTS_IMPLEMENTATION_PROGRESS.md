# Harness large-event correction — implementation progress

Date: 2026-10-29.
Branch: `fix/harness-large-events`, based on `89d0cb4` (`release/0.6.0`).
Plan: `HARNESS_LARGE_EVENTS_CORRECTION_PLAN.md`.

## Current delivery

The A/B mitigation is committed in `0c28365` and PR #39. The continuation adds
the C useful-content path and D adapter/read-path coverage, described below.
Synthetic checks and lightweight builds are implementation evidence; the E human
runtime/rollout gate remains separate and has not been accepted implicitly.

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

### Original B capacities (superseded by the continuation below)

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

## C/D continuation

### Ingestion and adapter state

- Native JSON is validated once before projection. The same reader validates
  discarded values, UTF-8, escapes, nesting, framing and duplicate decoded keys.
  Wide key sets spill into the existing private staging index instead of growing
  a resident map. Unknown useful fields remain backed by their complete source.
- Projected native envelopes retain bounded metadata and immutable byte-range
  handles. Large strings, arrays and result objects are not `RawMessage`, giant
  `Decoder.Token` strings, or full adapter maps. A handle refuses ordinary JSON
  marshaling, so it cannot become an authorization scope or actual tool value.
- Three small admission gates bound simultaneous projection, normalization and
  persistence work to eight operations each. Envelope/cache reservations have an
  8 MiB per-turn and 64 MiB shared resident budget. Native handoff stays synchronous;
  batched text is capped at 1 MiB and charged until persistence completes.
- Staging accounts native files, canonical temporary files, unpublished SQL bytes
  and private key indexes against 2 GiB; disk writes check the free-space reserve.
  Complete large frames use the resource budget rather than the old 2/4 MiB
  useful-content limits. Explicit small-only helpers retain bounded limits.
- Source frames have transport leases and cache references. Replaced/terminal
  OpenCode parts and Codex items retire their payloads after canonical publication;
  Pi retires command sources after tool completion. Closing the turn clears the
  remaining native snapshots. No per-frame worker/goroutine was introduced.
- JSONL/SSE incomplete frames have 30-second progress/idle handling. Native
  startup/history waits account for ongoing ingestion; HTTP history times blocked
  reads rather than imposing a tiny total duration on healthy transfers.

### Harness coverage

- OpenCode live text/reasoning/tool output and native children use content sources.
  Deltas use bounded append operations and an owned accumulation file instead of
  repeatedly concatenating the entire part. Terminal part caches keep identity and
  status. Permission/question fields are not part of summary omission.
- OpenCode HTTP pages are framed message by message. Initial token accounting is
  folded incrementally and history ID/cursor sets spill when necessary. Final
  reconciliation keeps compact message metadata, then reads one authoritative
  message at a time in chronological order. It does not retain an entire final
  native transcript. Native history is not rewritten.
- Pi JSONL handles useful `message_end`, reasoning, command/tool/MCP output and
  custom data through the same source contract. Redundant aggregates remain the
  exact A/B omissions. Its authenticated loopback child HTTP channel also stages
  event bodies before decoding compact metadata; large child output no longer
  inherits the former 2 MiB request-body allocation/limit.
- Codex notifications, responses and server requests preserve JSON-RPC numeric IDs.
  Command output, file changes, assistant/reasoning and MCP results use bounded
  sources. `thread/read(includeTurns:true)` iterates returned turns/items; native
  persisted-history verification and the exact Choice receipt remain authoritative.
- Oversized human-decision/elicitation/approval scope is rejected explicitly before
  any positive decision, including under YOLO. File-change data is not silently
  narrowed to a preview or an opaque remembered grant. Unsupported control metadata
  fails the owning operation instead of inventing successful settlement.

### SQLite publication and readers

- Additive schema version 2 introduces unpublished content generations and an
  event-to-generation publication pointer. Version-1 records keep their existing
  chunk path; the migration is transactional/repeatable and runs only after SQL
  authority verification. Existing histories/backups are not rewritten.
- Canonical JSON is written incrementally outside `app.mu`. Bounded batches stage
  SQL chunks with no transaction waiting for native/network input. A short normal
  session transaction publishes identity/ordinal/indexes/revision/digest together.
  Client invalidation is published after commit.
- Small text deltas keep the existing chunk/hash-checkpoint path. Appending to a
  large generation shares an immutable prefix plus an append range; it does not
  copy the existing large body or create a recursive generation chain per delta.
  Metadata changes reuse the canonical writer outside the publication transaction.
- Failed/cancelled generations remain undiscoverable. Active stages are protected
  by ownership tracking; startup and the existing runtime maintenance worker reclaim
  unreferenced generations in bounded transactions. Shared prefixes remain pinned
  by SQL references and read snapshots.
- History/content/export, search, linked/general reads and source excerpts use the
  generation-aware chunk reader. Text windows preserve Unicode-code-point offsets,
  without `[]rune(fullText)`. Whole-event fallback helpers are small-only.
- Metadata finalization/favorites prepare large changes before entering the writer
  gate. A capacity rejection remains an owning-operation failure, keeps prior
  committed content and does not set global `storageErr`. Actual SQL write failures
  retain storage protection.

### Clients and downloads

- Storage contract 3 is retained. Large history rows add explicitly named preview
  metadata/text; canonical `text` is not replaced by an undisclosed prefix.
- Details retain bounded pages and digest cancellation. Event and conversation
  downloads request an authenticated 256-bit, single-use ticket with a 60-second
  expiry. Consumption rechecks client revocation, session/graph ownership and the
  bound event digest or conversation revision. Tickets disappear on restart.
- The attachment streams SQL content directly, with progress-based writes and
  cancellation. At most two downloads pin read snapshots, leaving reader capacity
  for other clients. Ticket URLs are not logged and contain no persistent credential.
- Browser navigation uses the attachment response; Tauri and Flutter delegate the
  scoped ticket to the platform browser/download handler. The mobile bridge checks
  the exact frontend origin. There is no all-pages Blob fallback. The ordinary
  conversation export button uses the same ticket mechanism.

## Additional focused evidence

Windows, generated/sanitized input only:

- Useful 25 MiB OpenCode text, Pi final/tool output, Codex command/file-change
  output: passed with durable content references and bounded Unicode tail reads.
- Generation publication/reopen/export digest, virtual append without copying old
  chunks, metadata/favorite mutation, active-stage cleanup ownership: passed.
- Explicit content-capacity failure keeps another conversation writable; a genuine
  SQL staging failure activates storage protection and publishes no partial event.
- OpenCode native HTTP history: a 25 MiB summary plus 25 MiB useful text is projected
  without retaining summary patches, including reading the source after HTTP closes.
- A 20,000-key useful object spills duplicate-key bookkeeping and rejects a duplicate
  in its drained tail without materializing the result object.
- The desktop production build, `cargo check --locked` and targeted Flutter analysis
  passed during implementation. Updated final check results follow below.

### GC-sampled live heap probe

Command: `go test -tags dev -run '^TestHarnessLargeContentLiveHeap$' -count=1 -v .`

The probe generates native Pi tool results incrementally, normalizes/persists the
complete result, and samples `HeapAlloc` after GC every 100 ms. Baseline is taken
after opening the isolated SQLite fixture. This measures sampled live Go heap,
not allocations-per-call, process RSS, a browser heap or real provider execution.

| Generated useful content | Baseline heap | Peak sampled live heap | Additional live heap |
| --- | ---: | ---: | ---: |
| 1 MiB | 924,208 | 1,269,880 | 345,672 |
| 25 MiB | 959,688 | 2,092,360 | 1,132,672 |
| 100 MiB | 1,004,744 | 3,601,744 | 2,597,000 |
| 256 MiB | 1,036,344 | 3,668,616 | 2,632,272 |

Passed in 43.26 s (package time 46.176 s). The sampled additional heap stays below
16 MiB and does not grow approximately in proportion to payload size. This is
focused synthetic evidence, not acceptance of every installed-harness/runtime path.

## Human/runtime gate still required (E)

Real OpenCode/Pi/Codex provider turns, background native-child ordering, Stop under
real backpressure, interrupted-transfer/restart behavior, browser/Tauri/mobile
attachment handling, client heap and aggregate latency/writer-lock measurements
remain explicit acceptance checks. No model task, installer/release rollout,
production-data rewrite, full test suite or review agent has been run implicitly.

## Final lightweight checks for the continuation

- Focused Go command (46.833 s):
  `go test -tags dev -run 'Test(Protocol|JSONLines|Harness(Streamed|LargeUseful|Content)|SQLite(ContentGeneration|GenerationCleanup|EventStream|RecoverSession)|ContentDownloadTickets|OpenCode(History|NativeHistory|Usage)|GraphMCP)' -count=1 .`
  — passed.
- Conversation-export ticket/attachment coverage added to `TestContentDownloadTickets`
  — passed separately (2.732 s) after adding the export path.
- Desktop `npm run build` — passed (TypeScript and Vite; existing large-chunk warning).
- Tauri `cargo check --locked` — passed (17.17 s) with the updated frontend assets.
- Flutter `flutter analyze lib/features/web/host_webview_page.dart` — no issues.
- Linux amd64 `go build -tags dev` — cross-build passed; artifact was written to
  the approved temporary directory, not installed or run against production data.
- `git diff --check` — no whitespace errors.

The implementation is ready for the plan's human/runtime acceptance gate. This
record does not mark that gate or installer/release rollout complete.
