import { Pencil, X } from 'lucide-react';
import { useRef, useState, type ComponentProps } from 'react';
import { Menu, MenuItem } from '../../design-system/Menu';
import type { Project } from '../../engine';

export function ProjectActionsMenu({ project, onEdit, onRemove, onOpenChange, ...menu }: Omit<ComponentProps<typeof Menu>, 'label' | 'role' | 'children' | 'className'> & {
  project: Project;
  onEdit: (project: Project) => void;
  onRemove: (project: Project) => Promise<string | null>;
}) {
  const [removing, setRemoving] = useState(false);
  const [error, setError] = useState('');
  const pending = useRef(false);

  function changeOpen(open: boolean) {
    setError('');
    onOpenChange(open);
  }

  return <Menu {...menu} label={`${project.name} actions`} role="menu" className="project-context-menu" onOpenChange={changeOpen}>
    <MenuItem role="menuitem" disabled={removing} onClick={() => { changeOpen(false); onEdit(project); }}><Pencil />Edit</MenuItem>
    <MenuItem role="menuitem" className="project-remove-action" disabled={removing} onClick={async () => {
      if (pending.current) return;
      pending.current = true; setRemoving(true); setError('');
      try {
        const failure = await onRemove(project);
        if (failure) setError(failure); else changeOpen(false);
      } catch { setError('Could not remove the project. Try again.'); }
      finally { pending.current = false; setRemoving(false); }
    }}><X />{removing ? 'Removing...' : 'Remove'}</MenuItem>
    {error && <p className="form-error menu-feedback" role="alert">{error}</p>}
  </Menu>;
}
