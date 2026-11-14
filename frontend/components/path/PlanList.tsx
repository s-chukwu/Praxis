import { ListOrdered } from 'lucide-react';
import EmptyState from '@/components/ui/EmptyState';
import { courseNames, skillNames } from '@/lib/catalog';
import type { PlanReason, Progress } from '@/lib/baseline';

const DIMENSIONS: Record<string, string> = {
  accuracy: 'Accuracy',
  judgment: 'Judgment',
  safety: 'Safety',
  communication: 'Communication',
};

const OUTCOMES: Record<string, string> = {
  evidence_verified: 'Your work met the standard',
  needs_practice: 'Some skills need practice',
  review_pending: 'Your review is being checked',
};

function reasonText(r: PlanReason) {
  switch (r.code) {
    case 'competency_gap':
      return `Your review showed room to grow in ${skillNames[r.competency ?? ''] ?? 'this skill'}.`;
    case 'safety_gate':
      return 'Required for everyone before working with customers.';
    case 'role_orientation':
      return 'Everyone starts here.';
    case 'prerequisite':
      return 'Needed before a later course.';
    case 'integrated_challenge':
      return 'Brings your skills together.';
    default:
      return '';
  }
}

export default function PlanList({ progress }: { progress: Progress | null }) {
  const result = progress?.result;
  const modules = progress?.plan?.modules ?? [];

  if (!result) {
    return (
      <EmptyState
        icon={<ListOrdered className="h-5 w-5" aria-hidden />}
        title="Your plan comes after your baseline review"
        description="Once a reviewer scores your baseline task, your courses appear here in order, each with the reason it is in your plan."
        actionHref="/path/baseline"
        actionLabel="Go to baseline task"
      />
    );
  }

  return (
    <div className="space-y-6">
      <section aria-labelledby="result-title" className="animate-rise rounded-panel bg-surface p-6 shadow-panel">
        <h2 id="result-title" className="font-bold">
          {OUTCOMES[result.proposed_outcome ?? ''] ?? 'Your baseline result'}
        </h2>
        {result.weighted_score !== null && <p className="mt-1 text-sm text-muted">Overall {Math.round(result.weighted_score)} out of 100</p>}
        <ul className="mt-5 grid gap-4 sm:grid-cols-2">
          {result.scores.map((s) => (
            <li key={s.dimension}>
              <div className="flex justify-between text-sm">
                <span className="font-medium">{DIMENSIONS[s.dimension] ?? s.dimension}</span>
                <span className="text-muted">{s.score} of 4</span>
              </div>
              <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-line" aria-hidden>
                <div className="h-full rounded-full bg-brand transition-[width] duration-700 ease-out" style={{ width: `${(s.score / 4) * 100}%` }} />
              </div>
            </li>
          ))}
        </ul>
      </section>

      {modules.length === 0 ? (
        <div role="status" className="rounded-panel border border-dashed border-line bg-surface p-6">
          <p className="font-semibold">Your courses are being prepared</p>
          <p className="mt-1 text-sm text-muted">They appear here as soon as your plan is ready.</p>
        </div>
      ) : (
        <ol className="divide-y divide-line rounded-panel bg-surface shadow-panel">
          {modules.map((m) => {
            const why = m.reasons.map(reasonText).filter(Boolean)[0];
            return (
              <li key={m.course_code} className="flex gap-4 p-5">
                <span className="text-xl font-bold leading-7 text-brand">{m.position}</span>
                <div>
                  <p className="font-semibold">{m.title || courseNames[m.course_code] || m.course_code}</p>
                  {why && <p className="mt-0.5 text-sm text-muted">{why}</p>}
                </div>
              </li>
            );
          })}
        </ol>
      )}
    </div>
  );
}
