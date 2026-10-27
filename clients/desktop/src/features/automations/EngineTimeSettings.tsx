import { useEffect, useState } from 'react';
import { Button } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { request } from '../../engine';
import { DEFAULT_ENGINE_TIME_ZONE, getEngineSettings, type EngineSettings } from './api';

export function EngineTimeSettings() {
  const [settings, setSettings] = useState<EngineSettings>();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [saved, setSaved] = useState(false);
  useEffect(() => { getEngineSettings().then(setSettings, error => setError(error instanceof Error ? error.message : 'Could not load engine settings.')); }, []);
  return <form className="settings-form" onSubmit={async event => {
    event.preventDefault(); if (!settings || busy) return; setBusy(true); setError(''); setSaved(false);
    try { setSettings(await request<EngineSettings>('/api/settings/engine', 'PATCH', settings)); setSaved(true); window.dispatchEvent(new Event('engine-time-zone-changed')); }
    catch (error) { setError(error instanceof Error ? error.message : 'Could not save engine settings.'); } finally { setBusy(false); }
  }}><h4>Engine time zone</h4><label htmlFor="engine-time-zone">Time zone</label><Input id="engine-time-zone" list="engine-time-zones" required disabled={!settings || busy} value={settings?.timeZone ?? DEFAULT_ENGINE_TIME_ZONE} placeholder={DEFAULT_ENGINE_TIME_ZONE} onChange={event => { if (settings) setSettings({ ...settings, timeZone: event.target.value }); setSaved(false); }} /><datalist id="engine-time-zones">{[DEFAULT_ENGINE_TIME_ZONE, 'UTC', 'America/New_York', 'America/Los_Angeles', 'Europe/London', 'Europe/Paris', 'Asia/Tokyo', 'Australia/Sydney'].map(zone => <option key={zone} value={zone} />)}</datalist>
    {error && <p className="form-error" role="alert">{error}</p>}{saved && <p role="status">Saved</p>}<div className="dialog-actions"><Button type="submit" disabled={!settings || busy}>Save time zone</Button></div>
  </form>;
}
