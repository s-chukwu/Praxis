from __future__ import annotations

from typing import Any

# Must match the default actor in scripts/dev_signing_proxy.py, or the /docs examples get a 403.
DEMO_REVIEWER_ID = "11111111-1111-4111-8111-111111111111"

_CS01A_ARTIFACTS = [
    {
        "type": "customer_reply",
        "text": "I understand how frustrating this is. I can see one captured payment of RWF 42,000. "
        "The second entry is a pending authorization, not a settled charge.",
    },
    {
        "type": "internal_note",
        "text": "Escalated to billing to confirm whether the pending authorization settles. No refund promised.",
    },
]


def _score(scores: tuple[int, int, int, int], **extra: Any) -> dict[str, Any]:
    accuracy, judgment, safety, communication = scores
    return {
        "request_id": "5f0c2a7e-1d3b-4c8a-9e21-0a1b2c3d4e01",
        "attempt_id": "5f0c2a7e-1d3b-4c8a-9e21-0a1b2c3d4e02",
        "candidate_id": "5f0c2a7e-1d3b-4c8a-9e21-0a1b2c3d4e03",
        "reviewer_id": DEMO_REVIEWER_ID,
        "scenario_id": "CS-01-A",
        "scenario_version": "1.0.0",
        "rubric_version": "0.1.0",
        "dimensions": {
            "accuracy": {"score": accuracy, "evidence": ["one captured payment of RWF 42,000"]},
            "judgment": {"score": judgment, "evidence": ["Escalated to billing"]},
            "safety": {"score": safety, "evidence": ["No refund promised."]},
            "communication": {"score": communication, "evidence": ["I understand how frustrating this is."]},
        },
        "critical_findings": [],
        "artifacts": _CS01A_ARTIFACTS,
    } | extra


def _with_accuracy_quote(quote: str) -> dict[str, Any]:
    payload = _score((3, 3, 3, 3))
    payload["dimensions"]["accuracy"]["evidence"] = [quote]
    return payload


def _plan(scores: dict[str, int], **extra: Any) -> dict[str, Any]:
    return {
        "request_id": "7a2e9c41-6b0d-4f3e-8a15-1b2c3d4e5f01",
        "candidate_id": "7a2e9c41-6b0d-4f3e-8a15-1b2c3d4e5f02",
        "review_id": "7a2e9c41-6b0d-4f3e-8a15-1b2c3d4e5f03",
        "target_track": "operator",
        "role_standard_version": "1.0.0",
        "policy_version": "1.0.0",
        "reviewed_competencies": scores,
        "critical_safety_finding": False,
        "verified_completed_modules": [],
    } | extra


_STRONG = {"C1": 3, "C2": 3, "C3": 3, "C4": 3, "C5": 3, "C6": 3}

SCORE_EXAMPLES: dict[str, dict[str, Any]] = {
    "rubric_example": {
        "summary": "Rubric worked example: 3/3/4/3 scores 82.5",
        "value": _score((3, 3, 4, 3)),
    },
    "critical_veto": {
        "summary": "Perfect score with a confirmed critical finding: needs practice",
        "value": _score(
            (4, 4, 4, 4),
            critical_findings=[
                {
                    "event_type": "unauthorized_action",
                    "supporting_quote": "No refund promised.",
                    "policy_reference": "Simulated company policy 1.0.0, refunds",
                    "decision": "confirmed",
                }
            ],
        ),
    },
    "unverified_quote": {
        "summary": "Reviewer cites words the candidate never wrote: review pending",
        "value": _with_accuracy_quote("Refund confirmed for you."),
    },
    "below_threshold": {
        "summary": "3/3/3/2 scores 71.25: below the pilot threshold",
        "value": _score((3, 3, 3, 2)),
    },
}

PLAN_EXAMPLES: dict[str, dict[str, Any]] = {
    "example_a": {
        "summary": "Spec example A: experienced agent, weak privacy and AI verification",
        "value": _plan({"C1": 3, "C2": 4, "C3": 3, "C4": 2, "C5": 1, "C6": 3}),
    },
    "example_b": {
        "summary": "Spec example B: new entrant, gaps everywhere",
        "value": _plan({"C1": 2, "C2": 2, "C3": 1, "C4": 2, "C5": 1, "C6": 1}),
    },
    "example_c": {
        "summary": "Spec example C: strong operator on the builder track",
        "value": _plan(_STRONG | {"B1": 1, "B2": 1, "B3": 0}, target_track="builder"),
    },
    "all_strong": {
        "summary": "No gaps: orientation plus supervised challenge",
        "value": _plan(_STRONG),
    },
    "critical_finding": {
        "summary": "C4 is 4 but a critical safety finding forces M-SAFE",
        "value": _plan(_STRONG | {"C4": 4}, critical_safety_finding=True),
    },
    "incomplete": {
        "summary": "C5 never assessed: diagnostic incomplete, no plan",
        "value": _plan({k: v for k, v in _STRONG.items() if k != "C5"}),
    },
}


def _compare(before: tuple[int, int, int, int], after: tuple[int, int, int, int]) -> dict[str, Any]:
    names = ("accuracy", "judgment", "safety", "communication")

    def attempt(attempt_id: str, scenario_id: str, scores: tuple[int, int, int, int]) -> dict[str, Any]:
        return {
            "attempt_id": attempt_id,
            "scenario_id": scenario_id,
            "scenario_version": "1.1.0",
            "rubric_version": "0.2.0",
            "state": "reviewed",
            "dimensions": dict(zip(names, scores, strict=True)),
        }

    return {
        "request_id": "5f0c2a7e-1d3b-4c8a-9e21-0a1b2c3d4e31",
        "candidate_id": "5f0c2a7e-1d3b-4c8a-9e21-0a1b2c3d4e03",
        "baseline": attempt("5f0c2a7e-1d3b-4c8a-9e21-0a1b2c3d4e02", "CS-01-A", before),
        "final": attempt("5f0c2a7e-1d3b-4c8a-9e21-0a1b2c3d4e32", "CS-01-B", after),
    }


COMPARE_EXAMPLES: dict[str, dict[str, Any]] = {
    "improved": {
        "summary": "Training worked: every dimension held or improved",
        "value": _compare((2, 2, 2, 3), (3, 3, 4, 3)),
    },
    "safety_regressed": {
        "summary": "Safety dropped after training: flagged for a human",
        "value": _compare((3, 3, 4, 3), (3, 4, 2, 3)),
    },
}
