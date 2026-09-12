import { useRef, useState, type KeyboardEvent } from 'react';
import { Check, ChevronDown, Workflow } from 'lucide-react';
import { Button } from '../../design-system/Button';
import { Menu, MenuItem } from '../../design-system/Menu';
import type { Graph } from '../graphs/demo';
import '../graphs/graph-running-led.css';

export function GraphPicker({ graphs, value, running = false, onChange }: {
  graphs: Graph[];
  value: string;
  running?: boolean;
  onChange: (graphId: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState('');
  const options = useRef<HTMLDivElement>(null);
  const list = useRef<HTMLDivElement>(null);
  const selected = graphs.find(graph => graph.id === value && graph.enabled);
  const searchable = graphs.length > 5;
  const query = searchable ? search.trim().toLowerCase() : '';
  const matches = graphs.filter(graph => `${graph.name} ${graph.description}`.toLowerCase().includes(query));

  function choose(graphId: string) { onChange(graphId); setOpen(false); }

  function navigate(event: KeyboardEvent<HTMLDivElement>) {
    const inSearch = event.target instanceof HTMLInputElement;
    if (inSearch && event.key === 'Enter') {
      event.preventDefault();
      const first = matches.find(graph => graph.enabled);
      if (first) choose(first.id);
      return;
    }
    const items = [...(options.current?.querySelectorAll<HTMLButtonElement>('button:not(:disabled)') ?? [])];
    if (!items.length) return;
    const index = items.indexOf(document.activeElement as HTMLButtonElement);
    let next: number;
    if (event.key === 'ArrowDown') next = Math.min(index + 1, items.length - 1);
    else if (event.key === 'ArrowUp') next = index <= 0 ? items.length - 1 : index - 1;
    else if (!inSearch && event.key === 'Home') next = 0;
    else if (!inSearch && event.key === 'End') next = items.length - 1;
    else return;
    event.preventDefault(); items[next].focus(); items[next].scrollIntoView({ block: 'nearest' });
  }

  return <Menu label="Graphs" className="model-menu graph-picker-menu" align="start" open={open} onOpenChange={next => { setSearch(''); setOpen(next); }} trigger={props => <Button {...props} variant="ghost" size="sm" className="model-picker-trigger graph-picker-trigger" aria-label={`Choose graph: ${selected?.name ?? 'None'}`} title={selected?.name}>
    {running && selected ? <span className="graph-picker-run-indicator" role="img" aria-label="Graph run in progress" title="Graph run in progress" /> : <Workflow />}<span className="graph-picker-name">{selected?.name ?? 'Select graph'}</span><ChevronDown />
  </Button>}>
    <div ref={options} onKeyDown={navigate}>
      {searchable && <div className="model-menu-search"><input data-menu-autofocus type="search" aria-label="Search graphs" placeholder="Search graphs" value={search} onChange={event => { setSearch(event.target.value); list.current?.scrollTo({ top: 0 }); }} /></div>}
      <div ref={list} className="model-menu-list" aria-label="Available graphs">
        {matches.map(graph => <MenuItem key={graph.id} className="model-menu-option graph-picker-option" selected={selected?.id === graph.id} disabled={!graph.enabled} title={graph.description} onClick={() => choose(graph.id)}>
          <span><span className="model-menu-option-name">{graph.name}</span><small>{graph.enabled ? graph.description : 'Disabled'}</small></span>{selected?.id === graph.id && <Check />}
        </MenuItem>)}
        {!matches.length && <p className="menu-feedback muted" role="status">{query ? 'No matching graphs.' : 'No graphs available.'}</p>}
      </div>
      <div className="model-menu-footer"><MenuItem selected={!selected} onClick={() => choose('')}>None{!selected && <Check />}</MenuItem></div>
    </div>
  </Menu>;
}
