import { useEffect, useId, useRef } from 'react';
import { Button } from '../../design-system/Button';

export function DeleteTaskDialog({ name, reason, error, busy, onClose, onConfirm }: { name: string; reason: string; error?: string; busy?: boolean; onClose: () => void; onConfirm: () => void }) {
  const id = useId();
  const dialog = useRef<HTMLDialogElement>(null);
  useEffect(() => { dialog.current?.showModal(); }, []);
  return <dialog ref={dialog} className="settings-dialog task-delete-dialog" aria-labelledby={`${id}-title`} aria-describedby={`${id}-description`} onCancel={onClose} onClose={onClose}>
    <h2 id={`${id}-title`}>Delete task</h2><p id={`${id}-description`}>{reason || <>Delete “{name}”? Previous runs and sessions will be kept.</>}</p>
    {error && <p className="form-error" role="alert">{error}</p>}
    <div className="dialog-actions"><Button autoFocus disabled={busy} onClick={onClose}>Cancel</Button><Button variant="danger-soft" disabled={!!reason || busy} onClick={() => { if (!reason) onConfirm(); }}>{busy ? 'Deleting...' : 'Delete task'}</Button></div>
  </dialog>;
}
