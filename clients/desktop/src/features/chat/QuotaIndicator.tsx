import { useEffect, useState } from 'react';
import { request, type QuotaSnapshot, type Session } from '../../engine';

function remaining(seconds: number) {
  if (seconds < 60) return '<1 min';
  const [value, unit] = seconds >= 86400 ? [Math.ceil(seconds / 86400), 'day'] : seconds >= 3600 ? [Math.ceil(seconds / 3600), 'hour'] : [Math.ceil(seconds / 60), 'min'];
  return `${value} ${unit}${value === 1 || unit === 'min' ? '' : 's'}`;
}

export function QuotaIndicator({ session }: { session: Session }) {
  const [quota, setQuota] = useState<QuotaSnapshot | null>(null);
  const [now, setNow] = useState(Date.now());
  useEffect(() => {
    let active = true;
    let pending = false;
    setQuota(null);
    async function refresh() {
      if (pending || document.hidden) return;
      pending = true;
      try {
        const result = await request<{ quota: QuotaSnapshot | null }>(`/api/sessions/${encodeURIComponent(session.id)}/quota`, 'GET', undefined, 55000);
        if (active) setQuota(result.quota);
      } catch { if (active) setQuota(value => value ? { ...value, stale: true } : null); }
      finally { pending = false; }
    }
    void refresh();
    const timer = window.setInterval(() => void refresh(), 120000);
    document.addEventListener('visibilitychange', refresh);
    return () => { active = false; window.clearInterval(timer); document.removeEventListener('visibilitychange', refresh); };
  }, [session.id, session.model, session.resolvedModel]);
  useEffect(() => { const timer = window.setInterval(() => setNow(Date.now()), 30000); return () => window.clearInterval(timer); }, []);

  if (!quota || now - Date.parse(quota.observedAt) > 5 * 60000) return null;
  const windows = quota.windows.filter(window => window.resetsAt * 1000 > now)
    .map(window => ({ ...window, remainingPercent: Math.max(0, Math.min(100, 100 - window.usedPercent)) }));
  const lowest = windows.reduce<(typeof windows)[number] | undefined>((chosen, window) => !chosen || window.remainingPercent < chosen.remainingPercent ? window : chosen, undefined);
  if (!lowest) return null;
  const title = [quota.source, ...windows.map(window => `${window.name}: ${Math.round(window.remainingPercent)}% remaining · Resets ${new Date(window.resetsAt * 1000).toLocaleString('en-US')}`), `Updated ${new Date(quota.observedAt).toLocaleTimeString('en-US')}${quota.stale ? ' · Refresh unavailable' : ''}`].join('\n');
  return <span title={title} aria-label={title}>{Math.round(lowest.remainingPercent)}% remaining · {remaining(lowest.resetsAt - now / 1000)}</span>;
}
