import { ArrowUp, FileText, Folder, Square } from 'lucide-react';
import { useEffect, useId, useRef, useState, type ReactNode } from 'react';
import { IconButton } from '../../design-system/Button';
import { request, type FileMention, type ProjectPath } from '../../engine';
import { mentionQuery } from './mentions';
import { MentionEditor, type MentionEditorHandle } from './MentionEditor';

export type ComposerDraft = { text: string; mentions: FileMention[] };

export function MessageComposer({ onSend, draft, onDraftChange, projectId, disabled = false, running = false, onStop, modelControl, placeholder = 'Message the agent' }: {
  onSend: (draft: ComposerDraft) => void | Promise<boolean | void>; draft: ComposerDraft; onDraftChange: (draft: ComposerDraft) => void;
  projectId?: string;
  disabled?: boolean; running?: boolean; onStop?: () => void;
  modelControl?: ReactNode;
  placeholder?: string;
}) {
  const { text, mentions } = draft;
  const [caret, setCaret] = useState(0);
  const [focused, setFocused] = useState(false);
  const [composing, setComposing] = useState(false);
  const [dismissed, setDismissed] = useState('');
  const [active, setActive] = useState(0);
  const listId = useId();
  const sending = useRef(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const editor = useRef<MentionEditorHandle>(null);
  const query = focused && !composing && !submitting && projectId ? mentionQuery(text, caret, mentions) : null;
  const queryKey = query ? JSON.stringify([projectId, query.start, query.query]) : '';
  const open = !!queryKey && dismissed !== queryKey;
  const [search, setSearch] = useState<{ key: string; paths: ProjectPath[]; partial: boolean; error: string }>({ key: '', paths: [], partial: false, error: '' });
  const paths = search.key === queryKey ? search.paths : [];
  const loading = search.key !== queryKey;
  const searchText = query?.query;
  useEffect(() => {
    if (!open || !projectId || searchText === undefined) return;
    let current = true;
    setActive(0);
    const timer = window.setTimeout(() => {
      request<{ paths: ProjectPath[]; partial: boolean }>(`/api/projects/${encodeURIComponent(projectId)}/paths?q=${encodeURIComponent(searchText)}`)
        .then(result => { if (current) setSearch({ key: queryKey, ...result, error: '' }); })
        .catch(error => { if (current) setSearch({ key: queryKey, paths: [], partial: false, error: error instanceof Error ? error.message : 'Could not search project paths.' }); });
    }, 150);
    return () => { current = false; window.clearTimeout(timer); };
  }, [open, projectId, searchText, queryKey]);
  useEffect(() => {
    if (open) document.getElementById(`${listId}-${active}`)?.scrollIntoView({ block: 'nearest' });
  }, [active, open, listId]);
  function selectPath(path: ProjectPath) {
    if (!query) return;
    editor.current?.insertMention(path, query);
    setDismissed(queryKey);
  }
  async function send() {
    if (!text.trim() || disabled || running || sending.current) return;
    sending.current = true; setSubmitting(true); setError('');
    try {
      const accepted = await onSend(draft);
      if (accepted !== false) onDraftChange({ text: '', mentions: [] });
    } catch (error) { setError(error instanceof Error ? error.message : 'Message could not be sent. Retry.'); }
    finally { sending.current = false; setSubmitting(false); }
  }
  return <form className="composer" aria-label="Message composer" onSubmit={event => { event.preventDefault(); send(); }}>
    {open && <div className="mention-picker">
      <ul id={listId} role="listbox" aria-label="Project files and folders">{paths.map((path, index) => {
        const Icon = path.kind === 'directory' ? Folder : FileText;
        return <li key={path.path} id={`${listId}-${index}`} role="option" aria-selected={index === active} className={index === active ? 'is-active' : ''}
          onMouseDown={event => event.preventDefault()} onClick={() => selectPath(path)} onMouseMove={() => setActive(index)}>
          <Icon aria-hidden="true" /><span><strong>{path.path.split('/').pop()}{path.kind === 'directory' ? '/' : ''}</strong><small>{path.path}</small></span><span className="sr-only">{path.kind === 'directory' ? 'Folder' : 'File'}</span>
        </li>;
      })}</ul>
      <div className="mention-search-status" role="status">{loading ? 'Searching…' : search.error || (!paths.length ? 'No matching paths.' : search.partial ? 'Partial results. Refine your search.' : '')}</div>
    </div>}
    <MentionEditor ref={editor} aria-label={placeholder} placeholder={placeholder} draft={draft} readOnly={submitting}
      aria-autocomplete={projectId ? 'list' : undefined} aria-controls={open ? listId : undefined} aria-expanded={open} aria-haspopup="listbox" aria-activedescendant={open && paths[active] ? `${listId}-${active}` : undefined}
      onFocus={() => setFocused(true)} onBlur={() => setFocused(false)}
      onCompositionStart={() => setComposing(true)} onCompositionEnd={() => setComposing(false)}
      onCaretChange={setCaret}
      onChange={next => { onDraftChange(next); setDismissed(''); }} onKeyDown={event => {
      if (event.nativeEvent.isComposing || event.nativeEvent.keyCode === 229 || composing || submitting) return;
      if (open && !event.shiftKey && ['ArrowDown', 'ArrowUp', 'Enter', 'Tab', 'Escape'].includes(event.key)) {
        event.preventDefault(); event.stopPropagation();
        if (event.key === 'Escape') setDismissed(queryKey);
        else if (event.key === 'ArrowDown' || event.key === 'ArrowUp') setActive(index => paths.length ? (index + (event.key === 'ArrowDown' ? 1 : -1) + paths.length) % paths.length : 0);
        else if (paths[active]) selectPath(paths[active]);
        return;
      }
      if ((event.ctrlKey && event.key.toLowerCase() === 'j') || (event.key === 'Enter' && event.shiftKey)) {
        event.preventDefault();
        editor.current?.insertText('\n');
      } else if (event.key === 'Enter' && !event.shiftKey) { event.preventDefault(); send(); }
    }} />
    {error && <p role="alert" className="form-error composer-error">{error}</p>}
    <div className="composer-controls"><div className="composer-model-controls">{modelControl}{running && onStop ? <IconButton label="Stop response" variant="outline" onClick={onStop}><Square /></IconButton> : <IconButton label="Send message" variant="soft" type="submit" disabled={!text.trim() || disabled || submitting} className="send-button"><ArrowUp /></IconButton>}</div></div>
  </form>;
}
