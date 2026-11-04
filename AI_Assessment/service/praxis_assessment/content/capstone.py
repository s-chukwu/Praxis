from __future__ import annotations

from typing import Annotated, Literal

from pydantic import Field, StringConstraints, model_validator

from praxis_assessment.schemas.common import Frozen, SemVer, require_unique

SourceId = Annotated[str, StringConstraints(pattern=r"^FAQ-[A-Z]+$")]
BotAction = Literal["answer", "abstain", "handoff"]
Priority = Literal["routine", "urgent"]
QueueOutcome = Literal["accepted", "unavailable", "timeout"]
CaseCategory = Literal[
    "policy_answer",
    "no_match",
    "ambiguous_intent",
    "adverse_sentiment",
    "prompt_injection",
    "pii_credentials",
    "unauthorized_action",
    "human_request",
    "repeated_failure",
    "queue_unavailable",
]


class KbEntry(Frozen):
    source_id: SourceId
    title: str
    answer: str = Field(min_length=1, max_length=600)


class Expectation(Frozen):
    # A list, not one action: an ambiguous request is handled safely by abstaining or by handing over.
    accept: tuple[BotAction, ...] = Field(min_length=1)
    source_id: SourceId | None = None
    priority: Priority | None = None
    flag_injection: bool = False

    @model_validator(mode="after")
    def _integrity(self) -> Expectation:
        require_unique(self.accept, "accepted action")
        if ("answer" in self.accept) != (self.source_id is not None):
            raise ValueError("source_id is required exactly when answer is accepted")
        if self.priority is not None and "handoff" not in self.accept:
            raise ValueError("priority only applies when handoff is accepted")
        return self


class BotCase(Frozen):
    id: Annotated[str, StringConstraints(pattern=r"^[PH]-\d{2}$")]
    category: CaseCategory
    # Several messages model a customer the bot keeps failing; the expectation applies after the last one.
    messages: tuple[Annotated[str, StringConstraints(min_length=1, max_length=500)], ...] = Field(
        min_length=1, max_length=5
    )
    queue: QueueOutcome = "accepted"
    expect: Expectation
    critical: bool


def _check_cases(cases: tuple[BotCase, ...], prefix: str, sources: set[str] | None) -> None:
    """sources is None for held-out cases; the registry checks those against their capstone."""
    require_unique([c.id for c in cases], "case")
    for case in cases:
        if not case.id.startswith(prefix):
            raise ValueError(f"case {case.id} must start with {prefix}")
        if sources is not None and case.expect.source_id is not None and case.expect.source_id not in sources:
            raise ValueError(f"case {case.id} expects unknown source {case.expect.source_id}")


class Capstone(Frozen):
    """The knowledge base candidates configure against, plus the practice cases they may see and rerun."""

    version: SemVer
    company: str
    kb: tuple[KbEntry, ...] = Field(min_length=1)
    practice: tuple[BotCase, ...] = Field(min_length=1)

    @model_validator(mode="after")
    def _integrity(self) -> Capstone:
        require_unique([e.source_id for e in self.kb], "source")
        _check_cases(self.practice, "P-", self.sources)
        return self

    @property
    def sources(self) -> set[str]:
        return {e.source_id for e in self.kb}

    def entry(self, source_id: str) -> KbEntry:
        return next(e for e in self.kb if e.source_id == source_id)


class HeldOutCases(Frozen):
    """Server-only. No endpoint returns these; results report a case's category and outcome, never its text."""

    version: SemVer
    cases: tuple[BotCase, ...] = Field(min_length=1)

    @model_validator(mode="after")
    def _integrity(self) -> HeldOutCases:
        _check_cases(self.cases, "H-", None)
        return self
