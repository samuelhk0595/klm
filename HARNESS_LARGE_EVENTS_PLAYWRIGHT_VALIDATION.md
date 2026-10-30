# Large-event Playwright validation

Date: 2026-10-30. Browser-tested commit: `c4042ca`. Subsequent frontend
correction is described below and has build evidence only.

## Result

**Not fully accepted.** Assistant content of 25/256 MiB remains usable and bounded
in the browser, and streaming event/conversation downloads preserve the generated
content. The initial browser run found a UI defect: large tool/command/file-change
events in Agent work groups did not provide the content paging/download controls.
The follow-up correction has not yet been revalidated in the browser.

The requested fresh subagent prepared the isolated dataset and launched both
services, but its tool invocation did not return after startup. Its task was
cancelled. The parent agent inspected the running processes, reused the already
ready services, performed the Playwright actions below directly, and cleaned up.
No second server startup or foreground wait was attempted during that initial run.

## Isolation

- Worktree: `C:\Users\Samuel\Documents\Projects\Personal\klm-harness-large-events`.
- Test root: `C:\Users\Samuel\AppData\Local\Temp\opencode\klm-large-events-playwright-20261030-a91f`.
- Data root: `<test-root>\data`; skills root: `<test-root>\skills`.
- Binary: `<test-root>\klm-probe.exe`, compiled from this worktree with `-tags dev`.
- Fixture generation refuses to run unless the exact isolated `KLM_DEV_DATA_DIR`
  and absolute `KLM_USER_SKILLS_DIR` are set. It uses a fresh SQLite cutover and
  project directory below the test root, never the installed data directories.
- Synthetic project ID: `large-events-a91f`.
- Engine: `127.0.0.1:17331`, PID **50876**; Vite: `127.0.0.1:5173`, PID **51636**.
  Vite's worktree-local esbuild child was PID **45376**.
- Engine logs confirm loopback binding, storage contract 3 and explicit dev-only
  client-auth bypass. Consequently this browser run does not establish release
  pairing/cookie/revocation acceptance.
- Real installed engine PID **42452** (`7331`) and desktop PID **45668** (`7332`)
  remained running throughout. No installed CLI lifecycle command, installer,
  registry change, production-data access or model/provider prompt was used.

Startup logs show the engine and Vite ready at local 01:20:33/01:20:32. On takeover,
their launching shells had already exited. That establishes that services were
ready while the subagent was stuck; the exact tool-runner wait mechanism was not
observable from the remaining process list.

## Generated scenarios

The temporary Go probe used generated readers through the actual SSE/JSONL
projection, native adapter and SQL publication paths. It did not construct large
fixture strings in the browser. Each payload has unique begin/end markers and
Unicode (`é🐱`, `界😀`). The manifest records complete canonical hashes and sizes.

| Conversation | Useful payload | Canonical event bytes | Browser result |
| --- | ---: | ---: | --- |
| `oc-256-assistant` | 256 MiB + markers | 268,435,731 | Preview, Open content, Next/Previous and Download event passed. |
| `pi-25-assistant` | 25 MiB + markers | 26,214,699 | Preview and bounded page navigation passed. |
| `pi-100-tool` | 100 MiB + markers | 209,715,583 | **Failed UI controls** inside Agent work. Backend content/tail complete. |
| `cx-25-command` | 25 MiB + markers | 52,429,271 | **Failed UI controls** inside Agent work. Backend content/tail complete. |
| `cx-25-files` | 25 MiB diff + markers | 52,429,288 | **Failed UI controls** inside Agent work. Conversation export complete. |

Canonical sizes include the established normalized text and native-detail
representations, so some tool outputs occur in both fields. These measurements
must not be mistaken for evidence of storage deduplication.

The seed probe also verified an OpenCode 100 MiB summary followed by `session.error`,
Pi 25 MiB `agent_end` and 25 MiB `turn_end` followed by a separate `agent_settled`,
and cancellation of a blocked synchronous handoff while another small SQL event
persisted. Those are **backend probe evidence**, not real harness Stop/UI execution.

## Real Playwright actions

1. Opened `http://127.0.0.1:5173`, selected the synthetic project and expanded
   Ungrouped. The displayed project path was below the test root.
2. Opened `oc-256-assistant`: preview was 4,093 JS characters (bounded UTF-8
   preview), and history response was **4,555 bytes**, not 256 MiB.
3. Clicked Open content, Next part and Previous part. The displayed canonical
   page was **65,533–65,536 characters**. Each content response decoded to at
   most **65,536 bytes**; base64 HTTP responses were approximately **87,540 bytes**.
4. Opened `pi-25-assistant` and paged its content. Its preview/page had the same
   bounds. Total document text stayed around **70,162 characters**.
5. Expanded Agent work for `pi-100-tool`, `cx-25-command` and `cx-25-files`, and
   opened the tool detail row. No Open content or Download event controls appeared.
   The rows displayed only their title and completed status.
6. Used bounded browser API tail reads for every generated conversation. All five
   returned HTTP 200, matching manifest digest/total size, 65,536 decoded bytes,
   `hasMore=false`, and the exact Unicode end marker.
7. Clicked Download event on the OpenCode conversation. Playwright received the
   attachment and saved it below the test root, with no download failure.
8. Clicked Session log (the export action) on `cx-25-files`. Playwright received
   and saved the conversation attachment. Reusing its consumed ticket returned
   **404**. Ticket URLs/tokens are omitted from this report.
9. Repeated assistant content opening at a 390×844 browser viewport: preview and
   page remained bounded, and there was no horizontal document overflow. No
   `KlmMobile` bridge was present: this is **narrow browser**, not Flutter acceptance.

## Download integrity and browser heap

SHA-256 verification used `FileStream`/`SHA256.ComputeHash`, not whole-file reads.

- OpenCode event attachment: **268,435,731 bytes**;
  `e912987519dd1d35c1c0a97eca42316085b0c5ec7c602612917cc578a3dc5814`.
  This exactly matches the canonical manifest digest; the end marker was present.
- Conversation attachment: **52,429,693 bytes**;
  `b02aa5b170dba09c0a30a86131103b8832b6bde80a6bff53b4f0d0940762cb47`.
  An independent streaming GET of the isolated runtime export produced the same
  hash. The file-change end marker was present.
- Chromium CDP, forced GC before/after the conversation export:
  used heap **15,084,928 → 15,085,240 bytes**; backing storage **7,072,960 → 7,072,960**.
  Instrumented `URL.createObjectURL` calls during that export: **0**.
- Unforced `performance.memory` readings during opening/paging the 256 MiB event
  were approximately **25.2–29.9 MB**. These are sampled browser observations,
  not a formal peak-memory measurement of every download/harness scenario.

## Reproducible UI defect

Severity: large useful tool content is persisted but inaccessible through the
normal grouped-event details UI.

1. Select the synthetic project, then `pi-100-tool` (also reproducible on either
   Codex command/file-change conversation).
2. Expand Worked for …, then open the read/Command/File changes row.
3. Observe title/status only. The event has a valid content reference on the API,
   but no page/download controls are rendered.

Likely location: `clients/desktop/src/features/chat/ConversationEvents.tsx`.
`EventSequence` routes `event.content` through EventContent for standalone events,
but `ConversationWork` routes grouped entries through EventDetails. EventDetails
only inspects `event.text` and `event.data`; large content-reference events have
empty text and no inline native data. The grouped path therefore loses the reader.

### Follow-up correction

`EventDetails` now delegates content-reference events to the existing EventContent
reader, exposing the same bounded paging/download UI within grouped work details.
Reasoning with a content reference is no longer filtered out for having empty
inline text; its collapsed thought row uses the explicitly bounded preview.
No new content-loading or download mechanism was introduced.

`npm run build` passed (TypeScript and Vite; Vite build 42.69 s, existing large-chunk
warning). No full test suite was run. Browser revalidation of this correction is
**pending**, including the three originally failing cases and content-backed
reasoning. The initial screenshots/results above remain pre-fix evidence.

A follow-up attempt launched isolated engine PID 52156 and Vite PID 7928 at local
02:03:03. The user reported the command-return hang again and asked not to repeat
that workflow. Both services and Vite's identified esbuild child were terminated
without attempting further browser actions. Ports 17331/5173 were released;
installed listeners 7331/7332 remained owned by 42452/45668. No further background
server startup was attempted.

## Console and limitations

- `/api/sessions/<synthetic-id>/models` returned 502 in the isolated, unconfigured
  harness environment. Conversation/history/content/download navigation remained
  usable. This is recorded separately from the grouped-content defect.
- One expected 404 came from intentionally reusing the consumed ticket.
- No real provider/model task, installed conversation, release auth/pairing,
  native Tauri window, Flutter WebView, full suite or review agent was exercised.
- This run does not prove all crash/restart/Stop cases or aggregate concurrency
  measurements. Existing backend probe evidence is not relabeled as UI evidence.

## Artifacts and cleanup

Under the test root:

- `seed-manifest.json`, engine/Vite stdout/stderr logs and the isolated database.
- `oc-256-preview.png`, `oc-256-open-content.png`, `oc-256-narrow-browser.png`.
- `pi-100-tool-missing-content.png`, `pi-25-assistant-validation.png`,
  `cx-25-command-validation.png`, `cx-25-files-validation.png`.
- `oc-256-event.json` and `cx-25-files-conversation.json`.
- Temporary probe `large_events_playwright_a91f_test.go`, moved out of `engine/`
  after validation so it is not left in the normal project test suite.

Structured Playwright evidence is also in the harness-generated
`.playwright-mcp/large-events-a91f-evidence.json` artifact directory.

The Playwright tab was closed. Test PIDs 50876/51636/45376 were terminated only
after exact executable/parent/creation checks. Their test listeners 17331/5173
were released; installed listeners 7331/7332 still belonged to 42452/45668.
The test database/artifacts are retained in the temporary directory for inspection.
The browser results above predate the follow-up frontend correction and its commit.
