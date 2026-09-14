import type { ChoiceOutputField } from './choice';

export type TerminalDefinition = {
  name: string;
  command: string;
  outputFields?: ChoiceOutputField[];
};
