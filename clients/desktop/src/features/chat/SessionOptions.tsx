import { EllipsisVertical } from 'lucide-react';
import { useRef, useState } from 'react';
import { IconButton } from '../../design-system/Button';
import { Menu } from '../../design-system/Menu';
import { Toggle } from '../../design-system/Toggle';
import { request, type Session, type SessionResponse } from '../../engine';

export function SessionOptions({ session, disabled, onSnapshot }: {
  session: Session; disabled: boolean; onSnapshot: (snapshot: SessionResponse) => void;
}) {
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const pending = useRef(false);
  const locked = disabled || saving || session.status === 'running';
  async function save(yolo: boolean) {
    if (pending.current || locked) return;
    pending.current = true; setSaving(true); setError('');
    try {
      onSnapshot(await request<SessionResponse>(`/api/sessions/${encodeURIComponent(session.id)}/permissions`, 'PATCH', { yolo }));
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Could not change YOLO mode.');
    } finally { pending.current = false; setSaving(false); }
  }
  return <div className="session-options">
    {session.yolo && <span className="yolo-indicator" title="YOLO mode is active for this session">YOLO</span>}
    <Menu label="Session options" className="session-options-menu" open={open} onOpenChange={setOpen} trigger={props => <IconButton {...props} label="Session options" size="sm"><EllipsisVertical /></IconButton>}>
      <Toggle label="YOLO mode" checked={!!session.yolo} disabled={locked} onCheckedChange={value => void save(value)} title={session.status === 'running' ? 'Stop or finish the current turn to change YOLO mode' : 'Skip permission approvals and saved restrictions for this session'} />
      {error && <p className="form-error menu-feedback" role="alert">{error}</p>}
    </Menu>
  </div>;
}
