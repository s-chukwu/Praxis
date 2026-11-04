from __future__ import annotations

import re
from typing import Annotated, Literal

from pydantic import AfterValidator, Field, StringConstraints, model_validator

from praxis_assessment.content.capstone import Priority, SourceId
from praxis_assessment.schemas.common import Frozen, SemVer, Slug, require_unique

_WORDS = re.compile(r"\w+")


def words(text: str) -> tuple[str, ...]:
    """Lower-cased word tokens. Candidate phrases and customer messages are compared only through this."""
    return tuple(_WORDS.findall(text.lower()))


def _has_words(phrase: str) -> str:
    if not words(phrase):
        raise ValueError("a phrase needs at least one letter or digit")
    return phrase


# Plain words, never a pattern: a candidate-supplied regex could hang the evaluator.
Phrase = Annotated[
    str, StringConstraints(strip_whitespace=True, min_length=2, max_length=60), AfterValidator(_has_words)
]
Phrases = Annotated[tuple[Phrase, ...], Field(min_length=1, max_length=20)]


class Intent(Frozen):
    """Customer messages containing any phrase get the approved answer for source_id, word for word."""

    id: Slug
    phrases: Phrases
    source_id: SourceId


class Rule(Frozen):
    """Checked before intents. This is where the candidate decides what the bot must never answer itself."""

    id: Slug
    phrases: Phrases
    action: Literal["handoff", "abstain"]
    priority: Priority = "routine"
    flag_injection: bool = False


class BotConfig(Frozen):
    capstone_version: SemVer
    rules: tuple[Rule, ...] = Field(default=(), max_length=20)
    intents: tuple[Intent, ...] = Field(min_length=1, max_length=20)
    fallback: Literal["abstain", "handoff"]
    # After this many unanswered turns in a row the bot hands over instead of failing the customer again.
    handoff_after_failures: int = Field(ge=1, le=5)

    @model_validator(mode="after")
    def _integrity(self) -> BotConfig:
        require_unique([r.id for r in self.rules] + [i.id for i in self.intents], "rule or intent id")
        return self
