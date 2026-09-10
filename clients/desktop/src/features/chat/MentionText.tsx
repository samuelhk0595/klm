import type { ReactNode } from 'react';
import type { FileMention, MentionPreparation } from '../../engine';
import { MentionBadge } from './MentionBadge';
import { mentionToken } from './mentions';

export function MentionText({ text, mentions = [], preparation = [] }: { text: string; mentions?: FileMention[]; preparation?: MentionPreparation[] }) {
  const content: ReactNode[] = [];
  let offset = 0;
  for (const mention of [...mentions].sort((a, b) => a.start - b.start)) {
    if (mention.start < offset || text.slice(mention.start, mention.end) !== mentionToken(mention)) continue;
    content.push(text.slice(offset, mention.start));
    content.push(<MentionBadge key={mention.id} mention={mention} truncated={preparation.find(item => item.path === mention.path)?.truncated} />);
    offset = mention.end;
  }
  content.push(text.slice(offset));
  return <>{content}</>;
}
