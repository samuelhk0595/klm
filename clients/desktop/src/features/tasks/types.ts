import type { Harness, Session } from '../../engine';
import { formatEngineDate } from '../settings/engineTime';

// Shared configuration and presentation types. Runtime projections will replace
// TaskRun when the gated execution slice is implemented; no fixtures are loaded.
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
export const defaultTaskExecution = { harness: 'opencode' as Harness['id'], model: '', effort: '', allowParallelRuns: false, recoverMissedRuns: false };
export const harnessLabel: Record<Harness['id'], string> = { opencode: 'OpenCode', codex: 'Codex', pi: 'Pi' };
export type TaskDraft = {
  name: string; description: string; instructions: string; sessionFolder: string;
  allowedGraphIds: string[]; enabled: boolean; trigger: TaskTrigger; days: WeekDay[]; time: string;
  questions: 'klm' | 'telegram'; harness: Harness['id']; model: string; effort: string;
  allowParallelRuns: boolean; recoverMissedRuns: boolean;
};
export type TelegramLink = { bot: string; account: string; online: boolean; pairingCode?: string; recipientId?: string };
export const unlinkedTelegram: TelegramLink = { bot: '', account: '', online: false };
export function telegramStatus(link: TelegramLink) {
  return !link.bot || !link.account ? 'Not linked' : !link.online ? 'Offline' : `Linked to @${link.account}`;
}
export function telegramDeliveryProblem(link: TelegramLink) {
  return !link.bot || !link.account ? 'Telegram is not linked. Answer in KLM.' : !link.online ? 'Telegram is offline. Answer in KLM.' : '';
}
export type TaskRunStatus = 'queued' | 'running' | 'waiting' | 'attention' | 'finishing' | 'cancelling' | 'cancelled' | 'succeeded' | 'failed' | 'interrupted';
export type TaskRun = {
  id: string; taskId: string; taskName: string; origin: TaskTrigger; status: TaskRunStatus;
  requestedAt: string; startedAt?: string; endedAt?: string; observedAt: string;
  userWaitMs: number; userWaitStartedAt?: string;
  snapshot: TaskDraft & { engineTimezone: string; yolo: true };
  summary: string; report?: string; session: Session;
  input: { contentType: string; body: string; headers?: Record<string, string> };
};
export const runStatusLabel: Record<TaskRunStatus, string> = { queued: 'Queued', running: 'Running', waiting: 'Waiting for answer', attention: 'Needs attention', finishing: 'Finishing', cancelling: 'Cancelling', cancelled: 'Cancelled', succeeded: 'Succeeded', failed: 'Failed', interrupted: 'Interrupted' };
export const triggerOriginLabel: Record<TaskTrigger, string> = { manual: 'Manual', schedule: 'Schedule', webhook: 'Webhook' };
export const runIsActive = (run: TaskRun) => !['succeeded', 'failed', 'interrupted', 'cancelled'].includes(run.status);
export const formatRunDate = (date: string, timeZone: string) => formatEngineDate(date, timeZone);
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
export const runDuration = (run: TaskRun) => run.startedAt ? durationLabel(elapsedMs(run) - userWaitMs(run)) : '—';
export function taskTriggerLabel(task: TaskDraft) {
  if (task.trigger !== 'schedule') return task.trigger === 'webhook' ? 'Webhook' : 'Manual';
  const selected = weekDays.filter(day => task.days.includes(day.value));
  const label = selected.length === 7 ? 'Daily' : selected.length === 5 && selected.every(day => defaultTaskDays.includes(day.value)) ? 'Weekdays' : selected.map(day => day.short).join(', ');
  return `${label} at ${task.time}`;
}

export function taskConfig(task: TaskDraft): TaskDraft {
  const { name, description, instructions, sessionFolder, allowedGraphIds, enabled, trigger, days, time, questions, harness, model, effort, allowParallelRuns, recoverMissedRuns } = task;
  return { name, description, instructions, sessionFolder, allowedGraphIds, enabled, trigger, days, time, questions, harness, model, effort, allowParallelRuns, recoverMissedRuns };
}
