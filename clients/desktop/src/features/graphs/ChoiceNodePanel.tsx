import { useEffect, useId, useState } from 'react';
import { Plus, Trash2, X } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { RadioGroup } from '../../design-system/RadioGroup';
import { Toggle } from '../../design-system/Toggle';
import { agentSlug as choiceSlug } from '../agents/demo';
import { ChoiceOutputEditor } from './ChoiceOutputEditor';
import { ChoiceOutputPreview } from './ChoiceOutputPreview';
import { choiceOutputFields, choiceVariables, normalizeChoiceName, unknownOutputVariables, type ChoiceDefinition, type ChoicePayloadField } from './choice';

export function ChoiceNodePanel({ choice, otherChoices, onSave, onClose }: {
  choice: ChoiceDefinition;
  otherChoices: ChoiceDefinition[];
  onSave: (choice: ChoiceDefinition) => void;
  onClose: () => void;
}) {
  const id = useId();
  const [draft, setDraft] = useState(() => ({ ...choice, name: normalizeChoiceName(choice.name, true), outputFields: choiceOutputFields(choice) }));
  const slug = choiceSlug(draft.name);
  const duplicate = !!slug && otherChoices.some(item => item.id === slug);
  const nameError = duplicate ? 'A choice with this identifier already exists.' : draft.name.trim() && !slug ? 'Use at least one letter or number in the name.' : '';
  const fieldNames = draft.fields.map(field => field.name.trim());
  const duplicateFields = fieldNames.some((name, index) => name && fieldNames.indexOf(name) !== index);
  const validInput = !duplicateFields && fieldNames.every(Boolean);
  const outputNames = draft.outputFields.map(field => field.name.trim());
  const duplicateOutputFields = outputNames.some((name, index) => name && outputNames.indexOf(name) !== index);
  const unknownVariables = unknownOutputVariables(draft.outputFields, draft.fields);
  const validOutput = !duplicateOutputFields && outputNames.every(Boolean) && !unknownVariables.length;
  const valid = !!slug && !nameError && validInput && (draft.terminal || validOutput);
  useEffect(() => { document.getElementById(`${id}-name`)?.focus(); }, [id]);

  function updateField(key: string, update: Partial<ChoicePayloadField>) {
    setDraft(current => ({ ...current, fields: current.fields.map(field => field.key === key ? { ...field, ...update } : field) }));
  }

  return <aside className="graph-agent-panel graph-choice-panel nodrag nopan nowheel" role="dialog" aria-labelledby={`${id}-title`} onKeyDown={event => {
    if (event.key === 'Escape' && !event.defaultPrevented) { event.stopPropagation(); onClose(); }
  }}>
    <form className="graph-choice-form" onSubmit={event => {
      event.preventDefault();
      if (valid) onSave({ id: slug, name: normalizeChoiceName(draft.name, true), description: draft.description.trim(), terminal: draft.terminal, fields: draft.fields.map(field => ({ ...field, name: field.name.trim() })),
        sessionPolicy: !draft.terminal && draft.sessionPolicy === 'continue_target' ? 'continue_target' : undefined,
        outputFields: draft.terminal ? [] : draft.outputFields.map(field => ({ ...field, name: field.name.trim() })),
      });
    }}>
      <div className="graph-agent-panel-header"><h2 id={`${id}-title`}>Choice</h2><IconButton label="Close choice settings" onClick={onClose}><X /></IconButton></div>
      <div className="graph-agent-panel-field"><label htmlFor={`${id}-name`}>Name</label><Input id={`${id}-name`} required maxLength={80} pattern="[a-z0-9_]+" autoCapitalize="none" autoCorrect="off" spellCheck={false} value={draft.name} aria-invalid={!!nameError} aria-describedby={nameError ? `${id}-error` : undefined} onChange={event => setDraft(current => ({ ...current, name: normalizeChoiceName(event.target.value) }))} onBlur={() => setDraft(current => ({ ...current, name: normalizeChoiceName(current.name, true) }))} />{nameError && <p id={`${id}-error`} className="form-error" role="alert">{nameError}</p>}</div>
      <div className="graph-agent-panel-field"><label htmlFor={`${id}-description`}>Description</label><Input id={`${id}-description`} maxLength={240} value={draft.description} onChange={event => setDraft(current => ({ ...current, description: event.target.value }))} /></div>
      <RadioGroup label="Outcome" value={draft.terminal ? 'terminal' : 'continue'} options={[{ value: 'continue', label: 'Continue' }, { value: 'terminal', label: 'End graph' }]} onChange={value => setDraft(current => ({ ...current, terminal: value === 'terminal' }))} />
      <div className="graph-choice-payload">
        <div className="graph-agent-overrides-header"><h3>Input payload</h3><Button size="sm" variant="ghost" onClick={() => setDraft(current => ({ ...current, fields: [...current.fields, { key: crypto.randomUUID(), name: '', type: 'string', required: true }] }))}><Plus />Add field</Button></div>
        {draft.fields.map(field => <div key={field.key} className="graph-choice-field">
          <Input id={`${id}-${field.key}`} aria-label="Input field name" placeholder="Field name" required maxLength={80} value={field.name} onChange={event => updateField(field.key, { name: event.target.value })} />
          <Toggle label="Required" aria-label={`Required: ${field.name || 'payload field'}`} checked={field.required} onCheckedChange={required => updateField(field.key, { required })} />
          <IconButton label={`Remove payload field ${field.name || ''}`.trim()} onClick={() => setDraft(current => ({ ...current, fields: current.fields.filter(item => item.key !== field.key) }))}><Trash2 /></IconButton>
        </div>)}
        {duplicateFields && <p className="form-error" role="alert">Each payload field needs a unique name.</p>}
      </div>
      {!draft.terminal && <>
        <ChoiceOutputEditor fields={draft.outputFields} variables={choiceVariables(draft.fields)} onChange={outputFields => setDraft(current => ({ ...current, outputFields }))} />
        {duplicateOutputFields && <p className="form-error" role="alert">Each output field needs a unique name.</p>}
        {!!unknownVariables.length && <p className="form-error" role="alert">Unknown variable: {unknownVariables.map(variable => `{{${variable}}}`).join(', ')}</p>}
        {validInput && validOutput && !!draft.outputFields.length && <ChoiceOutputPreview key={JSON.stringify(draft.fields.map(field => [field.name.trim(), field.required]))} inputFields={draft.fields} outputFields={draft.outputFields} />}
        <RadioGroup label="Session policy" value={draft.sessionPolicy ?? 'new'} options={[{ value: 'new', label: 'New session' }, { value: 'continue_target', label: 'Continue target' }]} onChange={value => setDraft(current => ({ ...current, sessionPolicy: value === 'continue_target' ? 'continue_target' : undefined }))} />
      </>}
      <div className="graph-choice-form-actions"><Button onClick={onClose}>Cancel</Button><Button type="submit" variant="primary" disabled={!valid}>Apply</Button></div>
    </form>
  </aside>;
}
