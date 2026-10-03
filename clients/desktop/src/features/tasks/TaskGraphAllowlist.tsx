import { useId, useState } from 'react';
import { Workflow } from 'lucide-react';
import { Button } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { useAuthoringCatalog } from '../graphs/catalog';
import { graphListEntry } from '../graphs/files';

export function TaskGraphAllowlist({ projectId, value, onChange }: {
  projectId: string; value: string[]; onChange: (value: string[]) => void;
}) {
  const id = useId();
  const [search, setSearch] = useState('');
  const { catalog, loading, error, reload } = useAuthoringCatalog(projectId);
  const graphs = catalog?.graphs.map(graphListEntry) ?? [];
  const catalogError = error || catalog?.errors.join(' ') || '';
  const knownIds = new Set(graphs.map(graph => graph.id));
  const options = [
    ...graphs.map(graph => ({ ...graph, status: graph.enabled ? '' : 'Disabled' })),
    ...value.filter(graphId => !knownIds.has(graphId)).map(graphId => ({
      id: graphId, name: graphId, description: '', enabled: false,
      status: !catalog || loading || catalogError ? 'Saved selection' : 'Unavailable',
    })),
  ];
  const query = search.trim().toLowerCase();
  const matches = options.filter(graph => `${graph.name} ${graph.description}`.toLowerCase().includes(query));

  return <fieldset className="task-graph-allowlist" aria-describedby={!value.length ? `${id}-empty` : undefined}>
    <legend>Allowed graphs</legend>
    <div className="task-graph-heading">
      <span id={`${id}-empty`}>{value.length ? 'Selected graphs only' : 'No graphs allowed'}</span>
      {value.length > 0 && <Button variant="ghost" size="sm" onClick={() => onChange([])}>Clear selection</Button>}
    </div>
    {(options.length > 5 || search) && <Input type="search" aria-label="Search allowed graphs" placeholder="Search graphs" value={search} onChange={event => setSearch(event.target.value)} />}
    {matches.length > 0 ? <div className="task-graph-options">{matches.map(graph => {
      const selected = value.includes(graph.id);
      return <label key={graph.id} className="task-graph-option">
        <input type="checkbox" checked={selected} disabled={!selected && (!graph.enabled || loading || !!error)} onChange={event => onChange(event.target.checked ? [...value, graph.id] : value.filter(graphId => graphId !== graph.id))} />
        <Workflow aria-hidden="true" />
        <span className="task-graph-name"><strong>{graph.name}</strong>{graph.description && <small>{graph.description}</small>}</span>
        {graph.status && <span className="task-graph-status">{graph.status}</span>}
      </label>;
    })}</div> : <p className="task-graph-feedback" role="status">{loading || (!catalog && !catalogError) ? 'Loading graphs...' : catalogError ? 'Graph catalog unavailable.' : query ? 'No matching graphs.' : 'No project graphs yet.'}</p>}
    {catalogError && <div className="form-error" role="alert">{catalogError}<Button size="sm" disabled={loading} onClick={() => void reload().catch(() => {})}>Retry</Button></div>}
  </fieldset>;
}
