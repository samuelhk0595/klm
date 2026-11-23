import { act, cleanup, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { useRequestWindow } from './useRequestWindow';

const notices = vi.hoisted(() => new Set<() => void>());
vi.mock('./scopedConnection', () => ({ subscribeScope: (_kind: string, _id: string, listener: () => void) => { notices.add(listener); return () => notices.delete(listener); } }));
beforeEach(() => vi.useFakeTimers());
afterEach(() => { cleanup(); notices.clear(); vi.useRealTimers(); });
const items = (count: number) => Array.from({ length: count }, (_, index) => ({ id: String(index) }));

it('reconciles removed second-page requests even when the base page is unchanged', async () => {
  let inventory = items(60);
  const read = vi.fn(async (cursor: number | undefined) => {
    const offset = cursor ?? 0;
    return { items: inventory.slice(offset, offset + 25), next: offset + 25 < inventory.length ? offset + 25 : undefined };
  });
  const { result } = renderHook(() => useRequestWindow({ identity: 's/questions', sessionId: 's', revision: '1', base: { items: items(25), next: 25 }, key: item => item.id, read }));
  await act(async () => result.current.load());
  expect(result.current.items).toHaveLength(50);
  inventory = inventory.filter(item => item.id !== '30');
  await act(async () => { notices.forEach(notify => notify()); await vi.advanceTimersByTimeAsync(150); });
  expect(result.current.items.some(item => item.id === '30')).toBe(false);
  expect(result.current.items.some(item => item.id === '49')).toBe(true);
  expect(result.current.items).toHaveLength(50);
});

it('does not retry a failed extension on invalidations; explicit Retry resumes it', async () => {
  let rejectPage = true;
  const inventory = items(60);
  const read = vi.fn(async (cursor: number | undefined) => {
    if (cursor === 25 && rejectPage) throw new Error('Offline');
    const offset = cursor ?? 0;
    return { items: inventory.slice(offset, offset + 25), next: offset + 25 < inventory.length ? offset + 25 : undefined };
  });
  const { result } = renderHook(() => useRequestWindow({ identity: 's/queue', sessionId: 's', revision: '1', base: { items: items(25), next: 25 }, key: item => item.id, read }));
  await act(async () => result.current.load());
  expect(result.current.error).toBe('Offline');
  const requests = read.mock.calls.length;
  rejectPage = false;
  await act(async () => { notices.forEach(notify => notify()); await vi.advanceTimersByTimeAsync(30000); });
  expect(read).toHaveBeenCalledTimes(requests);
  await act(async () => result.current.retry());
  expect(result.current.items).toHaveLength(50);
  expect(result.current.error).toBe('');
});

it('rejects an in-flight page after the owning conversation changes', async () => {
  let finish!: (page: { items: { id: string }[]; next?: number }) => void;
  let signal!: AbortSignal;
  const read = vi.fn((_cursor: number | undefined, current: AbortSignal) => { signal = current; return new Promise<{ items: { id: string }[]; next?: number }>(resolve => { finish = resolve; }); });
  const { result, rerender } = renderHook(({ id }) => useRequestWindow({ identity: `${id}/queue`, sessionId: id, revision: '1', base: { items: [{ id }], next: 1 }, key: item => item.id, read }), { initialProps: { id: 'old' } });
  act(() => result.current.load());
  rerender({ id: 'new' });
  expect(signal.aborted).toBe(true);
  await act(async () => finish({ items: [{ id: 'stale' }] }));
  expect(result.current.items).toEqual([{ id: 'new' }]);
});

it('keeps live refreshes busy but silent while showing actual pagination progress', async () => {
  const inventory = items(60);
  let gate = Promise.resolve();
  let finish!: () => void;
  const read = vi.fn(async (cursor: number | undefined) => {
    await gate;
    const offset = cursor ?? 0;
    return { items: inventory.slice(offset, offset + 25), next: offset + 25 < inventory.length ? offset + 25 : undefined };
  });
  const { result } = renderHook(() => useRequestWindow({ identity: 's/questions', sessionId: 's', revision: '1', base: { items: items(25), next: 25 }, key: item => item.id, read }));
  await act(async () => result.current.load());
  expect(result.current.items).toHaveLength(50);

  gate = new Promise(resolve => { finish = resolve; });
  await act(async () => { notices.forEach(notify => notify()); await vi.advanceTimersByTimeAsync(150); });
  expect(result.current.busy).toBe(true);
  expect(result.current.loading).toBe(false);
  expect(result.current.items).toHaveLength(50);
  await act(async () => finish());
  expect(result.current.busy).toBe(false);

  gate = new Promise(resolve => { finish = resolve; });
  act(() => result.current.load());
  expect(result.current.busy).toBe(true);
  expect(result.current.loading).toBe(true);
  await act(async () => finish());
  expect(result.current.items).toHaveLength(60);
  expect(result.current.loading).toBe(false);
});
