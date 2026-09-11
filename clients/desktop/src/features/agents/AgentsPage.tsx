import { useState } from 'react';
import { MoreHorizontal, Pause, Play, Plus, Trash2 } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { ListTile } from '../../design-system/ListTile';
import { Menu, MenuItem } from '../../design-system/Menu';
import { AgentEditorDialog } from './AgentEditorDialog';
import { DeleteAgentDialog } from './DeleteAgentDialog';
import { AgentInUseDialog } from './AgentInUseDialog';
import { HarnessIcon } from '../chat/HarnessIcon';
import { agentSlug, mockAgentGraphReferences, mockAgentSettings, type Agent } from './demo';
import './agents.css';

const harnessNames = { opencode: 'OpenCode', pi: 'Pi', codex: 'Codex' };

function AgentMetadata({ agent }: { agent: Agent }) {
  const model = agent.model || mockAgentSettings.model;
  const effortValue = agent.effort || mockAgentSettings.effort;
  const defaultHarness = agent.defaultHarness ?? mockAgentSettings.defaultHarness;
  const effort = effortValue === 'xhigh' ? 'Extra high' : effortValue[0].toUpperCase() + effortValue.slice(1);
  return <>
    <span className="agent-model"><span className="sr-only">Model: </span>{model}</span>
    <span className="agent-effort" title={`Effort level: ${effort}`}><span className="sr-only">Effort level: </span>{effort}</span>
    <span className="agent-harness" title={`Default harness: ${harnessNames[defaultHarness]}`}>
      <span className="sr-only">Default harness: </span><span aria-hidden="true"><HarnessIcon harness={defaultHarness} /></span><span>{harnessNames[defaultHarness]}</span>
    </span>
  </>;
}

export function AgentsPage({ agents, onChange }: { agents: Agent[]; onChange: (agents: Agent[]) => void }) {
  const [search, setSearch] = useState('');
  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [blockedAction, setBlockedAction] = useState<{ agent: Agent; action: 'delete' | 'disable'; graphs: string[] } | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const selectedAgent = agents.find(agent => agent.id === selectedId);
  const deletingAgent = agents.find(agent => agent.id === deletingId);
  const query = search.trim().toLowerCase();
  const matches = agents.filter(agent => `${agent.name} ${agent.description}`.toLowerCase().includes(query));

  function blockIfUsed(agent: Agent, action: 'delete' | 'disable') {
    const graphs = agent.graphReferences ?? mockAgentGraphReferences[agent.id] ?? [];
    if (!graphs.length) return false;
    setBlockedAction({ agent, action, graphs });
    return true;
  }

  return <section className="agents-page" aria-label="Agents">
    <div className="agents-toolbar">
      <Input type="search" aria-label="Search agents" placeholder="Search agents" value={search} onChange={event => { setSearch(event.target.value); setOpenMenu(null); }} />
      <Button variant="primary" onClick={() => setCreating(true)}><Plus />New agent</Button>
    </div>
    {matches.length ? <ul className="agents-list">{matches.map(agent => <li key={agent.id}>
      <ListTile title={agent.name} description={agent.description} onClick={() => { setOpenMenu(null); setSelectedId(agent.id); }}
        metadata={<AgentMetadata agent={agent} />}
        status={!agent.enabled ? <span className="agent-disabled-status">Disabled</span> : undefined}
        actions={<Menu label={`${agent.name} actions`} role="menu" side="bottom" className="agent-actions-menu" open={openMenu === agent.id} onOpenChange={open => setOpenMenu(open ? agent.id : null)}
          trigger={props => <IconButton {...props} label={`${agent.name} actions`}><MoreHorizontal /></IconButton>}>
          <MenuItem role="menuitem" onClick={() => {
            setOpenMenu(null);
            if (agent.enabled && blockIfUsed(agent, 'disable')) return;
            onChange(agents.map(item => item.id === agent.id ? { ...item, enabled: !item.enabled } : item));
          }}>{agent.enabled ? <Pause /> : <Play />}{agent.enabled ? 'Disable' : 'Enable'}</MenuItem>
          <MenuItem role="menuitem" className="agent-delete-action" onClick={() => {
            setOpenMenu(null);
            if (!blockIfUsed(agent, 'delete')) setDeletingId(agent.id);
          }}><Trash2 />Delete</MenuItem>
        </Menu>} />
    </li>)}</ul> : <p className="agents-empty" role="status">{query ? 'No agents found' : 'No agents yet'}</p>}
    {(creating || selectedAgent) && <AgentEditorDialog key={selectedAgent?.id ?? 'new'} agent={selectedAgent} agents={agents}
      onClose={() => { setCreating(false); setSelectedId(null); }}
      onSave={draft => {
        const saved: Agent = { ...selectedAgent, ...draft, id: agentSlug(draft.name), enabled: selectedAgent?.enabled ?? true,
          graphReferences: selectedAgent ? selectedAgent.graphReferences ?? mockAgentGraphReferences[selectedAgent.id] ?? [] : [],
        };
        onChange(selectedAgent ? agents.map(item => item.id === selectedAgent.id ? saved : item) : [...agents, saved]);
        setCreating(false); setSelectedId(null); setSearch('');
      }} />}
    {deletingAgent && <DeleteAgentDialog agentName={deletingAgent.name} onClose={() => setDeletingId(null)} onConfirm={() => {
      setDeletingId(null);
      if (!blockIfUsed(deletingAgent, 'delete')) onChange(agents.filter(agent => agent.id !== deletingAgent.id));
    }} />}
    {blockedAction && <AgentInUseDialog agentName={blockedAction.agent.name} action={blockedAction.action} graphs={blockedAction.graphs} onClose={() => setBlockedAction(null)} />}
  </section>;
}
