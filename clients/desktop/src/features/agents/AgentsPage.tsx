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
import { request } from '../../engine';
import { authoringPath, type AgentFile } from '../graphs/files';
import { effortLabel } from '../graphs/agent-node';
import './agents.css';

export function AgentsPage({ projectId, agents, onChanged }: { projectId: string; agents: AgentFile[]; onChanged: () => Promise<void> }) {
  const [search, setSearch] = useState('');
  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const [editing, setEditing] = useState<AgentFile | 'new' | null>(null);
  const [deleting, setDeleting] = useState<AgentFile | null>(null);
  const [blocked, setBlocked] = useState<AgentFile | null>(null);
  const [blockedAction, setBlockedAction] = useState<'delete' | 'disable'>('disable');
  const [error, setError] = useState('');
  const [pending, setPending] = useState(false);
  const base = authoringPath(projectId);
  const query = search.trim().toLowerCase();
  const matches = agents.filter(agent => `${agent.name} ${agent.description}`.toLowerCase().includes(query));
  const definition = (agent: AgentFile) => ({ name: agent.name, description: agent.description, enabled: agent.enabled, defaultHarness: agent.defaultHarness, model: agent.model, effort: agent.effort, prompt: agent.prompt });

  async function toggle(agent: AgentFile) {
    setOpenMenu(null);
    if (agent.enabled && agent.graphReferences?.length) { setBlockedAction('disable'); setBlocked(agent); return; }
    setPending(true); setError('');
    try { await request(`${base}/agents/${encodeURIComponent(agent.id)}`, 'PATCH', { agent: { ...definition(agent), enabled: !agent.enabled }, revision: agent.revision }); await onChanged(); }
    catch (e) { setError(e instanceof Error ? e.message : 'Could not update agent.'); }
    finally { setPending(false); }
  }
  return <section className="agents-page" aria-label="Agents">
    <div className="agents-toolbar"><Input type="search" aria-label="Search agents" placeholder="Search agents" value={search} onChange={e => setSearch(e.target.value)} /><Button variant="primary" disabled={pending} onClick={() => setEditing('new')}><Plus />New agent</Button></div>
    {error && <p className="form-error" role="alert">{error}</p>}
    {matches.length ? <ul className="agents-list">{matches.map(agent => <li key={agent.id}>
      <ListTile title={agent.name} description={agent.description} onClick={() => { if (!pending) setEditing(agent); }}
        metadata={<><span className="agent-model">{agent.model}</span>{agent.effort && <span className="agent-effort">{effortLabel(agent.effort)}</span>}{agent.defaultHarness && <span className="agent-harness"><HarnessIcon harness={agent.defaultHarness} />{agent.defaultHarness}</span>}</>}
        status={!agent.enabled ? <span className="agent-disabled-status">Disabled</span> : undefined}
        actions={<Menu label={`${agent.name} actions`} role="menu" side="bottom" className="agent-actions-menu" open={openMenu === agent.id} onOpenChange={open => setOpenMenu(open ? agent.id : null)} trigger={props => <IconButton {...props} disabled={pending} label={`${agent.name} actions`}><MoreHorizontal /></IconButton>}>
          <MenuItem role="menuitem" onClick={() => void toggle(agent)}>{agent.enabled ? <Pause /> : <Play />}{agent.enabled ? 'Disable' : 'Enable'}</MenuItem>
          <MenuItem role="menuitem" className="agent-delete-action" onClick={() => { setOpenMenu(null); if (agent.graphReferences?.length) { setBlockedAction('delete'); setBlocked(agent); } else setDeleting(agent); }}><Trash2 />Delete</MenuItem>
        </Menu>} />
    </li>)}</ul> : <p className="agents-empty" role="status">{query ? 'No agents found' : 'No agents yet'}</p>}
    {editing && <AgentEditorDialog agent={editing === 'new' ? undefined : editing} agents={agents} onClose={() => setEditing(null)} onSave={async draft => {
      const old = editing === 'new' ? undefined : editing;
      await request(`${base}/agents${old ? `/${encodeURIComponent(old.id)}` : ''}`, old ? 'PATCH' : 'POST', { agent: { ...draft, enabled: old?.enabled ?? true }, revision: old?.revision ?? '' });
      await onChanged(); setEditing(null); setSearch('');
    }} />}
    {deleting && <DeleteAgentDialog agentName={deleting.name} onClose={() => { if (!pending) setDeleting(null); }} onConfirm={() => {
      if (pending) return;
      setPending(true); setError('');
      void request(`${base}/agents/${encodeURIComponent(deleting.id)}`, 'DELETE', { revision: deleting.revision })
        .then(async () => { setDeleting(null); await onChanged(); })
        .catch(e => { setDeleting(null); setError(e instanceof Error ? e.message : 'Could not delete agent.'); })
        .finally(() => setPending(false));
    }} />}
    {blocked && <AgentInUseDialog agentName={blocked.name} action={blockedAction} graphs={blocked.graphReferences ?? []} onClose={() => setBlocked(null)} />}
  </section>;
}
