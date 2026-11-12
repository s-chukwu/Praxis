import { BadgeCheck } from 'lucide-react';

const steps = [
  ['Baseline task', 'A real support scenario, 8 to 12 minutes.'],
  ['Reviewed by a person', 'A trained supervisor scores your work.'],
  ['Your learning plan', 'Only the courses for your gaps, in order.'],
  ['Final task', 'A new scenario shows what improved.'],
];

export default function AuthShowcase() {
  return (
    <div className="relative hidden flex-1 flex-col justify-center overflow-hidden bg-brand px-14 text-white lg:flex">
      <div className="max-w-[440px]">
        <h2 className="text-[34px] font-bold leading-[1.15]">Train for the support job AI is changing.</h2>
        <p className="mt-4 text-[17px] text-white/80">Skip what you already know. Prove what you learned.</p>

        <ol className="mt-12 space-y-0">
          {steps.map(([title, text], i) => (
            <li key={title} className="relative flex gap-4 pb-7 last:pb-0">
              {i < steps.length - 1 && <span aria-hidden className="absolute left-[15px] top-9 h-[calc(100%-36px)] w-0.5 bg-white/20" />}
              <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white text-sm font-bold text-brand-ink">
                {i + 1}
              </span>
              <div className="pt-1">
                <p className="font-semibold">{title}</p>
                <p className="text-sm text-white/75">{text}</p>
              </div>
            </li>
          ))}
        </ol>

        <p className="mt-12 flex items-start gap-2.5 text-sm text-white/80">
          <BadgeCheck className="mt-0.5 h-4 w-4 shrink-0" aria-hidden />
          Results are checked by people and show your work on practice tasks. They are not a job offer.
        </p>
      </div>
    </div>
  );
}
