import { Button } from '../../design-system/Button';
import { AuthoringProject } from '../agents/catalog';
import { AgentsPage } from '../agents/AgentsPage';
import { GraphsPage } from './GraphsPage';
import { useAuthoringCatalog } from './catalog';

export function ProjectAuthoring({ projectId, area }: { projectId: string; area: 'agents' | 'graphs' }) {
  const { catalog, error, reload, invalidate } = useAuthoringCatalog(projectId);
  return <AuthoringProject value={projectId}>
    {error && <div className="storage-error" role="alert">{error} <Button size="sm" onClick={() => void reload().catch(() => {})}>Retry</Button></div>}
    {catalog?.errors.map(message => <div className="storage-error" role="alert" key={message}>{message}</div>)}
    {catalog ? area === 'agents' ? <AgentsPage projectId={projectId} agents={catalog.agents} onChanged={invalidate} /> : <GraphsPage projectId={projectId} graphs={catalog.graphs} agents={catalog.agents} catalogError={!!error || catalog.errors.length > 0} onChanged={invalidate} /> : !error && <p className="small muted" role="status">Loading project files...</p>}
  </AuthoringProject>;
}
