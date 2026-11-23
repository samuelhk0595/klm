import { useEffect, useLayoutEffect, useRef, useState, type ReactNode } from 'react';
import { type EventPage, type Session, type SessionResponse } from '../../engine';
import { ConversationEvents } from './ConversationEvents';
import { HistoryLoader } from './HistoryLoader';
import { PendingRequestPages } from './PendingRequestPages';
import { prepareWindowChange } from './historyWindow';
import { HistoryPagingContext } from './HistoryPagingContext';
import { mergeSessions, prependHistory } from './sessionState';
import { readConversationWindow, subscribeScope } from './scopedConnection';
import { Button } from '../../design-system/Button';
import { isHistoryVisible, useHistoryVisibility } from './useHistoryVisibility';

export function ConversationHistory({ session: incoming, subagents, working, label, empty, onSnapshot, onEventChange, onAskSide, onOpenSubagent }: {
  session: Session;
  subagents?: Session[];
  working: boolean;
  label: string;
  empty?: ReactNode;
  onSnapshot: (session: SessionResponse) => void;
  onHistory: (page: EventPage) => void;
  onEventChange?: (sessionId: string, event: import('../../engine').EngineEvent) => void;
  onAskSide?: (messageId: string, passage: string) => void;
  onOpenSubagent?: (sessionId: string) => void;
}) {
  const history = useRef<HTMLDivElement>(null);
  const following = useRef(true);
  const visible = useHistoryVisibility(history, incoming.id);
  // Each mounted reader owns its paged window. The shared session stays a live
  // metadata/tail source; scrolling here never moves another reader's window.
  const [window, setWindow] = useState<Session>();
  const owned = window?.id === incoming.id ? window : undefined;
  const total = Math.max(owned?.history?.total ?? 0, incoming.history?.total ?? 0);
  const session = owned ? { ...incoming, events: owned.events, history: owned.history ? { ...owned.history, total, hasNewer: owned.history.startIndex + owned.events.length < total } : undefined } : incoming;
  const current = useRef(session); current.current = session;
  const restore = useRef<(() => void) | undefined>(undefined);
  const [readError, setReadError] = useState('');
  const readFailed = useRef(false);
  const retry = useRef<() => void>(() => {});
  const geometry = (previous: Session, next: Session) => history.current && prepareWindowChange(history.current,
    { events: previous.events, startIndex: previous.history?.startIndex ?? 0 },
    { events: next.events, startIndex: next.history?.startIndex ?? 0, hasEarlier: next.history?.hasMore, hasLater: next.history?.hasNewer });
  function page(page: EventPage) {
    const previous = current.current; const next = prependHistory([previous], page)[0];
    if (next === previous) return;
    const applyGeometry = geometry(previous, next); if (!applyGeometry) return;
    // Filling a compact viewport is not navigation away from the live tail.
    // A real upward scroll already cleared following; bounded eviction clears
    // it here only when the next window actually leaves newer events unloaded.
    following.current = following.current && !next.history?.hasNewer; restore.current = applyGeometry;
    current.current = next; setWindow(next);
  }
  useLayoutEffect(() => { restore.current?.(); restore.current = undefined; }, [session.events]);
  useLayoutEffect(() => { setWindow(undefined); restore.current = undefined; following.current = true; readFailed.current = false; setReadError(''); if (history.current) history.current.scrollTop = history.current.scrollHeight; }, [incoming.id]);
  useLayoutEffect(() => {
    // Adopt only the initial hydrated page. Later shared tail refreshes are
    // metadata inputs, never replacements for this reader's current window.
    if (incoming.history) setWindow(previous => previous?.id === incoming.id ? previous : incoming);
  }, [incoming.id, !!incoming.history]);
  useEffect(() => {
    if (!owned || !visible) return;
    const controller = new AbortController(); let busy = false, dirty = false;
    let timer: ReturnType<typeof setTimeout> | undefined;
    async function refresh(explicit = false) {
      if (explicit) readFailed.current = false;
      if (readFailed.current || controller.signal.aborted || !isHistoryVisible(history.current)) return;
      if (busy) { dirty = true; return; } busy = true;
      const requested = current.current;
      try {
        const result = await readConversationWindow(requested.id, requested, controller.signal, false, following.current);
        if (controller.signal.aborted || !isHistoryVisible(history.current)) return;
        const latest = current.current;
        if (latest.id !== requested.id) return;
        if (latest.history?.startIndex !== requested.history?.startIndex || latest.events.length !== requested.events.length) { dirty = true; return; }
        const next = mergeSessions([latest], [result], true)[0];
        if (!following.current) { const change = geometry(latest, next); if (!change) return; restore.current = change; }
        current.current = next; setWindow(next); setReadError('');
      } catch (cause) { if (!controller.signal.aborted) { readFailed.current = true; setReadError(cause instanceof Error ? cause.message : 'Could not refresh history.'); } }
      finally { busy = false; if (dirty && !controller.signal.aborted && !readFailed.current) { dirty = false; schedule(); } }
    }
    const schedule = () => { if (!timer) timer = setTimeout(() => { timer = undefined; void refresh(); }, 150); };
    retry.current = () => { void refresh(true); };
    const unsubscribe = subscribeScope('conversation', session.id, schedule);
    void refresh();
    return () => { controller.abort(); clearTimeout(timer); unsubscribe(); };
  }, [session.id, !!owned, visible]);
  useLayoutEffect(() => {
    if (visible && following.current && history.current) {
      history.current.scrollTo({ top: history.current.scrollHeight });
      history.current.dispatchEvent(new Event('klm:history-position'));
    }
  }, [session.events, session.updatedAt, session.history?.total, !!session.history, working, visible]);
  return <div className="chat-history" ref={history} role="log" tabIndex={0} aria-label={label} aria-live="polite" style={{ overflowAnchor: 'none' }} onScroll={event => {
    const element = event.currentTarget;
    following.current = !session.history?.hasNewer && element.scrollHeight - element.scrollTop - element.clientHeight < 24;
  }}>
    <HistoryLoader key={`history-loader/${session.id}`} session={session} onPage={page} active={visible} />
    {!session.history && !session.events.length
      ? <p className="muted" role="status">Loading history...</p>
      : session.events.length || working
        ? <HistoryPagingContext.Provider value={direction => <HistoryLoader session={session} onPage={page} direction={direction} active={visible} />}><ConversationEvents key={`conversation-events/${session.id}`} session={session} subagents={subagents} working={working} onSnapshot={onSnapshot} onEventChange={onEventChange} onAskSide={onAskSide} onOpenSubagent={onOpenSubagent} /></HistoryPagingContext.Provider>
         : empty}
    <HistoryLoader key={`newer/${session.id}`} session={session} onPage={page} direction="after" active={visible} background={following.current} />
    {readError && <p className="form-error" role="alert">{readError} <Button size="sm" onClick={() => retry.current()}>Retry</Button></p>}
    {session.role !== 'subagent' && <PendingRequestPages key={`pending/${session.id}`} session={session} onSnapshot={onSnapshot} />}
  </div>;
}
