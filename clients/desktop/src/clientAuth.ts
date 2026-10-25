import { invoke } from '@tauri-apps/api/core';
import { IS_DESKTOP, resolveEngineURL } from './platform';

const engine = resolveEngineURL();
let nativeCredential: string | null = null;
let blocked = false;
let generation = 0;
let initialized: Promise<void> | undefined;
let nativeDeletion: Promise<unknown> = Promise.resolve();
export const AUTH_REQUIRED_EVENT = 'klm-auth-required';

export function initializeClientAuth() {
  return initialized ??= (async () => {
    if (IS_DESKTOP) nativeCredential = await invoke<string | null>('client_credential_get', { engine });
  })().catch(error => { initialized = undefined; throw error; });
}

export function clientAuthGeneration() { return generation; }
export function isClientAuthCurrent(version: number) { return version === generation && !blocked; }

export function retryClientAuth() { generation++; blocked = false; }

export function invalidateClientAuth() {
  if (blocked) return;
  generation++;
  blocked = true;
  nativeCredential = null;
  if (IS_DESKTOP) nativeDeletion = invoke('client_credential_delete', { engine }).catch(() => {});
  window.dispatchEvent(new Event(AUTH_REQUIRED_EVENT));
}

export async function clientFetch(path: string, init: RequestInit = {}): Promise<Response> {
  await initializeClientAuth();
  if (blocked) return new Response(JSON.stringify({ error: 'Pair this client to access the engine.' }), { status: 401 });
  const version = generation;
  const headers = new Headers(init.headers);
  headers.set('X-KLM-CSRF', '1');
  if (IS_DESKTOP && nativeCredential) headers.set('Authorization', `Bearer ${nativeCredential}`);
  const response = await fetch(`${engine}${path}`, { ...init, headers, credentials: IS_DESKTOP ? 'omit' : 'include' });
  // Responses from a previous credential must neither invalidate its replacement
  // nor supply protected data after a newer denial/retry/pairing transition.
  if (!isClientAuthCurrent(version)) throw new DOMException('Client authentication changed.', 'AbortError');
  if (response.status === 401) invalidateClientAuth();
  return response;
}

export async function pairClient(base: string, name: string, code: string) {
  generation++;
  blocked = true;
  const response = await fetch(`${base}/api/auth/pair`, {
    method: 'POST', credentials: IS_DESKTOP ? 'omit' : 'include',
    headers: { 'Content-Type': 'application/json', 'X-KLM-CSRF': '1', ...(IS_DESKTOP ? { 'X-KLM-Native': '1' } : {}) },
    body: JSON.stringify({ name, code }), signal: AbortSignal.timeout(15000),
  });
  const result = await response.json();
  if (!response.ok) throw new Error(result.error || 'Could not pair this client.');
  if (IS_DESKTOP) {
    if (typeof result.credential !== 'string') throw new Error('The engine did not return a native credential.');
    await nativeDeletion;
    await invoke('client_credential_set', { engine: base, credential: result.credential });
    if (base === engine) nativeCredential = result.credential;
  }
  generation++;
  blocked = false;
}

// Fetch streaming supports native Authorization headers, unlike EventSource.
// Keep the small EventSource interface used by the existing reconnect helper.
export function createClientEventSource(url: string): EventSource {
  const controller = new AbortController();
  const source = {
    readyState: 0, onopen: null, onmessage: null, onerror: null,
    close() { setState(2); controller.abort(); },
  } as unknown as EventSource;
  const setState = (state: number) => Object.defineProperty(source, 'readyState', { value: state, configurable: true });
  void (async () => {
    try {
      if (!url.startsWith(`${engine}/`)) throw new Error('Engine stream destination changed.');
      const response = await clientFetch(url.slice(engine.length), { signal: controller.signal });
      if (!response.ok || !response.body || !response.headers.get('content-type')?.startsWith('text/event-stream')) throw new Error('Engine stream unavailable.');
      setState(1);
      source.onopen?.call(source, new Event('open'));
      const reader = response.body.pipeThrough(new TextDecoderStream()).getReader();
      let buffer = '';
      let data: string[] = [];
      while (!controller.signal.aborted) {
        const chunk = await reader.read();
        if (chunk.done) break;
        buffer += chunk.value;
        let newline: number;
        while ((newline = buffer.indexOf('\n')) >= 0) {
          const line = buffer.slice(0, newline).replace(/\r$/, '');
          buffer = buffer.slice(newline + 1);
          if (line === '') {
            if (data.length) source.onmessage?.call(source, new MessageEvent('message', { data: data.join('\n') }));
            data = [];
          } else if (line.startsWith('data:')) data.push(line.slice(5).replace(/^ /, ''));
        }
      }
      if (!controller.signal.aborted) throw new Error('Engine stream closed.');
    } catch {
      if (!controller.signal.aborted) {
        setState(2);
        source.onerror?.call(source, new Event('error'));
      }
    }
  })();
  return source;
}
