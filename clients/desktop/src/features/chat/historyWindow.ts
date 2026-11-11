import type { EngineEvent } from '../../engine';
import { scrollRoot } from '../pagination/pagination';
import { captureScrollAnchor } from '../pagination/scrollAnchor';

export type HistoryWindow = { events: EngineEvent[]; startIndex: number; hasEarlier?: boolean; hasLater?: boolean };
const selector = '[data-history-anchor], [data-work-id]';
const ids = (element: HTMLElement) => (element.dataset.historyEvents || element.dataset.workId || element.dataset.historyAnchor || '').split(' ');

function visible(item: HTMLElement, root: HTMLElement) {
  if (!item.getClientRects().length) return false;
  const rect = item.getBoundingClientRect();
  let top = Math.max(0, root.getBoundingClientRect().top);
  let bottom = Math.min(innerHeight, root.getBoundingClientRect().bottom);
  for (let parent = item.parentElement; parent && parent !== root; parent = parent.parentElement) {
    if (/(auto|scroll|hidden|clip)/.test(getComputedStyle(parent).overflowY)) {
      const clip = parent.getBoundingClientRect(); top = Math.max(top, clip.top); bottom = Math.min(bottom, clip.bottom);
    }
  }
  return rect.bottom > top && rect.top < bottom;
}

// Geometry belongs to this reading surface, never to every viewer of a session.
// Returning undefined defers prefetch when the user moved into the eviction edge
// while its request was in flight.
export function prepareWindowChange(root: HTMLElement, previous: HistoryWindow, next: HistoryWindow): (() => void) | undefined {
  const retained = new Set(next.events.map(event => event.id));
  const content = root.querySelector<HTMLElement>(':scope > .conversation-events');
  const candidates = content ? [...content.querySelectorAll<HTMLElement>(selector)].filter(item => item.closest('.chat-history, .graph-run-timeline') === root) : [];
  const anchor = candidates.find(item => !item.querySelector('[data-work-id]') && visible(item, root));
  // Protect the reading anchor, not every visible row. Compact work can make
  // both ends visible at once; vetoing the opposite edge would strand history
  // even after an explicit upward gesture. Its evicted space is retained below.
  if (anchor && !ids(anchor).some(id => retained.has(id))) return;

  const scrollers = [...root.querySelectorAll<HTMLElement>('.agent-work__items')].filter(item => item.getClientRects().length && item.closest('.chat-history, .graph-run-timeline') === root);
  const snapshots = [...scrollers, root].map(container => {
    const inner = container !== root;
    const body = inner ? container : container.querySelector<HTMLElement>(':scope > .conversation-events');
    const restore = captureScrollAnchor(container, inner ? '[data-work-id]' : selector, inner);
    const measure = (item: HTMLElement) => {
      const bounds = body!.getBoundingClientRect(); const rect = item.getBoundingClientRect();
      const top = rect.top - bounds.top + (inner ? container.scrollTop : 0);
      const bottom = (inner ? container.scrollHeight : bounds.height) - (rect.bottom - bounds.top + (inner ? container.scrollTop : 0));
      return { top, bottom };
    };
    const markers = body ? [...body.querySelectorAll<HTMLElement>(inner ? '[data-work-id]' : ':scope > [data-history-anchor]')].map(item => ({ keys: ids(item), ...measure(item) })) : [];
    const space = (direction: string) => [...container.querySelectorAll<HTMLElement>(`[data-history-space="${direction}"]`)].find(item => scrollRoot(item) === container);
    return () => {
      const surviving = body ? markers.flatMap(marker => {
        let target: HTMLElement | null = null;
        for (const id of marker.keys) {
          const key = CSS.escape(id);
          target = body.querySelector<HTMLElement>(inner ? `[data-work-id="${key}"]` : `:scope > [data-history-anchor="${key}"], :scope > [data-history-events~="${key}"]`);
          if (target) break;
        }
        return target ? [{ before: marker, after: measure(target) }] : [];
      }) : [];
      const first = surviving[0]; const last = surviving.at(-1);
      const before = space('before'); const after = space('after');
      if (first && last) {
        const startDelta = next.startIndex - previous.startIndex;
        const endDelta = next.startIndex + next.events.length - previous.startIndex - previous.events.length;
        if (before && startDelta) {
          const delta = first.before.top - first.after.top;
          before.style.height = `${Math.max(0, (parseFloat(before.style.height) || 0) + (startDelta > 0 ? Math.max(0, delta) : Math.min(0, delta)))}px`;
        }
        if (after && endDelta) {
          const delta = last.before.bottom - last.after.bottom;
          after.style.height = `${Math.max(0, (parseFloat(after.style.height) || 0) + (endDelta < 0 ? Math.max(0, delta) : Math.min(0, delta)))}px`;
        }
      }
      if (before && !next.hasEarlier) before.style.height = '0px';
      if (after && !next.hasLater) after.style.height = '0px';
      restore(); container.dispatchEvent(new Event('klm:history-position'));
    };
  });
  return () => snapshots.forEach(restore => restore());
}
