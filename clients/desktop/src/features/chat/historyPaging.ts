import type { EngineEvent } from '../../engine';

export const HISTORY_WINDOW_MESSAGES = 200;
// Separate transport/memory protection, not part of the message-page quota.
export const HISTORY_WINDOW_EVENTS = 8192;
export const isHistoryMessage = (event: EngineEvent) => event.type === 'user' || event.type === 'assistant' || event.type === 'agent_prompt';

export function retainHistoryWindow(events: EngineEvent[], unit: 'messages' | undefined, direction: 'before' | 'after') {
  if (unit !== 'messages') {
    const start = direction === 'after' ? Math.max(0, events.length - 200) : 0;
    return { start, events: events.slice(start, start + 200) };
  }
  let start = 0, stop = events.length, messages = 0;
  if (direction === 'after') {
    start = Math.max(0, events.length - HISTORY_WINDOW_EVENTS);
    for (let index = events.length - 1; index >= start; index--) {
      if (isHistoryMessage(events[index]) && ++messages > HISTORY_WINDOW_MESSAGES) { start = index + 1; break; }
    }
  } else {
    stop = Math.min(stop, HISTORY_WINDOW_EVENTS);
    for (let index = 0; index < stop; index++) {
      if (isHistoryMessage(events[index]) && ++messages > HISTORY_WINDOW_MESSAGES) { stop = index; break; }
    }
  }
  return { start, events: events.slice(start, stop) };
}
