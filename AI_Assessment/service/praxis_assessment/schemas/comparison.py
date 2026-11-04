from __future__ import annotations

from typing import Literal
from uuid import UUID

from pydantic import Field

from praxis_assessment.content.models import AttemptState
from praxis_assessment.schemas.common import ContentRef, Frozen, Rating, ScenarioId, SemVer, Slug


class ReviewedAttempt(Frozen):
    attempt_id: UUID
    scenario_id: ScenarioId
    scenario_version: SemVer
    rubric_version: SemVer
    state: AttemptState
    dimensions: dict[Slug, Rating] = Field(min_length=1, max_length=10)


class CompareRequest(Frozen):
    request_id: UUID
    candidate_id: UUID
    baseline: ReviewedAttempt
    final: ReviewedAttempt


class DimensionChange(Frozen):
    id: str
    baseline: int
    final: int
    delta: int
    direction: Literal["improved", "unchanged", "regressed"]


class Comparison(Frozen):
    dimensions: tuple[DimensionChange, ...]
    baseline_weighted_score: float
    final_weighted_score: float
    weighted_delta: float
    regressed_dimensions: tuple[str, ...]
    requires_human_review: bool


class CompareResponse(Frozen):
    request_id: UUID
    candidate_id: UUID
    baseline_attempt_id: UUID
    final_attempt_id: UUID
    rubric: ContentRef
    scenarios: ContentRef
    result: Comparison
