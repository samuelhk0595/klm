import { Archive, ArchiveRestore, Bot, Folder, FolderPlus, MoreHorizontal, PanelLeftClose, Pencil, Plus, Search, Settings, SquarePen, Workflow, X } from 'lucide-react';
import { useState } from 'react';
import { Button, IconButton } from '../../design-system/Button';
import { Menu, MenuItem } from '../../design-system/Menu';
import { ProjectActionsMenu } from '../projects/ProjectActionsMenu';
import type { Session, Project } from '../../engine';

function sessionRuntimeState(session: Session) {
  if (session.status === 'error') return { state: 'error', label: 'Runtime error' };
  if ((session.permissions?.length ?? 0) > 0) return { state: 'waiting', label: 'Waiting for approval' };
  if ((session.questions?.length ?? 0) > 0) return { state: 'waiting', label: 'Waiting for input' };
  if (session.status === 'running') return { state: 'running', label: 'Turn running' };
  if (session.runtimeActive) return { state: 'active', label: 'Runtime active' };
  return { state: 'off', label: 'Runtime stopped' };
}

type DragItem = { kind: 'session'; id: string } | { kind: 'folder'; name: string };

export function WorkspaceSidebar({ project, sessions, activeId, activeSection, collapsed, archivedCollapsed, onCollapsedChange, onArchivedCollapsedChange, onSelect, onNew, onCreateFolder, onRename, onPosition, onReorderFolder, onArchiveSession, onArchiveFolder, onClose, onAgents, onGraphs, onEditProject, onRemoveProject }: {
  project: Project; sessions: Session[]; activeId: string; collapsed: string[]; archivedCollapsed: boolean; onCollapsedChange: (folders: string[]) => void; onArchivedCollapsedChange: (collapsed: boolean) => void;
  onSelect: (id: string) => void; onNew: (folder?: string) => void; onCreateFolder: (name: string) => Promise<string | null>;
  onRename: (session: Session) => void; onPosition: (session: Session, folder: string, beforeId: string) => Promise<boolean>;
  onReorderFolder: (folder: string, before: string) => Promise<boolean>;
  onArchiveSession: (session: Session, archived: boolean) => Promise<boolean>; onArchiveFolder: (folder: string, archived: boolean) => Promise<boolean>;
  onClose: () => void; activeSection: 'chat' | 'design' | 'agents' | 'graphs'; onAgents: () => void; onGraphs: () => void;
  onEditProject: (project: Project) => void; onRemoveProject: (project: Project) => Promise<string | null>;
}) {
  const [projectMenuOpen, setProjectMenuOpen] = useState(false);
  const [sessionMenu, setSessionMenu] = useState('');
  const [folderMenu, setFolderMenu] = useState('');
  const [search, setSearch] = useState<string | null>(null);
  const [addingFolder, setAddingFolder] = useState(false);
  const [folderName, setFolderName] = useState('');
  const [folderError, setFolderError] = useState('');
  const [savingFolder, setSavingFolder] = useState(false);
  const [busyAction, setBusyAction] = useState('');
  const [actionError, setActionError] = useState('');
  const [dragging, setDragging] = useState<DragItem | null>(null);
  const [folderDropTarget, setFolderDropTarget] = useState('');
  const [sessionDropTarget, setSessionDropTarget] = useState<{ id: string; edge: 'before' | 'after' } | null>(null);
  const [folderOrderTarget, setFolderOrderTarget] = useState<{ name: string; edge: 'before' | 'after' } | null>(null);
  const archivedFolders = new Set(project.archivedFolders);
  const activeFolders = project.folders.filter(folder => !archivedFolders.has(folder));
  const archivedSessions = sessions.filter(session => session.archived && !archivedFolders.has(session.workspace));
  const query = (search ?? '').toLowerCase();
  const matchingArchivedFolders = project.archivedFolders.filter(folder => folder.toLowerCase().includes(query));
  const matchingArchivedSessions = archivedSessions.filter(session => session.title.toLowerCase().includes(query));

  async function runAction(key: string, action: () => Promise<boolean>, failure: string) {
    if (busyAction) return false;
    setBusyAction(key); setActionError('');
    try {
      const succeeded = await action();
      if (succeeded) { setSessionMenu(''); setFolderMenu(''); }
      else setActionError(failure);
      return succeeded;
    } catch { setActionError(failure); return false; }
    finally { setBusyAction(''); }
  }

  async function positionSession(session: Session, workspace: string, beforeId: string) {
    setFolderDropTarget(''); setSessionDropTarget(null);
    if (await runAction(`move:${session.id}`, () => onPosition(session, workspace, beforeId), 'Could not move the session. Try again.')) {
      onCollapsedChange(collapsed.filter(folder => folder !== workspace));
    }
  }

  function sessionRow(session: Session, archived = false) {
    const runtime = sessionRuntimeState(session);
    return <div key={session.id} className={`session-row ${dragging?.kind === 'session' && dragging.id === session.id ? 'is-dragging' : ''} ${sessionDropTarget?.id === session.id ? `is-drop-${sessionDropTarget.edge}` : ''} ${sessionMenu === session.id ? 'has-open-menu' : ''}`} draggable={!archived && !busyAction && search === null}
      onDragStart={event => { setDragging({ kind: 'session', id: session.id }); event.dataTransfer.effectAllowed = 'move'; event.dataTransfer.setData('application/x-klm-session', session.id); }}
      onDragEnd={() => { setDragging(null); setFolderDropTarget(''); setSessionDropTarget(null); setFolderOrderTarget(null); }}
      onDragOver={event => {
        if (dragging?.kind !== 'session' || dragging.id === session.id) return;
        event.preventDefault(); event.stopPropagation(); event.dataTransfer.dropEffect = 'move';
        const bounds = event.currentTarget.getBoundingClientRect();
        setFolderDropTarget('');
        setSessionDropTarget({ id: session.id, edge: event.clientY < bounds.top + bounds.height / 2 ? 'before' : 'after' });
      }}
      onDrop={event => {
        if (dragging?.kind !== 'session' || !sessionDropTarget || sessionDropTarget.id !== session.id) return;
        event.preventDefault(); event.stopPropagation();
        const moved = sessions.find(item => item.id === dragging.id);
        const peers = sessions.filter(item => !item.archived && item.workspace === session.workspace && item.id !== dragging.id);
        const index = peers.findIndex(item => item.id === session.id);
        if (!moved || index < 0) return;
        const beforeId = sessionDropTarget.edge === 'before' ? session.id : peers[index + 1]?.id ?? '';
        void positionSession(moved, session.workspace, beforeId);
      }}>
      <button className={`session-item ${session.id === activeId ? 'is-active' : ''}`} disabled={archived} aria-current={session.id === activeId ? 'page' : undefined} onClick={() => onSelect(session.id)}>
        <span>{session.title}</span>{!archived && <i className={`session-runtime-led is-${runtime.state}`} role="img" aria-label={runtime.label} title={runtime.label} />}
      </button>
      {archived
        ? <IconButton label={`Restore ${session.title}`} disabled={!!busyAction} onClick={() => void runAction(`restore:${session.id}`, () => onArchiveSession(session, false), 'Could not restore the session. Try again.')}><ArchiveRestore /></IconButton>
        : <Menu label={`${session.title} actions`} role="menu" className="workspace-action-menu" open={sessionMenu === session.id} onOpenChange={open => { setActionError(''); setSessionMenu(open ? session.id : ''); }} side="bottom" trigger={props => <IconButton {...props} label={`${session.title} actions`}><MoreHorizontal /></IconButton>}>
          <MenuItem role="menuitem" disabled={!!busyAction} onClick={() => { setSessionMenu(''); onRename(session); }}><Pencil />Rename</MenuItem>
          <MenuItem role="menuitem" disabled={!!busyAction} onClick={() => void runAction(`archive:${session.id}`, () => onArchiveSession(session, true), 'Could not archive the session. Try again.')}><Archive />Archive</MenuItem>
        </Menu>}
    </div>;
  }

  return <aside className="workspace-sidebar" aria-label="Sessions">
    <div className="brand-row"><div className="project-sidebar-heading"><strong title={project.name}>{project.name}</strong><ProjectActionsMenu project={project} onEdit={onEditProject} onRemove={onRemoveProject} open={projectMenuOpen} onOpenChange={setProjectMenuOpen} side="bottom" trigger={props => <IconButton {...props} label="Project settings" className="project-settings-button"><Settings /></IconButton>} /></div><IconButton label="Close sessions" onClick={onClose}><PanelLeftClose /></IconButton></div>
    <p className="project-folder-label" title={project.folder}><bdi dir="ltr">{project.folder}</bdi></p>
    <nav className="sidebar-navigation" aria-label="Project navigation">
      <Button variant="ghost" size="sm" onClick={() => onNew()}><SquarePen />New session</Button>
      <Button variant="ghost" size="sm" aria-current={activeSection === 'agents' ? 'page' : undefined} onClick={onAgents}><Bot />Agents</Button>
      <Button variant="ghost" size="sm" aria-current={activeSection === 'graphs' ? 'page' : undefined} onClick={onGraphs}><Workflow />Graphs</Button>
    </nav>
    <div className="workspace-list">
      <div className="workspace-heading"><h2>Sessions</h2><div className="workspace-heading-actions"><IconButton label="Search sessions" aria-expanded={search !== null} onClick={() => setSearch(search === null ? '' : null)}><Search /></IconButton><IconButton label="Create session folder" aria-expanded={addingFolder} onClick={() => setAddingFolder(!addingFolder)}><FolderPlus /></IconButton></div></div>
      {addingFolder && <form className="new-folder-form" onSubmit={async event => {
        event.preventDefault();
        if (!folderName.trim() || savingFolder) return;
        setSavingFolder(true);
        try {
          const failure = await onCreateFolder(folderName.trim());
          if (failure) { setFolderError(failure); return; }
          setFolderName(''); setFolderError(''); setAddingFolder(false);
        } finally { setSavingFolder(false); }
      }}><input autoFocus required maxLength={60} disabled={savingFolder} aria-label="Session folder name" placeholder="Topic name" value={folderName} onChange={event => setFolderName(event.target.value)} /><div><Button type="submit" size="sm" disabled={!folderName.trim() || savingFolder}>Create folder</Button><IconButton label="Cancel creating folder" disabled={savingFolder} onClick={() => setAddingFolder(false)}><X /></IconButton></div>{folderError && <p role="alert" className="form-error">{folderError}</p>}</form>}
      {search !== null && <input autoFocus className="search-input" aria-label="Search sessions" placeholder="Find a session..." value={search} onChange={event => setSearch(event.target.value)} />}
      {[...activeFolders, 'Ungrouped'].map(workspace => {
        const matches = sessions.filter(session => !session.archived && session.workspace === workspace && session.title.toLowerCase().includes(query));
        const expanded = search !== null || !collapsed.includes(workspace);
        const isDropTarget = folderDropTarget === workspace;
        return <div className={`workspace-group ${isDropTarget ? 'is-drop-target' : ''}`} key={workspace}
          onDragOver={event => { if (dragging?.kind !== 'session') return; event.preventDefault(); event.dataTransfer.dropEffect = 'move'; setSessionDropTarget(null); setFolderDropTarget(workspace); }}
          onDragLeave={event => { if (!(event.relatedTarget instanceof Node) || !event.currentTarget.contains(event.relatedTarget)) setFolderDropTarget(''); }}
          onDrop={event => {
            if (dragging?.kind !== 'session') return;
            event.preventDefault();
            const session = sessions.find(item => item.id === dragging.id);
            if (session) void positionSession(session, workspace, '');
          }}>
          <div className={`workspace-folder-row ${dragging?.kind === 'folder' && dragging.name === workspace ? 'is-dragging' : ''} ${folderOrderTarget?.name === workspace ? `is-drop-${folderOrderTarget.edge}` : ''}`}
            onDragStart={event => {
              if (workspace === 'Ungrouped') return;
              setDragging({ kind: 'folder', name: workspace }); event.dataTransfer.effectAllowed = 'move'; event.dataTransfer.setData('application/x-klm-folder', workspace);
            }}
            onDragEnd={() => { setDragging(null); setFolderDropTarget(''); setSessionDropTarget(null); setFolderOrderTarget(null); }}
            onDragOver={event => {
              if (!dragging) return;
              if (dragging.kind === 'session') {
                event.preventDefault(); event.stopPropagation(); event.dataTransfer.dropEffect = 'move'; setSessionDropTarget(null); setFolderDropTarget(workspace); return;
              }
              if (dragging.name === workspace) return;
              event.preventDefault(); event.stopPropagation(); event.dataTransfer.dropEffect = 'move';
              const bounds = event.currentTarget.getBoundingClientRect();
              setFolderOrderTarget({ name: workspace, edge: workspace === 'Ungrouped' || event.clientY < bounds.top + bounds.height / 2 ? 'before' : 'after' });
            }}
            onDrop={event => {
              if (!dragging) return;
              event.preventDefault(); event.stopPropagation();
              if (dragging.kind === 'session') {
                const session = sessions.find(item => item.id === dragging.id);
                if (session) void positionSession(session, workspace, '');
                return;
              }
              if (!folderOrderTarget || folderOrderTarget.name !== workspace) return;
              const remaining = activeFolders.filter(folder => folder !== dragging.name);
              const targetIndex = remaining.indexOf(workspace);
              const before = workspace === 'Ungrouped' ? '' : folderOrderTarget.edge === 'before' ? workspace : remaining[targetIndex + 1] ?? '';
              const folder = dragging.name;
              setFolderOrderTarget(null);
              void runAction(`reorder-folder:${folder}`, () => onReorderFolder(folder, before), 'Could not reorder the folder. Try again.');
            }}>
            <button className="workspace-folder" draggable={workspace !== 'Ungrouped' && !busyAction && search === null} aria-expanded={expanded} onClick={() => { if (search === null) onCollapsedChange(collapsed.includes(workspace) ? collapsed.filter(item => item !== workspace) : [...collapsed, workspace]); }}><Folder />{workspace}</button><IconButton label={`New session in ${workspace}`} onClick={() => { onCollapsedChange(collapsed.filter(item => item !== workspace)); onNew(workspace); }}><Plus /></IconButton>{workspace !== 'Ungrouped' && <Menu label={`${workspace} actions`} role="menu" className="workspace-action-menu" open={folderMenu === workspace} onOpenChange={open => { setActionError(''); setFolderMenu(open ? workspace : ''); }} side="bottom" trigger={props => <IconButton {...props} label={`${workspace} actions`}><MoreHorizontal /></IconButton>}><MenuItem role="menuitem" disabled={!!busyAction} onClick={() => void runAction(`archive-folder:${workspace}`, () => onArchiveFolder(workspace, true), 'Could not archive the folder. Try again.')}><Archive />Archive</MenuItem></Menu>}
          </div>
          {expanded && <div className="session-list">{matches.map(session => sessionRow(session))}{matches.length === 0 && <p className="empty-workspace">No sessions</p>}</div>}
        </div>;
      })}
      <div className="workspace-group">
        <div className="workspace-folder-row"><button className="workspace-folder" aria-expanded={search !== null || !archivedCollapsed} onClick={() => { if (search === null) onArchivedCollapsedChange(!archivedCollapsed); }}><Archive />Archived</button></div>
        {(search !== null || !archivedCollapsed) && <div className="session-list">
          {matchingArchivedFolders.map(folder => <div className="archived-row" key={folder}><span><Folder />{folder}</span><IconButton label={`Restore ${folder}`} disabled={!!busyAction} onClick={() => void runAction(`restore-folder:${folder}`, () => onArchiveFolder(folder, false), 'Could not restore the folder. Try again.')}><ArchiveRestore /></IconButton></div>)}
          {matchingArchivedSessions.map(session => sessionRow(session, true))}
          {matchingArchivedFolders.length === 0 && matchingArchivedSessions.length === 0 && <p className="empty-workspace">No sessions</p>}
        </div>}
      </div>
      {actionError && <p role="alert" className="form-error sidebar-action-error">{actionError}</p>}
    </div>
  </aside>;
}
