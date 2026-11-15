-- 011: a score of 0 or 1 no longer needs a supporting quote.
-- Scores of 2, 3 and 4 still need at least one quote. The maximum stays at 10.
ALTER TABLE public.review_dimension_scores
  DROP CONSTRAINT review_dimension_scores_evidence_check;

ALTER TABLE public.review_dimension_scores
  ADD CONSTRAINT review_dimension_scores_evidence_check
  CHECK (cardinality(evidence) <= 10 AND (score <= 1 OR cardinality(evidence) >= 1));