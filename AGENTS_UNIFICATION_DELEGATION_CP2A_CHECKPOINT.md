# B CP2a interruption checkpoint

Status: **paused by the user; integration in progress, not accepted or installed**.
Worker `e80963a1b2bdb9c249b765921feb7d28`.

The user reported a completed command whose response did not return and stopped
the worker. Do not repeat the last native probe just to recover this turn.

## Last completed native command

`go test -tags dev -run '^TestDelegationNativeBatchSmoke$' -count=1 -timeout 90s -v .`

The available output reported **PASS**, test 11.45s, package 11.562s:

```text
native batch2/join/reuse/Stop: providerRequests=6;
stopOutcomeError=Delegation finality is unconfirmed; stop the parent without returning a native tool result.
```

This used temporary schema3 SQLite and actual installed Pi adapters with a local
deterministic provider. The parent was a synthetic owned fixture actor; parent
native wrapper/RPC integration was not exercised. Two real child invocations
joined, reused their native histories, and a third pair was stopped while fixture
responses were held. Stop was explicitly not counted as successful completion.
The test used the engine's owned process teardown; its temporary directories were
cleaned by Go. No installed engine, real provider or user data was changed.

Before that, the pure Go contract target passed (0.126s); the hermetic manager
stop-latch and preparation transport fixtures also passed. No full suite ran.

## Work present in the workspace

- B-owned `delegation_repository.go` reuses A readers/CAS writers and provides
  admission lookup, cards, finalization and scoped result reads.
- `delegation_prepare.go` resolves/captures settings outside writer/app locks and
  admits child sessions/members/leases together. Capture authority is A's types.
- `delegation_service.go` owns bounded preparation, max4 dispatch, all-child join
  and cancellation/finalization. `delegation_delivery.go` adds native result
  correlation and bounded wait primitives.
- `delegation_child.go` invokes the existing adapters with leaf role/frozen
  preparation and immediate native/process/MCP retirement.
- Collector now retains every member in <=32KiB summaries with full-output refs;
  bounded UTF-8 result-only continuation does not expose capture/config payloads.
- `delegation_projection.go` provides batch/member DTOs; batch-list progress does
  not hydrate result histories. Public endpoint registration is not complete.
- `main.go`, `harness.go` and the three native adapters contain the initial leaf
  lifecycle/acceptance hooks. Root bridge/profile integration is incomplete.
- `delegation_bridge.go` was being implemented when interrupted; its private
  methods are not yet registered in linked bridge dispatch. The latest bridge
  changes have not been compiled. Two obvious local typing omissions (sql import,
  nativeCallId DTO field) were repaired while recording this checkpoint.

These are WIP facts, not a complete CP2a acceptance checklist.

## Preserve ownership

`sqlite_delegation.go` and `sqlite_validate_delegation.go` are A-owned. A concurrent
filename collision occurred; B preserved A's content and reconstructed only its
own helpers in `delegation_repository.go`. Do not restore the earlier B file over A.

Schema3 was reviewed/released by the manager. Use
`AGENTS_UNIFICATION_DELEGATION_STORAGE_HANDOFF.md`. Graph snapshot v2 handoff is
known, but graph execution/Choice integration remains outside this checkpoint.

## Next integration steps after explicit resumption

1. Compile/check the latest root bridge work once; do not repeat native smoke by
   default. Finish root controller cleanup/premature-completion checks.
2. Register bounded private bridge methods and load owned Pi/OpenCode wrapper
   assets; suppress native task only with operational attested replacement.
3. Register Codex v1 spec only on new native threads, persist manifest with first
   native identity atomically, handle dynamic calls/results and preserve legacy.
4. Finish per-alias delivery/replay rounds against A's guarded transition APIs;
   report minimal storage obstacles rather than bypass validators.
5. Integrate queue/steering/Stop/Send now, live owner YOLO and exact child permission/
   question routing; public children must remain read-only.
6. Add real read-only endpoints and private bounded output-read tools for C.
7. Wire bounded restart recovery through `runDelegationRecoveryBarrier` before
   scheduling/readiness and before owner graph recovery could accept stale evidence.
8. Resolve the requested thin `mcp_client.go` runtime-boundary ownership with the
   manager. No edit was made there while its Library ownership was ambiguous.
   Child MCP coverage currently fails closed rather than inventing settlement.

No CP2a final report or graph activation is implied. Resume only with the manager's
explicit direction following this user interruption.
