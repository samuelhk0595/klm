import { useLayoutEffect, useState, type RefObject } from 'react';

export function isHistoryVisible(element: HTMLElement | null): boolean {
  if (document.hidden || !element?.isConnected || element.closest('[hidden]') || !element.getClientRects().length) return false;
  const bounds = element.getBoundingClientRect();
  return bounds.width > 0 && bounds.height > 0 && getComputedStyle(element).visibility === 'visible';
}

// Focus and subagent navigation can hide a mounted reader through an ancestor,
// independently of the document's visibility or the session's metadata.
export function useHistoryVisibility(root: RefObject<HTMLElement | null>, identity: string) {
  const [visible, setVisible] = useState(false);
  useLayoutEffect(() => {
    const element = root.current;
    if (!element) return;
    const measure = () => { setVisible(isHistoryVisible(element)); };
    const mutations = new MutationObserver(measure);
    const resize = typeof ResizeObserver === 'undefined' ? undefined : new ResizeObserver(measure);
    for (let parent: HTMLElement | null = element; parent; parent = parent.parentElement) {
      mutations.observe(parent, { attributes: true, attributeFilter: ['hidden', 'class', 'style', 'open'] });
      resize?.observe(parent);
    }
    document.addEventListener('visibilitychange', measure);
    measure();
    return () => { mutations.disconnect(); resize?.disconnect(); document.removeEventListener('visibilitychange', measure); };
  }, [root, identity]);
  return visible;
}
