import { useId, useRef, useState } from 'react';
import { SearchSelect } from '../../design-system/SearchSelect';
import { Button } from '../../design-system/Button';
import { updateEngineTimeSettings } from '../../engine';
import { receiveEngineTime, reloadEngineTime, useEngineTime } from './engineTime';

const timezones = [...new Set(['UTC', ...Intl.supportedValuesOf('timeZone')])];

export function EngineTimeSettings() {
  const id = useId();
  const { settings, loading, error } = useEngineTime();
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState('');
  const pending = useRef(false);
  const timezone = settings?.timezone ?? '';
  async function change(zone: string) {
    if (!settings || pending.current || zone === timezone) return;
    pending.current = true; setSaving(true); setSaveError('');
    try { receiveEngineTime(await updateEngineTimeSettings(zone, settings.revision)); }
    catch (error) { setSaveError(error instanceof Error ? error.message : 'Could not save time zone.'); }
    finally { pending.current = false; setSaving(false); }
  }
  const options = [...new Set([...timezones, ...(timezone ? [timezone] : [])])];
  return <section aria-labelledby={`${id}-title`}>
    <div className="settings-panel-title"><h3 id={`${id}-title`}>Engine</h3></div>
    <div className="settings-field">
      <label htmlFor={`${id}-timezone`}>Global time zone</label>
      <SearchSelect id={`${id}-timezone`} label="Global time zone" value={timezone} disabled={!settings || loading || saving} placeholder={loading ? 'Loading...' : 'Select a time zone'} options={options.map(zone => ({ value: zone, label: zone }))} searchPlaceholder="Search time zones" onChange={zone => void change(zone)} />
      {settings && !timezone && <p className="form-error">Could not detect the engine time zone. Select one to continue.</p>}
      {saving && <p role="status">Saving...</p>}
      {(error || saveError) && <p className="form-error" role="alert">{error || saveError}<Button size="sm" disabled={loading || saving} onClick={() => { setSaveError(''); void reloadEngineTime(); }}>Reload</Button></p>}
    </div>
  </section>;
}
