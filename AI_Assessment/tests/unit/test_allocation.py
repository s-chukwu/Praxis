from __future__ import annotations

import random
from typing import Any

import pytest
from pydantic import ValidationError

from praxis_assessment.config import DEFAULT_CONTENT_DIR
from praxis_assessment.content import load_registry
from praxis_assessment.domain.allocation import allocate
from praxis_assessment.errors import InvalidInput
from praxis_assessment.schemas.learning_plan import LearningPlan, LearningPlanRequest
from support import OPERATOR_STRONG, plan_payload

REGISTRY = load_registry(DEFAULT_CONTENT_DIR)
STANDARD = REGISTRY.role_standard("1.0.0").content
POLICY = REGISTRY.allocation_policy("1.0.0").content
CURRICULUM = REGISTRY.curriculum("1.0.0").content


def plan(scores: dict[str, int] | None = None, **overrides: Any) -> LearningPlan:
    return allocate(STANDARD, POLICY, CURRICULUM, LearningPlanRequest.model_validate(plan_payload(scores, **overrides)))


def course_ids(result: LearningPlan) -> list[str]:
    return [m.course_id for m in result.modules]


def reasons_for(result: LearningPlan, course_id: str) -> list[str]:
    return [r.code for m in result.modules if m.course_id == course_id for r in m.reasons]


# The three worked examples in diagnostic_and_allocation.md. The spec only lists some scores for
# example A, so I fill the rest at the threshold.


def test_example_a_experienced_agent_gets_verification_and_privacy_only() -> None:
    result = plan({"C1": 3, "C2": 4, "C3": 3, "C4": 2, "C5": 1, "C6": 3})
    assert course_ids(result) == ["M-ROLE", "M-SAFE", "M-VERIFY", "M-CHALLENGE"]
    assert "M-COMM" not in course_ids(result)  # A03: strong communication is not retaught
    assert [g.code for g in result.gaps] == ["C4", "C5"]


def test_example_b_new_entrant_gets_every_operator_module() -> None:
    result = plan({"C1": 2, "C2": 2, "C3": 1, "C4": 2, "C5": 1, "C6": 1})
    assert course_ids(result) == ["M-ROLE", "M-COMM", "M-SAFE", "M-VERIFY", "M-TRIAGE", "M-CHALLENGE"]
    assert reasons_for(result, "M-VERIFY") == ["competency_gap", "competency_gap"]  # C1 and C5


def test_example_c_strong_operator_on_builder_track_skips_operator_content() -> None:
    result = plan(OPERATOR_STRONG | {"B1": 1, "B2": 1, "B3": 0}, target_track="builder")
    assert course_ids(result) == ["M-ROLE", "M-BOT-KB", "M-BOT-HANDOFF", "M-BOT-TEST"]
    assert [(w.course_id, w.reason) for w in result.waived_prerequisites] == [("M-SAFE", "competencies_demonstrated")]


def test_no_gaps_gives_orientation_plus_supervised_challenge() -> None:
    result = plan(OPERATOR_STRONG)
    assert result.status == "plan_ready"
    assert course_ids(result) == ["M-ROLE", "M-CHALLENGE"]
    assert reasons_for(result, "M-ROLE") == ["role_orientation"]
    assert result.gaps == ()


def test_reasons_point_at_the_reviewed_score() -> None:
    result = plan(OPERATOR_STRONG | {"C2": 1})
    comm = next(m for m in result.modules if m.course_id == "M-COMM")
    assert [(r.code, r.competency, r.reviewed_score) for r in comm.reasons] == [("competency_gap", "C2", 1)]


def test_missing_competency_is_incomplete_not_zero() -> None:  # A02
    scores = dict(OPERATOR_STRONG)
    del scores["C5"]
    result = plan(scores)
    assert result.status == "diagnostic_incomplete"
    assert result.missing_competencies == ("C5",)
    assert result.modules == ()
    assert "C5" not in [g.code for g in result.gaps]


def test_builder_track_requires_bot_scores() -> None:
    result = plan(OPERATOR_STRONG, target_track="builder")
    assert result.status == "diagnostic_incomplete"
    assert result.missing_competencies == ("B1", "B2", "B3")


def test_critical_finding_forces_safety_module_whatever_c4_says() -> None:  # A04
    result = plan(OPERATOR_STRONG | {"C4": 4}, critical_safety_finding=True)
    assert "M-SAFE" in course_ids(result)
    assert reasons_for(result, "M-SAFE") == ["safety_gate"]
    assert result.safety_gate.triggers == ("critical_safety_finding",)


def test_low_safety_score_records_both_gap_and_gate() -> None:
    result = plan(OPERATOR_STRONG | {"C4": 1})
    assert reasons_for(result, "M-SAFE") == ["competency_gap", "safety_gate"]
    assert result.safety_gate.triggers == ("safety_competency_below_threshold",)


def test_critical_finding_on_builder_track_blocks_safety_waiver() -> None:
    result = plan(OPERATOR_STRONG | {"B1": 1, "B2": 3, "B3": 3}, target_track="builder", critical_safety_finding=True)
    assert course_ids(result) == ["M-ROLE", "M-SAFE", "M-BOT-KB"]
    assert result.waived_prerequisites == ()


def test_prerequisite_is_waived_when_its_competency_is_demonstrated() -> None:
    result = plan(OPERATOR_STRONG | {"B1": 3, "B2": 3, "B3": 1}, target_track="builder")
    assert course_ids(result) == ["M-ROLE", "M-BOT-TEST"]
    assert [(w.course_id, w.required_by) for w in result.waived_prerequisites] == [("M-BOT-HANDOFF", "M-BOT-TEST")]


def test_prerequisite_is_pulled_in_when_not_demonstrated() -> None:
    # Under policy 1.0.0 every weak prerequisite is already a gap and M-ROLE is mandatory, so this path
    # only shows up once a policy drops the mandatory orientation. I still want it covered before it does.
    policy = POLICY.model_copy(update={"mandatory_modules": ()})
    request = LearningPlanRequest.model_validate(plan_payload(OPERATOR_STRONG | {"C1": 2}))
    result = allocate(STANDARD, policy, CURRICULUM, request)
    assert course_ids(result) == ["M-ROLE", "M-VERIFY", "M-CHALLENGE"]
    role = result.modules[0]
    assert [(r.code, r.required_by) for r in role.reasons] == [("prerequisite", "M-VERIFY")]


def test_verified_completed_module_satisfies_prerequisites() -> None:
    result = plan(OPERATOR_STRONG | {"C5": 1}, verified_completed_modules=["M-ROLE", "M-SAFE"])
    assert course_ids(result) == ["M-VERIFY", "M-CHALLENGE"]
    assert [(w.course_id, w.reason) for w in result.waived_prerequisites] == [
        ("M-ROLE", "verified_completed"),
        ("M-SAFE", "verified_completed"),
    ]


def test_completed_module_is_reassigned_while_the_gap_persists() -> None:
    result = plan(OPERATOR_STRONG | {"C5": 1}, verified_completed_modules=["M-VERIFY"])
    verify = next(m for m in result.modules if m.course_id == "M-VERIFY")
    assert verify.previously_completed is True
    assert [r.code for r in verify.reasons] == ["competency_gap"]


def test_competencies_outside_the_track_are_rejected() -> None:
    with pytest.raises(InvalidInput) as exc:
        plan(OPERATOR_STRONG | {"B1": 2})
    assert exc.value.details == {"unexpected": ["B1"], "track": "operator"}


def test_unknown_completed_module_is_rejected() -> None:
    with pytest.raises(InvalidInput) as exc:
        plan(verified_completed_modules=["M-NOPE"])
    assert exc.value.details == {"unknown": ["M-NOPE"]}


def test_cv_claims_cannot_be_submitted() -> None:  # A01
    with pytest.raises(ValidationError, match="Extra inputs are not permitted"):
        LearningPlanRequest.model_validate(plan_payload(claimed_skills=["C5"]))


@pytest.mark.parametrize("bad", [5, -1, 2.5, "3", True])
def test_scores_are_strict_integers(bad: Any) -> None:
    with pytest.raises(ValidationError):
        LearningPlanRequest.model_validate(plan_payload(OPERATOR_STRONG | {"C1": bad}))


def test_critical_flag_is_a_strict_boolean() -> None:
    with pytest.raises(ValidationError):
        LearningPlanRequest.model_validate(plan_payload(critical_safety_finding="false"))


def test_same_review_gives_identical_plan_regardless_of_key_order() -> None:  # A06
    scores = {"C1": 2, "C2": 4, "C3": 1, "C4": 2, "C5": 1, "C6": 3}
    shuffled = dict(reversed(scores.items()))
    assert plan(scores).model_dump_json() == plan(shuffled).model_dump_json()


def test_every_plan_respects_prerequisites_and_has_no_duplicates() -> None:  # A05
    # Exhaustive over the builder track would be 5^9 plans; a seeded sample covers the shapes I care about.
    rng = random.Random(20261029)
    codes = STANDARD.required("builder")
    for _ in range(300):
        scores = {code: rng.randint(0, 4) for code in codes}
        result = plan(scores, target_track="builder", critical_safety_finding=rng.random() < 0.2)
        ids = course_ids(result)
        assert len(ids) == len(set(ids))
        assert [m.order for m in result.modules] == list(range(1, len(ids) + 1))
        satisfied = {w.course_id for w in result.waived_prerequisites}
        for position, course_id in enumerate(ids):
            for prerequisite in CURRICULUM.course(course_id).prerequisites:
                assert prerequisite in satisfied or prerequisite in ids[:position], (scores, ids)
        if "M-CHALLENGE" in ids:
            assert ids[-1] == "M-CHALLENGE"
        for gap in result.gaps:
            assert set(POLICY.remediation[gap.code]) <= set(ids), (scores, ids)
