import { useRef, useState } from 'react';
import { Button } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { ENGINE_URL, rotateWebhookSecret, saveWebhookBinding, type WebhookBinding, type TaskDefinition } from '../../engine';
import { formatRunDate } from './types';

export const taskOperationId = () => Array.from(crypto.getRandomValues(new Uint32Array(4)), value => value.toString(16).padStart(8, '0')).join('');
type Settings = { publicOrigin: string; signatureHeader: string; deliveryHeader: string };
const defaults: Settings = { publicOrigin: '', signatureHeader: 'X-Webhook-Signature', deliveryHeader: 'Idempotency-Key' };

export function WebhookPanel({ task, binding, timezone, onChange }: { task: TaskDefinition; binding: WebhookBinding | null; timezone: string; onChange: (binding: WebhookBinding) => void }) {
  const [draft, setDraft] = useState<Settings | null>(null);
  const [secret, setSecret] = useState('');
  const [showSecret, setShowSecret] = useState(false);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [copyStatus, setCopyStatus] = useState('');
  const pending = useRef(false);
  const operation = useRef(taskOperationId());
  const settings = draft ?? binding ?? defaults;
  const localURL = binding ? `${ENGINE_URL.replace(/\/$/, '')}${binding.path}` : '';
  const url = binding?.publicOrigin ? `${binding.publicOrigin}${binding.path}` : localURL;
  function edit(change: Partial<Settings>) {
    setDraft({ publicOrigin: settings.publicOrigin, signatureHeader: settings.signatureHeader, deliveryHeader: settings.deliveryHeader, ...change });
    operation.current = taskOperationId();
  }
  async function save() {
    if (pending.current) return;
    pending.current = true; setBusy(true); setError('');
    try {
      const result = await saveWebhookBinding(task, { publicOrigin: settings.publicOrigin.trim(), signatureHeader: settings.signatureHeader.trim(), deliveryHeader: settings.deliveryHeader.trim() }, operation.current, binding?.revision ?? 0);
      if (result.secret) setSecret(result.secret);
      onChange(result.binding); setDraft(null); operation.current = taskOperationId();
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not save webhook.'); }
    finally { pending.current = false; setBusy(false); }
  }
  async function rotate() {
    if (pending.current) return;
    pending.current = true; setBusy(true); setError('');
    try { const result = await rotateWebhookSecret(task); setSecret(result.secret); onChange(result.binding); }
    catch (cause) { setError(cause instanceof Error ? cause.message : 'Could not rotate secret.'); }
    finally { pending.current = false; setBusy(false); }
  }
  async function copy(value: string, label: string) {
    try { await navigator.clipboard.writeText(value); setCopyStatus(`${label} copied.`); }
    catch { setError('Clipboard unavailable. Select and copy the value.'); }
  }
  return <section className="task-policy" aria-label="Webhook">
    <h3>Webhook</h3>
    {binding ? <>
      <div className="task-field"><label htmlFor={`webhook-url-${task.id}`}>{binding.publicOrigin ? 'Webhook URL' : 'Engine webhook URL'}</label><Input id={`webhook-url-${task.id}`} readOnly value={url} /><div><Button size="sm" onClick={() => void copy(url, 'URL')}>Copy URL</Button></div></div>
      {!binding.publicOrigin && <p>A public sender needs an HTTPS proxy or tunnel to this endpoint. Set its origin below to copy the public URL.</p>}
      <dl className="task-details-metadata">
        <div><dt>Method</dt><dd>POST</dd></div>
        <div><dt>Last delivery</dt><dd>{binding.lastReceiptAt ? `${formatRunDate(binding.lastReceiptAt, timezone)} · ${binding.lastDisposition}` : 'Not received'}</dd></div>
      </dl>
      <p>Send the secret as an Authorization: Bearer token, or sign the raw body with HMAC-SHA256 in <code>{binding.signatureHeader}</code>.</p>
    </> : <p>Create an endpoint to receive this task's payloads.</p>}
    {!task.deletedAt && <>
      <details className="task-details-disclosure">
        <summary>Endpoint settings</summary>
        <div className="task-webhook-settings">
          <div className="task-field"><label htmlFor={`origin-${task.id}`}>Public HTTPS origin (optional)</label><Input id={`origin-${task.id}`} disabled={busy} placeholder="https://hooks.example.com" value={settings.publicOrigin} onChange={event => edit({ publicOrigin: event.target.value })} /></div>
          <div className="task-field"><label htmlFor={`signature-${task.id}`}>Signature header</label><Input id={`signature-${task.id}`} disabled={busy} value={settings.signatureHeader} onChange={event => edit({ signatureHeader: event.target.value })} /></div>
          <div className="task-field"><label htmlFor={`delivery-${task.id}`}>Delivery ID header</label><Input id={`delivery-${task.id}`} disabled={busy} value={settings.deliveryHeader} onChange={event => edit({ deliveryHeader: event.target.value })} /></div>
          <p>Repeated delivery IDs reuse the accepted run. Without an ID, each request starts a new run.</p>
        </div>
      </details>
      <div className="task-details-actions"><Button disabled={busy || !settings.signatureHeader.trim() || !settings.deliveryHeader.trim()} onClick={() => void save()}>{busy ? 'Saving...' : binding ? 'Save webhook settings' : 'Create webhook'}</Button>{binding && <Button size="sm" disabled={busy} onClick={() => void rotate()}>Rotate secret</Button>}</div>
    </>}
    {secret && <div className="task-field"><label htmlFor={`secret-${task.id}`}>One-time webhook secret</label><Input id={`secret-${task.id}`} readOnly type={showSecret ? 'text' : 'password'} value={secret} /><div className="task-details-actions"><Button size="sm" onClick={() => void copy(secret, 'Secret')}>Copy secret</Button><Button size="sm" onClick={() => setShowSecret(value => !value)}>{showSecret ? 'Hide secret' : 'Show secret'}</Button><Button size="sm" onClick={() => { setSecret(''); setShowSecret(false); }}>Dismiss secret</Button></div></div>}
    {error && <p className="form-error" role="alert">{error}</p>}
    {copyStatus && <p role="status">{copyStatus}</p>}
  </section>;
}
