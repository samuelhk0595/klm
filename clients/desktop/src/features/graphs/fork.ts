import { normalizeChoiceName, type ChoiceOutputField } from './choice';

export type ForkBranch = {
  id: string;
  name: string;
  outputFields: ChoiceOutputField[];
  gitBranch: string;
};
export type ForkDefinition = { name: string; branches: ForkBranch[] };

export function createForkBranch(name = '', values: Record<string, string> = {}, gitBranch = ''): ForkBranch {
  return { id: crypto.randomUUID(), name, outputFields: Object.entries(values).map(([name, value]) => ({ key: crypto.randomUUID(), name, value })), gitBranch };
}

export function createFork(): ForkDefinition {
  return {
    name: 'Parallel checks',
    branches: [
      createForkBranch('Security review', { instructions: 'Review the changes for security issues.', focus: 'Access control and input validation' }, 'parallel/security-review'),
      createForkBranch('Tests', { instructions: 'Write tests for the implemented changes.', report: 'Test coverage and results' }, 'parallel/write-tests'),
    ],
  };
}

export function forkOutputError(fields: ForkBranch['outputFields']): string {
  const names = fields.map(field => normalizeChoiceName(field.name, true));
  if (names.some(name => !name)) return 'Each output field needs a name.';
  if (new Set(names).size !== names.length) return 'Each output field needs a unique name.';
  return '';
}

// Literal new-branch names, matching git check-ref-format --branch.
// Checkout-history expressions such as @{-1} are not new branch names.
export function gitBranchNameError(name: string): string {
  if (!name) return 'Enter a branch name.';
  if (name === 'HEAD') return 'HEAD is reserved by Git.';
  if (name.startsWith('-')) return 'A branch name cannot start with a hyphen.';
  if (/[\x00-\x20\x7f~^:?*\[\\]/.test(name)) return 'Spaces, control characters, ~, ^, :, ?, *, [, and \\ are not allowed.';
  if (name.includes('..')) return 'Consecutive dots are not allowed.';
  if (name.includes('@{')) return 'The sequence @{ is not allowed.';
  if (name.endsWith('.')) return 'A branch name cannot end with a dot.';
  const components = name.split('/');
  if (components.some(part => !part)) return 'Slashes must separate non-empty name components.';
  if (components.some(part => part.startsWith('.'))) return 'A name component cannot start with a dot.';
  if (components.some(part => part.endsWith('.lock'))) return 'A name component cannot end with .lock.';
  return '';
}
