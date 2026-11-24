# Agent Teams runner delivery

Read when continuing the generic runner support used by Werbert Agent Teams.
Source implementation is complete; human acceptance and deployment are separate.
The user commissioned this cross-checkout implementation on November 15, 2026;
earlier linked-worktree evidence was recorded November 16 UTC. Initial-CWD
compatibility was implemented November 18, 2026; its runtime acceptance is pending.
No commit, push, installation or deployment occurred in those source-delivery
stages. The user separately authorized the November 19 VPS publication below.

## November 19 VPS publication

The coordinated Linux release build is installed in the existing KLM service with
the matching Werbert backend/prompts/frontend. Authenticated runner capabilities
confirm `initial_cwd_workspace:true` and the required Teams extensions; live binary
hashes, health, original database/connection identities and retained records were
checked. This is deployment/readiness evidence, not provider/recovery or relaxed
directory execution acceptance. Local installed/personal services were not replaced.

The [publication and rollback receipt](../../Hermes/werbert/docs/testing/klm-foreman-vps-deployment-20261119.md)
records exact artifacts, private stopped backups, checks and remaining acceptance.

## Contracts and ownership

`engine/RUNNER_API.md` is the maintained protocol reference. Additions reuse
runner captures, operations, inputs, native contexts, jobs and effect ledgers:

- `POST /api/runner/captures` prepares an immutable engine-owned saved or embedded
  participant without inference. The original operation receipt freezes exact
  source/runtime/independent capability data.
- Prepared `POST /api/runner/work` consumes `preparation_id` and propagates a
  validated existing absolute initial CWD without changing the registered project.
- Exact worker input lookup and continuation preserve original execution/session,
  generation and native context. Optional Git/base/identity observations are
  provenance only and never future launch gates.
- `GET /api/runner/groups/{instanceKey}?project_id=...` reads retained owners;
  `POST .../cancel` takes `{operation_id,project_id}` and durably fences captures,
  delayed admissions, claims, inputs, effects, native replies and recovery before
  invoking existing cancellation. The immutable cancel receipt is not drainage.
- Opt-in `recovery_policy:"resume_drained"` now reconciles durable native/job
  ownership, committed application replies, unknown inputs, MCP effects and lost
  callbacks before automatically continuing the original context. It does not
  replay uncertain commands/edits/MCP effects or invent exit outcomes/decisions.
- Windows opted-in processes retain named Job and engine PID/creation identity;
  Linux reuses subreaper/pidfd supervision with a durable drainage receipt. Missing
  ownership evidence remains explicit. Cancelled instances never resume.

New mechanisms are advertised through runner capability extension flags.
Personal-session defaults, authentication, visibility isolation and existing
non-opted-in recovery remain separate. Team business identity, YAML authoring,
instance FIFO, leader/member roles, questions and authoritative delivery belong to
Werbert, not a new generic orchestration subsystem.

Main K files: `engine/runner_{preparation,workspace,groups,ownership,cold_jobs,
mcp_recovery,recovery}.go`, platform ownership/workspace helpers and existing
runner API/runtime/store/controls/effects/jobs plus narrow native/process/MCP seams.

## November 18 initial-CWD compatibility

The runner advertises `initial_cwd_workspace:true` and
`linked_worktree_binding:false`, retaining every other existing runner extension.
Werbert must negotiate the initial-CWD extension alongside immutable preparation,
worker input lookup, resume, cancellation fencing and cold reconciliation support.

Preparation defaults an omitted/empty workspace path to the registered project
directory. Shared/overlapping, non-Git, other/nested-repository and arbitrary
existing worktree directories are legal. Admission, input preparation, turn
scheduling and cold continuation check the actual absolute launch directory and
existing/nonremoved project policy; they do not require a Git base, exact linked
root, same repository or retained directory/Git identity. Observations remain
best effort. Stored `Session.ExecutionCWD` supplies continuation, native and owned
job launch paths. No worktree lifecycle or shell sandbox was added.

Path reservations were removed from preparation; integration/project/instance
binding, `PreserveSource:true` source resolution, immutable captures, native IDs,
receipts, owned domains, cancellation and confirmed drainage remain intact. All
legacy DTO fields/tags and preparation/create marshalled request digest encoding
are unchanged; no new request fields or SQL migration were needed. Old records
are read under the relaxed semantics without rewriting them. Personal sessions,
graph workspaces and personal webhook/schedule contracts retain their policies.

Changed files for this slice: `engine/runner_workspace.go`,
`engine/runner_preparation.go`, `engine/runner_capabilities.go`,
`engine/runner_api.go`, `engine/runner_runtime.go`, `engine/runner_recovery.go`,
`engine/RUNNER_API.md` and this document. Existing dirty/untracked Teams and user
work was preserved.

Validation: the requested Windows dev build passed (exit 0, no output), from
`engine/`, after verifying the temporary output parent:

```powershell
go build -tags dev -o C:\Users\Samuel\AppData\Local\Temp\opencode\klm-initial-cwd-dev.exe .
```

This was the only compilation/runtime check for this slice; no test suite or
provider/runtime probe ran. Runtime checks of shared, non-Git,
nested/other-repository paths, old receipt retries, unavailable launch directories
and original-context continuation remain human validation. The running
coordinator engine was not restarted, stopped or replaced; its capability response
therefore does not establish this source change. No platform-specific source was
changed, so no additional Linux cross-build is required for this slice.

## Earlier linked-worktree checks and evidence

The following results belong to the former mandatory linked-worktree contract.
They establish no runtime acceptance of the November 18 directory relaxation.

| Check | Recorded result |
| --- | --- |
| K dev backend build | PASS, Windows |
| K Linux amd64 cross-build | PASS, `GOOS=linux GOARCH=amd64 go build -tags dev -o <owned temporary output> .`; catches Linux-only imports/types, not Linux runtime acceptance |
| W backend build / frontend `npx tsc -b` | PASS in companion checkout; final frontend check followed Ticket participant integration |
| Focused cancellation uncertainty regression | PASS, 4.702 s |
| Slice 0 provider/drained restarts | PASS, 67.63 s, Codex 0.153.4 |
| Abrupt cold runner provider workload | PASS, 142.68 s, Windows/Codex; original member session/thread/capture/worktree and included recovery source |
| Group fence/isolation + real cold owned job | PASS, 3.021 s; retained output with unknown exit outcome, no replay |
| Original reply/unknown-input reconciliation | PASS, 0.991 s |
| W production Team provider workload | PASS, 236.18 s: expanded production roles/tools, whole-roster capture, same-member correction, separate W restart, abrupt K death, original contexts and one settled leader/Foreman result |
| Scoped real browser | Team YAML create/reload/detail, missing reference, draft/discard and retained Ticket presentation passed; populated native history and production discovery were not established |
| Diff whitespace | PASS in both checkouts |

No full suite, adversarial/review-agent round or broad harness matrix ran. Saved
specialists were owned leaf subagents; no independent sessions were created.
The isolated Pi version probe succeeded at 0.85.1, but new Team native recovery
proof is Codex 0.153.4 only. OpenCode 1.18.35 was available, not exercised here.

Raw production native evidence is retained at
`C:\Users\Samuel\AppData\Local\Temp\opencode\werbert-team-production-native-20261116-02\evidence.json`.
The retained artifact has exactly `INITIAL` and `CORRECTED`; execution/session/
generation/thread and applied captures survive the real correction/restart path.
The crash followed a completed edit, before the remaining sleep; the separate
owned-job check supplies interrupted live-job evidence. Temporary protected native
auth was removed, original auth stayed unchanged and owned processes exited.

## Companion specification and continuation

The companion checkout is `C:\Users\Samuel\Documents\Projects\Hermes\werbert`.
Maintained contract/evidence:

- `docs/contexts/agents-and-teams/features/agent-teams.md`
- `docs/testing/agent-teams-implementation.md`
- `engine/docs/foreman-api.md`, Agent Teams section
- `engine/docs/PROMPTS.md`, immutable Team role capture exception

Those documents are sufficient without temporary planning files. Raw diagnostics
and detailed leaf handoffs remain in its ignored `temp_doc/` while human
validation is pending. Preserve both checkouts' current dirty source; W includes
unrelated concurrent Directory/PromptEditor, Agents/Teams and Datasources work.

Remaining human checks: matching retained W/K native-history presentation, normal
connection/release UI/personal-session regression, multiple-member parallel work,
real human native decisions/callback restoration, Pi/OpenCode/Linux cold recovery
and selected real/nonqueryable MCP effects. Missing workspaces/history/credentials,
unconfirmed process ownership and unknown effects retain specific diagnostics.
No known blocker appeared on the demonstrated Windows/Codex production path.

Rebuild K and W together before manual use; old running binaries do not advertise
the new flags. K: `engine/` → `go build -tags dev .`; W: `engine/` →
`go build ./cmd/engine`; W frontend: `npm run build`. Use `WERBERT_MODE=foreman`,
explicit isolated data roots and the documented dev control/API connection.
Frontend Vite uses `WERBERT_ENGINE_URL` for its backend proxy. No installed/manual
service update is implied by source/build completion.
