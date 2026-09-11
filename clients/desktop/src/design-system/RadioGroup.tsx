import { useId, type ReactNode } from 'react';

export function RadioGroup<T extends string>({ label, value, options, onChange, disabled = false }: {
  label: string;
  value?: T;
  options: readonly { value: T; label: string; icon?: ReactNode }[];
  onChange: (value: T) => void;
  disabled?: boolean;
}) {
  const name = useId();
  return <fieldset className="radio-group" disabled={disabled}>
    <legend>{label}</legend>
    <div className="radio-group-options">{options.map(option => <label key={option.value} className="radio-option">
      <input type="radio" name={name} value={option.value} checked={value === option.value} onChange={() => onChange(option.value)} />
      {option.icon && <span className="radio-option-icon" aria-hidden="true">{option.icon}</span>}
      <span>{option.label}</span>
    </label>)}</div>
  </fieldset>;
}
