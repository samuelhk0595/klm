import { useEffect, useId, useRef } from 'react';
import { Button } from '../../design-system/Button';

export function DeleteGraphDialog({ graphName, onConfirm, onClose }: {
  graphName: string;
  onConfirm: () => void;
  onClose: () => void;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const id = useId();
  useEffect(() => { dialog.current?.showModal(); }, []);

  return <dialog ref={dialog} className="settings-dialog graph-delete-dialog" aria-labelledby={`${id}-title`} aria-describedby={`${id}-description`} onCancel={onClose} onClose={onClose}>
    <h2 id={`${id}-title`}>Delete graph</h2>
    <p id={`${id}-description`}>Are you sure you want to delete “{graphName}”?</p>
    <div className="dialog-actions"><Button autoFocus onClick={onClose}>Cancel</Button><Button variant="danger-soft" onClick={onConfirm}>Delete graph</Button></div>
  </dialog>;
}
