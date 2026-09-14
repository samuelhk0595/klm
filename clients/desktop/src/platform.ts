import { invoke, isTauri } from '@tauri-apps/api/core';

export const IS_DESKTOP = isTauri();

type MobileBridge = { postMessage: (message: string) => void };
const mobileBridge = (window as Window & { KlmMobile?: MobileBridge }).KlmMobile;
export const IS_MOBILE_HOST = typeof mobileBridge?.postMessage === 'function';

export function returnToHosts() {
  mobileBridge?.postMessage('home');
}

export function resolveEngineURL() {
  const override = import.meta.env.VITE_ENGINE_URL?.trim();
  if (override) return override.replace(/\/+$/, '');
  const hostname = IS_DESKTOP ? 'localhost' : window.location.hostname;
  const port = import.meta.env.DEV || window.location.port === '17332' ? '17331' : '7331';
  return `http://${hostname || 'localhost'}:${port}`;
}

export async function openFocus() {
  await invoke('open_focus');
}

export async function startWindowDrag() {
  const { getCurrentWindow } = await import('@tauri-apps/api/window');
  await getCurrentWindow().startDragging();
}
