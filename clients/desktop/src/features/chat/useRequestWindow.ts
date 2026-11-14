import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { subscribeScope } from './scopedConnection';
import { scrollRoot } from '../pagination/pagination';
import { captureScrollAnchor } from '../pagination/scrollAnchor';

export type RequestPage<T, C> = { items: T[]; next?: C };

// These mutable queues cannot safely append an old offset/afterId after another
// client removes an item. Rebuild the confirmed page window atomically instead.
export function useRequestWindow<T, C extends string | number>({ identity, sessionId, revision, base, key, read }: {
  identity: string; sessionId: string; revision: string; base: RequestPage<T, C>;
  key: (item: T) => string; read: (cursor: C | undefined, signal: AbortSignal) => Promise<RequestPage<T, C>>;
}) {
  const root = useRef<HTMLDivElement | null>(null);
  const latest = useRef({ identity, base, key, read }); latest.current = { identity, base, key, read };
  const [state, setState] = useState<{ identity: string; page?: RequestPage<T, C>; busy: boolean; loading: boolean; error: string }>({ identity, busy: false, loading: false, error: '' });
  const restore = useRef<(() => void) | undefined>(undefined);
  const controls = useRef({ load: () => {}, retry: () => {}, refresh: () => {}, remove: (_id: string) => {} });
  useLayoutEffect(() => { restore.current?.(); restore.current = undefined; }, [state.page]);
  useEffect(() => {
    const controller = new AbortController(); let page: RequestPage<T, C> | undefined;
    let confirmedPages = 1, wantedPages = 1, busy = false, dirty = false, failed = false;
    let wantedLoading = false;
    let timer: ReturnType<typeof setTimeout> | undefined;
    setState({ identity, busy: false, loading: false, error: '' });
    async function load(extend = false, retry = false) {
      if (controller.signal.aborted || failed && !retry) return;
      if (!page) confirmedPages = Math.max(1, Math.ceil(latest.current.base.items.length / 25));
      if (extend) wantedPages = confirmedPages + 1;
      if (extend || retry) wantedLoading = true;
      if (busy) { dirty = true; return; }
      if (retry) failed = false;
      // Live invalidations reconcile an already visible window silently. Only
      // pagination and explicit recovery need a loading row in the layout.
      const loading = wantedLoading; wantedLoading = false;
      busy = true; setState(old => ({ ...old, busy: true, loading, error: '' }));
      try {
        const items: T[] = []; let next: C | undefined; let consumed = 0;
        do {
          const result = await latest.current.read(next, controller.signal);
          if (result.next !== undefined && (result.next === next || !result.items.length)) throw new Error('Request pagination did not advance. Retry.');
          items.push(...result.items); next = result.next; consumed++;
        } while (next !== undefined && consumed < wantedPages);
        if (controller.signal.aborted) return;
        page = { items: [...new Map(items.map(item => [latest.current.key(item), item])).values()], next };
        confirmedPages = wantedPages = consumed;
        if (root.current) restore.current = captureScrollAnchor(scrollRoot(root.current));
        setState({ identity, page, busy: false, loading: false, error: '' });
      } catch (cause) {
        if (!controller.signal.aborted) { failed = true; setState(old => ({ ...old, busy: false, loading: false, error: cause instanceof Error ? cause.message : 'Could not refresh requests.' })); }
      } finally {
        busy = false;
        if (dirty && !failed && !controller.signal.aborted) { dirty = false; schedule(); }
      }
    }
    const refresh = () => { if ((page || busy) && !document.hidden) void load(); };
    const schedule = () => { if (!timer) timer = setTimeout(() => { timer = undefined; refresh(); }, 150); };
    controls.current = {
      load: () => { void load(true); }, retry: () => { void load(false, true); }, refresh: schedule,
      remove: id => { if (page) { page = { ...page, items: page.items.filter(item => latest.current.key(item) !== id) }; setState(old => ({ ...old, page })); } schedule(); },
    };
    const unsubscribe = subscribeScope('conversation', sessionId, schedule);
    document.addEventListener('visibilitychange', schedule);
    return () => { controller.abort(); clearTimeout(timer); unsubscribe(); document.removeEventListener('visibilitychange', schedule); restore.current = undefined; };
  }, [identity, sessionId]);
  useEffect(() => { controls.current.refresh(); }, [identity, revision]);
  const active = state.identity === identity ? state : undefined;
  return {
    root, items: active?.page?.items ?? base.items, next: active?.page ? active.page.next : base.next,
    busy: active?.busy ?? false, loading: active?.loading ?? false, error: active?.error ?? '', loaded: !!active?.page,
    load: () => { if (latest.current.identity === identity) controls.current.load(); },
    retry: () => { if (latest.current.identity === identity) controls.current.retry(); },
    refresh: () => { if (latest.current.identity === identity) controls.current.refresh(); },
    remove: (id: string) => { if (latest.current.identity === identity) controls.current.remove(id); },
  };
}
