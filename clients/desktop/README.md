# KLM Desktop

React 19 + TypeScript + Vite client for the standalone [Go engine](../../engine/README.md).
The original `../../code.html` is preserved. This is a browser client, not a native
desktop wrapper; native folder selection is handled by the local engine.

## Run

Start `go -C engine run .` from the project root. Then, from `clients/desktop`:

```sh
npm install
npm run dev
```

The client calls `http://127.0.0.1:7331` directly. Set `VITE_ENGINE_URL` to override
that endpoint. Production assets can be built with `npm run build` and served with
`npm run preview`. Both require the separate engine for application features.

## Structure

```text
src/
  engine.ts             HTTP client, shared API types, native picker request
  design-system/        Domain-independent primitives, tokens, and shared styles
  features/projects/    Project rail and creation form
  features/workspace/   Sessions sidebar and creation dialog
  features/chat/        Main/side chats, composers, consultation activity, harness events
  features/agents/      Agent catalog and editor with static fixtures
  features/graphs/      Graph list and temporary authoring canvas
  App.tsx               Engine state, SSE subscriptions, and application composition
  DesignSystem.tsx      Interactive gallery with isolated static examples
```

`tokens.css` owns semantic colors, typography, spacing, radii, shadows, and panel
dimensions. Primitive styles are in `design-system/styles.css`; app layout is in
`styles.css`. Import components directly and keep service calls outside generic
UI primitives. Inter fonts and Lucide icons are bundled without CDN dependencies.

## Authoring Prototypes

Agents and Graphs use static fixtures in the main application, not only in the
design-system gallery. The agent catalog, models, harness compatibility and graph
references are simulated. Agent deletion/disable checks use those reference
fixtures; they do not inspect canvas topology. Agent and graph lists are held in
memory per project until reload.

New graph opens a temporary canvas. Existing graph tiles edit their list metadata,
not topology. Apply saves node settings in the current canvas only; there is no
action to create a list record from that canvas or persist YAML/layout. Leaving
the canvas discards it. AI agent settings update immediately; other node panels
use Apply/Cancel. Cancelling configuration of a newly added node leaves its
initial draft on the canvas.

The initial node can be AI agent or Terminal command. The canvas catalog also
includes Choice, Fork and Join. AI agent and Join nodes connect to Choices.
Choice session policy is shown only for an AI agent or Join destination and is
cleared when its outgoing connection or destination is removed. Choice input/output
and Fork output field names use the same normalization as Choice names: lowercase,
spaces converted to underscores, accents folded, other symbols removed, and
edge underscores stripped on blur/apply. Payload values retain their text.
Form and connection checks support visual editing,
not validation of an executable graph: drafts may have unselected agents or
unconnected destinations. Fork output fields currently validate names only;
the variable picker offers run.input.task, without reference resolution or a
payload preview. The Choice preview has its own local validation and interpolation.

Worktree and branch settings describe intended execution; there are no Git
operations, shell command execution, graph runs or harness submissions here.
Integration, synchronization and execution contracts remain under refinement.

The main chat composer shows a graph picker at bottom left, opposite the model
and effort controls. It uses the selected project's graph fixtures/local list.
All graphs are listed; disabled ones cannot be selected. Up to five graph rows
are visible, with search only when the catalog has more than five graphs. None
clears the selection. Selection is kept per chat in browser memory until reload;
renames retain the selected ID, while disabling/deleting a graph clears its
selections in that project. It is not sent to the engine or used by Send yet.
The gallery's Composer example has six graphs to demonstrate search and scrolling.
Model and graph search inputs omit the thick external focus outline.
The graph menu aligns its left edge with the picker and extends rightward to
avoid covering the sessions sidebar. Other menus keep their existing alignment.

Selecting a graph adds a **Graph** tab beside **Chat** below the session title,
without switching tabs automatically. Graph shows a read-only React Flow canvas
using the authoring cards, with pan, zoom and Reset view. Nodes cannot be moved,
connected, removed or edited, and the authoring catalog is not exposed.
The composer and Session Status Bar appear only in Chat, preserving the draft while
Graph is open. None hides Graph and returns to Chat; deleting
or disabling the selected graph also removes the tab. Tab selection is local to
each chat. Changing graphs while viewing Graph loads the new drawing and fits it.

`features/graphs/graph-view-demo.ts` provides distinct fictional drawings keyed by
the three catalog graph IDs. Renaming a graph retains its drawing. These are not
saved builder canvases or executable definitions; New graph still has no save
action.

CSV export opens focused on its first activities. The current visual experiment
replaces the active element's LED border with a vertical pink-to-blue shimmer across
the card body, without tilt. It travels back and forth, taking 4.5 seconds per direction
with ease-in-out motion. The overlay follows
the card's corners and does not intercept clicks or clip connection handles.
Reduced-motion preferences disable the shimmer. The small dot beside Running
blinks in discrete steps: 500ms blue with glow, 500ms gray without glow. This applies
to the card and panel indicators. Reduced-motion preferences keep them steady.
All graph cards share the moderate `--shadow-card` resting shadow, including Choice
and Fork. The animated shimmer remains separate from that shadow. An explicit
initial-node flag displays **Start** in the agent or terminal header, in both the
authoring canvas and catalog drawings; it is not inferred from node position.
While a graph run is active, clicking any agent, Choice, Fork, Join or terminal card
(also keyboard-accessible) opens a floating panel above the canvas controls without resizing or moving nodes.
Close or Escape dismisses it and returns focus to that card. Reset view fits the
whole graph.

The panel has **Run**, **Input** and **Output** tabs in the same 54px header, replacing
the agent subtitle. The body remains 200px high. Run shows concise, non-expandable
logs, following new entries unless the user scrolls up. Input displays received JSON.
Output is blank while pending/running; a completed agent shows its selected Choice
and submitted payload. The Choice has its own input and transformed destination
output, rather than treating those two payloads as interchangeable.

`graph-run-demo.ts` provides fictional snapshots driven locally by `GraphView.tsx`:
Plan CSV export completes at 24 seconds, ready processes its payload, and Implement
export starts at 27 seconds. Planning/Choice results remain available when selecting
other cards and returning. Implementation logs loop while later elements remain
pending with empty tabs. Opening Graph for the first time starts the local CSV
simulation; selecting a graph alone does not. The clock, canvas viewport and inspected
activity remain mounted across Chat/Graph switches, with the canvas hidden in Chat.
A blinking LED beside the Graph tab label follows the preview's running state and
remains visible in Chat. During a run, the graph picker replaces its left graph icon
with the same blinking LED as the nodes; its name is always plain text. Idle/None
selections show the usual graph icon. Shared LED styles live in `graph-running-led.css`
and respect reduced-motion preferences. The text-shimmer experiment was removed;
the card keeps its pastel shimmer in `graph-shimmer.css`.
Changing/clearing the graph selection or leaving the session
view unmounts the preview and clears its running indicator. This is not a complete simulation of Fork/Join or graph
execution contracts. It runs no tools, writes no files and does not affect engine
state or persisted chat history. Timeline tracking and real execution remain pending.

Without an active graph run, clicking a card opens `NodeConfigurationPanel.tsx` in
the same right-side floating position and shell as the authoring panels. It renders
labels/values rather than disabled controls: agent name, selected agent and effective
harness/model/effort with inherited/override provenance; Choice description, outcome,
input fields, output templates, destination and eligible session policy; Fork branch
names, destinations, worktrees, Git branches, base revision and output fields; Join
agent, additional prompt and output Git branch; terminal name and command.
There are no Apply/Cancel, edit, add/remove or preview-editing actions. Close and
Escape return focus to the card. Required/Optional is provisional plain text rather
than a toggle. Bug fix's request_changes and Code review's outcomes include fictional
contract fields to demonstrate this view. CSV export remains the active-run example,
where even pending nodes open the bottom activity panel rather than configuration.

## Behavior

Project creation persists name, absolute directory, and a resized PNG icon through
the engine. The sidebar contains only the selected project's sessions, optionally
grouped into topic folders. Each real session chooses its harness at creation.
Below the project heading/path, compact icon-and-text rows show **New session**,
**Agents** and **Graphs**, before the Sessions list. New session uses a compose icon
and opens the existing creation dialog; it replaces the former large button.
The sidebar heading shows only the project name beside its settings button. The
Harness badge is reused in the empty main chat, below a KLM wordmark and above
“What would you like to work on?”. Both logo rows share the same compact width,
with tightly spaced KLM letters and 2px horizontal padding on the badge.
The composer has independent model and effort menus immediately left of Send.
The harness icon appears beside the model name; there is no lower-left harness label.
Clicking the model opens an anchored search menu with five model rows visible at a
time and scrolling for more. Each row identifies its connected provider. Selecting
a model saves immediately. Clicking effort opens a stepped slider; releasing it
or finishing a keyboard adjustment saves effort without changing the model.
The catalog carries an explicit connected-provider list, enforced by both engine
and picker. Automatically available free catalogs without a configured connection
are excluded. A model offered by OpenRouter remains under OpenRouter even when
its name contains OpenAI, Anthropic, or another model vendor.
Pi/Codex expose thinking
efforts; OpenCode exposes native model variants. Selections apply to the
next turn. Selection is disabled while running, and unsupported effort overrides
reset to Default when switching models. Harness default remains available; the
resolved model is displayed when the harness reports it.

Shared `design-system/Menu.tsx` and `Slider.tsx` primitives are demonstrated in the
Design system gallery. Menus are non-modal, viewport-aware anchored panels, dismissed
by outside clicks or Escape. Slider values are controlled and expose native keyboard
interaction and accessible value text. The model menu also supports arrow-key navigation.

SSE sends full session snapshots and reconnects automatically. A small state refresh
also discovers changes from other clients. The UI renders real user/assistant
messages, collapsible reasoning and command/MCP/tool events, and errors. Reasoning
availability and streaming granularity depend on the chosen harness and model.
Stop cancels the owned harness process. Drafts clear only after an accepted send.

The icon immediately after Session log opens the linked side conversation, replacing
the former details inspector. Side sessions are filtered from the workspace list.
Selecting text in a main-chat message exposes **Ask side agent**; exact passages
remain removable composer context until a question is sent. Subsequent selections
use the same side conversation. Sent context and source references persist in the engine.
Drafts and unsent selections are independent per main chat and last until reload.
The selected project, last selected session per project, and side-panel open state
per main session are saved in browser localStorage under the versioned, engine-scoped
`klm.workspace-navigation.v1:<engine-url>` key. Reloading restores that navigation.
Switching sessions restores only that session's side-panel state; opening a side chat
does not open or create side chats for other sessions. Missing projects/sessions fall
back to the first available entry. Storage failures leave in-memory navigation usable.

The side chat reuses message rendering, composer, model picker, permission/question
cards, and token/context status. Its status bar omits quota, skills, and MCP indicators.
Choose its harness before the first turn. SSE subscriptions include the visible side
and every running session, including hidden side conversations. Consultation output
from the answering agent is grouped under expandable question/answer activities with
lifecycle status and Cancel. The requesting agent's resumed reply appears in the
normal conversation automatically, including when reopening saved history.

Side width defaults to 352px and is remembered in `klm.side-agent.width.v1`. The
divider supports pointer dragging, ArrowLeft/ArrowRight (24px), Home, and End.
It retains at least 304px for the side and 400px for the main chat; when those
columns do not fit, the side occupies the full app panel with a close button and
keyboard focus containment. Closing the panel does not delete its native history.

Permission cards appear above the composer. Session snapshots include optional
`permissions`; decisions POST to `/api/sessions/{id}/permissions/{permissionID}`.
While a request is pending, Working changes to Waiting for approval. Buttons use
the backend's allowed choices and display the actual native grant lifetime.
Errors leave the request visible for recovery, and Stop remains available.

Question cards share that area but are separate from permission prompts. Options
are clickable buttons, multiple selections are available when requested, and
custom text is shown when allowed. Send answer submits all questions in the batch;
Dismiss cancels the request without choosing an answer. Working becomes Waiting
for your answer. Session snapshots include optional `questions`; answers POST to
`/api/sessions/{id}/questions/{questionID}/reply`. Secret inputs are masked and the
engine's answer audit is redacted; the receiving harness still receives their values.

Assistant replies render Markdown using react-markdown and remark-gfm: headings,
emphasis, lists, task lists, blockquotes, links, tables, and fenced code blocks.
User messages and tool output remain plain text. Code blocks and tables scroll
within the message. Copy preserves the original Markdown. Raw HTML is disabled,
unsafe URL schemes are filtered, and Markdown images appear as links rather than
automatically fetching remote resources. The Design system gallery includes a
Markdown example for visual validation. Syntax highlighting is not included yet.

Mermaid `flowchart` and `graph` blocks render inline between paragraphs in both
chats, with a preview capped at 320px high. **Expand** opens a modal with zoom,
**Fit**, scrolling, and Escape to close. The renderer loads on demand and uses
the current theme. It accepts `mermaid` fences, unlabelled flowchart code blocks,
and standalone flowchart headers followed by contiguous recognizable statements
without fences. Explicit `mermaid` fences are recommended for complex syntax.
Incomplete, invalid, or unsupported diagrams retain their source. Rendering is
debounced during streaming; copying a message keeps its original text. Generated
SVG is displayed as an isolated image, with HTML labels, agent configuration
directives, and remote image resources disabled. Other Mermaid diagram types are
not included in this slice. The Design system gallery has a **Flowcharts** example
for checking inline layout and expansion without running a harness.

The composer grows up to ten visible lines before scrolling. Enter sends;
Shift+Enter and Ctrl+J insert a newline. Project/session data and native harness
continuity survive restart; drafts are browser-memory-only. Feedback on message
buttons is local UI state and is not sent to a provider.

The window has a blue/lavender gradient backdrop and can be dragged by its header.
Header buttons do not drag it. Mobile retains the full-screen layout, and side
panels become toggleable overlays. Chat/Add project borders do not change on focus.

The moon/sun button at the bottom of the project rail switches between Light and
Graphite Dark. Light is the default. The browser stores the choice under `klm.theme`
and applies it before the app renders. Themes share the same layout and semantic
CSS tokens, including dialogs, permission cards, Markdown, and error states.

The footer below the composer shows live session input/output totals and context
tokens / model window and percentage as plain text, without a Context label.
Hover the context figures for unrounded counts;
the same details are available to screen readers. Input includes cached
tokens; output includes reported reasoning. Unknown values display `—`, not zero.
Values persist with the session and refresh when the harness reports usage, rather
than on each text delta. Older sessions populate on their next harness run.

The order is turns/steps, LLM time, cache hit, context, quota (when eligible), input/output.
Context replaces average TTFT. Turns, steps, LLM time, and cache hit
still use static reference values pending later metric decisions. Context is the
latest available snapshot, not cumulative session consumption or an estimate of
an unsent draft. Hover the usage figures for unrounded token counts.

Quota replaces tok/s with `83% · 2 days` (used percentage and time to reset).
The most-used unexpired window is shown; hover lists all returned windows, source,
reset timestamps, and observation time. Quota is hidden for intermediary providers,
API-key billing, unsupported routes, or unverified account matches. It refreshes
every two minutes while visible. Expired windows and snapshots older than five
minutes are hidden; a reset countdown never invents a fresh percentage.

Old mock localStorage data is not imported or deleted. Live sessions have no
fabricated responses, MCP connectivity indicators, or permissions/model
selectors that do not affect execution. The design-system gallery uses static
examples only. Native OS picker language may follow Windows settings; app copy
remains English.

See the root README for current execution, permission, and persistence boundaries.
