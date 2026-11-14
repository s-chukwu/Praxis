'use client';
import { useEffect, useState } from 'react';
import { Inbox } from 'lucide-react';
import PageHeader from '@/components/ui/PageHeader';
import EmptyState from '@/components/ui/EmptyState';
import { ButtonLink } from '@/components/ui/Button';
import { programs } from '@/lib/catalog';
import { friendly } from '@/lib/errors';
import { getQueue, type Priority, type QueueItem } from '@/lib/reviewer';

const PRIORITY: Record<Priority, { label: string; className: string }> = {
  overdue: { label: 'Overdue', className: 'bg-danger-soft text-danger' },
  due_soon: { label: 'Due soon', className: 'bg-progress-soft text-[#7A4E09]' },
  normal: { label: 'On time', className: 'bg-page text-muted' },
};

function relative(iso: string | null) {
  if (!iso) return '';
  const hours = Math.round((new Date(iso).getTime() - Date.now()) / 3_600_000);
  if (hours === 0) return 'now';
  return hours > 0 ? `in ${hours}h` : `${-hours}h ago`;
}

export default function ReviewerQueue() {
  const [items, setItems] = useState<QueueItem[] | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    getQueue()
      .then(setItems)
      .catch((e) => setError(friendly(e)));
  }, []);

  return (
    <div className="space-y-6">
      <PageHeader title="Review queue" description="Oldest first. Candidates appear as a reference so you mark the work, not the person." />

      {error && (
        <p role="alert" className="rounded-panel bg-danger-soft p-5 font-medium text-danger">
          {error}
        </p>
      )}
      {!items && !error && <div className="h-48 animate-pulse rounded-panel bg-surface shadow-panel" aria-label="Loading the queue" />}
      {items?.length === 0 && (
        <EmptyState icon={<Inbox className="h-5 w-5" aria-hidden />} title="Nothing to review" description="New baseline tasks appear here as soon as candidates submit them." />
      )}

      {items && items.length > 0 && (
        <ul className="divide-y divide-line rounded-panel bg-surface shadow-panel">
          {items.map((item) => {
            const taken = item.status !== 'submitted' && !item.mine;
            const p = PRIORITY[item.priority];
            return (
              <li key={item.attempt_id} className="flex flex-wrap items-center gap-x-6 gap-y-3 p-5">
                <div className="min-w-0 flex-1">
                  <p className="font-semibold tabular-nums">{item.candidate_ref}</p>
                  <p className="mt-0.5 text-sm text-muted">
                    {programs[item.program].name} · {item.stage === 'final' ? 'Final task' : 'Baseline task'} · Submitted {relative(item.submitted_at)}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <span className={`rounded-full px-2.5 py-1 text-xs font-semibold ${p.className}`}>{p.label}</span>
                  <span className="w-20 text-sm text-muted">Due {relative(item.due_at)}</span>
                </div>
                {taken ? (
                  <span className="w-28 text-center text-sm font-medium text-muted">Taken</span>
                ) : (
                  <ButtonLink href={`/reviewer/review/${item.attempt_id}`} variant={item.mine ? 'primary' : 'secondary'} className="w-28">
                    {item.mine ? 'Continue' : 'Start'}
                  </ButtonLink>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
