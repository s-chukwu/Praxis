from __future__ import annotations

from typing import Annotated, Literal
from uuid import UUID

from pydantic import Field, StringConstraints

from praxis_assessment.schemas.common import CompetencyCode, ContentRef, Frozen, SemVer
from praxis_assessment.schemas.scoring import Quote

Confidence = Literal["low", "medium", "high"]


class SkillsRequest(Frozen):
    request_id: UUID
    candidate_id: UUID
    # Our Go backend checks the cv_analysis consent before calling; I record which consent covered the call.
    consent_ref: UUID
    role_standard_version: SemVer
    text: Annotated[str, StringConstraints(min_length=1, max_length=20_000)]


class ModelSuggestion(Frozen):
    """One item as the model must return it. Anything looser is dropped, not repaired."""

    competency: CompetencyCode
    quote: Quote
    confidence: Confidence


class SkillSuggestion(Frozen):
    competency: CompetencyCode
    evidence_quote: str
    confidence: Confidence
    # A CV claim stays a suggestion until the candidate confirms it, and even then it is never a score.
    status: Literal["suggested"] = "suggested"


class SkillsResult(Frozen):
    suggestions: tuple[SkillSuggestion, ...] = Field(max_length=20)
    dropped: int  # model items rejected: unknown code, quote not in the CV, malformed or duplicate
    model: str
    requires_candidate_confirmation: Literal[True] = True


class SkillsResponse(Frozen):
    request_id: UUID
    candidate_id: UUID
    consent_ref: UUID
    role_standard: ContentRef
    result: SkillsResult
