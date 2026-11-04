from __future__ import annotations

from typing import Any

import pytest
from pydantic import ValidationError

from praxis_assessment.schemas.capstone import BotConfig, words
from support import bot_config


def test_a_careful_candidates_config_is_accepted() -> None:
    config = BotConfig.model_validate(bot_config())
    assert [r.id for r in config.rules][:2] == ["takeover", "money"]


def test_phrases_are_words_not_patterns() -> None:
    # A catastrophic-backtracking regex is just two words here; nothing is ever compiled from candidate input.
    assert words("(a+)+$ OR 1=1") == ("a", "or", "1", "1")


@pytest.mark.parametrize(
    ("change", "message"),
    [
        ({"intents": [{"id": "x", "phrases": ["!!"], "source_id": "FAQ-HOURS"}]}, "at least one letter or digit"),
        ({"intents": [{"id": "x", "phrases": ["a" * 61], "source_id": "FAQ-HOURS"}]}, "at most 60"),
        ({"intents": [{"id": "i" + "a" * n, "phrases": ["hi"], "source_id": "FAQ-HOURS"} for n in range(21)]},
         "at most 20"),
        ({"intents": [{"id": "human", "phrases": ["hi"], "source_id": "FAQ-HOURS"}]}, "duplicate rule or intent id"),
        ({"intents": [{"id": "x", "phrases": ["hi"], "source_id": "hours"}]}, "pattern"),
        ({"handoff_after_failures": 0}, "greater than or equal to 1"),
        ({"reply_text": "Sure, refund approved"}, "Extra inputs are not permitted"),  # A13: no free-text answers
    ],
)
def test_out_of_bounds_config_is_rejected(change: dict[str, Any], message: str) -> None:
    with pytest.raises(ValidationError, match=message):
        BotConfig.model_validate(bot_config(**change))
