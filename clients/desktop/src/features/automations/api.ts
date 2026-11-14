import { request, type Revision, type Harness, type Topic } from '../../engine';
import type { CapabilitySelection } from '../library/api';

export const DEFAULT_ENGINE_TIME_ZONE = 'America/Sao_Paulo';

export type AutomationSchedule = { time: string; date?: string; recurrence: 'once' | 'daily' | 'every_days' | 'weekly' | 'monthly'; everyDays: number; weekdays: number[]; monthDay: number; anchor?: string };
export type AutomationWebhook = { eventType: 'pull_request' | 'issues' | 'issue_comment'; action: string };
export type Automation = {
  sessionMode?: 'new_per_run' | 'reuse_created' | 'existing'; sessionId?: string; followUpPrompt?: string;
  id: string; projectId: string; revision: Revision; operationId: string; name: string; prompt: string;
  mode: 'webhook' | 'schedule'; enabled: boolean; parallel: boolean; harness: Harness['id']; model: string; effort: string; yolo: boolean;
  webhook?: AutomationWebhook;
  graphs: string[]; selection: CapabilitySelection; topicId: string; schedule: AutomationSchedule; nextScheduledAt?: string; afterSequence: Revision;
};
export type AutomationRow = { automation: Automation; queued: number; active: number; blockedReason?: string };
export type AutomationList = { items: AutomationRow[]; timeZone: string; topics: Topic[]; hasMore?: boolean; nextOffset?: number };
export type AutomationRunLink = { runId: string; sessionId: string; title: string };
export type EngineSettings = { timeZone: string; revision: Revision };
export const automationPath = (projectId: string, id?: string) => `/api/projects/${encodeURIComponent(projectId)}/automations${id ? `/${encodeURIComponent(id)}` : ''}`;
export const getEngineSettings = () => request<EngineSettings>('/api/settings/engine');

export function automationSaveBody(value: Automation, newTopic?: { name: string; color: string }) {
  if ([value.revision, value.afterSequence].some(number => typeof number === 'number' && (!Number.isSafeInteger(number) || number < 0))) {
    throw new Error('Invalid automation revision or sequence. Reload before saving.');
  }
  // Pick the save contract explicitly instead of echoing the read document.
  // Keep identity/sequence fields for existing creation-receipt fingerprints,
  // and preserve decimal values across native draft restore.
  return {
    automation: {
      id: value.id, projectId: value.projectId, afterSequence: String(value.afterSequence),
      revision: String(value.revision), operationId: value.operationId,
      name: value.name, prompt: value.prompt, followUpPrompt: value.followUpPrompt,
      sessionMode: value.sessionMode, sessionId: value.sessionId,
      mode: value.mode, enabled: value.enabled, parallel: value.parallel,
      webhook: value.webhook ? { eventType: value.webhook.eventType, action: value.webhook.action } : undefined,
      harness: value.harness, model: value.model, effort: value.effort, yolo: value.yolo,
      graphs: value.graphs, topicId: value.topicId,
      selection: {
        skills: value.selection.skills,
        mcpConnections: value.selection.mcpConnections.map(connection => ({ connectionId: connection.connectionId, tools: connection.tools })),
      },
      schedule: {
        date: value.schedule.date, time: value.schedule.time, recurrence: value.schedule.recurrence,
        everyDays: value.schedule.everyDays, weekdays: value.schedule.weekdays,
        monthDay: value.schedule.monthDay, anchor: value.schedule.anchor,
      },
    },
    ...(newTopic ? { newTopic: { name: newTopic.name, color: newTopic.color } } : {}),
  };
}

export const saveAutomation = (projectId: string, value: Automation, newTopic?: { name: string; color: string }) =>
  request<Automation>(automationPath(projectId, value.id), value.id ? 'PATCH' : 'POST', automationSaveBody(value, newTopic), 65000);

export function relativeNextRun(at?: string, current = Date.now()) {
  if (!at) return '';
  const hours = Math.max(0, Math.floor((Date.parse(at) - current) / 3600000));
  const days = Math.floor(hours / 24), remainder = hours % 24;
  const parts = [days ? `${days} ${days === 1 ? 'day' : 'days'}` : '', remainder ? `${remainder} ${remainder === 1 ? 'hour' : 'hours'}` : ''].filter(Boolean);
  return `Next run in ${parts.join(' and ') || 'less than an hour'}`;
}
