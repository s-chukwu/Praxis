from __future__ import annotations

import re
from collections.abc import Sequence

from praxis_assessment.content.capstone import QueueOutcome
from praxis_assessment.domain.bot import SPEAKER
from praxis_assessment.schemas.capstone import BotTurn, HandoffState, HandoffTicket

JOINED = f"{SPEAKER}: A person from our team has joined the chat."
# Never "transferring you" or "connected": until a person accepts, a person has only been requested.
UNAVAILABLE = f"{SPEAKER}: Our team can't take chats right now. Your case is saved and someone will contact you."
TIMED_OUT = f"{SPEAKER}: Nobody from our team was free in time. Your case is saved and someone will contact you."

_FLOWS: dict[QueueOutcome, tuple[tuple[HandoffState, ...], str]] = {
    "accepted": (("requested", "queued", "accepted_by_human"), JOINED),
    "unavailable": (("requested", "queue_unavailable"), UNAVAILABLE),
    "timeout": (("requested", "queued", "timeout"), TIMED_OUT),
}

MASK = "[redacted]"
_SECRETS = (
    # 13-19 digits, optionally grouped by spaces or dashes: the shape of a payment card number.
    re.compile(r"\b\d(?:[ -]?\d){12,18}\b"),
    # The value after a word like "password" or "pin". Bounded \S+ keeps matching linear on long input.
    re.compile(r"(?i)\b(password|passcode|pin|cvv|otp)\b(\s*(?:is|:|=)?\s*)(\S{1,64})"),
)
SUMMARY_LIMIT = 300


def hand_over(turn: BotTurn, messages: Sequence[str], queue: QueueOutcome) -> HandoffTicket:
    states, notice = _FLOWS[queue]
    return HandoffTicket(
        reason=turn.matched or "no_answer",
        priority=turn.priority or "routine",
        summary=redact(" / ".join(messages))[:SUMMARY_LIMIT],
        states=states,
        customer_notice=notice,
    )


def redact(text: str) -> str:
    text = _SECRETS[0].sub(MASK, text)
    return _SECRETS[1].sub(lambda m: f"{m.group(1)}{m.group(2)}{MASK}", text)


def find_secrets(text: str) -> list[str]:
    """The secret values themselves, so the evaluator can prove none of them came back out."""
    return [m.group(0) for m in _SECRETS[0].finditer(text)] + [m.group(3) for m in _SECRETS[1].finditer(text)]
