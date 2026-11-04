from __future__ import annotations

import pytest

import praxis_assessment.domain.handoff as handoff
from praxis_assessment.domain.capstone import evaluate
from praxis_assessment.schemas.capstone import BotConfig
from support import bot_config, fixture_registry

REGISTRY = fixture_registry()
CAPSTONE = REGISTRY.capstone("1.0.0").content
HELD_OUT = REGISTRY.held_out["1.0.0"].content.cases
CAREFUL = BotConfig.model_validate(bot_config())
# Answers the FAQs and nothing else: what a candidate who skipped the handoff work would submit.
NAIVE = BotConfig.model_validate(bot_config(rules=[], fallback="abstain", handoff_after_failures=5))


def test_a_careful_config_passes_both_suites() -> None:
    for cases, suite in ((CAPSTONE.practice, "practice"), (HELD_OUT, "held_out")):
        result = evaluate(CAPSTONE, CAREFUL, cases, suite)
        assert (result.passed, result.critical_failures) == (result.total, ())


def test_skipping_the_safety_rules_fails_critically() -> None:
    result = evaluate(CAPSTONE, NAIVE, HELD_OUT, "held_out")
    failed = {r.case_id: r.reasons for r in result.cases if not r.passed}
    assert failed["H-04"] == ("wrong_action",)  # the Air Canada pattern: no handoff on a refund promise
    assert failed["H-06"] == ("missed_injection",)  # the $1 "binding offer": it abstained but never flagged it
    assert {"H-04", "H-06", "H-11", "H-13"} <= set(result.critical_failures)
    assert "H-08" not in result.critical_failures  # not every failure is critical


def test_coverage_is_reported_per_category() -> None:
    result = evaluate(CAPSTONE, NAIVE, HELD_OUT, "held_out")
    by_category = {c.category: (c.passed, c.total) for c in result.categories}
    assert by_category["policy_answer"] == (3, 3)
    assert by_category["prompt_injection"] == (0, 2)
    assert sum(t for _, t in by_category.values()) == len(HELD_OUT) == 16


def test_held_out_results_never_quote_the_hidden_cases() -> None:  # A20
    body = evaluate(CAPSTONE, CAREFUL, HELD_OUT, "held_out").model_dump_json()
    assert not any(message in body for case in HELD_OUT for message in case.messages)
    assert all(r.ticket is None for r in evaluate(CAPSTONE, CAREFUL, HELD_OUT, "held_out").cases)


def test_a_redaction_bug_is_caught_as_a_critical_leak(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr(handoff, "redact", lambda text: text)
    result = evaluate(CAPSTONE, CAREFUL, HELD_OUT, "held_out")
    leak = next(r for r in result.cases if r.case_id == "H-10")  # "My password is Kora2024! ..."
    assert (leak.reasons, leak.critical_failure) == (("leaked_secret",), True)
