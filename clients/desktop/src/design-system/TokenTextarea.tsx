import { useLayoutEffect, useRef, type ReactNode, type TextareaHTMLAttributes } from 'react';
import { Textarea } from './Textarea';

export type TextToken = { start: number; end: number };

export function TokenTextarea({ value, tokens, className = '', onScroll, style, ...props }: Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'value' | 'defaultValue' | 'children'> & {
  value: string;
  tokens: readonly TextToken[];
}) {
  const root = useRef<HTMLDivElement>(null);
  const mirror = useRef<HTMLDivElement>(null);
  const content: ReactNode[] = [];
  let offset = 0;
  for (const token of [...tokens].sort((a, b) => a.start - b.start)) {
    if (token.start < offset || token.end > value.length || token.start >= token.end) continue;
    content.push(value.slice(offset, token.start));
    content.push(<span key={`${token.start}:${token.end}`} className="token-textarea-badge">{value.slice(token.start, token.end)}</span>);
    offset = token.end;
  }
  content.push(value.slice(offset));
  // Preserve the final empty line in the visual layer without changing the stored value.
  if (value.endsWith('\n')) content.push('\u200b');

  function syncScroll(input: HTMLTextAreaElement) {
    if (!mirror.current) return;
    mirror.current.scrollTop = input.scrollTop;
    mirror.current.scrollLeft = input.scrollLeft;
  }

  useLayoutEffect(() => {
    const input = root.current?.querySelector('textarea');
    if (input) syncScroll(input);
  }, [value]);

  // Keep editing, selection, copy/paste and undo in a native textarea; the mirror only paints tokens.
  return <div ref={root} className="token-textarea">
    <div ref={mirror} aria-hidden="true" className={`input textarea token-textarea-mirror ${className}`} style={style}>{content}</div>
    <Textarea {...props} value={value} className={`token-textarea-input ${className}`} style={style} onScroll={event => { syncScroll(event.currentTarget); onScroll?.(event); }} />
  </div>;
}
