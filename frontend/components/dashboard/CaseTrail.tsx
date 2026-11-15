import { Award, Check, Lock, Medal } from 'lucide-react';
import { milestones, type Milestone } from '@/lib/journey';
import type { Progress } from '@/lib/baseline';

function Mark({ m }: { m: Milestone }) {
  const Icon = m.kind === 'certificate' ? Award : m.kind === 'badge' ? Medal : Check;
  if (m.state === 'earned') {
    return (
      <span className="flex h-10 w-10 items-center justify-center rounded-full bg-brand text-white">
        <Icon className="h-[18px] w-[18px]" strokeWidth={2.4} aria-hidden />
      </span>
    );
  }
  if (m.state === 'current') {
    return (
      <span className="flex h-10 w-10 items-center justify-center rounded-full bg-surface text-progress ring-2 ring-progress">
        <Icon className="h-[18px] w-[18px]" strokeWidth={2.2} aria-hidden />
      </span>
    );
  }
  return (
    <span className="flex h-10 w-10 items-center justify-center rounded-full bg-page text-muted/70 ring-1 ring-line">
      <Lock className="h-4 w-4" aria-hidden />
    </span>
  );
}

// A support case moves through states; so does the learner. Agents read this trail at a glance.
export default function CaseTrail({ progress }: { progress: Progress | null }) {
  const list = milestones(progress);
  const earned = list.filter((m) => m.state === 'earned').length;

  return (
    <section aria-labelledby="trail-title" className="rounded-panel bg-surface p-6 shadow-panel">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h2 id="trail-title" className="text-lg font-bold">
          Your training record
        </h2>
        <p className="text-sm text-muted">
          <span className="font-semibold text-ink tabular-nums">{earned}</span> of {list.length} done
        </p>
      </div>

      <ol className="relative mt-6 grid grid-cols-1 gap-5 sm:grid-cols-5 sm:gap-2">
        {list.map((m, i) => (
          <li key={m.id} className="relative flex gap-4 sm:flex-col sm:items-center sm:gap-3 sm:text-center">
            {i < list.length - 1 && (
              <span
                aria-hidden
                className={`absolute left-5 top-10 h-[calc(100%-0.25rem)] w-0.5 sm:left-[calc(50%+1.5rem)] sm:top-5 sm:h-0.5 sm:w-[calc(100%-3rem)] ${
                  m.state === 'earned' ? 'bg-brand' : 'bg-line'
                }`}
              />
            )}
            <Mark m={m} />
            <div className="min-w-0 pt-1.5 sm:pt-0">
              <p className={`font-semibold leading-snug ${m.state === 'locked' ? 'text-muted' : 'text-ink'}`}>
                {m.title}
              </p>
              <p className="mt-0.5 text-[13px] leading-snug text-muted">{m.state === 'earned' && m.kind !== 'step' ? 'Earned' : m.detail}</p>
            </div>
          </li>
        ))}
      </ol>
    </section>
  );
}
