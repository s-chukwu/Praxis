-- Migration 005: gaps found when checking the database against the AI service.
-- Additive only (expand step), so nothing existing breaks.
--   1. attempts get scenario_id and stage (needed to call /score and /compare)
--   2. critical findings get their own table (event type, quote, policy reference, decision)
--   3. reviews store the weighted score, proposed outcome and rubric/scenario hashes for audit

-- 1. Scenario and stage on attempts
alter table attempts
  add column scenario_id text,
  add column stage text not null default 'baseline';

-- 2. Critical findings
create type finding_decision as enum ('confirmed', 'dismissed', 'pending');

create table critical_findings (
  id uuid primary key default gen_random_uuid(),
  review_id uuid not null references reviews (id) on delete cascade,
  event_type text not null,
  supporting_quote text not null
    check (char_length(btrim(supporting_quote)) between 12 and 500),
  policy_reference text not null,
  decision finding_decision not null default 'pending',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index critical_findings_review_idx on critical_findings (review_id);

create trigger critical_findings_updated_at
  before update on critical_findings
  for each row execute function set_updated_at();

-- Findings freeze when their review is finalized (same rule as dimension scores)
create trigger critical_findings_immutable
  before insert or update or delete on critical_findings
  for each row execute function block_finalized_dimension_changes();

alter table critical_findings enable row level security;

create policy "critical_findings: candidate read finalized own"
  on critical_findings for select to authenticated
  using (
    exists (
      select 1
      from reviews r
      join attempts a on a.id = r.attempt_id
      where r.id = critical_findings.review_id
        and r.finalized_at is not null
        and a.candidate_id = auth.uid()
    )
  );

create policy "critical_findings: reviewer read own"
  on critical_findings for select to authenticated
  using (
    current_user_role() = 'reviewer'
    and exists (
      select 1 from reviews r
      where r.id = critical_findings.review_id and r.reviewer_id = auth.uid()
    )
  );

create policy "critical_findings: reviewer write own draft"
  on critical_findings for insert to authenticated
  with check (
    current_user_role() = 'reviewer'
    and exists (
      select 1 from reviews r
      where r.id = critical_findings.review_id
        and r.reviewer_id = auth.uid()
        and r.finalized_at is null
    )
  );

create policy "critical_findings: reviewer edit own draft"
  on critical_findings for update to authenticated
  using (
    current_user_role() = 'reviewer'
    and exists (
      select 1 from reviews r
      where r.id = critical_findings.review_id
        and r.reviewer_id = auth.uid()
        and r.finalized_at is null
    )
  )
  with check (
    current_user_role() = 'reviewer'
    and exists (
      select 1 from reviews r
      where r.id = critical_findings.review_id
        and r.reviewer_id = auth.uid()
        and r.finalized_at is null
    )
  );

create policy "critical_findings: admin read"
  on critical_findings for select to authenticated
  using (current_user_role() = 'admin');

-- 3. Audit fields on reviews (frozen with the review once it is finalized)
alter table reviews
  add column weighted_score numeric(5, 2)
    check (weighted_score between 0 and 100),
  add column proposed_outcome text
    check (proposed_outcome in ('evidence_verified', 'needs_practice', 'review_pending')),
  add column outcome_reasons text[] not null default '{}',
  add column rubric_sha256 text
    check (rubric_sha256 ~ '^[0-9a-f]{64}$'),
  add column scenario_sha256 text
    check (scenario_sha256 ~ '^[0-9a-f]{64}$');
