import type { Harness } from '../../engine';
import { modelsForHarness, mockAgentSettings, type Agent } from '../agents/demo';

export type AgentNodeOverrides = { harness?: Harness['id']; model?: string; effort?: string };

export const harnessNames = { opencode: 'OpenCode', codex: 'Codex', pi: 'Pi' };

export function effortLabel(value: string) {
  return value === 'xhigh' ? 'Extra high' : value ? value[0].toUpperCase() + value.slice(1) : '';
}

export function resolveAgentNode(agent: Agent, overrides: AgentNodeOverrides) {
  const harness = overrides.harness ?? agent.defaultHarness ?? mockAgentSettings.defaultHarness;
  const models = modelsForHarness(harness);
  const modelName = overrides.model ?? agent.model ?? mockAgentSettings.model;
  const model = models.find(item => item.name === modelName);
  const effort = overrides.effort ?? agent.effort ?? mockAgentSettings.effort;
  return { harness, models, model, effort: model?.efforts.includes(effort) ? effort : '' };
}
