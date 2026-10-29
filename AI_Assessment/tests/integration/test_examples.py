from __future__ import annotations

import pytest

from praxis_assessment.api.examples import DEMO_REVIEWER_ID, PLAN_EXAMPLES, SCORE_EXAMPLES

EXPECTED_OUTCOMES = {
    "rubric_example": "evidence_verified",
    "critical_veto": "needs_practice",
    "unverified_quote": "review_pending",
    "below_threshold": "needs_practice",
}
EXPECTED_STATUS = {name: "plan_ready" for name in PLAN_EXAMPLES} | {"incomplete": "diagnostic_incomplete"}


def test_every_example_has_an_expectation() -> None:
    assert EXPECTED_OUTCOMES.keys() == SCORE_EXAMPLES.keys()
    assert EXPECTED_STATUS.keys() == PLAN_EXAMPLES.keys()


@pytest.mark.parametrize("name", SCORE_EXAMPLES)
def test_score_example(call, name: str) -> None:
    response = call(
        "POST", "/internal/v1/diagnostics/score", SCORE_EXAMPLES[name]["value"], actor=f"reviewer:{DEMO_REVIEWER_ID}"
    )
    assert response.status_code == 200, response.text
    assert response.json()["result"]["proposed_outcome"] == EXPECTED_OUTCOMES[name]


@pytest.mark.parametrize("name", PLAN_EXAMPLES)
def test_plan_example(call, name: str) -> None:
    response = call(
        "POST", "/internal/v1/learning-plans", PLAN_EXAMPLES[name]["value"], actor=f"reviewer:{DEMO_REVIEWER_ID}"
    )
    assert response.status_code == 200, response.text
    assert response.json()["result"]["status"] == EXPECTED_STATUS[name]
