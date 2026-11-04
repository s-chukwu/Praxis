-- Migration 006: learning plans, append-only audit log, idempotency keys.
-- Additive only. Go writes to all of these with the service role.
-- Candidates and staff only get read access through row level security.

-- 1. Learning plans (one plan per finalized review and policy version)
create table learning_plans (
  id uuid primary key default gen_random_uuid(),
  candidate_id uuid not null references profiles (id) on delete cascade,
  review_id uuid not null references reviews (id),
  request_id uuid,
  target_track text not null,
  status text not null check (status in ('plan_ready', 'diagnostic_incomplete')),
  role_standard_version text not null,
  policy_version text not null,
  curriculum_version text not null,
  role_standard_sha256 text check (role_standard_sha256 ~ '^[0-9a-f]{64}$'),
  policy_sha256 text check (policy_sha256 ~ '^[0-9a-f]{64}$'),
  curriculum_sha256 text check (curriculum_sha256 ~ '^[0-9a-f]{64}$'),
  safety_gate_required boolean not null default false,
  reassessment_required boolean not null default true,
  result jsonb not null,
  created_at timestamptz not null default now(),
  unique (review_id, policy_version)
);

create index learning_plans_candidate_idx on learning_plans (candidate_id);

-- Modules of a plan in prerequisite order. course_code matches courses.code,
-- but is not a foreign key yet because the catalog is not seeded.
create table learning_plan_modules (
  id uuid primary key default gen_random_uuid(),
  plan_id uuid not null references learning_plans (id) on delete cascade,
  position int not null check (position >= 1),
  course_code text not null,
  title text not null,
  covers text[] not null default '{}',
  reasons jsonb not null default '[]'::jsonb,
  previously_completed boolean not null default false,
  completed_at timestamptz,
  verified_at timestamptz,
  verified_by uuid references profiles (id),
  unique (plan_id, position),
  unique (plan_id, course_code)
);

alter table learning_plans enable row level security;
alter table learning_plan_modules enable row level security;

create policy "learning_plans: candidate read own"
  on learning_plans for select to authenticated
  using (candidate_id = auth.uid());

create policy "learning_plans: reviewer read own reviews"
  on learning_plans for select to authenticated
  using (
    current_user_role() = 'reviewer'
    and exists (
      select 1 from reviews r
      where r.id = learning_plans.review_id and r.reviewer_id = auth.uid()
    )
  );

create policy "learning_plans: admin read"
  on learning_plans for select to authenticated
  using (current_user_role() = 'admin');

create policy "plan_modules: candidate read own"
  on learning_plan_modules for select to authenticated
  using (
    exists (
      select 1 from learning_plans p
      where p.id = learning_plan_modules.plan_id and p.candidate_id = auth.uid()
    )
  );

create policy "plan_modules: reviewer read own reviews"
  on learning_plan_modules for select to authenticated
  using (
    current_user_role() = 'reviewer'
    and exists (
      select 1
      from learning_plans p
      join reviews r on r.id = p.review_id
      where p.id = learning_plan_modules.plan_id and r.reviewer_id = auth.uid()
    )
  );

create policy "plan_modules: admin read"
  on learning_plan_modules for select to authenticated
  using (current_user_role() = 'admin');

-- 2. Audit log (append only). Never put raw CV text, full candidate answers,
-- keys or hidden rubric answers in details.
create table audit_events (
  id bigint generated always as identity primary key,
  occurred_at timestamptz not null default now(),
  actor_id uuid,
  actor_role text not null,
  action text not null,
  entity_type text not null,
  entity_id uuid,
  request_id uuid,
  details jsonb not null default '{}'::jsonb
);

create index audit_events_entity_idx on audit_events (entity_type, entity_id, occurred_at desc);
create index audit_events_actor_idx on audit_events (actor_id, occurred_at desc);

create or replace function block_audit_changes()
returns trigger language plpgsql as $$
begin
  raise exception 'audit_events is append only';
end;
$$;

create trigger audit_events_append_only
  before update or delete on audit_events
  for each row execute function block_audit_changes();

create trigger audit_events_no_truncate
  before truncate on audit_events
  for each statement execute function block_audit_changes();

alter table audit_events enable row level security;

create policy "audit_events: admin read"
  on audit_events for select to authenticated
  using (current_user_role() = 'admin');

-- 3. Idempotency keys. Same key and same request hash replays the stored answer.
-- Same key with a different hash is a conflict (HTTP 409). Service role only.
create table idempotency_keys (
  scope text not null,
  key text not null,
  request_hash text not null check (request_hash ~ '^[0-9a-f]{64}$'),
  actor_id uuid,
  response_status int,
  response_body jsonb,
  created_at timestamptz not null default now(),
  expires_at timestamptz not null default now() + interval '24 hours',
  primary key (scope, key)
);

create index idempotency_keys_expiry_idx on idempotency_keys (expires_at);

alter table idempotency_keys enable row level security;
