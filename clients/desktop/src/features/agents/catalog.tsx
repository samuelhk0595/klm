import { createContext, useContext, useEffect, useState } from 'react';
import { request, type Harness, type ModelCatalog } from '../../engine';
import { modelsForHarness } from './demo';

export const AuthoringProject = createContext<string | undefined>(undefined);
export type AgentModel = { id: string; name: string; provider: string; efforts: string[]; defaultEffort?: string };
const cache = new Map<string, { expires: number; promise: Promise<AgentModel[]> }>();

export function useAgentModels(harness?: Harness['id']): { models: AgentModel[]; loading: boolean; error: string; retry: () => void } {
  const projectId = useContext(AuthoringProject);
  const [snapshot, setSnapshot] = useState<{ key: string; models: AgentModel[]; error: string }>({ key: '', models: [], error: '' });
  const [attempt, setAttempt] = useState(0);
  const key = `${projectId}/${harness}/${attempt}`;
  useEffect(() => {
    if (!projectId || !harness) return;
    let active = true;
    const cacheKey = `${projectId}/${harness}`;
    let entry = cache.get(cacheKey);
    if (!entry || entry.expires < Date.now()) {
      const promise = request<ModelCatalog>(`/api/projects/${encodeURIComponent(projectId)}/models/${harness}`, 'GET', undefined, 55000)
        .then(c => c.models.map(m => ({ id: m.id, name: m.name, provider: m.providerName, efforts: m.efforts ?? [], defaultEffort: m.defaultEffort })));
      entry = { expires: Date.now() + 120000, promise }; cache.set(cacheKey, entry);
    }
    entry.promise.then(models => { if (active) setSnapshot({ key, models, error: '' }); }, error => {
      cache.delete(cacheKey); if (active) setSnapshot({ key, models: [], error: error instanceof Error ? error.message : 'Could not load models.' });
    });
    return () => { active = false; };
  }, [projectId, harness, key]);
  if (!projectId) return { models: harness ? modelsForHarness(harness).map(m => ({ ...m, id: m.name })) : [], loading: false, error: '', retry: () => {} };
  return { models: snapshot.key === key ? snapshot.models : [], error: snapshot.key === key ? snapshot.error : '', loading: !!harness && snapshot.key !== key,
    retry: () => { cache.delete(`${projectId}/${harness}`); setAttempt(n => n + 1); } };
}
