export type ChoicePayloadField = {
  key: string;
  name: string;
  type: 'string';
  required: boolean;
};

export type ChoiceOutputField = { key: string; name: string; value: string };

export type ChoiceDefinition = {
  id: string;
  name: string;
  description: string;
  terminal: boolean;
  fields: ChoicePayloadField[];
  // Settings for the outgoing transition; omitted policy means a new session.
  sessionPolicy?: 'new' | 'continue_target';
  outputFields?: ChoiceOutputField[];
};

export function choiceOutputFields(choice: ChoiceDefinition): ChoiceOutputField[] {
  if (choice.outputFields) return choice.outputFields;
  // Preserve drafts created before the UI switched from a separate transition prompt.
  const legacy = choice as ChoiceDefinition & { transitionPrompt?: string };
  return legacy.transitionPrompt?.trim() ? [{ key: crypto.randomUUID(), name: 'prompt', value: legacy.transitionPrompt }] : [];
}

export function choiceVariables(fields: ChoicePayloadField[]) {
  return [...new Set(fields.map(field => field.name.trim()).filter(Boolean).map(name => `choice.${name}`)), 'run.input.task'];
}

export type OutputContext = 'choice' | 'fork' | 'terminal';
type TemplatePart = { text: string } | { reference: string };
const fieldNamePattern = /^[a-z0-9]+(?:[a-z0-9_]*[a-z0-9])?$/;

// Same restricted grammar as compileOutput in the engine; never evaluate code.
function parseOutputTemplate(value: string, context: OutputContext, fields: ChoicePayloadField[]): TemplatePart[] {
  const parts: TemplatePart[] = [];
  while (value.length) {
    const start = value.indexOf('{{');
    if (start < 0) {
      if (value.includes('}}')) throw new Error('Malformed template.');
      parts.push({ text: value });
      break;
    }
    if (value.slice(0, start).includes('}}')) throw new Error('Malformed template.');
    parts.push({ text: value.slice(0, start) });
    const end = value.indexOf('}}', start + 2);
    if (end < 0) throw new Error('Unclosed template.');
    const reference = value.slice(start + 2, end).trim();
    const valid = reference === 'run.input.task'
      || (context === 'choice' && reference.startsWith('choice.') && fields.some(field => field.name.trim() === reference.slice(7) && field.type === 'string'))
      || (context !== 'choice' && reference.startsWith('payload.') && fieldNamePattern.test(reference.slice(8)))
      || (context === 'terminal' && reference === 'command.result');
    if (!valid) throw new Error(`Unknown ${context} output reference: {{${reference}}}`);
    parts.push({ reference });
    value = value.slice(end + 2);
  }
  return parts;
}

export function outputMappingError(output: ChoiceOutputField[], context: OutputContext, fields: ChoicePayloadField[] = []): string {
  const names = output.map(field => normalizeChoiceName(field.name, true));
  if (names.some(name => !name)) return 'Each output field needs a name.';
  if (new Set(names).size !== names.length) return 'Each output field needs a unique name.';
  try {
    for (const field of output) parseOutputTemplate(field.value, context, fields);
    return '';
  } catch (error) { return error instanceof Error ? error.message : 'Invalid output mapping.'; }
}

export function outputVariables(context: OutputContext, payloadFields: string[] = []) {
  return [...new Set(payloadFields.filter(name => fieldNamePattern.test(name)).map(name => `payload.${name}`)), ...(context === 'terminal' ? ['command.result'] : []), 'run.input.task'];
}

export function previewOutput(context: OutputContext, fields: ChoicePayloadField[], outputFields: ChoiceOutputField[], input: unknown, task: string, commandResult?: string): { output: Record<string, string> | null; error: string } {
  if (!input || typeof input !== 'object' || Array.isArray(input)) return { output: null, error: 'Input values must be a JSON object.' };
  const values = input as Record<string, unknown>;
  if (context === 'choice') {
    for (const name of Object.keys(values)) {
      if (!fields.some(field => field.name.trim() === name)) return { output: null, error: `Undeclared input field "${name}".` };
    }
    for (const field of fields) {
      const name = field.name.trim();
      if (field.required && !Object.hasOwn(values, name)) return { output: null, error: `Required field "${name}" is missing.` };
    }
  }
  for (const [name, value] of Object.entries(values)) {
    if (typeof value !== 'string') return { output: null, error: `Field "${name}" must be a string.` };
  }
  const error = outputMappingError(outputFields, context, fields);
  if (error) return { output: null, error };
  try {
    const output = Object.fromEntries(outputFields.map(field => [normalizeChoiceName(field.name, true), parseOutputTemplate(field.value, context, fields).map(part => {
      if ('text' in part) return part.text;
      if (part.reference === 'run.input.task') return task;
      if (part.reference === 'command.result') {
        if (commandResult === undefined) throw new Error('Cannot resolve command.result: local result is absent.');
        return commandResult;
      }
      const name = part.reference.slice(part.reference.indexOf('.') + 1);
      if (Object.hasOwn(values, name)) return values[name] as string;
      if (context === 'choice' && fields.some(field => field.name.trim() === name && !field.required)) return '';
      throw new Error(`Cannot resolve ${part.reference}: local field is absent.`);
    }).join('')]));
    return { output, error: '' };
  } catch (error) { return { output: null, error: error instanceof Error ? error.message : 'Cannot resolve output.' }; }
}

export function previewChoiceOutput(fields: ChoicePayloadField[], outputFields: ChoiceOutputField[], input: unknown, task: string) {
  return previewOutput('choice', fields, outputFields, input, task);
}

export function normalizeChoiceName(value: string, trim = false) {
  const normalized = value.trimStart().normalize('NFKD').replace(/\p{M}/gu, '').toLowerCase()
    .replace(/\s/g, '_').replace(/[^a-z0-9_]/g, '');
  return trim ? normalized.replace(/^_+|_+$/g, '') : normalized;
}
