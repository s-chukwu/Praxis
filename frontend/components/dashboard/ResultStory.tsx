import { TrendingUp, Target } from 'lucide-react';
import type { Progress } from '@/lib/baseline';
import { DIMENSIONS, PILOT_THRESHOLD } from '@/lib/rubric';

const name = (id: string) => DIMENSIONS.find((d) => d.id === id)?.title ?? id;

// Leads with one strength and one focus area instead of a table of numbers.
export default function ResultStory({ result }: { result: NonNullable<Progress['result']> }) {
  if (result.scores.length === 0) return null;
  const sorted = [...result.scores].sort((a, b) => b.score - a.score);
  const strongest = sorted[0];
  const focus = sorted[sorted.length - 1];
  const even = strongest.score === focus.score;

  return (
    <section aria-labelledby="story-title" className="animate-rise rounded-panel bg-surface p-6 shadow-panel">
      <h2 id="story-title" className="text-lg font-bold">
        What your review found
      </h2>
      <div className="mt-5 grid gap-4 sm:grid-cols-2">
        <div className="flex gap-3">
          <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-brand-soft text-brand-ink">
            <TrendingUp className="h-[18px] w-[18px]" aria-hidden />
          </span>
          <div>
            <p className="text-sm text-muted">{even ? 'Even across the board' : 'Your strongest area'}</p>
            <p className="font-semibold">
              {name(strongest.dimension)} <span className="font-normal text-muted tabular-nums">· {strongest.score} of 4</span>
            </p>
          </div>
        </div>
        {!even && (
          <div className="flex gap-3">
            <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-progress-soft text-[#7A4E09]">
              <Target className="h-[18px] w-[18px]" aria-hidden />
            </span>
            <div>
              <p className="text-sm text-muted">Where your plan starts</p>
              <p className="font-semibold">
                {name(focus.dimension)} <span className="font-normal text-muted tabular-nums">· {focus.score} of 4</span>
              </p>
            </div>
          </div>
        )}
      </div>
      {result.weighted_score !== null && (
        <p className="mt-5 border-t border-line pt-4 text-[13px] text-muted">
          Overall <span className="font-semibold text-ink tabular-nums">{Math.round(result.weighted_score)}</span> of 100. The pilot pass line is {PILOT_THRESHOLD} and
          has not been validated yet.
        </p>
      )}
    </section>
  );
}
