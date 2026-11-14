import { PREVIEW, request } from './learner';

export type Priority = 'overdue' | 'due_soon' | 'normal';

export type QueueItem = {
  attempt_id: string;
  candidate_ref: string;
  program: 'operator' | 'builder';
  stage: 'baseline' | 'final';
  status: 'submitted' | 'under_review' | 'review_pending';
  submitted_at: string | null;
  due_at: string | null;
  priority: Priority;
  mine: boolean;
};

export type ReviewSummary = {
  id: string;
  attempt_id: string;
  finalized_at: string | null;
  weighted_score: number | null;
  proposed_outcome: string | null;
  created_at: string;
};

export type DimensionScore = { dimension: string; score: number; evidence: string[] };
export type Finding = { event_type: string; supporting_quote: string; policy_reference: string; decision: 'confirmed' | 'dismissed' | 'pending' };
export type Review = ReviewSummary & { scores: DimensionScore[]; findings: Finding[] };

export type ReviewCase = {
  attemptId: string;
  candidateRef: string;
  program: 'operator' | 'builder';
  answers: Record<string, string>;
  brief: string;
  checks: string[];
};

// Draft competency scores and notes have no backend column, so they wait in this browser until finalize.
export type Draft = { competencies: Record<string, number>; notes: string };

const hoursAgo = (h: number) => new Date(Date.now() - h * 3_600_000).toISOString();

const SAMPLE_QUEUE: QueueItem[] = [
  { attempt_id: 'sample-1', candidate_ref: 'C-4f2a91c0', program: 'operator', stage: 'baseline', status: 'submitted', submitted_at: hoursAgo(50), due_at: hoursAgo(2), priority: 'overdue', mine: false },
  { attempt_id: 'sample-2', candidate_ref: 'C-9b7d13e2', program: 'operator', stage: 'baseline', status: 'under_review', submitted_at: hoursAgo(40), due_at: hoursAgo(-8), priority: 'due_soon', mine: true },
  { attempt_id: 'sample-3', candidate_ref: 'C-27c8e5aa', program: 'builder', stage: 'baseline', status: 'submitted', submitted_at: hoursAgo(3), due_at: hoursAgo(-45), priority: 'normal', mine: false },
];

const SAMPLE_CASE: Omit<ReviewCase, 'attemptId' | 'candidateRef' | 'program'> = {
  brief:
    'You work at fictional electronics shop Inyange Tech. A customer writes: "My MoMo was debited twice for my headphones. Send my money back today." Order IT-3307 shows one completed payment of RWF 35,000. The payment provider log shows a second attempt with status \'failed - reversal pending\'. Reversals and refunds need finance team confirmation.',
  answers: {
    customer_reply:
      'Hello, I am sorry for the worry. I checked your order IT-3307 and I can see one completed payment of RWF 35,000. The second attempt failed and is waiting to be reversed, so you were not charged twice. I have asked our finance team to confirm the reversal and will update you within 24 hours.',
    internal_note: 'Customer reports double MoMo debit on IT-3307. Provider log shows second attempt failed, reversal pending. Escalated to finance for confirmation. No refund promised.',
    verification_steps: 'Checked order IT-3307 payment status\nChecked provider log for the second attempt\nAsked finance to confirm the reversal',
    ai_interaction: 'AI draft promised a refund today. I removed that because refunds need finance approval.',
  },
  checks: ['Does not promise a refund or a date finance has not confirmed', 'States what the records show: one completed payment', 'Escalates to finance with the order number'],
};

export async function getQueue(): Promise<QueueItem[]> {
  if (PREVIEW) return SAMPLE_QUEUE;
  return (await request<{ items: QueueItem[] }>('GET', '/v1/review-queue')).items;
}

export async function getMyReviews(): Promise<ReviewSummary[]> {
  if (PREVIEW) return [];
  return (await request<{ reviews: ReviewSummary[] }>('GET', '/v1/reviews')).reviews;
}

// Opening a case you already hold returns your draft, so this is safe on every page load.
export async function openReview(item: QueueItem): Promise<{ review: Review; reviewCase: ReviewCase }> {
  const base = { attemptId: item.attempt_id, candidateRef: item.candidate_ref, program: item.program };
  if (PREVIEW) {
    const review: Review = { id: `review-${item.attempt_id}`, attempt_id: item.attempt_id, finalized_at: null, weighted_score: null, proposed_outcome: null, created_at: hoursAgo(1), scores: [], findings: [] };
    return { review, reviewCase: { ...base, ...SAMPLE_CASE } };
  }
  const review = await request<Review>('POST', `/v1/attempts/${item.attempt_id}/review`);
  const [attempt, packet] = await Promise.all([
    request<{ answers?: Record<string, string> }>('GET', `/v1/attempts/${item.attempt_id}`),
    request<{ scenario: { candidate_brief: string }; key: { checks: string[] } }>('GET', `/v1/attempts/${item.attempt_id}/assessor-packet`),
  ]);
  return { review, reviewCase: { ...base, answers: attempt.answers ?? {}, brief: packet.scenario.candidate_brief, checks: packet.key.checks } };
}

export async function saveScores(reviewId: string, scores: DimensionScore[]): Promise<void> {
  if (PREVIEW) return;
  await request('PUT', `/v1/reviews/${reviewId}/scores`, { scores });
}

// Replaces the whole list, so an empty list clears findings the reviewer removed.
export async function saveFindings(reviewId: string, findings: Finding[]): Promise<void> {
  if (PREVIEW) return;
  await request('PUT', `/v1/reviews/${reviewId}/findings`, { findings });
}

// Finalizing asks the AI service for the weighted score, then builds the candidate's plan from the skill scores.
export async function finalizeReview(reviewId: string, draft: Draft): Promise<void> {
  if (PREVIEW) return;
  await request('POST', `/v1/reviews/${reviewId}/finalize`, { reviewed_competencies: draft.competencies, notes: draft.notes });
  await request('POST', `/v1/reviews/${reviewId}/learning-plan`);
}

const draftKey = (reviewId: string) => `praxis.review-draft.${reviewId}`;

export function loadDraft(reviewId: string): Draft {
  try {
    return { competencies: {}, notes: '', ...JSON.parse(localStorage.getItem(draftKey(reviewId)) ?? '{}') };
  } catch {
    return { competencies: {}, notes: '' };
  }
}

export function storeDraft(reviewId: string, draft: Draft) {
  try {
    localStorage.setItem(draftKey(reviewId), JSON.stringify(draft));
  } catch {
    // Private browsing: the draft stays in memory for this visit.
  }
}

export function clearDraft(reviewId: string) {
  try {
    localStorage.removeItem(draftKey(reviewId));
  } catch {
    // Nothing to clear.
  }
}
