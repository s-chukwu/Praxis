'use client';

// The rubric scale is 0 to 4, matching the backend and rubric 0.2.0.
export default function ScorePicker({ label, value, onChange }: { label: string; value: number | undefined; onChange: (score: number) => void }) {
  return (
    <div role="radiogroup" aria-label={label} className="flex shrink-0 gap-1">
      {[0, 1, 2, 3, 4].map((n) => (
        <button
          key={n}
          type="button"
          role="radio"
          aria-checked={value === n}
          onClick={() => onChange(n)}
          className={`h-9 w-9 rounded-control text-sm font-semibold tabular-nums transition ${
            value === n ? 'bg-brand text-white' : 'border border-line bg-surface text-ink hover:border-brand/50'
          }`}
        >
          {n}
        </button>
      ))}
    </div>
  );
}
