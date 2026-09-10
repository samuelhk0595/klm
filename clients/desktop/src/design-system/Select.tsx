import type { SelectHTMLAttributes } from 'react';

export function Select({ label, className = '', ...props }: SelectHTMLAttributes<HTMLSelectElement> & { label: string }) {
  return <select aria-label={label} className={`select ${className}`} {...props} />;
}
