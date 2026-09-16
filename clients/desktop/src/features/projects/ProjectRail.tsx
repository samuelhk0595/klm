import { Moon, Plus, Settings, Sun } from 'lucide-react';
import { useEffect, useState } from 'react';
import { IconButton } from '../../design-system/Button';
import { ProjectActionsMenu } from './ProjectActionsMenu';
import type { Project, Session } from '../../engine';
import { IS_MOBILE_HOST, returnToHosts } from '../../platform';

function projectRuntimeState(project: Project, sessions: Session[]) {
  let state = 'off';
  const projectSessions = sessions.filter(session => session.projectId === project.id);
  const byId = new Map(projectSessions.map(session => [session.id, session]));
  for (const session of sessions) {
    const parent = session.parentId ? byId.get(session.parentId) : undefined;
    if (session.projectId !== project.id || session.role === 'graph_node' || session.role === 'subagent' || session.archived || parent?.archived) continue;
    if (project.archivedFolders.includes(parent?.workspace ?? session.workspace)) continue;
    if (session.status === 'error') return { state: 'error', label: 'Session error' };
    if ((session.permissions?.length ?? 0) > 0 || (session.questions?.length ?? 0) > 0) state = 'waiting';
    else if (state === 'off' && (session.status === 'running' || session.runtimeActive)) state = 'active';
  }
  if (state === 'waiting') return { state, label: 'Waiting for input' };
  if (state === 'active') return { state, label: 'Runtime active' };
  return { state, label: '' };
}

export function ProjectRail({ projects, sessions, activeId, onSelect, onEdit, onRemove, onReorder, onAdd, onSettings, adding = false }: {
  projects: Project[]; sessions: Session[]; activeId: string; onSelect: (project: Project) => void; onEdit: (project: Project) => void; onAdd: () => void; adding?: boolean;
  onRemove: (project: Project) => Promise<string | null>;
  onReorder: (project: Project, beforeId: string) => Promise<boolean>;
  onSettings: () => void;
}) {
  const [contextMenu, setContextMenu] = useState<{ projectId: string; x: number; y: number } | null>(null);
  const [dragging, setDragging] = useState('');
  const [dropTarget, setDropTarget] = useState<{ id: string; edge: 'before' | 'after' } | null>(null);
  const [theme, setTheme] = useState<'light' | 'dark'>(() => document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light');
  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', theme === 'dark' ? '#111318' : '#f8f9fa');
    try { localStorage.setItem('klm.theme', theme); }
    catch { /* Theme switching still works when browser storage is unavailable. */ }
  }, [theme]);
  return <nav className="project-rail" aria-label="Projects">
    {IS_MOBILE_HOST
      ? <IconButton label="Hosts" className="rail-brand" data-klm-host-navigation onClick={returnToHosts}><span className="brand-mark" aria-hidden="true" /></IconButton>
      : <div className="rail-brand" title="KLM"><span className="brand-mark" aria-hidden="true" /><span className="sr-only">KLM</span></div>}
    <div className="project-icons">{projects.map(project => {
      const runtime = projectRuntimeState(project, sessions);
      return <ProjectActionsMenu key={project.id} project={project} onEdit={onEdit} onRemove={onRemove}
      open={contextMenu?.projectId === project.id} position={contextMenu?.projectId === project.id ? contextMenu : undefined}
      onOpenChange={open => { if (!open) setContextMenu(current => current?.projectId === project.id ? null : current); }}
      trigger={props => <button {...props} draggable={!adding} className={`project-icon status-${runtime.state} ${project.id === activeId ? 'is-active' : ''} ${dragging === project.id ? 'is-dragging' : ''} ${dropTarget?.id === project.id ? `is-drop-${dropTarget.edge}` : ''}`}
      aria-label={`Switch to ${project.name}${runtime.label ? `, ${runtime.label}` : ''}`} aria-current={project.id === activeId ? 'page' : undefined}
      title={runtime.label ? `${project.name}: ${runtime.label}` : project.name} onClick={() => { setContextMenu(null); onSelect(project); }}
      onDragStart={event => { setContextMenu(null); setDragging(project.id); event.dataTransfer.effectAllowed = 'move'; event.dataTransfer.setData('application/x-klm-project', project.id); }}
      onDragEnd={() => { setDragging(''); setDropTarget(null); }}
      onDragOver={event => {
        if (!dragging || dragging === project.id) return;
        event.preventDefault(); event.dataTransfer.dropEffect = 'move';
        const bounds = event.currentTarget.getBoundingClientRect();
        setDropTarget({ id: project.id, edge: event.clientY < bounds.top + bounds.height / 2 ? 'before' : 'after' });
      }}
      onDrop={event => {
        event.preventDefault();
        if (!dragging || !dropTarget || dropTarget.id !== project.id) return;
        const moved = projects.find(item => item.id === dragging);
        const remaining = projects.filter(item => item.id !== dragging);
        const targetIndex = remaining.findIndex(item => item.id === project.id);
        if (!moved || targetIndex < 0) return;
        const beforeId = dropTarget.edge === 'before' ? project.id : remaining[targetIndex + 1]?.id ?? '';
        setDropTarget(null); void onReorder(moved, beforeId);
      }}
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
    >{project.icon ? <img src={project.icon} alt="" /> : <span>{project.name.slice(0, 2).toUpperCase()}</span>}</button>} />;
    })}</div>
    <IconButton label={adding ? 'Choosing project directory' : 'Add project'} className="add-project-button" disabled={adding} onClick={onAdd}><Plus /></IconButton>
    <div className="project-rail-footer"><IconButton label={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'} onClick={() => setTheme(current => current === 'dark' ? 'light' : 'dark')}>{theme === 'dark' ? <Sun /> : <Moon />}</IconButton><IconButton label="Settings" onClick={onSettings}><Settings /></IconButton></div>
  </nav>;
}
