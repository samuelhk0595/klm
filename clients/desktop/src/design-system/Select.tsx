import type { SelectHTMLAttributes } from 'react';

export function Select({ label, variant = 'inline', className = '', ...props }: SelectHTMLAttributes<HTMLSelectElement> & { label: string; variant?: 'inline' | 'field' }) {
  return <select aria-label={label} className={`select ${variant === 'field' ? 'select--field' : ''} ${className}`} {...props} />;
}
