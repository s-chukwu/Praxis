# Diagnostics and personalized course allocation
Status v1.0 / all examples synthetic

## Source hierarchy
1. CV uploaded and parsed: **claimed** signals, no verification.
2. Candidate profile and self-rating: **claimed**.
3. Baseline task results on C1–C6 and a short B1–B3 design exercise: **observed** evidence, reviewer pending.
4. Authorized, documented human review: **verified-on-this-task**; not general employability.
5. Employer job-analysis requirement map: **required standards**, not a guess from CV content.

Never infer psychological traits, age, disability, ethnicity, or 'AI replacement risk' from CV. Never mark courses completed because an LLM guessed an existing skill. CV parser outputs `skill_id`, `quoted_evidence`, `source='cv'`, `status='claimed'`, and explicit unknowns; candidate confirms suggested entries.

## Allocation algorithm (MVP)
- Require candidate consent and target role; profile claims inform diagnostic selection but do not bypass baseline.
- Administer fixed diagnostic tasks spanning C1–C6, and optionally B1–B3 for workflow-builder track.
- Human reviewer grades 0..4 per competency, records evidence, safety issues and assessment versions.
- For each required skill: assign remediation if verified baseline <3; require safety course whenever C4<3 or an adjudicated critical finding exists.
- Select modules in a **directed acyclic prerequisite graph**; include prerequisites once, stable order; don't automatically assign all modules.
- Learning tasks include low-stakes practice and feedback. After practice, candidate gets separate, parallel final assessment; no released final answer keys.
- If all baseline skills >=3, allocate a compact supervised challenge plus mandatory role-policy orientation, **not** instant employer eligibility.
- If evidence is missing, assign a diagnostic, not label deficiency.
- On final, human verifies C1–C6 and B1–B3 if applicable; failures receive targeted retakes using fresh forms. Rate-limit retakes and offer appeals/accommodations.

## Allocation examples
**A. Experienced support agent:** CV claims communication and CRM. Reviewed baseline C2=4, C6=3, C5=1, C4=2. Assign AI verification and privacy remediation, their prerequisites, then final tasks; don't assign generic communication basics by default.
**B. New entrant:** CV has little evidence; baseline C1=2,C2=2,C3=1,C4=2,C5=1,C6=1. Assign workflow, safety, verification, triage and communication modules; may require extended practice before reassessment.
**C. Strong operator who wants builder track:** baseline C1–C6 >=3, B1=1,B2=1,B3=0. Assign bot knowledge setup, human handoff, bot regression testing; no need to repeat all operator content.

## Fairness and quality
Rubric measures task behavior, not spelling proficiency unrelated to role. Allow accommodations, use equivalent tests, and audit course allocation by skill rather than background. Missing data never means zero by default. Training module and rubric versions are immutable per attempt.

## Important limitation
The reference Python service accepts **reviewer-confirmed** competencies; it doesn't independently score natural-language answers. An LLM-assisted suggestion adapter, rubric evidence extraction, and reviewer UI are *future integration*, not secretly simulated by the prototype.
