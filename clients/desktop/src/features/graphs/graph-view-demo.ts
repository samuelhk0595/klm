import type { Edge } from '@xyflow/react';
import type { CanvasNode } from './GraphCanvas';
import type { ChoiceDefinition } from './choice';

// Read-only catalog fixtures, not saved builder topology or executable definitions.
type GraphDrawing = { nodes: CanvasNode[]; edges: Edge[] };

function agent(id: string, name: string, agentId: string, x: number, y: number, initial = false): CanvasNode {
  return { id, type: 'aiAgent', position: { x, y }, data: { name, agentId, initial } };
}

function choice(id: string, x: number, y: number, terminal = false, config: Partial<ChoiceDefinition> = {}): CanvasNode {
  return { id, type: 'choice', position: { x, y }, data: { choice: { id, name: id, description: '', terminal, fields: [], outputFields: [], ...config } } };
}

function edge(source: string, target: string, sourceHandle: string, targetHandle: string): Edge {
  return { id: `${source}/${sourceHandle}/${target}`, source, target, sourceHandle, targetHandle };
}

export const graphDrawings: Record<string, GraphDrawing> = {
  'csv-export': {
    nodes: [
      agent('plan', 'Plan CSV export', 'planner', 0, 160, true),
      choice('ready', 380, 210),
      agent('implement', 'Implement export', 'implementer', 720, 160),
      choice('submit', 1100, 210),
      { id: 'checks', type: 'fork', position: { x: 1440, y: 160 }, data: { fork: { name: 'Verify export', branches: [
        { id: 'security', name: 'Security review', gitBranch: 'csv/security', outputFields: [] },
        { id: 'tests', name: 'Tests', gitBranch: 'csv/tests', outputFields: [] },
      ] } } },
      agent('security', 'Review access controls', 'security-reviewer', 1820, 0),
      { id: 'tests', type: 'terminal', position: { x: 1820, y: 360 }, data: { terminal: { name: 'Run tests', command: 'npm test' } } },
      choice('security_report', 2200, 50),
      { id: 'integrate', type: 'join', position: { x: 2540, y: 160 }, data: { join: { agentId: 'implementer', prompt: '', outputBranch: 'csv/integrated' } } },
      choice('integrated', 2920, 210),
      agent('review', 'Review delivery', 'delivery-reviewer', 3260, 160),
      choice('approve', 3640, 210, true),
    ],
    edges: [
      edge('plan', 'ready', 'choices', 'sources'), edge('ready', 'implement', 'destination', 'input'),
      edge('implement', 'submit', 'choices', 'sources'), edge('submit', 'checks', 'destination', 'input'),
      edge('checks', 'security', 'security', 'input'), edge('checks', 'tests', 'tests', 'input'),
      edge('security', 'security_report', 'choices', 'sources'), edge('security_report', 'integrate', 'destination', 'branches'),
      edge('tests', 'integrate', 'output', 'branches'), edge('integrate', 'integrated', 'choices', 'sources'),
      edge('integrated', 'review', 'destination', 'input'), edge('review', 'approve', 'choices', 'sources'),
    ],
  },
  'bug-fix': {
    nodes: [
      agent('investigate', 'Investigate issue', 'planner', 0, 0, true), choice('ready', 380, 50),
      agent('fix', 'Implement fix', 'implementer', 720, 0), choice('submit', 1100, 50),
      agent('review', 'Review fix', 'delivery-reviewer', 1440, 0), choice('approve', 1820, 50, true),
      choice('request_changes', 1460, 300, false, {
        description: 'Return actionable findings to the implementation activity.',
        fields: [{ key: 'findings', name: 'findings', type: 'string', required: true }],
        outputFields: [
          { key: 'task', name: 'task', value: '{{run.input.task}}' },
          { key: 'findings', name: 'findings', value: '{{choice.findings}}' },
          { key: 'instructions', name: 'instructions', value: 'Address the review findings: {{choice.findings}}' },
        ],
        sessionPolicy: 'continue_target',
      }),
    ],
    edges: [
      edge('investigate', 'ready', 'choices', 'sources'), edge('ready', 'fix', 'destination', 'input'),
      edge('fix', 'submit', 'choices', 'sources'), edge('submit', 'review', 'destination', 'input'),
      edge('review', 'approve', 'choices', 'sources'), edge('review', 'request_changes', 'choices', 'sources'),
      edge('request_changes', 'fix', 'destination', 'input'),
    ],
  },
  'code-review': {
    nodes: [
      agent('review', 'Review changes', 'delivery-reviewer', 0, 120, true),
      choice('approve', 440, 40, true, { description: 'Complete the review with no remaining findings.', fields: [{ key: 'note', name: 'note', type: 'string', required: false }] }),
      choice('request_changes', 440, 280, true, { description: 'Complete the review with actionable findings.', fields: [{ key: 'findings', name: 'findings', type: 'string', required: true }] }),
    ],
    edges: [edge('review', 'approve', 'choices', 'sources'), edge('review', 'request_changes', 'choices', 'sources')],
  },
};
