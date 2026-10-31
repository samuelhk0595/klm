# A CP2b — global graph agents and frozen specialist catalog

Status: implemented for manager review. No graph child execution, Choice/drain or
new delegation tool registration is activated here. Schema remains version 3; v2
below is the **graph artifact format**, not another DB or snapshot store.

## Available backend behavior / C handoff

- Project graph Save validates the complete visible global + local agent view.
  Conflicting/invalid sources are not silently shadowed. Graphs remain project YAML.
- The existing project authoring response already carries origin metadata and now
  matches actual Save/new capture behavior for a global agent reference.
- Graph catalog initial-node name/description resolves visible globals as well as
  locals. Graph selection URLs/IDs remain unchanged; no extra UI setting exists.
- C can remove its temporary global-agent graph-picker filter after the manager
  approves this backend slice. Selecting an enabled global row now saves and
  captures normally. Real UI/provider/graph execution acceptance remains separate.

## Format compatibility

- Absent Version / explicit version 1 retain old source/hash/validation rules.
  Legacy snapshots are not relabelled, rewritten, adopted or live-recompiled.
- New SQL run capture creates `GraphSnapshot.Version=2`, `ProjectID`,
  `AgentOrigins`, `GlobalSources` and `DelegationCatalog` additively.
- Node `Agents`, `EffectiveAgents`, `Capabilities` remain the existing node-only
  execution contract. Delegation candidates do not become extra graph nodes.
- `Sources` retains only generated project-relative YAML/TOML paths, so existing
  scheduler/workspace mechanisms never receive an arbitrary/fake global path.
- Global node source bytes are canonical AgentDefinition TOML in `GlobalSources`,
  keyed by slug, with scope/UID in AgentOrigins. Their logical origin is
  `global/agents/<slug>.toml`, never a path opened against a project.
- Local sources retain exact raw TOML, including a marker if one already exists.
  Graph reading never adopts/writes one; markerless local provenance is captured
  for this run only and is not a new live session binding.
- V2 revision seals project/source/origin manifest plus the frozen catalog digest.
  Each captured capability retains its existing content hash/adapter checks.
  Offline validation checks definitions against captured text, canonical globals,
  local project/marker identity and configuration/capability seals, without I/O.

`revalidateGraphCapabilities` now includes both node and specialist captures and
rechecks the visible catalog revision/global canonical sources before reservation.
The existing scheduler's project-relative Sources loop can stay unchanged. This
is the same actual-start boundary, not a scheduling-time or child-start live read.

## Frozen catalog

```go
type GraphFrozenAgent struct {
    Definition AgentDefinition
    Source AgentSourceRef
    SourceRevision, SourceLocator, SourceText string
    Execution AgentExecutionCapture
}
type GraphFrozenAgentCatalog struct {
    Agents map[string]GraphFrozenAgent // public slug key, exact UID in each entry
    Complete bool
    Diagnostics []AgentDiagnostic
    DiagnosticsTruncated bool
    CatalogRevision, Digest string
}
```

The entire eligible **visible** catalog is captured, including agents not referenced
by nodes. Model/default validation and external skill/MCP captures happen once at
run start. Candidates with disabled/conflicting/invalid/unavailable configuration
produce frozen diagnostics and are not substituted into the eligible map. Referenced
nodes still fail strict validation/capture. Unknown roots remain incomplete, not
silently absent or deletions. Diagnostics max 100 with an explicit truncation flag.

At most **256 eligible definitions and 16 MiB serialized frozen-entry content**
(definition, raw/canonical source, provenance, resolved config and capability
metadata) are accepted. Content-addressed skill packages remain referenced rather
than inlined. Eligible overflow returns `delegation_catalog_too_large`; eligible
entries are never silently truncated. No configuration knob was added.

## Snapshot-only helpers for B

```go
lookupFrozenGraphAgent(snapshot GraphSnapshot, agent, sourceUID string) (GraphFrozenAgent, error)
composeFrozenGraphAgent(snapshot GraphSnapshot, agent, sourceUID, parentNodeID,
    targetHarness string, overrides AgentConfigOverrides,
    inheritParent bool) (AgentExecutionCapture, error)
```

Lookup accepts an exact name/slug, optional exact source UID, validates the v2
catalog and returns a deep copy. V1 reports `delegation_catalog_unavailable`;
missing/ineligible candidate reports `agent_not_captured`; UID mismatch reports
`source_identity_mismatch`. There is no live catalog/disk/provider fallback.

Composition uses that specialist capture and, only when the caller supplies an
explicitly authorized true inheritance flag, the actual parent's captured node
capabilities from **this run snapshot**. Caller must prove actual node/activation
ownership and human inheritance intent; a boolean/saved definition is not proof.

- Per-field overrides use A's single pure composer, including explicit empty/reset
  semantics. B must still validate an explicit target model/effort/harness/platform
  before child admission; the helper does no provider discovery or execution.
- Captured asset union uses only selected source/parent assets/contracts. Internal
  skills are rejected; differing same-ID skill/MCP contracts fail rather than pick
  a newer one. Missing skill/connection/tool reports `capability_not_captured`.
- Selection-removal overrides work; an override cannot import a new uncaptured
  skill/tool/connection from another candidate or the live Library. No administrative
  role/tools/tokens are inherited. Target adapter metadata can be resealed for the
  chosen managed harness using the same assets, without another Library capture.
- Returned Config + Source + ExternalCapabilities form the immutable child capture
  B should persist/use for the invocation. Owned run/workspace/parent and existing
  Choice/Job/MCP finality gates remain B's runtime responsibility.

Graph snapshot reference projection includes specialist external captures in the
existing capture/credential indexes. Current Library changes cannot rewrite an
active artifact; active-run credential retention still uses the existing graph
snapshot authority. No second artifact store or runner was introduced.

## Files / ownership

A edits in this checkpoint: `graph_snapshot.go`, `graph_contract.go`,
`graph_tools.go` (catalog lookup only), `authoring.go` (Save agent view only),
`sqlite_library_references.go` (new captured references), plus new
`graph_snapshot_v2.go`, `graph_frozen_agents.go`, `graph_frozen_capabilities.go`,
`graph_frozen_agents_probe_test.go`.

No changes to graph_adapter, graph execution/transitions/schedulers/worktrees,
runtime/main, mcp_client, bridge/API registration, agent_tools, B files or frontend.
B can consume the helpers after manager review; no graph tool was registered by A.

## Focused evidence

```text
go test -tags dev -run '^(TestGraphFrozenGlobalSaveCaptureAndLookup|TestAgentBindingLocalMarkerSnapshotValidation)$' -count=1 -timeout 90s .
workdir engine; tool timeout 120 seconds; exit 0; ok klm/engine 3.093s
```

The central temporary fixture saves a global-node graph through the actual Save
handler, compiles/captures v2, freezes non-node eligible specialists, diagnoses an
unusable model without breaking the valid node, preserves markerless source bytes,
rejects missing/wrong-UID lookups, merges parent-captured MCP metadata after a live
connection edit, preserves old prompt after live source edits, rejects changed
catalog at reservation revalidation, and recompiles offline without mutating the
sealed artifact. The second test retains v1-format marked-source/hash validation.

Gofmt succeeded. Initial compile-only name shadow (`bytes` counter vs bytes package)
was corrected; the focused selection above is the final result. `git diff --check`
on the changed tracked files exited 0 with only LF/CRLF notices (10-second timeout).
No full suite/redundant build, graph invocation, real provider/harness execution,
foreground server, installed data, package installation, commit or push occurred.
Fixture provider catalogs and pin metadata are controlled test data, not native
acceptance evidence. Manager review and final runtime/Playwright gates remain.
