import { useCallback, useEffect, useRef, useState, type FormEvent } from 'react';
import { ArrowLeft, Plus, Trash2 } from 'lucide-react';
import { Badge } from '../../design-system/Badge';
import { Button } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { ListTile } from '../../design-system/ListTile';
import { Toggle } from '../../design-system/Toggle';
import { EngineRequestError } from '../../engine';
import { beginLocalActivity } from '../engines/localActivities';
import { createGitIdentity, listGitIdentities, removeGitIdentity, testGitIdentity, updateGitIdentity, type GitIdentitiesResponse, type GitIdentity } from './gitIdentitiesApi';

type Draft = { identityId?: string; alias: string; token: string; replaceToken: boolean; makeDefault: boolean };
type Confirmation = 'replace' | 'default' | 'remove';

function failureMessage(error: unknown) {
  return error instanceof Error ? error.message : 'Could not update Git identities.';
}

export function ProjectGitIdentitiesSettings({ projectId, visible, onBusyChange }: {
  projectId: string;
  visible: boolean;
  onBusyChange: (busy: boolean) => void;
}) {
  const [identities, setIdentities] = useState<GitIdentity[]>([]);
  // Intentionally dialog-local: candidate tokens must never enter the draft store.
  const [draft, setDraft] = useState<Draft | null>(null);
  const retainedDrafts = useRef(new Map<string, Draft>());
  const [loading, setLoading] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const [listError, setListError] = useState('');
  const [error, setError] = useState('');
  const [verifiedLogin, setVerifiedLogin] = useState('');
  const [busy, setBusy] = useState<'saving' | 'testing' | 'default' | 'removing' | null>(null);
  const [confirmation, setConfirmation] = useState<Confirmation | null>(null);
  const active = useRef(false);
  const listing = useRef<AbortController | null>(null);
  const operation = useRef<AbortController | null>(null);
  const section = useRef<HTMLElement>(null);
  const returnFocus = useRef<HTMLElement | null>(null);
  const selected = identities.find(identity => identity.id === draft?.identityId);

  const load = useCallback(async () => {
    listing.current?.abort();
    const controller = new AbortController();
    listing.current = controller;
    setLoading(true); setListError('');
    try {
      const result = await listGitIdentities(projectId, controller.signal);
      if (!active.current || controller.signal.aborted || listing.current !== controller) return;
      setIdentities(result.identities); setLoaded(true);
    } catch (error) {
      if (active.current && !controller.signal.aborted && listing.current === controller) setListError(failureMessage(error));
    } finally {
      if (active.current && listing.current === controller) { listing.current = null; setLoading(false); }
    }
  }, [projectId]);

  useEffect(() => {
    active.current = true;
    return () => {
      active.current = false;
      listing.current?.abort(); operation.current?.abort();
      listing.current = null; operation.current = null;
      onBusyChange(false);
    };
  }, [onBusyChange]);

  useEffect(() => {
    if (visible && !loaded && !listError && !listing.current) void load();
  }, [visible, loaded, listError, load]);

  function restoreFocus() {
    requestAnimationFrame(() => {
      if (!active.current) return;
      const previous = returnFocus.current;
      const target = previous?.isConnected ? previous : section.current?.querySelector<HTMLElement>('[data-git-identity-add]');
      if (target?.getClientRects().length && !target.matches(':disabled')) target.focus();
    });
  }

  function edit(identity?: GitIdentity) {
    if (operation.current) return;
    returnFocus.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const key = identity?.id ?? '';
    const retained = retainedDrafts.current.get(key);
    retainedDrafts.current.delete(key);
    setDraft(retained ?? { identityId: identity?.id, alias: identity?.alias ?? '', token: '', replaceToken: !identity, makeDefault: identity?.isDefault ?? identities.length === 0 });
    setError(''); setVerifiedLogin(''); setConfirmation(null);
  }

  function closeEditor(discard = false) {
    if (draft && !discard) retainedDrafts.current.set(draft.identityId ?? '', draft);
    if (discard) retainedDrafts.current.delete(draft?.identityId ?? '');
    setDraft(null); setConfirmation(null); setError(''); setVerifiedLogin(''); restoreFocus();
  }

  async function run<T>(kind: NonNullable<typeof busy>, action: (signal: AbortSignal) => Promise<T>, onSuccess: (result: T) => void) {
    if (operation.current) return;
    const finished = beginLocalActivity('Updating Git identities');
    const controller = new AbortController();
    operation.current = controller;
    listing.current?.abort(); listing.current = null; setLoading(false);
    setBusy(kind); onBusyChange(true); setError(''); setVerifiedLogin('');
    try {
      const result = await action(AbortSignal.any([controller.signal, AbortSignal.timeout(15000)]));
      if (active.current && !controller.signal.aborted && operation.current === controller) onSuccess(result);
    } catch (error) {
      if (!active.current || controller.signal.aborted || operation.current !== controller) return;
      const uncertain = kind !== 'testing' && !(error instanceof EngineRequestError);
      setError(`${failureMessage(error)}${uncertain ? ' Change may be unconfirmed. Check the refreshed identities before trying again.' : ''}`);
      setConfirmation(null);
      if (uncertain) await load();
    } finally {
      finished();
      if (operation.current === controller) {
        operation.current = null;
        if (active.current) { setBusy(null); onBusyChange(false); }
      }
    }
  }

  function acceptList(result: GitIdentitiesResponse) {
    setIdentities(result.identities); setLoaded(true); setListError(''); closeEditor(true);
  }

  function validateDraft() {
    if (!draft) return false;
    if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]{0,39}$/.test(draft.alias)) {
      setError('Use 1–40 letters, digits, _ or -, starting with a letter or digit.'); return false;
    }
    if (identities.some(identity => identity.id !== draft.identityId && identity.alias.toLowerCase() === draft.alias.toLowerCase())) {
      setError('This alias is already in use.'); return false;
    }
    if (draft.replaceToken && !draft.token.trim()) { setError('Enter a GitHub token.'); return false; }
    return true;
  }

  function save() {
    if (!draft || !validateDraft()) return;
    const input = { alias: draft.alias, ...(draft.replaceToken ? { token: draft.token } : {}), ...(!draft.identityId ? { makeDefault: draft.makeDefault } : {}) };
    if (draft.identityId) {
      void run('saving', signal => updateGitIdentity(projectId, draft.identityId!, input, signal), acceptList);
    } else {
      void run('saving', signal => createGitIdentity(projectId, { ...input, token: draft.token }, signal), acceptList);
    }
  }

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (operation.current || confirmation || !validateDraft()) return;
    if (draft?.identityId && draft.replaceToken) setConfirmation('replace');
    else save();
  }

  function test() {
    if (!draft) return;
    if (draft.replaceToken && !draft.token.trim()) { setError('Enter a GitHub token.'); return; }
    const input = draft.replaceToken ? { token: draft.token } : { identityId: draft.identityId! };
    void run('testing', signal => testGitIdentity(projectId, input, signal), result => setVerifiedLogin(result.login));
  }

  return <section ref={section} className="git-identities-settings" hidden={!visible} aria-labelledby="project-git-identities-title">
    <div className="settings-panel-title"><h3 id="project-git-identities-title">Git identities</h3>
      {!draft && <Button data-git-identity-add size="sm" variant="primary" disabled={!!busy || loading || !loaded || !!listError} onClick={() => edit()}><Plus />Add</Button>}
    </div>
    {listError && <div className="git-identity-error"><p role="alert" className="form-error">{listError}</p><Button size="sm" disabled={!!busy || loading} onClick={() => void load()}>Retry</Button></div>}
    {!draft ? <>
      {loading && <p role="status" className="git-identity-hint">Loading Git identities…</p>}
      {loaded && !loading && !listError && identities.length === 0 && <p className="git-identity-hint">No Git identities. New work uses the shared Git account.</p>}
      <div className="git-identities-list" aria-busy={loading || !!busy}>
        {identities.map(identity => <ListTile key={identity.id} title={identity.alias} description={`@${identity.login}`} status={identity.isDefault ? <Badge>Default</Badge> : undefined} onClick={() => { if (!loading) edit(identity); }} />)}
      </div>
    </> : <>
      <div className="git-identity-editor-heading"><Button size="sm" variant="ghost" disabled={!!busy} onClick={() => closeEditor()}><ArrowLeft />Back</Button><h4>{draft.identityId ? 'Edit identity' : 'Add identity'}</h4></div>
      <form className="settings-form git-identity-form" onSubmit={submit} aria-busy={!!busy}>
        <fieldset disabled={!!busy || !!confirmation || loading}>
          <div className="settings-field"><label htmlFor="git-identity-alias">Alias</label><Input id="git-identity-alias" autoFocus={visible} required maxLength={40} autoComplete="off" autoCapitalize="none" spellCheck={false} value={draft.alias} onChange={event => { setDraft({ ...draft, alias: event.target.value }); setError(''); }} /></div>
          {selected && <div className="settings-field"><span className="git-identity-field-label">GitHub account</span><div className="git-identity-account"><span>@{selected.login}</span>{selected.isDefault && <Badge>Default</Badge>}</div></div>}
          {draft.identityId && <div className="settings-field"><Toggle label="Replace token" checked={draft.replaceToken} onCheckedChange={replaceToken => { setDraft({ ...draft, replaceToken, token: '' }); setVerifiedLogin(''); setError(''); }} /></div>}
          {draft.replaceToken && <div className="settings-field"><label htmlFor="git-identity-token">GitHub token</label><Input id="git-identity-token" type="password" required maxLength={4096} autoComplete="new-password" autoCapitalize="none" autoCorrect="off" spellCheck={false} value={draft.token} onChange={event => { setDraft({ ...draft, token: event.target.value }); setVerifiedLogin(''); setError(''); }} /></div>}
          {!draft.identityId && identities.length > 0 && <div className="settings-field"><Toggle label="Make default" checked={draft.makeDefault} onCheckedChange={makeDefault => setDraft({ ...draft, makeDefault })} /></div>}
        </fieldset>
        {verifiedLogin && <p className="settings-saved" role="status">Verified as @{verifiedLogin}</p>}
        {error && <p role="alert" className="form-error">{error}</p>}
        {!confirmation && <div className="dialog-actions">
          <Button disabled={!!busy || loading || draft.replaceToken && !draft.token.trim()} onClick={test}>{busy === 'testing' ? 'Testing…' : 'Test'}</Button>
          <Button type="submit" variant="primary" disabled={!!busy || loading || !draft.alias || draft.replaceToken && !draft.token.trim() || !!draft.identityId && !draft.replaceToken && draft.alias === selected?.alias}>{busy === 'saving' ? 'Saving…' : draft.identityId && draft.replaceToken ? 'Replace' : 'Save'}</Button>
        </div>}
      </form>
      {selected && !confirmation && <div className="git-identity-management">
        {!selected.isDefault && <Button size="sm" disabled={!!busy || loading} onClick={() => { setError(''); setConfirmation('default'); }}>Set default</Button>}
        <Button size="sm" variant="danger-soft" disabled={!!busy || loading} onClick={() => { setError(''); setConfirmation('remove'); }}><Trash2 />Remove</Button>
      </div>}
      {confirmation && <div className="git-identity-confirmation" role="group" aria-labelledby="git-identity-confirm-title">
        <h4 id="git-identity-confirm-title">{confirmation === 'replace' ? `Replace token for ${selected?.alias}?` : confirmation === 'default' ? `Make ${selected?.alias} the default?` : `Remove ${selected?.alias}?`}</h4>
        <p className="git-identity-hint">{confirmation === 'replace' ? 'Future commands using the previous token will fail.' : confirmation === 'default' ? 'New work will use this identity by default.' : identities.length === 1 ? 'New work will use the shared Git account.' : selected?.isDefault ? 'The oldest remaining identity will become the default.' : 'Future commands using this identity will fail.'}</p>
        <div className="dialog-actions"><Button size="sm" autoFocus disabled={!!busy} onClick={() => setConfirmation(null)}>Cancel</Button>
          <Button size="sm" variant={confirmation === 'remove' ? 'danger-soft' : 'primary'} disabled={!!busy} onClick={() => {
            if (confirmation === 'replace') save();
            else if (selected) void run(confirmation === 'remove' ? 'removing' : 'default', signal => confirmation === 'remove' ? removeGitIdentity(projectId, selected.id, signal) : updateGitIdentity(projectId, selected.id, { makeDefault: true }, signal), result => {
              setIdentities(result.identities); setConfirmation(null);
              if (confirmation === 'remove') closeEditor(true);
            });
          }}>{busy ? busy === 'removing' ? 'Removing…' : 'Saving…' : confirmation === 'remove' ? 'Remove' : confirmation === 'replace' ? 'Replace' : 'Set default'}</Button>
        </div>
      </div>}
    </>}
  </section>;
}
