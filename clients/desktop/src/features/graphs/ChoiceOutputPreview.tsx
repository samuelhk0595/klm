import { useId, useState } from 'react';
import { Input } from '../../design-system/Input';
import { Textarea } from '../../design-system/Textarea';
import { previewChoiceOutput, type ChoiceOutputField, type ChoicePayloadField } from './choice';

export function ChoiceOutputPreview({ inputFields, outputFields }: { inputFields: ChoicePayloadField[]; outputFields: ChoiceOutputField[] }) {
  const id = useId();
  const [task, setTask] = useState('Add CSV export for orders.');
  const [inputText, setInputText] = useState(() => JSON.stringify(Object.fromEntries(inputFields.filter(field => field.required).map(field => [field.name.trim(), `Example ${field.name.trim()}`])), null, 2));
  let result: ReturnType<typeof previewChoiceOutput>;
  try { result = previewChoiceOutput(inputFields, outputFields, JSON.parse(inputText), task); }
  catch { result = { output: null, error: 'Enter valid JSON for input values.' }; }

  return <details className="graph-choice-output-preview">
    <summary>Preview JSON</summary>
    <div className="graph-choice-preview-content">
      <div className="graph-agent-panel-field"><label htmlFor={`${id}-task`}>Task</label><Input id={`${id}-task`} value={task} onChange={event => setTask(event.target.value)} /></div>
      <div className="graph-agent-panel-field"><label htmlFor={`${id}-input`}>Input values</label><Textarea id={`${id}-input`} className="graph-choice-json-input" rows={4} value={inputText} spellCheck={false} onChange={event => setInputText(event.target.value)} /></div>
      {result.error ? <p className="form-error" role="status">{result.error}</p> : <pre className="graph-choice-json-preview" aria-label="Output JSON"><code>{JSON.stringify(result.output, null, 2)}</code></pre>}
    </div>
  </details>;
}
