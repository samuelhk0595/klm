# Automatic pagination — directed Playwright validation

## Environment and scope

- Chromium 158.0.8127.18 (build 1254), Playwright 1.65.0, headless Windows.
- Actual React components, production CSS and HTTP/SSE client code, served by
  Vite from the working tree. The full `App` layout was also exercised with
  React Strict Mode. Desktop viewport: 1440×900/1000; mobile viewport: 390×844.
- A separate loopback fixture API supplied deterministic histories, metadata,
  pending requests, delays, failures and invalidations. These are **browser UI
  integration checks**, not real-engine/SQLite or native WebView acceptance.
- Fixtures and Playwright/browser dependencies were isolated under
  `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-pagination-playwright`.
  Each bounded Node invocation started the fixture/Vite servers, ran Chromium,
  closed the browser and shut down the servers in `finally`.
- No installed engine/client was stopped, rebuilt, installed or restarted.
  No harness/model execution or real conversation mutation was performed.

## Reproduced product defects and corrections

### P1: compact history could not advance toward older messages

Fixture: 550 persisted events, a long compact tool group followed by a completed
assistant answer. Automatic filling reached the 200-event window. Both the
collapsed group and final answer fitted within the viewport.

Before correction, an upward wheel gesture left the first retained event at
`compact-350`: the page was rejected because its bounded slice evicted the final
answer, which was also visible. `prepareWindowChange` incorrectly protected
**every** visible row instead of the reading anchor. Repeated gestures could
not advance the window.

Correction: protect the first visible reading anchor; permit eviction at the
opposite edge while retaining its measured space. The post-fix browser run
advanced the window from `350–549` to `325–524`. A focused unit regression also
checks that a delayed page is still rejected when the user actually moved to
the answer that would be evicted.

### P2: automation inventory drained without scrolling

With the real `.settings-panel` sizing and `.automation-content` scroller, the
inventory's sentinel was outside the scrollable content. The browser fetched
all 65 fixture automations while the list remained at its top.

Correction: move the inventory's `AutoPage` inside `.automation-content`, after
the rows. Post-fix, the initial count stayed at 25 until scrolling. A failed
second-page request remained blocked across the 15-second poll and resumed only
after explicit Retry.

## Directed browser results

| Scenario | Result |
| --- | --- |
| Two readers of one 550-event chat; 12 older pages; bounded eviction and forward traversal | Passed. One reader traversed to `225–424` and back to the real tail; the other retained its original window. |
| Actual `App` layout with Strict Mode, 10 older pages | Passed. Reached and crossed the 200-event bound. |
| 550 consecutive tool events, expanded work's 126px inner scroller | Passed. Ten additional older-page gestures advanced the inner window beyond the initial 200; it remained bounded. |
| Collapsed compact work plus final answer, upward wheel | Failed before the anchor correction; passed afterward. |
| Chat at 390×844, 12 older pages and forward traversal | Passed. Same bounded, bidirectional traversal in the mobile viewport. This does not establish native touch/WebView behavior. |
| Library list: initial 25, scroll to next page, delayed old page during ` BETA ` search | Passed. No initial catalog drain; stale result did not replace the new query; normalized search paged to all 40 matching skills. |
| Large event: explicit Open, own scroller, 64KiB chunks, injected failure and Retry | Passed. All 16,000 UTF-8 lines matched the original exactly, including accents, CJK and emoji; no duplicate/missing content and no automatic retry loop. |
| Graph events: older/forward traversal, scrollbar direction and 200-event bound | Passed. |
| Graph picker: all 55 activations, choose an early activation after loading later pages | Passed. All catalog options remained after selection. |
| Graph: delayed older response, move to the opposite edge before it returns | Passed. The currently read tail anchor was retained. |
| Queue, permission and graph requests resolved on the second page through SSE invalidation | Passed. Removed IDs disappeared while the loaded window remained. Question cards were also rendered/paged, but external question resolution was not separately asserted. |
| Automation inventory and run history: fail offset 25, wait 16s without Retry, then Retry | Passed after fixing the inventory sentinel. Neither reader retried through polling; explicit Retry loaded the next page. |

The initial fixture run also had two **fixture/assertion issues**, corrected
before rerunning the affected scenarios: content request counting accidentally
included session metadata reads, and synthetic native questions initially used
the wrong field name. Queue traversal was adjusted to scroll its real nested
`.message-queue` rather than only the surrounding panel. These were not reported
as additional product defects.

## Artifacts and checks

Temporary scripts: `run-once.mjs`, `server.mjs`, `harness.tsx`, `chat.mjs`,
`work.mjs`, `app.mjs`, `compact.mjs` and `cases.mjs`. Screenshots and bounded JSON
results are in the same scratch directory. Historical failed-run JSON/screenshots
are retained; targeted rerun results document the corrections.

- Final `npx tsc -b`: passed.
- Targeted Vitest: **19 tests passed**, including the two new anchor regressions.
- Scoped `git diff --check`: passed. Fixture ports 5173/17331 were closed after
  validation; the pre-existing 7331/7332 listeners retained their original owners.
- No full test suite or additional independent review was run.

Remaining acceptance: authenticated real-engine/SQLite end-to-end traffic,
actual installed Tauri/WebView2 and Flutter WebView, real touch gestures, live
harness streaming and the rest of the original surface acceptance matrix.
The source corrections do not hot-update the already installed frontend bundle.

## Astra follow-up: A1–A3 corrections

The same reviewer subsequently reproduced three further defects in Chromium:
compact prefetch disabled live-tail following (A1), graph work groups lacked
internal graph paging (A2), and hidden mounted readers retained their own history
subscriptions (A3). These defects were corrected in the working tree.

- **A1:** older prefetch preserves following when the reader was following and
  the merged window still contains the real tail. Actual backward scrolling and
  eviction of newer events still leave following disabled.
- **A2:** `GraphEventPage` projects before/after edges into work-group slots through
  `HistoryPagingContext`. Each edge observes its real outer/inner scroller, uses
  graph offsets and calls the graph activity reader. It never uses the ordinary
  session-history endpoint. Group and outer paging share the same bounded window,
  request admission and geometry restoration.
- **A3:** `useHistoryVisibility` observes the root and ancestor hidden/class/style
  changes, layout size and document visibility. Hidden readers unsubscribe,
  abort refresh/paging reads and retain their window. Reopening revalidates that
  window and rearms automatic page detection. A retained refresh error still
  needs explicit Retry; hiding/showing does not silently release its error latch.

The new bounded `followup.mjs` browser run passed all three scenarios:

| Scenario | Post-fix result |
| --- | --- |
| Open 550-event compact chat; allow automatic 200-event fill; append event 550 with SSE | New assistant event appeared without a navigation gesture. After deliberate older navigation, a further append did not pull the reader back to the tail. |
| Transform graph activity events into tools, expand the 126px inner scroller, traverse nine older pages and return to graph-549 | Both internal graph edges were present, the retained group stayed at 200 events and no session-history endpoint was requested. |
| Hide the 50-event first reader while the second remains visible; invalidate the chat | No `after=499`/`after=524` reads from the hidden window. |
| Reopen with a delayed first page, hide through a CSS class while the read is pending, then reopen again | The hidden reader did not continue to its second page. Reopening revalidated both pages and retained the 50-event window. |

The existing directed graph scenario also passed after the paging refactor:
outer scrollbar traversal, 55-activation picker retention and delayed-response
anchor protection. Its first rerun exceeded the fixture's six-second cold Vite
navigation allowance while unit tests ran concurrently. The navigation allowance
was separated from interaction timeouts and set to 20 seconds; the standalone
rerun passed. Each invocation retained the 90-second overall deadline.

Final checks for this follow-up: `npx tsc -b` passed; **20 targeted unit tests
passed** in the same five focused files (including the added automatic-tail
regression). Scratch artifacts include `browser-followup-results.json` and the
`A1-compact-live-tail`, `A2-graph-inner-paging`, `A3-hidden-reader-pauses` screenshots.
Native clients, real-engine traffic and actual installed-app deployment remain
outside these fixture-based checks.
