from __future__ import annotations

from collections.abc import Sequence
from typing import Any

from pydantic import ValidationError

from praxis_assessment.content.models import RoleStandard
from praxis_assessment.domain.scoring import normalize
from praxis_assessment.schemas.skills import ModelSuggestion, SkillSuggestion

_RANK = {"high": 0, "medium": 1, "low": 2}


def vet_suggestions(
    standard: RoleStandard, sent_text: str, items: Sequence[Any]
) -> tuple[tuple[SkillSuggestion, ...], int]:
    """Keep only model suggestions I can check: a known competency and a quote that is really in the CV.

    sent_text is the masked text the model saw, so a quote containing a mask still verifies.
    """
    corpus = normalize(sent_text)
    best: dict[str, ModelSuggestion] = {}
    dropped = 0
    for item in items:
        try:
            suggestion = ModelSuggestion.model_validate(item)
        except ValidationError:
            dropped += 1
            continue
        # A quote the model paraphrased or invented is exactly the hallucination this check exists for.
        if suggestion.competency not in standard.codes or normalize(suggestion.quote) not in corpus:
            dropped += 1
            continue
        kept = best.get(suggestion.competency)
        if kept is not None:
            dropped += 1
            if _RANK[suggestion.confidence] >= _RANK[kept.confidence]:
                continue
        best[suggestion.competency] = suggestion

    vetted = tuple(
        SkillSuggestion(competency=s.competency, evidence_quote=s.quote, confidence=s.confidence)
        for s in sorted(best.values(), key=lambda s: s.competency)
    )
    return vetted, dropped
