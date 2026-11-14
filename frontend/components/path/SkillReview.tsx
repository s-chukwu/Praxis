'use client';
import { Check, Quote, X } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { skillNames } from '@/lib/catalog';
import type { Suggestion } from '@/lib/learner';

export type Decision = 'confirmed' | 'rejected';

export default function SkillReview({
  suggestions,
  decisions,
  onDecide,
}: {
  suggestions: Suggestion[];
  decisions: Record<string, Decision>;
  onDecide: (code: string, decision: Decision) => void;
}) {
  const answered = suggestions.filter((s) => decisions[s.competency]).length;

  return (
    <div>
      <div className="flex flex-wrap items-end justify-between gap-2">
        <p className="text-lg font-bold" aria-live="polite">
          {suggestions.length} {suggestions.length === 1 ? 'skill' : 'skills'} found in your CV
        </p>
        <p className="text-sm text-muted" aria-live="polite">
          {answered} of {suggestions.length} answered
        </p>
      </div>
      <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-line" aria-hidden>
        <div className="h-full rounded-full bg-brand transition-[width] duration-500 ease-out" style={{ width: `${(answered / suggestions.length) * 100}%` }} />
      </div>

      <ul className="mt-5 space-y-3">
        {suggestions.map((s, i) => {
          const d = decisions[s.competency];
          return (
            <li
              key={s.competency}
              style={{ animationDelay: `${i * 90}ms` }}
              className={`animate-rise rounded-panel border bg-surface p-5 transition-colors duration-300 ${
                d === 'confirmed' ? 'border-brand/50 bg-brand-soft/30' : d === 'rejected' ? 'border-line opacity-70' : 'border-line'
              }`}
            >
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div className="min-w-0 flex-1">
                  <p className="flex items-center gap-2 font-semibold">
                    {d === 'confirmed' && (
                      <span className="flex h-5 w-5 animate-pop items-center justify-center rounded-full bg-brand text-white">
                        <Check className="h-3 w-3" strokeWidth={3} aria-hidden />
                      </span>
                    )}
                    {skillNames[s.competency]}
                  </p>
                  <p className="mt-2 flex gap-2 text-sm text-muted">
                    <Quote className="mt-0.5 h-3.5 w-3.5 shrink-0" aria-hidden />
                    <span>{s.evidence_quote}</span>
                  </p>
                </div>
                <div className="flex gap-2" role="group" aria-label={`Is ${skillNames[s.competency]} right?`}>
                  <Button
                    variant={d === 'confirmed' ? 'primary' : 'secondary'}
                    aria-pressed={d === 'confirmed'}
                    onClick={() => onDecide(s.competency, 'confirmed')}
                  >
                    <Check className="h-4 w-4" aria-hidden /> Yes, that's me
                  </Button>
                  <Button
                    variant="secondary"
                    aria-pressed={d === 'rejected'}
                    className={d === 'rejected' ? 'border-ink/40 bg-page' : ''}
                    onClick={() => onDecide(s.competency, 'rejected')}
                  >
                    <X className="h-4 w-4" aria-hidden /> No
                  </Button>
                </div>
              </div>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
