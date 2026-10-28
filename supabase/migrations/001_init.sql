-- Migration 001: core schema for Praxis
-- Additive only (expand step). RLS is enabled here with no policies,
-- so every table is locked down until migration 002 adds the policies.

-- Types
create type user_role as enum ('candidate', 'reviewer', 'employer', 'admin');
create type consent_type as enum ('data_processing', 'cv_analysis', 'employer_sharing');
create type attempt_status as enum ('in_progress', 'submitted', 'under_review', 'reviewed');

-- Shared helper
create or replace function set_updated_at()
returns trigger language plpgsql as $$
begin
  new.updated_at = now();
  return new;
end;
$$;

-- Profiles (one row per Supabase auth user)
create table profiles (
  id uuid primary key references auth.users (id) on delete cascade,
  full_name text,
  role user_role not null default 'candidate',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create trigger profiles_updated_at
  before update on profiles
  for each row execute function set_updated_at();

-- Every new signup becomes a candidate. Role changes happen server side only.
create or replace function handle_new_user()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
begin
  insert into profiles (id, full_name, role)
  values (new.id, new.raw_user_meta_data ->> 'full_name', 'candidate');
  return new;
end;
$$;

create trigger on_auth_user_created
  after insert on auth.users
  for each row execute function handle_new_user();

-- Consents (append only: a new row records each grant or withdrawal)
create table consents (
  id uuid primary key default gen_random_uuid(),
  profile_id uuid not null references profiles (id) on delete cascade,
  type consent_type not null,
  granted boolean not null,
  recorded_at timestamptz not null default now()
);

create index consents_profile_type_idx on consents (profile_id, type, recorded_at desc);

-- Catalog (mirrors what the Python service returns from GET /internal/v1/catalog)
create table role_standards (
  id uuid primary key default gen_random_uuid(),
  track text not null,
  version text not null,
  created_at timestamptz not null default now(),
  unique (track, version)
);

create table competencies (
  id uuid primary key default gen_random_uuid(),
  role_standard_id uuid not null references role_standards (id) on delete cascade,
  code text not null,
  name text not null,
  unique (role_standard_id, code)
);

create table courses (
  id uuid primary key default gen_random_uuid(),
  code text not null unique,
  title text not null,
  created_at timestamptz not null default now()
);

-- Attempts (Go creates them and protects the answers)
create table attempts (
  id uuid primary key default gen_random_uuid(),
  candidate_id uuid not null references profiles (id) on delete cascade,
  target_track text not null,
  role_standard_version text not null,
  scenario_version text not null,
  rubric_version text not null,
  policy_version text not null,
  status attempt_status not null default 'in_progress',
  answers jsonb not null default '{}'::jsonb,
  submitted_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index attempts_candidate_idx on attempts (candidate_id);

create trigger attempts_updated_at
  before update on attempts
  for each row execute function set_updated_at();

-- Reviews (immutable once finalized)
create table reviews (
  id uuid primary key default gen_random_uuid(),
  attempt_id uuid not null references attempts (id) on delete cascade,
  reviewer_id uuid not null references profiles (id),
  reviewed_competencies jsonb not null default '{}'::jsonb,
  critical_safety_finding boolean not null default false,
  notes text,
  finalized_at timestamptz,
  created_at timestamptz not null default now()
);

-- Only one finalized review per attempt (Go maps a violation to HTTP 409)
create unique index reviews_one_finalized_per_attempt
  on reviews (attempt_id) where finalized_at is not null;

create or replace function block_finalized_review_changes()
returns trigger language plpgsql as $$
begin
  if old.finalized_at is not null then
    raise exception 'finalized reviews cannot be changed or deleted';
  end if;
  if tg_op = 'DELETE' then
    return old;
  end if;
  return new;
end;
$$;

create trigger reviews_immutable
  before update or delete on reviews
  for each row execute function block_finalized_review_changes();

-- Lock everything down until policies arrive in migration 002
alter table profiles enable row level security;
alter table consents enable row level security;
alter table role_standards enable row level security;
alter table competencies enable row level security;
alter table courses enable row level security;
alter table attempts enable row level security;
alter table reviews enable row level security;
