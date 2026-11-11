import type { ChoiceOutputField } from './choice';

export type TerminalDefinition = {
  shell?: 'powershell' | 'bash' | 'sh';
  name: string;
  command: string;
  outputFields?: ChoiceOutputField[];
};
