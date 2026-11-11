import { beforeEach, describe, expect, it, vi } from 'vitest';
import { request } from '../../engine';
import { automationSaveBody, saveAutomation, type Automation } from './api';

vi.mock('../../engine', () => ({ request: vi.fn() }));

const fixture = (): Automation => ({
  id: '', projectId: 'project', revision: '0', operationId: 'operation', name: 'Task',
  prompt: 'Initial task', followUpPrompt: 'Follow-up task', sessionMode: 'new_per_run',
  mode: 'webhook', enabled: true, parallel: false, harness: 'pi', model: 'model', effort: '', yolo: false,
  webhook: { eventType: 'issue_comment', action: 'created' },
  graphs: ['graph'], selection: { skills: ['skill'], mcpConnections: [{ connectionId: 'connection', tools: ['tool'] }] },
  topicId: 'topic', schedule: { recurrence: 'daily', time: '09:00', everyDays: 2, weekdays: [1], monthDay: 1 },
  afterSequence: '0',
});

beforeEach(() => vi.mocked(request).mockReset());

describe('automation save contract', () => {
  it.each(['schedule', 'webhook'] as const)('creates %s configurations with decimal revisions and an atomic new topic', async mode => {
    const value = { ...fixture(), mode };
    await saveAutomation('project', value, { name: 'New topic', color: 'blue' });
    const [path, method, body] = vi.mocked(request).mock.calls[0];
    expect(path).toBe('/api/projects/project/automations');
    expect(method).toBe('POST');
    expect(body).toMatchObject({ automation: { revision: '0', operationId: 'operation', mode }, newTopic: { name: 'New topic', color: 'blue' } });
    expect((body as ReturnType<typeof automationSaveBody>).automation).toMatchObject({ id: '', projectId: 'project', afterSequence: '0' });
  });

  it.each(['new_per_run', 'reuse_created', 'existing'] as const)('edits %s configurations without returning response-only or presentation fields', async sessionMode => {
    const value = Object.assign(fixture(), {
      id: 'saved', sessionMode, sessionId: sessionMode === 'new_per_run' ? '' : 'session',
      revision: '9007199254740993', afterSequence: '9007199254740995',
      creationRequest: 'engine-fingerprint', onceTriggeredAt: 'engine-receipt', futureMetadata: true,
      webhook: { eventType: 'issue_comment' as const, action: 'edited', label: 'UI label' },
    });
    const before = JSON.stringify(value);
    await saveAutomation('project', value);
    const [path, method, body] = vi.mocked(request).mock.calls[0];
    expect(path).toBe('/api/projects/project/automations/saved');
    expect(method).toBe('PATCH');
    const saved = (body as ReturnType<typeof automationSaveBody>).automation;
    expect(saved).toMatchObject({ revision: '9007199254740993', afterSequence: '9007199254740995', sessionMode, sessionId: value.sessionId, prompt: value.prompt, followUpPrompt: value.followUpPrompt });
    expect(saved.webhook).toEqual({ eventType: 'issue_comment', action: 'edited' });
    for (const field of ['creationRequest', 'onceTriggeredAt', 'futureMetadata']) expect(saved).not.toHaveProperty(field);
    expect(JSON.stringify(value)).toBe(before);
  });

  it('supports safe numeric revisions restored from older drafts but refuses lost precision', () => {
    expect(automationSaveBody({ ...fixture(), revision: 7 }).automation.revision).toBe('7');
    expect(() => automationSaveBody({ ...fixture(), revision: 9007199254740992 })).toThrow('Invalid automation revision');
  });
});
