-- Migration 007: make the signup trigger robust.
-- Safe to run more than once. It only replaces the trigger function and trigger,
-- and it adds a profile for any existing user who is missing one.
--
-- Changes from migration 001:
--   1. every name is fully qualified (public.profiles, public.user_role) and
--      search_path is empty, so the trigger cannot fail on a missing schema path
--   2. "on conflict do nothing", so a repeated insert cannot break signup
--   3. a backfill for users created while the trigger was failing

-- 0. The live log showed: column "full_name" of relation "profiles" does not exist.
-- Restore every column the trigger and the app rely on. Each line does nothing
-- if the column is already there, so this is safe to run more than once.
alter table public.profiles add column if not exists full_name text;
alter table public.profiles add column if not exists role public.user_role not null default 'candidate';
alter table public.profiles add column if not exists created_at timestamptz not null default now();
alter table public.profiles add column if not exists updated_at timestamptz not null default now();

create or replace function public.handle_new_user()
returns trigger
language plpgsql
security definer
set search_path = ''
as $$
begin
  insert into public.profiles (id, full_name, role)
  values (new.id, new.raw_user_meta_data ->> 'full_name', 'candidate'::public.user_role)
  on conflict (id) do nothing;
  return new;
end;
$$;

drop trigger if exists on_auth_user_created on auth.users;

create trigger on_auth_user_created
  after insert on auth.users
  for each row execute function public.handle_new_user();

-- Backfill: every existing auth user gets a candidate profile if they have none.
insert into public.profiles (id, role)
select u.id, 'candidate'::public.user_role
from auth.users u
where not exists (select 1 from public.profiles p where p.id = u.id);
