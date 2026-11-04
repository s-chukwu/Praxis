-- Migration 003: add the attempt states published in the AI service catalog.
-- Kept in its own file because Postgres does not let a new enum value be used
-- in the same transaction that adds it. Migration 004 uses these values.

alter type attempt_status add value if not exists 'assigned' before 'in_progress';
alter type attempt_status add value if not exists 'review_pending' after 'under_review';
alter type attempt_status add value if not exists 'voided';
