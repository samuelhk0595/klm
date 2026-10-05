import type { AuthoringCatalog, GraphFile } from './features/graphs/files';
import { resolveEngineURL } from './platform';

export const ENGINE_URL = resolveEngineURL();

export type Project = {
  id: string;
  name: string;
  folder: string;
  icon: string;
  folders: string[];
  archivedFolders: string[];
};

export type Harness = {
  id: 'pi' | 'opencode' | 'codex';
  name: string;
  available: boolean;
  error?: string;
};

export type EngineEvent = {
	consultationId?: string;
  id: string;
  type: 'user' | 'agent_prompt' | 'session_spawn' | 'assistant' | 'reasoning' | 'command' | 'mcp' | 'tool' | 'error' | 'status' | 'consultation' | 'subagent';
  text: string;
  title?: string;
  status?: string;
  favorite?: boolean;
  createdAt: string;
  data?: Record<string, unknown> & { mentions?: FileMention[]; mentionPreparation?: MentionPreparation[] };
};

export type SessionUsage = {
  inputTokens: number | null;
  outputTokens: number | null;
  context: { tokens: number | null; window: number | null } | null;
};

export type QueuedMessage = { id: string; text: string; mode: "queue" | "steer"; status: "queued" | "steering" | "sending" | "paused" | "uncertain"; error?: string; origin?: { sessionId: string; title: string } };

export type Session = {
	taskId?: string;
	taskRunId?: string;
  queue?: QueuedMessage[] | null;
  role?: 'side_agent' | 'subagent' | 'graph_node';
  selectedGraphId?: string;
  graph?: ConversationGraphState;
  parentId?: string;
  sources?: SourceReference[];
  id: string;
  projectId: string;
  title: string;
  workspace: string;
  harness: Harness['id'];
  model?: string;
  yolo?: boolean;
  effort?: string;
  resolvedModel?: string;
  resolvedEffort?: string;
  status: 'idle' | 'running' | 'error';
  archived?: boolean;
  runtimeActive?: boolean;
  events: EngineEvent[];
  history?: { revision: number; startIndex: number; total: number; nextCursor?: string; hasMore: boolean };
  permissions?: PermissionRequest[];
  questions?: QuestionRequest[];
  usage?: SessionUsage;
  createdAt: string;
  updatedAt: string;
};

export type SourceReference = { sessionId: string; messageId: string; passage: string };

export type SessionSummary = Omit<Session, 'events' | 'history' | 'sources'>;
export type EventPage = {
  sessionId: string; revision: number; startIndex: number; endIndex: number;
  total: number; events: EngineEvent[] | null; nextCursor?: string; hasMore: boolean;
};
export type SessionUpdate = {
  kind: 'reset' | 'delta'; revision: number; summary: SessionSummary; total: number;
  changes: { index: number; revision: number; event: EngineEvent }[];
  history?: EventPage;
};
export type SessionResponse = Session | SessionUpdate;
export type EngineSnapshot = Omit<EngineState, 'sessions'> & { sessions: SessionSummary[]; revision?: number };

export type ConversationGraphState = {
  sessionId: string;
  selectedGraphId: string;
  revision: number;
  run: GraphRunProjection | null;
  requests: GraphRequestProjection[];
};
export type GraphRunProjection = {
  id: string; graphId: string; active: boolean;
  status: 'starting' | 'running' | 'ending'; revision: number;
  snapshot: GraphFile;
  activeNodeIds: string[]; completedNodeIds: string[];
  completedChoiceIds: string[]; collectingJoinIds: string[];
};
export type GraphActivationSummary = { id: string; occurrence: number; status: string };
export type GraphNodeActivity = {
  runActive: boolean;
  activations: GraphActivationSummary[];
  activation: (GraphActivationSummary & {
    input: Record<string, string>; events: EngineEvent[]; error?: string;
    createdAt: string; updatedAt: string;
    submission?: { choice: { origin: string; id: string }; payload: Record<string, string>; acceptedAt?: string };
  }) | null;
};
export type GraphRequestProjection = {
  runId: string; graphId: string; nodeId: string; nodeName: string;
  activationId: string; sessionId: string; requestId: string;
  kind: 'permission' | 'question';
  permission?: PermissionRequest; question?: QuestionRequest;
};

// Graph revisions advance independently of the chat's updatedAt timestamp.
export function mergeGraphState(previous: ConversationGraphState | undefined, incoming: ConversationGraphState | undefined) {
  if (!incoming || (previous && incoming.revision <= previous.revision)) return previous;
  return incoming;
}

export const getAuthoringCatalog = (projectId: string) => request<AuthoringCatalog>(`/api/projects/${encodeURIComponent(projectId)}/authoring`);
export const getConversationGraph = (sessionId: string) => request<ConversationGraphState>(`/api/sessions/${encodeURIComponent(sessionId)}/graph`);
export const selectConversationGraph = (sessionId: string, selectedGraphId: string) => request<ConversationGraphState>(`/api/sessions/${encodeURIComponent(sessionId)}/graph`, 'PATCH', { selectedGraphId });

export type ProjectPath = { path: string; kind: 'file' | 'directory' };
// Offsets use UTF-16 units in canonical message text (full @path), not UTF-8 bytes
// or the visible length of the editor's basename-only badges.
export type FileMention = ProjectPath & { id: string; start: number; end: number };
export type MentionPreparation = ProjectPath & { mode: 'native' | 'prepared'; bytes?: number; truncated?: boolean };
export type MessageSubmission = { text: string; mentions: FileMention[]; sources?: SourceReference[] };

export type PermissionDecision = 'once' | 'session' | 'always' | 'reject' | 'deny_project' | 'allow_global' | 'deny_global';
export type QuestionItem = {
  id: string;
  header: string;
  text: string;
  options: { label: string; description: string }[];
  multiple: boolean;
  custom: boolean;
  secret: boolean;
};
export type QuestionRequest = {
  id: string;
  harness: Harness['id'];
  items: QuestionItem[];
  createdAt: string;
  resolving?: boolean;
};
export type PermissionRequest = {
  id: string;
  harness: Harness['id'];
  kind: string;
  title: string;
  description: string;
  patterns: string[];
  details?: Record<string, unknown>;
  command?: string;
  path?: string;
  scopeLabel?: string;
  dangerous?: boolean;
  decisions: PermissionDecision[];
  allowLabel?: string;
  createdAt: string;
  resolving?: boolean;
};

export type EngineState = {
  projects: Project[];
  sessions: Session[];
  harnesses: Harness[];
};

export type ModelOption = { id: string; name: string; provider: string; providerName: string; efforts: string[]; defaultEffort?: string };
export type ModelCatalog = { models: ModelOption[]; connectedProviders: { id: string; name: string; authType: string }[]; defaultModel?: string; defaultEffort?: string; effortLabel: string };
export type SessionMetadata = { sessionId: string; gitBranch?: string };
export type QuotaSnapshot = { source: string; observedAt: string; stale?: boolean; windows: { name: string; usedPercent: number; resetsAt: number }[] };
export type TranscriptionVocabularyEntry = { term: string; note: string };
export type TranscriptionSettings = {
  apiKeyConfigured: boolean;
  model: string;
  language: string;
  vocabulary: TranscriptionVocabularyEntry[];
};
export type TranscriptionSettingsUpdate = {
  apiKey?: string | null;
  model?: string;
  language?: string;
  vocabulary?: TranscriptionVocabularyEntry[];
};

export type TaskDefinition = import('./features/tasks/types').TaskDraft & {
  runtimeSummary?: string;
  id: string; projectId: string; revision: number; operationId: string;
  createdAt: string; updatedAt: string; authorizationId: string; deletedAt?: string;
};
const tasksPath = (projectId: string) => `/api/projects/${encodeURIComponent(projectId)}/tasks`;
export const getTasks = (projectId: string) => request<TaskDefinition[]>(tasksPath(projectId));
export const getTask = (projectId: string, taskId: string) => request<TaskDefinition>(`${tasksPath(projectId)}/${encodeURIComponent(taskId)}`);
export type WebhookBinding = { id: string; taskId: string; revision: number; path: string; publicOrigin: string; signatureHeader: string; deliveryHeader: string; lastReceiptAt: string; lastDisposition: string };
const taskPath = (task: TaskDefinition) => `${tasksPath(task.projectId)}/${encodeURIComponent(task.id)}`;
export const getWebhookBinding = async (task: TaskDefinition) => (await request<{ binding: WebhookBinding | null }>(`${taskPath(task)}/webhook`)).binding;
export const saveWebhookBinding = (task: TaskDefinition, settings: Pick<WebhookBinding, 'publicOrigin' | 'signatureHeader' | 'deliveryHeader'>, operationId: string, revision: number) => request<{ binding: WebhookBinding; secret?: string }>(`${taskPath(task)}/webhook`, 'POST', { ...settings, operationId, revision });
export const rotateWebhookSecret = (task: TaskDefinition) => request<{ secret: string; binding: WebhookBinding }>(`${taskPath(task)}/webhook/secret`, 'POST', {});
export const getTaskRuns = (task: TaskDefinition) => request<import('./features/tasks/types').TaskRun[]>(`${taskPath(task)}/runs`);
export const runTaskNow = (task: TaskDefinition, body: string, contentType: string, operationId: string) => request<{ runId: string; sessionId: string }>(`${taskPath(task)}/runs`, 'POST', { body, contentType, operationId });
export const cancelTaskRun = (task: TaskDefinition, runId: string) => request<{ status: string }>(`${taskPath(task)}/runs/${encodeURIComponent(runId)}/cancel`, 'POST', {});
export const saveTask = (projectId: string, task: import('./features/tasks/types').TaskDraft, operationId: string, previous?: TaskDefinition) => request<TaskDefinition>(`${tasksPath(projectId)}${previous ? `/${encodeURIComponent(previous.id)}` : ''}`, previous ? 'PATCH' : 'POST', { task, operationId, revision: previous?.revision ?? 0 });
export const setTaskEnabled = (task: TaskDefinition, enabled: boolean) => request<TaskDefinition>(`${tasksPath(task.projectId)}/${encodeURIComponent(task.id)}/enabled`, 'PATCH', { revision: task.revision, enabled });
export const deleteTask = (task: TaskDefinition) => request<TaskDefinition>(`${tasksPath(task.projectId)}/${encodeURIComponent(task.id)}`, 'DELETE', { revision: task.revision });
export const getTaskModels = (projectId: string, harness: Harness['id']) => request<ModelCatalog>(`/api/projects/${encodeURIComponent(projectId)}/models/${harness}`, 'GET', undefined, 60000);

export type EngineTimeSettings = { timezone: string; revision: number; updatedAt: string };
export const getEngineTimeSettings = () => request<EngineTimeSettings>('/api/settings/timezone');
export const updateEngineTimeSettings = (timezone: string, revision: number) => request<EngineTimeSettings>('/api/settings/timezone', 'PATCH', { timezone, revision });

export class EngineRequestError extends Error {
  constructor(message: string, public readonly status: number) { super(message); }
}

export async function request<T>(path: string, method = 'GET', body?: unknown, timeoutMs = 15000, signal?: AbortSignal): Promise<T> {
  let response: Response;
  const timeout = method === 'GET' ? AbortSignal.timeout(timeoutMs) : undefined;
  const requestSignal = signal && timeout ? AbortSignal.any([signal, timeout]) : signal ?? timeout;
  try {
    response = await fetch(`${ENGINE_URL}${path}`, {
      method,
      headers: { 'Content-Type': 'application/json' },
      credentials: 'omit',
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: requestSignal,
    });
  } catch (error) {
    if (signal?.aborted && signal.reason instanceof DOMException && signal.reason.name === 'TimeoutError') throw new Error('Engine request timed out. Delivery may be unconfirmed.');
    if (signal?.aborted) throw new DOMException('Request cancelled.', 'AbortError');
    if (timeout?.aborted || (error instanceof DOMException && error.name === 'TimeoutError')) throw new Error('Engine request timed out. Retry.');
    throw new Error(`Cannot connect to the engine at ${ENGINE_URL}. Start the engine and retry.`);
  }
  const result: unknown = await response.json().catch(() => null);
  if (!response.ok) {
    const detail = result && typeof result === 'object' && 'error' in result && typeof result.error === 'string' ? result.error : `Engine request failed (${response.status}).`;
    throw new EngineRequestError(detail, response.status);
  }
  if (result === null) throw new Error('The engine returned an invalid JSON response. Retry the request.');
  return result as T;
}

export async function pickDirectory(signal?: AbortSignal): Promise<string | null> {
  const result = await request<{ path: string | null }>('/api/dialogs/directory', 'POST', {}, 15000, signal);
  return result.path;
}

export const getTranscriptionSettings = () => request<TranscriptionSettings>('/api/transcription/settings');
export const updateTranscriptionSettings = (settings: TranscriptionSettingsUpdate) => request<TranscriptionSettings>('/api/transcription/settings', 'PATCH', settings);

export async function transcribeAudio(audio: Blob, filename: string, signal?: AbortSignal): Promise<string> {
  const form = new FormData();
  form.append('file', audio, filename);
  let response: Response;
  try {
    response = await fetch(`${ENGINE_URL}/api/transcriptions`, {
      method: 'POST',
      credentials: 'omit',
      body: form,
      signal,
    });
  } catch {
    throw new Error(`Cannot connect to the engine at ${ENGINE_URL}. Start the engine and retry.`);
  }
  const result: unknown = await response.json().catch(() => null);
  if (!response.ok) {
    const detail = result && typeof result === 'object' && 'error' in result && typeof result.error === 'string' ? result.error : `Transcription failed (${response.status}).`;
    throw new Error(detail);
  }
  if (!result || typeof result !== 'object' || !('text' in result) || typeof result.text !== 'string') {
    throw new Error('The engine returned an invalid transcription. Retry.');
  }
  return result.text;
}
