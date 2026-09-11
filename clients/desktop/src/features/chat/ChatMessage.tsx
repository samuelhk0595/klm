import { Atom, Check, Copy, FileText, ThumbsDown, ThumbsUp } from 'lucide-react';
import { useMemo, useState } from 'react';
import Markdown, { type Components } from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { IconButton } from '../../design-system/Button';
import type { Message } from '../workspace/demo';
import { MentionText } from './MentionText';
import { Flowchart } from './Flowchart';
import { normalizeFlowchartMarkdown } from './flowchartMarkdown';

const markdownPlugins = [remarkGfm];
const markdownComponents: Components = {
  a: ({ href, children }) => <a href={href || undefined} target={href?.startsWith('#') ? undefined : '_blank'} rel="noopener noreferrer">{children}</a>,
  // Model-generated images must not trigger unsolicited external requests.
  img: ({ src, alt }) => <a href={typeof src === 'string' && src ? src : undefined} target="_blank" rel="noopener noreferrer">{alt || 'Image'}</a>,
  table: ({ children }) => <div className="markdown-table" role="region" aria-label="Table" tabIndex={0}><table>{children}</table></div>,
  pre: ({ node, children }) => {
    const code = node?.children.find(child => child.type === 'element' && child.tagName === 'code');
    if (code?.type === 'element') {
      const source = code.children.map(child => child.type === 'text' ? child.value : '').join('').trimEnd();
      const classes = code.properties.className;
      const language = Array.isArray(classes) ? classes.find(value => String(value).startsWith('language-')) : undefined;
      if (language === 'language-mermaid' || (!language && /^\s*(?:flowchart|graph)\s+(?:TB|TD|BT|RL|LR)\b/.test(source))) return <Flowchart source={source} />;
    }
    return <pre tabIndex={0} aria-label="Code block">{children}</pre>;
  },
};

export function ChatMessage({ message }: { message: Message }) {
  const [copied, setCopied] = useState(false);
  const [copyError, setCopyError] = useState(false);
  const [feedback, setFeedback] = useState<'up' | 'down' | null>(null);
  const markdown = useMemo(() => message.role === 'assistant' ? normalizeFlowchartMarkdown(message.text) : message.text, [message.role, message.text]);
  async function copy() {
    try { await navigator.clipboard.writeText(message.text); setCopied(true); setCopyError(false); }
    catch { setCopyError(true); }
  }
  return <article data-message-id={message.id} className={`message message--${message.role}`} aria-label={`${message.role === 'user' ? 'Your' : 'Agent'} message`}>
    <div className="message-body">
      {message.context && <div className="context-injections">{message.context.map(file => <div key={file}><FileText /><span>Context injection</span><span aria-hidden="true">·</span><span className="context-file">{file}</span></div>)}</div>}
      {message.thought && <details className="thought"><summary><Atom /><strong>Think</strong><span>·</span><span className="thought-preview">{message.thought}</span></summary><p>{message.thought}</p></details>}
      {message.role === 'assistant' ? <div className="message-text markdown-body"><Markdown remarkPlugins={markdownPlugins} components={markdownComponents} skipHtml>{markdown}</Markdown></div> : <p className="message-text"><MentionText text={message.text} mentions={message.mentions} preparation={message.mentionPreparation} /></p>}
      <div className="message-actions"><IconButton label={copied ? 'Copied' : 'Copy message'} onClick={copy}>{copied ? <Check /> : <Copy />}</IconButton>{message.role === 'assistant' && <><IconButton label="Helpful response" aria-pressed={feedback === 'up'} onClick={() => setFeedback(feedback === 'up' ? null : 'up')}><ThumbsUp /></IconButton><IconButton label="Unhelpful response" aria-pressed={feedback === 'down'} onClick={() => setFeedback(feedback === 'down' ? null : 'down')}><ThumbsDown /></IconButton></>}</div>
      {copyError && <p role="status" className="small muted">Clipboard unavailable. Select the message to copy it.</p>}
    </div>
  </article>;
}
