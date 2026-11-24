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
| `POST /captures` | Retain an immutable participant preparation using its original operation receipt; no session or inference. See prepared work below. |
| `POST /conversations`; `POST /work` | Durable coordinator creation without inference, or asynchronous saved-Agent work admission. Return an immutable operation receipt with `result` containing the admitted execution. |
| `GET /conversations/{id}`; `GET /work/{id}` | Execution/capture, recent input ledger, exact request records, `attention_only` and `owned_jobs_running`. |
| `PATCH /conversations/{id}/configuration` | `{operation_id,expected_revision,capture}`; save desired capture, then apply a new generation only after drainage. Worker captures cannot be edited. |
| `POST /conversations/{id}/inputs`; `POST /work/{id}/inputs` | Attributed input `{operation_id,source_kind,source_id,mode,text,mentions?}`. Worker continuation is application input, not an interactive worker composer. |
| `GET /conversations/{id}/inputs?source_kind=…&source_id=…`; `GET /work/{id}/inputs?source_kind=…&source_id=…` | Exact durable source inclusion lookup, independent of recent-window size. |
| `DELETE /conversations/{id}/inputs/{inputId}` | Remove only an unclaimed queued human input. |
| `POST /conversations/{id}/stop`; `POST /work/{id}/cancel` | Exact observed turn or work cancellation. Receipt means requested; ordered execution events supply cessation or uncertainty. |
| `POST /groups/{instanceKey}/cancel` | `{operation_id,project_id}`; durably seal this integration/project instance before asynchronously draining admitted owners. Standard immutable operation receipt. |
| `GET /groups/{instanceKey}?project_id=...` | Current `{instance_key,project_id,cancel_requested,executions:[execution...]}`; includes all retained participant workers, not just workers known to the application before a lost response. |
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

### Prepared participants and initial CWD

Project Git identities are owned by KLM Project settings. Preparation, direct
runner admission and subsequent jobs retain the registered `project_id`'s
identity capture, independent of initial CWD. An optional output-only
`project_git` record contains the captured project, default and versioned alias
references; an empty identity list explicitly captures shared mode. Integrators
neither select a project by changing the workspace nor supply credential data.
Managed `shell_start` commands use the captured default or the operation-local
`klm-git-identity <alias> gh|git …` launcher. Job continuations and cold resume
retain the original capture. Replacement/removal invalidates the referenced old
credential for new launches; active executions cannot silently rebind. Legacy
executions without a capture can capture at fresh turn admission, but automatic
cold recovery without a capture is blocked. See [Project Git identities](README.md#project-git-identities).

The positive `extensions` capability flags are
`immutable_participant_preparation`, `initial_cwd_workspace`,
`worker_input_lookup`, `resume_drained`, `instance_cancellation_fence` and
`cold_process_reconciliation`. `linked_worktree_binding` is explicitly `false`.
Applications using relaxed directory semantics must require
`initial_cwd_workspace:true` alongside the other extensions they need; an older
runner requiring linked worktrees needs a coordinated upgrade. These flags
describe implemented mechanisms, not runtime acceptance.

`POST /captures` accepts `operation_id`, `project_id`, `instance_key`,
`workspace:{path,base}`, `capture` and optional `recovery_policy:"resume_drained"`.
`workspace`, its `path` and its legacy `base` may be omitted; an omitted/empty
path defaults to the registered project directory. A supplied path must be an
existing absolute directory on the engine host.
An embedded participant supplies its explicit capture. A saved participant also
supplies exact `agent`, `source_uid` and `agent_revision`; KLM resolves its saved
instructions, settings and independent Library snapshot with source adoption
disabled. Caller-supplied capability snapshots are replaced by engine resolution.
Configuration gates are incompatible with these immutable preparations.

The returned standard receipt has a `result` containing the normalized capture,
saved source when applicable, instance key and normalized initial CWD with any
observed workspace provenance. It is
retained in the existing operation ledger and retrieved through the existing
operation endpoint. Retrying its original operation/body returns the original
preparation, even after source changes. No capture-handle table or scope service
is introduced. The application must retain the complete roster before starting it;
this endpoint alone does not implement atomic roster admission. Whole-instance
cancellation uses the group fence below.

Initial CWD is launch metadata, not a repository or filesystem access boundary.
Shared and overlapping paths are legal across preparations, instances and
integrations. Non-Git directories, the registered checkout, other/nested
repositories and arbitrary existing worktrees are legal. Directory scope comes
from the captured task instructions and normal tool policy; commands may change
directory or use another repository. Runner lifecycle code never creates,
resets, cleans or removes worktrees and does not inspect shell commands to enforce
directory scope.

Preparation collects read-only Git and directory observations on a best-effort
basis. `base` remains a caller-selected reference; when resolvable in the initial
CWD, `base_commit` records its observed commit separately from `initial_head`.
`project_root` observes the registered project's Git root;
`common_git_dir` and `worktree_git_dir` observe the initial CWD's repository.
Missing Git, unresolved bases and unavailable observations do not reject work.
These fields and directory identities are provenance only: admission, inputs,
turn scheduling and cold continuation check the actual launch directory and the
existing/nonremoved project policy, without comparing repository or directory
identities across turns. Unavailable launch directories fail explicitly; actual
native/process launch errors remain execution diagnostics.

Wire and receipt compatibility is unchanged. All legacy workspace fields and JSON
tags remain, including their non-omitted empty-string encoding: `path`, `base`,
`base_commit`, `initial_head`, `project_root`, `common_git_dir`,
`worktree_git_dir`, `directory_identity`, `common_git_identity` and
`worktree_git_identity`. Preparation/create request structs, field order and
marshalled digest encoding are unchanged; no new request fields were added.
Original receipts, captures, observed facts and native IDs remain immutable and
readable, including earlier linked-worktree preparations. An original receipt
retry returns its original result, not refreshed directory observations.

`POST /work` may supply `{operation_id,key,project_id,preparation_id,prompt}`.
`preparation_id` is the owning preparation's original operation ID, resolved only
within the authenticated integration/project. Both embedded leaders and saved
members use this path. Capture/source overrides are rejected. Admission uses the
retained engine-owned data without rereading mutable Agent/Library definitions.
The captured path is stored as `Session.ExecutionCWD` and propagated to native
sessions, model validation, owned shell jobs and mention preparation. Inputs,
scheduling and cold continuation use the stored session CWD. Integration/project/
instance binding, exact source resolution with `PreserveSource:true`, native
identity and all capture/manifest, receipt, cancellation and drainage safeguards
remain in force. Direct saved-Agent/coordinator admission defaults to the
registered project directory. Personal sessions, graph workspace contracts and
personal webhook/schedule policies retain their existing behavior.

`GET /work/{id}` and conversation reads additionally return
`native_context:{id}` for the current generation's retained native identity.
It is an identity observation, not proof of process liveness or settlement.

### Whole-instance cancellation fence

Groups are integration-scoped stable `instance_key` values, bound to one registered
project. Captures bind the key before any execution exists; reusing that key for a
different project conflicts. Other integrations cannot read or cancel its retained
participants. An unused key reads as an empty uncancelled group for the requested
project and may be cancelled before any preparation exists.

Cancellation stores an immutable internal seal in the existing operation ledger;
no group/scope service or new schema table is introduced. In the same transaction,
existing workers acquire `cancel_requested`, new claims are sealed and unclaimed
inputs are paused. Only then are their normal native/job cancellation paths
invoked. Preparation and work admission recheck the seal transactionally; input,
effect, shell admission, native permission reply, turn claim and recovery also
respect it. Original immutable receipts remain readable after cancellation, but
reading/retrying an old receipt does not authorize another execution.

The cancel operation's receipt `result` is the small sealed group identity with
`executions:[]`; it is not a drainage observation. Use GET for current retained
executions. Prepared-only cancellation therefore needs no native process. The
engine retries unfinished opted-in/group cancellation after restart and while
late admitted effects settle. `cancelling` or `interrupted` remains visible when
cessation/effects cannot be confirmed; a seal or native completion alone does not
manufacture `cancelled`. There is no unseal or cancelled-worker resume operation.

### Opt-in cold continuation

Only prepared work explicitly captured with `resume_drained` uses this policy.
Existing personal sessions and older external executions retain their recovery
behavior. Complete-turn claims clear `processes_drained`; native runtime shutdown
records confirmed drainage separately from model completion. On startup, an
interrupted opted-in worker with reconciled drainage, its original native identity,
available initial CWD and settled jobs/application effects can become runnable in the
same execution/session/generation. One attributed `application_event` input keyed
by `recovery/<original turn ID>` contains recovery facts and unknown-inclusion
references. Original inputs, shell commands, native callbacks and MCP effects are
not replayed. Frozen application role Markdown supplies recovery behavior.

Opted-in native transports and shell jobs retain process-domain evidence before
execution. On Windows this is an unpredictable named kill-on-close Job plus the
original engine's process creation identity, with durable containment confirmation
before resuming the suspended target. Recovery first rejects a still-live original
engine, then inspects/drains that exact kernel Job; a root PID alone never supplies
kill authority or cessation proof. An unconfirmed pre-attachment start remains an
explicit admission uncertainty. Personal/legacy runtimes retain their unnamed Jobs.

Linux uses its existing subreaper/pidfd supervisor. The supervisor receives a
dedicated private receipt descriptor before target admission and fsyncs confirmed
domain drainage after engine-channel EOF or Close. Cold recovery requires that
receipt or a changed OS boot identity; disappearance of a supervisor/root PID is
not sufficient proof. Missing legacy ownership evidence remains explicit.

A reconciled interrupted shell job retains its original command, output bytes,
digest and workspace. Its state becomes `interrupted`, with an unknown exit
outcome and no manufactured exit code. It is physically settled, not classified as
successful. Recovery facts include the original job/result and unknown-inclusion
records so the captured role can inspect artifacts and continue unresolved work.

Application effects already admitted must settle through their original reply.
The two-second reconciliation loop automatically continues after that durable
reply arrives; recovery facts include the actual retained reply and explicitly do
not assert old native callback consumption. Unadmitted lost callbacks become
`callback_lost`, retaining their original identity/arguments rather than executing
the application effect. MCP effects additionally retain engine-owned
`mcp_pending`/`mcp_replied`/`mcp_not_sent`/`mcp_unknown` records in the existing call
ledger. A lost arbitrary tool-call result is a specific `mcp_result_unknown`
blocker; resource discovery/read queries are distinguished from effectful calls.
External applications cannot manufacture MCP settlement through tool-reply APIs.

Lost native request/answer provenance remains retained. Recovery supplies request
references without replaying protected permission/elicitation answers. If the
harness establishes an actual new callback, it gets a fresh runner request ID even
when the native request ID is reused; its real offered schema and current callback
still require an explicit human reply. No old answer is rebound automatically.

Specific `recovery_blocked` codes include `native_process_domain_missing`,
`native_process_domain_unconfirmed`, `owned_job_domain_missing`,
`owned_job_domain_unconfirmed`, `application_call_unsettled`,
`original_native_context_missing`, `workspace_unavailable`, `mcp_result_unknown`
and `recovery_context_exceeds_input_limit`. Typed native network timeouts receive
bounded-backoff automatic retries; other native failures retain their concrete
diagnostic rather than silently replacing the context. No generic manual Resume
gate is added. Cancelled executions never enter this continuation path.

Recorded cold provider proof under the former linked-worktree contract is
Windows/Codex 0.153.4: abrupt death of the owned
engine, named-domain reconciliation, original thread/session/capture/worktree and
included recovery input, with retained results and no repeated file edit. A focused
real-process Windows job check also retains pre-crash output without replay or a
fabricated exit code. Linux receipt-based recovery, other harness cold-resume
behavior, native request re-establishment and selected real-MCP variants remain
separate runtime acceptance paths; mechanism flags are not those acceptance claims.
That earlier evidence does not validate the relaxed initial-CWD behavior.

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

November 18 initial-CWD compatibility: the Windows dev build passed with
`go build -tags dev -o C:\Users\Samuel\AppData\Local\Temp\opencode\klm-initial-cwd-dev.exe .`
from `engine/`, after verifying the output parent. No runner was launched or
restarted. Shared/non-Git/other-or-nested-repository paths, old receipt retries,
unavailable launch directories and same-context continuation under relaxed
semantics remain runtime validation work. No platform-specific source changed,
so no Linux cross-build was run for this slice.

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
