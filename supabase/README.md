# Supabase

Postgres, Auth, and row level security for Praxis.

## Migrations

Migration files are in `supabase/migrations/`. Apply them in numeric order, starting with `001_init.sql`. Apply each file in the Supabase SQL Editor, one at a time. Do not edit an applied migration or run a file twice.

The set covers:

- Core tables, the signup trigger, and immutable reviews.
- Row level security policies.
- Attempt states and transitions.
- Scores, scenarios, findings, and audit records.
- Learning plans, audit log, and idempotency keys.
- Signup trigger fixes and baseline constraints.
- Learner profiles.
- Baseline reset support.
- Low scores without evidence quotes.

Two files use the `011_` prefix. Confirm which one is applied in each environment before adding the next migration.

## Roles

- `candidate`: sees only their own data.
- `reviewer`: sees submitted attempts and manages their own reviews.
- `employer`: no access yet.
- `admin`: read access to everything.

## Rules

1. Never edit an applied migration. Add a new numbered file instead.
2. Every new table must enable RLS and include its policies.
3. The Go backend uses the service role key, which skips RLS, so the Go code must check permissions itself.
4. Never commit keys or passwords. Use `.env`.
