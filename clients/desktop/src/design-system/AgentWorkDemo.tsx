import {
  CircleEllipsis,
  Database,
  FileCode2,
  FilePenLine,
  RotateCcw,
  Search,
  ShieldCheck,
  SquareTerminal,
  type LucideIcon,
} from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { AgentWork, type AgentWorkTone } from './AgentWork';
import { Button } from './Button';
import { Icon } from './Icon';

type DemoActivity = {
  completeAt: number;
  detail: string;
  diff?: { added: number; removed: number };
  icon: LucideIcon;
  label: string;
  startAt: number;
  tone: AgentWorkTone;
};

type DemoSegment = {
  activities: DemoActivity[];
  collapseAt: number;
  compactAt: number;
  response: string;
  responseEnd: number;
  responseStart: number;
  thoughtDuration: string;
  thoughts: Array<{ at: number; text: string }>;
  workDuration: string;
  workStart: number;
};

const DEMO_DURATION = 10_100;

// The source collapses each work block on the same frame its final activity completes.
const segments: DemoSegment[] = [
  {
    workStart: 0,
    compactAt: 2_000,
    collapseAt: 3_150,
    responseStart: 3_180,
    responseEnd: 3_650,
    thoughtDuration: '9.8s',
    workDuration: '19 seconds',
    thoughts: [
      { at: 350, text: 'Auditing authentication middleware AST for algorithm injection vulnerabilities and permissive defaults.' },
      { at: 700, text: 'Inspecting the global Express auth middleware call tree to locate the jwt.verify invocation.' },
      { at: 1_050, text: "The current verify implementation accepts undefined algorithms, exposing an algorithm confusion vulnerability ('none' cipher exploit)." },
      { at: 1_350, text: 'Tracing where process.env.JWT_SECRET is sourced to guarantee the cryptographic key is never exposed via response headers.' },
      { at: 1_620, text: 'Analyzing the token expiration timestamp format to prevent UNIX 2038 integer wrap-around errors.' },
      { at: 1_850, text: 'Formulating a strict algorithm schema requiring HS256 algorithm pinning across all protected API routes.' },
    ],
    activities: [
      { startAt: 2_000, completeAt: 2_350, label: 'Security Audit', detail: 'AST Check (3 files)', icon: ShieldCheck, tone: 'danger' },
      { startAt: 2_350, completeAt: 2_700, label: 'Read AST tree', detail: 'src/types/auth.d.ts', icon: FileCode2, tone: 'muted' },
      { startAt: 2_700, completeAt: 3_150, label: 'Search mitigation specs', detail: 'RFC 7519 algorithm confusion', icon: Search, tone: 'muted' },
    ],
    response: "I completed the security audit. Found an unpinned algorithm vulnerability in 'src/middleware/auth.ts' that could allow unsigned tokens. Beginning the patch and LRU revocation cache implementation now.",
  },
  {
    workStart: 3_700,
    compactAt: 5_650,
    collapseAt: 6_500,
    responseStart: 6_520,
    responseEnd: 6_960,
    thoughtDuration: '10.4s',
    workDuration: '17 seconds',
    thoughts: [
      { at: 3_950, text: 'Designing an in-memory sliding LRU cache with an automatic 15-minute TTL to track revoked token fingerprints.' },
      { at: 4_300, text: 'Ensuring thread safety and race-condition immunity during concurrent cache writes during cluster renewal handshakes.' },
      { at: 4_650, text: 'Adding a 10-second clock tolerance window (clockTolerance: 10) to guard against minor NTP drift between edge regions.' },
      { at: 5_000, text: 'Refactoring the jwt.verify wrapper to reject any payload lacking valid issuer and audience claim signatures.' },
      { at: 5_350, text: 'Injecting comprehensive TypeScript assertions to validate the patched session envelope.' },
    ],
    activities: [
      { startAt: 5_650, completeAt: 5_930, label: 'Patch file', detail: 'src/auth/jwt-verifier.ts', diff: { added: 3, removed: 1 }, icon: FilePenLine, tone: 'warning' },
      { startAt: 5_930, completeAt: 6_220, label: 'Inspect Schema', detail: 'sessions.revoked_tokens', icon: Database, tone: 'success' },
      { startAt: 6_220, completeAt: 6_500, label: 'Execute tests', detail: 'npm run test:auth -- --coverage', icon: SquareTerminal, tone: 'accent' },
    ],
    response: 'Security patch applied and verified with clean TypeScript compilation and 100% test coverage. Moving to adversarial regression testing and cluster canary deployment.',
  },
  {
    workStart: 7_000,
    compactAt: 8_500,
    collapseAt: 9_450,
    responseStart: 9_470,
    responseEnd: 10_050,
    thoughtDuration: '8.2s',
    workDuration: '17 seconds',
    thoughts: [
      { at: 7_180, text: 'Synthesizing adversarial signature matrix and staging canary environment.' },
      { at: 7_500, text: "Generating forged JWTs signed with algorithm 'none', asymmetric RSA keys, and corrupted header signatures." },
      { at: 7_820, text: 'Verifying that all malformed and expired tokens return an unambiguous HTTP 401 Unauthorized without stack traces.' },
      { at: 8_100, text: 'Benchmarking token verification latency under synthetic 5,000 req/s load to confirm p99 stays below 12ms.' },
      { at: 8_350, text: 'Running the canary health probe across staging-cluster-01 before promoting the configuration.' },
    ],
    activities: [
      { startAt: 8_500, completeAt: 8_800, label: 'Add test matrix', detail: 'tests/auth/adversarial.test.ts', diff: { added: 6, removed: 1 }, icon: SquareTerminal, tone: 'accent' },
      { startAt: 8_800, completeAt: 9_100, label: 'Run canary probe', detail: 'kubectl rollout status deployment/auth-service', icon: SquareTerminal, tone: 'accent' },
      { startAt: 9_100, completeAt: 9_450, label: 'Deploy Canary', detail: 'staging-cluster-01', icon: CircleEllipsis, tone: 'muted' },
    ],
    response: 'All 28 regression tests passed with zero regressions. The JWT auth middleware now pins HS256, validates revocation via the LRU cache, and successfully passed canary verification.',
  },
];

function WorkBlock({ elapsed, segment }: { elapsed: number; segment: DemoSegment }) {
  const active = elapsed < segment.collapseAt;
  const thoughtComplete = elapsed >= segment.compactAt;
  const visibleThoughts = segment.thoughts.filter(({ at }) => elapsed >= at);
  const visibleActivities = segment.activities.filter(({ startAt }) => elapsed >= startAt);

  return (
    <AgentWork
      activities={thoughtComplete ? visibleActivities.map(activity => ({
        id: activity.label,
        icon: activity.icon,
        label: activity.label,
        detail: activity.detail,
        diff: activity.diff,
        status: elapsed >= activity.completeAt ? 'completed' : 'running',
        tone: activity.tone,
      })) : []}
      durationLabel={segment.workDuration}
      status={active ? 'running' : 'completed'}
      streamedResponse={streamedText(segment, elapsed)}
      thoughts={[{
        id: `thought-${segment.workStart}`,
        status: thoughtComplete ? 'completed' : 'running',
        text: visibleThoughts.map(({ text }) => text).join('\n\n'),
        durationLabel: segment.thoughtDuration,
      }]}
    />
  );
}

function streamedText(segment: DemoSegment, elapsed: number): string {
  if (elapsed < segment.responseStart) return '';
  const progress = Math.min(1, (elapsed - segment.responseStart) / (segment.responseEnd - segment.responseStart));
  return segment.response.slice(0, Math.ceil(segment.response.length * progress));
}

export function agentWorkDemoStatus(elapsed: number): string {
  if (elapsed >= DEMO_DURATION) return 'Agent work demonstration complete.';
  let segment: DemoSegment | undefined;
  for (const candidate of segments) {
    if (elapsed < candidate.workStart) break;
    segment = candidate;
  }
  if (!segment) return 'Agent work demonstration ready.';
  if (elapsed >= segment.collapseAt) return 'Agent completed a work cycle and is responding.';
  if (elapsed < segment.compactAt) return 'Agent is thinking.';
  let activity: DemoActivity | undefined;
  for (const candidate of segment.activities) {
    if (elapsed < candidate.startAt) break;
    activity = candidate;
  }
  if (!activity) return 'Agent finished thinking.';
  return elapsed >= activity.completeAt ? `${activity.label} completed.` : `Agent is running ${activity.label}.`;
}

export function AgentWorkDemo() {
  const rootRef = useRef<HTMLDivElement>(null);
  const [reducedMotion, setReducedMotion] = useState(() => typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches);
  const [elapsed, setElapsed] = useState(reducedMotion ? DEMO_DURATION : 0);
  const [playVersion, setPlayVersion] = useState(0);

  useEffect(() => {
    const query = window.matchMedia('(prefers-reduced-motion: reduce)');
    const handleChange = ({ matches }: MediaQueryListEvent) => {
      setReducedMotion(matches);
      if (matches) setElapsed(DEMO_DURATION);
    };
    query.addEventListener('change', handleChange);
    return () => query.removeEventListener('change', handleChange);
  }, []);

  useEffect(() => {
    const root = rootRef.current;
    if (!root) return;
    if (!('IntersectionObserver' in window)) {
      setPlayVersion(1);
      return;
    }
    const observer = new IntersectionObserver(([entry]) => {
      if (!entry?.isIntersecting) return;
      setPlayVersion(1);
      observer.disconnect();
    }, { threshold: 0.35 });
    observer.observe(root);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (playVersion === 0) return;
    if (reducedMotion) {
      setElapsed(DEMO_DURATION);
      return;
    }
    const startedAt = performance.now();
    let animationFrame = 0;
    let lastPaint = 0;
    const tick = (now: number) => {
      const nextElapsed = Math.min(DEMO_DURATION, now - startedAt);
      if (now - lastPaint >= 32 || nextElapsed === DEMO_DURATION) {
        setElapsed(nextElapsed);
        lastPaint = now;
      }
      if (nextElapsed < DEMO_DURATION) animationFrame = requestAnimationFrame(tick);
    };
    animationFrame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(animationFrame);
  }, [playVersion, reducedMotion]);

  const replay = () => {
    if (reducedMotion) return;
    setElapsed(0);
    setPlayVersion(version => version + 1);
  };

  return (
    <div className="agent-work-demo" ref={rootRef}>
      <div className="agent-work-demo__toolbar">
        <span>10.1 second scripted specimen</span>
        <Button disabled={reducedMotion} onClick={replay} size="sm" variant="ghost" title={reducedMotion ? 'Replay is disabled by your reduced-motion preference' : undefined}><Icon glyph={RotateCcw} size={12} />Replay</Button>
      </div>
      <div aria-label="Animated Agent work example" className="agent-work-demo__stage">
        <div className="agent-work-demo__conversation">
          <p className="agent-work-demo__user-message">Can you audit our auth middleware for security vulnerabilities, patch any algorithm bypasses, and add regression tests?</p>
          <div className="agent-work-demo__transcript">
            {segments.map(segment => {
              if (elapsed < segment.workStart) return null;
              return (
                <div className="agent-work-demo__segment" key={segment.workStart}>
                  <WorkBlock elapsed={elapsed} segment={segment} />
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
