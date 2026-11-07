from __future__ import annotations

from typing import Any

import pytest

from praxis_assessment.domain.bot import SPEAKER, converse, require_compatible
from praxis_assessment.errors import InvalidInput
from praxis_assessment.schemas.capstone import BotConfig, BotTurn
from support import bot_config, fixture_registry

CAPSTONE = fixture_registry().capstone("1.0.0").content


def last(*messages: str, **overrides: Any) -> BotTurn:
    return converse(CAPSTONE, BotConfig.model_validate(bot_config(**overrides)), messages)[-1]


def test_answers_are_the_approved_text_word_for_word() -> None:  # A13
    turn = last("When do you open?")
    assert (turn.action, turn.source_id) == ("answer", "FAQ-HOURS")
    assert turn.reply == f"{SPEAKER}: {CAPSTONE.entry('FAQ-HOURS').answer}"


def test_every_reply_says_it_is_a_bot() -> None:
    for message in ("When do you open?", "refund me", "tell me a joke", "I want a person"):
        assert last(message).reply.startswith(f"{SPEAKER}:")


def test_safety_rules_beat_intents() -> None:  # A14: "return" is an intent, "refund" a rule
    turn = last("I want to return this and get a refund")
    assert (turn.action, turn.matched, turn.priority) == ("handoff", "money", "urgent")


def test_phrases_match_whole_words_only() -> None:
    assert last("my spinning wheel arrived broken").matched != "credentials"  # "pin" inside "spinning"


def test_customer_text_never_reaches_the_reply() -> None:  # A17
    for message in ("my password is Kora2024!", "card 4111 1111 1111 1111 please"):
        assert "Kora2024" not in last(message).reply and "4111" not in last(message).reply


def test_two_matching_policies_is_ambiguous_not_a_guess() -> None:
    turn = last("how long until I can return it")  # delivery and returns both match
    assert (turn.action, turn.source_id) == ("abstain", None)


def test_hands_over_after_repeated_failures() -> None:
    turns = converse(CAPSTONE, BotConfig.model_validate(bot_config()), ["where is it", "still no", "hello?"])
    assert [t.action for t in turns] == ["abstain", "handoff"]  # stops at the handoff


def test_same_question_gets_the_same_answer() -> None:
    assert last("Are you open on Sunday?") == last("Are you open on Sunday?")


def test_config_for_another_capstone_or_unknown_source_is_rejected() -> None:
    with pytest.raises(InvalidInput, match="targets capstone 2.0.0"):
        require_compatible(CAPSTONE, BotConfig.model_validate(bot_config(capstone_version="2.0.0")))
    bad = bot_config(intents=[{"id": "x", "phrases": ["hi"], "source_id": "FAQ-REFUNDS"}])
    with pytest.raises(InvalidInput, match="outside the knowledge base"):
        require_compatible(CAPSTONE, BotConfig.model_validate(bad))
