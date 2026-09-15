import { mergeGraphState, type EventPage, type Session, type SessionSummary, type SessionUpdate } from '../../engine';

function version(timestamp: string) {
  const [seconds, fraction = ''] = timestamp.replace(/Z$/, '').split('.');
  return `${seconds}.${fraction.padEnd(9, '0')}`;
}

export function mergeSessions(current: Session[], incoming: (SessionSummary | Session | SessionUpdate)[], live = false): Session[] {
  const sessions = new Map(current.map(session => [session.id, session]));
  for (const item of incoming) {
    const update = 'summary' in item ? item : undefined;
    const summary = update ? update.summary : item as SessionSummary | Session;
    if (summary.role === 'graph_node') continue;
    const previous = sessions.get(summary.id);
    const stale = update && previous?.history && update.revision < previous.history.revision;
    const base = previous && (stale || version(summary.updatedAt) < version(previous.updatedAt)) ? previous : summary;
    const graph = mergeGraphState(previous?.graph, summary.graph);
    let events = previous?.events ?? [];
    let history = previous?.history;
    if (!update && 'events' in item && Array.isArray(item.events) && base === summary) {
      events = item.events;
      history = { revision: 0, startIndex: 0, total: events.length, hasMore: false };
    }
    // Mutation responses contain only that commit's changes. They must not
    // advance the SSE cursor or they could skip an intervening streamed event.
    if (live && update?.history && (!history || update.revision >= history.revision)) {
      const page = update.history;
      events = page.events ?? [];
      history = { revision: update.revision, startIndex: page.startIndex, total: page.total, nextCursor: page.nextCursor, hasMore: page.hasMore };
    } else if (live && update && history && update.revision >= history.revision) {
      const nextEvents = events.slice();
      // The SSE reset supplies the initial window; deltas replace or append by
      // absolute index, including edits to an event already being streamed.
      for (const change of update.changes) {
        if (change.index < history.startIndex) continue;
        const index = change.index - history.startIndex;
        if (index <= nextEvents.length) nextEvents[index] = change.event;
      }
      events = nextEvents;
      history = { ...history, revision: update.revision, total: update.total };
    }
    sessions.set(summary.id, {
      ...base, events, history,
      runtimeActive: stale ? previous?.runtimeActive : summary.runtimeActive,
      ...(graph ? { graph, selectedGraphId: graph.selectedGraphId } : {}),
    });
  }
  return [...sessions.values()];
}

export function prependHistory(sessions: Session[], page: EventPage): Session[] {
  return sessions.map(session => {
    const history = session.history;
    if (session.id !== page.sessionId || !history || page.endIndex !== history.startIndex) return session;
    return {
      ...session,
      events: [...(page.events ?? []), ...session.events],
      history: { ...history, startIndex: page.startIndex, nextCursor: page.nextCursor, hasMore: page.hasMore },
    };
  });
}

export function isSessionResponse(value: unknown): value is Session | SessionUpdate {
  if (!value || typeof value !== 'object') return false;
  if ('summary' in value) {
    const update = value as SessionUpdate;
    return (update.kind === 'reset' || update.kind === 'delta') && typeof update.summary?.id === 'string'
      && typeof update.summary.updatedAt === 'string' && Number.isSafeInteger(update.revision)
      && Array.isArray(update.changes) && (update.kind !== 'reset' || !!update.history);
  }
  const session = value as Session;
  return typeof session.id === 'string' && typeof session.updatedAt === 'string' && Array.isArray(session.events);
}
