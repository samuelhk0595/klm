import { useCallback, useEffect, useRef, useState } from 'react';
import { CalendarClock, ClipboardList, MoreHorizontal, MousePointer2, Pencil, Plus, Trash2, Webhook } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { ListTile } from '../../design-system/ListTile';
import { Menu, MenuItem } from '../../design-system/Menu';
import { Select } from '../../design-system/Select';
import { cancelTaskRun, deleteTask, getWebhookBinding, getTask, getTaskRuns, getTasks, runTaskNow, saveTask, setTaskEnabled, type WebhookBinding, type TaskDefinition } from '../../engine';
import { TaskEditorDialog } from './TaskEditorDialog';
import { TaskDetailsDialog } from './TaskDetailsDialog';
import { DeleteTaskDialog } from './DeleteTaskDialog';
import { taskConfig, taskDeleteReason, taskTriggerLabel, unlinkedTelegram, type TaskDraft, type TaskRun } from './types';
import { taskOperationId } from './WebhookPanel';
import { useEngineTime } from '../settings/engineTime';
import './tasks.css';

const failure = (error: unknown) => error instanceof Error ? error.message : 'Could not update Tasks.';

export function TasksPage({ projectId, folders, refreshVersion, initialTaskId, onOpenSession, onTelegramSettings }: {
  projectId: string; folders: string[]; refreshVersion: number;
  initialTaskId?: string;
  onOpenSession: (id: string) => void; onTelegramSettings: () => void;
}) {
  const [tasks, setTasks] = useState<TaskDefinition[]>([]);
  const { settings: engineTime } = useEngineTime();
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState('all');
  const [editing, setEditing] = useState<TaskDefinition | 'new' | null>(null);
  const [detailsId, setDetailsId] = useState(initialTaskId ?? '');
  const [historicalTask, setHistoricalTask] = useState<TaskDefinition | null>(null);
  const [runs, setRuns] = useState<TaskRun[]>([]);
  const [binding, setBinding] = useState<WebhookBinding | null>(null);
  const [detailsError, setDetailsError] = useState('');
  const [detailVersion, setDetailVersion] = useState(0);
  const manualOperation = useRef({ key: '', id: '' });
  const [deleting, setDeleting] = useState<TaskDefinition | null>(null);
  const [deleteError, setDeleteError] = useState('');
  const [menu, setMenu] = useState('');
  const [notice, setNotice] = useState('');
  const requestVersion = useRef(0);
  const mutationPending = useRef(false);
  const details = tasks.find(task => task.id === detailsId) ?? (historicalTask?.id === detailsId ? historicalTask : undefined);
  const detailsRevision = details?.revision;
  useEffect(() => { if (initialTaskId) setDetailsId(initialTaskId); }, [initialTaskId]);
  useEffect(() => {
    if (!detailsId) { setRuns([]); setBinding(null); return; }
    let active = true;
    setDetailsError('');
    async function load() {
      const task = await getTask(projectId, detailsId);
      const [nextRuns, nextBinding] = await Promise.all([getTaskRuns(task), getWebhookBinding(task)]);
      if (active) { setHistoricalTask(task); setRuns(nextRuns); setBinding(nextBinding); }
    }
    void load().catch(cause => { if (active) setDetailsError(failure(cause)); });
    return () => { active = false; };
  }, [projectId, detailsId, detailsRevision, refreshVersion, detailVersion]);
  const reload = useCallback(async () => {
    const version = ++requestVersion.current;
    setLoading(true);
    try {
      const data = await getTasks(projectId);
      if (version === requestVersion.current) { setTasks(data); setLoaded(true); setError(''); }
    } catch (cause) { if (version === requestVersion.current) setError(failure(cause)); }
    finally { if (version === requestVersion.current) setLoading(false); }
  }, [projectId]);
  useEffect(() => { void reload(); return () => { requestVersion.current++; }; }, [reload, refreshVersion]);
  function receive(task: TaskDefinition) {
    requestVersion.current++;
    setLoading(false);
    setTasks(current => task.deletedAt ? current.filter(item => item.id !== task.id) : current.some(item => item.id === task.id) ? current.map(item => item.id === task.id && item.revision <= task.revision ? task : item) : [...current, task]);
  }
  async function save(draft: TaskDraft, operationId: string) {
    const task = await saveTask(projectId, taskConfig(draft), operationId, editing && editing !== 'new' ? editing : undefined);
    receive(task);
    setNotice(editing === 'new' ? 'Task created.' : 'Task updated.');
    setSearch(''); setFilter('all'); setEditing(null);
    void reload();
  }
  async function changeEnabled(task: TaskDefinition, enabled: boolean) {
    if (mutationPending.current) return;
    mutationPending.current = true; setSaving(true); setError('');
    try { receive(await setTaskEnabled(task, enabled)); void reload(); }
    catch (cause) { setError(failure(cause)); }
    finally { mutationPending.current = false; setSaving(false); }
  }
  async function remove() {
    if (!deleting || mutationPending.current) return;
    mutationPending.current = true; setSaving(true); setDeleteError('');
    try { receive(await deleteTask(deleting)); setDeleting(null); setDetailsId(''); setNotice('Task deleted.'); void reload(); }
    catch (cause) { setDeleteError(failure(cause)); }
    finally { mutationPending.current = false; setSaving(false); }
  }
  const query = search.trim().toLowerCase();
  const matches = tasks.filter(task => `${task.name} ${task.description}`.toLowerCase().includes(query) && (filter === 'all' || task.enabled === (filter === 'enabled')));
  const openDelete = (task: TaskDefinition) => { setDeleteError(''); setDeleting(task); };
  async function runNow(body: string, contentType: string) {
    if (!details || mutationPending.current) return;
    const key = JSON.stringify([details.id, contentType, body]);
    if (manualOperation.current.key !== key) manualOperation.current = { key, id: taskOperationId() };
    mutationPending.current = true; setSaving(true); setDetailsError('');
    try { await runTaskNow(details, body, contentType, manualOperation.current.id); manualOperation.current = { key: '', id: '' }; setNotice('Task run accepted.'); setDetailVersion(value => value + 1); }
    catch (cause) { setDetailsError(failure(cause)); }
    finally { mutationPending.current = false; setSaving(false); }
  }
  async function cancelRun(id: string) {
    if (!details || mutationPending.current) return;
    mutationPending.current = true; setSaving(true); setDetailsError('');
    try { await cancelTaskRun(details, id); setDetailVersion(value => value + 1); }
    catch (cause) { setDetailsError(failure(cause)); }
    finally { mutationPending.current = false; setSaving(false); }
  }

  return <section className="tasks-page" aria-label="Tasks">
    <div className="tasks-toolbar">
      <Input type="search" aria-label="Search tasks" placeholder="Search tasks" value={search} onChange={event => setSearch(event.target.value)} />
      <Select label="Task status" value={filter} onChange={event => setFilter(event.target.value)}><option value="all">All tasks</option><option value="enabled">Enabled</option><option value="disabled">Disabled</option></Select>
      <Button variant="primary" disabled={!loaded || !!error} onClick={() => setEditing('new')}><Plus />New task</Button>
    </div>
    {error && <div className="form-error" role="alert">{error}<Button size="sm" disabled={loading} onClick={() => void reload()}>Reload</Button></div>}
    <div className="tasks-notice" role="status">{notice}</div>
    {!loaded ? <p role="status">{loading ? 'Loading tasks...' : 'Tasks unavailable.'}</p> : matches.length ? <ul className="tasks-list">{matches.map(task => <li key={task.id}>
      <ListTile title={task.name} description={task.description} onClick={() => setDetailsId(task.id)}
        metadata={<><span>{task.trigger === 'schedule' ? <CalendarClock /> : task.trigger === 'webhook' ? <Webhook /> : <MousePointer2 />}{taskTriggerLabel(task)}</span><span className="task-enabled-status">{task.enabled ? 'Enabled' : 'Disabled'}</span>{task.runtimeSummary && <span>{task.runtimeSummary}</span>}</>}
        actions={<Menu label={`${task.name} actions`} role="menu" side="bottom" className="task-actions-menu" open={menu === task.id} onOpenChange={open => setMenu(open ? task.id : '')} trigger={props => <IconButton {...props} label={`${task.name} actions`}><MoreHorizontal /></IconButton>}>
          <MenuItem role="menuitem" onClick={() => { setMenu(''); setDetailsId(task.id); setEditing(task); }}><Pencil />Edit</MenuItem>
          <MenuItem role="menuitem" className="task-delete-action" onClick={() => { setMenu(''); openDelete(task); }}><Trash2 />Delete</MenuItem>
        </Menu>} />
    </li>)}</ul> : <div className="tasks-empty"><ClipboardList aria-hidden="true" /><h2>{tasks.length ? 'No tasks found' : 'No tasks yet'}</h2>{tasks.length ? <Button onClick={() => { setSearch(''); setFilter('all'); }}>Clear filters</Button> : <Button disabled={!!error} onClick={() => setEditing('new')}><Plus />New task</Button>}</div>}
    {details && !editing && <TaskDetailsDialog key={details.id} projectId={projectId} task={details} runs={runs.filter(run => run.taskId === details.id)} binding={binding?.taskId === details.id ? binding : null} onBindingChange={setBinding} onCancelRun={id => void cancelRun(id)} onRun={(body, contentType) => void runNow(body, contentType)} telegram={unlinkedTelegram} timezone={engineTime?.timezone ?? ''} error={detailsError || error} busy={saving} onClose={() => setDetailsId('')} onEdit={() => setEditing(details)} onDelete={() => openDelete(details)} onEnabledChange={enabled => void changeEnabled(details, enabled)} onOpenSession={sessionId => { setDetailsId(''); onOpenSession(sessionId); }} onTelegramSettings={onTelegramSettings} />}
    {editing && <TaskEditorDialog projectId={projectId} task={editing === 'new' ? undefined : editing} tasks={tasks} folders={folders} telegram={unlinkedTelegram} onTelegramSettings={onTelegramSettings} onClose={() => setEditing(null)} onSave={save} />}
    {deleting && <DeleteTaskDialog name={deleting.name} reason={taskDeleteReason(runs, deleting.id)} error={deleteError} busy={saving} onClose={() => { if (!saving) setDeleting(null); }} onConfirm={() => void remove()} />}
    {!details && detailsId && detailsError && <p className="form-error" role="alert">{detailsError}</p>}
  </section>;
}
