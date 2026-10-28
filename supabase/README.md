# Supabase

Database setup for Praxis (Postgres, Auth and row level security).

## Applying migrations

1. Open the Supabase dashboard and go to the SQL Editor.
2. Paste one migration file and click Run.
3. Run files in number order, starting with 001. Never run a file twice.

## Migrations

1. 001_init.sql: core tables, signup trigger, immutable reviews.
2. 002_rls.sql: row level security policies.

## Roles

1. candidate: sees only their own data.
2. reviewer: sees submitted attempts and manages their own reviews.
3. employer: no access yet.
4. admin: read access to everything.

## Rules

1. Never edit an applied migration. Add a new numbered file instead.
2. Every new table must enable RLS and include its policies.
3. The Go backend uses the service role key, which skips RLS, so Go must check permissions itself.
4. Never commit keys or passwords. Use .env.