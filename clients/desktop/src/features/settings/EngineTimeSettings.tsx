import { useId } from 'react';
import { SearchSelect } from '../../design-system/SearchSelect';

const timezones = [...new Set(['UTC', ...Intl.supportedValuesOf('timeZone')])];

export function EngineTimeSettings({ timezone, onChange }: { timezone: string; onChange: (timezone: string) => void }) {
  const id = useId();
  return <section aria-labelledby={`${id}-title`}>
    <div className="settings-panel-title"><h3 id={`${id}-title`}>Engine</h3></div>
    <div className="settings-field">
      <label htmlFor={`${id}-timezone`}>Global time zone</label>
      <SearchSelect id={`${id}-timezone`} label="Global time zone" value={timezone} options={timezones.map(zone => ({ value: zone, label: zone }))} searchPlaceholder="Search time zones" onChange={onChange} />
    </div>
  </section>;
}
