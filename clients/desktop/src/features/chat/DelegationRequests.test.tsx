import { act, cleanup, render, screen } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { request, type Session } from '../../engine';
import { DelegationRequests } from './DelegationRequests';

const notices = vi.hoisted(() => new Set<() => void>());
vi.mock('../../engine', () => ({ request: vi.fn() }));
vi.mock('./scopedConnection', () => ({ subscribeScope: (_kind: string, _id: string, listener: () => void) => { notices.add(listener); return () => notices.delete(listener); } }));
vi.mock('./PermissionCard', () => ({ PermissionCard: () => null }));
vi.mock('./QuestionCard', () => ({ QuestionCard: () => null }));

afterEach(() => { cleanup(); notices.clear(); vi.resetAllMocks(); vi.useRealTimers(); });

it('does not flash an empty child-request loading row on streaming invalidations', async () => {
  vi.useFakeTimers();
  const empty = { items: [], hasMore: false, nextAfterId: '' };
  vi.mocked(request).mockResolvedValueOnce(empty);
  const session = { id: 's', projectId: 'p', status: 'running' } as Session;
  render(<DelegationRequests session={session} projectName="Project" onSnapshot={vi.fn()} />);
  await act(async () => {});
  expect(screen.queryByRole('status')).not.toBeInTheDocument();

  let finish!: (page: typeof empty) => void;
  vi.mocked(request).mockImplementationOnce(() => new Promise(resolve => { finish = resolve; }));
  await act(async () => { notices.forEach(notify => notify()); await vi.advanceTimersByTimeAsync(150); });
  expect(request).toHaveBeenCalledTimes(2);
  expect(screen.queryByRole('status')).not.toBeInTheDocument();
  await act(async () => finish(empty));
  expect(screen.queryByRole('status')).not.toBeInTheDocument();
});
