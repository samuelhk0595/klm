import { useEffect, useRef, useState } from 'react';
import { AlertCircle, BookOpen, LockKeyhole, X } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { request, type Session } from '../../engine';
import { capabilityPath, emptySelection, selectionKey, selectionOf, type ConversationCapabilities } from './api';
import { SessionCapabilityPicker } from './SessionCapabilityPicker';
import './session-library.css';

export function SessionLibraryDialog({ session, onClose }: { session: Session; onClose: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null);
  const [state, setState] = useState<ConversationCapabilities>();
  const [draft, setDraft] = useState(emptySelection);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [attempt, setAttempt] = useState(0);
  const [confirmation, setConfirmation] = useState<'close' | 'reload' | null>(null);
  const locked = session.automation?.status === 'running';
  const dirty = !!state && selectionKey(draft) !== selectionKey(state.desired);
  const pending = !!state && (!state.applied || selectionKey(state.desired) !== selectionKey(state.applied.selection));
  const displayed = locked && state?.applied ? selectionOf(state.applied.selection) : draft;
  const toolCount = displayed.mcpConnections.reduce((total, connection) => total + connection.tools.length, 0);

  useEffect(() => {
    const element = dialog.current;
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    if (element && !element.open) element.showModal();
    return () => { queueMicrotask(() => { if (!element?.isConnected && previous?.isConnected) previous.focus(); }); };
  }, []);
  useEffect(() => {
    const controller = new AbortController();
    setLoading(true); setError('');
    request<ConversationCapabilities>(capabilityPath(session.id), 'GET', undefined, 15000, controller.signal).then(value => {
      if (!controller.signal.aborted) { setState(value); setDraft(selectionOf(value.desired)); }
    }, reason => {
      if (!controller.signal.aborted) setError(reason instanceof Error ? reason.message : 'Could not load selection.');
    }).finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [session.id, attempt]);
  useEffect(() => {
    if (!confirmation) return;
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    dialog.current?.querySelector<HTMLButtonElement>('.session-library-confirmation button')?.focus();
    return () => { if (previous?.isConnected) previous.focus(); };
  }, [confirmation]);

  function close() {
    if (saving) return;
    if (dirty) setConfirmation('close');
    else onClose();
  }
  function reload() {
    if (saving) return;
    if (dirty) setConfirmation('reload');
    else setAttempt(n => n + 1);
  }
  async function save() {
    if (!state || loading || saving || locked || !dirty) return;
    setSaving(true); setError('');
    try {
      await request<ConversationCapabilities>(capabilityPath(session.id), 'PATCH', { revision: state.revision, selection: draft });
      onClose();
    } catch (reason) { setError(reason instanceof Error ? reason.message : 'Could not save selection.'); }
    finally { setSaving(false); }
  }

  return <dialog ref={dialog} className="settings-dialog session-library-dialog" aria-labelledby="session-library-title" aria-describedby="session-library-session" onCancel={event => { event.preventDefault(); close(); }} onClose={onClose}>
    <header className="session-library-heading"><div><h2 id="session-library-title"><BookOpen aria-hidden="true" />Session Library</h2><p id="session-library-session">{session.title}</p></div><IconButton label="Close Library" disabled={saving} onClick={close}><X /></IconButton></header>
    {error && <div className="session-library-error" role="alert"><AlertCircle aria-hidden="true" /><p>{error}</p><Button size="sm" disabled={saving || loading} onClick={reload}>Reload selection</Button></div>}
    {locked && <div className="session-library-lock" role="status"><LockKeyhole aria-hidden="true" /><p>Read-only during this automation run.</p></div>}
    {!!state?.applied?.diagnostics?.length && <details className="session-library-diagnostics"><summary>Current turn notices <span>{state.applied.diagnostics.length}</span></summary><div>{state.applied.diagnostics.map(message => <p key={message}>{message}</p>)}</div></details>}
    <SessionCapabilityPicker value={displayed} onChange={setDraft} harness={session.harness} general={session.role === 'general_agent'} disabled={!state || loading || saving || locked} />
    <footer className="session-library-footer">
      {confirmation ? <div className="session-library-confirmation" role="alert"><div><strong>{confirmation === 'close' ? 'Discard unsaved changes?' : 'Reload the saved selection?'}</strong><p>Your changes in this dialog will be discarded.</p></div><div><Button onClick={() => setConfirmation(null)}>Keep editing</Button><Button variant="danger-soft" onClick={() => { if (confirmation === 'close') onClose(); else { setConfirmation(null); setAttempt(n => n + 1); } }}>{confirmation === 'close' ? 'Discard changes' : 'Reload selection'}</Button></div></div> : <>
        <div className="session-library-summary"><strong>{loading ? 'Loading selection…' : !state ? 'Selection unavailable' : `${displayed.skills.length} ${displayed.skills.length === 1 ? 'skill' : 'skills'} · ${displayed.mcpConnections.length} ${displayed.mcpConnections.length === 1 ? 'connection' : 'connections'} · ${toolCount} ${toolCount === 1 ? 'tool' : 'tools'}`}</strong>{state && <span aria-live="polite">{locked ? 'Current automation selection' : dirty ? 'Unsaved changes · Applies next turn' : pending ? 'Saved selection · Applies next turn' : 'Selection up to date'}</span>}</div>
        <div className="session-library-actions"><Button disabled={saving} onClick={close}>{dirty ? 'Cancel' : 'Close'}</Button>{!locked && <Button variant="primary" disabled={!state || loading || saving || !dirty} onClick={() => void save()}>{saving ? 'Saving…' : 'Save selection'}</Button>}</div>
      </>}
    </footer>
  </dialog>;
}
