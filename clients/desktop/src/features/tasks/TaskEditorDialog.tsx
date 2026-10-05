import { useEffect, useId, useRef, useState } from 'react';
import { Bot, CalendarClock, ClipboardList, MousePointer2, Webhook, X, Zap } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { Input } from '../../design-system/Input';
import { RadioGroup } from '../../design-system/RadioGroup';
import { Select } from '../../design-system/Select';
import { SearchSelect } from '../../design-system/SearchSelect';
import { Slider } from '../../design-system/Slider';
import { Textarea } from '../../design-system/Textarea';
import { Toggle } from '../../design-system/Toggle';
import { ToggleGroup } from '../../design-system/ToggleGroup';
import { TimePicker } from '../../design-system/TimePicker';
import { TaskGraphAllowlist } from './TaskGraphAllowlist';
import { getTaskModels, type ModelCatalog, type TaskDefinition } from '../../engine';
import { HarnessIcon } from '../chat/HarnessIcon';
import { defaultTaskDays, defaultTaskExecution, harnessLabel, telegramStatus, weekDays, type TaskDraft, type TaskTrigger, type TelegramLink } from './types';

export function TaskEditorDialog({ projectId, task, tasks, folders, telegram, onSave, onClose, onTelegramSettings }: {
  projectId: string;
  task?: TaskDefinition; tasks: TaskDefinition[]; folders: string[]; telegram: TelegramLink;
  onSave: (draft: TaskDraft, operationId: string) => Promise<void>; onClose: () => void; onTelegramSettings: () => void;
}) {
  const id = useId();
  const dialog = useRef<HTMLDialogElement>(null);
  useEffect(() => { dialog.current?.showModal(); }, []);
  const [draft, setDraft] = useState<TaskDraft>(() => task ? { ...task, days: [...task.days], allowedGraphIds: [...task.allowedGraphIds] } : { ...defaultTaskExecution, name: '', description: '', instructions: '', sessionFolder: 'Tasks', allowedGraphIds: [], enabled: true, trigger: 'manual', days: [...defaultTaskDays], time: '09:00', questions: 'klm' });
  const [section, setSection] = useState<'task' | 'orchestrator' | 'trigger'>('task');
  const [operationId] = useState(() => Array.from(crypto.getRandomValues(new Uint32Array(4)), value => value.toString(16).padStart(8, '0')).join(''));
  const [catalog, setCatalog] = useState<ModelCatalog | null>(null);
  const [modelLoading, setModelLoading] = useState(true);
  const [modelError, setModelError] = useState('');
  const [modelRefresh, setModelRefresh] = useState(0);
  const [saving, setSaving] = useState(false);
  const pending = useRef(false);
  const [saveError, setSaveError] = useState('');
  useEffect(() => {
    let active = true;
    setCatalog(null); setModelLoading(true); setModelError('');
    getTaskModels(projectId, draft.harness).then(value => { if (active) setCatalog(value); }, error => {
      if (active) setModelError(error instanceof Error ? error.message : 'Could not load models.');
    }).finally(() => { if (active) setModelLoading(false); });
    return () => { active = false; };
  }, [projectId, draft.harness, modelRefresh]);
  const sessionFolders = [...new Set(['Tasks', ...folders, draft.sessionFolder, 'Ungrouped'])];
  const duplicate = tasks.some(item => item.id !== task?.id && item.name.toLowerCase() === draft.name.trim().toLowerCase());
  const models = catalog?.models ?? [];
  const selectedModel = models.find(model => model.id === draft.model);
  const efforts = selectedModel?.efforts ?? [];
  const fieldsValid = !!draft.name.trim() && draft.name.trim().length <= 80 && !/[\u0000-\u001f\u007f-\u009f]/.test(draft.name) && draft.description.trim().length <= 240 && !!draft.instructions.trim() && draft.instructions.trim().length <= 12000 && !duplicate;
  const valid = fieldsValid && !modelLoading && !modelError && !!selectedModel && (efforts.length ? efforts.includes(draft.effort) : !draft.effort) && /^(?:[01]\d|2[0-3]):[0-5]\d$/.test(draft.time) && (draft.trigger !== 'schedule' || draft.days.length > 0);
  const effortLabel = (value: string) => value === 'xhigh' ? 'Extra high' : value ? value[0].toUpperCase() + value.slice(1) : '';

  const close = () => { if (!pending.current) onClose(); };
  return <dialog ref={dialog} className="settings-dialog task-editor-dialog" aria-labelledby={`${id}-title`} onCancel={event => { event.preventDefault(); close(); }} onClose={close}>
    <form className="task-editor" noValidate onSubmit={async event => {
      event.preventDefault();
      if (pending.current) return;
      if (!valid) {
        setSection(!fieldsValid ? 'task' : !selectedModel || modelError || efforts.length > 0 && !efforts.includes(draft.effort) ? 'orchestrator' : 'trigger');
        return;
      }
      pending.current = true; setSaving(true); setSaveError('');
      try { await onSave({ ...draft, enabled: task ? draft.enabled : true, name: draft.name.trim(), description: draft.description.trim(), instructions: draft.instructions.trim() }, operationId); }
      catch (error) { setSaveError(error instanceof Error ? error.message : 'Could not save task.'); }
      finally { pending.current = false; setSaving(false); }
    }}>
      <div className="task-editor-heading"><h2 id={`${id}-title`}>{task ? 'Edit task' : 'New task'}</h2><IconButton label="Close task editor" disabled={saving} onClick={close}><X /></IconButton></div>
      <div className="task-editor-layout" inert={saving}>
        <nav className="settings-sidebar task-editor-sidebar" aria-label="Task editor sections">
          <Button variant="ghost" className={section === 'task' ? 'is-active' : ''} aria-current={section === 'task' ? 'page' : undefined} aria-controls={`${id}-task`} onClick={() => setSection('task')}><ClipboardList />Task</Button>
          <Button variant="ghost" className={section === 'orchestrator' ? 'is-active' : ''} aria-current={section === 'orchestrator' ? 'page' : undefined} aria-controls={`${id}-orchestrator`} onClick={() => setSection('orchestrator')}><Bot />Orchestrator</Button>
          <Button variant="ghost" className={section === 'trigger' ? 'is-active' : ''} aria-current={section === 'trigger' ? 'page' : undefined} aria-controls={`${id}-trigger`} onClick={() => setSection('trigger')}><Zap />Trigger</Button>
        </nav>
        <div className="task-editor-content">
          <section id={`${id}-task`} className="task-editor-section" aria-label="Task" hidden={section !== 'task'}>
            <div className="task-field"><label htmlFor={`${id}-name`}>Name</label><Input id={`${id}-name`} autoFocus required maxLength={80} value={draft.name} aria-invalid={duplicate} aria-describedby={duplicate ? `${id}-name-error` : undefined} onChange={event => setDraft(current => ({ ...current, name: event.target.value }))} />{duplicate && <p id={`${id}-name-error`} className="form-error" role="alert">A task with this name already exists.</p>}</div>
            <div className="task-field"><label htmlFor={`${id}-description`}>Description</label><Input id={`${id}-description`} placeholder="Optional" maxLength={240} value={draft.description} onChange={event => setDraft(current => ({ ...current, description: event.target.value }))} /></div>
            <div className="task-field"><label htmlFor={`${id}-folder`}>Session folder</label><Select id={`${id}-folder`} label="Session folder" variant="field" value={draft.sessionFolder} onChange={event => setDraft(current => ({ ...current, sessionFolder: event.target.value }))}>{sessionFolders.map(folder => <option key={folder} value={folder}>{folder}</option>)}</Select></div>
            <div className="task-field"><label htmlFor={`${id}-instructions`}>Instructions</label><Textarea id={`${id}-instructions`} required rows={5} maxLength={12000} placeholder="What should this task accomplish?" value={draft.instructions} onChange={event => setDraft(current => ({ ...current, instructions: event.target.value }))} /></div>
            <TaskGraphAllowlist projectId={projectId} value={draft.allowedGraphIds} onChange={allowedGraphIds => setDraft(current => ({ ...current, allowedGraphIds }))} />
            {draft.allowedGraphIds.length > 0 && <p className="task-save-authorization">Saving authorizes the selected Allowed graphs to fulfill these instructions.</p>}
          </section>
          <section id={`${id}-orchestrator`} className="task-editor-section" aria-label="Orchestrator" hidden={section !== 'orchestrator'}>
            <RadioGroup<TaskDraft['harness']> label="Harness" value={draft.harness} options={(['opencode', 'codex', 'pi'] as const).map(harness => ({ value: harness, label: harnessLabel[harness], icon: <HarnessIcon harness={harness} /> }))} onChange={harness => setDraft(current => ({ ...current, harness, model: '', effort: '' }))} />
            <div className="task-field"><label htmlFor={`${id}-model`}>Model</label><SearchSelect key={draft.harness} id={`${id}-model`} label="Model" value={draft.model} options={[...models.map(model => ({ value: model.id, label: model.name, description: model.providerName || model.provider })), ...(draft.model && !selectedModel ? [{ value: draft.model, label: draft.model, description: modelLoading ? 'Loading...' : 'Unavailable' }] : [])]} searchPlaceholder="Search models" onChange={id => {
              const model = models.find(item => item.id === id);
              if (model) setDraft(current => ({ ...current, model: id, effort: model.efforts.includes(current.effort) ? current.effort : model.defaultEffort && model.efforts.includes(model.defaultEffort) ? model.defaultEffort : model.efforts[0] ?? '' }));
            }} /></div>
            {modelLoading && <p role="status">Loading models...</p>}
            {modelError && <p className="form-error" role="alert">{modelError}<Button size="sm" onClick={() => setModelRefresh(current => current + 1)}>Retry</Button></p>}
            {!modelLoading && !modelError && draft.model && !selectedModel && <p className="form-error" role="alert">The saved model is unavailable. Select an available model.</p>}
            {!!efforts.length && <Slider label="Effort level" min={0} max={efforts.length - 1} value={Math.max(0, efforts.indexOf(draft.effort))} valueText={effortLabel(draft.effort)} marks={efforts.map((effort, index) => ({ value: index, label: effortLabel(effort) }))} onValueChange={index => setDraft(current => ({ ...current, effort: efforts[index] }))} />}
            <section className="task-policy task-telegram-policy" aria-labelledby={`${id}-telegram-title`}>
              <div className="task-policy-heading"><h3 id={`${id}-telegram-title`}>Telegram routing</h3><Toggle label="" aria-label="Route questions to Telegram" aria-describedby={`${id}-telegram-description`} checked={draft.questions === 'telegram'} onCheckedChange={enabled => setDraft(current => ({ ...current, questions: enabled ? 'telegram' : 'klm' }))} /></div>
              <p id={`${id}-telegram-description`}>Send questions to Telegram while keeping them available in KLM.</p>
              {draft.questions === 'telegram' && <div className="task-telegram-link"><span>{telegramStatus(telegram)}</span><Button size="sm" onClick={onTelegramSettings}>{telegram.account ? 'Manage link' : 'Link Telegram'}</Button></div>}
            </section>
          </section>
          <section id={`${id}-trigger`} className="task-editor-section" aria-label="Trigger" hidden={section !== 'trigger'}>
            <RadioGroup<TaskTrigger> label="Trigger" value={draft.trigger} onChange={trigger => setDraft(current => ({ ...current, trigger }))} options={[{ value: 'manual', label: 'Manual', icon: <MousePointer2 /> }, { value: 'schedule', label: 'Schedule', icon: <CalendarClock /> }, { value: 'webhook', label: 'Webhook', icon: <Webhook /> }]} />
            {draft.trigger === 'webhook' && <section className="task-policy"><h3>Webhook</h3><p>Get the endpoint URL and secret in Task details after saving.</p></section>}
            {draft.trigger === 'schedule' && <div className="task-schedule">
              <fieldset className="task-days" aria-describedby={!draft.days.length ? `${id}-days-error` : undefined}>
                <legend>Run on</legend>
                <ToggleGroup label="Run on" value={draft.days} options={weekDays.map(day => ({ value: day.value, label: day.short, accessibleLabel: day.label }))} onValueChange={days => setDraft(current => ({ ...current, days }))} describedBy={!draft.days.length ? `${id}-days-error` : undefined} />
                {!draft.days.length && <p id={`${id}-days-error`} className="form-error" role="alert">Select at least one day.</p>}
              </fieldset>
              <div className="task-field"><label htmlFor={`${id}-time`}>Time</label><TimePicker id={`${id}-time`} label="Time" value={draft.time} onValueChange={time => setDraft(current => ({ ...current, time }))} /></div>
              <div className="task-policy task-schedule-recovery"><Toggle label="Recover missed runs" checked={draft.recoverMissedRuns} onCheckedChange={recoverMissedRuns => setDraft(current => ({ ...current, recoverMissedRuns }))} /><p>Recover only the latest missed firing.</p></div>
            </div>}
            <section className="task-policy task-parallel-policy" aria-labelledby={`${id}-parallel-title`}>
              <div className="task-policy-heading"><h3 id={`${id}-parallel-title`}>Parallel runs</h3><Toggle label="" aria-label="Allow parallel runs" aria-describedby={`${id}-parallel-description`} checked={draft.allowParallelRuns} onCheckedChange={allowParallelRuns => setDraft(current => ({ ...current, allowParallelRuns }))} /></div>
              <p id={`${id}-parallel-description`}>Allow multiple runs of this task at the same time.</p>
            </section>
          </section>
        </div>
      </div>
      {saveError && <p className="form-error" role="alert">{saveError}</p>}
      <div className="task-editor-footer">{task && <Toggle label="Enabled" disabled={saving} checked={draft.enabled} onCheckedChange={enabled => setDraft(current => ({ ...current, enabled }))} />}<div className="dialog-actions"><Button disabled={saving} onClick={close}>Cancel</Button><Button type="submit" variant="primary" disabled={!valid || saving}>{saving ? 'Saving...' : task ? 'Save changes' : 'Create task'}</Button></div></div>
    </form>
  </dialog>;
}
