import type { ReactNode } from 'react';
import { ChevronDown } from 'lucide-react';

export function Section({ title, children, collapsible = false }: { title: string; children: ReactNode; collapsible?: boolean }) {
  return collapsible ? (
    <details className="section" open>
      <summary className="section-title"><ChevronDown size={12} />{title}</summary>
      <div className="section-content">{children}</div>
    </details>
  ) : <section className="section"><h2 className="section-title">{title}</h2><div className="section-content">{children}</div></section>;
}
