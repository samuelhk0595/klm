import { act, cleanup, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import type { EventPage, Session, SessionUpdate } from '../../engine';
import { ConversationHistory } from './ConversationHistory';
import { readConversationWindow } from './scopedConnection';

const notices = vi.hoisted(() => new Set<() => void>());
vi.mock('./useHistoryVisibility', () => ({ useHistoryVisibility: () => true, isHistoryVisible: () => true }));

vi.mock('./PendingRequestPages', () => ({ PendingRequestPages: () => null }));
vi.mock('./ConversationEvents', () => ({ ConversationEvents: ({ session }: { session: Session }) => <div className="conversation-events"><span>{session.history?.startIndex}/{session.events.length}</span></div> }));
vi.mock('./HistoryLoader', () => ({ HistoryLoader: ({ session, direction = 'before', onPage }: { session: Session; direction?: string; onPage: (page: EventPage) => void }) => <button onClick={() => {
  const start = session.history!.startIndex;
  const first = direction === 'before' ? start - 25 : start + session.events.length;
  onPage({ sessionId: session.id, revision: 1, startIndex: first, endIndex: first + 25, total: 500, events: events(first, 25), hasMore: first > 0, direction: direction as 'before' | 'after' });
}}>{direction}</button> }));
vi.mock('./scopedConnection', () => ({
  subscribeScope: (_kind: string, _id: string, listener: () => void) => { notices.add(listener); return () => notices.delete(listener); },
  readConversationWindow: vi.fn(async (_id: string, current: Session) => ({ kind: 'reset', windowRefresh: true, summary: current, revision: 1, total: 500, changes: [], history: { ...current.history, sessionId: current.id, revision: 1, endIndex: current.history!.startIndex + current.events.length, events: current.events, total: 500 } })),
}));
const events = (start: number, count: number) => Array.from({ length: count }, (_, index) => ({ id: String(start + index), type: 'assistant' as const, text: '', createdAt: '' }));
const session: Session = { id: 's', projectId: 'p', title: 'Session', workspace: 'Ungrouped', harness: 'pi', status: 'idle', createdAt: '', updatedAt: '', events: events(300, 200), history: { revision: 1, startIndex: 300, total: 500, nextCursor: '300', afterCursor: '499', hasMore: true } };
beforeEach(() => { Object.defineProperty(HTMLElement.prototype, 'scrollTo', { configurable: true, value: vi.fn() }); });
afterEach(() => { cleanup(); notices.clear(); vi.restoreAllMocks(); Reflect.deleteProperty(HTMLElement.prototype, 'scrollTo'); });

it('lets one reader traverse beyond 200 without moving or asking permission from the other', async () => {
  const externalPage = vi.fn();
  render(<><ConversationHistory session={session} working={false} label="First" onSnapshot={vi.fn()} onHistory={externalPage} /><ConversationHistory session={session} working={false} label="Second" onSnapshot={vi.fn()} onHistory={externalPage} /></>);
  const first = screen.getByRole('log', { name: 'First' }); const second = screen.getByRole('log', { name: 'Second' });
  await act(async () => fireEvent.click(within(first).getByText('before')));
  await act(async () => fireEvent.click(within(first).getByText('before')));
  expect(within(first).getByText('250/200')).toBeInTheDocument();
  expect(within(second).getByText('300/200')).toBeInTheDocument();
  expect(externalPage).not.toHaveBeenCalled();
  await act(async () => fireEvent.click(within(first).getByText('after')));
  expect(within(first).getByText('275/200')).toBeInTheDocument();
  expect(within(second).getByText('300/200')).toBeInTheDocument();
});

it('rejects a refresh of an old window when a newer page was installed in flight', async () => {
  let finish!: (result: SessionUpdate) => void;
  let requested!: Session;
  vi.mocked(readConversationWindow).mockImplementationOnce(async (_id, current) => {
    requested = current!;
    return new Promise<SessionUpdate>(resolve => { finish = resolve; });
  });
  render(<ConversationHistory session={session} working={false} label="Child" onSnapshot={vi.fn()} onHistory={vi.fn()} />);
  const view = screen.getByRole('log', { name: 'Child' });
  await act(async () => fireEvent.click(within(view).getByText('before')));
  await act(async () => fireEvent.click(within(view).getByText('before')));
  await act(async () => finish({ kind: 'reset', windowRefresh: true, summary: requested, revision: 1, total: 500, changes: [], history: { ...requested.history!, sessionId: requested.id, endIndex: requested.history!.startIndex + requested.events.length, events: requested.events } }));
  expect(within(view).getByText('250/200')).toBeInTheDocument();
});

it('keeps following the real tail when automatic older prefetch only fills the window', async () => {
  const tail = { ...session, events: events(475, 25), history: { ...session.history!, startIndex: 475, nextCursor: '475' } };
  render(<ConversationHistory session={tail} working={false} label="Compact" onSnapshot={vi.fn()} onHistory={vi.fn()} />);
  await act(async () => fireEvent.click(within(screen.getByRole('log')).getByText('before')));
  await act(async () => { notices.forEach(notify => notify()); await new Promise(resolve => setTimeout(resolve, 200)); });
  expect(vi.mocked(readConversationWindow).mock.calls.at(-1)?.[4]).toBe(true);
});

it('retains prefetched history when a live refresh returns only the newest 25 events', async () => {
  const tail = { ...session, events: events(475, 25), history: { ...session.history!, startIndex: 475, nextCursor: '475' } };
  render(<ConversationHistory session={tail} working label="Live" onSnapshot={vi.fn()} onHistory={vi.fn()} />);
  await act(async () => {});
  const view = screen.getByRole('log', { name: 'Live' });
  for (let index = 0; index < 7; index++) await act(async () => fireEvent.click(within(view).getByText('before')));
  expect(within(view).getByText('300/200')).toBeInTheDocument();
  vi.mocked(readConversationWindow).mockImplementationOnce(async (_id, current) => ({
    kind: 'reset', windowRefresh: true, summary: current!, revision: 2, total: 501, changes: [],
    history: { ...current!.history!, sessionId: current!.id, revision: 2, startIndex: 476, endIndex: 501, total: 501, events: events(476, 25), hasMore: true },
  }));
  await act(async () => { notices.forEach(notify => notify()); await new Promise(resolve => setTimeout(resolve, 200)); });
  expect(within(view).getByText('301/200')).toBeInTheDocument();
});
