# Python assessment service contract and integration guide

## Boundary
`Go gateway -> internal Python service -> Go-owned persistence/Supabase`, with TLS/service authentication and identity/role checks in gateway and defense in depth in Python. The Python **reference implementation is local-only and in-memory**, with no auth, persistence, CV text extraction, or paid model. DO NOT expose it on a public interface or use live applicant data.

## Prototype endpoints
- `GET /health` health check.
- `GET /catalog` versioned competencies/course catalogue.
- `POST /assessment/plan`: takes `target_track`, candidate-confirmed `claimed_skills` and reviewer-confirmed `baseline_scores`; returns course list and gap diagnostics. **Do not make external job or applicant decisions from this response.**
- `POST /assessment/review`: authorized-reviewer-only in production; prototype calculates weighted score from human criterion ratings with evidence and flags pending validation.
- `POST /chatbot/simulate`: takes safe FAQ entries + message; returns constrained FAQ response or a human escalation; test harness only.

## Suggested durable tables (Go/database team)
`candidate_profile`, `claimed_skill` (source, status, original evidence), `job_role` (version), `competency`, `assessment_form`, `attempt`, `artifact`, `review` (criterion evidence, reviewer), `critical_finding` (human adjudication), `course` (version), `assignment` (reason + prerequisites), `lesson_progress`, `bot_configuration` (version), `bot_test_run`, `consent` (scope/time/revocation), `referral` (vacancy+manual approval), `audit_event` (append-only).

## Audit invariants
Store immutable role, scenario, rubric, curriculum and model/prompt version IDs, model output and reviewer score separately, and hash inputs. Score revisions need audit trail. Candidate cannot post reviewer score or employer eligibility. Never ship protected answer keys, policy overrides or internal prompts to browser. Train only on synthetic cases in pilot.

## Error semantics
422 malformed schema; 400 known but invalid decision graph/unsupported role; 403 on unauthorized reviewer in integrated system; 409 wrong version/duplicate finalized review; 429 quota; 503 optional AI suggestion unavailable, fail to human. Log request IDs only, redact PII.
