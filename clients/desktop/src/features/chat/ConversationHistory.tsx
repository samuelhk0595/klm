import { useEffect, useRef, type ReactNode } from 'react';
import type { EventPage, Session, SessionResponse } from '../../engine';
import { ConversationEvents } from './ConversationEvents';
import { HistoryLoader } from './HistoryLoader';

export function ConversationHistory({ session, subagents, working, label, empty, onSnapshot, onHistory, onEventChange, onAskSide, onOpenSubagent }: {
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
  useEffect(() => {
    if (following.current && history.current) history.current.scrollTo({ top: history.current.scrollHeight });
  }, [session.updatedAt, session.history?.total, !!session.history, working]);
  return <div className="chat-history" ref={history} role="log" aria-label={label} aria-live="polite" onScroll={event => {
    const element = event.currentTarget;
    following.current = element.scrollHeight - element.scrollTop - element.clientHeight < 24;
  }}>
    <HistoryLoader key={`history-loader/${session.id}`} session={session} onPage={onHistory} />
    {!session.history && !session.events.length
      ? <p className="muted" role="status">Loading history...</p>
      : session.events.length || working
        ? <ConversationEvents key={`conversation-events/${session.id}`} session={session} subagents={subagents} working={working} onSnapshot={onSnapshot} onEventChange={onEventChange} onAskSide={onAskSide} onOpenSubagent={onOpenSubagent} />
        : empty}
  </div>;
}
