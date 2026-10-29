from __future__ import annotations

from typing import Literal

from praxis_assessment.content.models import Rubric, Scenario
from praxis_assessment.domain.scoring import require_rubric_dimensions, weighted_score
from praxis_assessment.errors import InvalidInput, NotComparable
from praxis_assessment.schemas.comparison import CompareRequest, Comparison, DimensionChange


def compare_attempts(rubric: Rubric, baseline: Scenario, final: Scenario, request: CompareRequest) -> Comparison:
    _require_parallel_forms(baseline, final)
    versions = {(a.scenario_version, a.rubric_version) for a in (request.baseline, request.final)}
    # Deltas only mean something on one scale. After a rubric change, a reviewer re-scores the baseline.
    if versions != {(request.baseline.scenario_version, rubric.version)}:
        raise NotComparable(
            "baseline and final must be scored on the same scenario bank and rubric",
            details={
                "baseline": [request.baseline.scenario_version, request.baseline.rubric_version],
                "final": [request.final.scenario_version, request.final.rubric_version],
            },
        )
    for attempt in (request.baseline, request.final):
        # A delta against a score that can still change would be reported and then silently go stale.
        if attempt.state != "reviewed":
            raise NotComparable(
                f"attempt {attempt.attempt_id} is {attempt.state}, not reviewed",
                details={"attempt_id": str(attempt.attempt_id), "state": attempt.state},
            )
        require_rubric_dimensions(rubric, attempt.dimensions)

    before, after = request.baseline.dimensions, request.final.dimensions
    changes = tuple(
        DimensionChange(
            id=d.id,
            baseline=before[d.id],
            final=after[d.id],
            delta=after[d.id] - before[d.id],
            direction=_direction(after[d.id] - before[d.id]),
        )
        for d in rubric.dimensions
    )
    baseline_score, final_score = weighted_score(rubric, before), weighted_score(rubric, after)
    regressed = tuple(c.id for c in changes if c.direction == "regressed")
    return Comparison(
        dimensions=changes,
        baseline_weighted_score=float(baseline_score),
        final_weighted_score=float(final_score),
        weighted_delta=float(final_score - baseline_score),
        regressed_dimensions=regressed,
        # Going backwards after training is either a real problem or a scoring inconsistency; a person decides.
        requires_human_review=bool(regressed),
    )


def _direction(delta: int) -> Literal["improved", "unchanged", "regressed"]:
    return "improved" if delta > 0 else "regressed" if delta < 0 else "unchanged"


def _require_parallel_forms(baseline: Scenario, final: Scenario) -> None:
    if baseline.stage != "baseline" or final.stage != "final":
        raise InvalidInput(
            "compare needs a baseline form followed by a final form",
            details={"baseline_stage": baseline.stage, "final_stage": final.stage},
        )
    if baseline.parallel_form_id != final.id:
        raise InvalidInput(
            f"{final.id} is not the parallel form of {baseline.id}",
            details={"expected_final": baseline.parallel_form_id, "received": final.id},
        )
