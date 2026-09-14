import { useCallback, useEffect, useSyncExternalStore } from 'react';
import { getAuthoringCatalog } from '../../engine';
import type { AuthoringCatalog } from './files';

type CatalogState = { catalog: AuthoringCatalog | null; error: string; loading: boolean };
type Entry = { state: CatalogState; listeners: Set<() => void>; pending?: Promise<void> };
const entries = new Map<string, Entry>();
function entryFor(projectId: string): Entry {
  let entry = entries.get(projectId);
  if (!entry) {
    entry = { state: { catalog: null, error: '', loading: false }, listeners: new Set() };
    entries.set(projectId, entry);
  }
  return entry;
}
function publish(entry: Entry, state: CatalogState) {
  entry.state = state;
  entry.listeners.forEach(listener => listener());
}
// One catalog/request per project, shared by authoring and the conversation picker.
export function reloadAuthoringCatalog(projectId: string): Promise<void> {
  if (!projectId) return Promise.resolve();
  const entry = entryFor(projectId);
  if (entry.pending) return entry.pending;
  publish(entry, { ...entry.state, loading: true });
  entry.pending = getAuthoringCatalog(projectId).then(catalog => {
    publish(entry, { catalog, error: '', loading: false });
  }, error => {
    publish(entry, { ...entry.state, error: error instanceof Error ? error.message : 'Could not load project files.', loading: false });
    throw error;
  }).finally(() => { entry.pending = undefined; });
  return entry.pending;
}
export function useAuthoringCatalog(projectId: string) {
  const entry = entryFor(projectId);
  const subscribe = useCallback((listener: () => void) => {
    entry.listeners.add(listener);
    return () => { entry.listeners.delete(listener); };
  }, [entry]);
  const getSnapshot = useCallback(() => entry.state, [entry]);
  const state = useSyncExternalStore(subscribe, getSnapshot);
  const reload = useCallback(() => reloadAuthoringCatalog(projectId), [projectId]);
  const invalidate = useCallback(async () => {
    // A pre-mutation request must not satisfy post-CRUD invalidation.
    await entry.pending?.catch(() => {});
    await reloadAuthoringCatalog(projectId);
  }, [entry, projectId]);
  useEffect(() => { void reload().catch(() => {}); }, [reload]);
  return { ...state, reload, invalidate };
}
