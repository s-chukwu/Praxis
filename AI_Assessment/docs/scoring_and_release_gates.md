# Scoring, course allocation and referral review — decision policy
## Evidence classes
`claimed` (CV only), `suggested` (machine extraction), `candidate_confirmed` (still unverified), `observed` (attempted task), `reviewed` (authorized assessor), `qualified_for_referral_review` (authorized role signoff). Never collapse these.

## Form fairness
Use fixed, reviewed scenarios and different baseline/final parallel forms with equivalent skills/difficulty, held-out builder cases and item-level traceability. Annotate scenario IDs, form revision, rubrics, trained assessors, allowed tool use, completion time and accessibility accommodations. Run inter-rater checks with independent reviews; publish disagreement rate and manual moderation. Small 10–15-person pilot cannot validate population-level employment prediction.

## Scoring
Existing rubric has 4 criterion anchors 0–4 with weights accuracy 30%, judgment 25%, safety/privacy 30%, communication 15%. Report criterion evidence and weighted score. **70% or any single number is a pilot candidate threshold, not an industry-validated pass mark**. Before employment use, set per-role minimums and critical safety veto rules via employer review and documented pilot calibration. Do not use CVs to grant scores. Human must sign scores. For builder: require hidden-case safe-handoff criterion with independent evidence.

## Course selection
Complete every required baseline competency; if incomplete return `diagnostic_incomplete`, never guess. Map scores below the published remediation threshold to modules; recursively insert prerequisites; avoid duplicates; return reason codes, competency gaps and course-policy version. Re-evaluation after course completion is mandatory. Candidate can appeal mistaken input/score.

## Referral eligibility checklist
All mandatory scenarios submitted; no unresolved critical safety failure; adequate reviewed competency coverage and (builder track) a passed verified chatbot capstone and acknowledged human handoff; human reviewer signs versioned decision; candidate explicitly consents to employer sharing. Result: **eligible for employer review/referral**, never guaranteed hire. Maintain counts qualified → opted in → referred → interviewed → offered → hired with timestamps and employer evidence.

## Release blockers
Unauthenticated reviewer action; raw CV exposed; frontend contains hidden assessment answer key; a passing result after critical safety failure; skipped human approval; failing isolation test; non-idempotent duplicate referral; unversioned score; unresolved high/critical defect; bot claims human accepted when only queued; missing restore test for production. Stop release and report status rather than softening any of these.
