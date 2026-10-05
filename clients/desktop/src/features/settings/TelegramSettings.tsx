import { useId, useState } from 'react';
import { Bot, Check, Link2, Send, Unplug, UserRound } from 'lucide-react';
import { Button } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import type { TelegramLink } from '../tasks/types';
import './telegram.css';

// Deliberately no editable credential, network transport, or persisted token.
const fixtureBot = 'klm_fixture_bot';
const fixtureAccount = 'fixture_owner';

export function TelegramSettings({ link, onChange }: { link: TelegramLink; onChange: (link: TelegramLink) => void }) {
  const id = useId();
  const [confirm, setConfirm] = useState<'bot' | 'account' | null>(null);
  const [notice, setNotice] = useState('');

  function beginPairing() {
    // New/cancelled codes replace the only current code; no old receipt can be applied.
    const code = Array.from(crypto.getRandomValues(new Uint32Array(2)), value => value.toString(16).padStart(8, '0')).join('-').toUpperCase();
    onChange({ ...link, pairingCode: `KLM-${code}` });
    setNotice('');
  }

  return <section className="telegram-settings" aria-labelledby={`${id}-title`}>
    <div className="settings-panel-title"><h3 id={`${id}-title`}>Telegram</h3><span className={`telegram-status ${link.account && link.online ? 'is-linked' : ''}`}>{link.account && link.online ? <Check /> : <Send />}{!link.bot ? 'Not connected' : !link.online ? 'Offline' : link.account ? 'Linked' : 'Not linked'}</span></div>
    <div className="telegram-card">
      <div className="telegram-card-heading"><Bot /><h4>Bot</h4>{link.bot && <span className={`telegram-status ${link.online ? 'is-linked' : ''}`}>{link.online ? 'Connected' : 'Offline'}</span>}</div>
      {link.bot ? <>
        <div className="telegram-identity"><strong>@{link.bot}</strong><span>Engine-wide · Long polling</span></div>
        {!link.online && <p className="task-link-needed">Telegram delivery is unavailable. Questions remain in KLM.</p>}
        <Button size="sm" onClick={() => { onChange({ ...link, online: !link.online }); setNotice(''); }}>{link.online ? 'Simulate offline' : 'Restore connection'}</Button>
        {confirm === 'bot' ? <div className="telegram-confirm" role="group" aria-label="Disconnect bot confirmation"><p>Disconnect @{link.bot}? The linked account will also be removed.</p><div className="dialog-actions"><Button autoFocus size="sm" onClick={() => setConfirm(null)}>Cancel</Button><Button size="sm" variant="danger-soft" onClick={() => { onChange({ bot: '', account: '', online: false }); setConfirm(null); setNotice('Bot disconnected. Questions remain in KLM.'); }}><Unplug />Disconnect bot</Button></div></div> : <Button size="sm" onClick={() => setConfirm('bot')}><Unplug />Disconnect bot</Button>}
      </> : <form onSubmit={event => {
        event.preventDefault();
        onChange({ bot: fixtureBot, account: '', online: true });
        setNotice('');
      }}>
        <div className="settings-field"><label htmlFor={`${id}-token`}>Bot token</label><Input id={`${id}-token`} type="password" readOnly autoComplete="off" value="fixture-token" /></div>
        <div className="dialog-actions"><Button type="submit" variant="primary">Connect bot</Button></div>
      </form>}
    </div>
    <div className="telegram-card">
      <div className="telegram-card-heading"><UserRound /><h4>Recipient</h4></div>
      {link.account ? <>
        <div className="telegram-identity"><strong>@{link.account}</strong><span>Simulated bot interaction received · {link.recipientId}</span></div>
        {confirm === 'account' ? <div className="telegram-confirm" role="group" aria-label="Unlink account confirmation"><p>Unlink @{link.account} from Telegram task questions?</p><div className="dialog-actions"><Button autoFocus size="sm" onClick={() => setConfirm(null)}>Cancel</Button><Button size="sm" variant="danger-soft" onClick={() => { onChange({ ...link, account: '', recipientId: undefined, pairingCode: undefined }); setConfirm(null); setNotice('Account unlinked. Questions remain in KLM.'); }}>Unlink account</Button></div></div> : <Button size="sm" onClick={() => setConfirm('account')}><Unplug />Unlink account</Button>}
      </> : link.pairingCode && link.bot ? <div>
        <p className="telegram-pairing-instruction">Waiting for bot interaction</p>
        <div className="telegram-pairing-code"><code>{link.pairingCode}</code><Button size="sm" onClick={beginPairing}>New code</Button></div>
        <div className="dialog-actions"><Button onClick={() => { onChange({ ...link, pairingCode: undefined }); setNotice('Pairing cancelled.'); }}>Cancel</Button><Button variant="primary" disabled={!link.online} onClick={() => {
          if (!link.pairingCode || !link.online) return;
          // A local stand-in for an incoming update correlated to the current code.
          onChange({ ...link, account: fixtureAccount, recipientId: 'fixture-user-01 / fixture-chat-01', pairingCode: undefined });
          setNotice('Simulated pairing interaction received.');
        }}>Simulate bot interaction</Button></div>
      </div> : <><p className="muted">{link.bot ? 'No recipient linked. Questions remain in KLM.' : 'Connect a bot to link your account.'}</p><Button disabled={!link.bot || !link.online} onClick={beginPairing}><Link2 />Link account</Button></>}
    </div>
    <p className="telegram-notice" role="status">{notice}</p>
  </section>;
}
