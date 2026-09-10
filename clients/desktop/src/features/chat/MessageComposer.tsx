import { ArrowUp, Square } from 'lucide-react';
import { useEffect, useLayoutEffect, useRef, useState, type ReactNode } from 'react';
import { IconButton } from '../../design-system/Button';

export type ComposerDraft = { text: string; files: string[] };

function resizeInput(input: HTMLTextAreaElement) {
  const style = getComputedStyle(input);
  const maxHeight = 10 * parseFloat(style.lineHeight) + parseFloat(style.paddingTop) + parseFloat(style.paddingBottom);
  const scrollTop = input.scrollTop;
  input.style.overflowY = 'hidden';
  input.style.height = 'auto';
  input.style.height = `${Math.min(input.scrollHeight, maxHeight)}px`;
  input.style.overflowY = input.scrollHeight > maxHeight ? 'auto' : 'hidden';
  input.scrollTop = scrollTop;
}

export function MessageComposer({ onSend, draft, onDraftChange, disabled = false, running = false, onStop, modelControl, placeholder = 'Message the agent' }: {
  onSend: (text: string) => void | Promise<boolean | void>; draft: ComposerDraft; onDraftChange: (draft: ComposerDraft) => void;
  disabled?: boolean; running?: boolean; onStop?: () => void;
  modelControl?: ReactNode;
  placeholder?: string;
}) {
  const { text } = draft;
  const setText = (text: string) => onDraftChange({ ...draft, text });
  const sending = useRef(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const textarea = useRef<HTMLTextAreaElement>(null);
  const pendingCaret = useRef<number | null>(null);
  useLayoutEffect(() => {
    const input = textarea.current;
    if (!input) return;
    resizeInput(input);
    if (pendingCaret.current !== null) {
      input.setSelectionRange(pendingCaret.current, pendingCaret.current);
      if (pendingCaret.current === text.length) input.scrollTop = input.scrollHeight;
      pendingCaret.current = null;
    }
  }, [text]);
  useEffect(() => {
    const input = textarea.current;
    if (!input) return;
    let width = input.clientWidth;
    const observer = new ResizeObserver(() => {
      if (input.clientWidth !== width) { width = input.clientWidth; resizeInput(input); }
    });
    observer.observe(input);
    return () => observer.disconnect();
  }, []);
  async function send() {
    if (!text.trim() || disabled || running || sending.current) return;
    sending.current = true; setSubmitting(true); setError('');
    try {
      const accepted = await onSend(text.trim());
      if (accepted !== false) onDraftChange({ text: '', files: [] });
    } catch (error) { setError(error instanceof Error ? error.message : 'Message could not be sent. Retry.'); }
    finally { sending.current = false; setSubmitting(false); }
  }
  return <form className="composer" aria-label="Message composer" onSubmit={event => { event.preventDefault(); send(); }}>
    <textarea ref={textarea} aria-label={placeholder} placeholder={placeholder} rows={1} value={text} readOnly={submitting} onChange={event => setText(event.target.value)} onKeyDown={event => {
      if (event.nativeEvent.isComposing || submitting) return;
      if (event.ctrlKey && event.key.toLowerCase() === 'j') {
        event.preventDefault();
        const { selectionStart, selectionEnd } = event.currentTarget;
        pendingCaret.current = selectionStart + 1;
        setText(text.slice(0, selectionStart) + '\n' + text.slice(selectionEnd));
      } else if (event.key === 'Enter' && !event.shiftKey) { event.preventDefault(); send(); }
    }} />
    {error && <p role="alert" className="form-error composer-error">{error}</p>}
    <div className="composer-controls"><div className="composer-model-controls">{modelControl}{running && onStop ? <IconButton label="Stop response" variant="outline" onClick={onStop}><Square /></IconButton> : <IconButton label="Send message" variant="soft" type="submit" disabled={!text.trim() || disabled || submitting} className="send-button"><ArrowUp /></IconButton>}</div></div>
  </form>;
}
