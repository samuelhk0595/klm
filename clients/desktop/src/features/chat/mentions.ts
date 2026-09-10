import type { FileMention } from '../../engine';

export type TextSelection = { start: number; end: number };

export function mentionToken(mention: { path: string; kind: string }) {
  return `@${mention.path}${mention.kind === 'directory' ? '/' : ''}`;
}

export function mentionQuery(text: string, caret: number, mentions: FileMention[]) {
  if (caret < 0 || caret > text.length) return null;
  const start = text.lastIndexOf('@', caret - 1);
  if (start < 0 || (start > 0 && !/[\s([{]/.test(text[start - 1])) || mentions.some(mention => mention.start === start)) return null;
  const query = text.slice(start + 1, caret);
  if (query.includes('\n') || query.includes('\r') || query.length > 1024) return null;
  return { start, end: caret, query };
}
