# AI assessment engineering implementation sequence (future Python work)

This is **Emmanuel's sequence only**. Other engineers control their own feature designs.

1. **Freeze role blueprint and identifiers** — confirm C1–C6, operator vs builder, scenario forms, rubric anchors, module prerequisites, version IDs.
2. **Model schemas** — typed requests/responses, enums, reasons, evidence source provenance and state transition guards. Write schema tests before LLM integration.
3. **Deterministic assessment/review helper** — accept authorized human criterion evidence, calculate weighted results and critical-safety flags; never mark candidate hired.
4. **Deterministic allocator** — assign courses only from reviewed baseline evidence; topologically sort prerequisites and explain reasons; handle missing inputs.
5. **Skills suggestion adapter** — integrate file-text extraction received from Go; suggest skills with exact evidence; return unverified suggestions; add prompt injection tests.
6. **LLM feedback adapter** — schema-validated optional suggestions; evaluator must point to candidate evidence and scenario facts; fail safely on model outage.
7. **Builder capstone evaluator** — locked configuration, server-side held-out tests, safe FAQ grounding, explicit handoff state/acknowledgment and critical vetoes; never run arbitrary candidate code in MVP.
8. **Integrate through OpenAPI** — mock Go caller, auth and timeouts; repeatable consumer tests and trace IDs.
9. **Calibrate** — human assessor comparison, scenario difficulty review, fairness review, reported failures, documented employer validation.

## Principal engineering trade-offs

- **Rule-based course allocation** over opaque LLM recommendation: explainable and reproducible, but less adaptive; revisit with measured evidence.
- **Human signoff** over autonomous job-readiness decisions: slower and costlier but necessary at this stage.
- **Constrained configuration-based chatbot lab** over arbitrary executable code: safer and achievable in seven weeks, but doesn't verify professional programming skill. Describe graduates as *workflow configurators* until deeper engineering tests exist.
- **Go as persistence owner** over Python writing directly to shared database: stronger ownership and auditability, extra network call.
- **Versioned, fixed scenario bank** over generated graded questions: more comparable scores, smaller item pool and leakage risk; rotate held-out forms under access control.

## Non-goals of this folder

No frontend implementation, Go server/database migrations, infrastructure-as-code, employer marketplace, or real placements. This folder is specification-only until the Python project is intentionally started.
