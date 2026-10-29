from __future__ import annotations

import copy
from typing import Any

import pytest
from pydantic import ValidationError

from praxis_assessment.config import DEFAULT_CONTENT_DIR
from praxis_assessment.content import load_registry
from praxis_assessment.domain.scoring import score_attempt
from praxis_assessment.errors import InvalidInput
from praxis_assessment.schemas.scoring import MIN_QUOTE_CHARS, ScoreRequest, ScoreResult
from support import score_payload, with_scores

REGISTRY = load_registry(DEFAULT_CONTENT_DIR)
RUBRIC = REGISTRY.rubric("0.1.0").content
SCENARIO = REGISTRY.scenario_bank("1.0.0").content.scenario("CS-01-A")


def run(payload: dict[str, Any]) -> ScoreResult:
    return score_attempt(RUBRIC, SCENARIO, ScoreRequest.model_validate(payload))


def finding(decision: str, quote: str = "No refund promised.", event: str = "unauthorized_action") -> dict[str, str]:
    return {"event_type": event, "supporting_quote": quote, "policy_reference": "policy 1.0.0 s2", "decision": decision}


def test_rubric_worked_example_scores_82_5() -> None:
    result = run(score_payload())
    assert result.weighted_score == 82.5
    assert [(d.id, d.weighted_points) for d in result.dimensions] == [
        ("accuracy", 22.5),
        ("judgment", 18.75),
        ("safety", 30.0),
        ("communication", 11.25),
    ]
    assert result.proposed_outcome == "evidence_verified"
    assert result.score_source == "human_review"
    assert result.requires_signoff is True


@pytest.mark.parametrize(("score", "expected"), [(0, 0.0), (1, 25.0), (2, 50.0), (3, 75.0), (4, 100.0)])
def test_uniform_scores_map_linearly(score: int, expected: float) -> None:
    result = run(with_scores(accuracy=score, judgment=score, safety=score, communication=score))
    assert result.weighted_score == expected


def test_exactly_at_threshold_meets_it() -> None:
    result = run(with_scores(accuracy=3, judgment=3, safety=3, communication=3))
    assert result.weighted_score == 75.0
    assert result.proposed_outcome == "evidence_verified"
    assert "meets_pilot_threshold" in result.outcome_reasons


def test_one_step_below_threshold_needs_practice() -> None:
    result = run(with_scores(accuracy=3, judgment=3, safety=3, communication=2))
    assert result.weighted_score == 71.25
    assert result.proposed_outcome == "needs_practice"
    assert result.outcome_reasons == ("below_pilot_threshold",)


def test_threshold_is_reported_as_unvalidated() -> None:
    result = run(score_payload())
    assert result.pilot_threshold == 75.0
    assert result.pilot_threshold_is_validated is False


def test_result_does_not_depend_on_dimension_order() -> None:
    payload = score_payload()
    reordered = copy.deepcopy(payload)
    reordered["dimensions"] = dict(reversed(payload["dimensions"].items()))
    assert run(payload) == run(reordered)


def test_same_input_gives_byte_identical_output() -> None:
    payload = score_payload(critical_findings=[finding("dismissed")])
    assert run(payload).model_dump_json() == run(payload).model_dump_json()


def test_confirmed_critical_event_vetoes_a_perfect_score() -> None:
    result = run(
        with_scores(accuracy=4, judgment=4, safety=4, communication=4) | {"critical_findings": [finding("confirmed")]}
    )
    assert result.weighted_score == 100.0
    assert result.critical_veto is True
    assert result.proposed_outcome == "needs_practice"
    assert result.confirmed_critical_events == ("unauthorized_action",)
    assert "critical_finding_confirmed" in result.outcome_reasons


def test_pending_critical_event_holds_for_review() -> None:
    result = run(score_payload(critical_findings=[finding("pending")]))
    assert result.proposed_outcome == "review_pending"
    assert result.pending_critical_events == ("unauthorized_action",)
    assert result.critical_veto is False


def test_pending_outranks_confirmed() -> None:
    findings = [finding("confirmed"), finding("pending", event="secret_exposure")]
    assert run(score_payload(critical_findings=findings)).proposed_outcome == "review_pending"


def test_dismissed_critical_event_has_no_effect() -> None:
    result = run(score_payload(critical_findings=[finding("dismissed")]))
    assert result.proposed_outcome == "evidence_verified"
    assert result.critical_veto is False


def test_unknown_critical_event_type_is_rejected() -> None:
    with pytest.raises(InvalidInput) as exc:
        run(score_payload(critical_findings=[finding("confirmed", event="bad_vibes")]))
    assert exc.value.details["unknown"] == ["bad_vibes"]


def test_quote_absent_from_artifacts_holds_for_review() -> None:
    payload = score_payload()
    payload["dimensions"]["accuracy"]["evidence"] = ["Refund confirmed."]
    result = run(payload)
    assert result.proposed_outcome == "review_pending"
    assert [(i.source, i.ref, i.quote) for i in result.evidence_issues] == [
        ("dimension", "accuracy", "Refund confirmed.")
    ]


def test_critical_finding_quote_must_also_be_in_artifacts() -> None:
    result = run(score_payload(critical_findings=[finding("confirmed", quote="Your refund is approved")]))
    assert result.evidence_issues[0].source == "critical_finding"
    assert result.proposed_outcome == "review_pending"


def test_quote_match_tolerates_whitespace_but_not_wording() -> None:
    payload = score_payload()
    payload["dimensions"]["accuracy"]["evidence"] = ["RWF 42,000.   The second entry"]  # spans a newline
    assert run(payload).evidence_issues == ()
    payload["dimensions"]["accuracy"]["evidence"] = ["captured payment of rwf 42,000"]
    assert len(run(payload).evidence_issues) == 1


def test_quote_cannot_span_two_artifacts() -> None:
    payload = score_payload()
    payload["dimensions"]["judgment"]["evidence"] = ["not a settled charge. Escalated to billing"]
    assert len(run(payload).evidence_issues) == 1


def test_missing_dimension_is_rejected() -> None:
    payload = score_payload()
    del payload["dimensions"]["safety"]
    with pytest.raises(InvalidInput) as exc:
        run(payload)
    assert exc.value.details == {"missing": ["safety"], "unexpected": []}


def test_unknown_dimension_is_rejected() -> None:
    payload = score_payload()
    payload["dimensions"]["efficiency"] = {"score": 4, "evidence": ["Escalated to billing"]}
    with pytest.raises(InvalidInput) as exc:
        run(payload)
    assert exc.value.details["unexpected"] == ["efficiency"]


def test_artifact_type_outside_scenario_is_rejected() -> None:
    with pytest.raises(InvalidInput) as exc:
        run(score_payload(artifacts=[{"type": "triage_table", "text": "Escalated to billing"}]))
    assert exc.value.details["unexpected"] == ["triage_table"]


@pytest.mark.parametrize("bad", [5, -1, 2.5, "3", True, None])
def test_score_must_be_a_strict_integer_0_to_4(bad: Any) -> None:
    payload = score_payload()
    payload["dimensions"]["accuracy"]["score"] = bad
    with pytest.raises(ValidationError):
        ScoreRequest.model_validate(payload)


@pytest.mark.parametrize("evidence", [[], [""], ["   "]])
def test_every_dimension_needs_evidence(evidence: list[str]) -> None:
    payload = score_payload()
    payload["dimensions"]["accuracy"]["evidence"] = evidence
    with pytest.raises(ValidationError):
        ScoreRequest.model_validate(payload)


@pytest.mark.parametrize("quote", ["the", "No refund", "   ok   ", "x" * (MIN_QUOTE_CHARS - 1)])
def test_quotes_too_short_to_mean_anything_are_rejected(quote: str) -> None:
    payload = score_payload()
    payload["dimensions"]["accuracy"]["evidence"] = [quote]
    with pytest.raises(ValidationError, match="at least 12 characters"):
        ScoreRequest.model_validate(payload)


def test_short_critical_finding_quote_is_rejected() -> None:
    with pytest.raises(ValidationError):
        ScoreRequest.model_validate(score_payload(critical_findings=[finding("confirmed", quote="refund")]))


def test_quote_at_the_minimum_length_is_accepted() -> None:
    payload = score_payload()
    payload["dimensions"]["safety"]["evidence"] = ["No refund pr"]  # exactly 12, and present in the note
    assert len("No refund pr") == MIN_QUOTE_CHARS
    assert run(payload).evidence_issues == ()


@pytest.mark.parametrize("field", ["weighted_score", "proposed_outcome", "critical_veto", "score_source"])
def test_server_computed_fields_cannot_be_supplied(field: str) -> None:
    with pytest.raises(ValidationError, match="Extra inputs are not permitted"):
        ScoreRequest.model_validate(score_payload(**{field: "anything"}))


def test_artifacts_are_required() -> None:
    with pytest.raises(ValidationError):
        ScoreRequest.model_validate(score_payload(artifacts=[]))
