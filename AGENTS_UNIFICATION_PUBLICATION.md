# Agents — stacked publication integration

This publication is stacked above PR #40 (`feat/desktop-multi-engine`), itself
above PR #39 (`fix/harness-large-events`). Neither predecessor branch was changed.

- Source recovery: `delivery/agents-source`, commit `0943e8c`, based on existing
  harness/Linux commit `95fbd40`.
- Integrated prerequisite: `feat/harness-linux-prerequisite`, commit `46c7ab5`,
  reproduces `95fbd40` above #40 with its protocol-content and admission lifetimes
  both retained in `interactive.go`/`main.go`.
- Agents delivery: `feat/reusable-agents`, based on that prerequisite branch.
- Publication worktree:
  `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-agents-publish-20261031`.
  Shared source checkout and unrelated dirty/untracked work were preserved.

## Necessary schema integration

Original agents delivery/checkpoint reports describe its independently developed
catalog schema2 and delegation schema3. That numbering was never published in
this stack and is **historical pre-integration evidence**, not the final migration.
PR #39 already reserves schema2 for immutable content generations.

The manager-approved publication chain is:

1. Schema1 → **2**: existing content-generation migration from #39.
2. Schema2 → **3**: additive global-agent/source/catalog-intent/binding domains.
3. Schema3 → **4**: additive owned-child batch/call/invocation/child domains.

Fresh stores use schema4. Existing cutover authority is checked before upgrading.
Read-only table-layout checks precede admission/upgrades, preserve supported content
v2/catalog-v3/delegation-v4 layouts, and reject unpublished standalone agent-only
v2/v3 databases clearly without changing their version/data. They need an explicit
migration; no deletion, JSON fallback, automatic reinterpretation or downgrade.
Content-generation migration accepts valid higher supported layouts without
resetting their version. Existing histories, native identities and v1 graph
artifacts are not rewritten by these additive steps.

## Conflict integration

- Retained large-event protocol scope/cache/status finalization and owned-child
  admission/join/drain/delivery together; content-bound interactive requests still
  reject unsupported oversized approval scopes before delegation dispatch.
- Retained multi-engine identity isolation, protected draft persistence, captured
  edit revisions, local-activity/flush guards, and ACK-controlled composer clearing.
  Saved-agent selection/overrides and bound settings use the same draft store;
  explicit discard/reload clears those drafts before refreshing the exact source.
- Preserved both existing content-publication and new delegation transaction/error
  hooks. No new provider loop, graph execution, installed-data operation or UI redesign.

## Checks actually performed for this integration

- Gofmt on changed/conflict-resolved Go files: exit0, bounded90s command including
  the following focused test target (not a full suite).
- `go test -C engine -tags dev -run '^(TestPublicationSchemaChain|TestPublicationRejectStandalone)$' -count=1 -timeout 60s -v .`
  — **PASS**, package7.035s: four supported fresh/upgrade/reopen cases and two
  standalone-layout rejection cases. Rejection checks version retention and exact
  database-byte SHA-256 across refused open; supported upgrades retain a sentinel.
  This command also compiled the integrated Go package and its tests.
- `clients/desktop/node_modules/.bin/tsc.cmd --noEmit --pretty false --project tsconfig.json`
  — exit0,120s tool bound, integrated publication worktree.
- Frontend dependencies were reused through an ignored worktree-only junction to
  the source checkout's existing node_modules. No package/lock change or install.

CP-V1/CP-V2 Playwright and real Pi/OpenCode/new-Codex adapter evidence was measured
**before this stacked integration**, against a LOCAL simulated model provider.
It is not a post-conflict browser rerun, upstream-model proof or native graph
finality acceptance. The prior report's remaining platform/provider limits apply.
No Playwright/provider matrix or full test suite was repeated for publication.
