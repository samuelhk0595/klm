import { useEffect, useRef, useState } from 'react';
import { ArrowLeft, Download, PanelLeft, PanelRight, X } from 'lucide-react';
import { Button, IconButton } from './design-system/Button';
import { TabNav } from './design-system/TabNav';
import { WorkspaceSidebar } from './features/workspace/WorkspaceSidebar';
import { AgentsPage } from './features/agents/AgentsPage';
import { initialAgents, mockAgentGraphReferences, type Agent } from './features/agents/demo';
import { GraphsPage } from './features/graphs/GraphsPage';
import { initialGraphs, type Graph } from './features/graphs/demo';
import { SideChatPanel } from './features/chat/SideChatPanel';
import { CreateSessionDialog } from './features/workspace/CreateSessionDialog';
import { ConversationEvents } from './features/chat/ConversationEvents';
import { SessionStatusBar } from './features/chat/SessionStatusBar';
import { ModelPicker } from './features/chat/ModelPicker';
import { GraphPicker } from './features/chat/GraphPicker';
import { PermissionCard } from './features/chat/PermissionCard';
import { QuestionCard } from './features/chat/QuestionCard';
import { MessageComposer, type ComposerDraft } from './features/chat/MessageComposer';
import { DesignSystem } from './DesignSystem';
import { ProjectRail } from './features/projects/ProjectRail';
import { ProjectDialog } from './features/projects/ProjectDialog';
import { ENGINE_URL, pickDirectory, request, type EngineState, type Harness, type Project, type Session, type SourceReference } from './engine';
import { Select } from './design-system/Select';

function mergeSessions(current: Session[], incoming: Session[]): Session[] {
  // Preserve nanosecond precision and normalize Go's variable fractional digits.
  const version = (timestamp: string) => {
    const [seconds, fraction = ''] = timestamp.replace(/Z$/, '').split('.');
    return `${seconds}.${fraction.padEnd(9, '0')}`;
  };
  const sessions = new Map(current.map(session => [session.id, session]));
  for (const session of incoming) {
    const previous = sessions.get(session.id);
    if (!previous || version(session.updatedAt) > version(previous.updatedAt)) sessions.set(session.id, session);
  }
  return [...sessions.values()];
}

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : 'The engine request failed. Please retry.';
}

type WorkspaceNavigation = { activeProjectId: string; selectedSessions: Record<string, string>; openSideSessions: string[] };
const navigationStorageKey = `klm.workspace-navigation.v1:${ENGINE_URL}`;

function readWorkspaceNavigation(): WorkspaceNavigation {
  const fallback: WorkspaceNavigation = { activeProjectId: '', selectedSessions: {}, openSideSessions: [] };
  try {
    const saved: unknown = JSON.parse(localStorage.getItem(navigationStorageKey) ?? 'null');
    if (!saved || typeof saved !== 'object' || Array.isArray(saved)) return fallback;
    const value = saved as Partial<WorkspaceNavigation>;
    return {
      activeProjectId: typeof value.activeProjectId === 'string' ? value.activeProjectId : '',
      selectedSessions: value.selectedSessions && typeof value.selectedSessions === 'object' && !Array.isArray(value.selectedSessions)
        ? Object.fromEntries(Object.entries(value.selectedSessions).filter((entry): entry is [string, string] => typeof entry[1] === 'string')) : {},
      openSideSessions: Array.isArray(value.openSideSessions) ? [...new Set(value.openSideSessions.filter((id): id is string => typeof id === 'string'))] : [],
    };
  } catch { return fallback; }
}

export function App() {
  const [engine, setEngine] = useState<EngineState>({ projects: [], sessions: [], harnesses: [] });
  const { projects, sessions, harnesses } = engine;
  const [loaded, setLoaded] = useState(false);
  const [connectionError, setConnectionError] = useState('');
  const [actionError, setActionError] = useState('');
  const [sessionErrors, setSessionErrors] = useState<Record<string, string>>({});
  const [pendingSessions, setPendingSessions] = useState<Record<string, boolean>>({});
  const [pendingMessages, setPendingMessages] = useState<Record<string, boolean>>({});
  const pending = useRef(new Set<string>());
  const [refreshVersion, setRefreshVersion] = useState(0);
  const projectRevision = useRef(0);
  const navigation = useRef(0);
  const streams = useRef(new Map<string, EventSource>());
  const [workspaceNavigation, setWorkspaceNavigation] = useState(readWorkspaceNavigation);
  const { activeProjectId, selectedSessions, openSideSessions } = workspaceNavigation;
  const [addingProject, setAddingProject] = useState<string | null>(null);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [pickingDirectory, setPickingDirectory] = useState(false);
  const pickerPending = useRef(false);
  const [creatingSession, setCreatingSession] = useState<{ projectId: string; workspace: string } | null>(null);
  const project = projects.find(item => item.id === activeProjectId) ?? projects[0];
  const projectSessions = sessions.filter(item => item.projectId === project?.id && !item.parentId);
  const session = projectSessions.find(item => item.id === selectedSessions[project?.id ?? '']) ?? projectSessions[0];
  const activeId = session?.id ?? '';
  const working = session?.status === 'running' || !!pendingMessages[activeId];
  const awaitingPermission = !!session?.permissions?.length;
  const awaitingQuestion = !!session?.questions?.length;
  const createProject = projects.find(item => item.id === creatingSession?.projectId);
  const [view, setView] = useState<'chat' | 'design' | 'agents' | 'graphs'>('chat');
  const [projectAgents, setProjectAgents] = useState<Record<string, Agent[]>>({});
  const [projectGraphs, setProjectGraphs] = useState<Record<string, Graph[]>>({});
  const [selectedGraphs, setSelectedGraphs] = useState<Record<string, string>>({});
  const [agentsVisit, setAgentsVisit] = useState(0);
  const [graphsVisit, setGraphsVisit] = useState(0);
  const [creatingGraph, setCreatingGraph] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const sideOpen = openSideSessions.includes(activeId);
  const sideVisible = sideOpen && view === 'chat' && !sidebarOpen && !creatingSession && !editingProject && !pickingDirectory && addingProject === null;
  const sideSession = sessions.find(item => item.parentId === activeId);
  const [sideSources, setSideSources] = useState<Record<string, SourceReference[]>>({});
  const [sideErrors, setSideErrors] = useState<Record<string, string>>({});
  const openingSide = useRef(new Set<string>());
  const [drafts, setDrafts] = useState<Record<string, ComposerDraft>>({});
  const settings = useRef<HTMLDialogElement>(null);
  const history = useRef<HTMLDivElement>(null);
  const shell = useRef<HTMLDivElement>(null);
  const drag = useRef<{ pointerId: number; x: number; y: number; left: number; top: number; minX: number; maxX: number; minY: number; maxY: number } | null>(null);
  useEffect(() => {
    // Wait for engine data before persisting fallback selections; an initial
    // empty render must not erase the project/session saved before a reload.
    if (!loaded) return;
    try {
      localStorage.setItem(navigationStorageKey, JSON.stringify({
        ...workspaceNavigation,
        activeProjectId: project?.id ?? '',
        selectedSessions: project && activeId ? { ...selectedSessions, [project.id]: activeId } : selectedSessions,
      }));
    } catch { /* Navigation still works when browser storage is unavailable. */ }
  }, [loaded, workspaceNavigation, project?.id, activeId, selectedSessions]);
  function setSideOpen(open: boolean) {
    if (!activeId) return;
    setWorkspaceNavigation(current => ({ ...current, openSideSessions: open
      ? [...new Set([...current.openSideSessions, activeId])]
      : current.openSideSessions.filter(id => id !== activeId) }));
  }
  useEffect(() => {
    let active = true;
    let refreshing = false;
    async function refresh() {
      if (refreshing) return;
      refreshing = true;
      const revision = projectRevision.current;
      try {
        const next = await request<EngineState>('/api/state');
        if (!active) return;
        const projectsUnchanged = revision === projectRevision.current;
        setEngine(current => {
          const visibleProjects = projectsUnchanged ? next.projects : current.projects;
          const visibleIds = new Set(visibleProjects.map(project => project.id));
          return {
            projects: visibleProjects,
            sessions: mergeSessions(current.sessions, next.sessions).filter(session => visibleIds.has(session.projectId)),
            harnesses: next.harnesses,
          };
        });
        setLoaded(true);
        setConnectionError('');
      } catch (error) {
        if (active) setConnectionError(errorMessage(error));
      } finally {
        refreshing = false;
      }
    }
    void refresh();
    const timer = window.setInterval(() => void refresh(), 5000);
    return () => { active = false; window.clearInterval(timer); };
  }, [refreshVersion]);

  const streamIds = JSON.stringify(sessions.filter(item => item.status === 'running' || item.id === activeId || (sideVisible && item.id === sideSession?.id)).map(item => item.id).sort());
  useEffect(() => {
    const ids = new Set<string>(JSON.parse(streamIds) as string[]);
    for (const [id, source] of streams.current) {
      if (!ids.has(id)) { source.close(); streams.current.delete(id); }
    }
    for (const id of ids) {
      if (streams.current.has(id)) continue;
      const source = new EventSource(`${ENGINE_URL}/api/sessions/${encodeURIComponent(id)}/events`);
      streams.current.set(id, source);
      source.onopen = () => {
        if (streams.current.get(id) === source) setRefreshVersion(current => current + 1);
      };
      source.onmessage = event => {
        if (streams.current.get(id) !== source) return;
        try {
          const snapshot = JSON.parse(event.data) as Session;
          if (snapshot.id !== id || !Array.isArray(snapshot.events) || typeof snapshot.updatedAt !== 'string') throw new Error('Invalid snapshot');
          setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [snapshot]) }));
        } catch {
          setConnectionError('The engine sent an invalid session update. Retry to refresh the session.');
        }
      };
      source.onerror = () => {
        if (streams.current.get(id) === source) setConnectionError(`Live updates disconnected from ${ENGINE_URL}. Reconnecting automatically; check that the engine is running.`);
      };
    }
  }, [streamIds]);
  useEffect(() => {
    const sources = streams.current;
    return () => { sources.forEach(source => source.close()); sources.clear(); };
  }, []);

  useEffect(() => {
    const resetPosition = () => {
      drag.current = null;
      if (shell.current) {
        shell.current.style.left = '0px';
        shell.current.style.top = '0px';
        delete shell.current.dataset.dragging;
      }
    };
    window.addEventListener('resize', resetPosition);
    return () => window.removeEventListener('resize', resetPosition);
  }, []);
  useEffect(() => {
    const mobileSidebar = sidebarOpen && window.matchMedia('(max-width: 760px)').matches;
    if (!mobileSidebar) return;
    const panel = document.querySelector<HTMLElement>('.workspace-sidebar');
    if (!panel) return;
    const previous = document.activeElement as HTMLElement | null;
    const covered = Array.from(document.querySelectorAll<HTMLElement>('.main-panel, .side-chat-panel, .project-rail'));
    covered.forEach(element => { element.inert = true; });
    const focusable = () => Array.from(panel.querySelectorAll<HTMLElement>('button, input, select, summary, [tabindex="0"]')).filter(element => element.getClientRects().length > 0);
    focusable()[0]?.focus();
    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') { setSidebarOpen(false); }
      if (event.key === 'Tab') {
        const items = focusable(); const first = items[0]; const last = items[items.length - 1];
        if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last?.focus(); }
        else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus(); }
      }
    }
    const resize = () => { setSidebarOpen(false); };
    document.addEventListener('keydown', onKey); window.addEventListener('resize', resize);
    return () => { covered.forEach(element => { element.inert = false; }); document.removeEventListener('keydown', onKey); window.removeEventListener('resize', resize); previous?.focus(); };
  }, [sidebarOpen, project?.id]);
  useEffect(() => { if (sideVisible && session) void ensureSide(session); }, [sideVisible, activeId]);
  useEffect(() => { history.current?.scrollTo?.({ top: history.current.scrollHeight }); }, [session?.updatedAt, activeId, view, working]);
  function newSession(folder = 'Ungrouped') {
    if (!project) return;
    navigation.current += 1;
    setCreatingSession({ projectId: project.id, workspace: folder });
    setSidebarOpen(false);
  }
  async function updateSession(target: Session, action: 'messages' | 'stop' | 'move' | 'harness', body: (ComposerDraft & { sources?: SourceReference[] }) | { workspace: string } | { harness: Harness['id'] } | Record<string, never>) {
    const id = target.id;
    if (pending.current.has(id)) return false;
    pending.current.add(id);
    setPendingSessions(current => ({ ...current, [id]: true }));
    if (action === 'messages') setPendingMessages(current => ({ ...current, [id]: true }));
    setSessionErrors(current => ({ ...current, [id]: '' }));
    try {
      const next = await request<Session>(`/api/sessions/${encodeURIComponent(id)}${action === 'move' ? '' : `/${action}`}`, action === 'move' || action === 'harness' ? 'PATCH' : 'POST', body);
      setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [next]) }));
      return true;
    } catch (error) {
      setSessionErrors(current => ({ ...current, [id]: errorMessage(error) }));
      return false;
    } finally {
      pending.current.delete(id);
      setPendingSessions(current => ({ ...current, [id]: false }));
      if (action === 'messages') setPendingMessages(current => ({ ...current, [id]: false }));
    }
  }
  async function applyModelSettings(target: Session, model: string, effort: string) {
    const id = target.id;
    if (pending.current.has(id) || target.status === 'running') return false;
    pending.current.add(id); setPendingSessions(current => ({ ...current, [id]: true }));
    setSessionErrors(current => ({ ...current, [id]: '' }));
    try {
      const next = await request<Session>(`/api/sessions/${encodeURIComponent(id)}/settings`, 'PATCH', { model, effort });
      setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [next]) }));
      return true;
    } catch (error) { setSessionErrors(current => ({ ...current, [id]: errorMessage(error) })); return false; }
    finally { pending.current.delete(id); setPendingSessions(current => ({ ...current, [id]: false })); }
  }
  async function send(draft: ComposerDraft) {
    if (!session || session.status === 'running' || !draft.text.trim()) return false;
    return updateSession(session, 'messages', draft);
  }
  function receiveSession(snapshot: Session) {
    setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [snapshot]) }));
  }
  async function ensureSide(main: Session) {
    if (sessions.some(item => item.parentId === main.id) || openingSide.current.has(main.id)) return;
    openingSide.current.add(main.id);
    setSideErrors(current => ({ ...current, [main.id]: '' }));
    try { receiveSession(await request<Session>(`/api/sessions/${encodeURIComponent(main.id)}/side`, 'POST', {})); }
    catch (error) { setSideErrors(current => ({ ...current, [main.id]: errorMessage(error) })); }
    finally { openingSide.current.delete(main.id); }
  }
  function askSide(messageId: string, passage: string) {
    if (!session) return;
    setSideSources(current => ({ ...current, [session.id]: [...(current[session.id] ?? []), { sessionId: session.id, messageId, passage }] }));
    setSideOpen(true); setSidebarOpen(false);
  }
  async function sendSide(draft: ComposerDraft) {
    if (!sideSession || sideSession.status === 'running') return false;
    const mainId = activeId;
    const sources = sideSources[mainId] ?? [];
    const accepted = await updateSession(sideSession, 'messages', { ...draft, sources });
    if (accepted) setSideSources(current => ({ ...current, [mainId]: (current[mainId] ?? []).filter(source => !sources.includes(source)) }));
    return accepted;
  }
  function exportSession() {
    if (!session) return;
    const url = URL.createObjectURL(new Blob([JSON.stringify(session, null, 2)], { type: 'application/json' }));
    const link = document.createElement('a'); link.href = url; link.download = `session-${session.id}.json`; link.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  }
  function selectProject(next: Project) {
    navigation.current += 1;
    setWorkspaceNavigation(current => ({ ...current, activeProjectId: next.id })); setView('chat'); setSidebarOpen(false);
  }
  function openProjectEditor(target: Project) {
    navigation.current += 1;
    setEditingProject(target); setSidebarOpen(false);
  }
  async function openProjectPicker() {
    if (pickerPending.current) return;
    pickerPending.current = true;
    setPickingDirectory(true); setActionError(''); setSidebarOpen(false);
    const selection = ++navigation.current;
    try {
      const folder = await pickDirectory();
      if (folder !== null && navigation.current === selection) setAddingProject(folder);
    } catch (error) {
      setActionError(errorMessage(error));
    } finally {
      pickerPending.current = false;
      setPickingDirectory(false);
    }
  }
  async function addProject(input: Omit<Project, 'id' | 'folders'>): Promise<string | null> {
    const selection = navigation.current;
    try {
      const next = await request<Project>('/api/projects', 'POST', input);
      projectRevision.current += 1;
      setEngine(current => ({ ...current, projects: [...current.projects.filter(item => item.id !== next.id), next] }));
      if (navigation.current === selection) { setWorkspaceNavigation(current => ({ ...current, activeProjectId: next.id })); setView('chat'); }
      return null;
    } catch (error) { return errorMessage(error); }
  }
  async function editProject(projectId: string, input: Pick<Project, 'name' | 'icon'>): Promise<string | null> {
    try {
      const next = await request<Project>(`/api/projects/${encodeURIComponent(projectId)}`, 'PATCH', { name: input.name, icon: input.icon });
      projectRevision.current += 1;
      setEngine(current => ({ ...current, projects: current.projects.map(item => item.id === projectId ? next : item) }));
      return null;
    } catch (error) { return errorMessage(error); }
  }
  async function removeProject(projectId: string): Promise<string | null> {
    try {
      await request(`/api/projects/${encodeURIComponent(projectId)}`, 'DELETE');
      projectRevision.current += 1;
      setEngine(current => ({
        ...current,
        projects: current.projects.filter(item => item.id !== projectId),
        sessions: current.sessions.filter(item => item.projectId !== projectId),
      }));
      setWorkspaceNavigation(current => ({ ...current, activeProjectId: current.activeProjectId === projectId ? '' : current.activeProjectId }));
      return null;
    } catch (error) { return errorMessage(error); }
  }
  async function createSession(input: { projectId: string; title: string; workspace: string; harness: Harness['id']; model?: string }): Promise<string | null> {
    const selection = navigation.current;
    try {
      const next = await request<Session>('/api/sessions', 'POST', input);
      setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [next]) }));
      if (navigation.current === selection) {
        setWorkspaceNavigation(current => ({ ...current, activeProjectId: next.projectId, selectedSessions: { ...current.selectedSessions, [next.projectId]: next.id } }));
        setView('chat');
      }
      return null;
    } catch (error) { return errorMessage(error); }
  }
  async function createFolder(name: string): Promise<string | null> {
    if (!project) return 'Select a project first.';
    const projectId = project.id;
    try {
      const next = await request<Project>(`/api/projects/${encodeURIComponent(projectId)}/folders`, 'POST', { name });
      projectRevision.current += 1;
      setEngine(current => ({ ...current, projects: current.projects.map(item => item.id === projectId ? next : item) }));
      return null;
    } catch (error) { return errorMessage(error); }
  }
  function updateGraphs(graphs: Graph[]) {
    if (!project) return;
    const projectId = project.id;
    const changedNames = new Map<string, string | undefined>();
    for (const previous of projectGraphs[projectId] ?? initialGraphs) {
      const next = graphs.find(graph => graph.id === previous.id);
      if (previous.name !== next?.name) changedNames.set(previous.name, next?.name);
    }
    setProjectGraphs(current => ({ ...current, [projectId]: graphs }));
    const projectSessionIds = new Set(projectSessions.map(session => session.id));
    const availableGraphIds = new Set(graphs.filter(graph => graph.enabled).map(graph => graph.id));
    setSelectedGraphs(current => Object.fromEntries(Object.entries(current).filter(([sessionId, graphId]) => !projectSessionIds.has(sessionId) || availableGraphIds.has(graphId))));
    // Keep the existing simulated agent references consistent with renamed/deleted graphs.
    if (changedNames.size) setProjectAgents(current => ({ ...current, [projectId]: (current[projectId] ?? initialAgents).map(agent => ({
      ...agent,
      graphReferences: (agent.graphReferences ?? mockAgentGraphReferences[agent.id] ?? []).flatMap(name => {
        if (!changedNames.has(name)) return [name];
        const nextName = changedNames.get(name);
        return nextName ? [nextName] : [];
      }),
    })) }));
  }

  return <div ref={shell} className={`app-shell ${sidebarOpen ? 'sidebar-open' : ''} ${sideVisible && session ? 'side-agent-open' : ''}`}
    onPointerDown={event => {
      const target = event.target as HTMLElement;
      if (event.button !== 0 || !event.isPrimary || window.innerWidth <= 760 || !target.closest('.session-header') || target.closest('button, a, input, select, textarea')) return;
      const element = event.currentTarget;
      const bounds = element.getBoundingClientRect();
      drag.current = {
        pointerId: event.pointerId, x: event.clientX, y: event.clientY,
        left: parseFloat(element.style.left) || 0, top: parseFloat(element.style.top) || 0,
        minX: 8 - bounds.left, maxX: window.innerWidth - bounds.right - 8,
        minY: 8 - bounds.top, maxY: window.innerHeight - bounds.bottom - 8,
      };
      element.setPointerCapture(event.pointerId);
      element.dataset.dragging = 'true';
      event.preventDefault();
    }}
    onPointerMove={event => {
      const start = drag.current;
      if (!start || start.pointerId !== event.pointerId) return;
      const dx = Math.max(start.minX, Math.min(start.maxX, event.clientX - start.x));
      const dy = Math.max(start.minY, Math.min(start.maxY, event.clientY - start.y));
      event.currentTarget.style.left = `${start.left + dx}px`;
      event.currentTarget.style.top = `${start.top + dy}px`;
    }}
    onPointerUp={event => {
      if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId);
    }}
    onLostPointerCapture={() => { drag.current = null; if (shell.current) delete shell.current.dataset.dragging; }}
    onPointerCancel={() => { drag.current = null; if (shell.current) delete shell.current.dataset.dragging; }}
  >
    <ProjectRail projects={projects} activeId={project?.id ?? ''} onSelect={selectProject} onEdit={openProjectEditor} onRemove={target => removeProject(target.id)} onAdd={() => void openProjectPicker()} onSettings={() => { setSidebarOpen(false); settings.current?.showModal(); }} adding={pickingDirectory} />
    {sidebarOpen && <button className="panel-backdrop" aria-label="Close side panel" onClick={() => setSidebarOpen(false)} />}
    {project && <WorkspaceSidebar key={`sidebar:${project.id}`} project={project} sessions={projectSessions} activeId={view === 'chat' ? activeId : ''} activeSection={view} onNew={newSession} onCreateFolder={createFolder} onSelect={id => { navigation.current += 1; setWorkspaceNavigation(current => ({ ...current, selectedSessions: { ...current.selectedSessions, [project.id]: id } })); setView('chat'); setSidebarOpen(false); }} onClose={() => setSidebarOpen(false)} onAgents={() => { navigation.current += 1; setAgentsVisit(current => current + 1); setView('agents'); setSidebarOpen(false); }} onGraphs={() => { navigation.current += 1; setGraphsVisit(current => current + 1); setCreatingGraph(false); setView('graphs'); setSidebarOpen(false); }} onEditProject={openProjectEditor} onRemoveProject={target => removeProject(target.id)} />}
    <main className="main-panel">
      <header className="session-header"><div className="header-leading">{project && <IconButton label="Open sessions" className="mobile-nav" onClick={() => setSidebarOpen(true)}><PanelLeft /></IconButton>}{view === 'graphs' && creatingGraph && project ? <Button onClick={() => setCreatingGraph(false)}><ArrowLeft />Back to graphs</Button> : <h1>{view === 'design' ? 'Design system' : view === 'agents' ? 'Agents' : view === 'graphs' ? 'Graphs' : session?.title ?? project?.name ?? 'KLM'}</h1>}{view === 'chat' && session && <span className="mode-label">{session.status}</span>}</div><div className="header-actions">{session && view === 'chat' && <><Button size="sm" className="export-button" onClick={exportSession}>Session log<Download /></Button><IconButton label={sideOpen ? 'Close side agent' : 'Open side agent'} aria-expanded={sideOpen} onClick={() => { setSideOpen(!sideOpen); setSidebarOpen(false); }}><PanelRight /></IconButton></>}</div></header>
      {connectionError && <div role="alert" className="storage-error">{connectionError} <Button size="sm" onClick={() => setRefreshVersion(current => current + 1)}>Retry</Button></div>}
      {actionError && <div role="alert" className="storage-error">{actionError} <Button size="sm" disabled={pickingDirectory} onClick={() => void openProjectPicker()}>Retry</Button></div>}
      {(view === 'chat' || view === 'design') && <div className="session-navigation"><TabNav value={view} onChange={setView} items={[{ value: 'chat', label: 'Chat' }]} />{view === 'chat' && project && session && <Select label="Move session to folder" value={session.workspace} disabled={pendingSessions[session.id] || !!connectionError} onChange={event => void updateSession(session, 'move', { workspace: event.target.value })}>{[...project.folders, 'Ungrouped'].map(folder => <option key={folder} value={folder}>{folder}</option>)}</Select>}</div>}
      {view === 'design' ? <DesignSystem /> : view === 'agents' && project ? <AgentsPage key={`${project.id}:${agentsVisit}`} agents={projectAgents[project.id] ?? initialAgents} onChange={agents => setProjectAgents(current => ({ ...current, [project.id]: agents }))} /> : view === 'graphs' && project ? <GraphsPage key={`${project.id}:${graphsVisit}`} graphs={projectGraphs[project.id] ?? initialGraphs} agents={projectAgents[project.id] ?? initialAgents} onChange={updateGraphs} creating={creatingGraph} onCreate={() => setCreatingGraph(true)} /> : !loaded ? <div className="empty-chat"><h2>{connectionError ? 'Engine disconnected' : 'Connecting to engine...'}</h2></div> : !project ? <div className="empty-chat"><h2>Add a project</h2><Button onClick={() => void openProjectPicker()} disabled={pickingDirectory}>{pickingDirectory ? 'Choosing directory...' : 'Add project'}</Button></div> : !session ? <div className="empty-chat"><h2>Create a session</h2><Button onClick={() => newSession()} disabled={!!connectionError}>Create session</Button></div> : <>
        <div className="chat-history" ref={history} role="log" aria-label="Conversation" aria-live="polite">
          {session.events.length ? <ConversationEvents key={session.id} session={session} onSnapshot={receiveSession} onAskSide={askSide} /> : !working ? <div className="empty-chat"><h2>What would you like to work on?</h2></div> : null}
          {working && <div className="chat-working" role="status">{awaitingPermission ? 'Waiting for approval' : awaitingQuestion ? 'Waiting for your answer' : <><span className="working-spinner" aria-hidden="true" />Working</>}</div>}
        </div>
        {sessionErrors[session.id] && <div role="alert" className="storage-error">{sessionErrors[session.id]}</div>}
        <div className="composer-area">
          {(awaitingPermission || awaitingQuestion) && <div className="permission-queue">
            {session.permissions?.map(permission => <PermissionCard key={permission.id} permission={permission} sessionId={session.id} projectName={project.name} onResolved={snapshot => setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [snapshot]) }))} />)}
            {session.questions?.map(question => <QuestionCard key={question.id} question={question} sessionId={session.id} onResolved={snapshot => setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [snapshot]) }))} />)}
          </div>}
          <MessageComposer key={session.id} projectId={session.projectId} draft={drafts[session.id] ?? { text: '', mentions: [] }} onDraftChange={draft => setDrafts(current => ({ ...current, [session.id]: draft }))} onSend={send} disabled={!!connectionError || pendingSessions[session.id] || session.status === 'running' || awaitingPermission || awaitingQuestion} running={session.status === 'running'} onStop={() => void updateSession(session, 'stop', {})} graphControl={<GraphPicker key={session.id} graphs={projectGraphs[session.projectId] ?? initialGraphs} value={selectedGraphs[session.id] ?? ''} onChange={graphId => setSelectedGraphs(current => ({ ...current, [session.id]: graphId }))} />} modelControl={<ModelPicker key={session.id} session={session} disabled={!!connectionError || !!pendingSessions[session.id] || working} onSave={(model, effort) => applyModelSettings(session, model, effort)} />} />
          <SessionStatusBar session={session} />
        </div>
      </>}
    </main>
    {sideVisible && project && session && <SideChatPanel key={session.id} session={sideSession} project={project} harnesses={harnesses}
      draft={drafts[`side:${session.id}`] ?? { text: '', mentions: [] }} sources={sideSources[session.id] ?? []}
      error={sideSession ? sessionErrors[sideSession.id] ?? '' : sideErrors[session.id] ?? ''} disconnected={!!connectionError} pending={!!sideSession && !!pendingSessions[sideSession.id]}
      onClose={() => setSideOpen(false)} onRetry={() => void ensureSide(session)} onSnapshot={receiveSession}
      onDraftChange={draft => setDrafts(current => ({ ...current, [`side:${session.id}`]: draft }))}
      onRemoveSource={index => setSideSources(current => ({ ...current, [session.id]: (current[session.id] ?? []).filter((_, i) => i !== index) }))}
      onSend={sendSide} onStop={() => { if (sideSession) void updateSession(sideSession, 'stop', {}); }}
      onHarness={harness => { if (sideSession) void updateSession(sideSession, 'harness', { harness }); }}
      onModel={(model, effort) => sideSession ? applyModelSettings(sideSession, model, effort) : Promise.resolve(false)} />}
    {addingProject !== null && <ProjectDialog initialFolder={addingProject} onSave={addProject} onClose={() => { navigation.current += 1; setAddingProject(null); }} />}
    {editingProject && <ProjectDialog key={`edit-project:${editingProject.id}`} initialFolder={editingProject.folder} project={editingProject} onSave={input => editProject(editingProject.id, input)} onClose={() => setEditingProject(null)} />}
    {creatingSession && createProject && <CreateSessionDialog project={createProject} harnesses={harnesses} workspace={creatingSession.workspace} onCreate={createSession} onClose={() => { navigation.current += 1; setCreatingSession(null); }} />}
    <dialog ref={settings} aria-label="Settings" className="settings-dialog"><div className="dialog-heading"><h2>Settings</h2><IconButton label="Close settings" onClick={() => settings.current?.close()}><X /></IconButton></div><p>Projects and session history are saved locally by the engine. Unsent drafts are kept only until this page reloads.</p><p className="small muted">Engine: {ENGINE_URL}</p><Button onClick={() => { settings.current?.close(); setView('design'); }}>Design system</Button></dialog>
  </div>;
}
