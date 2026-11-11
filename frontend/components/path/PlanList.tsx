import EmptyState from '@/components/ui/EmptyState';

export type PlanItem = { order: number; code: string; title: string; reason?: string };

export default function PlanList({ items }: { items?: PlanItem[] }) {
  if (!items || items.length === 0) {
    return (
      <EmptyState
        title="Your learning plan is not ready"
        description="It will appear here once your baseline has been reviewed."
      />
    );
  }
  return (
    <ol className="divide-y divide-[#161B22] rounded-lg border border-[#161B22] bg-[#0D1117]">
      {items.map((i) => (
        <li key={i.code} className="flex items-center gap-4 p-4">
          <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border border-[#1C222B] text-xs text-gray-300">
            {i.order}
          </span>
          <div>
            <p className="text-sm font-medium text-white">{i.title}</p>
            <p className="mt-0.5 text-xs text-gray-400">
              {i.code}
              {i.reason ? ` · ${i.reason}` : ''}
            </p>
          </div>
        </li>
      ))}
    </ol>
  );
}