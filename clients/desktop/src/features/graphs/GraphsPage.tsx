import { lazy, Suspense, useState } from 'react';
import { MoreHorizontal, Pause, Play, Plus, Trash2 } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { ListTile } from '../../design-system/ListTile';
import { Menu, MenuItem } from '../../design-system/Menu';
import { GraphDetailsDialog } from './GraphDetailsDialog';
import { DeleteGraphDialog } from './DeleteGraphDialog';
import type { Graph } from './demo';
import type { Agent } from '../agents/demo';
import './graphs.css';

const GraphCanvas = lazy(() => import('./GraphCanvas').then(module => ({ default: module.GraphCanvas })));

export function GraphsPage({ graphs, agents, onChange, creating, onCreate }: { graphs: Graph[]; agents: Agent[]; onChange: (graphs: Graph[]) => void; creating: boolean; onCreate: () => void }) {
  const [search, setSearch] = useState('');
  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const editingGraph = graphs.find(graph => graph.id === editingId);
  const deletingGraph = graphs.find(graph => graph.id === deletingId);
  const query = search.trim().toLowerCase();
  const matches = graphs.filter(graph => `${graph.name} ${graph.description}`.toLowerCase().includes(query));

  if (creating) return <Suspense fallback={<section className="graphs-page" aria-label="Loading graph canvas" aria-busy="true" />}><GraphCanvas agents={agents} /></Suspense>;

  return <section className="graphs-page" aria-label="Graphs">
    <div className="graphs-toolbar"><Input type="search" aria-label="Search graphs" placeholder="Search graphs" value={search} onChange={event => { setSearch(event.target.value); setOpenMenu(null); }} /><Button variant="primary" onClick={onCreate}><Plus />New graph</Button></div>
    {matches.length ? <ul className="graphs-list">{matches.map(graph => <li key={graph.id}>
      <ListTile title={graph.name} description={graph.description} onClick={() => { setOpenMenu(null); setEditingId(graph.id); }}
        status={!graph.enabled ? <span className="graph-disabled-status">Disabled</span> : undefined}
        actions={<Menu label={`${graph.name} actions`} role="menu" side="bottom" className="graph-actions-menu" open={openMenu === graph.id} onOpenChange={open => setOpenMenu(open ? graph.id : null)}
          trigger={props => <IconButton {...props} label={`${graph.name} actions`}><MoreHorizontal /></IconButton>}>
          <MenuItem role="menuitem" onClick={() => { onChange(graphs.map(item => item.id === graph.id ? { ...item, enabled: !item.enabled } : item)); setOpenMenu(null); }}>{graph.enabled ? <Pause /> : <Play />}{graph.enabled ? 'Disable' : 'Enable'}</MenuItem>
          <MenuItem role="menuitem" className="graph-delete-action" onClick={() => { setDeletingId(graph.id); setOpenMenu(null); }}><Trash2 />Delete</MenuItem>
        </Menu>} />
    </li>)}</ul> : <p className="graphs-empty" role="status">{query ? 'No graphs found' : 'No graphs yet'}</p>}
    {editingGraph && <GraphDetailsDialog key={editingGraph.id} graph={editingGraph} graphs={graphs} onClose={() => setEditingId(null)} onSave={(name, description) => {
      const saved: Graph = { ...editingGraph, name, description };
      onChange(graphs.map(item => item.id === editingGraph.id ? saved : item));
      setSearch(''); setEditingId(null);
    }} />}
    {deletingGraph && <DeleteGraphDialog graphName={deletingGraph.name} onClose={() => setDeletingId(null)} onConfirm={() => {
      onChange(graphs.filter(graph => graph.id !== deletingGraph.id));
      setDeletingId(null);
    }} />}
  </section>;
}
