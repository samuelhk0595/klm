import { useState } from 'react';
import { AutoPage } from '../pagination/AutoPage';
import { X } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { request, type Session, type SessionResponse } from '../../engine';
import { agentExecutionBlocked, useSessionAgent } from '../agents/binding';
import { usePendingPage } from './PendingRequestPages';

const labels = { queued: 'Queued', steering: 'Next turn', sending: 'Sending', accepted: 'Next turn', interrupting: 'Interrupting', paused: 'Paused', uncertain: 'Delivery unconfirmed' };

export function MessageQueue({ session, disabled, onSnapshot, onInterrupt }: { session: Session; disabled?: boolean; onSnapshot: (value: SessionResponse) => void; onInterrupt: (sessionId: string) => void }) {
  useSessionAgent(session);
  const executionBlocked = agentExecutionBlocked(session);
  const [pending, setPending] = useState('');
  const [error, setError] = useState('');
  const queue = usePendingPage(session, 'queue', session.queue ?? []);
  const messages = queue.all;
  async function update(id: string, send: boolean) {
     if (pending || disabled || send && (agentExecutionBlocked(session) || session.delegation?.actions.interrupt === false)) return;
    const retryUncertain = send && messages.find(item => item.id === id)?.status === 'uncertain';
    if (retryUncertain && !window.confirm('Delivery was not confirmed. Retrying may deliver this message twice. Retry delivery?')) return;
    if (send) onInterrupt(session.id);
    setPending(id); setError('');
    try {
      onSnapshot(await request<SessionResponse>(`/api/sessions/${encodeURIComponent(session.id)}/queue/${encodeURIComponent(id)}${send ? '/send' : ''}`, send ? 'POST' : 'DELETE', send ? { retryUncertain } : undefined));
      if (!send) queue.resolved(id); else queue.refresh();
    } catch (error) { setError(error instanceof Error ? error.message : 'Could not update the queue.'); }
    finally { setPending(''); }
  }
  if (!messages.length && !error && !queue.error) return null;
  return <div ref={queue.root}><section className="message-queue" aria-label="Message queue">
    {messages.map(message => <div className="message-queue-item" key={message.id} data-history-anchor={`queue:${message.id}`}>
      <div className="message-queue-content"><small>{message.origin ? `${message.origin.kind === 'instruction' ? 'Instruction' : 'Initial prompt'} from ${message.origin.title} (${message.origin.sessionId}) · ` : ''}{labels[message.status]}</small><p>{message.text}</p>{message.error && <small className="form-error">{message.error}</small>}</div>
      <div className="message-queue-actions">
        {message.status !== 'sending' && message.status !== 'interrupting' && <Button size="sm" disabled={disabled || !!pending || executionBlocked || session.delegation?.actions.interrupt === false} onClick={() => void update(message.id, true)}>Send now</Button>}
        <IconButton label="Remove queued message" disabled={disabled || !!pending || message.status === 'sending' || message.status === 'accepted' || message.status === 'interrupting'} onClick={() => void update(message.id, false)}><X /></IconButton>
      </div>
    </div>)}
    <AutoPage hasMore={queue.more} cursor={`${session.id}/${queue.cursor}`} busy={!!pending || queue.busy} loading={queue.loading} error={queue.error} onLoad={queue.load} onRetry={queue.retry} />
    {error && <p role="alert" className="form-error">{error}</p>}
  </section></div>;
}
