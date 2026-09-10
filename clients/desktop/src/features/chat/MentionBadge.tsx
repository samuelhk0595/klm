import type { FileMention } from '../../engine';

export function mentionName(mention: Pick<FileMention, 'path'>) {
  return mention.path.split('/').pop() || mention.path;
}

export function MentionBadge({ mention, truncated = false }: { mention: FileMention; truncated?: boolean }) {
  const name = mentionName(mention);
  return <span className="mention-badge" data-kind={mention.kind} data-truncated={truncated || undefined}
    aria-label={`${mention.kind === 'directory' ? 'Folder' : 'File'}: ${name}${truncated ? ' (truncated)' : ''}`}
    title={truncated ? 'Truncated attachment' : undefined}>{name}</span>;
}

// The contenteditable owns its DOM so ordinary typing keeps browser selection/undo.
// Keep its badge markup equivalent to the history component above.
export function createMentionBadge(mention: FileMention) {
  const badge = document.createElement('span');
  badge.className = 'mention-badge';
  badge.dataset.kind = mention.kind;
  badge.dataset.mentionId = mention.id;
  badge.contentEditable = 'false';
  badge.textContent = mentionName(mention);
  badge.setAttribute('aria-label', `${mention.kind === 'directory' ? 'Folder' : 'File'}: ${mentionName(mention)}`);
  return badge;
}
