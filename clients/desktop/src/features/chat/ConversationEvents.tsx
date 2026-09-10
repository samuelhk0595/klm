import { useEffect, useRef, useState } from 'react';
import { MessagesSquare } from 'lucide-react';
import { Button } from '../../design-system/Button';
import { request, type EngineEvent, type Session } from '../../engine';
import { ChatMessage } from './ChatMessage';
import { SessionEvent } from './SessionEvent';

function ConsultationActivity({ event, output, sessionId, onSnapshot }: { event: EngineEvent; output: EngineEvent[]; sessionId: string; onSnapshot: (session: Session) => void }) {
  const [cancelling, setCancelling] = useState(false);
  const [error, setError] = useState('');
  const requestId = typeof event.data?.requestId === 'string' ? event.data.requestId : '';
  const answer = typeof event.data?.answer === 'string' ? event.data.answer : '';
  const failure = typeof event.data?.error === 'string' ? event.data.error : '';
  const active = event.status === 'queued' || event.status === 'answering';
  const pendingDelivery = event.data?.from === sessionId && (event.data?.delivery === 'pending' || event.data?.delivery === 'delivering');
  async function cancel() {
    if (cancelling) return;
    setCancelling(true); setError('');
    try { onSnapshot(await request<Session>(`/api/sessions/${encodeURIComponent(sessionId)}/consultations/${encodeURIComponent(requestId)}/cancel`, 'POST', {})); }
    catch (error) { setError(error instanceof Error ? error.message : 'Could not cancel the consultation.'); }
    finally { setCancelling(false); }
  }
  return <details className="session-event consultation-activity">
    <summary><MessagesSquare /><span>{event.title}</span><small className={event.status === 'failed' || event.status === 'interrupted' ? 'form-error' : 'muted'}>{event.status}</small></summary>
    <div className="consultation-body">
      <strong>Question</strong><p className="consultation-question">{event.text}</p>
      {answer && <><strong>Answer</strong><ChatMessage message={{ id: `${event.id}/answer`, role: 'assistant', text: answer }} /></>}
      {failure && <p className="form-error">{failure}</p>}
      {event.data?.delivery === 'interrupted' && <p className="form-error">Answer delivery interrupted by engine restart.</p>}
      {event.data?.delivery === 'failed' && <p className="form-error">Answer continuation failed.</p>}
      {(active || pendingDelivery) && <Button size="sm" disabled={cancelling} onClick={() => void cancel()}>{cancelling ? 'Cancelling...' : 'Cancel request'}</Button>}
      {error && <p role="alert" className="form-error">{error}</p>}
      {!!output.length && <details className="consultation-output"><summary>Agent activity</summary><div>{output.map(item => <SessionEvent key={item.id} event={item} />)}</div></details>}
    </div>
  </details>;
}

export function ConversationEvents({ session, onSnapshot, onAskSide }: {
  session: Session; onSnapshot: (session: Session) => void; onAskSide?: (messageId: string, passage: string) => void;
}) {
  const root = useRef<HTMLDivElement>(null);
  const [selection, setSelection] = useState<{ messageId: string; passage: string; left: number; top: number } | null>(null);
  useEffect(() => {
    if (!onAskSide) return;
    const readSelection = () => {
      const selected = window.getSelection();
      if (!selected?.rangeCount || selected.isCollapsed) return null;
      const range = selected.getRangeAt(0);
      const element = (node: Node) => node instanceof Element ? node : node.parentElement;
      const start = element(range.startContainer)?.closest('.message-text');
      const end = element(range.endContainer)?.closest('.message-text');
      const message = start?.closest<HTMLElement>('[data-message-id]');
      if (!start || start !== end || !message || message.closest('.consultation-activity') || !root.current?.contains(message)) return null;
      const passage = selected.toString();
      const bounds = range.getBoundingClientRect();
      return passage.trim() ? { messageId: message.dataset.messageId!, passage, left: Math.max(8, Math.min(bounds.left, window.innerWidth - 176)), top: Math.max(8, Math.min(bounds.top - 50, window.innerHeight - 52)) } : null;
    };
    const clear = () => setSelection(null);
    const onAction = (event: MouseEvent) => event.target instanceof Element && !!event.target.closest('.selection-action');
    const beginSelection = (event: MouseEvent) => { if (!onAction(event)) clear(); };
    const finishSelection = (event: MouseEvent) => {
      if (event.button === 0 && !onAction(event)) setSelection(readSelection());
    };
    const selectionChanged = () => {
      const current = readSelection();
      // Selection changes may dismiss an existing button, but only mouseup shows it.
      setSelection(previous => previous && current?.messageId === previous.messageId && current.passage === previous.passage ? previous : null);
    };
    document.addEventListener('mousedown', beginSelection);
    document.addEventListener('mouseup', finishSelection);
    document.addEventListener('selectionchange', selectionChanged);
    window.addEventListener('resize', clear);
    document.addEventListener('scroll', clear, true);
    return () => {
      document.removeEventListener('mousedown', beginSelection);
      document.removeEventListener('mouseup', finishSelection);
      document.removeEventListener('selectionchange', selectionChanged);
      window.removeEventListener('resize', clear);
      document.removeEventListener('scroll', clear, true);
    };
  }, [!!onAskSide, session.id]);
  // Only answering the other agent is background work. A continuation on the
  // requesting side is its response to the user, including in saved histories.
  const incoming = new Set(session.events
    .filter(event => event.type === 'consultation' && event.data?.to === session.id)
    .map(event => String(event.data?.requestId)));
  const grouped = new Map<string, EngineEvent[]>();
  for (const event of session.events) {
    if (!event.consultationId || !incoming.has(event.consultationId)) continue;
    const group = grouped.get(event.consultationId) ?? [];
    group.push(event); grouped.set(event.consultationId, group);
  }
  return <div ref={root} className="conversation-events">
    {session.events.filter(event => !event.consultationId || !incoming.has(event.consultationId)).map(event => event.type === 'consultation'
      ? <ConsultationActivity key={event.id} event={event} output={grouped.get(String(event.data?.requestId)) ?? []} sessionId={session.id} onSnapshot={onSnapshot} />
      : <SessionEvent key={event.id} event={event} />)}
    {selection && onAskSide && <Button size="sm" className="selection-action" style={{ left: selection.left, top: selection.top }} onPointerDown={event => event.preventDefault()} onClick={() => {
      onAskSide(selection.messageId, selection.passage); setSelection(null); window.getSelection()?.removeAllRanges();
    }}>Ask side agent</Button>}
  </div>;
}
