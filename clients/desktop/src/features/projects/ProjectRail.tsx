import { Moon, Plus, Settings, Sun } from 'lucide-react';
import { useEffect, useState } from 'react';
import { IconButton } from '../../design-system/Button';
import { ProjectActionsMenu } from './ProjectActionsMenu';
import type { Project } from '../../engine';

export function ProjectRail({ projects, activeId, onSelect, onEdit, onRemove, onAdd, onSettings, adding = false }: {
  projects: Project[]; activeId: string; onSelect: (project: Project) => void; onEdit: (project: Project) => void; onAdd: () => void; adding?: boolean;
  onRemove: (project: Project) => Promise<string | null>;
  onSettings: () => void;
}) {
  const [contextMenu, setContextMenu] = useState<{ projectId: string; x: number; y: number } | null>(null);
  const [theme, setTheme] = useState<'light' | 'dark'>(() => document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light');
  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', theme === 'dark' ? '#111318' : '#f8f9fa');
    try { localStorage.setItem('klm.theme', theme); }
    catch { /* Theme switching still works when browser storage is unavailable. */ }
  }, [theme]);
  return <nav className="project-rail" aria-label="Projects">
    <div className="rail-brand" title="KLM"><span className="brand-mark" aria-hidden="true" /><span className="sr-only">KLM</span></div>
    <div className="project-icons">{projects.map(project => <ProjectActionsMenu key={project.id} project={project} onEdit={onEdit} onRemove={onRemove}
      open={contextMenu?.projectId === project.id} position={contextMenu?.projectId === project.id ? contextMenu : undefined}
      onOpenChange={open => { if (!open) setContextMenu(current => current?.projectId === project.id ? null : current); }}
      trigger={props => <button {...props} className={`project-icon ${project.id === activeId ? 'is-active' : ''}`}
      aria-label={`Switch to ${project.name}`} aria-current={project.id === activeId ? 'page' : undefined}
      title={project.name} onClick={() => { setContextMenu(null); onSelect(project); }}
      onContextMenu={event => {
        event.preventDefault(); event.currentTarget.focus();
        const rect = event.currentTarget.getBoundingClientRect();
        setContextMenu({ projectId: project.id, x: event.clientX || rect.right, y: event.clientY || rect.bottom });
      }}
      onKeyDown={event => {
        if (event.key !== 'ContextMenu' && !(event.shiftKey && event.key === 'F10')) return;
        event.preventDefault();
        const rect = event.currentTarget.getBoundingClientRect();
        setContextMenu({ projectId: project.id, x: rect.right, y: rect.bottom });
      }}
    >{project.icon ? <img src={project.icon} alt="" /> : <span>{project.name.slice(0, 2).toUpperCase()}</span>}</button>} />)}</div>
    <IconButton label={adding ? 'Choosing project directory' : 'Add project'} className="add-project-button" disabled={adding} onClick={onAdd}><Plus /></IconButton>
    <div className="project-rail-footer"><IconButton label={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'} onClick={() => setTheme(current => current === 'dark' ? 'light' : 'dark')}>{theme === 'dark' ? <Sun /> : <Moon />}</IconButton><IconButton label="Settings" onClick={onSettings}><Settings /></IconButton></div>
  </nav>;
}
