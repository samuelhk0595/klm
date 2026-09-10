import { Folder, FolderPlus, PanelLeftClose, Plus, Search, Settings, Shapes, X } from 'lucide-react';
import { useState } from 'react';
import { Badge } from '../../design-system/Badge';
import { Button, IconButton } from '../../design-system/Button';
import { ProjectActionsMenu } from '../projects/ProjectActionsMenu';
import type { Session, Project } from '../../engine';

export function WorkspaceSidebar({ project, sessions, activeId, onSelect, onNew, onCreateFolder, onClose, onDesignSystem, onSettings, onEditProject, onRemoveProject }: {
  project: Project; sessions: Session[]; activeId: string; onSelect: (id: string) => void; onNew: (folder?: string) => void;
  onCreateFolder: (name: string) => Promise<string | null>;
  onClose: () => void; onDesignSystem: () => void; onSettings: () => void;
  onEditProject: (project: Project) => void;
  onRemoveProject: (project: Project) => Promise<string | null>;
}) {
  const [projectMenuOpen, setProjectMenuOpen] = useState(false);
  const [search, setSearch] = useState<string | null>(null);
  const [collapsed, setCollapsed] = useState<string[]>(['writing']);
  const [addingFolder, setAddingFolder] = useState(false);
  const [folderName, setFolderName] = useState('');
  const [folderError, setFolderError] = useState('');
  const [savingFolder, setSavingFolder] = useState(false);
  return <aside className="workspace-sidebar" aria-label="Sessions">
    <div className="brand-row"><div className="project-sidebar-heading"><strong title={project.name}>{project.name}</strong><Badge>Harness</Badge><ProjectActionsMenu project={project} onEdit={onEditProject} onRemove={onRemoveProject} open={projectMenuOpen} onOpenChange={setProjectMenuOpen} side="bottom" trigger={props => <IconButton {...props} label="Project settings" className="project-settings-button"><Settings /></IconButton>} /></div><IconButton label="Close sessions" onClick={onClose}><PanelLeftClose /></IconButton></div>
    <p className="project-folder-label" title={project.folder}><bdi dir="ltr">{project.folder}</bdi></p>
    <div className="new-session"><Button onClick={() => onNew()}><Plus />New Session</Button></div>
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
      {[...project.folders, 'Ungrouped'].map(workspace => {
        const matches = sessions.filter(session => session.workspace === workspace && session.title.toLowerCase().includes((search ?? '').toLowerCase()));
        const expanded = search !== null || !collapsed.includes(workspace);
        return <div className="workspace-group" key={workspace}>
          <div className="workspace-folder-row"><button className="workspace-folder" aria-expanded={expanded} onClick={() => setCollapsed(current => current.includes(workspace) ? current.filter(item => item !== workspace) : [...current, workspace])}><Folder />{workspace}</button><IconButton label={`New session in ${workspace}`} onClick={() => { setCollapsed(current => current.filter(item => item !== workspace)); onNew(workspace); }}><Plus /></IconButton></div>
          {expanded && <div className="session-list">{matches.map(session => <button key={session.id} className={`session-item ${session.id === activeId ? 'is-active' : ''}`} aria-current={session.id === activeId ? 'page' : undefined} onClick={() => onSelect(session.id)}><span>{session.title}</span><small>{session.status === 'running' ? 'Running' : session.harness === 'opencode' ? 'OpenCode' : session.harness === 'pi' ? 'Pi' : 'Codex'}</small></button>)}{matches.length === 0 && <p className="empty-workspace">No sessions</p>}</div>}
        </div>;
      })}
    </div>
    <div className="sidebar-footer"><Button variant="ghost" onClick={onDesignSystem}><Shapes />Design system</Button><Button variant="ghost" onClick={onSettings}><Settings />Settings</Button></div>
  </aside>;
}
