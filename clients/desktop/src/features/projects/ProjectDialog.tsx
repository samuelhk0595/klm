import { useEffect, useRef, useState } from 'react';
import { ImagePlus, X } from 'lucide-react';
import { Button, IconButton } from '../../design-system/Button';
import { pickDirectory, type Project } from '../../engine';

export function ProjectDialog({ initialFolder, project, onSave, onClose }: {
  initialFolder: string; project?: Project;
  onSave: (project: Omit<Project, 'id' | 'folders'>) => Promise<string | null>; onClose: () => void;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const [name, setName] = useState(project?.name ?? initialFolder.split(/[\\/]/).filter(Boolean).pop() ?? '');
  const [folder, setFolder] = useState(initialFolder);
  const [icon, setIcon] = useState(project?.icon ?? '');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  useEffect(() => { dialog.current?.showModal(); }, []);

  async function chooseIcon(file: File | undefined) {
    if (!file) return;
    setError('');
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type) || file.size > 5 * 1024 * 1024) {
      setError('Choose a PNG, JPEG, or WebP image smaller than 5 MB.'); return;
    }
    setBusy(true);
    const url = URL.createObjectURL(file);
    try {
      const image = new Image(); image.src = url; await image.decode();
      const canvas = document.createElement('canvas'); canvas.width = 96; canvas.height = 96;
      const context = canvas.getContext('2d');
      if (!context) throw new Error('Image processing unavailable');
      // Crop to a small square so stored project icons do not consume the storage quota.
      const size = Math.min(image.naturalWidth, image.naturalHeight);
      context.drawImage(image, (image.naturalWidth - size) / 2, (image.naturalHeight - size) / 2, size, size, 0, 0, 96, 96);
      setIcon(canvas.toDataURL('image/png'));
    } catch { setError('This image could not be opened. Please choose another image.'); }
    finally { URL.revokeObjectURL(url); setBusy(false); }
  }

  return <dialog ref={dialog} className="settings-dialog add-project-dialog" aria-labelledby="project-dialog-title" onCancel={event => { if (busy) event.preventDefault(); else onClose(); }} onClose={onClose}>
    <form onSubmit={async event => {
      event.preventDefault();
      if (busy || !name.trim() || !folder.trim()) return;
      setBusy(true); setError('');
      try {
        const failure = await onSave({ name: name.trim(), folder: folder.trim(), icon });
        if (failure) setError(failure); else onClose();
      } catch (error) { setError(error instanceof Error ? error.message : 'Could not save the project.'); }
      finally { setBusy(false); }
    }}>
      <div className="dialog-heading"><h2 id="project-dialog-title">{project ? 'Edit project' : 'Add project'}</h2><IconButton label={project ? 'Cancel editing project' : 'Cancel adding project'} disabled={busy} onClick={onClose}><X /></IconButton></div>
      <label className="project-field" htmlFor="project-picture">Project icon</label>
      <div className="project-picture-picker project-icon-trigger">
        <div className="project-picture-preview">{icon ? <img src={icon} alt="" /> : <ImagePlus aria-hidden="true" />}</div>
        <span className="button button--outline button--md">{icon ? 'Change icon' : 'Choose icon'}</span>
        <input id="project-picture" type="file" className="native-picker-overlay" aria-label={icon ? 'Change project icon' : 'Choose project icon'} accept="image/png,image/jpeg,image/webp" disabled={busy} onChange={event => { void chooseIcon(event.target.files?.[0]); event.target.value = ''; }} />
      </div>
      <label className="project-field" htmlFor="project-name">Project name</label>
      <input id="project-name" required maxLength={60} disabled={busy} value={name} onChange={event => setName(event.target.value)} placeholder="e.g. My application" />
      {!project && <><label className="project-field" htmlFor="project-directory">Project directory</label>
      <div className="folder-picker">
        <input id="project-directory" readOnly value={folder} title={folder} />
        <Button disabled={busy} onClick={async () => {
          setBusy(true); setError('');
          try { const selected = await pickDirectory(); if (selected) setFolder(selected); }
          catch (error) { setError(error instanceof Error ? error.message : 'Could not open the directory picker.'); }
          finally { setBusy(false); }
        }}>Change directory</Button>
      </div></>}
      {error && <p role="alert" className="form-error">{error}</p>}
      <div className="dialog-actions"><Button disabled={busy} onClick={onClose}>Cancel</Button><Button type="submit" variant="primary" disabled={busy || !folder.trim() || !name.trim()}>{busy ? 'Please wait...' : project ? 'Save changes' : 'Create project'}</Button></div>
    </form>
  </dialog>;
}
