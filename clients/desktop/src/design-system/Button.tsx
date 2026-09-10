import type { ButtonHTMLAttributes, ReactNode } from 'react';

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'outline' | 'ghost' | 'primary' | 'soft';
  size?: 'sm' | 'md';
};

export function Button({ variant = 'outline', size = 'md', className = '', type = 'button', ...props }: ButtonProps) {
  return <button type={type} className={`button button--${variant} button--${size} ${className}`} {...props} />;
}

export function IconButton({ label, children, className = '', ...props }: Omit<ButtonProps, 'children'> & { label: string; children: ReactNode }) {
  return <Button variant="ghost" className={`icon-button ${className}`} aria-label={label} title={label} {...props}>{children}</Button>;
}
