import { useLayoutEffect, useRef, useState } from 'react';
import { ChevronDown, Clock3 } from 'lucide-react';
import * as RadixToggleGroup from '@radix-ui/react-toggle-group';
import { Button } from './Button';
import { Menu } from './Menu';

const hours = Array.from({ length: 24 }, (_, value) => String(value).padStart(2, '0'));
const minutes = Array.from({ length: 60 }, (_, value) => String(value).padStart(2, '0'));

function TimeColumn({ label, values, value, onChange, autoFocus = false }: {
  label: string; values: string[]; value: string; onChange: (value: string) => void; autoFocus?: boolean;
}) {
  const column = useRef<HTMLDivElement>(null);
  useLayoutEffect(() => {
    const list = column.current;
    const selected = list?.querySelector<HTMLElement>('[data-state="on"]');
    if (list && selected) list.scrollTop = selected.offsetTop - (list.clientHeight - selected.offsetHeight) / 2;
  }, []);

  return <div className="time-picker-column">
    <span className="time-picker-column-label">{label}</span>
    <RadixToggleGroup.Root ref={column} type="single" orientation="vertical" loop aria-label={label}
      className="time-picker-options" value={value} onValueChange={next => { if (next) onChange(next); }}>
      {values.map(option => <RadixToggleGroup.Item key={option} value={option} className="time-picker-option"
        data-menu-autofocus={autoFocus && option === value ? '' : undefined}>{option}</RadixToggleGroup.Item>)}
    </RadixToggleGroup.Root>
  </div>;
}

export function TimePicker({ id, label, value, onValueChange, disabled = false }: {
  id?: string; label: string; value: string; onValueChange: (value: string) => void; disabled?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const valid = /^([01]\d|2[0-3]):[0-5]\d$/.test(value);
  const [hour, minute] = valid ? value.split(':') : ['09', '00'];

  return <div className="time-picker">
    <Menu label={label} role="dialog" side="bottom" className="time-picker-panel" open={open && !disabled} onOpenChange={setOpen}
      trigger={props => <Button {...props} id={id} disabled={disabled} className="time-picker-trigger" aria-label={`${label}: ${valid ? value : 'Select time'}`}>
        <Clock3 aria-hidden="true" /><span>{valid ? value : 'Select time'}</span><ChevronDown aria-hidden="true" />
      </Button>}>
      <div className="time-picker-columns">
        <TimeColumn label="Hour" values={hours} value={hour} autoFocus onChange={next => onValueChange(`${next}:${minute}`)} />
        <span className="time-picker-separator" aria-hidden="true">:</span>
        <TimeColumn label="Minute" values={minutes} value={minute} onChange={next => onValueChange(`${hour}:${next}`)} />
      </div>
      <div className="time-picker-footer"><span>24-hour time</span><Button size="sm" variant="soft" onClick={() => { if (!valid) onValueChange(`${hour}:${minute}`); setOpen(false); }}>Done</Button></div>
    </Menu>
  </div>;
}
