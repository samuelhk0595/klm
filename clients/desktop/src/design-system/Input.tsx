import type { InputHTMLAttributes } from 'react';

export function Input({ className = '', type = 'text', ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return <input type={type} className={`input ${className}`} {...props} />;
}
