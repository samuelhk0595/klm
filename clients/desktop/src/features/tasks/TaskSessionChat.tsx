import { useEffect, useRef, useState } from 'react';
import { EllipsisVertical } from 'lucide-react';
import { IconButton } from '../../design-system/Button';
import { Menu } from '../../design-system/Menu';
import { Toggle } from '../../design-system/Toggle';
import { ChatMessage } from '../chat/ChatMessage';
import { MessageComposer, type ComposerDraft } from '../chat/MessageComposer';
import { harnessLabel, runDuration, runIsActive, runStatusLabel, runUserWait, telegramDeliveryProblem, triggerOriginLabel, type TaskRun, type TelegramLink } from './prototype';
import { TaskRunSnapshot } from './TaskRunSnapshot';

// Uses the normal chat components with local handlers, without engine-side sessions.
export function TaskSessionChat({ run, timezone, telegram, draft, onDraftChange, onSend, onFavorite, onYoloChange }: {
  run: TaskRun; timezone: string; telegram: TelegramLink; draft: ComposerDraft; onDraftChange: (draft: ComposerDraft) => void;
  onSend: (text: string) => void; onFavorite: (eventId: string, favorite: boolean) => void;
  onYoloChange: (yolo: boolean) => void;
}) {
  const session = run.session;
  const [optionsOpen, setOptionsOpen] = useState(false);
  const history = useRef<HTMLDivElement>(null);
  const following = useRef(true);
  useEffect(() => {
    if (following.current) history.current?.scrollTo({ top: history.current.scrollHeight });
  }, [session.id, session.events.length]);
  return <>
    <div className="chat-history" ref={history} role="log" aria-label="Conversation" aria-live="polite" onScroll={event => {
      const element = event.currentTarget;
      following.current = element.scrollHeight - element.scrollTop - element.clientHeight < 24;
    }}>
      <section className="task-chat-context" aria-label="Task run">
        <div className="task-chat-metadata"><strong className={`task-run-status is-${run.status}`}>{runStatusLabel[run.status]}</strong><span>{triggerOriginLabel[run.origin]}</span><span>Execution {runDuration(run)}</span><span>User wait {runUserWait(run)}</span></div>
        {run.status !== 'queued' && run.summary && <p>{run.summary}</p>}
        {run.status === 'waiting' && run.snapshot.questions === 'telegram' && telegramDeliveryProblem(telegram) && <p className="task-link-needed">{telegramDeliveryProblem(telegram)}</p>}
        <TaskRunSnapshot run={run} timezone={timezone} />
      </section>
      {session.events.map(event => event.type === 'error'
        ? <div key={event.id} className="storage-error" role="alert">{event.text}</div>
        : event.type === 'agent_prompt' ? <div key={event.id} className="task-chat-instructions"><strong>Task instructions</strong><p>{event.text}</p></div>
        : <ChatMessage key={event.id} message={{ id: event.id, role: event.type === 'assistant' ? 'assistant' : 'user', text: event.text, status: event.type === 'assistant' ? 'completed' : undefined, favorite: event.favorite }} onFavorite={event.type === 'assistant' ? async favorite => onFavorite(event.id, favorite) : undefined} />)}
    </div>
    <div className="composer-area"><MessageComposer key={session.id} draft={draft} onDraftChange={onDraftChange} onSend={message => onSend(message.text)} modelControl={<>
      <div className="session-options">
        {session.yolo && <span className="yolo-indicator" title="YOLO mode is active for this session">YOLO</span>}
        <Menu label="Session options" className="session-options-menu" open={optionsOpen} onOpenChange={setOptionsOpen} trigger={props => <IconButton {...props} label="Session options" size="sm"><EllipsisVertical /></IconButton>}>
          <Toggle label="YOLO mode" checked={!!session.yolo} disabled={runIsActive(run)} title={runIsActive(run) ? 'Available after the task run ends' : 'YOLO mode for this session'} onCheckedChange={onYoloChange} />
        </Menu>
      </div>
      <span className="task-chat-model">{harnessLabel[session.harness]} · {session.model}{session.effort && ` · ${session.effort}`}</span>
    </>} /></div>
  </>;
}
