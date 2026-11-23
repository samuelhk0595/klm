import { afterEach, expect, it, vi } from 'vitest';
import type { EngineEvent } from '../../engine';
import { prepareWindowChange } from './historyWindow';

afterEach(() => { document.body.replaceChildren(); vi.restoreAllMocks(); });
const events = (ids: string[]): EngineEvent[] => ids.map(id => ({ id, type: 'tool', text: '', createdAt: '' }));
function box(element: HTMLElement, top: number, height: number) {
  const rect = { x: 0, y: top, top, bottom: top + height, left: 0, right: 800, width: 800, height, toJSON: () => ({}) } as DOMRect;
  vi.spyOn(element, 'getBoundingClientRect').mockReturnValue(rect);
  vi.spyOn(element, 'getClientRects').mockReturnValue([rect] as unknown as DOMRectList);
}
function fixture(workTop: number) {
  const root = document.createElement('div'); root.className = 'chat-history';
  const body = document.createElement('div'); body.className = 'conversation-events';
  const work = document.createElement('div'); work.dataset.historyAnchor = 'tool-0'; work.dataset.historyEvents = 'tool-0 tool-1';
  const answer = document.createElement('div'); answer.dataset.historyAnchor = 'answer';
  body.append(work, answer); root.append(body); document.body.append(root);
  box(root, 0, 600); box(body, 0, 200); box(work, workTop, 40); box(answer, 80, 100);
  return root;
}

it('allows compact history paging when the reading anchor survives but the opposite visible answer is evicted', () => {
  const root = fixture(10);
  const restore = prepareWindowChange(root, { events: events(['tool-0', 'tool-1', 'answer']), startIndex: 25 }, { events: events(['older', 'tool-0', 'tool-1']), startIndex: 24, hasEarlier: true, hasLater: true });
  expect(restore).toBeTypeOf('function');
});

it('defers a delayed page when the user has moved to the answer that would be evicted', () => {
  const root = fixture(-100);
  const restore = prepareWindowChange(root, { events: events(['tool-0', 'tool-1', 'answer']), startIndex: 25 }, { events: events(['older', 'tool-0', 'tool-1']), startIndex: 24, hasEarlier: true, hasLater: true });
  expect(restore).toBeUndefined();
});
