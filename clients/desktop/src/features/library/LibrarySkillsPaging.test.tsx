import { act, cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { request } from '../../engine';
import { LibrarySkillsPaging } from './LibrarySkillsPaging';
import type { LibraryInventory } from './api';

vi.mock('../../engine', () => ({ request: vi.fn() }));
vi.mock('../../engineIdentity', () => ({ currentEngineIdentity: () => 'engine', engineIdentityGeneration: () => 1 }));
vi.mock('../pagination/AutoPage', () => ({ AutoPage: ({ hasMore, busy, onLoad }: { hasMore: boolean; busy: boolean; onLoad: () => void }) => <button disabled={!hasMore || busy} onClick={onLoad}>Next</button> }));
afterEach(() => { cleanup(); vi.resetAllMocks(); });
const inventory = (search: string, clientGeneration = 1): LibraryInventory => ({ clientGeneration, search, skillsRoot: '/skills', skills: [], skillsHasMore: true, skillsNextOffset: 25, connections: [], recoveryPending: false, harnesses: {} });

it('discards a late skills page after switching search/inventory', async () => {
  let finish!: (value: LibraryInventory) => void;
  vi.mocked(request).mockReturnValue(new Promise(resolve => { finish = resolve; }));
  const onPage = vi.fn();
  const view = render(<LibrarySkillsPaging inventory={inventory('alpha')} query="alpha" onPage={onPage} />);
  fireEvent.click(screen.getByText('Next'));
  const signal = vi.mocked(request).mock.calls[0][4] as AbortSignal;
  view.rerender(<LibrarySkillsPaging inventory={inventory('beta', 2)} query="beta" onPage={onPage} />);
  expect(signal.aborted).toBe(true);
  await act(async () => finish(inventory('alpha')));
  expect(onPage).not.toHaveBeenCalled();
});

it('matches the server-normalized search and merges only the originating inventory', async () => {
  const original = inventory('alpha'); const replacement = inventory('alpha', 2);
  vi.mocked(request).mockResolvedValue({ ...original, skillsHasMore: false, skills: [{ id: 'more', name: 'More', description: '', source: '', valid: true }] });
  const onPage = vi.fn();
  render(<LibrarySkillsPaging inventory={original} query=" ALPHA " onPage={onPage} />);
  expect(screen.getByText('Next')).toBeEnabled();
  await act(async () => fireEvent.click(screen.getByText('Next')));
  const apply = onPage.mock.calls[0][0] as (previous: LibraryInventory) => LibraryInventory;
  expect(apply(replacement)).toBe(replacement);
  expect(apply(original).skills.map(item => item.id)).toEqual(['more']);
});
