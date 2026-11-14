import { useRef, useState } from 'react';
import { Button } from '../../design-system/Button';
import { request, type DelegationRecovery as Recovery, type SessionResponse } from '../../engine';

export function DelegationRecovery({ sessionId, recovery, onSnapshot }: { sessionId: string; recovery: Recovery; onSnapshot: (value: SessionResponse) => void }) {
  const [confirming, setConfirming] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const operation = useRef(crypto.randomUUID());
  const pending = useRef(false);
  async function recover() {
    if (pending.current) return;
    pending.current = true; setBusy(true); setError('');
    try {
      const result = await request<SessionResponse>(`/api/sessions/${encodeURIComponent(sessionId)}/delegation/batches/${encodeURIComponent(recovery.batchId)}/recover`, 'POST', {
        action: 'continue_without_replay', operationId: operation.current, revision: recovery.revision,
      });
      onSnapshot(result);
    } catch (cause) { setError(cause instanceof Error ? cause.message : 'Recovery could not be confirmed. Refresh or retry this decision.'); }
    finally { pending.current = false; setBusy(false); }
  }
  return <section className="permission-card" aria-label="Subagent recovery">
    <div className="permission-heading"><strong>Subagent recovery required</strong></div>
    <p className="permission-description">Previous child execution or result delivery could not be confirmed.</p>
    {recovery.canRecover ? <>
      {confirming && <p className="permission-description">Continue using the saved results. Previous work will not be replayed, and unconfirmed children will remain quarantined. Review existing changes before retrying unfinished work.</p>}
      <div className="permission-actions">
        {confirming ? <><Button size="sm" variant="primary" disabled={busy} onClick={() => void recover()}>{busy ? 'Recovering...' : 'Continue without replay'}</Button><Button size="sm" disabled={busy} onClick={() => setConfirming(false)}>Cancel</Button></>
          : <Button size="sm" onClick={() => setConfirming(true)}>Recover subagents</Button>}
      </div>
    </> : <p className="form-error">{recovery.reason}</p>}
    {error && <p className="form-error" role="alert">{error}</p>}
  </section>;
}
