import { useEffect, useId } from 'react';
import { RotateCcw, X } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { RadioGroup } from '../../design-system/RadioGroup';
import { SearchSelect } from '../../design-system/SearchSelect';
import { Slider } from '../../design-system/Slider';
import { HarnessIcon } from '../chat/HarnessIcon';
import { mockAgentSettings, type Agent } from '../agents/demo';
import type { Harness } from '../../engine';
import { effortLabel, resolveAgentNode, type AgentNodeOverrides } from './agent-node';

export function AgentNodePanel({ agents, name, agentId, overrides, onNameChange, onSelectAgent, onOverridesChange, onClose }: {
  agents: Agent[];
  name: string;
  agentId: string;
  overrides: AgentNodeOverrides;
  onNameChange: (name: string) => void;
  onSelectAgent: (id: string) => void;
  onOverridesChange: (overrides: AgentNodeOverrides) => void;
  onClose: () => void;
}) {
  const id = useId();
  const available = agents.filter(agent => agent.enabled);
  const agent = available.find(item => item.id === agentId);
  const settings = agent ? resolveAgentNode(agent, overrides) : undefined;
  const hasOverrides = Object.values(overrides).some(value => value !== undefined);
  useEffect(() => { document.getElementById(`${id}-name`)?.focus(); }, [id]);

  return <aside className="graph-agent-panel nodrag nopan nowheel" role="dialog" aria-labelledby={`${id}-title`} onKeyDown={event => {
    if (event.key === 'Escape' && !event.defaultPrevented) { event.stopPropagation(); onClose(); }
  }}>
    <div className="graph-agent-panel-header"><h2 id={`${id}-title`}>AI agent</h2><IconButton label="Close node settings" onClick={onClose}><X /></IconButton></div>
    <div className="graph-agent-panel-field"><label htmlFor={`${id}-name`}>Name</label><Input id={`${id}-name`} maxLength={80} placeholder="Node name" value={name} onChange={event => onNameChange(event.target.value)} /></div>
    <div className="graph-agent-panel-field"><label htmlFor={`${id}-agent`}>Agent</label><SearchSelect id={`${id}-agent`} label="Agent" value={agentId} options={available.map(item => ({ value: item.id, label: item.name, description: item.description }))} onChange={onSelectAgent} placeholder="Select agent" searchPlaceholder="Search agents" emptyMessage={available.length ? 'No agents found' : 'No enabled agents available'} /></div>
    {agent && settings && <div className="graph-agent-overrides">
      <div className="graph-agent-overrides-header"><h3>Overrides</h3>{hasOverrides && <Button size="sm" variant="ghost" onClick={() => onOverridesChange({})}><RotateCcw />Reset</Button>}</div>
      <RadioGroup<Harness['id']> label="Harness" value={settings.harness} options={[
        { value: 'opencode', label: 'OpenCode', icon: <HarnessIcon harness="opencode" /> },
        { value: 'codex', label: 'Codex', icon: <HarnessIcon harness="codex" /> },
        { value: 'pi', label: 'Pi', icon: <HarnessIcon harness="pi" /> },
      ]} onChange={harness => onOverridesChange({ ...overrides, harness: harness === (agent.defaultHarness ?? mockAgentSettings.defaultHarness) ? undefined : harness })} />
      <div className="graph-agent-panel-field"><label htmlFor={`${id}-model`}>Model</label><SearchSelect key={`${agent.id}:${settings.harness}`} id={`${id}-model`} label="Model" value={settings.model?.name ?? ''} options={settings.models.map(model => ({ value: model.name, label: model.name, description: model.provider }))} placeholder="Select model" searchPlaceholder="Search models" emptyMessage={settings.models.length ? 'No models found' : 'No models available from connected providers'} onChange={name => {
        const model = settings.models.find(item => item.name === name);
        if (!model) return;
        const baseEffort = agent.effort ?? mockAgentSettings.effort;
        const effort = model.efforts.includes(settings.effort) ? settings.effort : model.efforts.includes(baseEffort) ? baseEffort : model.efforts[0];
        onOverridesChange({ ...overrides, model: name === (agent.model ?? mockAgentSettings.model) ? undefined : name, effort: effort === baseEffort ? undefined : effort });
      }} />{!settings.model && <p className="form-error" role="status">Select a model available in this harness.</p>}</div>
      {settings.model && <Slider label="Effort level" min={0} max={Math.max(0, settings.model.efforts.length - 1)} value={Math.max(0, settings.model.efforts.indexOf(settings.effort))} valueText={effortLabel(settings.effort)} marks={settings.model.efforts.map((effort, index) => ({ value: index, label: effortLabel(effort) }))} onValueChange={index => {
        const effort = settings.model?.efforts[index];
        if (effort) onOverridesChange({ ...overrides, effort: effort === (agent.effort ?? mockAgentSettings.effort) ? undefined : effort });
      }} />}
    </div>}
  </aside>;
}
