CREATE UNIQUE INDEX IF NOT EXISTS attempts_one_active_baseline_per_candidate
ON public.attempts (candidate_id)
WHERE stage = 'baseline' AND status <> 'voided';

DROP INDEX IF EXISTS public.attempts_one_baseline_per_candidate;

