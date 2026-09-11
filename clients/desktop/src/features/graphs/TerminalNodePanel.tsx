import { useId, useState } from 'react';
import { X } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { Textarea } from '../../design-system/Textarea';
import type { TerminalDefinition } from './terminal';

export function TerminalNodePanel({ terminal, onSave, onClose }: {
  terminal: TerminalDefinition;
  onSave: (terminal: TerminalDefinition) => void;
  onClose: () => void;
}) {
  const id = useId();
  const [draft, setDraft] = useState(terminal);
  const valid = !!draft.name.trim() && !!draft.command.trim();

  return <aside className="graph-agent-panel nodrag nopan nowheel" role="dialog" aria-labelledby={`${id}-title`} onKeyDown={event => {
    if (event.key === 'Escape' && !event.defaultPrevented) { event.stopPropagation(); onClose(); }
  }}>
    <form className="graph-choice-form" onSubmit={event => {
      event.preventDefault();
      if (valid) onSave({ name: draft.name.trim(), command: draft.command });
    }}>
      <div className="graph-agent-panel-header"><h2 id={`${id}-title`}>Terminal command</h2><IconButton label="Close terminal settings" onClick={onClose}><X /></IconButton></div>
      <div className="graph-agent-panel-field"><label htmlFor={`${id}-name`}>Name</label><Input id={`${id}-name`} autoFocus required maxLength={80} placeholder="Update repository" value={draft.name} onChange={event => setDraft(current => ({ ...current, name: event.target.value }))} /></div>
      <div className="graph-agent-panel-field"><label htmlFor={`${id}-command`}>Command</label><Textarea id={`${id}-command`} className="graph-terminal-command-input" required rows={7} placeholder="git pull" spellCheck={false} autoCapitalize="none" autoCorrect="off" value={draft.command} onChange={event => setDraft(current => ({ ...current, command: event.target.value }))} /></div>
      <div className="graph-choice-form-actions"><Button onClick={onClose}>Cancel</Button><Button type="submit" variant="primary" disabled={!valid}>Apply</Button></div>
    </form>
  </aside>;
}
