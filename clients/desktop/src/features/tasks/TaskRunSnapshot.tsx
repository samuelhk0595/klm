import { ChevronDown } from 'lucide-react';
import { formatRunDate, harnessLabel, taskTriggerLabel, type TaskRun } from './types';

export function TaskRunSnapshot({ run, timezone }: { run: TaskRun; timezone: string }) {
  const settings = run.snapshot;
  return <details className="task-details-disclosure task-run-snapshot">
    <summary><ChevronDown />Captured settings</summary>
    <dl className="task-details-metadata">
      <div><dt>Content type</dt><dd>{run.input.contentType}</dd></div>
      <div><dt>Orchestrator</dt><dd>{harnessLabel[settings.harness]} · {settings.model}{settings.effort && ` · ${settings.effort}`}</dd></div>
      <div><dt>Run policy</dt><dd>{settings.allowParallelRuns ? 'Parallel' : 'FIFO'} · YOLO on</dd></div>
      <div><dt>Session folder at admission</dt><dd>{settings.sessionFolder}</dd></div>
      <div><dt>Questions</dt><dd>{settings.questions === 'telegram' ? 'KLM + Telegram' : 'KLM'}</dd></div>
      <div><dt>Configured trigger</dt><dd>{taskTriggerLabel(settings)}</dd></div>
      <div><dt>Accepted</dt><dd>{formatRunDate(run.requestedAt, timezone)}</dd></div>
      <div><dt>Started</dt><dd>{run.startedAt ? formatRunDate(run.startedAt, timezone) : 'Not started'}</dd></div>
      <div><dt>{run.endedAt ? 'Ended' : 'Observed'}</dt><dd>{formatRunDate(run.endedAt ?? run.observedAt, timezone)}</dd></div>
      <div><dt>Allowed graphs</dt><dd>{settings.allowedGraphIds.join(', ') || 'None'}</dd></div>
      {settings.trigger === 'schedule' && <div><dt>Missed-run recovery</dt><dd>{settings.recoverMissedRuns ? 'Latest missed firing only' : 'Off'}</dd></div>}
    </dl>
    <p className="task-instructions">{settings.instructions}</p>
    <details className="task-details-disclosure"><summary>Trigger payload</summary><pre className="task-instructions">{run.input.body || '(empty)'}</pre>{run.input.headers && Object.keys(run.input.headers).length > 0 && <pre className="task-instructions">{JSON.stringify(run.input.headers, null, 2)}</pre>}</details>
  </details>;
}
