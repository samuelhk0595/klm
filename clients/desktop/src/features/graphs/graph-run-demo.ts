// UI-only snapshots. Payloads and file operations below are fictional, never executed.
export type RunLogEvent = { time: string; kind: 'Thinking' | 'Read' | 'Search' | 'Command' | 'Write' | 'Choice' | 'Dispatch'; text: string };
export type NodeActivity = {
  status: 'pending' | 'running' | 'completed';
  events: RunLogEvent[];
  input?: Record<string, unknown>;
  output?: { choice?: string; payload: Record<string, unknown> };
};

const task = 'Add CSV export for orders, respecting the active filters and the user’s access to orders.';
const plan = 'Reuse the authorized order query. Export every matching row with CSV escaping. Add the download action and verify filters and cross-account access.';
const choiceInput = { plan };
const choiceOutput = { task, plan, instructions: 'Implement the export and verify the acceptance criteria in the plan.' };
const planEvents: RunLogEvent[] = [
  { time: '00:00', kind: 'Thinking', text: 'Mapping order filters and access rules.' },
  { time: '00:03', kind: 'Read', text: 'src/features/orders/OrdersPage.tsx' },
  { time: '00:06', kind: 'Search', text: 'Order query and permission checks' },
  { time: '00:09', kind: 'Read', text: 'src/api/orders.ts' },
  { time: '00:12', kind: 'Thinking', text: 'Reuse the authorized query and export all matching rows.' },
  { time: '00:15', kind: 'Command', text: 'git diff --stat' },
  { time: '00:18', kind: 'Write', text: 'docs/plans/csv-export.md' },
  { time: '00:21', kind: 'Thinking', text: 'Checking the plan against the access constraints.' },
  { time: '00:24', kind: 'Choice', text: 'ready' },
];
const implementationEvents: RunLogEvent[] = [
  { time: '00:00', kind: 'Read', text: 'docs/plans/csv-export.md' },
  { time: '00:03', kind: 'Read', text: 'src/api/orders.ts' },
  { time: '00:06', kind: 'Thinking', text: 'Sharing the existing filter and authorization logic.' },
  { time: '00:09', kind: 'Write', text: 'src/api/orders/export.ts' },
  { time: '00:12', kind: 'Write', text: 'src/features/orders/ExportButton.tsx' },
  { time: '00:15', kind: 'Thinking', text: 'Checking CSV escaping and empty results.' },
  { time: '00:18', kind: 'Read', text: 'src/features/orders/OrdersPage.tsx' },
];
const pendingActivity: NodeActivity = { status: 'pending', events: [] };

export function getNodeActivity(graphId: string, nodeId: string, step: number): NodeActivity {
  if (graphId !== 'csv-export') return pendingActivity;
  if (nodeId === 'plan') return {
    status: step < 8 ? 'running' : 'completed', input: { task },
    events: planEvents.slice(0, Math.min(step + 1, planEvents.length)),
    // The agent submits the selected Choice's INPUT payload.
    output: step >= 8 ? { choice: 'ready', payload: choiceInput } : undefined,
  };
  if (nodeId === 'ready' && step >= 8) return {
    status: step === 8 ? 'running' : 'completed', input: choiceInput,
    events: [{ time: '00:00', kind: 'Dispatch', text: step === 8 ? 'Preparing the destination payload.' : 'Payload delivered to Implement export.' }],
    // The Choice assembles a separate OUTPUT payload for its destination.
    output: step >= 9 ? { payload: choiceOutput } : undefined,
  };
  if (nodeId === 'implement' && step >= 9) return {
    status: 'running', input: choiceOutput,
    events: implementationEvents.slice(0, step - 8),
  };
  return pendingActivity;
}
