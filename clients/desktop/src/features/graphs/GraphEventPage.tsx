import { useEffect, useRef, useState } from 'react';
import type { GraphNodeActivity } from '../../engine';
import { AutoPage } from '../pagination/AutoPage';
import { scrollRoot } from '../pagination/pagination';

// This edge uses graph offsets. It can live either in the outer timeline or in
// an expanded work group's presentation slot, with that group's actual scroller.
export function GraphEventPage({ activity, identity, direction, busy, error, onLoad, onRetry }: {
  activity: NonNullable<GraphNodeActivity['activation']>; identity: string; direction: 'before' | 'after';
  busy: boolean; error: string; onLoad: (offset: number, explicit: boolean) => void; onRetry: () => void;
}) {
  const target = useRef<HTMLDivElement>(null);
  const explicit = useRef(false);
  const [intent, setIntent] = useState<'before' | 'after'>('before');
  const [gesture, setGesture] = useState(0);
  useEffect(() => {
    const root = target.current?.closest<HTMLElement>('.agent-work__items, .graph-run-timeline');
    if (!root) return;
    let lastScroll = root.scrollTop, lastGesture = 0, touch: number | undefined;
    const position = () => { lastScroll = root.scrollTop; };
    const navigate = (event: Event) => {
      if (event.target instanceof HTMLElement && event.target !== root && scrollRoot(event.target) !== root) return;
      let next: 'before' | 'after';
      if (event instanceof WheelEvent) { if (!event.deltaY) return; next = event.deltaY < 0 ? 'before' : 'after'; }
      else if (event instanceof KeyboardEvent) {
        if (event.target instanceof Element && event.target.closest('input, textarea, select, button, [contenteditable="true"]')) return;
        if (!['ArrowUp', 'PageUp', 'Home', 'ArrowDown', 'PageDown', 'End', ' '].includes(event.key)) return;
        next = ['ArrowUp', 'PageUp', 'Home'].includes(event.key) || event.key === ' ' && event.shiftKey ? 'before' : 'after';
      } else if (typeof TouchEvent !== 'undefined' && event instanceof TouchEvent) {
        const y = event.touches[0]?.clientY; if (y === undefined) return;
        if (touch === undefined || touch === y) { touch = y; return; }
        next = y > touch ? 'before' : 'after'; touch = y;
      } else {
        const delta = root.scrollTop - lastScroll; lastScroll = root.scrollTop;
        if (Math.abs(delta) < 1) return; next = delta < 0 ? 'before' : 'after';
      }
      setIntent(next); explicit.current = true;
      if (Date.now() - lastGesture > 350) { lastGesture = Date.now(); setGesture(value => value + 1); }
    };
    const touchStart = () => { touch = undefined; };
    root.addEventListener('wheel', navigate, { passive: true }); root.addEventListener('keydown', navigate);
    root.addEventListener('scroll', navigate, { passive: true }); root.addEventListener('touchmove', navigate, { passive: true });
    root.addEventListener('touchstart', touchStart, { passive: true }); root.addEventListener('klm:history-position', position);
    return () => {
      root.removeEventListener('wheel', navigate); root.removeEventListener('keydown', navigate); root.removeEventListener('scroll', navigate);
      root.removeEventListener('touchmove', navigate); root.removeEventListener('touchstart', touchStart); root.removeEventListener('klm:history-position', position);
    };
  }, [identity, activity.id]);
  const more = direction === 'before' ? activity.hasEarlier : activity.hasMore;
  const offset = direction === 'before' ? activity.previousOffset : activity.nextOffset;
  function load() {
    const root = target.current?.closest<HTMLElement>('.agent-work__items, .graph-run-timeline');
    if (!root || busy || offset === undefined) return;
    if (activity.events.length >= 200 && (intent !== direction || root.scrollHeight <= root.clientHeight + 1 && !explicit.current)) return;
    const initiated = explicit.current; explicit.current = false; onLoad(offset, initiated);
  }
  return <div ref={target} style={{ flexShrink: 0 }}><AutoPage hasMore={!!more && (activity.events.length < 200 || intent === direction)} cursor={`${identity}/${activity.id}/${direction}/${offset}/${intent}/${gesture}`} busy={busy} error={error} onLoad={load} onRetry={onRetry}><div data-history-space={direction} aria-hidden="true" style={{ height: 0 }} /></AutoPage></div>;
}
