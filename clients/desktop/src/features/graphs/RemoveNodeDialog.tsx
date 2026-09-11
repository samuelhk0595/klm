import { useEffect, useId, useRef } from 'react';
import { Button } from '../../design-system/Button';

export function RemoveNodeDialog({ nodeName, kind = 'node', onConfirm, onClose }: {
  nodeName: string;
  kind?: 'node' | 'choice';
  onConfirm: () => void;
  onClose: () => void;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const id = useId();
  useEffect(() => { dialog.current?.showModal(); }, []);

  return <dialog ref={dialog} className="settings-dialog graph-remove-node-dialog" aria-labelledby={`${id}-title`} aria-describedby={`${id}-description`} onCancel={onClose} onClose={onClose}>
    <h2 id={`${id}-title`}>{kind === 'choice' ? 'Remove choice' : 'Remove node'}</h2>
    <p id={`${id}-description`}>Remove “{nodeName}” from this graph?</p>
    <div className="dialog-actions"><Button autoFocus onClick={onClose}>Cancel</Button><Button variant="danger-soft" onClick={onConfirm}>Remove</Button></div>
  </dialog>;
}
