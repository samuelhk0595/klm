import { useState } from 'react';
import type { Harness, Session } from '../../engine';

// UI-only state. No task configuration or Telegram credentials reach the engine.
export type TaskTrigger = 'manual' | 'schedule' | 'webhook';
export const weekDays = [
  { value: 'mon', label: 'Monday', short: 'Mon' },
  { value: 'tue', label: 'Tuesday', short: 'Tue' },
  { value: 'wed', label: 'Wednesday', short: 'Wed' },
  { value: 'thu', label: 'Thursday', short: 'Thu' },
  { value: 'fri', label: 'Friday', short: 'Fri' },
  { value: 'sat', label: 'Saturday', short: 'Sat' },
  { value: 'sun', label: 'Sunday', short: 'Sun' },
] as const;
export type WeekDay = typeof weekDays[number]['value'];
export const defaultTaskDays: WeekDay[] = ['mon', 'tue', 'wed', 'thu', 'fri'];
export const defaultTaskExecution = { harness: 'opencode' as Harness['id'], model: 'GPT-5.6 Terra', effort: 'high', allowParallelRuns: false, recoverMissedRuns: false };
export const harnessLabel: Record<Harness['id'], string> = { opencode: 'OpenCode', codex: 'Codex', pi: 'Pi' };
// A fixture starting zone, not a decision about the engine's initial timezone.
export const initialEngineTimezone = 'UTC';
export type TaskDraft = {
  name: string;
  description: string;
  instructions: string;
  sessionFolder: string;
  allowedGraphIds: string[];
  enabled: boolean;
  trigger: TaskTrigger;
  days: WeekDay[];
  time: string;
  // KLM always shows questions; 'telegram' enables additional Telegram routing.
  questions: 'klm' | 'telegram';
  harness: Harness['id'];
  model: string;
  effort: string;
  allowParallelRuns: boolean;
  recoverMissedRuns: boolean;
};
export type PrototypeTask = TaskDraft & { id: string };
export type TelegramLink = { bot: string; account: string; online: boolean; pairingCode?: string; recipientId?: string };
export function telegramStatus(link: TelegramLink) {
  return !link.bot || !link.account ? 'Not linked' : !link.online ? 'Offline' : `Linked to @${link.account}`;
}
export function telegramDeliveryProblem(link: TelegramLink) {
  return !link.bot || !link.account ? 'Telegram is not linked. Answer in KLM.' : !link.online ? 'Telegram is offline. Answer in KLM.' : '';
}
export type TaskRunStatus = 'queued' | 'running' | 'waiting' | 'succeeded' | 'failed' | 'interrupted';
export type TaskRun = {
  id: string;
  taskId: string;
  taskName: string;
  origin: TaskTrigger;
  status: TaskRunStatus;
  requestedAt: string;
  startedAt?: string;
  endedAt?: string;
  observedAt: string;
  userWaitMs: number;
  userWaitStartedAt?: string;
  // Captured at admission; task edits and later manual chat never modify this.
  snapshot: TaskDraft & { engineTimezone: string; yolo: true };
  summary: string;
  session: Session;
};
export const runStatusLabel: Record<TaskRunStatus, string> = { queued: 'Queued', running: 'Running', waiting: 'Waiting for answer', succeeded: 'Succeeded', failed: 'Failed', interrupted: 'Interrupted' };
export const triggerOriginLabel: Record<TaskTrigger, string> = { manual: 'Manual', schedule: 'Schedule', webhook: 'Webhook' };
export const runIsActive = (run: TaskRun) => runIsActiveStatus(run.status);
export const formatRunDate = (date: string, timeZone: string) => new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeStyle: 'short', timeZone }).format(new Date(date));
export const taskSessionTitle = (run: TaskRun, timeZone: string) => `${run.session.title} · ${formatRunDate(run.requestedAt, timeZone)}`;
export function taskDeleteReason(runs: TaskRun[], taskId: string) {
  return runs.some(run => run.taskId === taskId && runIsActive(run)) ? 'Cannot delete while runs are queued, running, or waiting for an answer.' : '';
}
function elapsedMs(run: TaskRun) {
  return run.startedAt ? Math.max(0, Date.parse(run.endedAt ?? run.observedAt) - Date.parse(run.startedAt)) : 0;
}
function userWaitMs(run: TaskRun) {
  const openWait = run.userWaitStartedAt ? Math.max(0, Date.parse(run.endedAt ?? run.observedAt) - Date.parse(run.userWaitStartedAt)) : 0;
  return Math.min(elapsedMs(run), run.userWaitMs + openWait);
}
function durationLabel(ms: number) {
  const seconds = Math.floor(ms / 1000);
  if (seconds >= 3600) return `${Math.floor(seconds / 3600)}h ${Math.floor(seconds % 3600 / 60)}m`;
  return seconds < 60 ? `${seconds}s` : `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
}
export const runUserWait = (run: TaskRun) => run.startedAt ? durationLabel(userWaitMs(run)) : '—';
export function runDuration(run: TaskRun) {
  if (!run.startedAt) return '—';
  return durationLabel(elapsedMs(run) - userWaitMs(run));
}
function captureTask(task: TaskDraft, engineTimezone: string): TaskRun['snapshot'] {
  return { ...task, days: [...task.days], allowedGraphIds: [...task.allowedGraphIds], engineTimezone, yolo: true };
}

function localId() {
  // getRandomValues also works in the LAN browser's non-secure context.
  return Array.from(crypto.getRandomValues(new Uint32Array(4)), value => value.toString(16)).join('-');
}

const initialTasks: PrototypeTask[] = [
  { ...defaultTaskExecution, id: 'morning-review', name: 'Morning project review', description: 'Review open work and identify the next priorities.', instructions: 'Review the current project, summarize open work, and suggest the next priorities. Ask before making changes.', sessionFolder: 'Tasks', allowedGraphIds: [], enabled: true, trigger: 'schedule', days: defaultTaskDays, time: '09:00', questions: 'klm' },
  { ...defaultTaskExecution, id: 'release-notes', name: 'Prepare release notes', description: 'Draft release notes from recent project changes.', instructions: 'Draft release notes from recent changes. Group them into features, fixes, and improvements, and ask which release range to use.', sessionFolder: 'Tasks', allowedGraphIds: [], enabled: true, trigger: 'manual', days: weekDays.map(day => day.value), time: '09:00', questions: 'telegram' },
  { ...defaultTaskExecution, id: 'incoming-issue', name: 'Triage incoming issues', description: 'Summarize an incoming issue and propose next steps.', instructions: 'Review the incoming issue, identify missing context, and propose next steps.', sessionFolder: 'Tasks', allowedGraphIds: [], enabled: false, trigger: 'webhook', days: weekDays.map(day => day.value), time: '09:00', questions: 'klm' },
  { ...defaultTaskExecution, id: 'release-readiness', name: 'Check release readiness', description: 'Check release prerequisites before preparing a release.', instructions: 'Review the release checklist and report missing prerequisites. Ask before making changes.', sessionFolder: 'Tasks', allowedGraphIds: [], enabled: true, trigger: 'manual', days: defaultTaskDays, time: '09:00', questions: 'klm' },
];

function sampleRun(projectId: string, task: PrototypeTask, key: string, startedAt: Date, seconds: number, status: TaskRunStatus, summary: string, observedAt: string): TaskRun {
  const start = startedAt.toISOString();
  const eventAt = new Date(startedAt.getTime() + seconds * 1000).toISOString();
  const end = status === 'running' ? observedAt : eventAt;
  const id = `task-run:${projectId}:${key}`;
  return {
    id, taskId: task.id, taskName: task.name, origin: task.trigger, status,
    requestedAt: start, startedAt: start, endedAt: runIsActiveStatus(status) ? undefined : end,
    observedAt: runIsActiveStatus(status) ? observedAt : end, summary,
    userWaitMs: 0, userWaitStartedAt: status === 'waiting' ? eventAt : undefined,
    // Automatic fixture runs were accepted before the task was disabled.
    snapshot: captureTask({ ...task, enabled: true }, initialEngineTimezone),
    // A normal, independent top-level chat shape. Never inserted into engine state.
    session: {
      id: `task-session:${projectId}:${key}`, projectId, title: task.name,
      workspace: task.sessionFolder, harness: task.harness, model: task.model, effort: task.effort, yolo: true, status: 'idle', createdAt: start, updatedAt: end,
      events: [
        { id: `${id}:input`, type: 'agent_prompt', text: task.instructions, createdAt: start },
        { id: `${id}:outcome`, type: status === 'failed' || status === 'interrupted' ? 'error' : 'assistant', text: summary, createdAt: end },
      ],
    },
  };
}

function runIsActiveStatus(status: TaskRunStatus) {
  return status === 'queued' || status === 'running' || status === 'waiting';
}

function queuedRun(projectId: string, task: PrototypeTask, key: string, date: string, engineTimezone: string): TaskRun {
  const id = `task-run:${key}`;
  return {
    id, taskId: task.id, taskName: task.name, origin: 'manual', status: 'queued', requestedAt: date, observedAt: date,
    snapshot: captureTask(task, engineTimezone), userWaitMs: 0,
    summary: '',
    session: { id: `task-session:${key}`, projectId, title: task.name, workspace: task.sessionFolder,
      harness: task.harness, model: task.model, effort: task.effort, yolo: true, status: 'idle', createdAt: date, updatedAt: date,
      events: [{ id: `${id}:input`, type: 'agent_prompt', text: task.instructions, createdAt: date }] },
  };
}

function sampleRuns(projectId: string, observedAt: string): TaskRun[] {
  const now = new Date(observedAt);
  const runs: TaskRun[] = [];
  // Bounded fixture generation, not a scheduler: six past weekday snapshots.
  for (let offset = 0; offset < 14 && runs.length < 6; offset++) {
    const date = new Date(now);
    date.setUTCDate(date.getUTCDate() - offset);
    date.setUTCHours(9, 0, 0, 0);
    if (date.getUTCDay() === 0 || date.getUTCDay() === 6 || date.getTime() + 135000 >= now.getTime()) continue;
    runs.push(sampleRun(projectId, initialTasks[0], `review-${runs.length}`, date, 135, 'succeeded', 'Priorities summarized. Two open items need a decision before implementation.', observedAt));
  }
  runs.push(sampleRun(projectId, initialTasks[1], 'notes-waiting', new Date(now.getTime() - 8 * 60000), 75, 'waiting', 'Which release range should the notes cover?', observedAt));
  runs.push(sampleRun(projectId, initialTasks[1], 'notes-completed', new Date(now.getTime() - 7 * 86400000), 210, 'succeeded', 'Release notes drafted with features, fixes, and improvements.', observedAt));
  runs[runs.length - 1].userWaitMs = 60000;
  // Five queued runs exercise waiting-to-start presentation alongside history.
  for (let index = 0; index < 5; index++) {
    runs.push(queuedRun(projectId, initialTasks[1], `${projectId}:notes-queued-${index}`, new Date(now.getTime() - (5 - index) * 60000).toISOString(), initialEngineTimezone));
  }
  runs.push(sampleRun(projectId, initialTasks[2], 'issue-failed', new Date(now.getTime() - 5 * 60000), 18, 'failed', 'The incoming issue did not include a repository or issue URL.', observedAt));
  runs.push(sampleRun(projectId, initialTasks[2], 'issue-interrupted', new Date(now.getTime() - 2 * 86400000), 42, 'interrupted', 'Execution was interrupted. Its history is retained; it has not been restarted.', observedAt));
  // An independent FIFO successor after failure, retained even while the task is disabled.
  runs.push(sampleRun(projectId, initialTasks[2], 'issue-running', new Date(now.getTime() - 4 * 60000), 60, 'running', 'Reviewing the issue context.', observedAt));
  runs.push(sampleRun(projectId, initialTasks[2], 'issue-completed', new Date(now.getTime() - 3 * 86400000), 96, 'succeeded', 'Issue summarized. Reproduction steps and an affected version are still needed.', observedAt));
  runs.push(sampleRun(projectId, initialTasks[3], 'readiness-failed', new Date(now.getTime() - 3 * 3600000), 42, 'failed', 'The release checklist could not be read: docs/release-checklist.md was not found.', observedAt));
  return runs.sort((a, b) => Date.parse(b.requestedAt) - Date.parse(a.requestedAt));
}

function sampleNextFiring(observedAt: string) {
  const now = new Date(observedAt);
  // One read-only fixture timestamp for the unchanged sample schedule.
  for (let offset = 0; offset < 7; offset++) {
    const date = new Date(now);
    date.setUTCDate(date.getUTCDate() + offset);
    date.setUTCHours(9, 0, 0, 0);
    if (date > now && date.getUTCDay() !== 0 && date.getUTCDay() !== 6) return date.toISOString();
  }
}

export function taskTriggerLabel(task: TaskDraft) {
  if (task.trigger !== 'schedule') return task.trigger === 'webhook' ? 'Webhook' : 'Manual';
  const selected = weekDays.filter(day => (task.days ?? defaultTaskDays).includes(day.value));
  const label = selected.length === 7 ? 'Daily' : selected.length === 5 && selected.every(day => defaultTaskDays.includes(day.value)) ? 'Weekdays' : selected.map(day => day.short).join(', ');
  return `${label} at ${task.time}`;
}

export function useTaskPrototype() {
  const [byProject, setByProject] = useState<Record<string, PrototypeTask[]>>({});
  const [observedAt] = useState(() => new Date().toISOString());
  const [runsByProject, setRunsByProject] = useState<Record<string, TaskRun[]>>({});
  const [changedSchedules, setChangedSchedules] = useState<Record<string, string[]>>({});
  const [telegram, setTelegram] = useState<TelegramLink>({ bot: '', account: '', online: false });
  const [timezone, setTimezone] = useState(initialEngineTimezone);
  const [timezoneChanged, setTimezoneChanged] = useState(false);
  const runsFor = (projectId: string) => runsByProject[projectId] ?? sampleRuns(projectId, observedAt);
  return {
    tasksFor: (projectId: string) => byProject[projectId] ?? initialTasks,
    runsFor,
    nextFiringFor: (projectId: string, task: PrototypeTask) => !timezoneChanged && task.enabled && task.trigger === 'schedule' && task.id === 'morning-review' && !changedSchedules[projectId]?.includes(task.id) ? sampleNextFiring(observedAt) : undefined,
    saveTask: (projectId: string, draft: TaskDraft, id?: string) => {
      const taskId = id ?? `task-${localId()}`;
      const previous = (byProject[projectId] ?? initialTasks).find(task => task.id === id);
      if (previous && (previous.trigger !== draft.trigger || previous.time !== draft.time || previous.days.join() !== draft.days.join())) {
        setChangedSchedules(current => ({ ...current, [projectId]: [...new Set([...(current[projectId] ?? []), taskId])] }));
      }
      setByProject(current => {
        const tasks = current[projectId] ?? initialTasks;
        return { ...current, [projectId]: id ? tasks.map(task => task.id === id ? { ...draft, id } : task) : [{ ...draft, id: taskId }, ...tasks] };
      });
    },
    deleteTask: (projectId: string, id: string) => {
      const reason = taskDeleteReason(runsFor(projectId), id);
      if (reason) return reason;
      setByProject(current => ({ ...current, [projectId]: (current[projectId] ?? initialTasks).filter(task => task.id !== id) }));
      return '';
    },
    runTask: (projectId: string, task: PrototypeTask) => {
      if (!(byProject[projectId] ?? initialTasks).some(item => item.id === task.id)) return;
      const date = new Date().toISOString();
      const run = queuedRun(projectId, task, localId(), date, timezone);
      setRunsByProject(current => ({ ...current, [projectId]: [run, ...(current[projectId] ?? sampleRuns(projectId, observedAt))] }));
      return run.session.id;
    },
    updateSession: (projectId: string, sessionId: string, update: (session: Session) => Session) => {
      setRunsByProject(current => ({ ...current, [projectId]: (current[projectId] ?? sampleRuns(projectId, observedAt)).map(run => run.session.id === sessionId ? { ...run, session: update(run.session) } : run) }));
    },
    appendMessage: (projectId: string, sessionId: string, text: string) => {
      const date = new Date().toISOString();
      setRunsByProject(current => ({ ...current, [projectId]: (current[projectId] ?? sampleRuns(projectId, observedAt)).map(run => run.session.id === sessionId ? { ...run, session: { ...run.session, updatedAt: date, events: [...run.session.events, { id: `task-message:${localId()}`, type: 'user', text, createdAt: date }] } } : run) }));
    },
    telegram, setTelegram, timezone,
    setTimezone: (zone: string) => { if (zone !== timezone) { setTimezone(zone); setTimezoneChanged(true); } },
  };
}
