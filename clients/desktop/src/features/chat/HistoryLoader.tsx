import { useLayoutEffect, useRef, useState } from 'react';
import { Button } from '../../design-system/Button';
import { request, type EventPage, type Session } from '../../engine';

export function HistoryLoader({ session, onPage }: { session: Session; onPage: (page: EventPage) => void }) {
  const container = useRef<HTMLDivElement>(null);
  const anchor = useRef<{ element: HTMLElement; height: number; top: number } | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  useLayoutEffect(() => {
    if (!anchor.current) return;
    const { element, height, top } = anchor.current;
    element.scrollTop = top + element.scrollHeight - height;
    anchor.current = null;
  }, [session.events]);
  async function load() {
    if (loading || !session.history?.nextCursor) return;
    setLoading(true); setError('');
    try {
      const page = await request<EventPage>(`/api/sessions/${encodeURIComponent(session.id)}/history?cursor=${encodeURIComponent(session.history.nextCursor)}`);
      const element = container.current?.closest<HTMLElement>('.chat-history');
      if (element) anchor.current = { element, height: element.scrollHeight, top: element.scrollTop };
      onPage(page);
    } catch (error) { setError(error instanceof Error ? error.message : 'Could not load history.'); }
    finally { setLoading(false); }
  }
  if (!session.history?.hasMore) return null;
  return <div ref={container}>
    <Button size="sm" disabled={loading} onClick={() => void load()}>{loading ? 'Loading...' : 'Load earlier messages'}</Button>
    {error && <p className="form-error" role="alert">{error}</p>}
  </div>;
}
