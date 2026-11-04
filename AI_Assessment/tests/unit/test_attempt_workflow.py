from __future__ import annotations

from typing import Any

import pytest
from pydantic import ValidationError

from praxis_assessment.config import DEFAULT_CONTENT_DIR
from praxis_assessment.content import load_registry
from praxis_assessment.content.models import AttemptWorkflow

WORKFLOW = load_registry(DEFAULT_CONTENT_DIR).attempt_workflow("1.0.0").content
STATES = ["assigned", "in_progress", "submitted", "under_review", "review_pending", "reviewed", "voided"]
ROLES = ["candidate", "system", "reviewer", "admin"]


@pytest.mark.parametrize(
    ("source", "target", "role"),
    [
        ("assigned", "in_progress", "candidate"),
        ("in_progress", "submitted", "system"),  # timebox expired
        ("under_review", "reviewed", "reviewer"),
        ("reviewed", "under_review", "admin"),  # appeal
        ("review_pending", "voided", "admin"),
    ],
)
def test_expected_transitions_are_allowed(source: str, target: str, role: str) -> None:
    assert WORKFLOW.allows(source, target, role)


def test_reviewed_is_locked_except_for_an_admin_appeal() -> None:
    exits = [(t, r) for t in STATES for r in ROLES if WORKFLOW.allows("reviewed", t, r)]
    assert exits == [("under_review", "admin")]


def test_candidates_never_touch_review_states() -> None:
    reachable = {t for s in STATES for t in STATES if WORKFLOW.allows(s, t, "candidate")}
    assert reachable == {"in_progress", "submitted"}


def test_nothing_leaves_voided() -> None:
    assert not any(WORKFLOW.allows("voided", t, r) for t in STATES for r in ROLES)


def _workflow(**overrides: Any) -> dict[str, Any]:
    return WORKFLOW.model_dump(mode="json") | overrides


@pytest.mark.parametrize(
    ("transitions", "message"),
    [
        ([{"source": "voided", "target": "assigned", "roles": ["admin"]}], "voided is final"),
        ([{"source": "assigned", "target": "assigned", "roles": ["admin"]}], "to itself"),
        ([{"source": "submitted", "target": "in_progress", "roles": ["admin", "admin"]}], "duplicate role"),
        ([{"source": "assigned", "target": "in_progress", "roles": ["root"]}], "roles"),
    ],
)
def test_broken_workflow_is_rejected(transitions: list[dict[str, Any]], message: str) -> None:
    with pytest.raises(ValidationError, match=message):
        AttemptWorkflow.model_validate(_workflow(transitions=[*WORKFLOW.model_dump()["transitions"], *transitions]))


def test_unreachable_state_is_rejected() -> None:
    kept = [t for t in WORKFLOW.model_dump()["transitions"] if t["target"] != "review_pending"]
    with pytest.raises(ValidationError, match=r"unreachable from assigned: \['review_pending'\]"):
        AttemptWorkflow.model_validate(_workflow(transitions=kept))
