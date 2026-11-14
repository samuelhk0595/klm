import { useEffect, useRef, useState } from 'react';
import { request, type EventPage, type Session } from '../../engine';
import { AutoPage } from '../pagination/AutoPage';
import { PAGE_SIZE, scrollRoot, serializeHistory } from '../pagination/pagination';

type HistoryRequest = { promise: Promise<EventPage>; controller: AbortController; readers: number };
const requests = new Map<string, HistoryRequest>();
export function HistoryLoader({ session, onPage, direction = 'before', active = true, background = false }: { session: Session; onPage: (page: EventPage) => void; direction?: 'before' | 'after'; active?: boolean; background?: boolean }) {
  const container = useRef<HTMLDivElement>(null);
  const controller = useRef<AbortController | null>(null);
  const flight = useRef(false);
  const current = useRef(session); current.current = session;
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [intent, setIntent] = useState('before');
  const [gesture, setGesture] = useState(0);
  const explicit = useRef(false);
  const [wake, setWake] = useState(0);
  useEffect(() => {
    if (active) { setWake(value => value + 1); return; }
    controller.current?.abort(); flight.current = false; setBusy(false);
  }, [active]);
  useEffect(() => {
    const root = container.current?.closest<HTMLElement>('.agent-work__items, .chat-history');
    flight.current = false; explicit.current = false; setBusy(false); setError(''); setIntent('before'); setGesture(0);
    if (root) root.dataset.historyDirection = 'before';
    let touch: number | undefined; let lastGesture = 0; let lastScroll = root?.scrollTop ?? 0;
    const position = () => { lastScroll = root?.scrollTop ?? 0; };
    const intent = (event: Event) => {
      if (!root) return;
      if (event.target instanceof HTMLElement && event.target !== root && scrollRoot(event.target) !== root) return;
      if (event instanceof WheelEvent) {
        if (!event.deltaY) return;
        root.dataset.historyDirection = event.deltaY < 0 ? 'before' : 'after';
      }
      if (event instanceof KeyboardEvent) {
        if (event.target instanceof Element && event.target.closest('input, textarea, select, button, [contenteditable="true"]')) return;
        if (!['ArrowUp', 'PageUp', 'Home', 'ArrowDown', 'PageDown', 'End', ' '].includes(event.key)) return;
        root.dataset.historyDirection = ['ArrowUp', 'PageUp', 'Home'].includes(event.key) || event.key === ' ' && event.shiftKey ? 'before' : 'after';
      }
      if (typeof TouchEvent !== 'undefined' && event instanceof TouchEvent && event.touches[0]) {
        const y = event.touches[0].clientY;
        if (touch === undefined || touch === y) { touch = y; return; }
        root.dataset.historyDirection = y > touch ? 'before' : 'after'; touch = y;
      }
      if (event.type === 'scroll') {
        const delta = root.scrollTop - lastScroll; lastScroll = root.scrollTop;
        if (Math.abs(delta) < 1) return;
        root.dataset.historyDirection = delta < 0 ? 'before' : 'after';
      }
      setIntent(root.dataset.historyDirection ?? 'before');
      if (Date.now() - lastGesture > 350) { lastGesture = Date.now(); explicit.current = true; setGesture(value => value + 1); }
    };
    const touchStart = () => { touch = undefined; };
    root?.addEventListener('wheel', intent, { passive: true }); root?.addEventListener('keydown', intent); root?.addEventListener('touchmove', intent, { passive: true });
    root?.addEventListener('scroll', intent, { passive: true }); root?.addEventListener('touchstart', touchStart, { passive: true }); root?.addEventListener('klm:history-position', position);
    return () => {
      controller.current?.abort();
      root?.removeEventListener('wheel', intent); root?.removeEventListener('keydown', intent); root?.removeEventListener('touchmove', intent);
      root?.removeEventListener('scroll', intent); root?.removeEventListener('touchstart', touchStart); root?.removeEventListener('klm:history-position', position);
    };
  }, [session.id]);
  const history = session.history;
  const cursor = direction === 'before' ? history?.nextCursor : history?.afterCursor;
  const more = direction === 'before' ? history?.hasMore : history?.hasNewer;
  async function load() {
    const root = container.current?.closest<HTMLElement>('.agent-work__items, .chat-history');
    if (!active || flight.current || !cursor || !root || !root.getClientRects().length) return;
    if (session.events.length >= 200 && (root.scrollHeight <= root.clientHeight + 1 && !explicit.current || root.dataset.historyDirection && root.dataset.historyDirection !== direction)) return;
    explicit.current = false;
    const abort = new AbortController(); controller.current = abort;
    flight.current = true; setBusy(true); setError('');
    let release: (() => void) | undefined;
    try {
      const key = `${session.id}/${direction}/${cursor}`;
      let pending = requests.get(key);
      if (!pending) {
        const shared = new AbortController();
        pending = { controller: shared, readers: 0, promise: serializeHistory(session.id, () => request<EventPage>(`/api/sessions/${encodeURIComponent(session.id)}/history?${direction}=${encodeURIComponent(cursor!)}&limit=${PAGE_SIZE}`, 'GET', undefined, 15000, shared.signal)) };
        requests.set(key, pending);
        void pending.promise.finally(() => { if (requests.get(key) === pending) requests.delete(key); }).catch(() => {});
      }
      const shared = pending; shared.readers++;
      let released = false;
      release = () => {
        if (released) return; released = true;
        if (!--shared.readers) { shared.controller.abort(); if (requests.get(key) === shared) requests.delete(key); }
      };
      abort.signal.addEventListener('abort', release, { once: true });
      const page = await shared.promise;
      if (abort.signal.aborted || document.hidden || !root.isConnected || !root.getClientRects().length) return;
      const now = current.current.history;
      if (!now || (direction === 'before' ? now.nextCursor : now.afterCursor) !== cursor) return;
      if (!page.events?.length || (direction === 'before' ? page.startIndex >= now.startIndex : page.endIndex <= now.startIndex + current.current.events.length)) throw new Error('History did not advance. Retry.');
      onPage({ ...page, direction });
    } catch (cause) { if (!abort.signal.aborted) setError(cause instanceof Error ? cause.message : 'Could not load history.'); }
    finally { release?.(); if (controller.current === abort && !abort.signal.aborted) { flight.current = false; setBusy(false); } }
  }
  const space = <div data-history-space={direction} aria-hidden="true" style={{ height: 0, flexShrink: 0 }} />;
  return <div ref={container} style={{ flexShrink: 0 }}><AutoPage hasMore={active && !!more && (session.events.length < 200 || intent === direction)} cursor={`${session.id}/${direction}/${cursor ?? ''}/${intent}/${gesture}/${wake}`} busy={busy} loading={busy && !background} error={error} onLoad={() => void load()}>{space}</AutoPage></div>;
}
