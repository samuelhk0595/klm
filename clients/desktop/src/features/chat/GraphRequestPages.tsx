import { AutoPage } from '../pagination/AutoPage';
import { request, type ConversationGraphState, type Session } from '../../engine';
import { PermissionCard } from './PermissionCard';
import { QuestionCard } from './QuestionCard';
import { useRequestWindow } from './useRequestWindow';
import type { GraphRequestProjection } from '../../engine';

const key = (item: GraphRequestProjection) => `${item.sessionId}/${item.kind}/${item.requestId}`;

export function GraphRequestPages({ session, projectName, onResolved }: { session: Session; projectName: string; onResolved: () => void }) {
  const page = useRequestWindow<GraphRequestProjection, number>({
    identity: `${session.id}/graph/${session.graph?.run?.id ?? ''}`, sessionId: session.id, revision: String(session.graph?.revision ?? ''), key,
    base: { items: session.graph?.requests ?? [], next: session.graph?.requestsHasMore ? session.graph.requestsNextOffset ?? 25 : undefined },
    read: async (offset, signal) => {
      const result = await request<ConversationGraphState>(`/api/sessions/${encodeURIComponent(session.id)}/graph?limit=25&offset=${offset ?? 0}`, 'GET', undefined, 15000, signal);
      if (result.run?.id !== session.graph?.run?.id) throw new Error('Graph requests changed. Reload the conversation.');
      return { items: result.requests, next: result.requestsHasMore ? result.requestsNextOffset : undefined };
    },
  });
  const base = new Set(session.graph?.requests.map(key));
  const items = page.loaded ? page.items.filter(item => !base.has(key(item))) : [];
  const resolved = (item: GraphRequestProjection) => { page.remove(key(item)); onResolved(); };
  return <div ref={page.root}>{items.map(item => <div key={key(item)} data-history-anchor={`graph-request:${key(item)}`}>{item.permission
    ? <PermissionCard permission={{ ...item.permission, id: item.requestId }} sessionId={item.sessionId} projectName={projectName} origin={`Graph ${item.graphId} · ${item.nodeName}`} onResolved={() => resolved(item)} />
    : item.question ? <QuestionCard question={{ ...item.question, id: item.requestId }} sessionId={item.sessionId} origin={`Graph ${item.graphId} · ${item.nodeName}`} onResolved={() => resolved(item)} /> : null}</div>)}
    <AutoPage hasMore={page.next !== undefined} cursor={`${session.id}/${page.next}`} busy={page.busy} loading={page.loading} error={page.error} onLoad={page.load} onRetry={page.retry} /></div>;
}
