import { ChevronDown, RefreshCw, Check } from 'lucide-react';
import { useEffect, useRef, useState, type KeyboardEvent } from 'react';
import { Button, IconButton } from '../../design-system/Button';
import { Menu, MenuItem } from '../../design-system/Menu';
import { Slider } from '../../design-system/Slider';
import { HarnessIcon } from './HarnessIcon';
import { request, type ModelCatalog, type Session } from '../../engine';

const effortOrder = ['none', 'off', 'minimal', 'low', 'medium', 'high', 'xhigh', 'max'];
function effortRank(value: string) {
  const index = effortOrder.indexOf(value.toLowerCase());
  return index < 0 ? effortOrder.length : index;
}
function effortLabel(value: string) {
  return value === 'xhigh' ? 'Extra high' : value ? value[0].toUpperCase() + value.slice(1) : 'Default';
}

export function ModelPicker({ session, disabled, onSave }: {
  session: Session; disabled: boolean; onSave: (model: string, effort: string) => Promise<boolean>;
}) {
  const [catalog, setCatalog] = useState<ModelCatalog | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [refresh, setRefresh] = useState(0);
  const [menu, setMenu] = useState<'model' | 'effort' | null>(null);
  const [search, setSearch] = useState('');
  const [previewEffort, setPreviewEffort] = useState('');
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState('');
  const pending = useRef(false);
  const loadedAt = useRef(0);
  const modelList = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let active = true;
    setLoading(true); setError('');
    request<ModelCatalog>(`/api/sessions/${encodeURIComponent(session.id)}/models${refresh ? '?refresh=true' : ''}`, 'GET', undefined, 55000)
      .then(value => { if (active) { setCatalog(value); loadedAt.current = Date.now(); } })
      .catch(error => { if (active) setError(error instanceof Error ? error.message : 'Could not load models.'); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [session.id, refresh]);

  const connected = new Set(catalog?.connectedProviders?.map(provider => provider.id) ?? []);
  const models = catalog?.models.filter(item => connected.has(item.provider)) ?? [];
  const current = models.find(item => item.id === (session.model || session.resolvedModel || catalog?.defaultModel));
  const currentEffort = session.effort || session.resolvedEffort || current?.defaultEffort || (!session.model ? catalog?.defaultEffort : '') || '';
  const efforts = [...(current?.efforts ?? [])].sort((a, b) => effortRank(a) - effortRank(b));
  const effortIndex = Math.max(0, efforts.indexOf(previewEffort));
  const filtered = models.filter(option => `${option.name} ${option.id} ${option.providerName}`.toLowerCase().includes(search.toLowerCase()));
  const locked = disabled || saving;

  function openMenu(next: 'model' | 'effort', open: boolean) {
    if (!open) { setMenu(current => current === next ? null : current); return; }
    setSaveError(''); setSearch(''); setPreviewEffort(currentEffort);
    if (!loading && Date.now() - loadedAt.current > 120000) setRefresh(value => value + 1);
    setMenu(next);
  }
  async function save(model: string, effort: string, close: boolean) {
    if (pending.current || disabled) return;
    if (model === (session.model ?? '') && effort === (session.effort ?? '')) { if (close) setMenu(null); return; }
    pending.current = true; setSaving(true); setSaveError('');
    try {
      if (await onSave(model, effort)) { if (close) setMenu(null); }
      else { setSaveError('Could not apply selection. Try again.'); setPreviewEffort(currentEffort); }
    } catch { setSaveError('Could not apply selection. Try again.'); setPreviewEffort(currentEffort); }
    finally { pending.current = false; setSaving(false); }
  }
  function chooseModel(id: string) {
    const next = models.find(option => option.id === (id || catalog?.defaultModel));
    const effort = next?.efforts.includes(session.effort ?? '') ? session.effort! : '';
    void save(id, effort, true);
  }
  function navigateModels(event: KeyboardEvent<HTMLElement>) {
    const options = [...(modelList.current?.querySelectorAll<HTMLButtonElement>('button:not(:disabled)') ?? [])];
    if (!options.length) return;
    const index = options.indexOf(document.activeElement as HTMLButtonElement);
    let next: number;
    if (event.key === 'ArrowDown') next = Math.min(index + 1, options.length - 1);
    else if (event.key === 'ArrowUp') next = index <= 0 ? options.length - 1 : index - 1;
    else if (index >= 0 && event.key === 'Home') next = 0;
    else if (index >= 0 && event.key === 'End') next = options.length - 1;
    else return;
    event.preventDefault(); options[next].focus(); options[next].scrollIntoView({ block: 'nearest' });
  }

  return <div className="model-picker-controls">
    <Menu label="Models" className="model-menu" open={menu === 'model'} onOpenChange={open => openMenu('model', open)} trigger={props => <Button {...props} variant="ghost" size="sm" className="model-picker-trigger" disabled={locked} aria-label="Choose model">
      <HarnessIcon harness={session.harness} /><span className="model-picker-name">{current?.name || session.model || session.resolvedModel || 'Harness default'}</span><ChevronDown />
    </Button>}>
      <div className="model-menu-search"><input data-menu-autofocus type="search" aria-label="Search models" placeholder="Search models" value={search} onChange={event => { setSearch(event.target.value); modelList.current?.scrollTo({ top: 0 }); }} onKeyDown={navigateModels} disabled={saving} /><IconButton label="Refresh models" disabled={loading || saving} onClick={() => setRefresh(value => value + 1)}><RefreshCw /></IconButton></div>
      {error && <p className="form-error menu-feedback" role="alert">{error}</p>}
      {loading ? <p className="menu-feedback muted" role="status">Loading models...</p> : <div ref={modelList} className="model-menu-list" onKeyDown={navigateModels} aria-label="Available models">
        {filtered.map(option => <MenuItem key={option.id} className="model-menu-option" selected={session.model === option.id} disabled={locked} title={option.id} onClick={() => chooseModel(option.id)}>
          <span><span className="model-menu-option-name">{option.name || option.id}</span><small>{option.providerName}</small></span>{session.model === option.id && <Check />}
        </MenuItem>)}
        {!filtered.length && <p className="menu-feedback muted">{search ? 'No matching models.' : 'No connected models available.'}</p>}
      </div>}
      <div className="model-menu-footer"><MenuItem selected={!session.model} disabled={locked || loading || !!error} onClick={() => chooseModel('')}>Harness default{!session.model && <Check />}</MenuItem></div>
      {saveError && <p className="form-error menu-feedback" role="alert">{saveError}</p>}
    </Menu>
    {!!efforts.length && <Menu label="Effort" className="effort-menu" open={menu === 'effort'} onOpenChange={open => openMenu('effort', open)} trigger={props => <Button {...props} variant="ghost" size="sm" className="model-picker-trigger effort-picker-trigger" disabled={locked || loading} aria-label="Choose effort">
      {effortLabel(currentEffort)}<ChevronDown />
    </Button>}>
      <Slider label="Effort" min={0} max={Math.max(0, efforts.length - 1)} value={effortIndex} valueText={effortLabel(efforts[effortIndex] ?? '')} marks={efforts.map((value, index) => ({ value: index, label: effortLabel(value) }))} disabled={locked || loading}
        onValueChange={value => setPreviewEffort(efforts[value])} onValueCommit={value => { const effort = efforts[value]; if (effort !== undefined) void save(session.model ?? '', effort, false); }} />
      {saving && <div className="effort-menu-footer"><span className="small muted" role="status">Saving...</span></div>}
      {saveError && <p className="form-error menu-feedback" role="alert">{saveError}</p>}
    </Menu>}
  </div>;
}
