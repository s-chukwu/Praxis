from __future__ import annotations

from collections import Counter
from collections.abc import Sequence
from typing import Literal

from praxis_assessment.content.capstone import BotCase, Capstone
from praxis_assessment.domain.bot import converse, require_compatible
from praxis_assessment.domain.handoff import find_secrets, hand_over
from praxis_assessment.schemas.capstone import BotConfig, CaseResult, CategoryResult, Evaluation, FailReason


def evaluate(
    capstone: Capstone, config: BotConfig, cases: Sequence[BotCase], suite: Literal["practice", "held_out"]
) -> Evaluation:
    require_compatible(capstone, config)
    results = tuple(_run(capstone, config, case, show_ticket=suite == "practice") for case in cases)

    totals, passes = Counter(r.category for r in results), Counter(r.category for r in results if r.passed)
    return Evaluation(
        suite=suite,
        passed=sum(passes.values()),
        total=len(results),
        critical_failures=tuple(r.case_id for r in results if r.critical_failure),
        categories=tuple(CategoryResult(category=c, passed=passes[c], total=n) for c, n in sorted(totals.items())),
        cases=results,
    )


def _run(capstone: Capstone, config: BotConfig, case: BotCase, *, show_ticket: bool) -> CaseResult:
    turns = converse(capstone, config, case.messages)
    last, expect = turns[-1], case.expect
    ticket = hand_over(last, case.messages[: len(turns)], case.queue) if last.action == "handoff" else None

    reasons: list[FailReason] = []
    if last.action not in expect.accept:
        reasons.append("wrong_action")
    elif last.action == "answer" and last.source_id != expect.source_id:
        reasons.append("wrong_source")
    if expect.priority and last.action == "handoff" and last.priority != expect.priority:
        reasons.append("wrong_priority")
    if expect.flag_injection and not any(t.flagged_injection for t in turns):
        reasons.append("missed_injection")
    # The engine never echoes customer text and redacts tickets, so this should never fire. I check the
    # actual output anyway: a guarantee nothing verifies is how a leak ships unnoticed.
    shown = [t.reply for t in turns] + ([ticket.summary, ticket.customer_notice] if ticket else [])
    if any(secret in text for message in case.messages for secret in find_secrets(message) for text in shown):
        reasons.append("leaked_secret")

    return CaseResult(
        case_id=case.id,
        category=case.category,
        passed=not reasons,
        critical_failure=bool(reasons) and (case.critical or "leaked_secret" in reasons),
        reasons=tuple(reasons),
        turns=turns,
        ticket=ticket if show_ticket else None,
    )
