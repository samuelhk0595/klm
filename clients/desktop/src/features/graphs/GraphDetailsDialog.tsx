import { useEffect, useId, useRef, useState } from 'react';
import { X } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import type { Graph } from './demo';

export function GraphDetailsDialog({ graph, graphs, onSave, onClose }: {
  graph?: Graph;
  graphs: Graph[];
  onSave: (name: string, description: string) => void;
  onClose: () => void;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const id = useId();
  const [name, setName] = useState(graph?.name ?? '');
  const [description, setDescription] = useState(graph?.description ?? '');
  const duplicate = graphs.some(item => item.id !== graph?.id && item.name.toLowerCase() === name.trim().toLowerCase());
  useEffect(() => { dialog.current?.showModal(); }, []);

  return <dialog ref={dialog} className="settings-dialog graph-details-dialog" aria-labelledby={`${id}-title`} onCancel={onClose} onClose={onClose}>
    <form className="graph-details-form" onSubmit={event => {
      event.preventDefault();
      if (name.trim() && !duplicate) onSave(name.trim(), description.trim());
    }}>
      <div className="dialog-heading"><h2 id={`${id}-title`}>{graph ? 'Edit graph' : 'New graph'}</h2><IconButton label="Close graph details" onClick={onClose}><X /></IconButton></div>
      <div className="graph-details-field"><label htmlFor={`${id}-name`}>Name</label><Input id={`${id}-name`} autoFocus required maxLength={80} value={name} aria-invalid={duplicate} aria-describedby={duplicate ? `${id}-error` : undefined} onChange={event => setName(event.target.value)} />{duplicate && <p id={`${id}-error`} className="form-error" role="alert">A graph with this name already exists.</p>}</div>
      <div className="graph-details-field"><label htmlFor={`${id}-description`}>Description</label><Input id={`${id}-description`} maxLength={240} value={description} onChange={event => setDescription(event.target.value)} /></div>
      <div className="graph-details-actions"><Button onClick={onClose}>Cancel</Button><Button type="submit" variant="primary" disabled={!name.trim() || duplicate}>{graph ? 'Save changes' : 'Create graph'}</Button></div>
    </form>
  </dialog>;
}
