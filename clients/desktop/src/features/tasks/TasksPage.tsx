import { useState } from 'react';
import { CalendarClock, ClipboardList, MoreHorizontal, MousePointer2, Pencil, Plus, Trash2, Webhook } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Badge } from '../../design-system/Badge';
import { Input } from '../../design-system/Input';
import { ListTile } from '../../design-system/ListTile';
import { Menu, MenuItem } from '../../design-system/Menu';
import { Select } from '../../design-system/Select';
import { TaskEditorDialog } from './TaskEditorDialog';
import { TaskDetailsDialog } from './TaskDetailsDialog';
import { DeleteTaskDialog } from './DeleteTaskDialog';
import { runIsActive, runStatusLabel, taskDeleteReason, taskTriggerLabel, type PrototypeTask, type TaskDraft, type TaskRun, type TelegramLink } from './prototype';
import './tasks.css';

export function TasksPage({ projectId, tasks, runs, folders, telegram, timezone, initialTaskId, nextFiringFor, onSave, onDelete, onRun, onOpenSession, onTelegramSettings }: {
  projectId: string;
  timezone: string;
  tasks: PrototypeTask[];
  runs: TaskRun[];
  initialTaskId?: string;
  nextFiringFor: (task: PrototypeTask) => string | undefined;
  folders: string[];
  telegram: TelegramLink;
  onSave: (draft: TaskDraft, id?: string) => void;
  onDelete: (id: string) => string;
  onRun: (task: PrototypeTask) => void;
  onOpenSession: (id: string) => void;
  onTelegramSettings: () => void;
}) {
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState('all');
  const [editing, setEditing] = useState<PrototypeTask | 'new' | null>(null);
  const [detailsId, setDetailsId] = useState(initialTaskId ?? '');
  const details = tasks.find(task => task.id === detailsId);
  const [deleting, setDeleting] = useState<PrototypeTask | null>(null);
  const [menu, setMenu] = useState('');
  const [notice, setNotice] = useState('');
  const query = search.trim().toLowerCase();
  const matches = tasks.filter(task => `${task.name} ${task.description}`.toLowerCase().includes(query) && (filter === 'all' || task.enabled === (filter === 'enabled')));

  return <section className="tasks-page" aria-label="Tasks">
    <div className="tasks-toolbar">
      <Input type="search" aria-label="Search tasks" placeholder="Search tasks" value={search} onChange={event => setSearch(event.target.value)} />
      <Select label="Task status" value={filter} onChange={event => setFilter(event.target.value)}><option value="all">All tasks</option><option value="enabled">Enabled</option><option value="disabled">Disabled</option></Select>
      <Button variant="primary" onClick={() => setEditing('new')}><Plus />New task</Button>
    </div>
    <div className="tasks-notice" role="status">{notice}</div>
    {matches.length ? <ul className="tasks-list">{matches.map(task => {
      const taskRuns = runs.filter(run => run.taskId === task.id).sort((a, b) => Date.parse(b.requestedAt) - Date.parse(a.requestedAt));
      const latest = taskRuns.find(run => run.status === 'waiting' || run.status === 'running') ?? taskRuns.find(runIsActive) ?? taskRuns[0];
      const queued = taskRuns.filter(run => run.status === 'queued').length;
      return <li key={task.id}>
      <ListTile title={task.name} description={task.description} onClick={() => setDetailsId(task.id)}
        metadata={<><span>{task.trigger === 'schedule' ? <CalendarClock /> : task.trigger === 'webhook' ? <Webhook /> : <MousePointer2 />}{taskTriggerLabel(task)}</span>{latest && latest.status !== 'succeeded' && <span className={`task-run-status is-${latest.status}`}>{latest.status === 'running' || latest.status === 'waiting' || latest.status === 'failed' ? <Badge tone="accent">{latest.status === 'failed' ? 'Last run · Failed' : runStatusLabel[latest.status]}</Badge> : <>{runIsActive(latest) ? '' : 'Last run · '}{runStatusLabel[latest.status]}</>}</span>}{queued > 0 && <span>{queued} queued</span>}<span className="task-enabled-status">{task.enabled ? 'Enabled' : 'Disabled'}</span></>}
        actions={<Menu label={`${task.name} actions`} role="menu" side="bottom" className="task-actions-menu" open={menu === task.id} onOpenChange={open => setMenu(open ? task.id : '')} trigger={props => <IconButton {...props} label={`${task.name} actions`}><MoreHorizontal /></IconButton>}>
          <MenuItem role="menuitem" onClick={() => { setMenu(''); setDetailsId(task.id); setEditing(task); }}><Pencil />Edit</MenuItem>
          <MenuItem role="menuitem" className="task-delete-action" onClick={() => { setMenu(''); setDeleting(task); }}><Trash2 />Delete</MenuItem>
        </Menu>} />
    </li>; })}</ul> : <div className="tasks-empty"><ClipboardList aria-hidden="true" /><h2>{tasks.length ? 'No tasks found' : 'No tasks yet'}</h2>{tasks.length ? <Button onClick={() => { setSearch(''); setFilter('all'); }}>Clear filters</Button> : <Button onClick={() => setEditing('new')}><Plus />New task</Button>}</div>}
    {details && !editing && <TaskDetailsDialog projectId={projectId} task={details} runs={runs.filter(run => run.taskId === details.id)} nextFiring={nextFiringFor(details)} telegram={telegram} timezone={timezone} onClose={() => setDetailsId('')} onEdit={() => setEditing(details)} onDelete={() => setDeleting(details)} onEnabledChange={enabled => onSave({ ...details, enabled }, details.id)} onRun={() => onRun(details)} onOpenSession={sessionId => { setDetailsId(''); onOpenSession(sessionId); }} onTelegramSettings={onTelegramSettings} />}
    {editing && <TaskEditorDialog projectId={projectId} task={editing === 'new' ? undefined : editing} tasks={tasks} folders={folders} telegram={telegram} onTelegramSettings={onTelegramSettings} onClose={() => setEditing(null)} onSave={draft => { onSave(draft, editing === 'new' ? undefined : editing.id); setNotice(editing === 'new' ? 'Task created.' : 'Task updated.'); setSearch(''); setFilter('all'); setEditing(null); }} />}
    {deleting && <DeleteTaskDialog name={deleting.name} reason={taskDeleteReason(runs, deleting.id)} onClose={() => setDeleting(null)} onConfirm={() => { const reason = onDelete(deleting.id); setNotice(reason || 'Task deleted.'); if (!reason) setDeleting(null); }} />}
  </section>;
}
