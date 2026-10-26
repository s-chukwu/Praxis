# Threat model, validation and release gates

## Threats
- CV and chatbot text may inject instructions into any LLM; never trust these documents as authority; parse into constrained fields and use strict schema, no tool access.
- Candidate impersonates reviewer / tampers with competency scores: RBAC, signed gateway claims, resource-level authorization, immutable attempts, RLS, integration tests. Prototype does not implement these, so must stay local.
- CV upload contains executable content, archive bomb, PII: whitelist PDF/DOCX MIME+magic, size/page limits, sandbox extraction, malware scanning, deletion policy; do not send raw file to LLM.
- Answer keys are exposed via frontend/static files: isolate assessor-only packets; distinct baseline/final forms; avoid including hidden keys in compiled frontend.
- Bot prompt injection and oversharing: test adversarial inputs; sandbox with static FAQ, no privileged actions and no external network.
- Reviewer bias / false pass: two-reviewer calibration; separate task evidence from identity; appeals; uncertainty; no AI-only readiness decision.
- Course allocation gaming: don't trust CV keyword match or candidate-submitted scores; audited baseline scoring; assignment reproducibility tests.
- Scenario memorization / retakes: different but blueprint-matched final forms, limited retakes and alternative forms.
- Unauthorized employer access: explicit revocable candidate consent; scoped and expiring permissions; separate employer views.

## Validation plan before partner-employer placement
1. Two supervisors complete a role-task analysis and confirm that chatbot config belongs to actual jobs offered.
2. Independent assessment authors produce equivalently difficult baseline/final forms; second reviewer checks ambiguity, relevance, leakage, accessibility.
3. Use ≥24 synthetic scoring examples and two independent reviewers, record agreement and disagreements; compare AI suggestions separately.
4. Test course allocation under missing evidence, claimed-but-unverified skills, score boundary, prerequisite cycles and adverse security finding.
5. Test chatbot with legitimate and hostile instructions; require human handoff on security/payment/account concerns; verify no live operations.
6. Pilot 10–15 candidates only with consent, usability feedback, paired results and limitations; cannot prove employment predictive validity.
7. Employer reviews evidence samples and explicitly states requirements; measure real referrals, interviews and hires; do not promise guaranteed employment.
8. Conduct privacy/legal review relevant to Rwanda and hiring locations, formal threat assessment, penetration tests and deletion/appeal exercises.

## Acceptance gates
- Pass pytest suite on every PR, lint/type checks, OpenAPI compatibility tests, dependency audit and CI secret scanning.
- Gate1: rubric/course/role versions signed by product+two domain reviewers.
- Gate2: human scoring evidence and critical event adjudication implemented server side.
- Gate3: CV and assessment authorization, immutable audit, explicit consent and deletion enforced.
- Gate4: bot sandbox isolation and injection regression tests.
- Gate5: documented evidence of scoring reliability, accessible assessments, complaint/appeal procedure, and employer-approved criteria.

## Reference standards (engineering context; not academic literature)
NIST AI RMF Generative AI Profile: https://doi.org/10.6028/NIST.AI.600-1
OWASP LLM Top 10: https://genai.owasp.org/llm-top-10/
OWASP Web Top 10: https://top10.owasp.org/2025/
EEOC employment-test guidance (US example, not Rwanda law): https://www.eeoc.gov/laws/guidance/employment-tests-and-selection-procedures
