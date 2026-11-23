import { useEffect, useRef, useState, type RefObject } from 'react';
import { EDGE_MARGIN, scrollRoot } from './pagination';

export function useNearEdge(target: RefObject<HTMLElement | null>, { enabled, token, busy, error, load, root }: {
  enabled: boolean; token: string; busy?: boolean; error?: string; load: () => void; root?: RefObject<HTMLElement | null>;
}) {
  const latest = useRef(load); latest.current = load;
  const attempted = useRef('');
  const [layout, setLayout] = useState(0);
  useEffect(() => {
    const element = target.current;
    if (!element || !enabled || busy || error) return;
    const container = root?.current ?? scrollRoot(element);
    const ancestors: HTMLElement[] = [];
    for (let parent = element.parentElement; parent; parent = parent.parentElement) ancestors.push(parent);
    let frame = 0;
    const check = () => {
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(() => {
        if (!root && scrollRoot(element) !== container) { setLayout(value => value + 1); return; }
        if (document.hidden || !element.getClientRects().length || getComputedStyle(element).visibility !== 'visible') { attempted.current = ''; return; }
        if (attempted.current === token) return;
        const bounds = element.getBoundingClientRect();
        const viewport = { top: 0, bottom: innerHeight, left: 0, right: innerWidth };
        // A nested scroller can itself be outside a surrounding panel. Its
        // sentinel must not fetch simply because it is near that hidden root.
        for (let parent: HTMLElement | null = container; parent && parent !== document.documentElement; parent = parent.parentElement) {
          const style = getComputedStyle(parent);
          const rect = parent.getBoundingClientRect();
          if (parent === container || /(auto|scroll|hidden|clip)/.test(style.overflowY)) {
            viewport.top = Math.max(viewport.top, rect.top); viewport.bottom = Math.min(viewport.bottom, rect.bottom);
          }
          if (parent === container || /(auto|scroll|hidden|clip)/.test(style.overflowX)) {
            viewport.left = Math.max(viewport.left, rect.left); viewport.right = Math.min(viewport.right, rect.right);
          }
        }
        if (viewport.bottom <= viewport.top || viewport.right <= viewport.left) return;
        if (bounds.right >= viewport.left && bounds.left <= viewport.right && bounds.top <= viewport.bottom + EDGE_MARGIN && bounds.bottom >= viewport.top - EDGE_MARGIN) {
          attempted.current = token;
          latest.current();
        }
      });
    };
    const observer = typeof IntersectionObserver === 'undefined' ? undefined : new IntersectionObserver(check, { root: container === document.documentElement ? null : container, rootMargin: `${EDGE_MARGIN}px` });
    observer?.observe(element);
    const resize = typeof ResizeObserver === 'undefined' ? undefined : new ResizeObserver(check);
    resize?.observe(element); ancestors.forEach(parent => resize?.observe(parent));
    const mutations = typeof MutationObserver === 'undefined' ? undefined : new MutationObserver(check);
    ancestors.forEach(parent => mutations?.observe(parent, { attributes: true, attributeFilter: ['class', 'style', 'hidden', 'open'] }));
    ancestors.forEach(parent => parent.addEventListener('scroll', check, { passive: true }));
    window.addEventListener('scroll', check, { passive: true });
    document.addEventListener('visibilitychange', check); window.addEventListener('resize', check);
    check();
    return () => {
      cancelAnimationFrame(frame); observer?.disconnect(); resize?.disconnect(); mutations?.disconnect();
      ancestors.forEach(parent => parent.removeEventListener('scroll', check)); window.removeEventListener('scroll', check);
      document.removeEventListener('visibilitychange', check); window.removeEventListener('resize', check);
    };
  }, [target, root, enabled, token, busy, error, layout]);
}
