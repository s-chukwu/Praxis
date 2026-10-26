# Praxis role standard: AI-assisted customer support operator (CS-AI-01)
Status: pilot hypothesis. Must be signed off by employers/supervisors. Not a professional qualification.

## What the role entails
Read case context and policy; classify urgency; compose accurate and empathetic messages; use an AI drafting assistant without accepting hallucinations; handle restricted information safely; decide when a human must approve/refund/review; configure and test a bounded FAQ chatbot and handoff. A candidate need not program in Python to be competent at this role. The course may expose tool configuration and debugging without teaching model training.

## Skill IDs and assessment tasks
| ID | Skill | Evidence | Assessment | Critical? |
|---|---|---|---|---|
| C1 | Fact checking and comprehension | Confirmed vs unknown distinction | Billing reply | No |
| C2 | Communication | Empathy and accurate next steps | Angry-customer response | No |
| C3 | Workflow and escalation | Approval dependency, correct routing | Billing, account takeover | Yes when unauthorized action |
| C4 | Privacy / AI safety | Redaction, ignores injection, no secrets | Privacy and injection scenarios | Yes |
| C5 | AI verification | Reject unsupported AI-generated claims | Rewrite flawed assistant response | Yes when material deception |
| C6 | Triage/documentation | Category, priority, internal note | Four-ticket batch | Yes on security compromise |
| B1 | Chatbot knowledge configuration | Curated FAQ+source policy, explicit unknowns | Configured FAQ bot lab | No |
| B2 | Handoff and escalation | Correct rule-based human transfer | Bot lab critical scenarios | Yes |
| B3 | Chatbot testing & trade-off analysis | Test log, failure analysis and iterative correction | Bot lab challenge cases | No |

## Proficiency levels
0 = not observed / unsafe; 1 = attempts but substantial gaps; 2 = works with close supervision; 3 = can perform controlled simulation with minimal coaching; 4 = robust performance, justified escalation, documented edge cases. These levels require assessor evidence, not self-rated confidence.

## Two different outcomes
**Foundational support track** tests C1–C6. **AI workflow-builder endorsement (pilot only)** additionally tests B1–B3. Don't require chatbot design of all customer service roles without evidence from a job analysis. A candidate may complete foundational training without earning the builder endorsement.

## Employer fit
Each participating employer supplies role-specific tasks, approved policies, tools permitted, required capabilities and reviewer. Candidate's readiness profile must describe *verified evidence in simulations* and limitations; employer decides hiring independently. A pass is not a job guarantee.
