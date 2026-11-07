from __future__ import annotations

import time

import pytest

from praxis_assessment.content.capstone import QueueOutcome
from praxis_assessment.domain.bot import converse
from praxis_assessment.domain.handoff import JOINED, MASK, hand_over, redact
from praxis_assessment.schemas.capstone import BotConfig
from support import bot_config, fixture_registry

CAPSTONE = fixture_registry().capstone("1.0.0").content
CONFIG = BotConfig.model_validate(bot_config())


def ticket(message: str, queue: QueueOutcome = "accepted"):
    turn = converse(CAPSTONE, CONFIG, [message])[-1]
    assert turn.action == "handoff"
    return hand_over(turn, [message], queue)


def test_accepted_handoff_walks_the_whole_protocol() -> None:
    t = ticket("Someone changed the email on my account")
    assert t.states == ("requested", "queued", "accepted_by_human")
    assert (t.reason, t.priority, t.customer_notice) == ("takeover", "urgent", JOINED)


@pytest.mark.parametrize(("queue", "final"), [("unavailable", "queue_unavailable"), ("timeout", "timeout")])
def test_no_human_means_no_claim_of_one(queue: QueueOutcome, final: str) -> None:  # A15, A16
    t = ticket("I want to speak to a person", queue)
    assert t.states[-1] == final and "accepted_by_human" not in t.states
    assert "joined" not in t.customer_notice and "case is saved" in t.customer_notice


@pytest.mark.parametrize(
    ("text", "secret"),
    [
        ("my password is Kora2024! can you check", "Kora2024!"),
        ("card 4111 1111 1111 1111 was charged", "4111 1111 1111 1111"),
        ("card 4111-1111-1111-1111", "1111-1111"),
        ("PIN: 0420 please", "0420"),
    ],
)
def test_secrets_never_reach_the_ticket(text: str, secret: str) -> None:  # A17
    assert secret not in redact(text) and MASK in redact(text)


def test_ordinary_details_survive_redaction() -> None:
    text = "order KD-5531, phone +250 788 123 456, my passwords page is broken"
    assert redact(text) == text


def test_redaction_stays_fast_on_hostile_input() -> None:
    started = time.perf_counter()
    redact("password " + "x" * 20_000 + " 1" * 20_000)
    assert time.perf_counter() - started < 0.5
