# Assessment engine verification and exit criteria

**Scope:** tests for AI/assessment responsibilities only. No claim that the full Praxis platform has been tested.

## Mandatory deterministic contract tests

| ID | Case | Expected result |
|---|---|---|
| A01 | CV claims a skill, no diagnostic | skill stays `claimed`; no verified score or course exemption |
| A02 | Missing required competency | `diagnostic_incomplete`; missing ≠ 0 |
| A03 | Reviewed C5 weak, C2 strong | AI-verification course assigned; unnecessary beginner communication excluded |
| A04 | C4 weak/critical privacy finding | safety remediation required regardless of overall weighted score |
| A05 | Module prerequisite DAG | prerequisite first, no cycle and no duplicate modules |
| A06 | Same reviewed input/policy version | same ordered allocation and reason codes |
| A07 | Unsupported track/scenario/rubric version | explicit error; no silent default |
| A08 | Duplicate retry with same idempotency key | no duplicate review or assessment state change |
| A09 | Prompt injection embedded in uploaded CV or ticket | remains data; cannot change instruction or score |
| A10 | AI scoring provider times out | no auto-pass; reviewer fallback |
| A11 | Candidate tries to provide reviewer ratings | rejected by authorization layer; service verifies caller scope |
| A12 | Revoked employer consent | assessment evidence not made employer-visible |
| A13 | Bot answers approved FAQ | cites exact allowed KB source, no invented policy |
| A14 | Bot receives refund execution request | no refund; creates escalation request |
| A15 | Human handoff ticket queued but not accepted | no false `accepted_by_human` claim |
| A16 | Human queue unavailable | safe status and unresolved case retained |
| A17 | Candidate leaks password/card data from bot | critical safety failure, manual review required |
| A18 | New rubric version during old attempt | historical score remains tied to original rubric/version |
| A19 | Final and baseline same answer text | detect leakage risk; parallel scenario form with matched competencies |
| A20 | Bot assessment hidden test requested by client | protected answer key never returned to candidate |

## Calibration before real assessments

1. Obtain role-task review from qualified customer-support supervisors.
2. Have at least two independent human reviewers score a varied, synthetic response set; record per-criterion agreement and adjudication outcomes.
3. Test and document baseline/final form difficulty and safety criticality; do not assume two scenarios are equivalent just because they sound similar.
4. Compare optional model-generated suggestions against human evidence; report errors per criterion and examples of unsafe model scores. Model agreement alone is **not** validation of job performance.
5. Confirm candidate appeal/second-review path and accessibility accommodations.
6. Obtain employer feedback on the relevance of the practical bot capstone for the **builder** track.

## Release blockers

Unauthenticated privileged grading, hidden keys in browser assets, raw CV in logs, contradictory rubric versions, silent missing scores, critical-safety override by aggregate score, bot falsely confirming handoff, no human verification for referral, or absent employer-sharing consent. Reject release until fixed and retested.

## Definition of implementation complete (future, not yet achieved)

AI service passes local unit tests, exposes frozen OpenAPI, passes Go consumer contract tests and negative-security cases, produces replayable scoring/plan results, and has documented reviewer calibration. This documentation package contains **no runnable AI service** and claims **no integration tests**.
