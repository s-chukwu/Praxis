# Praxis — AI / Assessment Engineering Specifications

**Owner:** Emmanuel (AI / Assessment Engineer)  
**Scope:** This folder contains **only specifications** for the Python AI/assessment service and the interfaces that other teams will consume. There is **no Python implementation**, frontend code, Go backend code, deployment infrastructure, or assignment of another engineer's implementation work in this deliverable.

**Status:** engineering draft for the seven-week customer-support pilot; not a validated hiring instrument. The workflow-builder track is a distinct additional specialization, not a required skill for every customer-support operator.

## Read in this order

1. [`docs/competency_standard.md`](docs/competency_standard.md) — which abilities the pilot measures; role boundaries (C1–C6, optionally B1–B3).
2. [`docs/scenario.md`](docs/scenario.md) — baseline and final simulated workplace tasks (contains assessor guidance: **restricted**).
3. [`docs/rubric.md`](docs/rubric.md) — score anchors, evidence requirements and safety gates.
4. [`docs/diagnostic_and_allocation.md`](docs/diagnostic_and_allocation.md) — map reviewed skill gaps to course modules and prerequisites.
5. [`docs/curriculum.md`](docs/curriculum.md) — course IDs, outcomes, prerequisite graph, training evidence.
6. [`docs/chatbot_lab.md`](docs/chatbot_lab.md) and [`docs/practical_capstone_and_handoff.md`](docs/practical_capstone_and_handoff.md) — builder-track capstone and verified human handoff.
7. [`docs/scoring_and_release_gates.md`](docs/scoring_and_release_gates.md) — qualification decision boundaries.
8. [`docs/assessment_api.md`](docs/assessment_api.md) and [`docs/integration_contract.md`](docs/integration_contract.md) — how other teams communicate with **our** Python service.
9. [`docs/security_and_validation.md`](docs/security_and_validation.md), [`docs/assessment_test_plan.md`](docs/assessment_test_plan.md) — threats and verification before implementation.

## Candidate journey (the assessment engine's viewpoint)

`CV/profile claim (unverified) → candidate confirms → diagnostic task → human-reviewed score → gap-to-course allocation → practice → separate final task → reviewed competency evidence → (builder track) verified bot-capstone → candidate-authorized employer-review eligibility`.

- **CV text never makes someone competent or job-ready by itself.**
- **Courses are assigned from reviewer-confirmed diagnostic evidence**, not from CV claims alone.
- **Completing a course does not qualify a candidate**; the final practical assessment and human review are mandatory.
- **Passing produces eligibility for employer consideration**, not guaranteed employment.

## What other engineers need to know

The only public integration documents for this module are [`docs/integration_contract.md`](docs/integration_contract.md) and [`docs/assessment_api.md`](docs/assessment_api.md). Frontend calls the Go API, **not Python directly**. Go owns users, identity, authorization, durable records, and employer consent; Python provides bounded assessment/plan/evaluation decisions and evidence. DevOps and QA use the same service contract for deployment and tests. This document **does not specify their implementations**.

## Document control and protected content

These specifications include sample scenarios and assessor-only notes. **Do not publish assessor keys, held-out test cases, sensitive prompts or secrets in a candidate-visible/public repository.** Before the first candidate accesses the platform, extract sensitive material to restricted storage and publish a candidate-safe scenario projection. Keep policy/rubric/scenario/course versions immutable for historical attempts.

**Source of truth:** role requirements = `competency_standard.md`; scenario facts = `scenario.md`; scores = `rubric.md`; course mapping = `diagnostic_and_allocation.md`; course descriptions = `curriculum.md`; referral rules = `scoring_and_release_gates.md`; service interface boundary = `integration_contract.md`. Conflicts must be resolved in a reviewed change before coding.

## Acceptance before Python implementation

The team must approve role scope, baseline/final scenario coverage, reviewer scoring anchors, safety-blocking rules, and the contract IDs; a real employer/support supervisor must review the occupational standard. The Python service implementation can begin against draft inputs, but **must not make autonomous hiring decisions**.

## Where Python code will go

Implement the Python assessment API under `service/` and its automated tests under `tests/`. These directories are currently placeholders; the Markdown specifications in `docs/` remain the contracts to implement.
