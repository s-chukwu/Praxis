import PageHeader from '@/components/ui/PageHeader';
import { ButtonLink } from '@/components/ui/Button';
import { ANCHORS } from '@/lib/rubric';

export const metadata = { title: 'Calibration' };

const STEPS = [
  'Your lead picks one submitted case and every reviewer scores it alone.',
  'Compare scores area by area. A gap of more than one point means the anchors are being read differently.',
  'Talk through the quotes each reviewer chose, then agree which anchor fits.',
  'Write down any rule you agreed so the next session starts from it.',
];

export default function Calibration() {
  return (
    <div className="max-w-3xl space-y-8">
      <PageHeader title="Calibration" description="Keep every reviewer scoring the same work the same way." />

      <section aria-labelledby="how-title" className="rounded-panel bg-surface p-6 shadow-panel">
        <h2 id="how-title" className="font-bold">
          How a session works
        </h2>
        <ol className="mt-3 list-decimal space-y-2 pl-5 text-sm">
          {STEPS.map((s) => (
            <li key={s}>{s}</li>
          ))}
        </ol>
        <p className="mt-4 text-sm text-muted">Practice cases inside Praxis are not ready yet. Until then, your lead runs sessions with a real case from the queue.</p>
      </section>

      <section aria-labelledby="anchors-title" className="space-y-3">
        <h2 id="anchors-title" className="text-lg font-bold">
          The line between scores
        </h2>
        <div className="grid gap-3 sm:grid-cols-2">
          {ANCHORS.slice(1).map(([n, text], i) => (
            <div key={n} className="rounded-panel bg-surface p-5 shadow-panel">
              <p className="text-sm font-semibold text-muted">
                {ANCHORS[i][0]} or {n}?
              </p>
              <p className="mt-2 text-sm">
                Give <strong>{n}</strong> only if: {text}
              </p>
            </div>
          ))}
        </div>
      </section>

      <ButtonLink href="/reviewer/guidelines" variant="secondary">
        Read the full guidelines
      </ButtonLink>
    </div>
  );
}
