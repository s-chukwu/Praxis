-- Migration 008: enforce one baseline attempt per candidate.
-- A candidate may have other attempt stages, but only one baseline attempt.

create unique index attempts_one_baseline_per_candidate
    on attempts (candidate_id)
    where stage = 'baseline';