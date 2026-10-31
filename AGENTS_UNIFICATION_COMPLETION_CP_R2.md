# Completion owner R — CP-R2

Publication note: original storage schema3 evidence below predates integration
with PRs #39/#40. Final stacked storage is schema4 with the preserved content2 →
catalog3 → delegation4 chain; see `AGENTS_UNIFICATION_PUBLICATION.md`.

Status: **integrated candidate delivered; stopped for manager review**.
Owner `92467a25db7be6788388129d542b401f`; manager
`a05369eaa5c9151b61e93fb38cb35c52`. No CP-R3 or final validation-session creation
is implied by report acceptance.

## Graph runtime

- V2 SQL node launch attaches the existing captured snapshot and exact
  run/node/activation/engine-turn/workspace/snapshot digest to its turn. V1 does
  not gain the new tool or get relabelled/recompiled from live configuration.
- Graph parents now use the CP-R1 wrappers/service/adapters. Specialists resolve
  through lookupFrozenGraphAgent/composeFrozenGraphAgent only. Unnamed children
  use effective parent settings and an explicit empty frozen Library unless the
  human-request-only inheritance boolean is true. No live catalog/source/Library
  resolution occurs inside graph child preparation or the graph parent model step.
- Reused children require the same engine parent, run, workspace, canonical cwd
  and fixed harness. Their previous immutable capture/history is retained; they
  do not acquire an ordinary live source binding. Children have leaf capabilities,
  never Choice, graph execution/session administration or further delegation.
- Preparation/admission and Choice reservation use the existing app mutex/writer.
  Choice rejects an active preparation or batch/delivery fence. Admission rechecks
  the actual open activation and current turn in the same commit. Choice acceptance
  rechecks after join/drain/native delivery, then uses the existing seal/drain/
  acceptance transitions. No runner/scheduler/worktree replacement or Linux graph
  support expansion. Graph-context shutdown cancels children; chat Stop retains
  its independent-graph boundary.
- Graph child finality errors propagate as graphUnconfirmedError. Ordinary and
  graph alias correlation share the same guarded delivery helpers. Batch state now
  reflects preparing/running as actual child dispatch/native acceptance occurs.
- agents_list/agent_get are registered on V2 graph parents against frozen entries
  only. subagent_output_read is parent/activation-scoped. Existing Codex new-thread
  manifest compatibility and the complete v1 dynamic spec/digest are unchanged.

### Minimal missing metadata seam

GraphFrozenAgentCatalog now optionally includes **public ModelCatalog metadata**
for harnesses already used by captured nodes/eligible specialists. It is captured
at the existing run-start boundary, JSON-cloned to exclude private provider options/
endpoints/auth data, sealed by the existing catalog digest, and included in its
existing 16 MiB bound. This permits legitimate model/effort overrides and defaults
without live provider discovery in a graph node. No DB/schema/ledger/protocol
version or second snapshot store was added. V1 is untouched; older V2 artifacts
without this optional field retain their original digest and accept their existing
captured effective pairs rather than inventing a live catalog.

## Shared UI

- Managed batch history cards consume the actual member/output APIs and show all
  children, terminal counts, waiting/delivery, factual outcome, title/key, harness,
  effective model/effort and reused invocation number. Managed per-member transport
  events do not produce duplicate legacy cards. Existing native child cards/history
  remain, with failed/cancelled labels no longer called completed indiscriminately.
- History opens a native-dialog read-only ConversationHistory backed by the actual
  child window/history/scoped updates. It stays local to the viewer and never adds
  an owned child to top-level inventory or Happening now. Child views do not render
  pending-request reply controls, composer, model settings or favorite mutations.
- Output follows digest-bound UTF-8 pages (16 KiB) with First/Next navigation; it
  does not concatenate an unbounded result. No captured configuration is returned.
- Native PermissionCard/QuestionCard (including elicitation) render in the owning
  main, side/general/Focus/mobile-web composer via the actual request projection.
  Replies target the visible owner and exact child request; scopes/grants and human
  decisions remain native. Shared scope invalidations coalesce rather than cancel
  the same-key in-flight read; focus restores reads paused while hidden.
- Session GET now carries delegation availability/waiting/actions. The composer
  queues when steering is fenced, keeps Stop available, and respects send/interrupt
  actions; queue Send now and model/agent controls use factual action flags. Codex
  legacy-history unavailability is a short fact inside Session options, not a
  persistent tutorial. YOLO retains its existing live control/server downgrade gate.
- Graph node Activity renders the same managed batch card against its exact run/
  activation API scope, including opening child histories and bounded output. Escape
  closes the child dialog without dismissing the enclosing graph inspection.

## Final API/DTO additions

CP-R1 routes remain below `/api/sessions/{id}/delegation`:
`GET /batches`, `/batches/{batchId}/members`,
`/batches/{batchId}/invocations/{invocationId}/output`, `/children`, `/requests`.
Paging/digest/UTF-8 bounds are unchanged. Request entries now also carry childTitle.

Graph batch/member/output equivalents are below:

```text
/api/sessions/{visibleOwnerId}/graph/runs/{runId}/activations/{activationId}/delegation
```

They verify the visible run owner and actual activation/node session, then restrict
every batch/member/output read to that activation. Public node chat/control routes
remain hidden. Generic read-only child session/history/content/events reads validate
the managed graph child's retained run/activation/workspace ownership before serving.

`GET /api/sessions/{ownerId}/delegation/requests` includes that visible owner's
active graph-child requests. Replies use existing owner permission/question routes,
not a public child mutation. Child pending changes invalidate both the engine node
parent and visible graph owner. Session GET adds:

```text
delegation: { availability: {available, code?, detail?}, waiting,
              actions: {send, queue, steer, interrupt, settings, stop} }
```

No child composer, top-level child creation, graph Stop endpoint or new mutation
protocol was added.

## Files

New Go: `engine/delegation_graph.go`, `engine/delegation_graph_smoke_test.go`.
Graph/runtime integration: `main.go`, `sqlite_graph_execution.go`,
`sqlite_graph_transitions.go`, `graph_adapter.go`, `harness.go`, `agent_tools.go`,
`delegation_bridge.go`, `delegation_prepare.go`, `delegation_service.go`,
`delegation_child.go`, `delegation_delivery.go`, `delegation_repository.go`,
`delegation_projection.go`, `delegation_api.go`, `delegation_requests.go`,
`sqlite_execution_requests.go`, `api.go`, `sqlite_api.go`.
Frozen public model metadata: `graph_frozen_agents.go`, `graph_snapshot_v2.go`.
Other R-touched CP-R1 Go files were gofmt-formatted only; no unrelated cleanup.
A's `sqlite_delegation.go` / `sqlite_validate_delegation.go` remain unchanged.

New frontend: `features/chat/DelegationActivity.tsx`, `DelegationRequests.tsx`,
`delegation.css`. Integration: `engine.ts`, `App.tsx`, chat `ConversationEvents`,
`ConversationHistory`, `MessageComposer`, `MessageQueue`, `ModelPicker`,
`SessionOptions`, `SideChatPanel`, agents `AgentMessageComposer`, workspace
`FocusSessionChat`, graphs `GraphRunActivity`. All frontend paths are under
`clients/desktop/src`. Product/coordination docs updated.

## Checks actually performed

- Gofmt on the Go files changed by R (including compact CP-R1 additions), with a
  final formatting pass after the frozen model metadata/default-alignment change.
  Commands had 10-second bounds; exit0.
- Targeted Go compile during wiring caught a string/uint64 revision mismatch in
  frozen discovery paging; corrected to the existing revisionForItems helper.
- Initial graph fixture used a display name instead of the required slug; its
  invalid-request failure was corrected in the fixture, not by relaxing validation.
- `go test -tags dev -run '^TestDelegationGraphJoinChoice$' -count=1 -timeout 45s -v .`
  from engine, tool bound60s: final **PASS1.64s / package1.785s, exit0**.
  Earlier incremental passes were package2.024s and1.929s before the final small
  frozen-harness default alignment. This one central fixture covers real V2
  Save/capture, two-child admission/join, exclusion before join and before native
  result delivery, exact graph ownership, frozen low-effort override after removing
  live catalog access, accepted Choice after delivery, rejection of later children,
  and the owner-scoped graph batch HTTP projection.
- `.\node_modules\.bin\tsc.cmd --noEmit --pretty false --project tsconfig.json`
  from clients/desktop, tool bound120s: final exit0/no diagnostics. Rechecked only
  after frontend integration changes; no build/install/full frontend suite.

The graph fixture is **local storage/coordinator evidence**. Runner output and
native/process/tool finality flags are synthetic fixture data; it starts no native
process/provider, graph scheduler, registered graph execution or worktree. No Pi
native-parent/old native-batch smoke was repeated for CP-R2. No installed engine,
real user data, package installation, commit/push or session/agent spawning.

## Short isolated DEV validation handoff (not executed here)

1. Choose a fresh absolute temporary root under the approved temp parent, with
   existing bin/data/skills/log parents verified before writes. Keep the actual
   installed engine and its data directory untouched. Use an isolated temporary
   project folder when registering a validation project.
2. In engine, set KLM_DEV_DATA_DIR to that root's data directory and absolute
   KLM_USER_SKILLS_DIR to its skills directory. For local dev only set
   KLM_DEV_DISABLE_CLIENT_AUTH=1 and KLM_API_BIND_HOST=127.0.0.1. Configure disposable
   native profiles/local deterministic providers before any native turn; do not
   copy real credentials. Existing pinned installations are sufficient.
3. Build `go build -tags dev -o <root>\bin\klm-dev.exe .`, then run that binary's
   `start` with a finite60–90s command budget. It backgrounds its worker and uses
   the isolated data/control identity. API remains fixed on17331. If occupied, stop
   and report the conflict rather than killing another process or changing ports.
4. In clients/desktop set VITE_ENGINE_URL=http://127.0.0.1:17331 and run the existing
   Vite entry in an owned background process/PID with separate stdout/stderr logs,
   `--host 127.0.0.1 --port 5173 --strictPort`. Bound API/frontend readiness to20s.
   Keep the process out of a foreground/watch tool call. No npm install is needed.
5. Validate catalog/session binding plus a task with two children, request reply,
   full output/history, reuse and Stop, then the frozen graph path when expressly
   authorized for that validation. Use normal UI/actual adapters, not inserted
   transcript events as evidence of model execution.
6. Cleanup through the **same isolated binary/environment** with `stop`; terminate
   only the recorded owned Vite PID. Retain logs/evidence for the manager. Do not
   stop the installed engine or unrelated development processes.

## Remaining gate

Integrated OpenCode/Codex native execution, graph native seal/drain behavior with
real adapters, side/general combinations, actual MCP permission/elicitation paths,
and browser/native/mobile-web visual behavior remain runtime/manual acceptance.
No visual/provider result is claimed from Go/TypeScript fixtures. The manager must
review this candidate and release the next checkpoint/create the final independent
runtime/Playwright session. R remains stopped until that explicit direction.
