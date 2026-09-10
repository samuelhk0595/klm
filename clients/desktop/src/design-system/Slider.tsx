import { useId, type CSSProperties } from 'react';

const adjustmentKeys = new Set(['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End', 'PageUp', 'PageDown']);

export function Slider({ label, value, min = 0, max = 100, step = 1, valueText, marks, disabled = false, onValueChange, onValueCommit }: {
  label: string;
  value: number;
  min?: number;
  max?: number;
  step?: number;
  valueText?: string;
  marks?: { value: number; label: string }[];
  disabled?: boolean;
  onValueChange: (value: number) => void;
  onValueCommit?: (value: number) => void;
}) {
  const id = useId();
  const position = (markValue: number) => max > min ? Math.max(0, Math.min(100, (markValue - min) / (max - min) * 100)) : 50;
  const progress = position(value);
  return <div className="slider" data-disabled={disabled || undefined}>
    <div className="slider-heading"><label htmlFor={id}>{label}</label><output htmlFor={id}>{valueText ?? value}</output></div>
    <div className="slider-scale" data-marked={!!marks?.length || undefined}>
    <input id={id} type="range" className="slider-input" min={min} max={max} step={step} value={value} aria-valuetext={valueText} disabled={disabled || max <= min}
      style={{ '--slider-progress': `calc(${progress}% + var(--slider-thumb-size) * ${0.5 - progress / 100})` } as CSSProperties}
      onChange={event => onValueChange(event.currentTarget.valueAsNumber)}
      onPointerDown={event => event.currentTarget.setPointerCapture(event.pointerId)}
      onPointerUp={event => onValueCommit?.(event.currentTarget.valueAsNumber)}
      onKeyUp={event => { if (adjustmentKeys.has(event.key)) onValueCommit?.(event.currentTarget.valueAsNumber); }} />
    {!!marks?.length && <div className="slider-marks" aria-hidden="true">{marks.map(mark => <span key={mark.value} title={mark.label} style={{ left: `${position(mark.value)}%` }} className={mark.value === value ? 'is-active' : undefined}>{mark.label}</span>)}</div>}
    </div>
  </div>;
}
