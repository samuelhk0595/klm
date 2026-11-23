import { useEffect, useRef, useState } from 'react';
import { Button } from '../../design-system/Button';
import { ENGINE_URL, request, type EngineEvent } from '../../engine';
import { openContentDownload } from '../../platform';
import { AutoPage } from '../pagination/AutoPage';

type ContentPage = { data: string; offset: number; nextOffset: number; hasMore: boolean; totalBytes: number };
type ContentState = { nextOffset: number; hasMore: boolean; chunks: string[]; tail: Uint8Array };
const decode = (data: string) => Uint8Array.from(atob(data), c => c.charCodeAt(0));

// Keep at most three incomplete UTF-8 bytes between requests. Decoding is
// transactional: a failed page leaves both the offset and the tail untouched.
function decodeChunk(data: Uint8Array, tail: Uint8Array, more: boolean) {
  const bytes = new Uint8Array(tail.length + data.length);
  bytes.set(tail); bytes.set(data, tail.length);
  let end = bytes.length;
  if (more && end) {
    let start = end - 1;
    while (start > 0 && (bytes[start] & 0xc0) === 0x80) start--;
    const first = bytes[start];
    const length = first >= 0xc2 && first <= 0xdf ? 2 : first >= 0xe0 && first <= 0xef ? 3 : first >= 0xf0 && first <= 0xf4 ? 4 : 1;
    if (end - start < length) end = start;
  }
  return { text: new TextDecoder('utf-8', { fatal: true, ignoreBOM: true }).decode(bytes.subarray(0, end)), tail: bytes.slice(end) };
}

export function EventContent({ event }: { event: EngineEvent }) {
  const content = event.content!;
  const identity = JSON.stringify([content.sessionId, content.domain, content.key, content.ordinal, content.digest]);
  return <ContentReader key={identity} event={event} />;
}

function ContentReader({ event }: { event: EngineEvent }) {
  const content = event.content!;
  const [opened, setOpened] = useState(false);
  const [page, setPage] = useState<ContentState>({ nextOffset: 0, hasMore: true, chunks: [], tail: new Uint8Array() });
  const currentPage = useRef(page); currentPage.current = page;
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const controller = useRef<AbortController | undefined>(undefined);
  const flight = useRef(false);
  const [downloading, setDownloading] = useState(false);
  const [downloadError, setDownloadError] = useState('');
  const downloadController = useRef<AbortController | undefined>(undefined);
  useEffect(() => () => { controller.current?.abort(); downloadController.current?.abort(); }, []);
  const path = (offset: number) => content.domain
    ? `/api/sessions/${encodeURIComponent(content.sessionId)}/graph/content?domain=${encodeURIComponent(content.domain)}&key=${encodeURIComponent(content.key ?? '')}&digest=${encodeURIComponent(content.digest)}&offset=${offset}`
    : `/api/sessions/${encodeURIComponent(content.sessionId)}/content?ordinal=${content.ordinal}&digest=${encodeURIComponent(content.digest)}&offset=${offset}`;
  async function load() {
    const previous = currentPage.current;
    if (flight.current || !previous.hasMore) return;
    flight.current = true;
    const current = new AbortController(); controller.current = current; setBusy(true); setError('');
    try {
      const result = await request<ContentPage>(path(previous.nextOffset), 'GET', undefined, 15000, current.signal);
      if (current.signal.aborted) return;
      const bytes = decode(result.data);
      if (result.offset !== previous.nextOffset || result.nextOffset !== result.offset + bytes.length || result.nextOffset > result.totalBytes || result.hasMore !== (result.nextOffset < result.totalBytes) || result.hasMore && !bytes.length) throw new Error('Content did not advance. Retry.');
      const decoded = decodeChunk(bytes, previous.tail, result.hasMore);
      const next = { nextOffset: result.nextOffset, hasMore: result.hasMore, chunks: [...previous.chunks, decoded.text], tail: decoded.tail };
      currentPage.current = next; setPage(next);
    } catch (cause) { if (!current.signal.aborted) setError(cause instanceof Error ? cause.message : 'Could not read content.'); }
    finally { if (controller.current === current) { flight.current = false; if (!current.signal.aborted) setBusy(false); } }
  }
  async function download() {
    if (downloadController.current && !downloadController.current.signal.aborted) return;
    const current = new AbortController(); downloadController.current = current; setDownloading(true); setDownloadError('');
    try {
      const ticket = await request<{ path: string }>(`/api/sessions/${encodeURIComponent(content.sessionId)}/content-downloads`, 'POST', {
        ordinal: content.ordinal, domain: content.domain, key: content.key, digest: content.digest,
      }, 15000, current.signal);
      if (current.signal.aborted) return;
      if (!/^\/api\/content-downloads\/[A-Za-z0-9_-]{43}\?contract=3$/.test(ticket.path)) throw new Error('Invalid download response.');
      await openContentDownload(ENGINE_URL + ticket.path);
    } catch (cause) { if (!current.signal.aborted) setDownloadError(cause instanceof Error ? cause.message : 'Could not download content.'); }
    finally { if (downloadController.current === current) { downloadController.current = undefined; if (!current.signal.aborted) setDownloading(false); } }
  }
  return <section className="session-event"><strong>{event.title || 'Message content'}</strong>
    {!opened && event.previewText && <div><span>Preview</span><pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{event.previewText}</pre></div>}
    <div>
      {!opened && <Button size="sm" onClick={() => setOpened(true)}>Open content</Button>}
      <Button size="sm" disabled={downloading} onClick={() => void download()}>Download event</Button>
    </div>
    {opened && <div role="region" aria-label={event.title || 'Message content'} tabIndex={0} style={{ maxHeight: 'min(60vh, 480px)', overflowY: 'auto', overscrollBehavior: 'contain', overflowAnchor: 'none' }}>
      <pre style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere', maxHeight: 'none', overflow: 'visible' }}>{page.chunks.map((chunk, index) => <span key={index}>{chunk}</span>)}</pre>
      <AutoPage hasMore={page.hasMore} cursor={`${content.digest}/${page.nextOffset}`} busy={busy} error={error} onLoad={() => void load()} />
    </div>}
    {downloadError && <p role="alert" className="form-error">{downloadError} <Button size="sm" disabled={downloading} onClick={() => void download()}>Retry</Button></p>}
  </section>;
}
