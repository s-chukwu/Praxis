'use client';
import { useEffect, useState } from 'react';
import { Clock, Hourglass } from 'lucide-react';
import { ButtonLink } from '@/components/ui/Button';
import { getBaseline, type Attempt, type Progress } from '@/lib/baseline';
import { ago, hoursLeft, REVIEW_HOURS } from '@/lib/journey';
import { report } from '@/lib/errors';

const REQUIRED: [string, string][] = [
  ['customer_reply', 'Your reply to the customer'],
  ['internal_note', 'Internal note for your team'],
  ['verification_steps', 'How you would check your answer'],
];

function Panel({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section aria-labelledby="next-title" className="flex flex-col rounded-panel bg-ink p-6 text-white shadow-panel">
      <h2 id="next-title" className="text-sm font-semibold text-white/70">
        Your next step
      </h2>
      <p className="mt-2 text-2xl font-bold leading-snug [text-wrap:balance]">{title}</p>
      {children}
    </section>
  );
}

export default function NextStep({ progress }: { progress: Progress | null }) {
  const status = progress?.baseline?.status;
  const [attempt, setAttempt] = useState<Attempt | null>(null);

  // Only a half-finished task needs its answers, to say exactly where the learner stopped.
  useEffect(() => {
    if (status === 'in_progress' || status === 'assigned') getBaseline().then(setAttempt).catch(report);
  }, [status]);

  if (progress?.plan) {
    const mods = progress.plan.modules;
    const index = Math.max(0, mods.findIndex((m) => !m.completed_at));
    const next = mods[index];
    return (
      <Panel title={next ? next.title : 'Your learning plan'}>
        <p className="mt-2 text-white/75">
          Course {index + 1} of {mods.length}, chosen from your review. Start here; everything else waits.
        </p>
        <ButtonLink href="/path/plan" size="lg" className="mt-6 lg:mt-auto">
          Open my plan
        </ButtonLink>
      </Panel>
    );
  }

  if (progress?.result) {
    return (
      <Panel title="Your plan is on its way">
        <p className="mt-2 text-white/75">Your review is done. The courses it chose for you appear here as soon as they are ready.</p>
      </Panel>
    );
  }

  if (status === 'submitted' || status === 'under_review' || status === 'review_pending') {
    const left = hoursLeft(progress?.baseline?.submitted_at ?? null);
    return (
      <Panel title="Your work is with a reviewer">
        <p className="mt-2 text-white/75">A trained support supervisor is scoring your answers against the same rubric every candidate gets.</p>
        <div className="mt-5 rounded-control bg-white/[0.07] p-4">
          <p className="flex items-center gap-2 font-semibold">
            <Hourglass className="h-4 w-4 text-progress" aria-hidden />
            {left === null ? 'Reviews take up to 48 hours' : left > 0 ? `Due within ${left} ${left === 1 ? 'hour' : 'hours'}` : 'Taking a little longer than usual'}
          </p>
          {left !== null && (
            <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-white/10" aria-hidden>
              <div className="h-full rounded-full bg-progress" style={{ width: `${((REVIEW_HOURS - left) / REVIEW_HOURS) * 100}%` }} />
            </div>
          )}
        </div>
        <p className="mt-4 text-sm text-white/60">Come back for your result and the courses picked for you.</p>
      </Panel>
    );
  }

  if (status === 'in_progress' || status === 'assigned') {
    const done = REQUIRED.filter(([id]) => attempt?.answers?.[id]?.trim()).length;
    const nextField = REQUIRED.find(([id]) => !attempt?.answers?.[id]?.trim());
    return (
      <Panel title="Pick up where you left off">
        <div className="mt-4 flex gap-1.5" aria-hidden>
          {REQUIRED.map(([id]) => (
            <span key={id} className={`h-1.5 flex-1 rounded-full ${attempt?.answers?.[id]?.trim() ? 'bg-progress' : 'bg-white/15'}`} />
          ))}
        </div>
        <p className="mt-3 text-white/80">
          {done} of 3 answers written{attempt?.updated_at && <span className="text-white/55"> · saved {ago(attempt.updated_at)}</span>}
        </p>
        {nextField && <p className="mt-1 text-sm text-white/60">Next: {nextField[1]}</p>}
        <ButtonLink href="/path/baseline" size="lg" className="mt-6 lg:mt-auto">
          Continue the case
        </ButtonLink>
      </Panel>
    );
  }

  return (
    <Panel title="Your first case is waiting">
      <p className="mt-2 text-white/75">One real customer support case. It shows what you already know, so your training can skip it.</p>
      <p className="mt-4 flex items-center gap-2 text-sm text-white/75">
        <Clock className="h-4 w-4" aria-hidden /> About 12 minutes, no timer. Leave and come back any time.
      </p>
      <ButtonLink href="/path/baseline" size="lg" className="mt-6 lg:mt-auto">
        Open the case
      </ButtonLink>
    </Panel>
  );
}
