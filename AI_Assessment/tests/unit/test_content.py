from __future__ import annotations

from pathlib import Path

import pytest

from praxis_assessment.config import DEFAULT_CONTENT_DIR
from praxis_assessment.content import load_registry
from praxis_assessment.errors import ContentError, UnsupportedVersion
from support import edit_json


def test_shipped_content_loads_and_pins_resolve() -> None:
    registry = load_registry(DEFAULT_CONTENT_DIR)
    rubric = registry.rubric(registry.current.rubric).content
    assert sum(d.weight_bp for d in rubric.dimensions) == 10_000
    assert rubric.pilot_fixture_min_weighted_score == 75
    assert registry.role_standard("1.0.0").content.tracks["operator"] == ("C1", "C2", "C3", "C4", "C5", "C6")


def test_reformatting_a_file_does_not_change_its_hash(content_dir: Path) -> None:
    original = load_registry(content_dir).rubric("0.1.0").sha256
    edit_json(content_dir / "rubric" / "0.1.0.json", lambda _: None)  # rewrites with different whitespace
    assert load_registry(content_dir).rubric("0.1.0").sha256 == original


def test_unknown_version_is_an_explicit_error() -> None:
    with pytest.raises(UnsupportedVersion) as exc:
        load_registry(DEFAULT_CONTENT_DIR).rubric("9.9.9")
    assert exc.value.details["available"] == ["0.1.0"]


def _weights_off(data: dict) -> None:
    data["dimensions"][0]["weight_bp"] += 1


def _drop_parallel_twin(data: dict) -> None:
    data["scenarios"] = [s for s in data["scenarios"] if s["id"] != "CS-01-B"]


def _cycle(data: dict) -> None:
    next(c for c in data["courses"] if c["id"] == "M-ROLE")["prerequisites"] = ["M-VERIFY"]


def _remediation_gap(data: dict) -> None:
    del data["remediation"]["C6"]


def _remediation_wrong_course(data: dict) -> None:
    data["remediation"]["C2"] = ["M-TRIAGE"]


def _remediation_unknown_course(data: dict) -> None:
    data["remediation"]["C2"] = ["M-NOPE"]


def _challenge_with_prerequisites(data: dict) -> None:
    next(c for c in data["courses"] if c["id"] == "M-CHALLENGE")["prerequisites"] = ["M-ROLE"]


def _unknown_competency(data: dict) -> None:
    data["dimensions"][0]["competencies"] = ["C9"]


def _fake_validated_cutoff(data: dict) -> None:
    data["pilot_fixture_is_validated_cutoff"] = True


def _unbound_rubric(data: dict) -> None:
    data["scenarios"][0]["rubric_version"] = "0.9.0"


def _mismatched_blueprint(data: dict) -> None:
    data["scenarios"][0]["competencies"] = ["C1"]


@pytest.mark.parametrize(
    ("relative", "mutate", "message"),
    [
        ("rubric/0.1.0.json", _weights_off, "weights sum"),
        ("rubric/0.1.0.json", _unknown_competency, "unknown competencies"),
        ("rubric/0.1.0.json", _fake_validated_cutoff, "pilot_fixture_is_validated_cutoff"),
        ("scenarios/1.0.0.json", _drop_parallel_twin, "not reciprocal"),
        ("scenarios/1.0.0.json", _unbound_rubric, "missing rubric"),
        ("scenarios/1.0.0.json", _mismatched_blueprint, "blueprint"),
        ("curriculum/1.0.0.json", _cycle, "cycle"),
        ("curriculum/1.0.0.json", _challenge_with_prerequisites, "must have no prerequisites"),
        ("allocation_policy/1.0.0.json", _remediation_gap, r"missing \['C6'\]"),
        ("allocation_policy/1.0.0.json", _remediation_wrong_course, "does not cover it"),
        ("allocation_policy/1.0.0.json", _remediation_unknown_course, "unknown course M-NOPE"),
    ],
)
def test_broken_content_refuses_to_load(content_dir: Path, relative: str, mutate, message: str) -> None:
    edit_json(content_dir / relative, mutate)
    with pytest.raises(ContentError, match=message):
        load_registry(content_dir)


def test_file_name_must_match_declared_version(content_dir: Path) -> None:
    (content_dir / "rubric" / "0.1.0.json").rename(content_dir / "rubric" / "0.1.1.json")
    edit_json(content_dir / "current.json", lambda d: d.update(rubric="0.1.1"))
    with pytest.raises(ContentError, match="declares version 0.1.0"):
        load_registry(content_dir)


def test_current_pins_must_exist(content_dir: Path) -> None:
    edit_json(content_dir / "current.json", lambda d: d.update(curriculum="2.0.0"))
    with pytest.raises(ContentError, match="pins a missing version"):
        load_registry(content_dir)
