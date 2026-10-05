# Tasks — configuration and generic webhooks

`TASKS_IMPLEMENTATION_PLAN.md` is the approved product contract. Its section 8
decisions remain open unless explicitly resolved by the user.

## Implemented

- Project-scoped Task definitions in the engine's existing state transaction and
  journal storage, with stable IDs, revision checks, creation operation IDs and
  deletion tombstones. No fixtures or browser-local definitions are loaded.
- Tasks below New session; clicking opens Details by default. Edit replaces the
  details dialog; Save/Cancel returns. Enable/Disable/Delete remains in the header
  overflow on either tab. Deletion retains confirmation and displays API errors.
- The sectioned Task / Orchestrator / Trigger editor keeps one draft across
  sections and an overlaid Settings dialog. Failed saves retain the draft. A
  conflicting save does not overwrite the latest engine revision: preserve any
  needed draft text, cancel and reopen to use the current definition.
- Real per-project harness model/effort discovery and save-time model validation.
  Missing models are shown and require an explicit replacement. Missing/disabled
  saved graphs remain visible and removable; adding graphs requires an enabled
  catalog entry. Existing missing folder references are retained, not remapped.
- Required trimmed name/instructions, optional description, case-insensitive
  duplicate-name checks, and 80/240/12000 **UTF-16 code-unit** limits shared with
  the API. Weekdays are unique and time is minute-resolution 24-hour `HH:mm`.
- New definitions are always enabled. Telegram routing, parallelism and missed-run
  recovery default off. The default destination is `Tasks`; saving creates no
  folder, session or execution. `Ungrouped` retains its existing meaning.
- Save records an authorization ID associated with the revision's instructions
  and selected graphs. The explicit authorization copy is retained. Empty graph
  selection grants none. Task runs consume a distinct captured Task grant;
  no human event is fabricated and ordinary chat keeps its user-event contract.
- Typed API requests live in `src/engine.ts`. Existing multiplexed update SSE emits
  `tasks_changed` invalidations; the mounted Tasks view reloads shared definitions.
  Connection failures are shown as errors, not as an empty inventory.

## Runtime boundary

Generic webhook Tasks use real engine admission, normal sessions, captured Task
settings/grants/input, FIFO/captured parallel policy, explicit completion, Needs
attention, cancellation/settlement and restart interruption. Details provides Create
webhook and a copyable engine endpoint without requiring a repository/public origin.
Optional public origin and signature/delivery headers can be edited later.
Run now accepts generic text/JSON/form input, including for disabled Tasks, with
manual origin and no binding prerequisite. Event/provider interpretation belongs to
saved instructions and graphs. A Task can finish without a graph; an empty allowlist
still authorizes none. Scheduling/catch-up and Telegram delivery remain unimplemented.
Setup and pending human acceptance are in `TASKS_WEBHOOK_HANDOFF.md` at the root.

The run presentation helper retains the latest five non-queued entries plus every
older waiting entry, without duplicates; queued entries remain separate/FIFO.
`TaskRun` now uses the real engine projection, with captured payload/metadata, separate
execution/user-wait metrics, reports and normal session identity. `TaskSessionChat.tsx`
remains removed; normal chat's existing between-turn YOLO control is untouched.

The engine refuses deletion while work is queued/active/waiting/awaiting attention.
Allowed deletion retains tombstones, runs and ordinary sessions/backlinks.

Telegram simulation is no longer mounted. Its Settings section shows Not connected;
the unused `TelegramSettings.tsx` layout awaits real transport integration. No fake
bot credential, pairing success, scheduler or catch-up is active. Only the focused
authenticated generic webhook ingress is implemented; live deployment has not been checked.

## Engine timezone

Settings → Engine now reads/writes a persisted engine-global IANA timezone with
revision checks and live SSE synchronization. Initialization detects the engine
OS zone once (Windows WinRT; supported Unix system sources), then persists it.
Failed detection is recorded as unconfigured and requires explicit selection.
Neither OS changes nor the browser's timezone silently replace the saved value.
The engine embeds tzdata for installed Windows machines without a Go SDK.

Approved change policy: reformat historical dates without changing stored instants;
preserve accepted runs/snapshots; future scheduling uses the new zone, with no
catch-up generated solely by the change. The future scheduler must skip nonexistent
DST times and use only the first occurrence of repeated times. Scheduling itself
is not implemented. Shared Task and quota date formatting uses the saved zone;
an unavailable/unconfigured zone shows explicit ISO instants instead of device time.

## Storage and API

Implemented routes:

- `GET /api/projects/{id}/tasks`
- `POST /api/projects/{id}/tasks` — `{task, operationId}`
- `PATCH /api/projects/{id}/tasks/{taskId}` — `{task, revision}`
- `PATCH /api/projects/{id}/tasks/{taskId}/enabled` — `{enabled, revision}`
- `DELETE /api/projects/{id}/tasks/{taskId}` — `{revision}`
- `GET /api/projects/{id}/tasks/{taskId}` — retained definition/history identity
- `GET/POST /api/projects/{id}/tasks/{taskId}/webhook` — nonsecret endpoint settings
- `POST /api/projects/{id}/tasks/{taskId}/webhook/secret` — explicit secret rotation
- `GET/POST /api/projects/{id}/tasks/{taskId}/runs` — bounded history/manual payload
- `POST /api/projects/{id}/tasks/{taskId}/runs/{runId}/cancel` — separate Cancel run
- `POST /hooks/tasks/{bindingId}` — only this route belongs at public HTTPS ingress
- `GET /api/settings/timezone`
- `PATCH /api/settings/timezone` — `{timezone, revision}`

Startup reads versions 1/2 and replays the existing journal before the additive
Tasks migration. It writes a fully replayed `state.pre-tasks-<id>.json` recovery
copy, then adds `taskSchema: 1` to version 2. Strict older readers refuse the new
field rather than discard Tasks. The backup is a checkpoint, not a live second
store. A rollback requires a stopped engine and the matching checkpoint/journal
recovery procedure; do not merge an older checkpoint with newer journal entries.

## Manual acceptance (not yet performed)

Focused checks completed: `go build -tags dev` with its executable written outside
the repository, and `npx tsc --noEmit -p tsconfig.json` in `clients/desktop`. No test
suite or runtime/real-data migration check was performed.

1. Start the updated development engine and frontend. Open Tasks; existing projects
   have no sample definitions or task sessions.
2. Create a manual task, choose a real harness/model/effort and save. Confirm enabled,
   default-off policies and no new `Tasks` sidebar folder. Reload and restart the
   engine: the definition should remain.
3. Create scheduled/webhook definitions. Check weekday/time persistence and that
   saving/enabling creates no run. No next firing is shown yet.
4. Edit fields in all sections, open/close Telegram Settings, and return to the
   unchanged draft. Save and Cancel return to Details; Details is the initial tab.
5. Open two clients. Edit/disable in one and observe the second update. Keep an
   editor open in both, save one, then save the stale draft: expect a conflict and
   retained draft rather than an overwrite. Test unavailable models/catalog errors.
6. Exercise duplicate names, blank name/instructions and length boundaries through
   UI and API. Replay the same creation operation: only one definition is created.
7. Cancel deletion, then confirm it. Reload to confirm removal; repeated/stale
   mutations must not restore or overwrite the definition.
8. Confirm normal engine-backed chat, folder and session controls remain available.
9. Settings → Engine should initially show the engine computer's timezone. Change
   it and observe a second client's Settings and quota timestamp tooltips update.
   Restart and verify persistence. Changing the computer's OS zone later must not
   overwrite it. Failed initialization must require an explicit selection.

## Remaining gates

Finish/settlement and missing-finish recovery; whole-task/queued cancellation;
Task grant integration/revocation/graph changes; fresh/reuse corrections;
concurrency transitions/resource and shared-write policy; catch-up revision and
eligibility reconciliation; partial-branch user-wait timing; Telegram pairing,
unsupported forms/replacement/provenance; webhook manual input/authentication/
ingress; archive, project removal and deleted-task navigation.

Telegram secrets remain explicitly permitted for the first delivery version.
Token protection, verified identity and ordinary KLM private-answer masking remain
required; confidential transport/redaction/retention improvements stay deferred.
