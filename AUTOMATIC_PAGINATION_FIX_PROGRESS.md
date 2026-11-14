# Automatic pagination restoration

## Loading indicator flicker follow-up — 2026-11-11

Streaming invalidations were being presented as pagination progress. In particular,
`DelegationRequests`, mounted immediately above the composer, rendered Loading on
every child-request refresh even with an empty result. Both sidebar pagination
footers also rendered Loading from the same project-wide busy flag, including when
there were no more pages.

- Separated request exclusion (`busy`) from visible loading feedback in AutoPage.
- Kept child-request, pending-request, queue and graph-request revalidation silent;
  actual page extension and explicit Retry retain feedback and error recovery.
- Sidebar feedback now identifies initial/session/topic/retry reads, so background
  invalidations do not add rows and only the appropriate footer shows progress.
  Confirmed empty-topic state no longer disappears during background reads.
- Live-tail history catch-up is silent; ordinary history pagination retains progress.
  Existing history-window preservation and its 200-event limit were retained.
- Checks passed: TypeScript `--noEmit`, scoped `git diff --check`, and 13 focused tests
  across AutoPage, DelegationRequests, useRequestWindow, ConversationHistory paging
  and history-window geometry. A failing new loading-state assertion was corrected
  by capturing the feedback flag before scheduling React's state updater.

Real streaming/scroll/layout validation in the updated desktop remains pending.
This frontend correction was not installed or deployed in this follow-up.

## Plan and scope

Implemented from `AUTOMATIC_PAGINATION_FIX_PLAN.md`, authored by the saved Plan
Writer child. The initial child execution was interrupted by an engine restart;
the parent recovered the recorded child and resent the complete context before
receiving the completed plan. No independent conversation was created.

The earlier history observer was introduced by `d3a38c3` and removed in `a255836`.
The correction uses current SQLite cursors rather than reverting those files.

## Implemented

- Shared near-edge detection, 300 px margin, real scroll-root discovery, resize
  and visibility wakeup, cursor-attempt suppression and error recovery. The
  observer only acts on visible surfaces. A failed cursor retains Retry even if
  background revalidation clears the feature's transient error.
- Explicit 25-item requests across the converted lists. History items are
  persisted events; grouping and response byte limits are preserved.
- Chat before/after cursors, directional contiguous/overlapping merge, per-event
  revision reconciliation and the existing 200-event bound. Revalidation reads
  the retained range in bounded pages instead of `startIndex + 50`.
- Reading-anchor markers and restoration; following streaming requires the real
  tail. Shared conversation reads are serialized between mounted viewers.
  Main, general, side, Focus/mobile and subagent views use the shared behavior.
- Session/topic metadata accumulation, removal of four-page rotation, retained
  boundary revalidation, sidebar anchor preservation and incomplete-empty-state
  handling. Collapsed groups do not initiate session pagination.
- Happening now, feed and grouped activity use near-edge paging. Feed boundary,
  retention, bookmarks and explicit grouped-activity Update semantics remain in
  their existing feature-owned readers.
- Graph activation picker and bidirectional activity, Tasks columns, webhook
  inventory, automation inventory/run history, topic/session/agent pickers,
  agent references, Library skills/deletion references, queues and request pages.
- Paged native select use cases now use SearchSelect with feature-owned search
  and pagination footers. Generic UI primitives do not own endpoint behavior.
- Go handlers accept bounded page limits for projected graph requests,
  activations/events, automation runs, pending items and delegated requests.
  Existing defaults/maxima and response byte protection remain; session detail
  requests may request a 25-item pending prefix without changing mutation defaults.
- Existing grouped-activity test fixtures no longer depend on production Load
  more buttons. The history merge test includes bounded bidirectional traversal
  and rejection of a disjoint page.

The restored compact subagent row remains the rendered projection. Pagination
does not load unopened child transcripts. Single-event content retains its separate
byte/chunk contract. Show more/Show less presentation controls and analytics period
navigation retain their own semantics.

## Follow-up implementation after `1766fb1`

The current branch already contained the automatic list and bidirectional history
implementation when this follow-up started. The surface inventory was checked
against that code rather than replaced. Existing unrelated local work was preserved.

- Increased the shared prefetch margin to 300 px. Nested roots are intersected
  with their surrounding clipping panels and the viewport before paging. Ancestor
  scroll, resize and visibility/layout changes wake the detector; hidden surfaces
  rearm when shown again.
- Added history-window geometry capture for mounted readers before applying a
  page. Evicted rows leave measured prefix/suffix space, which is consumed on
  reverse traversal. The existing 200-event window and both cursors remain.
  Prefetch pauses rather than evicting visible rows. Restored positions are not
  interpreted as new scrollbar navigation.
- History direction also follows scrollbar movement. Nested content scroll and
  unrelated keyboard actions do not change the outer history's direction. Shared
  history requests cancel when their last mounted reader detaches.
- Work-group identities survive overlapping history pages, preserving their
  expansion state when the first event in a group changes. Anchor fallback can
  find a surviving group/member. Session rows and topic headers now carry stable
  identities for sidebar revalidation anchors.
- Removed Previous part / Next part from large event content. Open content starts
  a separate scroll reader; subsequent server-sized chunks append on proximity.
  Incomplete UTF-8 bytes carry across chunks without lookahead requests. Page
  failures retain the same offset and explicit Retry. Download event stays an
  explicit, independently cancellable action.

### Follow-up checks

- `npx tsc -b` in `clients/desktop`: passed during implementation.
- Final `npx tsc -b` and scoped `git diff --check`: passed.
- No test suite, browser/runtime validation, engine build, installation or restart
  was performed for this frontend-only follow-up.

## Astra review and corrections

The user-requested independent Astra review (session
`7d9b90254200bb41ce00a79553046123`) returned ten findings against HEAD `1766fb1`
and the then-current working tree. It used static inspection, not browser
reproduction or test execution. The following corrections supersede the earlier
shared-window implementation described above:

| Finding | Correction |
| --- | --- |
| F1: one reader vetoes another's 200-event traversal | Each `ConversationHistory` adopts its own initial hydrated page and owns its before/after window and revalidation. Paging never mutates the shared tail or asks another reader to discard its anchor. |
| F2: compact work's inner scroller cannot page beyond 200 | Feature-owned history loaders occupy presentation slots inside the expanded work scroller, only at the outer timeline's boundary groups. They detect wheel, touch, keyboard and scrollbar direction on that scroller; geometry restores inner anchors relative to their containers as well as the outer anchor. Nested consultation projections do not inherit parent paging. |
| F3: child refresh reinstalls an older window | Both per-reader refresh and the modal's source refresh check the requested window again before committing, discard superseded results and schedule the current range. |
| F4: graph scrollbar/Space does not switch direction | Timeline scroll deltas update direction; layout restoration records the resulting position separately. Space and Shift+Space join the existing navigation keys. |
| F5: activation selection discards catalog pages | Changing the selected activation resets only its event window; the accumulated catalog and cursor survive. Changing session/run/node resets the catalog scope. |
| F6: graph paging removes the captured visible row | The proposed bounded slice is checked against currently visible anchors before commit. Unsafe prefetch is deferred; accepted changes preserve measured space and geometry, including expanded work. A later deliberate gesture can request a deferred page again. |
| F7: a late skills page overwrites a new query | Search/engine/inventory generations cancel old loads; the commit also checks exact inventory identity and merges only skills into that same state. Query comparison uses the server's lowercase/trim normalization. |
| F8: Library sentinel outside the real scroller | The picker now renders skills paging after the rows inside `.session-picker-list`. |
| F9: paged requests/queue retain resolved items or reset to page one | A feature-owned request-window reader revalidates the confirmed page count from the beginning on conversation invalidation/revision changes and commits atomically with anchors. Queue, permissions, questions and graph requests use this reconciliation; mutation controls retain exact IDs. Failed extensions do not increase the confirmed count or retry on invalidation. |
| F10: automation polling repeats failed extensions | Automation inventory and run-history readers stop after an error. Polling, editing transitions and queued refreshes cannot reissue the failed page; explicit Retry releases the failure latch. |

### Checks after corrections

- `npx tsc -b` in `clients/desktop`: passed during corrections.
- Targeted Vitest run: **17 tests passed** across
  `ConversationHistory.paging.test.tsx`, `useRequestWindow.test.tsx`,
  `sessionState.test.ts` and `LibrarySkillsPaging.test.tsx`. These cover independent
  readers, stale refresh rejection, mutable request revalidation, failure latching,
  identity cancellation, normalized Library search and the existing bidirectional
  merge contract. The first narrower run had 15 passing tests; the two Library
  regressions were added afterward.
- Final `npx tsc -b` and scoped `git diff --check`: passed. After the final
  initial-page ownership adjustment, the two focused conversation-window
  regressions were rerun and passed.
- No full suite, browser/native runtime check, second independent review, engine
  build, installation or restart was performed for these frontend corrections.

The unit fixtures do not establish native scroll geometry. Manually exercise the
review's ten scenarios, especially work groups longer than 200 events, two visible
readers, delayed graph responses, 50+ activations, second-page resolution in another
client, Library search during a pending page, and an automation pagination failure
left open for more than 15 seconds before Retry.

### Same-reviewer follow-up: A1–A3

The user requested a second review from the same Astra session. It returned two
P1 findings (automatic compact prefetch stopped live-tail following; graph work
groups still lacked internal paging) and one P2 finding (hidden readers retained
their private subscriptions). All three received source corrections:

- Preserve following during automatic fill while the merged window retains the
  real tail, preserving deliberate backward-reading behavior.
- Project feature-owned graph-offset paging into expanded work slots through
  `GraphEventPage` and the existing presentation context; share bounded geometry
  and request admission with the outer activity timeline.
- Observe actual root/ancestor/document visibility, unsubscribe and cancel hidden
  reads, retain the local window and revalidate on return. Retain explicit Retry
  after a refresh failure across visibility transitions.

All three directed Chromium scenarios passed after correction, including a
hidden in-flight cancellation and restoration through a CSS class. The earlier
outer graph/picker/delayed-anchor scenario also passed after the refactor.
`npx tsc -b` and 20 focused unit tests passed. Exact fixture scope and artifacts
are recorded in `AUTOMATIC_PAGINATION_PLAYWRIGHT_VALIDATION.md`; native and
real-engine acceptance are still separate.

## Lightweight checks performed

The checks below belong to the earlier implementation, before this follow-up.

- `npx tsc -b` in `clients/desktop`: passed after correcting two introduced type
  errors (optional graph access and an unreachable abort).
- `go build -tags dev -o "C:\Users\Samuel\AppData\Local\Temp\opencode\klm-pagination-check.exe" .`
  in `engine`: passed. The executable was not started or installed.
- `npx --no-install vitest run src/features/chat/sessionState.test.ts`: passed,
  10 tests. No full test suite or adversarial review was run.
- `npx --no-install vitest run src/features/klm-agent/GithubActivityCard.review.test.tsx -t "keeps page two"`:
  passed, one selected existing test; six other cases skipped. This verifies
  second-page retention and explicit Update with the edge-trigger fixture, not
  browser IntersectionObserver geometry.
- Final `npx tsc -b` and the scoped `git diff --check`: passed.

## Human acceptance still pending

Directed Chromium/Playwright UI validation was subsequently requested and run.
See `AUTOMATIC_PAGINATION_PLAYWRIGHT_VALIDATION.md` for the exact fixture scope,
passing scenarios, the compact-history anchor deadlock and automation sentinel
defects found/fixed, and native/real-engine checks that remain pending. Final
TypeScript compilation and 19 focused unit regressions passed after those fixes.

Use the detailed matrix in the plan. In particular:

1. Scroll chat above and below a 200-event boundary, with grouped tools and
   concurrent streaming; check the visible line stays in place.
2. Display the same chat in main and Focus; change session/project/engine and
   filters while requests are pending.
3. Navigate beyond four pages of sessions and topics, including collapsed and
   archived groups, search and insertions/removals during revalidation.
4. Exercise every converted surface in desktop, Focus and a mobile viewport;
   verify Network requests use 25 and observers belong to the actual scroller.
5. Fail a pagination request while its edge remains visible; confirm no automatic
   retry loop and that Retry resumes. Confirm end-of-list stops requests.
6. Verify a short viewport fills sequentially, while collapsed work groups do not
   drain a hidden transcript. At the history bound further deliberate scrolling
   can advance without automatic opposite-edge ping-pong.
7. Validate graph polling/activation selection, pending request resolution in
   another view, automation session reuse, feed retention and explicit Update.
8. Traverse history in both directions after measured prefix/suffix space exists;
   include scrollbar dragging, expanded/collapsed work and two visible readers.
   Confirm visible-row protection does not strand navigation in the intended layouts.
9. Open large content containing accented text and emoji across byte boundaries.
   Scroll its own region, fail/retry a chunk and change the content digest while
   reading; confirm earlier text stays available and the outer chat stays put.
10. Hide/show or resize nested panels and closed pickers. Confirm invisible roots
    do not page and visible short regions fill sequentially with the 300 px margin.

Compilation and the targeted merge tests are not browser/runtime acceptance.
