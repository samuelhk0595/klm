---
reviewed: false
last_revision: "2026-11-14T16:22:09-03:00"
---

# First Werbert Factory delivery and KLM runner adaptations

**Status: implementation plan for human review. No implementation is authorized by this document.**

Read before implementing the first local KLM-backed Werbert journey. Deliver one complete path:

**Discover local KLM → Connect → select an existing project → configure Foreman → Chat → Ticket → one saved-Agent Assignment → asynchronous worker → human intervention in Chat → retained result → separate objective assessment.**

This planning/revision work changes documentation only, including alignment of the companion specs. The user and originating conversation must validate the complete plan before implementation. Mock implementation, earlier Werbert deployments, session creation and static source inspection establish neither approval nor runner compatibility.

**Approved review adjustments:** in human message
`cae9b2ab-3c7f-4272-8a34-1c6677bc2ca1`, the user accepted moving the short harness
compatibility proof ahead of implementation and deferring background conversion
of long MCP calls. The same message requires **all Factory workers always in YOLO
mode**, as Foreman already is. These three decisions supersede the earlier worker
non-YOLO proposal and MCP-job envelope; they are not approval to implement the
whole plan.

## 1. Scope, sources and decisions

Roots used below:

- **K:** `C:\Users\Samuel\Documents\Projects\Personal\klm`
- **W:** `C:\Users\Samuel\Documents\Projects\Hermes\werbert`

Authoritative inputs are [Persistent Factory specification](FACTORY_PERSISTENT_AGENT_SPEC.md) (**F**), [runner proposal](FACTORY_RUNNER_CONTRACT.md) (**R**), and the user's current small-delivery brief. The maintained [mock experience](../../Hermes/werbert/frontend/docs/foreman-mock-experience.md) and [FX ledger](../../Hermes/werbert/frontend/docs/foreman-experience-deltas.md) distinguish requested changes from proposals. This plan reconciles them; it does not mark any mock proposal accepted.

Inspected context includes both `AGENTS.md` files, K `product.md`, W `docs/INDEX.md`, documentation guide, `PRODUCT.md`, `ARCHITECTURE.md`, `docs/CURRENT_STATE.md`, Factory/Ticket/direct-Agent specifications and architecture decision index. The `writing-for-agents` skill was read. W's incumbent memory, graph, automation, managed-clone and multi-assignment contracts describe its independent runner, not requirements to transplant into this slice. Preserve that product's records and code. This requested plan belongs at K root rather than W's temporary-document directory.

### Required first delivery

- One KLM engine on the same computer, under the same OS user as Werbert's backend; release API `127.0.0.1:7331`. Explicit development configuration uses `17331`, not a second production connection slot.
- Existing KLM projects become Factories in their **same current checkout**. Add Factory remains accessible. No inference or execution on discovery, project selection, Settings inspection or history reads.
- Versionable explicit harness/model/effort, **Domain-specific instructions**, skill IDs and MCP connection/tool allowlists. KLM owns installation, catalog readiness and secrets.
- Persistent Foreman Chat; fixed YOLO for Foreman **and every Factory worker**; composer shows the **applied** capture read-only. Settings owns edits; neither role offers a YOLO toggle.
- Multiple independent Tickets, each with at most one saved-Agent Assignment. Delegation returns promptly; the worker does not share Foreman's turn lock or cancellation context.
- Five-second foreground shell handoff and job waits, serial next-step/next-turn/passive admission, durable attributed results and event-driven wakeups. No blocking join or model sleep/status-poll loop.
- Exact native questions, permissions and MCP elicitations remain answerable in Foreman Chat even if Foreman fails or is stopped. Native outcome and objective assessment remain separate.
- Stop Foreman targets its current turn only. Confirmed Cancel Ticket targets that Ticket's work. Read-only Open chat shows real worker messages/work in event-updated tabs.
- Factory executions are absent from **all ordinary KLM projections**, including exact-ID reads and activity/request aggregates. Personal prompts, tools, settings, permissions, queue/steering/Stop and compaction retain their existing semantics whether Factory is connected or not.
- Native compaction, durable history and bounded evidence reads. No mandatory AI Memory.

### Proposed implementation choices, awaiting review

Use one additive runner module in KLM and one thin local connector/business module in Werbert, inside their existing Go processes and SQLite stores. Keep ordinary KLM execution as the default; only a server-owned external capture selects runner behavior. Reuse adapters, Library snapshots, process ownership and history storage rather than building another harness stack.

Use a distinct KLM-backed Werbert admission namespace and `/foreman` entry for this first delivery. Preserve the incumbent routes, data and independent-runner behavior. The compact Chat/Tickets/Settings navigation and Library modal's **Use selection → Save settings** staging remain proposals, consistent with the mock, rather than accepted replacements for the incumbent product.

**Approved worker execution policy:** capture `yolo=true` for every external
Factory worker, independently of the saved Agent definition. Saved KLM Agent
configurations currently have no YOLO field (`K/engine/agent_types.go`); do not add
one or change personal defaults to implement this fixed external-role rule. Show
the effective mode read-only in the Assignment/worker capture. No non-YOLO default
or worker-policy editor remains in this first delivery. YOLO bypasses eligible
execution approvals, not task scope, tool allowlists, necessary task questions or
MCP elicitation. The permission-reply path remains for an actual upstream request
that still needs explicit human action; do not manufacture or auto-answer one.

## 2. Existing seams and essential gaps

| Inspected implementation | Reuse / necessary adaptation |
| --- | --- |
| K `client_auth.go`, `cli_windows.go`, `cli_linux.go`, `sqlite_api.go` | Paired credentials, protected per-user control, engine identity and project inventory exist. Even `/api/health` requires authentication. Add bounded read-only local discovery metadata; retain existing `pair` redemption. |
| K `agent_catalog_api.go`, `agent_resolution.go`, `agent_turn.go`, `capabilities.go`, `library_api.go` | Reuse saved-Agent source/revision resolution and selected capability capture. Foreman requires an explicit model before calling a resolver that otherwise fills a default. Use current complete catalogs; never silently drop unavailable selections. |
| K `store.go`, `sqlite_session_repository.go`, `harness.go`, `runtime.go` | Reuse technical session/history/native runtime records. Introduce immutable external origin and runner role, plus runner admission; adding an arbitrary new `Role` alone is insufficient. |
| K `graph_adapter.go:bridgeTools`, `harness.go`, `linked_bridge.go` | Existing negative role checks can give an unfamiliar role ordinary linked tools/graph hooks. Add a positive external branch before those defaults, including server-side dispatch rejection of ordinary role tools. |
| K `message_queue.go`, `sqlite_execution_queue.go`, `pi-permissions.ts`, `opencode-permissions-plugin.mjs`, `codex_interactive.go` | Native steering and some acknowledgement/inclusion distinctions already exist. Add externally attributed immutable input batches and passive/queue rules without changing personal queue paths. |
| K `platform_windows.go`, `platform_linux.go`, `interactive.go` | Reuse owned process launch/drainage. There is no demonstrated five-second shell-to-job API in the inspected paths. A process shutdown timeout is not a shell job handoff. |
| K `sqlite_activity.go`, `sqlite_api.go`, `sqlite_updates.go`, `api.go:chatSessionHandler` | Some ordinary lists filter `role=''`; other exact-ID paths only exclude graph nodes or restrict subagent mutations. Enforce external visibility throughout reads, publication and controls, before pagination/counting. |
| W `server.go:New`, `store.go`, `factory_inbox.go`, `factory_dispatcher_boundary.go` | SQLite transaction/receipt/inbox patterns are useful. Startup currently constructs and starts native Factory, Ticket and automation services. New KLM-backed rows must never enter those services. |
| W `ticket_assignment_store.go`, `ticket_assignment_*`, `factory_ticket_tools.go` | Reuse business validation patterns, not its native assignment executor, dependency/capacity/retry framework. Current routine-progress outbox can wake coordination; the new path must retain routine reports without that wake. |
| W `frontend/src/features/foreman-mock/` | Actual components are wired to a fixture reducer, fixed catalogs, local arrays and synthetic events. `ExecutorChat.tsx` currently reconstructs reports, not a full native transcript. Replace those data seams, preserving the requested controls and source-shaped Chat. |

### Harness compatibility work is on the critical path

KLM currently pins managed Pi **0.85.1**, OpenCode **1.18.35**, Codex **0.153.4** in `harness_discovery.go`. W's later Pi/Codex deployment receipts do not validate these adapters. Keep KLM's pins for this work unless a concrete blocker requires a separately reviewed change. Do not upgrade shared harness defaults as a shortcut.

| Harness | Existing building blocks | First-delivery gap / focused proof required |
| --- | --- | --- |
| Pi | Owned RPC extension, registered tools, selected Library, `turn_end`/`before_agent_start` steering, native session-file checks | External-role manifest; exact input inclusion attribution; replace/suppress native shell paths with owned job tools; retain background jobs across yielded turns; verify abort and native compaction. |
| OpenCode | Controlled profile, MCP bridge, pre-tool gate, `experimental.chat.messages.transform`, `noReply` persistence before model snapshot inclusion | External role/manifest and job-tool enforcement. Current YOLO plugin enables configured tools broadly: external deny/allowlists must remain authoritative after that code. Disable native `task` and alternative shell bypasses for external roles. Verify that boundary claim is not later than the eligible model snapshot. |
| Codex | App-server, per-thread MCP projection, `turn/steer`, client input IDs, request handling, dynamic-tool infrastructure for owned delegation | External manifest capture/start/resume; distinguish steer ACK from actual input inclusion. Prove pinned Codex can disable/replace built-in shell/unified-exec paths while exposing owned jobs; native exec timeouts alone are insufficient. Confirm exact next-step behavior and unknown-delivery handling with these versions. |

Capabilities must report each mechanism independently, with an actionable diagnostic. Merely installing a harness does not mean external execution is supported. Complete support is intended for **all three**. If a pinned harness cannot enforce shell replacement, safe next-step delivery or the external tool surface, report that specific incompatibility before proposing reduced coverage. Do not quietly ship a prompt-only or next-turn-only substitute.

## 3. Ordered implementation work

Each numbered item is a delivery step, not an independent framework or agent assignment. Owners describe code responsibility. Proposed new filenames are marked **new**; verify the working tree again before editing because both repositories contain extensive unrelated changes.

### 0 — Prove the pinned harness mechanisms first

**Owner: KLM engine. Run first, after implementation authorization.**

Before the persistent runner/API or Werbert business loop, run a short isolated
probe for each pinned Pi/OpenCode/Codex target. Reuse the current adapter/plugin
seams with only the minimal disposable tool/profile support needed to establish:
external operational tools, fixed YOLO on coordinator and worker, five-second
shell handoff with one continuing process, bounded job waits, event-driven job
completion, next-step input inclusion and Stop/drainage. Include a necessary task
question or MCP elicitation, neither answered by YOLO. The ordinary personal path
must keep its existing settings and tool behavior.

Use temporary directories and harmless commands; no application-wide framework,
provider/harness upgrade or production data is needed. These probes establish
native feasibility, not completed runner durability. Record each concrete support
gap, especially Codex shell replacement and actual next-step inclusion. If a
required mechanism cannot be supported, report it before implementing steps 1–7
or narrowing coverage. Do not convert MCP calls into jobs in these probes.

**Done when:** actual pinned-harness evidence establishes the required mechanism
or identifies a specific incompatibility for a user decision. A matching prompt,
protocol signature or fixture without the native harness is insufficient.

### 1 — Establish the external execution boundary

**Owner: KLM engine.**

Files: `engine/store.go`, `harness.go` (including `adapter`), `graph_adapter.go`, `linked_bridge.go`, `sqlite_schema.go`, `sqlite_indexes.go`, `sqlite_store.go`; **new** `runner_records.go`, `sqlite_runner.go`, `runner_runtime.go`.

1. Add immutable server-owned `origin=external_runner` and `integrationId` to technical execution metadata, and explicit coordinator/worker roles. Legacy/missing origin remains personal. Clients cannot set these through ordinary session creation/patching.
2. Keep runner coordination/work IDs separate from hidden technical session/native generation IDs. Store opaque Werbert references without making KLM understand Tickets.
3. Reuse existing execution resources through a small external capture/hook on the current adapter. It selects the captured application manifest, runner input source and event sink. Ordinary admissions execute the existing path unchanged. External workers have independent runtime/process ownership, not an owned blocking subagent batch.
4. Select the external tool catalog positively. Exclude linked/general-agent tools, topic assignment, session creation, graph/automation tools, native delegation and `subagent_run`; selected Library remains separate. Validate the same restrictions at tool dispatch, not only discovery.
5. Persist external admissions and operation receipts before launching anything. Unique durable conversation key and Assignment key prevent another worker even if a caller changes its operation ID after a lost response.

**Done when:** external technical records can exist without starting inference, own the right project directory and cannot acquire ordinary role capabilities. Old records/defaults remain unchanged.

### 2 — Close ordinary KLM projection and permission leakage

**Owner: KLM engine; no KLM visual redesign.**

Files: `engine/api.go`, `sqlite_api.go`, `sqlite_activity.go`, `sqlite_topic_views.go`, `sqlite_updates.go`, `sqlite_change_bus.go`, `sqlite_history.go`, `content_downloads.go`, `sqlite_general_history.go`, `general_agent_tools.go`, `sqlite_execution_analytics.go`, `analytics*.go`, `permissions.go`, `sqlite_execution_requests.go`, `session_controls.go`, `live_yolo.go` and affected Library/Agent reference projections.

- Centralize an ordinary-visible predicate using origin, and apply it to list/search/topic counts/order, archived lists, activity/Focus/Happening now, pending counts/cards, exact-ID metadata/history/events/content/export/downloads, general-agent discovery/read/control and ordinary subscriptions. Guard routes that bypass `chatSessionHandler` as well as that wrapper. Retained hidden IDs must not work as personal deep links or control targets.
- Route external change notices exclusively to runner subscribers. Global/project invalidations must not publish hidden IDs, titles, pending payloads or hidden activity. Exclude external runs from ordinary analytics and visible running-session counters while retaining their technical accounting privately.
- Audit Library/Agent usage/deletion/reference diagnostics: preserve actual in-use protection without listing external execution identities in ordinary UI. No Factory notifications may be redirected to the general agent on failure/restart.
- Keep external request/decision ownership separate from personal grant lookup/write. With fixed YOLO on Foreman and workers, do not introduce a configurable worker-grant product or persist Factory-wide decisions into personal project/global grants. If an actual upstream request still requires a human reply, retain its exact offered scope and execution owner; never relabel a broad grant as narrow. Ordinary decision menus/scopes remain unchanged.
- Foreman and worker captures always use YOLO for eligible execution approval. Manifest selection, task authority, necessary task questions and MCP elicitation remain explicit. Ordinary YOLO edits cannot reach hidden contexts, and the external fixed mode never changes personal defaults.

**Done when:** an external working/waiting execution is invisible through all ordinary entry points, including known IDs and aggregates, while the runner can still inspect/reply to its evidence. This is presentation/ownership separation, not filesystem or same-user security isolation.

### 3 — Implement the narrow runner contract and durable adapter hooks

**Owner: KLM engine.**

Files: **new** `engine/runner_api.go`, `runner_tools.go`, `runner_inputs.go`, `runner_events.go`; extend `api.go`, `message_queue.go`, `pi_interactive.go`, `pi-permissions.ts`, `opencode_interactive.go`, `opencode-permissions-plugin.mjs`, `opencode_library.go`, `codex_interactive.go`, `codex_library.go`, `runtime.go` and the owned bridge.

Implement the small surface in §4 using current authentication, SQLite transactions, content storage, selected Library snapshots and adapters. Do not add a plugin marketplace, callback registration framework or a public generic scheduler.

- One active coordinator turn; each eligible step atomically claims a bounded immutable batch. Steer enters the next eligible step; queue waits for turn completion; passive is retained and enters the next eligible context without starting a turn. Arrival after claim waits for a later boundary. Automation/native user-role encoding does not alter producer provenance.
- Track retained → claimed → native accepted → included separately. Stop/transport loss can produce unknown inclusion. Resume by original identity/evidence, not text equality or replay. Drain tools and effects before reusing/changing native context.
- Stop is not a persistent Factory pause: after confirmed cessation, a newly arriving actionable input may start a fresh serial turn. The stopped turn's claimed input is not automatically replayed. Settings edits apply only after the old capture's boundary is drained, including its relevant application-tool effects.
- Application tool calls persist owner/generation/native-call/name/arguments and emit `tool_call`. Werbert replies with the same call identity and operation receipt. Long business work is always admission plus later events; a tool call never waits for a worker's lifetime.
- Native questions/permissions/elicitations commit their pending record and runner event together. Request replies recheck the actual owning runtime/request, current state and offered choice/schema. Keep native delivery/resumption separate from HTTP acceptance.
- External history events include actual assistant messages, exposed reasoning, tool input/output/status, job references and terminal evidence. Reuse bounded history/content encoding rather than reconstructing worker chat from summary reports. No private tokens or raw upstream credentials enter events.

**Done when:** a controlled peer can submit the same work/input/tool reply twice without a second effect, deliver an event at the correct boundary, and read/reply to exact pending native requests without any ordinary-session route.

### 4 — Enforce five-second jobs and verify each harness seam

**Owner: KLM engine.**

Files: **new** `engine/runner_jobs.go` and small owned tool adapters if needed; reuse `platform_windows.go`, `platform_linux.go`, `interactive.go`, `protocol_content.go` and the harness files in step 3.

- Provide owned `shell_start`, `job_read`, `job_wait`, `job_cancel` operations under external-role manifests. Default and maximum foreground and wait duration are **5,000 ms**. Changing a model argument cannot bypass the cap.
- Persist a command intent keyed by execution + native call ID before spawn. Start once in the captured directory/environment and owned process domain. If still running at five seconds, return its job ID and bounded output while the **same process** continues. The timer releases the wait, not the process. A crash between intent and confirmed start is uncertain, never permission to spawn again.
- Jobs belong to the external execution, with originating turn recorded. Ordinary turn completion/yield does not kill a still-running job; job completion wakes the owning coordinator/worker with a stable attributed result. A worker cannot be declared settled while owned jobs/effects remain unsettled. Foreman configuration-generation changes wait for its old generation's jobs; independent workers never hold that boundary.
- Stop the observed Foreman turn and its owned activity/jobs. Cancel Ticket stops that worker and its jobs. Keep job ownership across a later turn resuming the same work; do not orphan a job because the model yielded. Confirm process/tool drainage before saying Cancelled or admitting a conflicting generation. Local process cessation does not prove a remote side effect was undone.
- Disable native alternative shell/delegation tools for the external profile or wrap every exposed shell path. Five-second prompt instructions or setting native kill-timeouts are unacceptable substitutes. Reuse step 0's native evidence rather than discovering compatibility only after the business loop is wired; recheck only mechanisms changed since that proof.
- **Approved non-shell scope:** application tools are bounded reads/admissions. Native file tools and selected MCP calls retain their existing supported execution/timeout/request behavior. The five-second foreground handoff applies to shell/job waits only. Do not add MCP-job records, pending-job result envelopes or another asynchronous MCP invocation layer in this delivery. A long MCP/model request can delay the next eligible step; do not promise universal five-second responses or kill/retry it merely to achieve that number. Keep exact elicitation and uncertain-outcome handling intact. Async MCP conversion is deferred until a concrete need justifies a later slice.

**Done when:** on each intended harness a command lasting over five seconds starts once, foreground wait releases, another input can be admitted, eventual output arrives once and Stop/cancel accounts for the owned process. Version-specific incompatibilities are reported rather than waived.

### 5 — Connect Werbert locally and initialize versionable Factories

**Owners: Werbert backend; minimal KLM local control addition.**

Files: W `engine/internal/engine/server.go`, `store.go`, `types.go`; **new** `foreman_runner_client.go`, `foreman_connection.go`, `foreman_definition.go`, `foreman_store.go`, `foreman_api.go`, platform-specific protected credential helper. K `engine/cli_windows.go`, `cli_linux.go` plus a small shared local-control identity response.

1. Add a read-only local control discovery response containing engine ID, API port, readiness and runner protocol version. It uses the existing user-protected pipe/socket; no HTTP auth exemption. Werbert also probes the fixed loopback API port. A listening port/401 alone is only a candidate, not verified KLM identity. No pair code is issued by discovery.
2. On explicit Connect, request `pair` through that local channel, redeem `/api/auth/pair` using `X-KLM-Native: 1`, verify authenticated engine identity/capabilities, then establish/reconnect the stable integration. Keep codes out of shell logging, browser state and project files. If the local control reply is unavailable/incompatible, show that fact; never fall back to unauthenticated access.
3. Store the credential and binding outside project files in protected per-user storage (Windows DPAPI; existing appropriate private local storage on another supported host). Use `X-KLM-Read-Contract: 3` and the engine identity on authenticated reads. Revocation/expiry returns to Connect; reconnect verifies the original identity and integration, not a display-name match. Preserve old records on mismatch.
4. Protect Werbert's local mutation boundary too: loopback listener, exact allowed origin/Host and CSRF checks for browser operations. A website must not cause the backend to pair or submit work through the user's local control authority. Reuse W's origin handling; no account/authentication product is needed for this local slice. Browser never receives the KLM bearer credential.
5. Read project inventory through KLM. Store a unique `(engineId, projectId)` → Factory/Chat association using KLM's project path, not W's managed-clone repository model. Do not insert fake GitHub repository IDs or clone anything.
6. Initialize only missing `.werbert/factory.yaml` using F §6.2 schema 1. Existing valid files load; invalid/legacy files remain byte-preserved and block new admission visibly. Concurrent initialization resolves the same association and never overwrites an existing file. Settings use a content digest revision, strict schema and atomic single-file replacement with stale-write detection.
7. Keep YAML authoritative for desired settings. Re-read/revalidate before every new turn; active steering keeps the current capture. Save during a turn records pending settings. The connector installs a new captured generation after confirmed drainage. A valid file write followed by a failed runner update is **saved/pending with an error**, not applied or rolled back deceptively. Restart derives the same pending update from file digest and operation receipt.
8. Freeze worker source UID/revision, effective settings and Library at admission; workers never inherit Foreman selections. Scope unavailable catalog/library errors to the affected next admission. Ordinary open/read operations and Settings save start no model work.

Give this local KLM-backed entry a small explicit startup composition that opens the store and connector without initializing the legacy environment installer, memory bootstrap, native Factory/assignment executor or automation scheduler. Keep the incumbent startup available independently; a `.werbert/factory.yaml` file never switches runtime modes. Preserve legacy records/read routes rather than migrating or executing them through the connector. This is a startup branch, not a configurable multi-runner framework.

**Done when:** Connect needs no manual pairing-code copy, project selection is idempotent, settings are versionable and real, and opening/loading the Factory launches no native execution.

### 6 — Add the small Werbert business loop

**Owner: Werbert backend.**

Files: **new** W `engine/internal/engine/foreman_runtime.go`, `foreman_tools.go`, `foreman_tickets.go`, `foreman_events.go`; reuse transaction/receipt patterns from `factory_inbox.go`, evidence formatting from `ticket_assignment_evidence.go`, and prompt loader `engine/prompts/load.go`. **New prompts:** `engine/prompts/foreman.md`, `foreman-input.md`, `factory-worker.md`.

- Add small namespaced `foreman_*` records (§5) to the existing SQLite store. Reuse validation helpers where they do not pull in graph/attempt/capacity/memory behavior. Keep the new admission path separate from W's existing `FactoryRuntime`, `TicketScheduler`, `assignment_native_*`, memory bootstrap and automation startup selectors. One registered execution has one runner, KLM.
- Persist human Chat input before forwarding it. Deduplicate source input → Ticket; operational questions create no Ticket. Foreman tools open a Ticket with objective/criteria and admit one captured saved Agent under it. A database unique constraint enforces one Assignment per Ticket, including while runner acceptance is uncertain.
- Journal Assignment + outgoing runner operation before HTTP; record runner work ID/receipt before returning the tool result. After a lost response, look up the original operation/Assignment key. A committed delegation survives Stop even if its native tool reply never arrived. No automatic replacement, retry attempts or dependency scheduling.
- Worker owns the concrete acknowledgement before substantial work, using a purpose-bound application tool; publish one source-bound acknowledgement. Foreman does not emit a second delegation announcement. A quick operational answer does not need an acknowledgement. Enforce valid current source/purpose and stable deduplication on publication.
- Purpose-bound publication is the authority for unsolicited Chat notifications. Retain raw native assistant/reasoning/tool events as inspectable work evidence, but do not blindly mirror every worker report or duplicate a published acknowledgement/result as another assistant message. Keep actual streamed work available in the source-shaped disclosures and executor reader.
- Reports record evidence; a structured `human_action` creates/reuses the exact current request and an actionable relay in the same local transaction. Omitted field preserves it; explicit clear can resolve a worker-authored information request. Changed question gets a new identity. A report/ready status cannot clear or approve a native permission/elicitation.
- Routine progress without an answer or human action stays internal and does not wake Foreman. Relevant results, consultations, job completions and requests wake the appropriate owner. Consultations return promptly with an ID, then an attributed answer event; Foreman answers from original scope/context or asks the human. Workers gain no nested delegation or administrative authority.
- One connector consumes KLM's ordered event pages. For each event, commit deduplication, business-state changes, Chat request/publication or outgoing Foreman input intent and the cursor together. Deliver pending intents with stable IDs; SSE only wakes this drain. Use startup/reconnect catch-up and bounded transport reconnection backoff, not model polling or a second native scheduler.
- Publish structured native request cards immediately from retained events, independently of Foreman's prose. Store exact human answer and outgoing reply before delivery. Request/publication/reply/native-resumption facts stay separate. A stale or replaced request refuses a conflicting answer. A question answer never becomes a permission grant.
- Worker finish records structured result and evidence; KLM supplies native finality. Foreman separately assesses the original objective against current settled evidence with a reason and result links. Completed requires satisfied assessment and confirmed worker/job/effect settlement. A successful exit or unfinished PR cannot complete the Ticket. Definitive unmet work is Failed with a reason; uncertain native work is Interrupted. Keep a settled result visibly assessment-pending if Foreman is unavailable, without automatically rerunning work.
- Keep Stop and Cancel independent. Cancel HTTP requires the confirmation's captured Ticket/Assignment/work identity and expected state/version; dismiss/Escape makes no request. Persist Cancelling before calling KLM; mark Cancelled only after cessation. No rollback of completed file effects.
- Use native compaction. On an idle compatible restart continue the verified native identity; on configuration change open a new native generation under the same Chat with role, recent conversation and bounded live-work overview. Historical evidence remains readable. Lost active turns remain interrupted/unknown until reconciled; input/effects are not replayed to manufacture completion.

**Done when:** a real human request can produce one worker, an exact human intervention in Chat, a retained result and a separate assessment, with concise source-directed replies and no routine-progress narration.

### 7 — Integrate the existing frontend rather than rewrite it

**Owner: Werbert frontend.**

Files: W `frontend/src/main.tsx`; existing `features/foreman-mock/{context,types,ForemanMockApp,ConnectionPage,ForemanSettingsPage,RuntimeControls,DomainInstructionsEditor,LibrarySelection,ForemanChatPage,TicketsPage}.tsx` (and `.ts` where applicable), `chat/{ChatTimeline,MessageComposer,RequestCards,ExecutorChat,MessageQueue,ReadonlyRuntime}.tsx`; **new** `features/foreman/{transport,store,types}.ts` and real entry/provider.

Extract/reuse the existing view components behind a narrow data/actions interface. Keep fixture reducer/ScenarioPanel exclusively at `/dev/foreman`; production `/foreman` mounts a real provider before legacy Setup. This is a feature-local seam, not an application-wide state-framework migration. Keep scoped KLM CSS and Werbert design primitives; KLM frontend remains the comparison reference and is not modified for this UI.

| Mock action / view | Real integration |
| --- | --- |
| discovery/Connect/project selection | Backend connection state, catalog/project inventory, stable Factory initialization; remove connection timers from the real provider. |
| runtime choices / Library / instructions | Real per-harness catalogs and Refresh; no automatic model fallback. Keep domain label, full-page Markdown Write/Preview, Escape/Collapse draft retention, explicit modal allowlists and compact saved/applied/pending status. |
| Send / Queue / removal / Stop | Backend receipt plus pending state; exact IDs persisted for unknown-delivery lookup. Use desired/applied capture separately. Keep normal next-step Send and explicit next-turn Queue, without Send now. |
| request cards | Shared authoritative request in Chat and Ticket view; render the native schema/choices, not fixture `repository/includeTests` fields. Real MCP URL flow opens only the exact request's supported URL; link opening is not acceptance. |
| Cancel Ticket | Existing confirmation remains; send only on Confirm, with captured worker identity and expected version. Render intent then actual cessation. |
| Open chat | Independent paged native event/history/content reader keyed by existing work ID, plus captured worker model/effort. SSE invalidates/appends its window. No new execution/session or interactive executor endpoints. Closing tab changes only local view state. |
| results / assessment | Bounded previews and digest-bound full content; actual result links; separate assessment-pending/satisfied/unsatisfied evidence. Download does not interpret EOF as success. |

Replace `MockState`'s scenario/classification fields and fixture-native state transitions with server DTOs. Drafts, open tabs and scroll anchors can remain client-local; execution/history/requests/receipts cannot. Use decimal-string revisions/cursors rather than incrementing the fixture's JavaScript counter. Real Ticket DTOs allow an unassigned Ticket/admission failure instead of requiring `assignment` before one exists. Keep original human text distinct from the derived objective.

Preserve the KLM-shaped timeline/composer interaction: bounded independent reader windows, tail following only at the real tail, paused anchors, Copy and Foreman-message favorites. Store favorites as Werbert UI metadata; worker tabs have no favorite mutation. Connect file mentions through KLM's existing project-path catalog and retain mention metadata. Reuse KLM's existing transcription endpoint through the backend for the source-shaped microphone flow when configured; show its actual availability/error, never fixture transcripts. These are thin integrations, not new file-browser/transcription products.

**Done when:** the same user journey works through real backend data; requested FX controls survive integration; opening/closing inspector tabs has no execution effect; old live routes and `/dev/foreman` remain independent.

## 4. Minimum public contract

Use existing KLM project/model/Agent/Library reads. The browser calls Werbert only; Werbert calls KLM. Route spellings below are proposed implementation choices, not existing endpoints.

### KLM runner surface

Retain R's resource-shaped surface in one small module rather than adding a generic RPC platform:

| Surface | Minimum behavior |
| --- | --- |
| `GET /api/runner/capabilities?projectId=…` | Engine/protocol identity, per-harness mechanism support, readiness diagnostics and supported request/job kinds. |
| `POST /api/runner/connections` | Create or explicitly reconnect stable local integration under authenticated client. Reconnect binds retained integration key/identity; never match by display name. |
| `POST /api/runner/conversations`; `GET /api/runner/conversations/{id}` | Idempotent coordinator creation without inference; current capture/generation/turn, inputs, pending requests and evidence references. |
| `PATCH /api/runner/conversations/{id}/configuration` | CAS desired capture; apply at a drained boundary; invalid desired configuration blocks new admission. |
| `POST /api/runner/conversations/{id}/inputs`; `DELETE …/inputs/{inputId}` | Immutable source-attributed input with steer/queue/passive mode; delete only an unclaimed queued human input. |
| `POST /api/runner/work`; `GET /api/runner/work/{id}` | One asynchronous saved-Agent work admission per external Assignment key; captured resources, current state and finality. |
| `GET /api/runner/events`; `GET /api/runner/content/{id}` | Retained bounded event/history reads, exclusive opaque cursors/high-watermark and immutable digest-bound content. Optional conversation/work filter serves independent readers; consumer catch-up stays integration-wide. |
| `GET /api/runner/updates` | SSE wake hints scoped to the integration; credentials rechecked/revocation closes stream. No ordinary update subscription. |
| `POST /api/runner/tool-calls/{id}/reply`; `POST /api/runner/requests/{id}/reply` | Distinct application-tool result and explicit human-native answer paths; exact owners and immutable reply identity. |
| `POST /api/runner/conversations/{id}/stop`; `POST /api/runner/work/{id}/cancel` | Expected current turn / exact work, with request receipt separate from cessation. |
| `GET /api/runner/operations/{operationId}` | Original immutable receipt, including pending/unknown intent, separate from current execution state. |

All mutations require an owner-scoped bounded `operationId`, normalized request digest and expected revision/target when applicable. Same body returns the same receipt; changed body conflicts. IDs for integration/ownership/generation come from authenticated capture, not model claims. Preserve existing full paired-client trust: this is not a new tenant-isolation guarantee.

Consolidations versus the proposed inventory: one event reader also supplies filtered history; the existing content machinery supplies large outputs; one receipt store covers all runner mutations; existing catalogs are reused unchanged. SSE and protected discovery metadata are the two small additions. Shell jobs are owned model tools/events, not a separate public job-management product. Configuration stays an explicit pending operation rather than silently changing the immutable creation request.

### Werbert browser surface

Implement a feature-scoped `/api/v1/foreman` group: connection discovery/connect/status, KLM project/catalog reads, Factory initialize/read, definition read/CAS save, Chat input/queue removal/Stop, Ticket list/detail/Cancel, exact request reply, event/history/content/operation reads and an update stream. Reuse existing HTTP decoding/error/content helpers. No graph/automation/Agent-authoring/Library-authoring endpoints. An unavailable resource is an explicit error, not a fixture fallback.

### Application tool manifest

Keep a fixed revisioned manifest, not configurable executable callbacks:

- **Foreman:** Agent list/read; Ticket open/read/assign; bounded history/result reads; consultation answer; purpose-bound Chat reply; objective assessment. Ticket admission validates the retained real human source and task scope. Cancellation, if requested through Chat, must lead to the same explicit UI confirmation, not bypass FX-014 through a model tool.
- **Worker:** source-bound acknowledgement; report with optional structured human action; consultation request/read; owned evidence read; structured finish. Immutable Assignment supplies ownership. No model-selected owner ID grants authority.
- Replies have `purpose` (`acknowledgement`, `attention`, `answer`, `result`, `inability`), source input/request/result reference and text. Attention must match the current question; a result must reference usable evidence. Store purpose/source deduplication and publication receipt. Proposed 100-word ceiling applies to notification prose, not full question/options/evidence.

## 5. Minimum persistent records and lifecycle rules

Use current SQLite storage infrastructure. These are logical record groups, not a requirement for a table per noun; implement indexed IDs/state/revisions plus existing bounded JSON/content patterns. Keep growing histories out of giant aggregate rows. No shared database or cross-application transaction.

| Owner | Records and essential uniqueness |
| --- | --- |
| Werbert | Connection: engine ID, stable integration ID, protected credential reference, committed event cursor. Factory: unique engine/project, registered path, Chat ID, desired file digest and last applied capture. |
| Werbert | Chat input/event: stable ID, original text/mentions, producer/source, delivery mode/state and runner input/batch reference. Ticket: unique Chat/human-input source, objective/criteria, state/version, optional single Assignment. |
| Werbert | Assignment: unique Ticket, saved Agent full source/revision, task, outgoing operation/work ID and capture; result reference and separate assessment `{basisWorkRevision, outcome, reason, resultRefs}`. Reports/requests/Chat publications append separately. |
| Werbert | Operations/outbox: owner/action/operation ID, immutable body digest, intent/result receipt. Incoming events deduplicate `(engineId,eventId)`; cursor advances in their local projection transaction. Human replies retain exact request/reply ID/body and delivery state. |
| KLM | Integration binding; conversation key unique per integration/project; native generation/configuration capture; external work key unique per integration. Hidden technical sessions retain existing native/history records. |
| KLM | Inputs/batches: source kind/ID, payload digest, mode, generation/turn/step, retained/claimed/accepted/included/unknown state. At most one active coordinator turn; a batch cannot be claimed by two steps. |
| KLM | Application calls and native requests: execution/generation/turn/call identity, captured manifest/schema/choices/scope, pending/replied/included/resolved or unknown state. One immutable reply per current request/call. |
| KLM | Shell jobs: owner, command digest, originating turn/native call, start intent/process identity, bounded output/content references, exit/cancel/drain evidence. Unique execution/call prevents restart-on-retry. MCP invocations are not new job kinds in this delivery. |
| KLM | Ordered runner events and operation receipts: integration/sequence/type/owner IDs, content references, request digest and original receipt. Technical event + associated state commits atomically. |

Ticket labels retain the eight requested states: **Queued, Working, Waiting for user, Completed, Failed, Cancelling, Cancelled, Interrupted**. With one worker, Waiting for user requires evidence that no executable path, including an owned background job, can proceed without the answer; a pending question alone is insufficient. A settled worker can remain Working with **assessment pending**, but must not appear to still be executing.

Necessary durable transaction boundaries:

1. Retain human input and receipt before delivery; retain Ticket + unique Assignment + outgoing admission intent before HTTP.
2. KLM persists work + receipt before native launch. Werbert recovers acceptance by original operation/key, never a replacement session.
3. Worker report/request + local relay intent commit together. Native request + KLM event commit together. Event ingestion + local projection/publication intent + cursor commit together.
4. Human answer + exact outgoing reply commit before native forwarding. Publication in Chat is distinct from person seeing/responding and from native continuation.
5. Finish evidence + terminal event commit together; assessment checks a current settled basis. Cancellation intent races against that same current work/version, preserving prior effects.

**Recovery policy:** restart resumes durable *delivery* only when it is idempotent or reconciliation proves no native inclusion. Unknown native input/tool/shell effects remain interrupted/unknown; never automatically repeat them. Preserve compatible idle native continuation. Reconnect to known continuing worker identities. A lost native pending callback remains visible as unavailable/interrupted, not recreated in a new worker. New unrelated work need not wait on unrelated historical uncertainty; the conflicting owner/generation remains fenced. No automatic worker retry policy is introduced.

## 6. Reconciliation of mock deltas and pinned DeepSeek mechanism

### Experience ledger

| FX | First-delivery treatment / approval boundary |
| --- | --- |
| 001 | Preserve requested KLM timeline/composer, fixed read-only applied runtime and YOLO. Pixel/interaction acceptance remains pending. |
| 002 | Preserve next-step Send, explicit next-turn queue and exact unsent removal; no interrupt-and-send. Real native boundary proof remains required. |
| 003 | Keep isolated dev fixtures; production SQLite replaces browser execution snapshots. `/foreman` production entry is a proposal, not inferred approval of the mock route. |
| 004 | Propose the compact navigation only for the new local entry. Preserve incumbent live navigation/routes; no global product removal. |
| 005 | Replace fictional connection/definition/runner/media data with actual services. Missing adapters produce real unavailability; no production scenario controls. |
| 006 | Required independent Chat request visibility is retained. Identity disclosure density and receipt wording remain presentation proposals. |
| 007 | Remove fixed review criteria/classification/outcomes. Actual Foreman chooses an existing Agent for the retained objective; no hard-coded Code Reviewer product role. |
| 008 | Preserve source-shaped styles/interaction and independent bounded readers; replace local history arrays with cursor/content reads. Visual fidelity is a human check. |
| **009** | **Requested:** KLM-style harness icons/names, searchable model picker/Reload and effort slider in Settings. Explicit model selection, supported default effort only. |
| **010** | **Requested:** Domain-specific instructions label. Storage remains `foreman.instructions`. |
| **011** | **Requested:** full-page document-sheet Markdown editor; Escape/Collapse retains the same draft. No rich-text format or implicit save. |
| **012** | **Requested:** KLM-style Library modal, search/detail/Refresh and exact allowlists. **Proposal:** Use selection stages the draft; Save settings commits atomically. No Library authoring or automatic tool selection. |
| **013** | **Requested:** remove redundant success/no-change prose. Retain compact saved/applied/pending state and actionable errors. |
| **014** | **Requested:** exact Ticket/Agent cancellation confirmation; dismiss records nothing; intent is not cessation. Server rejects stale ownership/state. |
| **015** | **Requested:** read-only event-updated executor tab with worker capture. Real native messages/work are required, not only synthesized report cards. Replies stay in Foreman Chat/Ticket request surfaces. |

The originating user has requested adjustments, not confirmed acceptance of the whole mock. During an authorized implementation, update W's maintained experience/FX documents with actual mappings/checks and any new difference; do not rewrite their current proposal labels as approvals.

### Internal mechanism, not notification-only similarity

Pinned reference: [`deepseek-ai/dsh-superagent-prototype@8c32a9dc5d72c033816c6a8110b37f40cf94fba4`](https://github.com/deepseek-ai/dsh-superagent-prototype/tree/8c32a9dc5d72c033816c6a8110b37f40cf94fba4). For this plan, its `cordis.patch.yml`, `src/app/delivery.ts`, `src/app/notices.ts` and `src/app/tools/executor.ts` were fetched directly; F lists the other pinned behavioral references.

- Router and direct-worker presets set default **and maximum** shell and job waits to 5,000 ms. Implement at the runner/tool layer, preserving one live job. The prototype's nested standard-preset subagents use 60 seconds; nested delegation is outside this slice rather than falsely advertised as five-second coverage.
- `AgentDelivery` checks durable history and pending inboxes, retains stable producer message IDs, distinguishes steer/queue/passive injection and flushes before acknowledgement. Reproduce that distinction with KLM's runner ledger and actual native boundaries, not ordinary `/messages` calls with invented human events.
- `reportRelay` returns no relay for routine progress without an answer; a question can relay while work continues. Correct the inspected W routine-progress wake behavior in the **new** path.
- Executor report + attention/result relay are atomic. Human-action identity/routing and worker-owned acknowledgement survive retry. Publication/answer/settlement remain separate evidence.
- Intentional platform adaptations: KLM owns shell jobs/native lifecycle; W owns business records and Chat publication. Structured native permissions and elicitations are an addition to the reference's unattended executor model. Durable request cards must not depend on model success. Native harness compaction is used; no AI Memory bootstrap is inherited.
- Approved first-delivery policy: Foreman and every Factory worker always run YOLO, without changing personal KLM settings. Necessary task questions and MCP elicitation remain human actions. Async conversion of long MCP calls is deferred; the shell/job limit is not a universal tool-response limit.
- The 100-word notification limit and compact navigation/staging choices remain proposals from F/FX. This plan preserves them for review, not as newly manufactured approvals.

## 7. Small focused checks and manual acceptance

**This planning task ran no build, test suite, provider execution or server.** Verification performed here is source/spec inspection, local-link/file-reference checking and a documentation-only readback/whitespace check. The following checks belong to later authorized implementation.

### Lightweight engineering checks

Use one relevant Go build/type check per changed backend and one frontend TypeScript/Vite build when integrating its changes. Run small existing targeted checks where applicable; do not run full suites or overlapping broad workflows. If adding focused tests is authorized, prioritize only these new invariants in temporary stores/peers:

1. Same admission/source/key → one work; changed same-ID body conflicts; lost acceptance/committed-delegation-before-Stop resolves original receipt.
2. Input arrives after claim → later batch; passive causes no idle wake; serial coordinator; exact queued deletion refuses claimed input.
3. External records excluded from ordinary list/count/exact-ID/stream/control and permission namespaces; legacy origin still follows the old path.
4. Report/request + relay/cursor transaction rollback; repeated/stale answers; result cannot become Completed without current settled assessment.

Step **0** performs the short native feasibility probe per **Pi, OpenCode and Codex** before the substantial implementation. The later acceptance pass checks the integrated path and only repeats a native mechanism if changed or unresolved, rather than launching overlapping probe workflows. Observe actual native calls, fixed YOLO on both external roles, one command process, foreground release after five seconds, bounded waits, asynchronous shell completion, steering inclusion, selected tools, necessary question/elicitation, Stop/drainage and compatible continuation. One ordinary comparison confirms its shell/queue/catalog behavior did not acquire Factory policy. Unsupported mechanisms remain recorded gaps; fixture/build passes do not close them.

### Complete manual journey

Use a disposable registered project and a small existing saved Agent with a harmless task that intentionally asks one concrete question and produces inspectable output. A Code Reviewer is a useful workload, not a required specialist. Installation/provider readiness remains in KLM.

| Check | Observable acceptance |
| --- | --- |
| Local entry | Discovery reaches 7331 without pairing; Connect completes local handshake without manual code copy; reload retains binding. Expired/revoked/incompatible/mismatched engine shows a real actionable state without duplicate work. |
| Project/configuration | Select project twice → one Factory/Chat and preserved existing YAML bytes. Add Factory works. Missing definition starts unconfigured. Invalid/legacy definition is preserved. Stale save conflicts while retaining draft. |
| Requested Settings controls | Validate FX-009–013 in light/dark, keyboard/Escape/focus return and a narrow layout. Refresh does not select new tools. Composer still shows applied read-only runtime/YOLO. |
| Real delegation | Human text retained; substantive work produces one Ticket and one Assignment with exact Agent source/revision; one concrete worker acknowledgement. Operational status question creates no Ticket. Foreman yields while worker runs. |
| Internal mechanism | A command exceeds five seconds without rerunning; inspect same job and eventual output. New actionable input enters next eligible step; explicitly queued input waits for next turn; passive/routine progress causes no idle turn/narration. |
| Human action | Foreman and worker captures both show fixed YOLO; eligible execution approvals are not prompted as an on-request policy. A necessary task question or selected MCP elicitation still appears in Foreman Chat without opening Tickets. Stop/fail Foreman and verify the card remains; reply to the exact current request. If an actual upstream permission remains outside YOLO bypass, preserve its offered scope and explicit reply rather than auto-approving it. |
| Dedup/reconnection | Drop one HTTP acknowledgement and one event connection; recover original operation/cursor. Exactly one Assignment, request, answer and notification. Backend/UI reconnects without resubmitting the human task. |
| Worker inspector | Open chat shows actual messages, tool work, request and eventual result live; worker model/effort is accurate. Switch/close/reopen tabs; no input/reply/Stop/favorite mutation exists and no execution is created. |
| Result/assessment | Worker result arrives first; separate Foreman reason/evidence establishes satisfied or unmet objective. Unfinished PR/native success alone never says Completed. Full output and digest/chunk continuation remain inspectable. |
| Stop versus Cancel | With two independent Tickets active, Stop Foreman leaves both workers running. Cancel dialog dismiss/Escape does nothing; Confirm cancels only the captured worker and shows Cancelling until confirmed. Uncertain cessation remains Interrupted. |
| Boundary/continuity | Change Settings during a turn; steering and active worker retain captures; later drained Foreman generation applies new settings under the same Chat ID. Native compaction retains transcript/work references. Engine loss preserves interrupted evidence without replay. |
| Unaffected KLM | Before connection, while connected/working/waiting, and after disconnect: ordinary session creation/search/topics/archive, direct known-ID access, Happening now/Focus, requests, analytics and general-agent discovery expose no Factory work. Personal model/effort/Library/YOLO, native tools/shell behavior, queue/steering/Stop and compaction behave as before. |

No deployment, preview server, full suite or review agent is part of this plan-writing task. Later human acceptance is the release gate for behavior; compilation and targeted probes establish only their recorded scope.

## 8. Deferred work and blockers

**Deferred:** remote/multiple engines; worktrees/branch management; automatic cleanup/commit/push; graph/automation/webhook integration; multiple Assignments/dependencies/teams; automatic worker retries/replacement pipelines; new Factory-specific quotas or human-wait parking scheduler; Agent/Library authoring and installation in Werbert; mandatory AI Memory, managed consolidation/self-improvement; operational legacy migration; new deployment/installers; asynchronous MCP/non-shell job conversion; configurable worker execution-mode/grant editors; generic integration/plugin frameworks and exhaustive crash/provider permutations.

These deferrals do not remove pending native-request delivery, durable receipts, cancellation truthfulness, current-checkout safety, full retained history or the five-second/input mechanism. Those are first-delivery necessities.

**Blocking decisions for writing this plan: none.** The early native proof, deferral of MCP jobs and always-YOLO workers are explicitly approved adjustments. The old worker non-YOLO proposal is superseded. Local-control discovery response, production entry/navigation, notification ceiling and staged Library save remain proposed choices for review. Approval of the three adjustments is not approval to implement the complete plan.

**Implementation/acceptance dependencies:** (1) approval of this plan and its remaining proposals; (2) actual pinned-harness proof of external tools, next-step inclusion and enforced shell replacement, especially Codex; (3) an available existing saved Agent and selected Library resources for the human journey. No concrete harness impossibility was established by static inspection. If a required mechanism proves unavailable, bring the exact failure back before narrowing coverage or substituting semantics.

**Handoff:** report this file once to originating conversation `b54fa8f416b75f42f74d4ac541d1edcc`. Its review and the user's validation precede implementation; delivery acknowledgement is not plan approval.
