import Link from 'next/link';
import { Check } from 'lucide-react';
import type { Learner } from '@/lib/learner';

type State = 'done' | 'next' | 'later';

// Only the CV step can be finished today; the rest move on as our Go backend reports attempts and reviews.
export function pathSteps(learner: Learner) {
  return [
    { title: 'Setup', text: 'Program chosen and CV added.', href: '/path/cv', state: 'done' as State },
    { title: 'Baseline task', text: 'One support scenario, 8 to 12 minutes.', href: '/path/baseline', state: 'next' as State },
    { title: 'Review', text: 'A supervisor scores your work.', state: 'later' as State },
    { title: 'Learning plan', text: 'Courses for your gaps, in order.', href: '/path/plan', state: 'later' as State },
    { title: 'Final task', text: 'A new scenario shows what improved.', state: 'later' as State },
  ].map((s, i) => (i === 0 && !learner.cvSubmitted ? { ...s, text: 'Program chosen. CV not added yet.' } : s));
}

export default function PathList({ learner }: { learner: Learner }) {
  return (
    <ol className="rounded-panel bg-surface p-2 shadow-panel">
      {pathSteps(learner).map((s, i, all) => (
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
                  : 'bg-page text-muted ring-1 ring-line'
            }`}
          >
            {s.state === 'done' ? <Check className="h-4 w-4" strokeWidth={3} aria-label="Done" /> : i + 1}
          </span>
          <div className="min-w-0 flex-1 pt-1">
            <p className={`font-semibold ${s.state === 'later' ? 'text-muted' : 'text-ink'}`}>{s.title}</p>
            <p className="text-sm text-muted">{s.text}</p>
          </div>
          {s.href && s.state !== 'later' && (
            <Link href={s.href} className="self-center text-sm font-semibold text-brand-ink hover:underline">
              {s.state === 'done' ? 'Edit' : 'Start'}
            </Link>
          )}
        </li>
      ))}
    </ol>
  );
}
