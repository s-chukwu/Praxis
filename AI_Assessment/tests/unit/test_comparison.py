from __future__ import annotations

import uuid
from typing import Any

import pytest

from praxis_assessment.config import DEFAULT_CONTENT_DIR
from praxis_assessment.content import load_registry
from praxis_assessment.domain.comparison import compare_attempts
from praxis_assessment.errors import InvalidInput, NotComparable
from praxis_assessment.schemas.comparison import CompareRequest, Comparison

REGISTRY = load_registry(DEFAULT_CONTENT_DIR)
RUBRIC = REGISTRY.rubric("0.2.0").content
BANK = REGISTRY.scenario_bank("1.1.0").content


def attempt(scenario_id: str, scores: dict[str, int], **overrides: Any) -> dict[str, Any]:
    return {
        "attempt_id": str(uuid.uuid4()),
        "scenario_id": scenario_id,
        "scenario_version": "1.1.0",
        "rubric_version": "0.2.0",
        "state": "reviewed",
        "dimensions": scores,
    } | overrides


BEFORE = {"accuracy": 2, "judgment": 2, "safety": 3, "communication": 3}
AFTER = {"accuracy": 3, "judgment": 4, "safety": 3, "communication": 2}


def compare(baseline: dict[str, Any], final: dict[str, Any]) -> Comparison:
    request = CompareRequest.model_validate(
        {"request_id": str(uuid.uuid4()), "candidate_id": str(uuid.uuid4()), "baseline": baseline, "final": final}
    )
    return compare_attempts(
        RUBRIC, BANK.scenario(request.baseline.scenario_id), BANK.scenario(request.final.scenario_id), request
    )


def test_reports_each_dimension_and_the_weighted_change() -> None:
    result = compare(attempt("CS-01-A", BEFORE), attempt("CS-01-B", AFTER))
    assert [(d.id, d.delta, d.direction) for d in result.dimensions] == [
        ("accuracy", 1, "improved"),
        ("judgment", 2, "improved"),
        ("safety", 0, "unchanged"),
        ("communication", -1, "regressed"),
    ]
    assert (result.baseline_weighted_score, result.final_weighted_score) == (61.25, 77.5)
    assert result.weighted_delta == 16.25


def test_any_regression_needs_a_human() -> None:
    assert compare(attempt("CS-01-A", BEFORE), attempt("CS-01-B", AFTER)).requires_human_review is True
    steady = compare(attempt("CS-01-A", BEFORE), attempt("CS-01-B", BEFORE))
    assert steady.regressed_dimensions == ()
    assert steady.requires_human_review is False


@pytest.mark.parametrize(
    ("baseline", "final", "message"),
    [
        ("CS-01-A", "CS-02-B", "not the parallel form"),
        ("CS-01-B", "CS-01-A", "baseline form followed by a final form"),  # swapped stages
        ("CS-01-A", "CS-01-A", "baseline form followed by a final form"),  # same form twice
        ("CS-04-A", "CS-04-B", "baseline form followed by a final form"),  # security regression pair
    ],
)
def test_only_a_baseline_and_its_twin_compare(baseline: str, final: str, message: str) -> None:
    with pytest.raises(InvalidInput, match=message):
        compare(attempt(baseline, BEFORE), attempt(final, AFTER))


@pytest.mark.parametrize("state", ["submitted", "under_review", "review_pending"])
def test_unreviewed_attempt_is_not_comparable(state: str) -> None:
    with pytest.raises(NotComparable, match="not reviewed"):
        compare(attempt("CS-01-A", BEFORE), attempt("CS-01-B", AFTER, state=state))


def test_attempts_on_different_rubrics_are_not_comparable() -> None:  # A18
    old = attempt("CS-01-A", BEFORE, scenario_version="1.0.0", rubric_version="0.1.0")
    with pytest.raises(NotComparable, match="same scenario bank and rubric"):
        compare(old, attempt("CS-01-B", AFTER))


def test_missing_dimension_is_rejected_not_zeroed() -> None:
    partial = {k: v for k, v in AFTER.items() if k != "safety"}
    with pytest.raises(InvalidInput) as exc:
        compare(attempt("CS-01-A", BEFORE), attempt("CS-01-B", partial))
    assert exc.value.details["missing"] == ["safety"]
