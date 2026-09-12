import { useEffect, useRef, useState } from 'react';
import { Panel } from '@xyflow/react';
import { Atom, Diamond, FilePenLine, FileText, Forward, Search, Terminal, X } from 'lucide-react';
import { IconButton } from '../../design-system/Button';
import { TabNav } from '../../design-system/TabNav';
import type { NodeActivity } from './graph-run-demo';

const eventIcons = { Thinking: Atom, Read: FileText, Search, Command: Terminal, Write: FilePenLine, Choice: Diamond, Dispatch: Forward };
const statusLabels = { pending: 'Pending', running: 'Running', completed: 'Completed' };

export function GraphRunActivity({ nodeName, activity, onClose }: {
  nodeName: string; activity: NodeActivity; onClose: () => void;
}) {
  const [tab, setTab] = useState<'run' | 'input' | 'output'>('run');
  const log = useRef<HTMLDivElement>(null);
  const following = useRef(true);
  useEffect(() => {
    if (tab === 'run' && following.current && log.current) log.current.scrollTop = log.current.scrollHeight;
  }, [tab, activity.events.length]);

  return <Panel position="bottom-center" className="graph-run-activity-position nodrag nopan nowheel">
    <section className="graph-run-activity" aria-label={`${nodeName} activity`}>
      <header className="graph-run-activity-heading">
        <strong className="graph-run-activity-title" title={nodeName}>{nodeName}</strong>
        <TabNav value={tab} onChange={setTab} items={[{ value: 'run', label: 'Run' }, { value: 'input', label: 'Input' }, { value: 'output', label: 'Output' }]} />
        <span className={`graph-run-status graph-run-status--${activity.status}`}>{statusLabels[activity.status]}</span>
        <IconButton label="Close node activity" onClick={onClose}><X /></IconButton>
      </header>
      {tab === 'run' ? <div ref={log} className="graph-run-event-log" role="log" aria-label="Node events" aria-live="polite" tabIndex={0}
        onScroll={event => { const element = event.currentTarget; following.current = element.scrollHeight - element.scrollTop - element.clientHeight < 24; }}>
        {activity.events.map((event, index) => {
          const Icon = eventIcons[event.kind];
          return <div className="graph-run-event" key={`${event.time}:${event.kind}`}><time>{event.time}</time><Icon aria-hidden="true" /><span className="graph-run-event-kind">{event.kind}</span><span className="graph-run-event-text">{event.text}</span>{activity.status === 'running' && index === activity.events.length - 1 && <span className="graph-run-event-cursor" aria-hidden="true" />}</div>;
        })}
      </div> : <div className="graph-run-payload" role="region" aria-label={tab === 'input' ? 'Input payload' : 'Output payload'} tabIndex={0}>
        {tab === 'input' ? activity.input !== undefined && <pre>{JSON.stringify(activity.input, null, 2)}</pre> : activity.status === 'completed' && activity.output !== undefined && <>
          {activity.output.choice && <div className="graph-run-output-choice"><span>Choice</span><code>{activity.output.choice}</code></div>}
          <pre>{JSON.stringify(activity.output.payload, null, 2)}</pre>
        </>}
      </div>}
    </section>
  </Panel>;
}
