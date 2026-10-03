import { useEffect, useId, useRef, useState } from 'react';
import { CalendarClock, ChevronDown, MoreHorizontal, MousePointer2, Pause, Pencil, Play, Timer, Trash2, UserRound, Webhook, Workflow, X } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Badge } from '../../design-system/Badge';
import { Menu, MenuItem } from '../../design-system/Menu';
import { TabNav } from '../../design-system/TabNav';
import { useAuthoringCatalog } from '../graphs/catalog';
import { graphListEntry } from '../graphs/files';
import { formatRunDate, harnessLabel, runDuration, runUserWait, runStatusLabel, taskTriggerLabel, telegramDeliveryProblem, telegramStatus, triggerOriginLabel, type PrototypeTask, type TaskRun, type TelegramLink } from './prototype';
import { TaskRunSnapshot } from './TaskRunSnapshot';

export function TaskDetailsDialog({ projectId, task, runs, nextFiring, telegram, timezone, onClose, onEdit, onDelete, onEnabledChange, onRun, onOpenSession, onTelegramSettings }: {
  projectId: string; task: PrototypeTask; runs: TaskRun[]; nextFiring?: string; telegram: TelegramLink;
  timezone: string;
  onClose: () => void; onEdit: () => void; onDelete: () => void; onEnabledChange: (enabled: boolean) => void;
  onRun: () => void; onOpenSession: (id: string) => void; onTelegramSettings: () => void;
}) {
  const id = useId();
  const dialog = useRef<HTMLDialogElement>(null);
  const [menu, setMenu] = useState(false);
  const [tab, setTab] = useState<'details' | 'runs'>('details');
  useEffect(() => { dialog.current?.showModal(); }, []);
  const { catalog, loading, error, reload } = useAuthoringCatalog(projectId);
  const graphs = catalog?.graphs.map(graphListEntry) ?? [];
  const catalogError = error || catalog?.errors.join(' ');
  const sorted = [...runs].sort((a, b) => Date.parse(b.requestedAt) - Date.parse(a.requestedAt));
  const queued = sorted.filter(run => run.status === 'queued').reverse();
  const history = sorted.filter(run => run.status !== 'queued');

  function runRow(run: TaskRun) {
    const OriginIcon = run.origin === 'schedule' ? CalendarClock : run.origin === 'webhook' ? Webhook : MousePointer2;
    const execution = runDuration(run);
    const userWait = runUserWait(run);
    return <li key={run.id}>
      <div className="task-run-row"><details className="task-run">
        <summary><ChevronDown aria-hidden="true" /><span className="task-run-date"><time dateTime={run.requestedAt}>{formatRunDate(run.requestedAt, timezone)}</time><span className="task-run-origin"><OriginIcon aria-hidden="true" />{triggerOriginLabel[run.origin]}</span></span>{run.status !== 'queued' && <span className={`task-run-status is-${run.status}`}>{run.status === 'waiting' || run.status === 'failed' ? <Badge tone="accent">{runStatusLabel[run.status]}</Badge> : runStatusLabel[run.status]}</span>}{run.status !== 'queued' && <span className="task-run-duration"><span role="img" aria-label={`Execution time: ${execution}`} title={`Execution time: ${execution}`}><Timer aria-hidden="true" />{execution}</span><span role="img" aria-label={`User wait: ${userWait}`} title={`User wait: ${userWait}`}><UserRound aria-hidden="true" />{userWait}</span></span>}</summary>
        {run.status !== 'queued' && run.summary && <p className={run.status === 'failed' ? 'form-error' : ''}>{run.summary}</p>}
        <TaskRunSnapshot run={run} timezone={timezone} />
      </details><Button size="sm" onClick={() => onOpenSession(run.session.id)} aria-label={`Open session for ${formatRunDate(run.requestedAt, timezone)} run`}>Open session</Button></div>
    </li>;
  }

  return <dialog ref={dialog} className="settings-dialog task-details-dialog" aria-labelledby={`${id}-title`} onCancel={onClose} onClose={onClose}>
    <div className="task-details">
      <div className="task-editor-heading"><h2 id={`${id}-title`}>{task.name}</h2><div className="task-details-heading-actions">
          <Menu label="Task actions" role="menu" side="bottom" className="task-actions-menu" open={menu} onOpenChange={setMenu} trigger={props => <IconButton {...props} label="Task actions"><MoreHorizontal /></IconButton>}>
            <MenuItem role="menuitem" onClick={() => { setMenu(false); onEnabledChange(!task.enabled); }}>{task.enabled ? <Pause /> : <Play />}{task.enabled ? 'Disable' : 'Enable'}</MenuItem>
            <MenuItem role="menuitem" className="task-delete-action" onClick={() => { setMenu(false); onDelete(); }}><Trash2 />Delete</MenuItem>
          </Menu>
          <IconButton label="Close task details" onClick={onClose}><X /></IconButton>
        </div>
      </div>
      <div className="task-details-tabs"><TabNav<'details' | 'runs'> label="Task sections" value={tab} onChange={setTab} items={[{ value: 'details', label: 'Details' }, { value: 'runs', label: 'Runs' }]} /></div>
      <section className="task-details-panel" aria-label="Task configuration" hidden={tab !== 'details'}>
        <dl className="task-details-metadata">
          <div><dt>Trigger</dt><dd>{taskTriggerLabel(task)}</dd></div>
          <div><dt>Orchestrator</dt><dd>{harnessLabel[task.harness]} · {task.model}{task.effort && ` · ${task.effort}`}</dd></div>
          <div><dt>Run policy</dt><dd>{task.allowParallelRuns ? 'Parallel' : 'FIFO'} · YOLO on</dd></div>
          {task.trigger === 'schedule' && <div><dt>Missed-run recovery</dt><dd>{task.recoverMissedRuns ? 'Latest missed firing only' : 'Off'}</dd></div>}
          <div><dt>Session folder</dt><dd>{task.sessionFolder}</dd></div>
          <div><dt>Questions</dt><dd>{task.questions === 'telegram' ? 'KLM + Telegram' : 'KLM'}</dd></div>
          <div><dt>Telegram</dt><dd><span className={telegramDeliveryProblem(telegram) ? 'task-link-needed' : ''}>{telegramStatus(telegram)}</span><Button variant="ghost" size="sm" onClick={onTelegramSettings}>{telegram.account ? 'Manage link' : 'Link Telegram'}</Button></dd></div>
          {task.enabled && task.trigger === 'schedule' && nextFiring && <div className="task-next-firing"><dt>Next scheduled run</dt><dd><time dateTime={nextFiring}>{formatRunDate(nextFiring, timezone)}</time></dd></div>}
        </dl>
        <details className="task-details-disclosure"><summary><ChevronDown />Instructions</summary><p className="task-instructions">{task.instructions}</p></details>
        <details className="task-details-disclosure"><summary><ChevronDown />Allowed graphs <span className="muted">{task.allowedGraphIds.length}</span></summary>
          {task.allowedGraphIds.length ? <ul className="task-allowed-graphs">{task.allowedGraphIds.map(graphId => {
            const graph = graphs.find(item => item.id === graphId);
            return <li key={graphId}><Workflow /><span>{graph?.name ?? graphId}</span>{!graph ? <small>{loading ? 'Loading...' : !catalog || catalogError ? 'Catalog unavailable' : 'Unavailable'}</small> : !graph.enabled ? <small>Disabled</small> : null}</li>;
          })}</ul> : <p className="muted">No graphs allowed.</p>}
          {catalogError && <div className="form-error" role="alert">{catalogError}<Button size="sm" disabled={loading} onClick={() => void reload().catch(() => {})}>Retry</Button></div>}
        </details>
        <div className="task-details-actions">
          <Button autoFocus onClick={onEdit}><Pencil />Edit</Button>
          <Button variant="primary" onClick={onRun}><Play />Run now</Button>
        </div>
      </section>
      <section className="task-details-panel" aria-label="Task runs" hidden={tab !== 'runs'}>
        {queued.length > 0 && <section className="task-run-history task-run-queue" aria-labelledby={`${id}-queue`}>
          <h3 id={`${id}-queue`}>Queued runs <small className="muted">{queued.length}</small></h3>
          <ul className="task-runs">{queued.map(runRow)}</ul>
        </section>}
        <section className="task-run-history" aria-labelledby={`${id}-history`}>
          <h3 id={`${id}-history`}>Latest runs</h3>
          {history.length ? <ul className="task-runs">{history.slice(0, 5).map(runRow)}</ul> : <p className="muted">No execution history yet.</p>}
        </section>
      </section>
    </div>
  </dialog>;
}
