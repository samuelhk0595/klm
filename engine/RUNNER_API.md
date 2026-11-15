# External runner protocol 1

Read when implementing an authenticated local application integration. The
Factory implementation is commissioned by the current user; this technical
contract does not establish human acceptance of the Factory specifications or UI.

## Ownership and storage

SQLite publication schema 6 adds the `runner_*` admission, input, operation,
application-call, request, generation, event and job ledgers. Ordinary sessions
keep their existing metadata and behavior. Runner technical sessions have
immutable `origin=external_runner`, integration/execution identity and
`runner_coordinator` or `runner_worker` role. Public ordinary session routes,
subscriptions and projections exclude them. External adapters select the
application manifest positively and do not receive ordinary linked/general-agent,
graph or delegation tools. Both external roles capture YOLO without editing
personal defaults or saved Agent definitions.

The application owns business identity and objective assessment. KLM owns native
contexts, inclusion evidence, pending native requests, process domains and shell
jobs. Native completion is not objective success. External native transports are
retired after a turn; separate owned job domains can continue and wake the same
execution. Unknown starts/inclusion/effects are retained and never replayed to
manufacture success.

## Connection and HTTP

The existing per-user named pipe / Unix socket accepts read-only `discover` and
returns `{engine_id,api_port,runner_protocol,ready}`. It issues no code. Explicit
Connect uses existing `pair` and `/api/auth/pair`, then binds a stable integration
key through `POST /api/runner/connections`. Existing authentication remains intact.

Runner requests send the paired bearer credential, `X-KLM-Native: 1`,
`X-KLM-Read-Contract: 3`, `X-KLM-Engine-ID` and, after connection,
`X-KLM-Runner-ID`. The browser must not receive the credential or pairing code.
Mutations use bounded `operation_id` values. Reused operation identities with
different normalized targets/bodies conflict. A conversation key and Assignment
key independently prevent another execution when operation IDs change.

| Route under `/api/runner` | Contract |
| --- | --- |
| `GET /capabilities` | Protocol/engine identity, per-harness mechanisms and readiness diagnostics; probe evidence is distinct from complete runtime acceptance. |
| `POST /connections`; `GET /connections/{integrationId}` | Bind/reconnect stable key/identity; confirm current client binding. |
| `POST /captures/validate` | Validate `{project_id,capture}` without admitting a conversation or model turn. |
| `POST /conversations`; `POST /work` | Durable coordinator creation without inference, or asynchronous saved-Agent work admission. Return an immutable operation receipt with `result` containing the admitted execution. |
| `GET /conversations/{id}`; `GET /work/{id}` | Execution/capture, recent input ledger, exact request records, `attention_only` and `owned_jobs_running`. |
| `PATCH /conversations/{id}/configuration` | `{operation_id,expected_revision,capture}`; save desired capture, then apply a new generation only after drainage. Worker captures cannot be edited. |
| `POST /conversations/{id}/inputs`; `POST /work/{id}/inputs` | Attributed input `{operation_id,source_kind,source_id,mode,text,mentions?}`. Worker continuation is application input, not an interactive worker composer. |
| `GET /conversations/{id}/inputs?source_kind=…&source_id=…` | Exact durable source inclusion lookup, independent of recent-window size. |
| `DELETE /conversations/{id}/inputs/{inputId}` | Remove only an unclaimed queued human input. |
| `POST /conversations/{id}/stop`; `POST /work/{id}/cancel` | Exact observed turn or work cancellation. Receipt means requested; ordered execution events supply cessation or uncertainty. |
| `POST /tool-calls/{callId}/admit` | Reserve an application effect before its transaction. Stop seals new admissions and waits for admitted effects. |
| `POST /tool-calls/{callId}/reply` | `{operation_id,result,is_error?}`; immutable application result, distinct from native inclusion. |
| `POST /requests/{requestId}/reply` | `{operation_id,answer}`; revalidate owning live callback and offered native schema/decisions before recording/delivering. |
| `GET /events?after=…`; `GET /updates` | Ordered retained events, exclusive decimal cursor, bounded count/bytes; SSE only wakes catch-up. |
| `GET /operations/{operationId}` | Original immutable receipt, not current execution status. |
| `GET /conversations/{id}/history`; `GET /work/{id}/history` | Existing bounded native history pages, newest tail by default; exclusive `before`/`after` ordinals. |
| `GET /conversations/{id}/content`; `GET /work/{id}/content` | Existing digest-bound native event chunks using `ordinal`, `digest`, `offset`. Optional `source_session_id` addresses a retained owned generation. |
| `GET /jobs/{jobId}/content` | Finalized retained command bytes, digest, byte offsets, base64 chunk and EOF. EOF is not execution success. |

Captures contain explicit harness/model/effort, assembled instructions, selection,
JSON-schema manifest and effective `yolo=true`. Saved-Agent work additionally
supplies exact `agent`, `source_uid`, `agent_revision` and task. Source adoption is
disabled for runner resolution: existing TOML bytes are preserved. Missing or
unavailable selections fail explicitly rather than being stripped.

## Delivery and lifecycle events

`steer` enters the next eligible native step; `queue` waits for turn completion.
`passive` can join an active turn but alone does not start an idle turn. Source
kind/ID remain explicit even when native transport uses user-role messages.
Claim, native acceptance and observed inclusion remain separate; an ACK is not
inclusion. Mention snapshots use the existing contained, bounded preparation.

Coordinators with `configuration_gate=true` emit `capture_request` before a new
complete turn. This is a fixed lifecycle request, not a model tool. The owning
application replies through the same call-reply route with
`{capture,context?,error?}` after re-reading its definition. KLM validates the
capture and installs it at a drained generation boundary. This closes the queue /
external-file-edit boundary without making KLM interpret Factory files.

Events include `execution`, `input`, `input_included`, `history_changed`, `job`,
`tool_call`, `effect_admitted`, `tool_reply`, `request`, `request_reply` and
`capture_request`. `history_changed` includes the actual `source_session_id` and
affected ordinals; readers retrieve real native messages/work rather than reports.

Owned tools are `shell_start`, `job_read`, `job_wait`, `job_cancel`. Command intent
precedes spawn. The same process continues after foreground release. Waits reserve
framing time within the 5,000 ms ceiling. Full output is retained separately from
its bounded preview. Stop/cancel drains the owned process domain; remote MCP
effects are not assumed undone. Selected MCP calls keep their existing synchronous
timeout/elicitation behavior and are not converted into jobs.

Human answer payloads are protected using the engine credential storage. Linux
therefore preserves the same database/key backup and explicit missing-key rules.
Question answers, permission decisions and elicitation actions are separate; none
becomes new human graph/session authorization. External grants cannot enter
personal project/global rule scopes.

## Recorded checks and acceptance boundary

Targeted SQLite/HTTP checks cover origin immutability, operation-body conflicts,
unique Assignment keys, ordinary exact-ID/inventory/subscription exclusion and
retained native history. Real isolated pinned Pi 0.85.1, OpenCode 1.18.35 and Codex
0.153.4 runner checks reached operational calls, exact unanswered native questions,
original worker receipt recovery, Foreman-only Stop and independent work cancel.
Original auth and saved-Agent bytes remained unchanged; owned jobs/processes were
drained. No human answer was fabricated.

Native integrated checks use `KLM_RUN_FACTORY_NATIVE_PROBE=1` and
`KLM_FACTORY_PROBE_HARNESS=pi|opencode|codex` with the focused
`TestRunnerNativeCodexOwnedBoundary` test (historical test name). These are
explicit real-provider opt-ins, not part of normal validation.

The representative authenticated local-discovery/pairing and Pi browser journey
now includes actual human answer delivery/resumption, retained output and separate
objective assessment; see `../FACTORY_BROWSER_VALIDATION.md` and
`../FACTORY_IMPLEMENTATION_COMPLETION.md` for source cohorts and driver limits.
Installed-engine rollout, selected upstream MCP variants, restart permutations,
native compaction and Linux runtime acceptance remain human checks. Full suites,
installers and deployment were not performed for this delivery.
