import { useEffect, useSyncExternalStore } from 'react';
import { getEngineTimeSettings, type EngineTimeSettings } from '../../engine';

type State = { settings: EngineTimeSettings | null; loading: boolean; error: string };
let state: State = { settings: null, loading: false, error: '' };
let pending: Promise<void> | undefined;
const listeners = new Set<() => void>();
const subscribe = (listener: () => void) => { listeners.add(listener); return () => { listeners.delete(listener); }; };
const snapshot = () => state;
function publish(next: State) { state = next; listeners.forEach(listener => listener()); }

export function receiveEngineTime(settings: EngineTimeSettings) {
  if (!state.settings || settings.revision >= state.settings.revision) publish({ settings, loading: false, error: '' });
}

export function reloadEngineTime(): Promise<void> {
  if (pending) return pending;
  publish({ ...state, loading: true, error: '' });
  pending = getEngineTimeSettings().then(receiveEngineTime, error => {
    publish({ ...state, loading: false, error: error instanceof Error ? error.message : 'Could not load engine time zone.' });
  }).finally(() => { pending = undefined; if (state.loading) publish({ ...state, loading: false }); });
  return pending;
}

export function useEngineTime() {
  const value = useSyncExternalStore(subscribe, snapshot);
  useEffect(() => { if (!state.settings) void reloadEngineTime(); }, []);
  return value;
}

export function formatEngineDate(date: string | number, timezone: string, options: Intl.DateTimeFormatOptions = { dateStyle: 'medium', timeStyle: 'short' }) {
  const instant = new Date(date);
  // Until the engine has a zone, retain an explicit UTC ISO instant rather than
  // silently granting the browser timezone authority over presentation.
  if (!timezone) return instant.toISOString();
  try { return new Intl.DateTimeFormat('en-US', { ...options, timeZone: timezone }).format(instant); }
  catch { return instant.toISOString(); }
}
