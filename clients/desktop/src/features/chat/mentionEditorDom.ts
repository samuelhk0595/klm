import type { FileMention } from '../../engine';
import { mentionToken, type TextSelection } from './mentions';

type Segment = { node: Node; start: number; end: number; atomic: boolean };

// Map visible DOM positions to the canonical message, where a short badge still
// occupies the UTF-16 length of its full @project/path token.
export function readMentionEditor(root: HTMLElement, known: Map<string, FileMention>) {
  let text = '';
  const mentions: FileMention[] = [];
  const segments: Segment[] = [];
  const boundaries = new Map<Node, number[]>();
  const visit = (node: Node) => {
    const start = text.length;
    if (node instanceof Text) {
      text += node.data;
      segments.push({ node, start, end: text.length, atomic: false });
      return;
    }
    if (!(node instanceof HTMLElement)) return;
    const mention = known.get(node.dataset.mentionId ?? '');
    if (mention) {
      text += mentionToken(mention);
      mentions.push({ ...mention, start, end: text.length });
      segments.push({ node, start, end: text.length, atomic: true });
      return;
    }
    if (node.tagName === 'BR') {
      // A sole BR is the browser's empty-line placeholder, not message content.
      if (node.parentNode?.childNodes.length !== 1) text += '\n';
      segments.push({ node, start, end: text.length, atomic: true });
      return;
    }
    const offsets = [text.length];
    Array.from(node.childNodes).forEach((child, index) => {
      // Native IME/mobile editing can produce blocks instead of newline text nodes.
      if (index > 0 && child instanceof HTMLElement && /^(DIV|P)$/.test(child.tagName)) text += '\n';
      visit(child);
      offsets.push(text.length);
    });
    boundaries.set(node, offsets);
  };
  visit(root);
  const offsetAt = (node: Node, offset: number): number => {
    const segment = segments.find(segment => segment.node === node || (segment.atomic && segment.node.contains(node)));
    if (segment) return segment.atomic ? (offset === 0 ? segment.start : segment.end) : segment.start + Math.min(offset, segment.end - segment.start);
    return boundaries.get(node)?.[offset] ?? text.length;
  };
  const selection = window.getSelection();
  let range: TextSelection | null = null;
  if (selection?.anchorNode && selection.focusNode && root.contains(selection.anchorNode) && root.contains(selection.focusNode)) {
    const anchor = offsetAt(selection.anchorNode, selection.anchorOffset);
    const focus = offsetAt(selection.focusNode, selection.focusOffset);
    range = { start: Math.min(anchor, focus), end: Math.max(anchor, focus) };
  }
  return { text, mentions, segments, range };
}

export function selectMentionEditor(root: HTMLElement, known: Map<string, FileMention>, start: number, end = start) {
  const { segments } = readMentionEditor(root, known);
  const range = document.createRange();
  const point = (offset: number): [Node, number] => {
    for (const segment of segments) {
      if (offset > segment.end) continue;
      if (!segment.atomic) return [segment.node, Math.max(0, offset - segment.start)];
      const parent = segment.node.parentNode!;
      const index = Array.prototype.indexOf.call(parent.childNodes, segment.node) as number;
      return [parent, index + (offset > segment.start ? 1 : 0)];
    }
    return [root, root.childNodes.length];
  };
  range.setStart(...point(start));
  range.setEnd(...point(end));
  const selection = window.getSelection();
  selection?.removeAllRanges();
  selection?.addRange(range);
}
