import type { ReactNode } from 'react';

export function Badge({ children, tone = 'neutral' }: { children: ReactNode; tone?: 'neutral' | 'accent' }) {
  return <span className={`badge badge--${tone}`}>{children}</span>;
}

export function StatusIndicator({ label, status = 'connected' }: { label: string; status?: 'connected' | 'pending' | 'offline' }) {
  return <span className="status-indicator"><span className={`status-dot status-dot--${status}`} aria-hidden="true" />{label}</span>;
}
