import type { ReactNode } from 'react';

export function TabNav<T extends string>({ items, value, onChange }: {
  items: readonly { value: T; label: ReactNode; onClose?: () => void; closeLabel?: string }[];
  value: T;
  onChange: (value: T) => void;
}) {
  return <nav className="tab-nav" aria-label="Session views">{items.map(item => item.onClose ? (
    <span key={item.value} className={`tab-group ${value === item.value ? 'is-active' : ''}`}>
      <button type="button" className="tab" aria-current={value === item.value ? 'page' : undefined} onClick={() => onChange(item.value)}>{item.label}</button>
      <button type="button" className="tab-close" aria-label={item.closeLabel ?? 'Close tab'} onClick={item.onClose}>×</button>
    </span>
  ) : <button key={item.value} type="button" className={`tab ${value === item.value ? 'is-active' : ''}`} aria-current={value === item.value ? 'page' : undefined} onClick={() => onChange(item.value)}>{item.label}</button>)}</nav>;
}
