import { useEffect, useId, useRef, useState } from 'react';
import { X } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { RadioGroup } from '../../design-system/RadioGroup';
import { SearchSelect } from '../../design-system/SearchSelect';
import { Slider } from '../../design-system/Slider';
import { Textarea } from '../../design-system/Textarea';
import { HarnessIcon } from '../chat/HarnessIcon';
import type { Harness } from '../../engine';
import { agentSlug, initialAgents, modelsForHarness, mockAgentSettings, type Agent } from './demo';

type AgentDraft = {
  name: string;
  description: string;
  defaultHarness: Harness['id'];
  model: string;
  effort: string;
  prompt: string;
};

function effortLabel(value: string) {
  return value === 'xhigh' ? 'Extra high' : value ? value[0].toUpperCase() + value.slice(1) : '';
}

export function AgentEditorDialog({ agent, agents, onSave, onClose }: {
  agent?: Agent;
  agents: Agent[];
  onSave: (draft: AgentDraft) => void;
  onClose: () => void;
}) {
  const id = useId();
  const dialog = useRef<HTMLDialogElement>(null);
  useEffect(() => { dialog.current?.showModal(); }, []);
  const [draft, setDraft] = useState<Omit<AgentDraft, 'defaultHarness'> & { defaultHarness?: Harness['id'] }>(() => ({
    name: agent?.name ?? '',
    description: agent?.description ?? '',
    defaultHarness: agent ? agent.defaultHarness ?? mockAgentSettings.defaultHarness : undefined,
    model: agent ? agent.model || mockAgentSettings.model : '',
    effort: agent ? agent.effort || mockAgentSettings.effort : '',
    prompt: agent?.prompt ?? initialAgents.find(item => item.id === agent?.id)?.prompt ?? '',
  }));
  const slug = agentSlug(draft.name);
  const duplicate = !!slug && agents.some(item => item.id !== agent?.id && (item.id === slug || agentSlug(item.name) === slug));
  const nameError = duplicate ? 'An agent with this identifier already exists.' : draft.name.trim() && !slug ? 'Use at least one letter or number in the name.' : '';
  const models = draft.defaultHarness ? modelsForHarness(draft.defaultHarness) : [];
  const selectedModel = models.find(model => model.name === draft.model);
  const efforts = selectedModel?.efforts ?? [];
  const valid = !!slug && !nameError && !!draft.prompt.trim() && !!selectedModel && efforts.includes(draft.effort);

  function selectModel(modelName: string) {
    const model = models.find(item => item.name === modelName);
    if (!model) return;
    setDraft(current => ({ ...current, model: model.name,
      effort: model.efforts.includes(current.effort) ? current.effort : model.efforts.includes('high') ? 'high' : model.efforts[0],
    }));
  }

  return <dialog ref={dialog} className="settings-dialog agent-editor-dialog" aria-labelledby={`${id}-title`} onCancel={onClose} onClose={onClose}>
    <form className="agent-editor" onSubmit={event => {
      event.preventDefault();
      if (valid && draft.defaultHarness) onSave({ ...draft, defaultHarness: draft.defaultHarness, name: draft.name.trim(), description: draft.description.trim() });
    }}>
      <div className="agent-editor-header">
        <h2 id={`${id}-title`}>{agent ? 'Edit agent' : 'New agent'}</h2>
        <IconButton label="Close agent editor" onClick={onClose}><X /></IconButton>
      </div>
      <div className="agent-editor-grid">
        <div className="agent-editor-field"><label htmlFor={`${id}-name`}>Name</label><Input id={`${id}-name`} autoFocus required maxLength={80} value={draft.name} aria-invalid={!!nameError} aria-describedby={nameError ? `${id}-name-error` : undefined} onChange={event => setDraft(current => ({ ...current, name: event.target.value }))} />{nameError && <p id={`${id}-name-error`} className="form-error" role="alert">{nameError}</p>}</div>
        <div className="agent-editor-field"><label htmlFor={`${id}-slug`}>Identifier</label><Input id={`${id}-slug`} readOnly value={slug} /></div>
        <div className="agent-editor-field agent-editor-wide"><label htmlFor={`${id}-description`}>Description</label><Input id={`${id}-description`} maxLength={240} value={draft.description} onChange={event => setDraft(current => ({ ...current, description: event.target.value }))} /></div>
        <div className="agent-editor-wide"><RadioGroup<Harness['id']> label="Default harness" value={draft.defaultHarness} options={[
          { value: 'opencode', label: 'OpenCode', icon: <HarnessIcon harness="opencode" /> },
          { value: 'codex', label: 'Codex', icon: <HarnessIcon harness="codex" /> },
          { value: 'pi', label: 'Pi', icon: <HarnessIcon harness="pi" /> },
        ]} onChange={harness => setDraft(current => {
          const model = modelsForHarness(harness).find(item => item.name === current.model);
          return { ...current, defaultHarness: harness, model: model?.name ?? '', effort: model ? current.effort : '' };
        })} /></div>
        {draft.defaultHarness && <div className="agent-editor-field agent-editor-wide"><label htmlFor={`${id}-model`}>Model</label><SearchSelect key={draft.defaultHarness} id={`${id}-model`} label="Model" value={draft.model} options={models.map(model => ({ value: model.name, label: model.name, description: model.provider }))} onChange={selectModel} placeholder="Select model" searchPlaceholder="Search models" emptyMessage={models.length ? 'No models found' : 'No models available from connected providers'} /></div>}
        {!!efforts.length && <div className="agent-editor-wide"><Slider label="Effort level" min={0} max={Math.max(0, efforts.length - 1)} value={Math.max(0, efforts.indexOf(draft.effort))} valueText={effortLabel(draft.effort)} marks={efforts.map((effort, index) => ({ value: index, label: effortLabel(effort) }))} onValueChange={index => setDraft(current => ({ ...current, effort: efforts[index] }))} /></div>}
        <div className="agent-editor-field agent-editor-wide"><label htmlFor={`${id}-prompt`}>Prompt</label><Textarea id={`${id}-prompt`} required rows={6} value={draft.prompt} onChange={event => setDraft(current => ({ ...current, prompt: event.target.value }))} /></div>
      </div>
      <div className="agent-editor-actions"><Button onClick={onClose}>Cancel</Button><Button type="submit" variant="primary" disabled={!valid}>{agent ? 'Save changes' : 'Create agent'}</Button></div>
    </form>
  </dialog>;
}
