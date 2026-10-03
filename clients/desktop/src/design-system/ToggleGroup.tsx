import * as RadixToggleGroup from '@radix-ui/react-toggle-group';

export function ToggleGroup<T extends string>({ label, value, options, onValueChange, disabled = false, describedBy, className = '' }: {
  label: string;
  value: readonly T[];
  options: readonly { value: T; label: string; accessibleLabel?: string; disabled?: boolean }[];
  onValueChange: (value: T[]) => void;
  disabled?: boolean;
  describedBy?: string;
  className?: string;
}) {
  return <RadixToggleGroup.Root type="multiple" orientation="horizontal" loop disabled={disabled}
    aria-label={label} aria-describedby={describedBy} className={`toggle-group ${className}`} value={[...value]}
    onValueChange={values => onValueChange(options.filter(option => values.includes(option.value)).map(option => option.value))}>
    {options.map(option => <RadixToggleGroup.Item key={option.value} value={option.value} disabled={option.disabled}
      className="toggle-group-item" aria-label={option.accessibleLabel ?? option.label}>
      {option.label}
    </RadixToggleGroup.Item>)}
  </RadixToggleGroup.Root>;
}
