import { afterEach, expect, it, vi } from 'vitest';

const base = 'https://engine.example.test';
const invoke = vi.fn();

async function transport(native = false, engine = base) {
  vi.resetModules();
  vi.doMock('./platform', () => ({ IS_DESKTOP: native, resolveEngineURL: () => engine }));
  vi.doMock('@tauri-apps/api/core', () => ({ invoke }));
  return import('./clientAuth');
}

afterEach(() => { vi.unstubAllGlobals(); vi.clearAllMocks(); });

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>(done => { resolve = done; });
  return { promise, resolve };
}

it('sends cookie credentials and multipart CSRF headers, then blocks requests after a 401', async () => {
  const fetch = vi.fn()
    .mockResolvedValueOnce(new Response('{}', { status: 200 }))
    .mockResolvedValue(new Response('{}', { status: 401 }));
  vi.stubGlobal('fetch', fetch);
  const auth = await transport();
  const form = new FormData(); form.append('file', new Blob(['audio']), 'audio.m4a');
  await auth.clientFetch('/api/transcriptions', { method: 'POST', body: form });
  const [, init] = fetch.mock.calls[0];
  expect(init.credentials).toBe('include');
  expect(init.headers.get('X-KLM-CSRF')).toBe('1');
  expect(init.headers.has('Content-Type')).toBe(false);
  const denied = vi.fn(); window.addEventListener(auth.AUTH_REQUIRED_EVENT, denied);
  await auth.clientFetch('/api/auth/session');
  await auth.clientFetch('/api/updates');
  expect(fetch).toHaveBeenCalledTimes(2);
  expect(denied).toHaveBeenCalledTimes(1);
  window.removeEventListener(auth.AUTH_REQUIRED_EVENT, denied);
});

it('loads native credentials from OS storage and sends them only as bearer headers', async () => {
  const credential = 'a'.repeat(43);
  invoke.mockResolvedValue(credential);
  const fetch = vi.fn().mockResolvedValue(new Response('{}'));
  vi.stubGlobal('fetch', fetch);
  const auth = await transport(true);
  await auth.clientFetch('/api/auth/session');
  expect(invoke).toHaveBeenCalledWith('client_credential_get', { engine: base });
  const [url, init] = fetch.mock.calls[0];
  expect(url).toBe(`${base}/api/auth/session`);
  expect(init.credentials).toBe('omit');
  expect(init.headers.get('Authorization')).toBe(`Bearer ${credential}`);
  expect(localStorage.getItem('credential')).toBeNull();
});

it('parses fragmented SSE and aborts the authenticated transport when closed', async () => {
  let streamController: ReadableStreamDefaultController<Uint8Array>;
  const body = new ReadableStream<Uint8Array>({ start(controller) { streamController = controller; } });
  const fetch = vi.fn().mockResolvedValue(new Response(body, { headers: { 'Content-Type': 'text/event-stream' } }));
  vi.stubGlobal('fetch', fetch);
  const auth = await transport();
  const source = auth.createClientEventSource(`${base}/api/updates?ids=x`);
  const message = vi.fn(); source.onmessage = message;
  await vi.waitFor(() => expect(source.readyState).toBe(1));
  const encoder = new TextEncoder();
  streamController!.enqueue(encoder.encode(': keepalive\n\ndata: {"kind":'));
  streamController!.enqueue(encoder.encode('"inventory"}\r\n\r\n'));
  await vi.waitFor(() => expect(message).toHaveBeenCalledTimes(1));
  expect(message.mock.calls[0][0].data).toBe('{"kind":"inventory"}');
  source.close();
  expect(fetch.mock.calls[0][1].signal.aborted).toBe(true);
  streamController!.close();
});

it('ignores an old 401 arriving after pairing without deleting the replacement native credential', async () => {
  const first = deferred<Response>();
  const late = deferred<Response>();
  const oldCredential = 'a'.repeat(43);
  const newCredential = 'b'.repeat(43);
  invoke.mockImplementation(async command => command === 'client_credential_get' ? oldCredential : undefined);
  const fetch = vi.fn()
    .mockReturnValueOnce(first.promise)
    .mockReturnValueOnce(late.promise)
    .mockResolvedValueOnce(new Response(JSON.stringify({ credential: newCredential })))
    .mockResolvedValue(new Response('{}'));
  vi.stubGlobal('fetch', fetch);
  const auth = await transport(true);
  const firstRequest = auth.clientFetch('/api/auth/session');
  const lateResult = auth.clientFetch('/api/state').catch(error => error);
  await vi.waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));
  first.resolve(new Response('{}', { status: 401 }));
  expect((await firstRequest).status).toBe(401);
  await auth.pairClient(base, 'Replacement', 'test-code');
  late.resolve(new Response('{}', { status: 401 }));
  expect(await lateResult).toMatchObject({ name: 'AbortError' });
  expect(invoke.mock.calls.filter(([command]) => command === 'client_credential_delete')).toHaveLength(1);
  expect(invoke).toHaveBeenCalledWith('client_credential_set', { engine: base, credential: newCredential });
  await auth.clientFetch('/api/auth/session');
  expect(fetch.mock.calls[3][1].headers.get('Authorization')).toBe(`Bearer ${newCredential}`);
});

it('uses engine B existing native storage after reload without pairing or forwarding engine A credential', async () => {
  const engineB = 'https://paired.example.test';
  const credentialA = 'a'.repeat(43);
  const credentialB = 'b'.repeat(43);
  invoke.mockImplementation(async (command, args) => command === 'client_credential_get' ? (args.engine === base ? credentialA : credentialB) : undefined);
  const fetch = vi.fn().mockRejectedValueOnce(new Error('Engine A unavailable')).mockResolvedValue(new Response('{}'));
  vi.stubGlobal('fetch', fetch);
  const authA = await transport(true);
  await expect(authA.clientFetch('/api/auth/session')).rejects.toThrow('Engine A unavailable');
  const authB = await transport(true, engineB);
  await authB.clientFetch('/api/auth/session');
  expect(invoke).toHaveBeenCalledWith('client_credential_get', { engine: engineB });
  expect(fetch.mock.calls[1][0]).toBe(`${engineB}/api/auth/session`);
  expect(fetch.mock.calls[1][1].headers.get('Authorization')).toBe(`Bearer ${credentialB}`);
  expect(fetch.mock.calls.some(([url]) => url.includes('/api/auth/pair'))).toBe(false);
});
