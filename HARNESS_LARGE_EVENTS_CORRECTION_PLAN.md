# Harness large-event correction plan

Status: proposed implementation plan. No runtime change or rollout is authorized by this document alone.
Date: 2026-10-29.

## 1. Objective and completion boundary

Make OpenCode, Pi and Codex ingestion tolerate legitimately large events without
requiring an equally large in-memory JSON object, losing useful content, or
cancelling a turn because an unused display field exceeded a small frame limit.

The correction must span transport, parsing, adapter state, persistence, internal
readers and clients. Increasing 2 MiB to 32 MiB, truncating every output, or moving
the same allocation into SQLite/the browser does not meet this objective.

Deliver in small increments. The first increment fixes the observed optional-data
failure; it must not be described as completion of the cross-harness correction.
Normal execution remains SQLite-only. Removing legacy persistence is a separate
recorded follow-up, not a prerequisite or part of this change.

## 2. Evidence and current failure points

### Observed failures

- KLM session `f226c76da02b97b65576c076f3edcf1c`, native OpenCode session
  `ses_e7a157ff6ffeP4gnIHWkmwAgA6`, failed on 2026-10-29 at local 18:58:44 and
  19:05:35 with `OpenCode event frame exceeded KLM's 2 MiB limit`.
- Native event records contain `message.updated` payloads of 25,796,474 and
  3,839,631 bytes, dominated by `info.summary.diffs[].patch`, with 3,241 and 86
  file entries respectively. The corresponding native message updates precede
  the errors. These payloads were stored by OpenCode, not accepted by KLM.
- The user also reported Pi rejecting a JSON line over 4 MiB after 10m55s.
  The exact offending Pi event has not been identified. Do not assume it was
  `agent_end`; retain coverage for useful large results and final messages too.

### Code boundaries to change or verify

| Boundary | Present behavior |
| --- | --- |
| `engine/opencode_interactive.go`, `openCodeEvents` | Scanner materializes a complete SSE line/frame under 2 MiB before JSON decoding and semantic filtering. |
| Same file, `putPart`, text accumulation, permission/question parsing | Additional 2 MiB checks can reject data even after fixing SSE. Maps retain entire native parts. |
| `engine/interactive.go`, `engine/harness.go`, `jsonLines` | Pi/Codex JSONL line is accumulated under 4 MiB, then decoded into a map. Overflow cancels the owned process. |
| `engine/harness.go`, Pi adapter | `agent_end` and `turn_end` are ignored only after ingestion. Tool results retain raw native fields; useful output can be duplicated in text and data. |
| `engine/codex_interactive.go` | Full items, file diffs, aggregated output, tool results and some history responses pass through the same JSONL reader. Graph verification uses `thread/read` with `includeTurns:true`. |
| `engine/opencode_history_projection.go` | Selectively omits large tool diff previews for HTTP history only. Does not cover SSE or `info.summary.diffs`. |
| `engine/sqlite_session_repository.go`, `sqlite_event_stream.go` | Storage is chunked, but writers first receive `Event` strings/maps and encode complete values. A 256 MiB record limit and whole-event fallback reads remain. |
| `engine/sqlite_linked_execution.go` | `linked_read` decodes a whole event and converts its text to runes before returning a small excerpt. |
| `engine/sqlite_general_history.go`, `sqlite_content.go` | Already offer bounded, digest-bound content reads; reuse these semantics. |
| `clients/desktop/src/features/chat/EventContent.tsx` | Viewing is paginated; download accumulates all pages into a browser Blob. |
| `ConversationEvents.tsx`, `SessionEvent.tsx` | Details can stringify/render native data in full when supplied as an inline event. |

## 3. Required semantics: decide by field, not only byte count

Maintain a small explicit policy table per supported protocol/version. Unknown
fields are not automatically disposable. Separate native input from normalized
KLM events; arbitrary native objects must not remain the universal adapter API.

| Class | Examples | Required treatment |
| --- | --- | --- |
| Control | IDs, correlation, turn completion/error, tool identity, usage, steering acknowledgements, graph receipts | Keep exact bounded metadata. Preserve order, identity and existing state-machine meaning. |
| Human decision | Permission operation/scope, file changes to approve, questions/options, MCP elicitation | Preserve full semantics. Large descriptions/details may be paged, but authorization cannot be inferred from a preview or shortened scope. |
| Useful content | User/assistant text, reasoning exposed by the harness, command output, tool/MCP results, file-change details | Retain the complete supported content in KLM-owned storage, with a small preview and bounded reads. No silent clipping of canonical content. |
| Redundant/unused presentation | OpenCode `info.summary.diffs`; repeated Pi `agent_end` messages and `turn_end` content already represented by authoritative events | Validate and drain without retaining, only after the protocol policy proves it is safe. Do not discard associated control fields or synthesize successful completion. |
| Unknown | New event type or unclassified field | Bound/spool rather than blindly retain or silently drop. Unknown control semantics must fail explicitly or remain unsupported according to the adapter's existing compatibility rules. |

OpenCode summaries are not permission diffs. A field named `patch` or `diff` is
not sufficient grounds for omission: policy must match event type and exact path.
Keep native harness history intact. KLM presentation filtering does not rewrite
the harness's model context, workspace files, native database or tool results.

## 4. Target ingestion design

### 4.1 Incremental framing and projection

Implement a small shared input layer with JSONL and SSE framing adapters:

1. Read bounded byte blocks; do not require Scanner to deliver an entire line.
2. Parse JSON structure and string contents incrementally. `json.Decoder.Decode`
   into `map[string]any`, `RawMessage` for the entire event, or `Decoder.Token`
   returning one giant string is not an acceptable large-field implementation.
3. Route fields to small metadata, discard sinks or temporary content sinks.
4. Complete/validate the frame before publishing its durable normalized event
   or applying terminal/control transitions.

Support arbitrary JSON key order: `type`, method, role and session IDs may follow
large fields. Stage uncertain fields under a bounded disk budget until policy
and ownership are known. Do not require upstream key-order changes.

JSONL must preserve LF framing and escaped newlines. SSE must preserve blank-line
boundaries, multiline `data:`, CRLF, comments/heartbeats and reconnect semantics.
Validate UTF-8, escaping, nesting and complete JSON even for discarded values.
Handle duplicate policy/identity keys unambiguously; do not filter using one
interpretation and authorize using another.

Only carry compact envelopes and content handles through Go channels. Replace
the implicit `64 frames` memory budget with byte accounting as well as an item
count. Bound producers, queued envelopes, staging work and live adapter caches.
Retain cancellation/Stop responsiveness under backpressure; do not start an
unbounded goroutine or writer for each frame.

### 4.2 Content handles, not giant strings/maps

Introduce an internal representation carrying native identity, normalized event
metadata, previews and streaming sources for large text/data fields. Small events
keep the ordinary lightweight path. A handle must not escape as if it were the
actual result value in tool semantics or human decisions.

Adapters must stop retaining full payloads in `infos`, `parts`, `items`, usage
maps, final batches and child-session caches. Retain metadata and generation
handles only; release superseded native snapshots after no active reader needs
them. Apply bounded append/replace operations using the existing native event ID.
Avoid repeated full copies of an accumulated text on every delta.

### 4.3 Initial operational budgets

These are implementation defaults to measure, not new universal payload limits:

- 64 KiB I/O/content chunks; previews at most 4 KiB per textual field.
- At most 1 MiB of resident normalized envelope metadata; larger useful arrays
  become streamed/indexed collections rather than one resident envelope.
- At most 8 MiB of queued resident ingestion data per active turn and 64 MiB
  engine-wide; enforce byte reservations before enqueueing and fair consumption
  across sessions, so concurrency cannot multiply this without admission.
- Temporary staging initially capped at 2 GiB engine-wide, reserving 512 MiB of
  free disk. Account for concurrent turns, replacement generations and cleanup.
  Reuse the engine data-directory ownership rules; never store secrets in logs.
- An incomplete frame has progress/idle deadlines and cancellation, rather than
  inheriting a tiny fixed total duration that rejects a healthy large transfer.

The goal is memory independent of payload size, not acceptance of infinite input.
If actual disk/CPU/protocol capacity is exhausted, report the specific cause and
scope it to the owning operation/turn. A content-budget rejection is not a database
failure and must not set global `storageErr`. Genuine persistence failure retains
the existing data-integrity protection. Never silently lose required content.

## 5. Persistence and read-path design

### Atomic, bounded publication

- Stage large content outside `app.mu` and outside a long-lived SQLite writer
  transaction. Do not hold the global writer while waiting for native/network I/O.
- Persist bounded chunks into an unpublished content generation. Use explicit
  owner/turn/native-event identity, length and digest; readers cannot discover a
  generation before its complete frame is validated and committed.
- A short final transaction publishes the generation, event metadata, ordinal,
  indexes and revision together. Publish client invalidations only after commit.
  An incomplete final event must never settle a tool, queue entry or graph node.
- Preserve already committed deltas if a later snapshot/frame fails. Failed new
  generations must not overwrite earlier valid content. Avoid copying a whole
  large generation inside the final transaction just to change its visible key.
- Clean uncommitted generations after cancellation, process death and restart.
  Keep generations referenced by committed events or active bounded readers;
  ownership-aware cleanup must not race another active turn.

### Reuse the canonical event contract

Retain `data_session_events` identity/index/revision semantics and lossless KLM
event JSON on export. Extend content storage with generation addressing only where
the current fixed record-key chunks cannot publish atomically without large copy
transactions. Existing records continue to read through their current chunk path.
Any new tables require a versioned, transactional, repeatable SQLite migration;
do not rewrite all histories, delete legacy backups or fall back to JSON storage.

Write canonical event JSON incrementally from staged fields. Do not route the
large path back through `Event.Text`, `Event.Data`, `json.Marshal` or the existing
256 MiB whole-record allocation. Whole-record limits may remain on explicitly
small-only helpers, but cannot define capacity for the new streaming content path.

Keep public history/content responses compatible with contract 3 where possible:
small inline events and existing digest-bound, paginated content references.
An additive preview must be explicitly labeled and never replace canonical `text`
with an undisclosed prefix. If a required wire change cannot be represented safely,
version/negotiate it with all clients before enabling it; do not silently break old
desktop/browser/mobile clients. Do not assume a new wire version is necessary.

### Internal consumers are part of acceptance

Inventory every `eventByID`, whole `session_events` decode and native history read
reached by changed flows. Convert text excerpts/search, linked/general-agent reads,
final-answer extraction, graph settlement/receipts, source context and exports to
bounded field readers or streamed reconstruction. Preserve current character vs
byte offset semantics and digest/revision validation. A preview is never a final
answer, graph output, authorization reference or complete search domain.

Useful large native data remains inspectable through bounded details/export. Do
not promise byte-for-byte native transport archives: the canonical KLM event is
normalized and expressly omits policy-approved redundant fields. Avoid storing
duplicate large output in both text and native details when a shared internal
content generation can reconstruct the established output representation.

## 6. Harness-specific work

### OpenCode

- Apply the policy to live SSE **and** startup/final HTTP history. Cover
  `message.updated.properties.info.summary.diffs` (wire envelope path) and
  history `[].info.summary.diffs`, plus the tool-preview paths already handled.
- Drain an ignored `session.diff` only after confirming it is a presentation-only
  event for the supported version. Preserve permission/file-change semantics.
- Remove the whole-map and 2 MiB assumptions from `putPart`, accumulated text,
  native-child parts and `infos`/usage caches; replacing Scanner alone is insufficient.
- Stream HTTP pages message by message. Pagination by 64/32/1 messages does not
  solve a single enormous message. Share field policy across both transports.
- Perform ownership and baseline checks using small metadata; large data for
  unrelated/old messages must not poison the active turn. Validate/drain it safely.
- Preserve error/completion ordering and final reconciliation. Never replace a
  stream error by assumed native success or replay uncertain input automatically.

### Pi

- Supply a Pi projection policy to the shared JSONL reader. Recognize discarded
  aggregate message bodies before creating the current map, including late tags.
- Handle `message_end`, `tool_execution_*`, custom/extension events and images or
  other supported result data through content handles. Keep useful final output.
- Preserve `agent_settled`, errors, RPC response IDs, native-child routing,
  steering consumption, permissions/questions and usage. Ignoring `agent_end`
  must not invent `agent_settled` or discard settlement work.
- Inspect supported installed Pi event serialization; current message deltas may
  already omit partial snapshots. Do not assume that bounds final/aggregate events.
- Add sanitized event-type/byte-count diagnostics so the next Pi failure is
  attributable without recording prompts, results, credentials or raw JSON.

### Codex

- Use the shared JSONL machinery with a Codex-specific policy, covering responses
  as well as notifications and server requests. Keep JSON-RPC correlation intact.
- Stream `commandExecution.aggregatedOutput`, `fileChange.changes`, MCP results,
  assistant items and corresponding deltas; retire superseded item payloads.
- Audit `thread/start`, `thread/resume`, `thread/read(includeTurns:true)` and graph
  final verification. Iterate returned turns/items instead of decoding a complete
  transcript. Native persisted-history verification must remain authoritative.
- Preserve approval scopes and `item/completed` / turn completion / graph Choice
  ordering. Never omit a file diff that determines the operation being approved.

## 7. Client behavior

- Reuse the shared frontend for desktop, browser and Flutter WebView. No separate
  Flutter chat implementation or redesign is needed.
- Large events retain identity/type/status and a bounded preview. Open details
  fetches bounded content; changing the digest cancels stale requests as today.
- Do not pass giant metadata objects to `JSON.stringify`, Markdown renderers,
  component props or global session caches. Limit rendered pages/cache bytes.
- Replace `EventContent.download()` accumulation with a streaming attachment
  response. Since a normal browser navigation cannot attach the engine's auth
  header, create a download ticket through the existing authenticated request:
  single-use, unguessable, 60-second expiry, bound to client/session/event/digest.
  The ticket authorizes only that content; recheck client revocation and ownership
  when consuming it, do not log its URL, and expire it on engine restart. Stream
  the response to the platform download handler, not a JavaScript Blob. Keep any
  immutable content generation needed by the in-progress download alive.
  Verify Tauri and mobile WebView download handling as well as browser navigation;
  implement the required native download callback if the WebView does not handle
  attachments. No persistent credentials in URLs or whole-file-memory fallback.
- For oversized human-decision details, offer complete paged/downloadable scope
  or reject that request explicitly. Do not approve, narrow, or silently omit it.
  Keep labels in English, using existing design-system components.

## 8. Delivery sequence and gates

### A. Reproduce and record the contract

Record supported versions, event/field classifications, known consumers and the
baseline failures. Use synthetic/sanitized fixtures shaped like the observed
OpenCode events; keep real session payloads out of the repository. Obtain the
Pi session/export if available, without blocking coverage of both aggregate and
useful-content paths. Exit: precise failing boundaries and accepted field policy.

File groups: `opencode_interactive.go`, `opencode_history_projection.go`,
`interactive.go`, `harness.go`, `pi_interactive.go`, `pi_subagents.go`,
`codex_interactive.go`; add only focused reader/policy helpers and small fixtures.

### B. First working correction: discard-only projection

Implement incremental framing/projection and policy-approved omission for
OpenCode summaries and Pi redundant aggregates. Leave the prior useful-content
path explicit until C/D. Add bounded diagnostics and flow control at the reader.
Exit: 25 MiB/100 MiB unused data is drained and the next normal event is processed;
Stop remains responsive. This is a partial mitigation, not the final solution.

### C. Complete useful-content vertical slice

Implement staged content, atomic SQL publication, bounded readers and client
details/download. Wire one representative OpenCode tool result and assistant text
through the entire path. Exit: a 25 MiB useful result remains complete after
restart and can be read/exported without a full-size allocation in engine/client.

File groups: `stream.go`, `sqlite_session_repository.go`, `sqlite_event_stream.go`,
`sqlite_content.go`, `sqlite_history.go`, SQLite schema/migration ownership,
`sqlite_linked_execution.go`, `sqlite_general_history.go`, graph final readers;
frontend `engine.ts`, `EventContent.tsx`, `ConversationEvents.tsx`, `SessionEvent.tsx`.
Inspect current edits before modifying these files: unrelated harness-installation
work is already present in several of them.

### D. Complete all adapter and control paths

Migrate Pi and Codex useful events, native children, final snapshots and all native
history responses. Complete human-decision detail handling and internal consumers.
Remove obsolete generic 2/4 MiB fatal branches only for paths now covered by the
new reader; retain semantic/resource validation with attributable errors.
Exit: no fallback rehydrates a whole large payload in these execution paths.

### E. Validate and roll out

Run focused checks per changed boundary, followed by explicit human validation
on real OpenCode, Pi and Codex. Publish evidence with implemented/pending statuses.
Installer/release rollout, full suites, review agents and commits/pushes require
their normal explicit authorization; this plan does not request them implicitly.

## 9. Acceptance matrix

The following are proposed focused checks for implementation, not checks already
run. Prefer small unit/component fixtures and streaming probes. Any substantial
integration/review campaign requires explicit authorization under project policy.

| Scenario | Required outcome |
| --- | --- |
| OpenCode summary events shaped like 25 MiB / 3.66 MiB observed failures | Turn continues; summary patches are not retained; next completion/error event still processed. |
| Same fields via HTTP history and user-message updates | No resume/reconciliation failure or large cached metadata object. |
| Pi `agent_end`/`turn_end` with 25 MiB repeated messages | Drain correctly; do not duplicate history, cancel process or fabricate settlement. |
| Pi useful final message/tool result over 4 MiB | Preserve complete useful data and continue; do not treat it as disposable aggregate content. |
| Codex large file-change/tool/output and history response | Full semantics available; correlation and final graph verification intact. |
| 1 MiB, 25 MiB, 100 MiB, 256 MiB useful payloads | All fit the documented resource budget; canonical content digest/length matches after restart and export. |
| Incremental deltas followed by large authoritative final snapshot | Same event identity, exact content, no duplicate output or accidental loss of valid deltas. |
| Large field before type/ID; escaped newline; UTF-8 split across chunks; multiline SSE | Same result as small equivalent event; no parser desynchronization. |
| Malformed/truncated JSON or duplicate identity/policy keys | Explicit owning-operation failure; no partial publish, mistaken permission decision or guessed recovery. |
| Cancellation during staging; crash before/after publication | No dangling visible event/reference; reclaim only unpublished/unreferenced data; Stop remains responsive. |
| Storage budget/disk failure; concurrent other session | Clear attributable failure; capacity limits do not set global storage failure; real SQL failures retain integrity protection. |
| Large permission/question/elicitation | Exact requested scope/options retained, accessible and correlated; only explicit human response resolves it. |
| Bounded history, linked/general reads, search, graph settlement, export | No whole-event reconstruction merely to extract small fields; no preview mistaken for full content. |
| Open details/download on desktop, browser and mobile WebView | Correct ownership/auth; bounded rendering; no all-pages-in-Blob download for large payloads. |

### Memory/performance evidence

Measure peak live heap (not only allocations per call), queue bytes, staging bytes,
writer-lock duration and client heap with generated streams, not prebuilt giant
fixture strings. A 10x payload increase must not cause an approximately 10x live
heap increase. Initial acceptance target: additional parser/adapter live heap at
most 16 MiB per large active stream, excluding fixed process/SQLite caches, with
an explicit aggregate concurrency budget. Record the baseline and measurement
method; if the target changes, explain why rather than quietly weakening it.

No SQL transaction or `app.mu` span may include waiting for input. Another chat
must remain able to persist small events during large ingestion. Memory stability
alone is insufficient if staging or event caches grow indefinitely across turns.

## 10. Definition of done

- [ ] All three harnesses handle known large legitimate payloads end to end.
- [ ] Optional omissions are path-specific, documented and do not affect execution.
- [ ] Useful content is complete and durable; control/human-decision semantics are preserved.
- [ ] Live streams, final snapshots and resume/history use consistent policies.
- [ ] Engine, internal readers, UI and downloads remain bounded in memory.
- [ ] Partial writes, Stop, restart and capacity failure have explicit outcomes.
- [ ] Diagnostics identify harness/session/phase/type/byte counts without payload leakage.
- [ ] Focused evidence and remaining human-validation gaps are recorded separately.
- [ ] No production data rewrite, legacy fallback or implicit authorization/replay was introduced.
