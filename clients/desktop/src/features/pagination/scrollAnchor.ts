export function captureScrollAnchor(root: HTMLElement, selector = '[data-history-anchor], [data-work-id]', relativeToRoot = false) {
  root.style.overflowAnchor = 'none';
  const viewport = root.getBoundingClientRect();
  const candidates = [...root.querySelectorAll<HTMLElement>(selector)];
  const visible = candidates.filter(item => {
    if (!item.getClientRects().length) return false;
    const rect = item.getBoundingClientRect();
    let top = viewport.top; let bottom = viewport.bottom;
    for (let parent = item.parentElement; parent && parent !== root; parent = parent.parentElement) {
      if (/(auto|scroll|hidden|clip)/.test(getComputedStyle(parent).overflowY)) {
        const clip = parent.getBoundingClientRect(); top = Math.max(top, clip.top); bottom = Math.min(bottom, clip.bottom);
      }
    }
    return rect.bottom > top && rect.top < bottom;
  });
  const preferred = visible.find(item => !item.querySelector('[data-work-id]')) ?? visible[0];
  const anchors = [preferred, ...visible.filter(item => item !== preferred)].filter((item): item is HTMLElement => !!item).map(element => ({
    element,
    attribute: element.hasAttribute('data-work-id') ? 'data-work-id' : 'data-history-anchor',
    id: element.getAttribute(element.hasAttribute('data-work-id') ? 'data-work-id' : 'data-history-anchor'),
    group: element.dataset.historyEvents?.split(' '),
    top: element.getBoundingClientRect().top - (relativeToRoot ? viewport.top : 0),
  }));
  const scrollTop = root.scrollTop;
  return () => {
    for (const anchor of anchors) {
      const target = anchor.element.isConnected ? anchor.element : anchor.id ? root.querySelector<HTMLElement>(`[${anchor.attribute}="${CSS.escape(anchor.id)}"]`) ?? anchor.group?.map(id => root.querySelector<HTMLElement>(`[data-history-events~="${CSS.escape(id)}"]`)).find(Boolean) : undefined;
      if (target) { root.scrollTop += target.getBoundingClientRect().top - (relativeToRoot ? root.getBoundingClientRect().top : 0) - anchor.top; return; }
    }
    root.scrollTop = scrollTop;
  };
}
