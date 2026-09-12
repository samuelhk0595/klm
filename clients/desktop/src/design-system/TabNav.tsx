import type { ReactNode } from 'react';

export function TabNav<T extends string>({ items, value, onChange }: {
  items: readonly { value: T; label: ReactNode }[];
  value: T;
  onChange: (value: T) => void;
}) {
  return <nav className="tab-nav" aria-label="Session views">{items.map(item => (
    <button key={item.value} type="button" className={`tab ${value === item.value ? 'is-active' : ''}`} aria-current={value === item.value ? 'page' : undefined} onClick={() => onChange(item.value)}>{item.label}</button>
  ))}</nav>;
}
