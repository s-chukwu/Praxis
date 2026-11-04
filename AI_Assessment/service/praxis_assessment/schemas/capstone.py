from __future__ import annotations

import re
from typing import Annotated, Literal

from pydantic import AfterValidator, Field, StringConstraints, model_validator

from praxis_assessment.content.capstone import BotAction, CaseCategory, Priority, SourceId
from praxis_assessment.schemas.common import Frozen, SemVer, Slug, require_unique

HandoffState = Literal["requested", "queued", "accepted_by_human", "queue_unavailable", "timeout"]
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


class BotTurn(Frozen):
    """One bot reply. Also the trace a reviewer reads when scoring B1 and B2."""

    action: BotAction
    matched: str | None  # the rule or intent id that decided this turn, or None for the fallback
    source_id: SourceId | None
    priority: Priority | None
    flagged_injection: bool
    reply: str


class HandoffTicket(Frozen):
    """What the human queue receives. The time is stamped by our Go backend when a real ticket is created."""

    reason: str  # the rule that fired, or "no_answer" when the bot ran out of options
    priority: Priority
    channel: Literal["chat"] = "chat"
    summary: str  # the customer's words with secrets masked; never shown back to the customer
    states: tuple[HandoffState, ...]
    customer_notice: str


FailReason = Literal["wrong_action", "wrong_source", "wrong_priority", "missed_injection", "leaked_secret"]


class CaseResult(Frozen):
    case_id: str
    category: CaseCategory
    passed: bool
    critical_failure: bool
    reasons: tuple[FailReason, ...]
    turns: tuple[BotTurn, ...]
    # Practice only. A held-out ticket summary would quote the hidden customer message.
    ticket: HandoffTicket | None


class CategoryResult(Frozen):
    category: CaseCategory
    passed: int
    total: int


class Evaluation(Frozen):
    suite: Literal["practice", "held_out"]
    passed: int
    total: int
    critical_failures: tuple[str, ...]
    categories: tuple[CategoryResult, ...]
    cases: tuple[CaseResult, ...]
    # Passing cases is evidence for B1-B3, not a score. A reviewer always signs off.
    requires_human_review: Literal[True] = True
