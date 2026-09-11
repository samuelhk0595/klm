const flowchartHeader = /^\s{0,3}(?:flowchart|graph)\s+(?:TB|TD|BT|RL|LR)\s*;?\s*$/;
const flowchartStatement = /^(?:%%|subgraph\b|end\s*;?$|direction\s+(?:TB|TD|BT|RL|LR)\b|(?:classDef|class|style|linkStyle|click)\s|[\p{L}\p{N}_][\p{L}\p{N}_-]*\s*(?:\[|\(|\{|>|@\{|--|==|-\.|~~~|&))/u;

// Some harness responses omit fences. Only wrap a standalone flowchart header
// followed by recognizable statements; leave surrounding prose and code intact.
export function normalizeFlowchartMarkdown(text: string): string {
  const lines = text.split('\n');
  const output: string[] = [];
  let fence: { marker: string; length: number } | undefined;
  for (let index = 0; index < lines.length; index++) {
    const line = lines[index];
    const delimiter = /^\s{0,3}(`{3,}|~{3,})(.*)$/.exec(line);
    if (fence) {
      output.push(line);
      if (delimiter && delimiter[1][0] === fence.marker && delimiter[1].length >= fence.length && !delimiter[2].trim()) fence = undefined;
      continue;
    }
    if (delimiter) {
      fence = { marker: delimiter[1][0], length: delimiter[1].length };
      output.push(line);
      continue;
    }
    if (!flowchartHeader.test(line)) {
      output.push(line);
      continue;
    }
    let end = index + 1;
    while (end < lines.length && flowchartStatement.test(lines[end].trim())) end++;
    if (end === index + 1) {
      output.push(line);
      continue;
    }
    output.push('', '```mermaid', ...lines.slice(index, end), '```', '');
    index = end - 1;
  }
  return output.join('\n');
}
