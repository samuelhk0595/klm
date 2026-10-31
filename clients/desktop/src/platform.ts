import { invoke, isTauri } from '@tauri-apps/api/core';

export const IS_DESKTOP = isTauri();

type MobileBridge = { postMessage: (message: string) => void };
const mobileBridge = (window as Window & { KlmMobile?: MobileBridge }).KlmMobile;
export const IS_MOBILE_HOST = typeof mobileBridge?.postMessage === 'function';

const engineURLStorageKey = 'klm.engine-url.v1';

export function returnToHosts() {
  mobileBridge?.postMessage('home');
}

export async function openContentDownload(url: string) {
  if (IS_DESKTOP) {
    await invoke('open_content_download', { url });
    return;
  }
  const downloadBridge = (window as Window & { KlmDownload?: MobileBridge }).KlmDownload;
  if (IS_MOBILE_HOST) {
    if (!downloadBridge) throw new Error('Update the mobile client to download content.');
    downloadBridge.postMessage(url);
    return;
  }
  const link = document.createElement('a');
  link.href = url;
  link.download = 'event.json';
  link.rel = 'noreferrer';
  link.click();
}

export function resolveEngineURL() {
  const override = import.meta.env.VITE_ENGINE_URL?.trim();
  if (override) return normalizeEngineURL(override);
  try {
    const configured = localStorage.getItem(engineURLStorageKey)?.trim();
    if (configured) return normalizeEngineURL(configured);
  } catch { /* Invalid legacy selections or unavailable storage fall back to the URL-derived engine. */ }
  const hostname = IS_DESKTOP ? 'localhost' : window.location.hostname;
  if (!IS_DESKTOP && hostname === 'klm.codemob.com.br') return 'https://klmengine.codemob.com.br';
  const port = import.meta.env.DEV || window.location.port === '17332' ? '17331' : '7331';
  return `http://${hostname || 'localhost'}:${port}`;
}

export function configureEngineURL(value: string) {
  const configured = selectableEngineURL(value);
  localStorage.setItem(engineURLStorageKey, configured);
  return configured;
}

export function selectableEngineURL(value: string) {
  const configured = normalizeEngineURL(value);
  const override = import.meta.env.VITE_ENGINE_URL?.trim();
  if (override && configured !== normalizeEngineURL(override)) {
    throw new Error('Engine URL is fixed by VITE_ENGINE_URL. Update the development override to switch engines.');
  }
  return configured;
}

export function reloadClient() { window.location.reload(); }

export function normalizeEngineURL(value: string) {
  let address = value.trim();
  const hasScheme = /^[a-z][a-z\d+.-]*:\/\//i.test(address);
  if (!hasScheme) {
    // A bare IPv6 literal needs brackets before URL parsing.
    if (/^[\da-f:]+$/i.test(address) && address.includes("::")) address = "[" + address + "]";
    const candidate = new URL("http://" + address);
    const ip = candidate.hostname.startsWith("[") || /^\d+\.\d+\.\d+\.\d+$/.test(candidate.hostname);
    address = (ip ? "http://" : "https://") + address;
  }
  const url = new URL(address);
  if (url.username || url.password) throw new Error("The engine URL cannot include credentials.");
  const ip = url.hostname.startsWith("[") || /^\d+\.\d+\.\d+\.\d+$/.test(url.hostname);
  // URL.port omits explicit :80/:443, so inspect the original authority.
  const authority = address.slice(address.indexOf("://") + 3).split(/[/?#]/, 1)[0];
  // Bare IP input uses the engine port; absolute URLs retain standard default
  // ports so re-normalizing a canonical URL is idempotent (including :80/:443).
  if (ip && !hasScheme && !/:\d+$/.test(authority)) url.port = "7331";
  if (url.protocol !== 'http:' && url.protocol !== 'https:') throw new Error('Use an HTTP or HTTPS URL.');
  const loopback = url.hostname === 'localhost' || url.hostname === '127.0.0.1' || url.hostname === '[::1]';
  if (url.protocol !== 'https:' && !loopback) throw new Error('Use HTTPS for remote engines.');
  if (url.search || url.hash) throw new Error('The engine URL cannot include a query or fragment.');
  const configured = url.toString().replace(/\/+$/, '');
  return configured;
}

export async function openFocus() {
  await invoke('open_focus');
}

export async function startWindowDrag() {
  const { getCurrentWindow } = await import('@tauri-apps/api/window');
  await getCurrentWindow().startDragging();
}
