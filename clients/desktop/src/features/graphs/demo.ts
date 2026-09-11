// UI-only graph records. Topology and execution will be configured in the graph builder.
export type Graph = { id: string; name: string; description: string; enabled: boolean };

export const initialGraphs: Graph[] = [
  { id: 'csv-export', name: 'CSV export', description: 'Plan, implement, and verify order exports with security and test checks.', enabled: true },
  { id: 'bug-fix', name: 'Bug fix', description: 'Investigate a reported issue, implement a fix, and review the result.', enabled: true },
  { id: 'code-review', name: 'Code review', description: 'Review changes and gather actionable feedback before delivery.', enabled: true },
];
