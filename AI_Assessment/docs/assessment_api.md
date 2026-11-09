# Python assessment service contract and integration guide

## Boundary
`Go gateway -> internal Python service -> Go-owned persistence/Supabase`. Go authenticates users and checks permissions; Python verifies that every internal call was signed by Go and checks the actor role again as defense in depth. Python stores nothing. Only `skills/suggest` calls an LLM (Gemini); everything else is deterministic. DO NOT expose port 8000 publicly or use live applicant data.

## Implemented endpoints (milestones 1–5)
The live OpenAPI schema is at `http://localhost:8000/openapi.json` when `PRAXIS_ENV=local`; it is disabled in production.

| Method + path | Allowed actor | Purpose |
|---|---|---|
| `GET /health` | none (unsigned) | liveness |
| `GET /internal/v1/catalog` | any signed actor | every loaded version of role standard, rubric, scenarios, curriculum, allocation policy and attempt workflow; `current` pins for new attempts; SHA-256 per version. No assessor material. |
| `GET /internal/v1/scenarios/{id}/assessor-packet?version=X` | `reviewer` or `admin` | the scenario plus its grading key (checks, critical events). Never forward to a candidate. |
| `POST /internal/v1/diagnostics/score` | `reviewer` only; `reviewer_id` must equal the signed actor id and differ from `candidate_id` | weighted rubric score, critical-safety veto, evidence check, proposed outcome |
| `POST /internal/v1/diagnostics/compare` | `reviewer` or `system` | per-dimension change from a reviewed baseline to its reviewed parallel final form; flags any regression for a human |
| `POST /internal/v1/capstones/evaluate` | `candidate` (own `candidate_id`, `practice` only), `reviewer`, `system` | runs a candidate's bot config against the practice or held-out cases |
| `POST /internal/v1/learning-plans` | `reviewer` or `system` | skills gaps, safety gate and prerequisite-ordered modules from reviewed scores |
| `POST /internal/v1/skills/suggest` | `system` only | suggested competencies from CV text, each with a quote from the CV; the candidate confirms them |

**Attempt workflow.** `attempt_workflows` lists every allowed state change and who may make it (`reviewed` is locked except for an admin appeal; `voided` is final). Python publishes the table; our Go backend must enforce it on every attempt write, in the same transaction.

The old `GET /catalog` is removed. `POST /assessment/plan` is replaced by `POST /internal/v1/learning-plans`, which takes reviewed scores only, never CV claims.

### Service authentication (our Go backend must implement)
Every `/internal/v1` request carries:

```
X-Praxis-Timestamp: <unix seconds, within ±300 s of the service clock>
X-Praxis-Actor:     reviewer:<uuid> | candidate:<uuid> | admin:<uuid> | system
X-Praxis-Signature: hex(HMAC-SHA256(AI_SERVICE_SHARED_SECRET, canonical))

canonical = "v1\n" + timestamp + "\n" + METHOD + "\n" + path_with_raw_query + "\n" + actor + "\n" + hex(SHA-256(raw body))
```

The secret is at least 32 bytes and comes from `AI_SERVICE_SHARED_SECRET` (in `.env.example`; set a real value in your local `.env`).

**Replay window.** A signed request stays valid for ±300 s and carries no nonce, so a captured request can be replayed inside that window. That is acceptable only because every endpoint today is a pure computation that stores nothing. Before any endpoint changes state, add a single-use nonce to the canonical string and a shared store to reject repeats.

**Rotating the secret without downtime.** Python accepts `AI_SERVICE_SHARED_SECRET_PREVIOUS` alongside the current secret; our Go backend always signs with the current one.
1. Python: set `AI_SERVICE_SHARED_SECRET=<new>` and `AI_SERVICE_SHARED_SECRET_PREVIOUS=<old>`, deploy. Both now verify.
2. Go: switch to `<new>`, deploy.
3. Python: remove `AI_SERVICE_SHARED_SECRET_PREVIOUS`, deploy. `<old>` is now rejected.

The service refuses to start if the previous secret is set without a current one, equals it, or is under 32 bytes. Without it the service still serves `/health` locally but answers 503 `service_auth_not_configured` on internal routes, and refuses to start with `PRAXIS_ENV=production`. Optional `X-Request-ID` (8–64 chars `[A-Za-z0-9-]`) is echoed back and logged; otherwise one is generated.

### Assessor keys and Azure Key Vault
In production the service reads the assessor keys and the capstone held-out cases from Key Vault once, at startup, with the app's user-assigned managed identity. It never reads the JSON files baked into the image. Locally, with none of the variables below set, it reads the dev copies from the folder in `PRAXIS_PROTECTED_CONTENT_DIR` (use `AI_Assessment/tests/fixtures`, which holds `assessor_keys/` and `capstone_held_out/`); without it the service refuses to start. Neither folder is under `content/data`, so the image never contains them.

| Variable | Example | Notes |
|---|---|---|
| `AZURE_KEY_VAULT_URL` | `https://praxis-kv-26c0.vault.azure.net/` | must be `https://` |
| `ASSESSOR_KEYS_SECRET_NAME` | `assessor-keys` | |
| `CAPSTONE_HELD_OUT_SECRET_NAME` | `capstone-held-out` | same format, keyed by capstone version |
| `AZURE_CLIENT_ID` | client ID of `praxis-ai-identity` | selects the user-assigned identity |

All four are required when `PRAXIS_ENV=production`, and setting only some of them refuses to start. The secret is a JSON object mapping each scenario bank version to that version's `assessor_keys` document, as an object or as the file text: `{"1.0.0": {...}, "1.1.0": {...}}`. Every scenario bank the service loads needs an entry. A missing secret, denied access, invalid JSON or a missing version stops startup; the running revision keeps serving. Errors name the Azure error type and HTTP status, never the secret or Azure's message. A Key Vault secret holds at most 25 KB (about 4 KB per version today), so plan a second secret before the map approaches that. **Real keys go into Key Vault only, never into git.**

### POST /internal/v1/diagnostics/score
Request (all fields required unless noted; unknown fields are rejected):

```json
{
  "request_id": "uuid", "attempt_id": "uuid", "candidate_id": "uuid", "reviewer_id": "uuid",
  "scenario_id": "CS-01-A", "scenario_version": "1.1.0", "rubric_version": "0.2.0",
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
  "rubric": {"version": "0.2.0", "sha256": "..."}, "scenarios": {"version": "1.1.0", "sha256": "..."},
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

Retries (all POST endpoints): they are pure functions with no side effects, so retrying the same body returns the same result. Suggested Go client: 2 s timeout, at most 2 retries on connection errors or 5xx, never on 4xx. Duplicate finalized reviews are prevented by Go and the database unique index, not here.

### POST /internal/v1/diagnostics/compare
Send both reviewed attempts with the dimension scores the reviewer signed:

```json
{
  "request_id": "uuid", "candidate_id": "uuid",
  "baseline": {"attempt_id": "uuid", "scenario_id": "CS-01-A", "scenario_version": "1.1.0", "rubric_version": "0.2.0",
               "state": "reviewed", "dimensions": {"accuracy": 3, "judgment": 3, "safety": 4, "communication": 3}},
  "final":    {"attempt_id": "uuid", "scenario_id": "CS-01-B", "scenario_version": "1.1.0", "rubric_version": "0.2.0",
               "state": "reviewed", "dimensions": {"accuracy": 3, "judgment": 4, "safety": 2, "communication": 3}}
}
```

The response gives each dimension's `baseline`, `final`, `delta` and `direction` (`improved`/`unchanged`/`regressed`), both weighted scores, `weighted_delta`, `regressed_dimensions` and `requires_human_review` (true on any regression). Rules:
- `final.scenario_id` must be the baseline's parallel form, baseline stage first. The security pair (CS-04) never compares. Otherwise 422.
- Both attempts must be `reviewed` and scored on the same scenario bank and rubric, otherwise 409 `not_comparable`. After a rubric change, a reviewer re-scores the baseline under the new rubric first.
- Every rubric dimension is required; a missing one is 422, never a zero.

### POST /internal/v1/capstones/evaluate
Builder-track capstone. The candidate's bot is configuration only: `rules` (phrases that hand over or abstain, checked first), `intents` (phrases mapped to an approved knowledge-base `source_id`), a `fallback` and `handoff_after_failures`. There is no field for free-text replies and no code runs. Phrases match whole words only, never as patterns.

```json
{"request_id": "uuid", "attempt_id": "uuid", "candidate_id": "uuid", "suite": "practice",
 "config": {"capstone_version": "1.0.0", "rules": [...], "intents": [...], "fallback": "abstain", "handoff_after_failures": 2}}
```

The bot answers only with approved text, labels every reply as automated, never says a person joined before `accepted_by_human`, and masks card numbers and passwords in handoff tickets. The response has `passed`/`total`, `critical_failures` (case ids), per-category coverage, each case's turns and fail reasons (`wrong_action`, `wrong_source`, `wrong_priority`, `missed_injection`, `leaked_secret`), and `config_sha256`; store that hash with the result. `requires_human_review` is always true: a reviewer scores B1-B3.
- `practice` cases are public (also in `/catalog`); candidates may rerun them.
- `held_out` cases are server-only. Results report case id, category and outcome but never the customer messages or tickets. Our Go backend starts this run once per locked attempt; candidates get 403.

### POST /internal/v1/skills/suggest
Call only after the candidate's `cv_analysis` consent is recorded; pass its id as `consent_ref`.

```json
{"request_id": "uuid", "candidate_id": "uuid", "consent_ref": "uuid", "role_standard_version": "1.0.0", "text": "plain CV text, max 20000 chars"}
```

Emails and long numbers (phones, ID numbers) are masked before the text goes to Gemini. The model returns competency codes with quotes; I keep a suggestion only if the code is in the role standard and the quote appears word for word in the CV, one per competency. The response lists `suggestions` (`competency`, `evidence_quote`, `confidence`, `status: "suggested"`), how many model items were `dropped`, and the `model` used. Suggestions are claims for the candidate to confirm: they never score, waive a course or reach `learning-plans`.

- **Timeout:** an LLM call can take several seconds. Give this endpoint about 20 s and do not retry it; I already retry once inside.
- **503 `model_unavailable`** (`details.reason`: `timeout`, `http_429`, `unusable_answer`, `not_configured`...): send the CV to manual review. Never treat it as "no skills".
- Config: `GEMINI_API_KEY` (no key: always 503, the service still starts) and `GEMINI_MODEL` (default `gemini-2.5-flash`). CV text is never logged.

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
| 409 | `not_comparable` | compare attempts that are not both `reviewed`, or were scored on different scenario/rubric versions |
| 409 | `unsupported_version` | version not loaded, or scenario bound to a different rubric; `details.available` lists loaded versions |
| 413 | `payload_too_large` | body over 256 KiB |
| 422 | `invalid_request` | schema failure; `details.errors` gives `loc`, `msg`, `type` and never echoes input values |
| 422 | `invalid_input` | valid JSON that breaks a rubric/scenario rule (missing dimension, unknown event type, unexpected artifact, unknown scenario) |
| 503 | `service_auth_not_configured` | shared secret not set |
| 503 | `model_unavailable` | skill suggestions: the LLM timed out, was rate limited, answered badly or has no key |

Every response carries `Cache-Control: no-store`; it may hold candidate evidence or assessor keys.

Logs hold one JSON line per request (`request_id`, method, path, status, duration). Bodies are never logged.
