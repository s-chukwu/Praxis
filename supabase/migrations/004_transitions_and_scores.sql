-- Migration 004: attempt state machine, per-dimension scores, no self-review.
-- Run only after 003 has succeeded (it uses the new enum values).

-- 1. Attempt state machine (copy of attempt states version 1.0.0 from the AI catalog)
create table attempt_transitions (
  version text not null default '1.0.0',
  source attempt_status not null,
  target attempt_status not null,
  allowed_roles text[] not null,
  primary key (version, source, target)
);

insert into attempt_transitions (source, target, allowed_roles) values
  ('assigned',       'in_progress',    '{candidate}'),
  ('in_progress',    'submitted',      '{candidate,system}'),
  ('submitted',      'under_review',   '{reviewer}'),
  ('under_review',   'review_pending', '{reviewer}'),
  ('review_pending', 'under_review',   '{reviewer}'),
  ('under_review',   'reviewed',       '{reviewer}'),
  ('reviewed',       'under_review',   '{admin}'),
  ('assigned',       'voided',         '{admin}'),
  ('in_progress',    'voided',         '{admin}'),
  ('submitted',      'voided',         '{admin}'),
  ('under_review',   'voided',         '{admin}'),
  ('review_pending', 'voided',         '{admin}');

alter table attempt_transitions enable row level security;

create policy "attempt_transitions: read"
  on attempt_transitions for select to authenticated
  using (true);

alter table attempts alter column status set default 'assigned';

-- Only legal moves are allowed. Role checks apply to direct client access.
-- The Go backend uses the service role (no auth.uid), so Go must check roles itself.
create or replace function enforce_attempt_transition()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
declare
  allowed text[];
  actor text;
begin
  if new.status is not distinct from old.status then
    return new;
  end if;

  select allowed_roles into allowed
  from attempt_transitions
  where version = '1.0.0' and source = old.status and target = new.status;

  if allowed is null then
    raise exception 'illegal attempt transition from % to %', old.status, new.status;
  end if;

  if auth.uid() is not null then
    actor := current_user_role()::text;
    if actor is null or not (actor = any (allowed)) then
      raise exception 'role % cannot move an attempt from % to %', actor, old.status, new.status;
    end if;
  end if;

  return new;
end;
$$;

create trigger attempts_transition
  before update of status on attempts
  for each row execute function enforce_attempt_transition();

-- Attempt policies updated for the new states
drop policy "attempts: candidate create own" on attempts;
create policy "attempts: candidate create own"
  on attempts for insert to authenticated
  with check (candidate_id = auth.uid() and status = 'assigned');

drop policy "attempts: candidate edit while open" on attempts;
create policy "attempts: candidate edit while open"
  on attempts for update to authenticated
  using (candidate_id = auth.uid() and status in ('assigned', 'in_progress'))
  with check (candidate_id = auth.uid() and status in ('assigned', 'in_progress', 'submitted'));

drop policy "attempts: staff read submitted" on attempts;
create policy "attempts: staff read submitted"
  on attempts for select to authenticated
  using (
    current_user_role() = 'admin'
    or (current_user_role() = 'reviewer'
        and status in ('submitted', 'under_review', 'review_pending', 'reviewed'))
  );

-- 2. A reviewer can never review their own attempt
create or replace function block_self_review()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
  if exists (
    select 1 from attempts
    where id = new.attempt_id and candidate_id = new.reviewer_id
  ) then
    raise exception 'a reviewer cannot review their own attempt';
  end if;
  return new;
end;
$$;

create trigger reviews_no_self_review
  before insert or update of attempt_id, reviewer_id on reviews
  for each row execute function block_self_review();

-- 3. Per-dimension scores (scores 0 to 4, one to ten evidence quotes)
create table review_dimension_scores (
  id uuid primary key default gen_random_uuid(),
  review_id uuid not null references reviews (id) on delete cascade,
  dimension text not null,
  score smallint not null check (score between 0 and 4),
  evidence text[] not null check (cardinality(evidence) between 1 and 10),
  created_at timestamptz not null default now(),
  unique (review_id, dimension)
);

-- Scores freeze when their review is finalized
create or replace function block_finalized_dimension_changes()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
declare
  finalized timestamptz;
begin
  select finalized_at into finalized
  from reviews
  where id = coalesce(new.review_id, old.review_id);

  if finalized is not null then
    raise exception 'scores of a finalized review cannot be changed';
  end if;

  if tg_op = 'DELETE' then
    return old;
  end if;
  return new;
end;
$$;

create trigger review_dimension_scores_immutable
  before insert or update or delete on review_dimension_scores
  for each row execute function block_finalized_dimension_changes();

alter table review_dimension_scores enable row level security;

create policy "dimension_scores: candidate read finalized own"
  on review_dimension_scores for select to authenticated
  using (
    exists (
      select 1
      from reviews r
      join attempts a on a.id = r.attempt_id
      where r.id = review_dimension_scores.review_id
        and r.finalized_at is not null
        and a.candidate_id = auth.uid()
    )
  );

create policy "dimension_scores: reviewer read own"
  on review_dimension_scores for select to authenticated
  using (
    current_user_role() = 'reviewer'
    and exists (
      select 1 from reviews r
      where r.id = review_dimension_scores.review_id and r.reviewer_id = auth.uid()
    )
  );

create policy "dimension_scores: reviewer write own draft"
  on review_dimension_scores for insert to authenticated
  with check (
    current_user_role() = 'reviewer'
    and exists (
      select 1 from reviews r
      where r.id = review_dimension_scores.review_id
        and r.reviewer_id = auth.uid()
        and r.finalized_at is null
    )
  );

create policy "dimension_scores: reviewer edit own draft"
  on review_dimension_scores for update to authenticated
  using (
    current_user_role() = 'reviewer'
    and exists (
      select 1 from reviews r
      where r.id = review_dimension_scores.review_id
        and r.reviewer_id = auth.uid()
        and r.finalized_at is null
    )
  )
  with check (
    current_user_role() = 'reviewer'
    and exists (
      select 1 from reviews r
      where r.id = review_dimension_scores.review_id
        and r.reviewer_id = auth.uid()
        and r.finalized_at is null
    )
  );

create policy "dimension_scores: admin read"
  on review_dimension_scores for select to authenticated
  using (current_user_role() = 'admin');
