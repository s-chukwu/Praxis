'use client';
import { useEffect, useState } from 'react';
import { ClipboardList } from 'lucide-react';
import PageHeader from '@/components/ui/PageHeader';
import EmptyState from '@/components/ui/EmptyState';
import { ButtonLink } from '@/components/ui/Button';
import { friendly } from '@/lib/errors';
import { getMyReviews, type ReviewSummary } from '@/lib/reviewer';

const OUTCOMES: Record<string, string> = {
  evidence_verified: 'Meets the standard',
  needs_practice: 'Needs practice',
  review_pending: 'Held for checking',
};

const day = (iso: string) => new Date(iso).toLocaleDateString(undefined, { day: 'numeric', month: 'short' });

export default function MyReviews() {
  const [reviews, setReviews] = useState<ReviewSummary[] | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    getMyReviews()
      .then(setReviews)
      .catch((e) => setError(friendly(e)));
  }, []);

  return (
    <div className="space-y-6">
      <PageHeader title="My reviews" description="Drafts to finish first, then reviews you have submitted." />
      {error && (
        <p role="alert" className="rounded-panel bg-danger-soft p-5 font-medium text-danger">
          {error}
        </p>
      )}
      {!reviews && !error && <div className="h-40 animate-pulse rounded-panel bg-surface shadow-panel" aria-label="Loading your reviews" />}
      {reviews?.length === 0 && (
        <EmptyState
          icon={<ClipboardList className="h-5 w-5" aria-hidden />}
          title="No reviews yet"
          description="Reviews you start from the queue appear here."
          actionHref="/reviewer"
          actionLabel="Open the queue"
        />
      )}
      {reviews && reviews.length > 0 && (
        <ul className="divide-y divide-line rounded-panel bg-surface shadow-panel">
          {reviews.map((r) => (
            <li key={r.id} className="flex flex-wrap items-center gap-4 p-5">
              <div className="min-w-0 flex-1">
                <p className="font-semibold">{r.finalized_at ? OUTCOMES[r.proposed_outcome ?? ''] ?? 'Submitted' : 'Draft'}</p>
                <p className="mt-0.5 text-sm text-muted">
                  Started {day(r.created_at)}
                  {r.finalized_at && ` · Submitted ${day(r.finalized_at)}`}
                </p>
              </div>
              {r.finalized_at ? (
                r.weighted_score !== null && <span className="text-lg font-bold tabular-nums">{Math.round(r.weighted_score)}</span>
              ) : (
                <ButtonLink href={`/reviewer/review/${r.attempt_id}`}>Continue</ButtonLink>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
