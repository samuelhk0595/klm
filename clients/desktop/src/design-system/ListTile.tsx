import type { ReactNode } from 'react';

export function ListTile({ title, description, metadata, onClick, actions, status, className = '' }: {
  title: string;
  description?: string;
  metadata?: ReactNode;
  onClick: () => void;
  actions?: ReactNode;
  status?: ReactNode;
  className?: string;
}) {
  return <div className={`list-tile ${className}`}>
    <button type="button" className="list-tile-content" onClick={onClick}>
      <span className="list-tile-heading"><strong>{title}</strong>{status}</span>
      {description && <span className="list-tile-description">{description}</span>}
      {metadata && <span className="list-tile-metadata">{metadata}</span>}
    </button>
    {actions && <div className="list-tile-actions">{actions}</div>}
  </div>;
}
