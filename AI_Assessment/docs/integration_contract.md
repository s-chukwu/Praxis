# AI / Assessment service — integration boundary

**Owner:** AI / Assessment Engineering. **Audience:** other teams integrating with this service. **Status:** proposed API interface; **not yet implemented**. This is a contract draft, not a deployed OpenAPI document. Consult `assessment_api.md` for existing prototype conventions; these internal endpoints supersede prototype-only paths **when the Go/Python contract is approved**.

## Ownership

| Boundary | AI/assessment Python service | External responsibility (interface expectation only) |
|---|---|---|
| CV | Extract suggested skills from already-authorized text, evidence spans and uncertainty; do **not** grant competency | Go obtains upload/consent, parses/stores originals securely; UI asks candidate to confirm claims |
| Baseline | Provide versioned assessment schema and optional draft feedback; compute reviewed scores | Go creates attempts, protects answers, authorizes reviewers, saves immutable review |
| Course allocation | Given **reviewed scores**, choose skill-gap modules deterministically and explain why | Go verifies review, persists plan; UI displays modules |
| Final exam | Calculate rubric result; flag missing evidence and critical safety issues | Trained reviewer decides verified competency |
| Bot capstone | Evaluate restricted bot configuration on server-held test cases and handoff event evidence | Go handles persistent config, queue ticket and actual human acknowledgment |
| Referrals | Return decision evidence only; never hire or contact employers | Go checks consent, employer permissions and final referral workflow |

## Direction of calls

`Next.js browser → Go API → internal Python assessment service → response to Go → Go persistence → frontend`. The Python service has **no direct browser endpoint in production**. Mutual authorization or short-lived service tokens, TLS, request tracing and rate limits are mandatory before live traffic. Do not pass CV raw text in logs or domain events.

## Proposed internal endpoint responsibilities (versioned)

| Path | Input | Output | Side effect | Caller |
|---|---|---|---|---|
| `POST /internal/v1/skills/suggest` | sanitized candidate text, consent ref, trace ID | suggested `skill_id`, evidence quote, confidence/unknown flag | none | authenticated Go service |
| `POST /internal/v1/diagnostics/score` | attempt ID, scenario/rubric versions, criterion evidence and authorized reviewed ratings | criterion breakdown, safety flags, feedback | none; Go stores review | authenticated Go service |
| `POST /internal/v1/learning-plans` | track, versions, **reviewed** C/B competencies, critical findings | gap IDs, ordered course IDs, reason codes, policy version | none | authenticated Go service |
| `POST /internal/v1/capstones/evaluate` | versioned bot config reference and locked attempt ID | held-out category results, safety blocks, reviewer-required flag | no real-world customer action | authenticated Go service |
| `GET /internal/v1/catalog` | optional version | role competencies, course/assessment metadata without hidden keys | none | authenticated Go service |

### Shared request envelope

```json
{
  "request_id": "uuid",
  "candidate_id": "uuid",
  "attempt_id": "uuid",
  "target_track": "operator",
  "role_standard_version": "1.0.0",
  "scenario_version": "1.0.0",
  "rubric_version": "0.1.0",
  "policy_version": "1.0.0"
}
```

Each operation carries the subset it requires, plus a signed caller context out-of-band. Responses MUST include the corresponding versions, `request_id`, stable reason codes and `requires_human_review` where relevant. Avoid making personal data mandatory for score computation.

### Example allocation (illustrative values, NOT a final endpoint schema)

```json
{
  "target_track": "operator",
  "review_id": "review-123",
  "reviewed_competencies": {"C1":3,"C2":4,"C3":3,"C4":2,"C5":1,"C6":3},
  "critical_safety_finding": false,
  "role_standard_version": "1.0.0",
  "allocation_policy_version": "1.0.0"
}
```

Response must include `diagnostic_complete`, `assigned_modules` (including prerequisite order), `gap_competencies` and per-module `reason_codes`. If a required skill is missing, return `diagnostic_incomplete` and **do not assume a zero score**. Include request and version IDs once contract is frozen.

## Errors and non-negotiable behavior

- `422`: schema invalid; `409`: stale assessment/policy version or duplicate finalized review; `403`: service authorization failed; `429`: quota; `503`: external model temporarily unavailable.
- On model outage, **fail to reviewer/manual assessment**, never invent scores or convert missing data to pass.
- Requests may be retried: idempotency key + replay-safe response for same input/version; incompatible replay returns conflict.
- No PII, raw CV, full candidate answers, API keys, or hidden rubric keys in ordinary application logs or emitted events.
- Treat CV, candidate response, course uploads and model text as untrusted; protect against prompt injection and tool abuse.
- Eligibility is an **audited human decision** outside Python; Python may report evidence but cannot transition directly to `hired`.

## Contract freeze checklist

Before simultaneous implementation, the AI and backend engineers agree and publish: OpenAPI schema, canonical identifiers/enums, role/policy versioning, error examples, signed auth model, timeout/retry limits, trace format, API mock fixtures, and consumer contract tests. A Markdown example is not a substitute for machine-readable OpenAPI.
