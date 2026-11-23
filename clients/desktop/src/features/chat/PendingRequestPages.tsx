import { AutoPage } from '../pagination/AutoPage';
import { request, type PermissionRequest, type QuestionRequest, type Session, type SessionResponse } from '../../engine';
import { PermissionCard } from './PermissionCard';
import { QuestionCard } from './QuestionCard';
import { useSessionNavigation } from './SessionNavigation';
import { useRequestWindow } from './useRequestWindow';

export function usePendingPage<T extends { id: string }>(session: Session, kind: 'permissions' | 'questions' | 'queue', base: T[]) {
  const window = useRequestWindow<T, string>({
    identity: `${session.id}/${kind}`, sessionId: session.id,
    revision: JSON.stringify([session.updatedAt, session.history?.revision, base, session.pendingMore?.[kind]]),
    base: { items: base, next: session.pendingMore?.[kind] ? base.at(-1)?.id ?? '' : undefined }, key: item => item.id,
    read: async (after, signal) => {
      const page = await request<{ items: T[]; hasMore: boolean; nextAfterId: string }>(`/api/sessions/${encodeURIComponent(session.id)}/pending?limit=25&kind=${kind}${after ? `&afterId=${encodeURIComponent(after)}` : ''}`, 'GET', undefined, 15000, signal);
      return { items: page.items, next: page.hasMore ? page.nextAfterId : undefined };
    },
  });
  return { ...window, all: window.items, items: window.loaded ? window.items.filter(item => !base.some(loaded => loaded.id === item.id)) : [], cursor: window.next ?? '', more: window.next !== undefined, resolved: window.remove };
}

export function PendingRequestPages({ session, onSnapshot }: { session: Session; onSnapshot: (value: SessionResponse) => void }) {
  const permissions = usePendingPage(session, 'permissions', session.permissions ?? []);
  const questions = usePendingPage(session, 'questions', session.questions ?? []);
  const navigation = useSessionNavigation();
  return <>
    <div ref={permissions.root}>{permissions.items.map(permission => <div key={permission.id} data-history-anchor={`permission:${permission.id}`}><PermissionCard permission={permission as PermissionRequest} sessionId={session.id} projectName={navigation.projects.get(session.projectId)?.name ?? ''} onResolved={value => { permissions.resolved(permission.id); onSnapshot(value); }} /></div>)}
      <AutoPage hasMore={!!permissions.more} cursor={`${session.id}/permissions/${permissions.cursor}`} busy={permissions.busy} loading={permissions.loading} error={permissions.error} onLoad={permissions.load} onRetry={permissions.retry} /></div>
    <div ref={questions.root}>{questions.items.map(question => <div key={question.id} data-history-anchor={`question:${question.id}`}><QuestionCard question={question as QuestionRequest} sessionId={session.id} onResolved={value => { questions.resolved(question.id); onSnapshot(value); }} /></div>)}
      <AutoPage hasMore={!!questions.more} cursor={`${session.id}/questions/${questions.cursor}`} busy={questions.busy} loading={questions.loading} error={questions.error} onLoad={questions.load} onRetry={questions.retry} /></div>
  </>;
}
