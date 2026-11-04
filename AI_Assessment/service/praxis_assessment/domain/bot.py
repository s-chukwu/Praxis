from __future__ import annotations

from collections.abc import Sequence

from praxis_assessment.content.capstone import Capstone
from praxis_assessment.errors import InvalidInput
from praxis_assessment.schemas.capstone import BotConfig, BotTurn, words

# Every reply says it comes from a bot. Customers who took an AI agent for a person is how invented
# policies get believed.
SPEAKER = "KoraDesk virtual assistant"
ABSTAIN_REPLY = f"{SPEAKER}: I can't help with that here. You can ask for a person at any time."
HANDOFF_REPLY = f"{SPEAKER}: I've asked a person from our team to help. I'll tell you as soon as they join."


def require_compatible(capstone: Capstone, config: BotConfig) -> None:
    if config.capstone_version != capstone.version:
        raise InvalidInput(
            f"config targets capstone {config.capstone_version}, not {capstone.version}",
            details={"config": config.capstone_version, "capstone": capstone.version},
        )
    if unknown := sorted({i.source_id for i in config.intents} - capstone.sources):
        raise InvalidInput("intents point at sources outside the knowledge base", details={"unknown": unknown})


def converse(capstone: Capstone, config: BotConfig, messages: Sequence[str]) -> tuple[BotTurn, ...]:
    """Play a conversation through the bot. Stops at the first handoff, as a real session would."""
    turns: list[BotTurn] = []
    failures = 0
    for message in messages:
        turn = _respond(capstone, config, words(message), failures)
        turns.append(turn)
        if turn.action == "handoff":
            break
        failures = failures + 1 if turn.matched is None else 0
    return tuple(turns)


def _respond(capstone: Capstone, config: BotConfig, said: tuple[str, ...], failures: int) -> BotTurn:
    # Rules run first and in the candidate's order, so a safety rule always beats an intent.
    for rule in config.rules:
        if _mentions(said, rule.phrases):
            handoff = rule.action == "handoff"
            return BotTurn(
                action=rule.action,
                matched=rule.id,
                source_id=None,
                priority=rule.priority if handoff else None,
                flagged_injection=rule.flag_injection,
                reply=HANDOFF_REPLY if handoff else ABSTAIN_REPLY,
            )

    hits = {i.source_id: i.id for i in config.intents if _mentions(said, i.phrases)}
    # Two different policies matching means I don't know what was asked; guessing is how bots invent answers.
    if len(hits) == 1:
        source_id, intent_id = next(iter(hits.items()))
        reply = f"{SPEAKER}: {capstone.entry(source_id).answer}"
        return BotTurn(action="answer", matched=intent_id, source_id=source_id, priority=None,
                       flagged_injection=False, reply=reply)  # fmt: skip

    gives_up = config.fallback == "handoff" or failures + 1 >= config.handoff_after_failures
    return BotTurn(
        action="handoff" if gives_up else "abstain",
        matched=None,
        source_id=None,
        priority="routine" if gives_up else None,
        flagged_injection=False,
        reply=HANDOFF_REPLY if gives_up else ABSTAIN_REPLY,
    )


def _mentions(said: tuple[str, ...], phrases: Sequence[str]) -> bool:
    """True when any phrase appears as consecutive whole words, so "pin" never matches "spinning"."""
    for phrase in phrases:
        target = words(phrase)
        size = len(target)
        if any(said[i : i + size] == target for i in range(len(said) - size + 1)):
            return True
    return False
