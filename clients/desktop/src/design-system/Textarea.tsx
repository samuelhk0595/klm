import type { TextareaHTMLAttributes } from 'react';

export function Textarea({ className = '', rows = 6, ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea rows={rows} className={`input textarea ${className}`} {...props} />;
}
