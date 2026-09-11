import { useEffect, useId, useRef } from 'react';
import { Button } from '../../design-system/Button';

export function AgentInUseDialog({ agentName, action, graphs, onClose }: {
  agentName: string;
  action: 'delete' | 'disable';
  graphs: string[];
  onClose: () => void;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const id = useId();
  useEffect(() => { dialog.current?.showModal(); }, []);

  return <dialog ref={dialog} className="settings-dialog agent-in-use-dialog" aria-labelledby={`${id}-title`} aria-describedby={`${id}-description`} onCancel={onClose} onClose={onClose}>
    <h2 id={`${id}-title`}>{action === 'delete' ? 'Cannot delete agent' : 'Cannot disable agent'}</h2>
    <p id={`${id}-description`}>Remove all references to “{agentName}” from the following graphs before {action === 'delete' ? 'deleting' : 'disabling'} it.</p>
    <ul className="agent-graph-references">{graphs.map(graph => <li key={graph}>{graph}</li>)}</ul>
    <div className="dialog-actions"><Button autoFocus onClick={onClose}>Close</Button></div>
  </dialog>;
}
