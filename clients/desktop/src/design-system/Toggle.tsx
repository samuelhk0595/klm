import type { ButtonHTMLAttributes } from 'react';

export function Toggle({ label, checked, onCheckedChange, disabled = false, className = '', ...props }: Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'onClick' | 'children' | 'type' | 'role' | 'aria-checked'> & {
  label: string;
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
}) {
  return <button {...props} type="button" role="switch" aria-checked={checked} disabled={disabled} className={`toggle ${className}`} onClick={() => onCheckedChange(!checked)}>
    <span className="toggle-track" aria-hidden="true"><span className="toggle-thumb" /></span><span>{label}</span>
  </button>;
}
