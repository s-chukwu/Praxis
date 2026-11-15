-- Migration 011: a score of 0 or 1 may have no evidence quote.
-- A low score often rests on what the candidate left out, and a missing step has nothing to quote.
-- Scores of 2 or more still need 1 to 10 quotes. Matches the AI service and the Go store rule.

do $$
declare
  c text;
begin
  -- 004 declared the check inline, so find it by its definition rather than guessing its name.
  for c in
    select conname from pg_constraint
    where conrelid = 'public.review_dimension_scores'::regclass
      and contype = 'c'
      and pg_get_constraintdef(oid) like '%cardinality(evidence)%'
  loop
    execute format('alter table public.review_dimension_scores drop constraint %I', c);
  end loop;
end $$;

alter table public.review_dimension_scores
  add constraint review_dimension_scores_evidence_rule
  check (cardinality(evidence) <= 10 and (score <= 1 or cardinality(evidence) >= 1));
