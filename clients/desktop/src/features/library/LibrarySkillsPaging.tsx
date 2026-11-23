import { useEffect, useRef, useState, type Dispatch, type SetStateAction } from 'react';
import { AutoPage } from '../pagination/AutoPage';
import { request } from '../../engine';
import { currentEngineIdentity, engineIdentityGeneration } from '../../engineIdentity';
import type { LibraryInventory } from './api';

export function LibrarySkillsPaging({ inventory, onPage, query = inventory.search ?? '' }: { inventory: LibraryInventory; onPage: Dispatch<SetStateAction<LibraryInventory | undefined>>; query?: string }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const requestRef = useRef<AbortController | null>(null);
  const search = query.trim().toLowerCase();
  const matches = (inventory.search ?? '').trim().toLowerCase() === search;
  const scope = JSON.stringify([currentEngineIdentity(), engineIdentityGeneration(), inventory.skillsRoot, inventory.clientGeneration, search]);
  const latest = useRef({ scope, inventory }); latest.current = { scope, inventory };
  useEffect(() => {
    requestRef.current?.abort(); requestRef.current = null; setBusy(false); setError('');
    return () => { requestRef.current?.abort(); };
  }, [scope]);
  async function load(offset: number) {
    if (requestRef.current || !matches) return;
    const controller = new AbortController(); requestRef.current = controller; setBusy(true); setError('');
    try {
      const page = await request<LibraryInventory>(`/api/library?limit=25&skillsOffset=${offset}&search=${encodeURIComponent(inventory.search ?? '')}`, 'GET', undefined, 20000, controller.signal);
      if (!controller.signal.aborted && latest.current.scope === scope && latest.current.inventory === inventory) {
        if (page.skillsHasMore && (page.skillsNextOffset ?? 0) <= offset) throw new Error('Skill pagination did not advance. Retry.');
        onPage(previous => previous === inventory ? { ...previous, skills: [...new Map([...previous.skills, ...page.skills].map(item => [item.id, item])).values()], skillsHasMore: page.skillsHasMore, skillsNextOffset: page.skillsNextOffset } : previous);
      }
    } catch (error) { if (!controller.signal.aborted) setError(error instanceof Error ? error.message : 'Could not load skills.'); }
    finally { if (requestRef.current === controller) { requestRef.current = null; if (!controller.signal.aborted) setBusy(false); } }
  }
  return <AutoPage hasMore={!!inventory.skillsHasMore && matches} cursor={`${scope}/${inventory.skillsNextOffset}`} busy={busy} error={error} onLoad={() => void load(inventory.skillsNextOffset ?? 0)} />;
}
