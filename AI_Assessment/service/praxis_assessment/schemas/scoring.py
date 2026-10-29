from __future__ import annotations

from typing import Annotated, Literal
from uuid import UUID

from pydantic import Field, StringConstraints

from praxis_assessment.schemas.common import ContentRef, Frozen, Rating, ScenarioId, SemVer, Slug

# A quote has to be long enough to point at something. Without a floor, "the" or "ok" passes the
# verbatim check against any answer and proves nothing. 12 characters is roughly three short words;
# the reviewer UI should still have reviewers select text rather than type it.
MIN_QUOTE_CHARS = 12
Quote = Annotated[str, StringConstraints(strip_whitespace=True, min_length=MIN_QUOTE_CHARS, max_length=500)]

ProposedOutcome = Literal["evidence_verified", "needs_practice", "review_pending"]
OutcomeReason = Literal[
    "critical_finding_pending",
    "evidence_not_found",
    "critical_finding_confirmed",
    "below_pilot_threshold",
    "meets_pilot_threshold",
]


class DimensionRating(Frozen):
    score: Rating
    evidence: tuple[Quote, ...] = Field(min_length=1, max_length=10)


class CriticalFinding(Frozen):
    event_type: Slug
    supporting_quote: Quote
    policy_reference: Annotated[str, StringConstraints(strip_whitespace=True, min_length=1, max_length=200)]
    decision: Literal["confirmed", "dismissed", "pending"]


class Artifact(Frozen):
    type: Slug
    text: Annotated[str, StringConstraints(min_length=1, max_length=20_000)]


class ScoreRequest(Frozen):
    request_id: UUID
    attempt_id: UUID
    candidate_id: UUID
    reviewer_id: UUID
    scenario_id: ScenarioId
    scenario_version: SemVer
    rubric_version: SemVer
    dimensions: dict[Slug, DimensionRating] = Field(min_length=1, max_length=10)
    critical_findings: tuple[CriticalFinding, ...] = Field(default=(), max_length=10)
    artifacts: tuple[Artifact, ...] = Field(min_length=1, max_length=10)


class DimensionScore(Frozen):
    id: str
    score: int
    weight_bp: int
    weighted_points: float


class EvidenceIssue(Frozen):
    source: Literal["dimension", "critical_finding"]
    ref: str
    quote: str
    reason: Literal["not_found_in_artifacts"]


class ScoreResult(Frozen):
    dimensions: tuple[DimensionScore, ...]
    weighted_score: float
    pilot_threshold: float
    pilot_threshold_is_validated: Literal[False]
    critical_veto: bool
    confirmed_critical_events: tuple[str, ...]
    pending_critical_events: tuple[str, ...]
    evidence_issues: tuple[EvidenceIssue, ...]
    proposed_outcome: ProposedOutcome
    outcome_reasons: tuple[OutcomeReason, ...]
    score_source: Literal["human_review"]
    requires_signoff: Literal[True]


class ScoreResponse(Frozen):
    request_id: UUID
    attempt_id: UUID
    scenario_id: str
    stage: str
    rubric: ContentRef
    scenarios: ContentRef
    result: ScoreResult
