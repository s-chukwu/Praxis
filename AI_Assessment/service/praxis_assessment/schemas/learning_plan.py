from __future__ import annotations

from typing import Literal
from uuid import UUID

from pydantic import Field, StrictBool

from praxis_assessment.content.models import CourseId, Track
from praxis_assessment.schemas.common import CompetencyCode, ContentRef, Frozen, Rating, SemVer

ReasonCode = Literal["role_orientation", "competency_gap", "safety_gate", "prerequisite", "integrated_challenge"]
SafetyTrigger = Literal["safety_competency_below_threshold", "critical_safety_finding"]


# No field for CV claims or self-ratings, on purpose: a claim must never shorten someone's training.
class LearningPlanRequest(Frozen):
    request_id: UUID
    candidate_id: UUID
    review_id: UUID
    target_track: Track
    role_standard_version: SemVer
    policy_version: SemVer
    reviewed_competencies: dict[CompetencyCode, Rating] = Field(max_length=20)
    critical_safety_finding: StrictBool
    # Modules our Go backend has recorded as completed AND verified. They satisfy prerequisites;
    # they never cancel a gap that the latest review still shows.
    verified_completed_modules: tuple[CourseId, ...] = Field(default=(), max_length=20)


class CompetencyScore(Frozen):
    code: str
    reviewed_score: int


class Reason(Frozen):
    code: ReasonCode
    competency: str | None = None
    reviewed_score: int | None = None
    required_by: str | None = None


class AssignedModule(Frozen):
    order: int
    course_id: str
    title: str
    covers: tuple[str, ...]
    reasons: tuple[Reason, ...]
    previously_completed: bool


class WaivedModule(Frozen):
    course_id: str
    required_by: str
    reason: Literal["competencies_demonstrated", "verified_completed"]


class SafetyGate(Frozen):
    required: bool
    module: str
    triggers: tuple[SafetyTrigger, ...]


class LearningPlan(Frozen):
    status: Literal["plan_ready", "diagnostic_incomplete"]
    threshold: int
    missing_competencies: tuple[str, ...]
    demonstrated: tuple[CompetencyScore, ...]
    gaps: tuple[CompetencyScore, ...]
    safety_gate: SafetyGate
    modules: tuple[AssignedModule, ...]
    waived_prerequisites: tuple[WaivedModule, ...]
    reassessment_required: Literal[True]


class LearningPlanResponse(Frozen):
    request_id: UUID
    candidate_id: UUID
    review_id: UUID
    target_track: Track
    role_standard: ContentRef
    policy: ContentRef
    curriculum: ContentRef
    result: LearningPlan
