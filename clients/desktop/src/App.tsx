import { lazy, Suspense, useCallback, useEffect, useMemo, useRef, useState, type FormEvent } from 'react';
import { Download, Focus, GitBranch, PanelLeft, PanelRight, X } from 'lucide-react';
import { configureEngineURL, IS_DESKTOP, openFocus, startWindowDrag } from './platform';
import { Button, IconButton } from './design-system/Button';
import { Input } from './design-system/Input';
import { Badge } from './design-system/Badge';
import { TabNav } from './design-system/TabNav';
import { WorkspaceSidebar } from './features/workspace/WorkspaceSidebar';
import { ProjectAuthoring } from './features/graphs/ProjectAuthoring';
import { useAuthoringCatalog } from './features/graphs/catalog';
import { graphListEntry } from './features/graphs/files';
import { SideChatPanel } from './features/chat/SideChatPanel';
import { CreateSessionDialog } from './features/workspace/CreateSessionDialog';
import { SessionStatusBar } from './features/chat/SessionStatusBar';
import { ModelPicker } from './features/chat/ModelPicker';
import { isSessionResponse, mergeSessions, prependHistory } from './features/chat/sessionState';
import { ConversationHistory } from './features/chat/ConversationHistory';
import { SubagentView } from './features/chat/SubagentView';
import { GraphPicker } from './features/chat/GraphPicker';
import { PermissionCard } from './features/chat/PermissionCard';
import { QuestionCard } from './features/chat/QuestionCard';
import { MessageComposer, type ComposerDraft } from './features/chat/MessageComposer';
import { DesignSystem } from './DesignSystem';
import { ProjectRail } from './features/projects/ProjectRail';
import { ProjectDialog } from './features/projects/ProjectDialog';
import { ENGINE_URL, getConversationGraph, mergeGraphState, selectConversationGraph, request, type ConversationGraphState, type EngineState, type EngineSnapshot, type EventPage, type Harness, type Project, type Session, type SessionResponse, type SessionMetadata, type SourceReference } from './engine';
import { Select } from './design-system/Select';

const GraphView = lazy(() => import('./features/graphs/GraphView').then(module => ({ default: module.GraphView })));
type SessionTab = 'chat' | 'graph' | `subagent:${string}`;

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
  const [focusError, setFocusError] = useState('');
  async function focus() {
    try { await openFocus(); setFocusError(''); }
    catch (error) { setFocusError(error instanceof Error ? error.message : String(error)); }
  }
  const [sessionErrors, setSessionErrors] = useState<Record<string, string>>({});
  const [pendingSessions, setPendingSessions] = useState<Record<string, boolean>>({});
  const [pendingMessages, setPendingMessages] = useState<Record<string, boolean>>({});
  const pending = useRef(new Set<string>());
  const [refreshVersion, setRefreshVersion] = useState(0);
  const [sessionMetadata, setSessionMetadata] = useState<SessionMetadata | null>(null);
  const metadataRequest = useRef(0);
  const projectRevision = useRef(0);
  const navigation = useRef(0);
  const streams = useRef(new Map<string, EventSource>());
  const [workspaceNavigation, setWorkspaceNavigation] = useState(readWorkspaceNavigation);
  const { activeProjectId, selectedSessions, openSideSessions } = workspaceNavigation;
  const [addingProject, setAddingProject] = useState<string | null>(null);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [creatingSession, setCreatingSession] = useState<{ projectId: string; workspace: string } | null>(null);
  const project = projects.find(item => item.id === activeProjectId) ?? projects[0];
  const projectSessions = sessions.filter(item => item.projectId === project?.id && !item.parentId && item.role !== 'graph_node');
  const session = projectSessions.find(item => item.id === selectedSessions[project?.id ?? '']) ?? projectSessions[0];
  const activeId = session?.id ?? '';
  const gitBranch = sessionMetadata?.sessionId === activeId ? sessionMetadata.gitBranch ?? '' : '';
  const working = session?.status === 'running' || !!pendingMessages[activeId];
  const awaitingPermission = !!session?.permissions?.length;
  const awaitingQuestion = !!session?.questions?.length;
  const createProject = projects.find(item => item.id === creatingSession?.projectId);
  const [view, setView] = useState<'chat' | 'design' | 'agents' | 'graphs'>('chat');
  const { catalog, error: catalogError, loading: catalogLoading, reload: reloadCatalog } = useAuthoringCatalog(project?.id ?? '');
  const catalogGraphs = useMemo(() => catalog?.graphs.map(graphListEntry) ?? [], [catalog]);
  const [graphErrors, setGraphErrors] = useState<Record<string, string>>({});
  const [selectingGraphs, setSelectingGraphs] = useState<Record<string, boolean>>({});
  const graphSelectionPending = useRef(new Set<string>());
  const [sessionTabs, setSessionTabs] = useState<Record<string, SessionTab>>({});
  const [openSubagentTabs, setOpenSubagentTabs] = useState<Record<string, string[]>>({});
  const selectedGraphId = session?.graph?.selectedGraphId ?? session?.selectedGraphId ?? '';
  const graphRun = session?.graph?.run;
  const graphRunInProgress = !!selectedGraphId && !!graphRun?.active && graphRun.graphId === selectedGraphId;
  const catalogSelectedGraph = catalog?.graphs.find(graph => graph.id === selectedGraphId);
  const selectedGraph = graphRunInProgress ? graphRun!.snapshot : catalogSelectedGraph;
  const requestedTab = sessionTabs[activeId] ?? 'chat';
  const activeTab: SessionTab = requestedTab === 'graph' && !selectedGraphId ? 'chat' : requestedTab;
  const openSubagentIds = openSubagentTabs[activeId] ?? [];
  const activeSubagentId = activeTab.startsWith('subagent:') ? activeTab.slice('subagent:'.length) : '';
  const activeSubagent = sessions.find(item => item.id === activeSubagentId && item.parentId === activeId && item.role === 'subagent');
  const graphRequests = session?.graph?.requests ?? [];
  useEffect(() => {
    void reloadCatalog().catch(() => {});
  }, [reloadCatalog, selectedGraphId, graphRun?.id, view]);
  const [agentsVisit, setAgentsVisit] = useState(0);
  const [graphsVisit, setGraphsVisit] = useState(0);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const sideOpen = openSideSessions.includes(activeId);
  const sideVisible = sideOpen && view === 'chat' && !sidebarOpen && !creatingSession && !editingProject && addingProject === null;
  const sideSession = sessions.find(item => item.parentId === activeId && item.role === 'side_agent');
  const [sideSources, setSideSources] = useState<Record<string, SourceReference[]>>({});
  const [sideErrors, setSideErrors] = useState<Record<string, string>>({});
  const openingSide = useRef(new Set<string>());
  const [drafts, setDrafts] = useState<Record<string, ComposerDraft>>({});
  const settings = useRef<HTMLDialogElement>(null);
  const [settingsEngineURL, setSettingsEngineURL] = useState(ENGINE_URL);
  const [settingsError, setSettingsError] = useState('');
  function connectEngine(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      configureEngineURL(settingsEngineURL);
      window.location.reload();
    } catch (error) {
      setSettingsError(error instanceof Error ? error.message : 'Enter a valid engine URL.');
    }
  }
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
        const next = await request<EngineSnapshot>('/api/state');
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

  useEffect(() => {
    if (!activeId) {
      setSessionMetadata(null);
      return;
    }
    const controller = new AbortController();
    setSessionMetadata(current => current?.sessionId === activeId ? current : { sessionId: activeId });
    async function refreshMetadata() {
      const requestId = ++metadataRequest.current;
      try {
        const next = await request<SessionMetadata>(`/api/sessions/${encodeURIComponent(activeId)}/metadata`, 'GET', undefined, 15000, controller.signal);
        if (!controller.signal.aborted && requestId === metadataRequest.current && next.sessionId === activeId) setSessionMetadata(next);
      } catch {
        if (!controller.signal.aborted && requestId === metadataRequest.current) setSessionMetadata({ sessionId: activeId });
      }
    }
    void refreshMetadata();
    const onFocus = () => void refreshMetadata();
    window.addEventListener('focus', onFocus);
    return () => { controller.abort(); window.removeEventListener('focus', onFocus); };
  }, [activeId, session?.status]);

  const applyGraph = useCallback((snapshot: ConversationGraphState) => {
    setEngine(current => ({ ...current, sessions: current.sessions.map(item => {
      if (item.id !== snapshot.sessionId) return item;
      const graph = mergeGraphState(item.graph, snapshot);
      return graph === item.graph ? item : { ...item, graph, selectedGraphId: graph?.selectedGraphId };
    }) }));
  }, []);
  // Hydrate/reconcile selection after CRUD, and poll even while the chat is idle.
  useEffect(() => {
    if (!activeId) return;
    let active = true;
    let pending = false;
    async function refreshGraph() {
      if (pending) return;
      pending = true;
      try {
        const snapshot = await getConversationGraph(activeId);
        if (active) { applyGraph(snapshot); setGraphErrors(current => ({ ...current, [activeId]: '' })); }
      } catch (error) {
        if (active) setGraphErrors(current => ({ ...current, [activeId]: errorMessage(error) }));
      } finally { pending = false; }
    }
    void refreshGraph();
    const timer = window.setInterval(() => void refreshGraph(), 5000);
    return () => { active = false; window.clearInterval(timer); };
  }, [activeId, catalog, refreshVersion, applyGraph]);
  useEffect(() => {
    if (!selectedGraphId) setSessionTabs(current => current[activeId] === 'graph' ? { ...current, [activeId]: 'chat' } : current);
  }, [activeId, selectedGraphId]);
  async function chooseGraph(id: string, graphId: string) {
    if (graphSelectionPending.current.has(id)) return;
    graphSelectionPending.current.add(id);
    setSelectingGraphs(current => ({ ...current, [id]: true }));
    setGraphErrors(current => ({ ...current, [id]: '' }));
    try { applyGraph(await selectConversationGraph(id, graphId)); }
    catch (error) { setGraphErrors(current => ({ ...current, [id]: errorMessage(error) })); }
    finally { graphSelectionPending.current.delete(id); setSelectingGraphs(current => ({ ...current, [id]: false })); }
  }
  const streamSet = new Set(sessions.filter(item => item.role !== 'graph_node' && item.role !== 'subagent' && (item.status === 'running' || item.runtimeActive || item.graph?.run?.active || item.id === activeId || (sideVisible && item.id === sideSession?.id))).map(item => item.id));
  sessions.filter(item => item.role === 'subagent' && item.status === 'running').forEach(item => streamSet.add(item.id));
  session?.events.forEach(event => {
    const childSessionId = event.type === 'subagent' && ['running', 'pending', 'started'].includes(event.status ?? 'running') && typeof event.data?.childSessionId === 'string' ? event.data.childSessionId : '';
    if (childSessionId) streamSet.add(childSessionId);
  });
  openSubagentIds.forEach(id => streamSet.add(id));
  const streamIds = JSON.stringify([...streamSet].sort());
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
          const snapshot: unknown = JSON.parse(event.data);
          if (!isSessionResponse(snapshot) || ('summary' in snapshot ? snapshot.summary.id : snapshot.id) !== id) throw new Error('Invalid snapshot');
          setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [snapshot], true) }));
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
      const next = await request<SessionResponse>(`/api/sessions/${encodeURIComponent(id)}${action === 'move' ? '' : `/${action}`}`, action === 'move' || action === 'harness' ? 'PATCH' : 'POST', body);
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
      const next = await request<SessionResponse>(`/api/sessions/${encodeURIComponent(id)}/settings`, 'PATCH', { model, effort });
      setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [next]) }));
      return true;
    } catch (error) { setSessionErrors(current => ({ ...current, [id]: errorMessage(error) })); return false; }
    finally { pending.current.delete(id); setPendingSessions(current => ({ ...current, [id]: false })); }
  }
  async function send(draft: ComposerDraft) {
    if (!session || session.status === 'running' || !draft.text.trim()) return false;
    return updateSession(session, 'messages', draft);
  }
  function receiveSession(snapshot: SessionResponse) {
    setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [snapshot]) }));
  }
  function receiveHistory(page: EventPage) {
    setEngine(current => ({ ...current, sessions: prependHistory(current.sessions, page) }));
  }
  function graphRequestResolved(conversationId: string) {
    // Callback responses describe the private node session, not a public chat.
    void getConversationGraph(conversationId).then(applyGraph, error => setGraphErrors(current => ({ ...current, [conversationId]: errorMessage(error) })));
  }
  function openSubagent(sessionId: string) {
    if (!activeId) return;
    setOpenSubagentTabs(current => ({
      ...current,
      [activeId]: (current[activeId] ?? []).includes(sessionId) ? current[activeId] : [...(current[activeId] ?? []), sessionId],
    }));
    setSessionTabs(current => ({ ...current, [activeId]: `subagent:${sessionId}` }));
  }
  function closeSubagent(sessionId: string) {
    setOpenSubagentTabs(current => ({ ...current, [activeId]: (current[activeId] ?? []).filter(id => id !== sessionId) }));
    setSessionTabs(current => current[activeId] === `subagent:${sessionId}` ? { ...current, [activeId]: 'chat' } : current);
  }
  async function ensureSide(main: Session) {
    if (sessions.some(item => item.parentId === main.id && item.role === 'side_agent') || openingSide.current.has(main.id)) return;
    openingSide.current.add(main.id);
    setSideErrors(current => ({ ...current, [main.id]: '' }));
    try { receiveSession(await request<SessionResponse>(`/api/sessions/${encodeURIComponent(main.id)}/side`, 'POST', {})); }
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
  async function exportSession() {
    if (!session) return;
    const id = session.id;
    try {
      const snapshot = await request<Session>(`/api/sessions/${encodeURIComponent(id)}/export`);
      const url = URL.createObjectURL(new Blob([JSON.stringify(snapshot, null, 2)], { type: 'application/json' }));
      const link = document.createElement('a'); link.href = url; link.download = `session-${id}.json`; link.click();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    } catch (error) { setSessionErrors(current => ({ ...current, [id]: errorMessage(error) })); }
  }
  function selectProject(next: Project) {
    navigation.current += 1;
    setWorkspaceNavigation(current => ({ ...current, activeProjectId: next.id })); setView('chat'); setSidebarOpen(false);
  }
  function openProjectEditor(target: Project) {
    navigation.current += 1;
    setEditingProject(target); setSidebarOpen(false);
  }
  function openProjectPicker() {
    navigation.current += 1;
    setSidebarOpen(false);
    setAddingProject('');
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
      const next = await request<SessionResponse>('/api/sessions', 'POST', input);
      const created = 'summary' in next ? next.summary : next;
      setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [next]) }));
      if (navigation.current === selection) {
        setWorkspaceNavigation(current => ({ ...current, activeProjectId: created.projectId, selectedSessions: { ...current.selectedSessions, [created.projectId]: created.id } }));
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

  const tabItems: { value: SessionTab; label: string; onClose?: () => void; closeLabel?: string }[] = [{ value: 'chat', label: 'Chat' }];
  if (selectedGraphId) tabItems.push({ value: 'graph', label: 'Graph' });
  for (const id of openSubagentIds) {
    const child = sessions.find(item => item.id === id && item.parentId === activeId && item.role === 'subagent');
    const source = session?.events.find(event => event.type === 'subagent' && event.data?.childSessionId === id);
    const label = child?.title || source?.title || 'Subagent';
    tabItems.push({ value: `subagent:${id}`, label, closeLabel: `Close ${label}`, onClose: () => closeSubagent(id) });
  }

  return <div ref={shell} className={`app-shell ${sidebarOpen ? 'sidebar-open' : ''} ${sideVisible && session ? 'side-agent-open' : ''}`}
    onPointerDown={event => {
      const target = event.target as HTMLElement;
      if (event.button !== 0 || !event.isPrimary || !target.closest('.session-header') || target.closest('button, a, input, select, textarea')) return;
      if (IS_DESKTOP) { void startWindowDrag().catch(error => setFocusError(errorMessage(error))); return; }
      if (window.innerWidth <= 760) return;
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
    <ProjectRail projects={projects} sessions={sessions} activeId={project?.id ?? ''} onSelect={selectProject} onEdit={openProjectEditor} onRemove={target => removeProject(target.id)} onAdd={() => void openProjectPicker()} onSettings={() => { setSidebarOpen(false); setSettingsEngineURL(ENGINE_URL); setSettingsError(''); settings.current?.showModal(); }} />
    {sidebarOpen && <button className="panel-backdrop" aria-label="Close side panel" onClick={() => setSidebarOpen(false)} />}
    {project && <WorkspaceSidebar key={`sidebar:${project.id}`} project={project} sessions={projectSessions} activeId={view === 'chat' ? activeId : ''} activeSection={view} onNew={newSession} onCreateFolder={createFolder} onSelect={id => { navigation.current += 1; setWorkspaceNavigation(current => ({ ...current, selectedSessions: { ...current.selectedSessions, [project.id]: id } })); setView('chat'); setSidebarOpen(false); }} onClose={() => setSidebarOpen(false)} onAgents={() => { navigation.current += 1; setAgentsVisit(current => current + 1); setView('agents'); setSidebarOpen(false); }} onGraphs={() => { navigation.current += 1; setGraphsVisit(current => current + 1); setView('graphs'); setSidebarOpen(false); }} onEditProject={openProjectEditor} onRemoveProject={target => removeProject(target.id)} />}
    <main className="main-panel">
      <header className="session-header"><div id="workspace-header-leading" className="header-leading">{project && <IconButton label="Open sessions" className="mobile-nav" onClick={() => setSidebarOpen(true)}><PanelLeft /></IconButton>}<h1>{view === 'design' ? 'Design system' : view === 'agents' ? 'Agents' : view === 'graphs' ? 'Graphs' : session?.title ?? project?.name ?? 'KLM'}</h1>{view === 'chat' && session && gitBranch && <span className="mode-label session-header-branch" title={gitBranch}><GitBranch aria-hidden="true" /><span>{gitBranch}</span></span>}</div><div id="workspace-header-actions" className="header-actions">
        {session && view === 'chat' && <Button size="sm" className="export-button" onClick={exportSession}>Session log<Download /></Button>}
        {IS_DESKTOP && <IconButton label="Focus" onClick={() => void focus()}><Focus /></IconButton>}
        {session && view === 'chat' && <IconButton label={sideOpen ? 'Close side agent' : 'Open side agent'} aria-expanded={sideOpen} onClick={() => { setSideOpen(!sideOpen); setSidebarOpen(false); }}><PanelRight /></IconButton>}
      </div></header>
      {focusError && <div role="alert" className="storage-error">{focusError} <Button size="sm" onClick={() => void focus()}>Retry</Button></div>}
      {connectionError && <div role="alert" className="storage-error">{connectionError} <Button size="sm" onClick={() => setRefreshVersion(current => current + 1)}>Retry</Button></div>}
      {(view === 'chat' || view === 'design') && <div className="session-navigation"><TabNav<SessionTab> value={view === 'design' ? 'chat' : activeTab} onChange={tab => { setView('chat'); setSessionTabs(current => ({ ...current, [activeId]: tab })); }} items={view === 'chat' ? tabItems : [{ value: 'chat', label: 'Chat' }]} />{view === 'chat' && project && session && <Select label="Move session to folder" value={session.workspace} disabled={pendingSessions[session.id] || !!connectionError} onChange={event => void updateSession(session, 'move', { workspace: event.target.value })}>{[...project.folders, 'Ungrouped'].map(folder => <option key={folder} value={folder}>{folder}</option>)}</Select>}</div>}
      {view === 'design' ? <DesignSystem /> : (view === 'agents' || view === 'graphs') && project ? <ProjectAuthoring key={`${project.id}:${view}:${agentsVisit}:${graphsVisit}`} projectId={project.id} area={view} /> : !loaded ? <div className="empty-chat"><h2>{connectionError ? 'Engine disconnected' : 'Connecting to engine...'}</h2></div> : !project ? <div className="empty-chat"><h2>Add a project</h2><Button onClick={() => void openProjectPicker()}>Add project</Button></div> : !session ? <div className="empty-chat"><h2>Create a session</h2><Button onClick={() => newSession()} disabled={!!connectionError}>Create session</Button></div> : <>
        {activeTab === 'chat' && <ConversationHistory session={session} subagents={sessions.filter(item => item.parentId === activeId && item.role === 'subagent')} working={working} label="Conversation" onSnapshot={receiveSession} onHistory={receiveHistory} onAskSide={askSide} onOpenSubagent={openSubagent} empty={<div className="empty-chat"><div className="empty-chat-brand" role="img" aria-label="KLM Harness"><span className="empty-chat-wordmark" aria-hidden="true"><span>K</span><span>L</span><span>M</span></span><Badge>Harness</Badge></div><h2>What would you like to work on?</h2></div>} />}
        {activeTab.startsWith('subagent:') && <SubagentView session={activeSubagent} onSnapshot={receiveSession} onHistory={receiveHistory} />}
        {(catalogError || !!catalog?.errors.length) && <div role="alert" className="storage-error">{catalogError || catalog?.errors.join(' ')} <Button size="sm" disabled={catalogLoading} onClick={() => void reloadCatalog().catch(() => {})}>Retry</Button></div>}
        {graphErrors[session.id] && <div role="alert" className="storage-error">{graphErrors[session.id]} <Button size="sm" onClick={() => setRefreshVersion(current => current + 1)}>Retry</Button></div>}
        {selectedGraph ? <Suspense fallback={activeTab === 'graph' ? <section className="graph-view-loading" aria-label="Loading graph canvas" aria-busy="true" /> : null}><GraphView key={`${session.id}:${selectedGraphId}:${graphRunInProgress ? graphRun!.id : 'preview'}`} graph={selectedGraph} agents={catalog?.agents ?? []} visible={activeTab === 'graph'} run={graphRunInProgress ? graphRun : undefined} /></Suspense> : activeTab === 'graph' && <section className="graph-view-loading" role="status">{catalogLoading ? 'Loading graph...' : 'Selected graph is unavailable.'} <Button size="sm" disabled={catalogLoading} onClick={() => void reloadCatalog().catch(() => {})}>Reload</Button></section>}
        {sessionErrors[session.id] && <div role="alert" className="storage-error">{sessionErrors[session.id]}</div>}
        <div className="composer-area" hidden={activeTab !== 'chat'}>
          {(awaitingPermission || awaitingQuestion || graphRequests.length > 0) && <div className="permission-queue">
            {session.permissions?.map(permission => <PermissionCard key={permission.id} permission={permission} sessionId={session.id} projectName={project.name} onResolved={snapshot => setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [snapshot]) }))} />)}
            {session.questions?.map(question => <QuestionCard key={question.id} question={question} sessionId={session.id} onResolved={snapshot => setEngine(current => ({ ...current, sessions: mergeSessions(current.sessions, [snapshot]) }))} />)}
            {graphRequests.map(item => {
              const graphName = graphRun?.id === item.runId ? graphRun.snapshot.definition.name : catalog?.graphs.find(graph => graph.id === item.graphId)?.definition.name ?? item.graphId;
              const node = graphRun?.id === item.runId ? graphRun.snapshot.definition.nodes[item.nodeId] : undefined;
              const nodeName = node?.name || (node?.type === 'join' ? `Join · ${node.agent || item.nodeId}` : item.nodeName || item.nodeId);
              const origin = `${graphName || 'Graph'} · ${nodeName}`;
              const key = `${item.sessionId}:${item.kind}:${item.requestId}`;
              return item.kind === 'permission' && item.permission
                ? <PermissionCard key={key} permission={{ ...item.permission, id: item.requestId }} sessionId={item.sessionId} projectName={project.name} origin={origin} onResolved={() => graphRequestResolved(session.id)} />
                : item.kind === 'question' && item.question
                  ? <QuestionCard key={key} question={{ ...item.question, id: item.requestId }} sessionId={item.sessionId} origin={origin} onResolved={() => graphRequestResolved(session.id)} /> : null;
            })}
          </div>}
          <MessageComposer key={session.id} projectId={session.projectId} draft={drafts[session.id] ?? { text: '', mentions: [] }} onDraftChange={draft => setDrafts(current => ({ ...current, [session.id]: draft }))} onSend={send} disabled={!!connectionError || pendingSessions[session.id] || session.status === 'running' || awaitingPermission || awaitingQuestion} running={session.status === 'running'} onStop={() => void updateSession(session, 'stop', {})} graphControl={<GraphPicker key={session.id} graphs={catalogGraphs} value={selectedGraphId} selectedName={catalogSelectedGraph?.definition.name ?? selectedGraph?.definition.name} running={graphRunInProgress} disabled={!!connectionError || !!selectingGraphs[session.id] || !session.graph} loading={catalogLoading} error={catalogError || catalog?.errors.join(' ')} onReload={() => void reloadCatalog().catch(() => {})} onChange={graphId => void chooseGraph(session.id, graphId)} />} modelControl={<ModelPicker key={session.id} session={session} disabled={!!connectionError || !!pendingSessions[session.id] || working} onSave={(model, effort) => applyModelSettings(session, model, effort)} />} />
          <SessionStatusBar session={session} />
        </div>
      </>}
    </main>
    {sideVisible && project && session && <SideChatPanel key={session.id} session={sideSession} project={project} harnesses={harnesses}
      draft={drafts[`side:${session.id}`] ?? { text: '', mentions: [] }} sources={sideSources[session.id] ?? []}
      error={sideSession ? sessionErrors[sideSession.id] ?? '' : sideErrors[session.id] ?? ''} disconnected={!!connectionError} pending={!!sideSession && !!pendingSessions[sideSession.id]}
      onClose={() => setSideOpen(false)} onRetry={() => void ensureSide(session)} onSnapshot={receiveSession} onHistory={receiveHistory}
      onDraftChange={draft => setDrafts(current => ({ ...current, [`side:${session.id}`]: draft }))}
      onRemoveSource={index => setSideSources(current => ({ ...current, [session.id]: (current[session.id] ?? []).filter((_, i) => i !== index) }))}
      onSend={sendSide} onStop={() => { if (sideSession) void updateSession(sideSession, 'stop', {}); }}
      onHarness={harness => { if (sideSession) void updateSession(sideSession, 'harness', { harness }); }}
      onModel={(model, effort) => sideSession ? applyModelSettings(sideSession, model, effort) : Promise.resolve(false)} />}
    {addingProject !== null && <ProjectDialog initialFolder={addingProject} onSave={addProject} onClose={() => { navigation.current += 1; setAddingProject(null); }} />}
    {editingProject && <ProjectDialog key={`edit-project:${editingProject.id}`} initialFolder={editingProject.folder} project={editingProject} onSave={input => editProject(editingProject.id, input)} onClose={() => setEditingProject(null)} />}
    {creatingSession && createProject && <CreateSessionDialog project={createProject} harnesses={harnesses} workspace={creatingSession.workspace} onCreate={createSession} onClose={() => { navigation.current += 1; setCreatingSession(null); }} />}
    <dialog ref={settings} aria-label="Settings" className="settings-dialog"><div className="dialog-heading"><h2>Settings</h2><IconButton label="Close settings" onClick={() => settings.current?.close()}><X /></IconButton></div><form className="engine-settings-form" onSubmit={connectEngine}><label htmlFor="engine-url">Engine URL</label><Input id="engine-url" type="url" required autoCapitalize="none" autoCorrect="off" spellCheck={false} placeholder="http://localhost:7331" value={settingsEngineURL} aria-invalid={!!settingsError} aria-describedby={settingsError ? 'engine-url-error' : undefined} onChange={event => { setSettingsEngineURL(event.target.value); setSettingsError(''); }} />{settingsError && <p id="engine-url-error" className="form-error" role="alert">{settingsError}</p>}<div className="dialog-actions"><Button variant="primary" type="submit">Connect</Button></div></form><p>Projects and session history are saved locally by the engine. Unsent drafts are kept only until this page reloads.</p><Button onClick={() => { settings.current?.close(); setView('design'); }}>Design system</Button></dialog>
  </div>;
}
