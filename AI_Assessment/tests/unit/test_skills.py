from __future__ import annotations

from praxis_assessment.domain.skills import instructions_for, mask_contacts, vet_suggestions
from support import fixture_registry

STANDARD = fixture_registry().role_standard("1.0.0").content
CV = "Handled 60 billing tickets a day at a telecom call centre.\nEscalated suspected account takeovers to security."


def item(competency: str, quote: str, confidence: str = "medium") -> dict[str, str]:
    return {"competency": competency, "quote": quote, "confidence": confidence}


def test_keeps_suggestions_quoted_from_the_cv() -> None:
    kept, dropped = vet_suggestions(STANDARD, CV, [item("C4", "Escalated suspected account takeovers")])
    assert [(s.competency, s.status) for s in kept] == [("C4", "suggested")] and dropped == 0


def test_drops_invented_or_paraphrased_evidence() -> None:
    kept, dropped = vet_suggestions(STANDARD, CV, [item("C2", "Won employee of the month twice"),
                                                   item("C4", "Escalated suspected account hijacks")])  # fmt: skip
    assert (kept, dropped) == ((), 2)


def test_drops_unknown_codes_and_malformed_items() -> None:
    bad = [item("C9", "Handled 60 billing tickets"), item("C1", "too short"), {"competency": "C1"}, "C1", None]
    assert vet_suggestions(STANDARD, CV, bad) == ((), 5)


def test_one_suggestion_per_competency_keeping_the_most_confident() -> None:
    items = [item("C3", "Handled 60 billing tickets", "low"), item("C3", "telecom call centre", "high")]
    kept, dropped = vet_suggestions(STANDARD, CV, items)
    assert [(s.evidence_quote, s.confidence) for s in kept] == [("telecom call centre", "high")] and dropped == 1


def test_quotes_survive_whitespace_differences_only() -> None:
    kept, _ = vet_suggestions(STANDARD, CV, [item("C3", "call  centre.\n Escalated suspected")])
    assert len(kept) == 1


def test_contact_details_are_masked_before_leaving_the_service() -> None:
    text = "Mail amari.k+cv@gmail.com or +250 788 123 456. National ID 1199880012345678."
    assert mask_contacts(text) == "Mail [email] or [number]. National ID [number]."


def test_masking_keeps_the_evidence_a_skill_needs() -> None:
    text = "Handled 60 tickets a day in 2023, order KD-5531, rated 4.8/5"
    assert mask_contacts(text) == text


def test_instructions_name_every_competency_and_never_contain_the_cv() -> None:
    instructions = instructions_for(STANDARD)
    assert all(f"{c.code}: {c.name}" in instructions for c in STANDARD.competencies)
    assert "untrusted data" in instructions and "billing tickets" not in instructions  # A09: the CV travels apart
