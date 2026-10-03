import { useEffect, useId, useRef } from 'react';
import { Button } from '../../design-system/Button';

export function DeleteTaskDialog({ name, reason, onClose, onConfirm }: { name: string; reason: string; onClose: () => void; onConfirm: () => void }) {
  const id = useId();
  const dialog = useRef<HTMLDialogElement>(null);
  useEffect(() => { dialog.current?.showModal(); }, []);
  return <dialog ref={dialog} className="settings-dialog task-delete-dialog" aria-labelledby={`${id}-title`} aria-describedby={`${id}-description`} onCancel={onClose} onClose={onClose}>
    <h2 id={`${id}-title`}>Delete task</h2><p id={`${id}-description`}>{reason || <>Delete “{name}”? Previous runs and sessions will be kept.</>}</p>
    <div className="dialog-actions"><Button autoFocus onClick={onClose}>Cancel</Button><Button variant="danger-soft" disabled={!!reason} onClick={() => { if (!reason) onConfirm(); }}>Delete task</Button></div>
  </dialog>;
}
