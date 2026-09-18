export type SortPositions = Map<string, number>;

export function captureSortPositions(container: HTMLElement, selector: string, attribute: string) {
  const positions: SortPositions = new Map();
  for (const element of container.querySelectorAll<HTMLElement>(selector)) {
    const key = element.dataset[attribute];
    if (key) positions.set(key, element.getBoundingClientRect().top);
  }
  return positions;
}

export function animateSortPositions(container: HTMLElement, selector: string, attribute: string, previous: SortPositions) {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
  for (const element of container.querySelectorAll<HTMLElement>(selector)) {
    const key = element.dataset[attribute];
    const before = key ? previous.get(key) : undefined;
    if (before === undefined) continue;
    for (const animation of element.getAnimations()) animation.cancel();
    const delta = before - element.getBoundingClientRect().top;
    if (Math.abs(delta) < 1) continue;
    element.animate([{ transform: `translateY(${delta}px)` }, { transform: 'translateY(0)' }], {
      duration: 160,
      easing: 'cubic-bezier(.2, .8, .2, 1)',
    });
  }
}
