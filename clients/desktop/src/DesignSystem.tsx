import { useState } from 'react';
import { Plus, Settings } from 'lucide-react';
import { Button, IconButton } from './design-system/Button';
import { Badge, StatusIndicator } from './design-system/Badge';
import { Section } from './design-system/Section';
import { Select } from './design-system/Select';
import { TabNav } from './design-system/TabNav';
import { Menu, MenuItem } from './design-system/Menu';
import { Slider } from './design-system/Slider';
import { MessageComposer, type ComposerDraft } from './features/chat/MessageComposer';
import { ChatMessage } from './features/chat/ChatMessage';
import { initialSessions } from './features/workspace/demo';

const markdownExample = [
  '## Session summary',
  '',
  '**Markdown** supports *emphasis*, ~~strikethrough~~, and `inline code`.',
  '',
  '1. Choose a harness.',
  '2. Send a message.',
  '   - Review the response.',
  '',
  '> Keep each change focused.',
  '',
  '```typescript',
  'const session = { harness: "opencode", status: "idle" };',
  '```',
  '',
  '| Harness | Status |',
  '| :--- | :--- |',
  '| OpenCode | Ready |',
  '',
  '- [x] Create session',
  '- [ ] Send next message',
  '',
  '[Markdown reference](https://commonmark.org/help/)',
].join('\n');

export function DesignSystem() {
  const [tab, setTab] = useState('chat');
  const [notice, setNotice] = useState('');
  const [draft, setDraft] = useState<ComposerDraft>({ text: '', mentions: [] });
  const [menuOpen, setMenuOpen] = useState(false);
  const [level, setLevel] = useState(40);
  return <div className="design-system">
    <Badge tone="accent">KLM foundations / 0.1</Badge><h2>A quiet workspace for complex work.</h2><p className="muted">Tokens and components extracted from the supplied Harness page. Shared primitives below power the actual workspace.</p>
    <Section title="Color"><div className="swatch-grid">{['surface', 'sidebar', 'subtle', 'border', 'text', 'muted', 'accent', 'accent-soft', 'success', 'warning'].map(color => <div key={color} className="swatch"><div style={{ background: `var(--color-${color})` }} /><code>{color}</code></div>)}</div></Section>
    <Section title="Typography"><div className="type-samples"><h2>Inter / Workspace heading</h2><p>Body / A clear, focused conversation with your agent.</p><span className="small muted">Metadata / Context injection · AGENTS.md</span><code>Monospace / clients/desktop</code></div></Section>
    <Section title="Spacing and shape"><div className="spacing-samples">{[1, 2, 3, 4, 6, 8].map(space => <div key={space}><span style={{ width: `var(--space-${space})` }} /><code>space-{space}</code></div>)}</div><p className="small muted">6px controls · 8px buttons · 12px panels · 16px messages · 24px composer</p></Section>
    <Section title="Buttons"><div className="component-row"><Button variant="primary" onClick={() => setNotice('Primary button clicked')}>Primary</Button><Button onClick={() => setNotice('New session button clicked')}><Plus />New Session</Button><Button variant="ghost" onClick={() => setNotice('Ghost button clicked')}>Ghost</Button><Button variant="soft" onClick={() => setNotice('Soft button clicked')}>Soft</Button><Button disabled>Disabled</Button><IconButton label="Example settings" onClick={() => setNotice('Icon button clicked')}><Settings /></IconButton></div></Section>
    <Section title="Badges and status"><div className="component-row"><Badge>Harness</Badge><Badge tone="accent">Ready</Badge><StatusIndicator label="Connected" /><StatusIndicator status="pending" label="In progress" /><StatusIndicator status="offline" label="Offline" /></div></Section>
    <Section title="Navigation and controls"><TabNav value={tab} onChange={setTab} items={[{ value: 'chat', label: 'Chat' }, { value: 'graph', label: 'Graph' }]} /><p className="small muted">Selected: {tab}</p><Select label="Example permissions"><option>Workspace Write</option><option>Read Only</option></Select></Section>
    <Section title="Collapsible section" collapsible><p className="muted">Native details and summary provide keyboard-accessible disclosure without a JavaScript dependency.</p></Section>
    <Section title="Menu and slider"><div className="component-row"><Menu label="Example menu" open={menuOpen} onOpenChange={setMenuOpen} trigger={props => <Button {...props}>Open menu</Button>}>{['Rename', 'Duplicate', 'Archive'].map(action => <MenuItem key={action} onClick={() => { setNotice(`${action} selected`); setMenuOpen(false); }}>{action}</MenuItem>)}</Menu><div style={{ width: 240 }}><Slider label="Level" value={level} valueText={`${level}%`} step={10} onValueChange={setLevel} marks={[{ value: 0, label: '0' }, { value: 50, label: '50' }, { value: 100, label: '100' }]} /></div></div></Section>
    <Section title="Messages"><div className="message-examples">{initialSessions[0].messages.map(message => <ChatMessage key={message.id} message={message} />)}</div></Section>
    <Section title="Markdown"><ChatMessage message={{ id: 'markdown-example', role: 'assistant', text: markdownExample }} /></Section>
    <Section title="Composer"><MessageComposer draft={draft} onDraftChange={setDraft} onSend={draft => setNotice(`Example message: ${draft.text}`)} /></Section>
    <p role="status" className="gallery-notice">{notice}</p>
  </div>;
}
