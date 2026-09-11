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

export function unknownOutputVariables(outputFields: ChoiceOutputField[], fields: ChoicePayloadField[]) {
  const available = new Set(choiceVariables(fields));
  const used = outputFields.flatMap(field => [...field.value.matchAll(/\{\{\s*([^{}]*?)\s*\}\}/g)].map(match => match[1].trim()));
  return [...new Set(used.filter(variable => !available.has(variable)))];
}

// Local preview only; harness submission validation and execution remain outside this UI.
export function previewChoiceOutput(fields: ChoicePayloadField[], outputFields: ChoiceOutputField[], input: unknown, task: string): { output: Record<string, string | null> | null; error: string } {
  if (!input || typeof input !== 'object' || Array.isArray(input)) return { output: null, error: 'Input values must be a JSON object.' };
  const values = input as Record<string, unknown>;
  for (const field of fields) {
    const name = field.name.trim();
    const present = Object.prototype.hasOwnProperty.call(values, name);
    if (field.required && !present) return { output: null, error: `Required field "${name}" is missing.` };
    if (present && typeof values[name] !== 'string') return { output: null, error: `Field "${name}" must be a string.` };
  }
  const unknown = unknownOutputVariables(outputFields, fields);
  if (unknown.length) return { output: null, error: `Unknown variable: {{${unknown[0]}}}` };
  function resolve(variable: string): string | null {
    const path = variable.trim();
    if (path === 'run.input.task') return task;
    const name = path.slice('choice.'.length);
    return path.startsWith('choice.') && Object.prototype.hasOwnProperty.call(values, name) && typeof values[name] === 'string' ? values[name] as string : null;
  }
  const output = Object.fromEntries(outputFields.map(field => {
    const expression = /^\{\{\s*([^{}]*?)\s*\}\}$/.exec(field.value);
    const value = expression ? resolve(expression[1]) : field.value.replace(/\{\{\s*([^{}]*?)\s*\}\}/g, (_, variable: string) => resolve(variable) ?? 'null');
    return [field.name.trim(), value];
  }));
  return { output, error: '' };
}

export function normalizeChoiceName(value: string, trim = false) {
  const normalized = value.trimStart().normalize('NFKD').replace(/\p{M}/gu, '').toLowerCase()
    .replace(/\s/g, '_').replace(/[^a-z0-9_]/g, '');
  return trim ? normalized.replace(/^_+|_+$/g, '') : normalized;
}
