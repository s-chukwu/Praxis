import Link from 'next/link';
import { Check } from 'lucide-react';
import type { Learner } from '@/lib/learner';
import type { Progress } from '@/lib/baseline';

export type State = 'done' | 'next' | 'waiting' | 'later';
export type Step = { title: string; text: string; href?: string; action?: string; state: State };

// Each step follows the backend: submitted moves on to Review, reviewed moves on to the plan.
export function pathSteps(learner: Learner, progress: Progress | null): Step[] {
  const status = progress?.baseline?.status;
  const submitted = status === 'submitted' || status === 'under_review' || status === 'review_pending' || status === 'reviewed';
  const reviewed = Boolean(progress?.result);
  const planReady = Boolean(progress?.plan);
  const started = status === 'in_progress' || status === 'assigned';

  return [
    {
      title: 'Setup',
      text: learner.cvSubmitted ? 'Program chosen and CV added.' : 'Program chosen. CV not added yet.',
      href: '/path/cv',
      action: 'Edit',
      state: 'done',
    },
    {
      title: 'Baseline task',
      text: submitted ? 'Submitted. Thank you.' : started ? 'Started. Your answers are saved.' : 'One support scenario, 8 to 12 minutes.',
      href: '/path/baseline',
      action: started ? 'Continue' : 'Start',
      state: submitted ? 'done' : 'next',
    },
    {
      title: 'Review',
      text: reviewed ? 'Scored by a reviewer.' : submitted ? 'A reviewer is scoring your work.' : 'A supervisor scores your work.',
      state: reviewed ? 'done' : submitted ? 'waiting' : 'later',
    },
    {
      title: 'Learning plan',
      text: planReady ? 'Your courses are ready, in order.' : reviewed ? 'Your plan is being prepared.' : 'Courses for your gaps, in order.',
      href: '/path/plan',
      action: 'Open',
      state: planReady ? 'next' : reviewed ? 'waiting' : 'later',
    },
    { title: 'Final task', text: 'A new scenario shows what improved.', state: 'later' },
  ];
}

export default function PathList({ learner, progress }: { learner: Learner; progress: Progress | null }) {
  return (
    <ol className="rounded-panel bg-surface p-2 shadow-panel">
      {pathSteps(learner, progress).map((s, i, all) => (
        <li key={s.title} className="relative flex gap-4 rounded-control p-4">
          {i < all.length - 1 && (
            <span aria-hidden className={`absolute left-[31px] top-[48px] h-[calc(100%-32px)] w-0.5 ${s.state === 'done' ? 'bg-brand' : 'bg-line'}`} />
          )}
          <span
            className={`relative flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-bold ${
              s.state === 'done'
                ? 'bg-brand text-white'
                : s.state === 'next'
                  ? 'bg-surface text-brand ring-2 ring-brand'
                  : s.state === 'waiting'
                    ? 'bg-progress-soft text-progress ring-2 ring-progress'
                  : 'bg-page text-muted ring-1 ring-line'
            }`}
          >
            {s.state === 'done' ? <Check className="h-4 w-4" strokeWidth={3} aria-label="Done" /> : i + 1}
          </span>
          <div className="min-w-0 flex-1 pt-1">
            <p className={`font-semibold ${s.state === 'later' ? 'text-muted' : 'text-ink'}`}>{s.title}</p>
            <p className="text-sm text-muted">{s.text}</p>
          </div>
          {s.href && (s.state === 'done' || s.state === 'next') && (
            <Link href={s.href} className="self-center text-sm font-semibold text-brand-ink hover:underline">
              {s.state === 'done' && s.title !== 'Setup' ? 'View' : s.action}
            </Link>
          )}
        </li>
      ))}
    </ol>
  );
}
