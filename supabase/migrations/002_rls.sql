-- Migration 002: row level security policies
-- The Go backend connects with the service role, which bypasses RLS.
-- These policies protect direct access (anon and authenticated keys).
-- Anonymous users get no policies, so they see nothing.

-- Helper: current user's role. Security definer avoids policy recursion on profiles.
create or replace function current_user_role()
returns user_role
language sql
stable
security definer
set search_path = public
as $$
  select role from profiles where id = auth.uid();
$$;

-- profiles
create policy "profiles: read own"
  on profiles for select to authenticated
  using (id = auth.uid());

create policy "profiles: staff read all"
  on profiles for select to authenticated
  using (current_user_role() in ('reviewer', 'admin'));

create policy "profiles: update own, role locked"
  on profiles for update to authenticated
  using (id = auth.uid())
  with check (id = auth.uid() and role = current_user_role());

-- consents (append only: no update or delete policies)
create policy "consents: read own"
  on consents for select to authenticated
  using (profile_id = auth.uid());

create policy "consents: insert own"
  on consents for insert to authenticated
  with check (profile_id = auth.uid());

-- catalog (read only for any signed in user, writes only through the service role)
create policy "role_standards: read"
  on role_standards for select to authenticated
  using (true);

create policy "competencies: read"
  on competencies for select to authenticated
  using (true);

create policy "courses: read"
  on courses for select to authenticated
  using (true);

-- attempts
create policy "attempts: candidate read own"
  on attempts for select to authenticated
  using (candidate_id = auth.uid());

create policy "attempts: candidate create own"
  on attempts for insert to authenticated
  with check (candidate_id = auth.uid() and status = 'in_progress');

create policy "attempts: candidate edit while open"
  on attempts for update to authenticated
  using (candidate_id = auth.uid() and status = 'in_progress')
  with check (candidate_id = auth.uid() and status in ('in_progress', 'submitted'));

create policy "attempts: staff read submitted"
  on attempts for select to authenticated
  using (
    current_user_role() = 'admin'
    or (current_user_role() = 'reviewer'
        and status in ('submitted', 'under_review', 'reviewed'))
  );

-- reviews
create policy "reviews: candidate read finalized own"
  on reviews for select to authenticated
  using (
    finalized_at is not null
    and exists (
      select 1 from attempts a
      where a.id = reviews.attempt_id and a.candidate_id = auth.uid()
    )
  );

create policy "reviews: reviewer read own"
  on reviews for select to authenticated
  using (current_user_role() = 'reviewer' and reviewer_id = auth.uid());

create policy "reviews: reviewer create"
  on reviews for insert to authenticated
  with check (current_user_role() = 'reviewer' and reviewer_id = auth.uid());

create policy "reviews: reviewer edit own draft"
  on reviews for update to authenticated
  using (current_user_role() = 'reviewer' and reviewer_id = auth.uid() and finalized_at is null)
  with check (current_user_role() = 'reviewer' and reviewer_id = auth.uid());

create policy "reviews: admin read"
  on reviews for select to authenticated
  using (current_user_role() = 'admin');
