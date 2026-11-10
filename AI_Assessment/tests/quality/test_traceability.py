from __future__ import annotations

import re
from pathlib import Path

TESTS = Path(__file__).parents[1]
PLAN = TESTS.parent / "docs" / "assessment_test_plan.md"
# Owned by our Go backend: consent revocation never reaches this service.
ELSEWHERE = {"A12"}


def test_every_test_plan_requirement_has_a_tagged_test() -> None:
    required = set(re.findall(r"^\| (A\d\d) \|", PLAN.read_text(), re.M)) - ELSEWHERE
    comments = (c for path in TESTS.rglob("test_*.py") for c in re.findall(r"#(.*)", path.read_text()))
    tagged = {tag for comment in comments for tag in re.findall(r"\bA\d\d\b", comment)}
    assert len(required) == 19
    assert sorted(required - tagged) == []
