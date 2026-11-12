import { ListOrdered } from 'lucide-react';
import EmptyState from '@/components/ui/EmptyState';
import { courseNames } from '@/lib/catalog';

export type PlanItem = { order: number; code: string; reason?: string };

export default function PlanList({ items }: { items?: PlanItem[] }) {
  if (!items || items.length === 0) {
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
    <ol className="divide-y divide-line rounded-panel bg-surface shadow-panel">
      {items.map((i) => (
        <li key={i.code} className="flex gap-4 p-5">
          <span className="text-xl font-bold leading-7 text-brand">{i.order}</span>
          <div>
            <p className="font-semibold">{courseNames[i.code] ?? i.code}</p>
            {i.reason && <p className="mt-0.5 text-sm text-muted">{i.reason}</p>}
          </div>
        </li>
      ))}
    </ol>
  );
}
