import { useEffect, useRef, useState } from 'react';
import { X } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import type { Harness, Project } from '../../engine';
import { bindDraftDiscard, flushDrafts, useEngineDraft } from '../drafts/draftStore';
import { beginLocalActivity } from '../engines/localActivities';

export function CreateSessionDialog({ project, harnesses, topicId, onCreate, onClose }: {
  project: Project; harnesses: Harness[]; topicId?: string;
  onCreate: (input: { projectId: string; title: string; harness: Harness['id']; model?: string; topicId?: string }) => Promise<string | null>;
  onClose: () => void;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const key = `form.session.new:${project.id}:${topicId ?? ''}`;
  const forgetDrafts = bindDraftDiscard();
  const [harness, setHarness] = useEngineDraft<Harness['id']>(`${key}.harness`, harnesses.find(item => item.available)?.id ?? 'pi');
  const [title, setTitle] = useEngineDraft(`${key}.title`, '');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  useEffect(() => { dialog.current?.showModal(); }, []);
  const selected = harnesses.find(item => item.id === harness);
  return <dialog ref={dialog} className="settings-dialog add-project-dialog" aria-labelledby="create-session-title" onCancel={event => { if (busy) event.preventDefault(); else onClose(); }} onClose={onClose}>
    <form onSubmit={async event => {
      event.preventDefault();
      if (busy || !selected?.available) return;
      const finished = beginLocalActivity('Creating session');
      setBusy(true); setError('');
      try {
        await flushDrafts();
        const failure = await onCreate({ projectId: project.id, title: title.trim() || 'New session', harness, ...(topicId ? { topicId } : {}) });
        if (failure) setError(failure); else { forgetDrafts(key); onClose(); }
      } catch (error) { setError(error instanceof Error ? error.message : 'Could not create the session.'); }
      finally { setBusy(false); finished(); }
    }}>
      <div className="dialog-heading"><h2 id="create-session-title">New session</h2><IconButton label="Cancel creating session" disabled={busy} onClick={onClose}><X /></IconButton></div>
      <label className="project-field" htmlFor="session-title">Session name</label>
      <input id="session-title" autoFocus maxLength={200} placeholder="New session" value={title} onChange={event => setTitle(event.target.value)} disabled={busy} />
      <fieldset className="harness-picker" disabled={busy}><legend className="project-field">Harness</legend>{harnesses.map(item => <label key={item.id} className={`harness-option ${item.id === harness ? 'is-active' : ''}`}><input type="radio" name="harness" value={item.id} checked={item.id === harness} disabled={!item.available} onChange={() => setHarness(item.id)} /><span>{item.name}</span>{!item.available && <small>Not installed</small>}</label>)}</fieldset>
      {!selected?.available && <p className="form-error" role="alert">{selected?.error ?? 'No harness is available. Install Pi, OpenCode, or Codex, then restart the engine.'}</p>}
      {error && <p role="alert" className="form-error">{error}</p>}
      <div className="dialog-actions"><Button disabled={busy} onClick={() => { forgetDrafts(key); onClose(); }}>Cancel</Button><Button type="submit" variant="primary" disabled={busy || !selected?.available}>{busy ? 'Creating...' : 'Create session'}</Button></div>
    </form>
  </dialog>;
}
