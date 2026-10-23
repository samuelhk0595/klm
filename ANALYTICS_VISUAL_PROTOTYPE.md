# Analytics and KLM Agent visual prototype

## Scope and status

The visual layout is approved. Analytics now reads the engine's compact
`GET /api/analytics` snapshot. Execution collection is prospective and durable;
historical creation and valid graph-run facts remain separate from unavailable
historical execution/token measurements. Deterministic Analytics fixtures and the
temporary development Home sample posts have been removed. Home and independent
Focus feeds use real retained posts only; Happening now remains live.

Implementation details, adapter/coverage limitations, exact checks and pending
originating-agent review/runtime acceptance are recorded in
`ANALYTICS_IMPLEMENTATION_PROGRESS.md`. Compilation is not runtime acceptance.

## Entry and controls

Open the rail's KLM logo, then **Analytics** in the left navigation. The menu starts
at 240px and can be resized from its right edge (160–360px, constrained by the
available window). Its width is remembered per browser origin and engine. Double-click
the divider to reset; focused dividers accept Left/Right and Home/End. Narrow layouts
use horizontal Home / Analytics navigation without resizing or replacing the saved width. Overview, Topics and Graphs share the
period and project filters. Previous/next arrows and the selector browse 13
Monday–Sunday local weeks. All time covers all retained evidence, including dates
before the 13-week calendar. Comparison is optional; the current week compares
through the same local weekday/time in the previous week. Percentages are withheld
when either metric is incomplete/unavailable or the previous value is zero.

Chart points support hover, focus and selection; its metric selector switches
between agent-hours and tokens. Calendar cells select their week. The calendar
always shows the 13-week history for the selected project scope, independent of
the selected week. Project choices use the real nonremoved registered inventory.
Archived history is included; the general agent's own work is outside project scope.

## Observed accounting semantics

- All totals, charts, rankings and tables derive from one immutable engine snapshot.
  Sessions used, agents, native subagents and graphs use period-wide identity sets.
  Sessions created come from retained inventory, including unused ordinary sessions.
- Agents executed includes main, side and graph-node agents. Native subagents are
  shown separately. Unsupported native-child time/usage is unavailable, not inferred
  from task duration or added to parent effort. Technical coverage remains in the API.
- Observed execution excludes idle-between-turns and proven blocked waits.
  Nonserial waits without exclusive-blocked evidence are unclassified and expose
  partial time coverage. The blocked-wait figure is separate; retained runtime idle
  is never counted as effort.
- Graph-run hours sum each run's wall-clock duration, not its agents' durations.
  Graph-node execution already belongs to agent-hours. These two time measures
  are never added into a combined effort total. Activity is not a success score.
- Topic grouping uses project-scoped IDs and current owner-session membership.
  Side, node and supported child evidence follows explicit owner links. Equal names
  remain distinct. Historical topic-movement attribution is not reconstructed.
- Missing measurements display as `—`; available values have no Partial labels or
  coverage-only hover/accessibility warnings, per the user's presentation decision.
  Live indicators and internal coverage-based comparison safeguards remain.
  Calendar cells distinguish unavailable history, observed zero and future dates.
  Tokens are normalized native reported deltas attributed to engine receipt time.

## Chat and Home integration

`KlmAgentWorkspace.tsx` provides the navigation and bottom-docked chat. The dock
starts minimized. Message KLM Agent or the bottom launcher opens it; the header
supports expand/restore, minimize, close and the existing Focus floating action.
On wide Home layouts, the minimized/reopen control aligns with the Happening now
column and touches the page bottom. The expanded dock is anchored to the same
centered layout. Close and Minimize both show the same minimized bar.
Minimize and close hide the mounted chat;
they do not call Stop or discard drafts. Home and Analytics share that same dock.

The supplied App conversation/composer, queue, Stop, permission/question controls,
harness picker and drafts remain in use. Desktop Happening now still uses the
existing SessionOverlay and its composer routing, now inside the dock. Mobile
project-session sheets remain independent, as do Focus widgets/floating chats.
Project navigation, session rails and mobile host bridges retain their behavior.
Home fixes the contextual column on the right; its resize remains.
On wide windows, Home centers the entire navigation/timeline/context composition.
Its maximum width is the current left menu + 840px timeline + preferred right
column. The feed header contains search in place of the former Activity title and
subtitle, alongside refresh/settings. The right column starts with Happening now;
mobile keeps its existing search/host controls. Posts stay at a maximum of 800px; the remaining timeline space provides
up to 20px on each side, with excess window space outside the menus. Side-only
post shadows appear only when the timeline's available content width exceeds
800px. At or below 800px posts fill that width without added outer gutters or
shadows. Analytics keeps its full available layout, and narrow/mobile navigation
retains its horizontal presentation.

## Validation

The original prototype passed `npx tsc --noEmit`. Current implementation commands
and results are recorded in `ANALYTICS_IMPLEMENTATION_PROGRESS.md`.

Manual acceptance:

1. Open KLM Agent → Home; check the real feed, search, categories, Details,
   refresh, bookmarks, Happening now ordering and Tasks entry.
2. Open Analytics; exercise tabs, weeks, comparison, chart metric, project filter
   and calendar. Check the partial-week and All time presentations and table totals.
3. Draft a message, minimize/close/reopen, and switch Home ↔ Analytics. Verify the
   draft survives. Exercise existing send, queue, Stop and pending request controls
   with a real harness when desired; minimize/close must leave work running.
4. Open a Happening now session; check overlay draft isolation, close/backdrop,
   and Open in project. Verify normal project chats and return navigation.
5. Check light/dark theme variants, native desktop, resized browser Focus and a
   narrow/mobile host. Check independent Focus windows/widgets and mobile sheets,
   including Show projects and Back to hosts.

Visual layout is approved. Originating-agent implementation review/corrections
precede the still-pending real harness and interaction acceptance checks.
