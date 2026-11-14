import PageHeader from '@/components/ui/PageHeader';
import { skillNames } from '@/lib/catalog';
import { ANCHORS, CRITICAL_EVENTS, DIMENSIONS, PILOT_THRESHOLD, RUBRIC_VERSION } from '@/lib/rubric';

export const metadata = { title: 'Reviewer guidelines' };

const RULES = [
  ['Mark the work, not the person', 'Candidates appear as a reference code. If you recognise someone, hand the case back and tell your lead.'],
  ['Quote, do not summarise', 'Every rubric score needs a quote copied exactly from the answers. The system checks each quote and holds back reviews it cannot find.'],
  ['Score every skill', 'The learning plan is built from the C1 to C6 scores (and B1 to B3 for builders), so none can be skipped.'],
  ['Review within 48 hours', 'Cases turn amber 12 hours before the deadline and red once it passes. Oldest cases come first.'],
  ['Results are not a job promise', 'A strong review makes a candidate eligible for employer review. Never tell a candidate they have a job.'],
];

export default function Guidelines() {
  return (
    <div className="max-w-3xl space-y-8">
      <PageHeader title="Reviewer guidelines" description={`How to score a baseline task with rubric ${RUBRIC_VERSION}.`} />

      <section aria-labelledby="rules-title" className="space-y-3">
        <h2 id="rules-title" className="text-lg font-bold">
          The rules
        </h2>
        <ol className="space-y-3">
          {RULES.map(([title, text], i) => (
            <li key={title} className="flex gap-4 rounded-panel bg-surface p-5 shadow-panel">
              <span className="text-lg font-bold text-brand">{i + 1}</span>
              <div>
                <p className="font-semibold">{title}</p>
                <p className="mt-0.5 text-sm text-muted">{text}</p>
              </div>
            </li>
          ))}
        </ol>
      </section>

      <section aria-labelledby="areas-title" className="space-y-3">
        <h2 id="areas-title" className="text-lg font-bold">
          The four rubric areas
        </h2>
        <div className="divide-y divide-line rounded-panel bg-surface shadow-panel">
          {DIMENSIONS.map((d) => (
            <div key={d.id} className="flex gap-4 p-5">
              <span className="w-12 shrink-0 text-lg font-bold tabular-nums text-brand">{d.weight}%</span>
              <div>
                <p className="font-semibold">{d.title}</p>
                <p className="mt-0.5 text-sm text-muted">{d.evaluates}</p>
                <p className="mt-1 text-xs text-muted">Covers {d.skills.map((s) => skillNames[s]).join(' and ')}</p>
              </div>
            </div>
          ))}
        </div>
        <p className="text-sm text-muted">
          The weighted score is out of 100. The pilot line is {PILOT_THRESHOLD}, and it has not been validated yet, so treat it as a guide.
        </p>
      </section>

      <section aria-labelledby="levels-title" className="space-y-3">
        <h2 id="levels-title" className="text-lg font-bold">
          What each score means
        </h2>
        <dl className="divide-y divide-line rounded-panel bg-surface shadow-panel">
          {ANCHORS.map(([n, text]) => (
            <div key={n} className="flex gap-4 p-5">
              <dt className="w-8 shrink-0 text-lg font-bold tabular-nums">{n}</dt>
              <dd className="text-sm">{text}</dd>
            </div>
          ))}
        </dl>
      </section>

      <section aria-labelledby="critical-title" className="space-y-3">
        <h2 id="critical-title" className="text-lg font-bold">
          Critical events
        </h2>
        <p className="text-sm text-muted">Any one of these, once confirmed, fails the safety check whatever the other scores are.</p>
        <ul className="divide-y divide-line rounded-panel border border-danger/20 bg-surface">
          {CRITICAL_EVENTS.map(([title, text]) => (
            <li key={title} className="p-5">
              <p className="font-semibold text-danger">{title}</p>
              <p className="mt-0.5 text-sm text-muted">{text}</p>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
