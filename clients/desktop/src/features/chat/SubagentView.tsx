import type { EventPage, Session, SessionResponse } from '../../engine';
import { ConversationHistory } from './ConversationHistory';
import { SessionStatusBar } from './SessionStatusBar';

export function SubagentView({ session, onSnapshot, onHistory }: {
  session?: Session;
  onSnapshot: (session: SessionResponse) => void;
  onHistory: (page: EventPage) => void;
}) {
  if (!session) return <section className="subagent-view"><div className="empty-chat"><h2>Opening subagent...</h2></div></section>;
  const model = session.resolvedModel || session.model || 'Unknown model';
  const effort = session.resolvedEffort || session.effort || 'Unknown effort';
  return <section className="subagent-view" aria-label={`${session.title} subagent`}>
    <ConversationHistory session={session} working={session.status === 'running'} label={`${session.title} activity`} empty={<div className="empty-chat"><h2>No activity reported</h2></div>} onSnapshot={onSnapshot} onHistory={onHistory} />
    <footer className="subagent-status"><span title={model}>Model: {model}</span><span>Effort: {effort}</span><SessionStatusBar session={session} side /></footer>
  </section>;
}
