import { useRef, useState } from 'react';
import { ShieldQuestion } from 'lucide-react';
import { Button } from '../../design-system/Button';
import { request, type PermissionDecision, type PermissionRequest, type Session } from '../../engine';

export function PermissionCard({ permission, sessionId, projectName, onResolved }: {
  permission: PermissionRequest; sessionId: string; projectName: string; onResolved: (session: Session) => void;
}) {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const pending = useRef(false);
  const busy = submitting || permission.resolving;
  async function decide(decision: PermissionDecision) {
    if (pending.current || permission.resolving) return;
    pending.current = true; setSubmitting(true); setError('');
    try {
      onResolved(await request<Session>(`/api/sessions/${encodeURIComponent(sessionId)}/permissions/${encodeURIComponent(permission.id)}`, 'POST', { decision }));
    } catch (error) { setError(error instanceof Error ? error.message : 'Permission decision could not be sent. Please retry.'); }
    finally { pending.current = false; setSubmitting(false); }
  }
  const harness = permission.harness === 'opencode' ? 'OpenCode' : permission.harness === 'pi' ? 'Pi' : 'Codex';
  return <section className="permission-card" aria-label={`${harness} permission request`}>
    <div className="permission-heading"><ShieldQuestion aria-hidden="true" /><strong>{permission.title || 'Permission required'}</strong><span>{harness}</span></div>
    {permission.description && permission.description !== permission.patterns.join('\n') && <p className="permission-description">{permission.description}</p>}
    {permission.patterns.length > 0 && <ul className="permission-patterns">{permission.patterns.map((pattern, index) => <li key={index}><code>{pattern}</code></li>)}</ul>}
    {permission.details && <details className="permission-details"><summary>Request details</summary><pre>{JSON.stringify(permission.details, null, 2)}</pre></details>}
    {permission.decisions.includes('always') && <p className="permission-scope">Remembered approvals apply only to this exact request in {projectName}, using {harness}.</p>}
    <div className="permission-actions">
      {permission.decisions.includes('once') && <Button size="sm" variant="primary" disabled={busy} onClick={() => void decide('once')}>{permission.allowLabel || 'Allow'}</Button>}
      {permission.decisions.includes('session') && <Button size="sm" disabled={busy} onClick={() => void decide('session')}>Allow this session</Button>}
      {permission.decisions.includes('always') && <Button size="sm" disabled={busy} onClick={() => void decide('always')}>Always allow</Button>}
      {permission.decisions.includes('reject') && <Button size="sm" variant="ghost" disabled={busy} onClick={() => void decide('reject')}>Deny</Button>}
    </div>
    {error && <p className="form-error" role="alert">{error}</p>}
  </section>;
}
