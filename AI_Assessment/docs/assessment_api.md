# Python assessment service contract and integration guide

## Boundary
`Go gateway -> internal Python service -> Go-owned persistence/Supabase`. Go authenticates users and checks permissions; Python verifies that every internal call was signed by Go and checks the actor role again as defense in depth. Python stores nothing and calls no model in milestones 1–2. DO NOT expose port 8000 publicly or use live applicant data.

## Implemented endpoints (milestones 1–3)
The live OpenAPI schema is at `http://localhost:8000/openapi.json` when `PRAXIS_ENV=local`; it is disabled in production.

| Method + path | Allowed actor | Purpose |
|---|---|---|
| `GET /health` | none (unsigned) | liveness |
| `GET /internal/v1/catalog` | any signed actor | every loaded version of role standard, rubric, scenarios, curriculum; `current` pins for new attempts; SHA-256 per version. No assessor material. |
| `GET /internal/v1/scenarios/{id}/assessor-packet?version=X` | `reviewer` or `admin` | the scenario plus its grading key (checks, critical events). Never forward to a candidate. |
| `POST /internal/v1/diagnostics/score` | `reviewer` only; `reviewer_id` must equal the signed actor id and differ from `candidate_id` | weighted rubric score, critical-safety veto, evidence check, proposed outcome |
| `POST /internal/v1/learning-plans` | `reviewer` or `system` | skills gaps, safety gate and prerequisite-ordered modules from reviewed scores |

The old `GET /catalog` is removed. `POST /assessment/plan` is replaced by `POST /internal/v1/learning-plans`, which takes reviewed scores only, never CV claims. The chatbot endpoint arrives in milestone 4.

### Service authentication (Go must implement)
Every `/internal/v1` request carries:

```
X-Praxis-Timestamp: <unix seconds, within ±300 s of the service clock>
X-Praxis-Actor:     reviewer:<uuid> | candidate:<uuid> | admin:<uuid> | system
X-Praxis-Signature: hex(HMAC-SHA256(AI_SERVICE_SHARED_SECRET, canonical))

canonical = "v1\n" + timestamp + "\n" + METHOD + "\n" + path_with_raw_query + "\n" + actor + "\n" + hex(SHA-256(raw body))
```

The secret is at least 32 bytes and comes from `AI_SERVICE_SHARED_SECRET` (in `.env.example`; set a real value in your local `.env`).

**Replay window.** A signed request stays valid for ±300 s and carries no nonce, so a captured request can be replayed inside that window. That is acceptable only because every endpoint today is a pure computation that stores nothing. Before any endpoint changes state, add a single-use nonce to the canonical string and a shared store to reject repeats.

**Rotating the secret without downtime.** Python accepts `AI_SERVICE_SHARED_SECRET_PREVIOUS` alongside the current secret; Go always signs with the current one.
1. Python: set `AI_SERVICE_SHARED_SECRET=<new>` and `AI_SERVICE_SHARED_SECRET_PREVIOUS=<old>`, deploy. Both now verify.
2. Go: switch to `<new>`, deploy.
3. Python: remove `AI_SERVICE_SHARED_SECRET_PREVIOUS`, deploy. `<old>` is now rejected.

The service refuses to start if the previous secret is set without a current one, equals it, or is under 32 bytes. Without it the service still serves `/health` locally but answers 503 `service_auth_not_configured` on internal routes, and refuses to start with `PRAXIS_ENV=production`. Optional `X-Request-ID` (8–64 chars `[A-Za-z0-9-]`) is echoed back and logged; otherwise one is generated.

### POST /internal/v1/diagnostics/score
Request (all fields required unless noted; unknown fields are rejected):

```json
{
  "request_id": "uuid", "attempt_id": "uuid", "candidate_id": "uuid", "reviewer_id": "uuid",
  "scenario_id": "CS-01-A", "scenario_version": "1.0.0", "rubric_version": "0.1.0",
  "dimensions": {
    "accuracy":      {"score": 3, "evidence": ["exact quote from an artifact"]},
    "judgment":      {"score": 3, "evidence": ["..."]},
    "safety":        {"score": 4, "evidence": ["..."]},
    "communication": {"score": 3, "evidence": ["..."]}
  },
  "critical_findings": [
    {"event_type": "unauthorized_action", "supporting_quote": "...", "policy_reference": "...", "decision": "confirmed | dismissed | pending"}
  ],
  "artifacts": [{"type": "customer_reply", "text": "candidate's submitted text"}]
}
```

- Scores are integers 0–4 (`"3"`, `2.5`, `true` are rejected). Every dimension of the named rubric must be present, each with 1–10 evidence quotes.
- Every quote (dimension evidence and `supporting_quote`) is 12–500 characters after trimming. Shorter quotes like "the" would match any answer and prove nothing.
- `critical_findings` is optional (max 10); `event_type` must be in the rubric's `critical_event_types`.
- `artifacts` (1–10, ≤ 20 000 chars each) must use types the scenario expects. Every quote must appear in them word for word (whitespace differences tolerated), or the result is held for review.
- Body limit 256 KiB.

Response 200:

```json
{
  "request_id": "uuid", "attempt_id": "uuid", "scenario_id": "CS-01-A", "stage": "baseline",
  "rubric": {"version": "0.1.0", "sha256": "..."}, "scenarios": {"version": "1.0.0", "sha256": "..."},
  "result": {
    "dimensions": [{"id": "accuracy", "score": 3, "weight_bp": 3000, "weighted_points": 22.5}],
    "weighted_score": 82.5, "pilot_threshold": 75.0, "pilot_threshold_is_validated": false,
    "critical_veto": false, "confirmed_critical_events": [], "pending_critical_events": [],
    "evidence_issues": [], "proposed_outcome": "evidence_verified",
    "outcome_reasons": ["meets_pilot_threshold"], "score_source": "human_review", "requires_signoff": true
  }
}
```

`proposed_outcome` precedence: any pending critical finding or unverified quote → `review_pending`; else any confirmed critical finding → `needs_practice` (veto, whatever the score); else `weighted_score >= pilot_threshold` → `evidence_verified`, otherwise `needs_practice`. It is a proposal for the reviewer to sign, not a decision.

Retries (both POST endpoints): they are pure functions with no side effects, so retrying the same body returns the same result. Suggested Go client: 2 s timeout, at most 2 retries on connection errors or 5xx, never on 4xx. Duplicate finalized reviews are prevented by Go and the database unique index, not here.

### POST /internal/v1/learning-plans
Request (unknown fields, including `claimed_skills`, are rejected):

```json
{
  "request_id": "uuid", "candidate_id": "uuid", "review_id": "uuid",
  "target_track": "operator | builder",
  "role_standard_version": "1.0.0", "policy_version": "1.0.0",
  "reviewed_competencies": {"C1": 3, "C2": 4, "C3": 3, "C4": 2, "C5": 1, "C6": 3},
  "critical_safety_finding": false,
  "verified_completed_modules": ["M-ROLE"]
}
```

- `reviewed_competencies`: strict integers 0–4, only codes in the track (`operator`: C1–C6; `builder`: C1–C6 and B1–B3). Send exactly what the finalized review contains; do not fill gaps with zeros.
- `critical_safety_finding`: strict boolean, true when any adjudicated critical finding exists on the review.
- `verified_completed_modules` (optional): modules Go has recorded as completed and verified. They satisfy prerequisites but never cancel a gap the latest review still shows.

Response 200 (`status` is `plan_ready` or `diagnostic_incomplete`; both are 200):

```json
{
  "request_id": "uuid", "candidate_id": "uuid", "review_id": "uuid", "target_track": "operator",
  "role_standard": {"version": "1.0.0", "sha256": "..."},
  "policy": {"version": "1.0.0", "sha256": "..."},
  "curriculum": {"version": "1.0.0", "sha256": "..."},
  "result": {
    "status": "plan_ready", "threshold": 3, "missing_competencies": [],
    "demonstrated": [{"code": "C1", "reviewed_score": 3}],
    "gaps": [{"code": "C4", "reviewed_score": 2}, {"code": "C5", "reviewed_score": 1}],
    "safety_gate": {"required": true, "module": "M-SAFE", "triggers": ["safety_competency_below_threshold"]},
    "modules": [
      {"order": 1, "course_id": "M-ROLE", "title": "...", "covers": ["C1", "C3"],
       "reasons": [{"code": "role_orientation"}], "previously_completed": false},
      {"order": 2, "course_id": "M-SAFE", "title": "...", "covers": ["C4"],
       "reasons": [{"code": "competency_gap", "competency": "C4", "reviewed_score": 2}, {"code": "safety_gate"}],
       "previously_completed": false}
    ],
    "waived_prerequisites": [],
    "reassessment_required": true
  }
}
```

Rules (allocation policy 1.0.0):
1. Any track competency missing → `diagnostic_incomplete`, `modules: []`. Missing never counts as zero.
2. Score below `threshold` (3) → the module(s) in the policy's remediation map (C1, C5 → M-VERIFY; C2 → M-COMM; C3, C6 → M-TRIAGE; C4 → M-SAFE; B1–B3 → M-BOT-KB, M-BOT-HANDOFF, M-BOT-TEST).
3. C4 below 3 or `critical_safety_finding` → M-SAFE with reason `safety_gate`, whatever else applies.
4. M-ROLE is always assigned (`role_orientation`) unless already completed and verified.
5. M-CHALLENGE is added last after any C1–C6 gap, or alone with M-ROLE when there are no gaps. A builder with only B gaps does not repeat operator content.
6. Prerequisites are added (`prerequisite`, with `required_by`) unless completed and verified, or every competency they cover was reviewed at 3 or above; skipped ones are listed in `waived_prerequisites`.
7. Modules come in the curriculum's fixed prerequisite-first order; the same review always gives the same plan.

`409 unsupported_version` when the policy version isn't loaded or is bound to a different role standard; `422 invalid_input` for competencies outside the track or unknown course IDs.

## Suggested durable tables (Go/database team)
`candidate_profile`, `claimed_skill` (source, status, original evidence), `job_role` (version), `competency`, `assessment_form`, `attempt`, `artifact`, `review` (criterion evidence, reviewer), `critical_finding` (human adjudication), `course` (version), `assignment` (reason + prerequisites), `lesson_progress`, `bot_configuration` (version), `bot_test_run`, `consent` (scope/time/revocation), `referral` (vacancy+manual approval), `audit_event` (append-only).

## Audit invariants
Store immutable role, scenario, rubric, curriculum and model/prompt version IDs, model output and reviewer score separately, and hash inputs. Score revisions need audit trail. Candidate cannot post reviewer score or employer eligibility. Never ship protected answer keys, policy overrides or internal prompts to browser. Assessor keys live apart from scenarios (`content/data/assessor_keys/`), and the service refuses to start unless each scenario bank has exactly one matching key set. **Before real candidates**, move that directory out of git into a server-only store; the pilot keys are already public in `scenario.md`. Our Go backend must also stop a reviewer from being assigned their own attempt; Python only rejects `reviewer_id` = `candidate_id`. Train only on synthetic cases in pilot.

## Error semantics
Every error has one shape: `{"error": {"code", "message", "request_id", "details"}}`.

| Status | `code` | When |
|---|---|---|
| 401 | `unauthenticated` | missing, stale or wrong signature; malformed actor. Checked before the body schema. |
| 403 | `forbidden` | actor role not allowed, `reviewer_id` ≠ signed actor, or `reviewer_id` = `candidate_id` |
| 404 / 405 | `not_found` / `method_not_allowed` | unknown route |
| 409 | `unsupported_version` | version not loaded, or scenario bound to a different rubric; `details.available` lists loaded versions |
| 413 | `payload_too_large` | body over 256 KiB |
| 422 | `invalid_request` | schema failure; `details.errors` gives `loc`, `msg`, `type` and never echoes input values |
| 422 | `invalid_input` | valid JSON that breaks a rubric/scenario rule (missing dimension, unknown event type, unexpected artifact, unknown scenario) |
| 503 | `service_auth_not_configured` | shared secret not set |
| 429 / 503 | reserved | quota and model outage, from milestone 5 |

Every response carries `Cache-Control: no-store`; it may hold candidate evidence or assessor keys.

Logs hold one JSON line per request (`request_id`, method, path, status, duration). Bodies are never logged.
