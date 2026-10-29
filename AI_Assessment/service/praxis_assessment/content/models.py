from __future__ import annotations

import heapq
import re
from fractions import Fraction
from typing import Annotated, Literal

from pydantic import Field, PrivateAttr, StringConstraints, model_validator

from praxis_assessment.errors import InvalidInput
from praxis_assessment.schemas.common import (
    CompetencyCode,
    Frozen,
    ScenarioId,
    SemVer,
    Slug,
    require_unique,
)

Track = Literal["operator", "builder"]
Stage = Literal["baseline", "final", "practice", "security_regression"]
CourseId = Annotated[str, StringConstraints(pattern=r"^M-[A-Z-]+$")]

# Pilot value, not validated. Shipped twins share no three-word run at all, so this only trips on a near copy.
MAX_TWIN_OVERLAP = Fraction(3, 10)


class Competency(Frozen):
    code: CompetencyCode
    name: str
    critical_when: str | None


class RoleStandard(Frozen):
    id: str
    version: SemVer
    title: str
    competencies: tuple[Competency, ...]
    tracks: dict[Track, tuple[CompetencyCode, ...]]

    @model_validator(mode="after")
    def _integrity(self) -> RoleStandard:
        codes = [c.code for c in self.competencies]
        require_unique(codes, "competency")
        for track, required in self.tracks.items():
            require_unique(required, f"{track} competency")
            if unknown := set(required) - set(codes):
                raise ValueError(f"track {track} references unknown competencies {sorted(unknown)}")
        return self

    @property
    def codes(self) -> frozenset[str]:
        return frozenset(c.code for c in self.competencies)

    def required(self, track: Track) -> tuple[str, ...]:
        return self.tracks[track]


class Dimension(Frozen):
    id: Slug
    weight_bp: int = Field(gt=0, le=10_000)
    competencies: tuple[CompetencyCode, ...] = Field(min_length=1)
    evaluates: str


class Rubric(Frozen):
    version: SemVer
    role_standard_version: SemVer
    scale_max: Literal[4]
    anchors: dict[Literal["0", "1", "2", "3", "4"], str]
    dimensions: tuple[Dimension, ...] = Field(min_length=1)
    critical_event_types: tuple[Slug, ...] = Field(min_length=1)
    pilot_fixture_min_weighted_score: float = Field(ge=0, le=100)
    pilot_fixture_is_validated_cutoff: Literal[False]

    @model_validator(mode="after")
    def _integrity(self) -> Rubric:
        require_unique([d.id for d in self.dimensions], "dimension")
        require_unique(self.critical_event_types, "critical event type")
        # Weights live in basis points so the weighted score is exact; float weights drift at the threshold.
        total = sum(d.weight_bp for d in self.dimensions)
        if total != 10_000:
            raise ValueError(f"dimension weights sum to {total} bp, expected 10000")
        if len(self.anchors) != self.scale_max + 1:
            raise ValueError("every point on the scale needs an anchor")
        return self


class Scenario(Frozen):
    id: ScenarioId
    stage: Stage
    pair: str
    parallel_form_id: ScenarioId | None
    rubric_version: SemVer
    competencies: tuple[CompetencyCode, ...] = Field(min_length=1)
    timebox_minutes: int = Field(gt=0, le=120)
    expected_artifacts: tuple[Slug, ...] = Field(min_length=1)
    candidate_brief: str


class ScenarioBank(Frozen):
    version: SemVer
    scenarios: tuple[Scenario, ...] = Field(min_length=1)

    @model_validator(mode="after")
    def _integrity(self) -> ScenarioBank:
        require_unique([s.id for s in self.scenarios], "scenario")
        by_id = {s.id: s for s in self.scenarios}
        for s in self.scenarios:
            if s.parallel_form_id is None:
                continue
            twin = by_id.get(s.parallel_form_id)
            # A parallel form that measures a different blueprint makes baseline-to-final deltas meaningless.
            if twin is None or twin.parallel_form_id != s.id:
                raise ValueError(f"{s.id}: parallel form is missing or not reciprocal")
            if set(twin.competencies) != set(s.competencies) or twin.pair != s.pair:
                raise ValueError(f"{s.id} and {twin.id} do not share a blueprint")
            # Twins that read alike let a candidate carry a baseline answer into the final.
            if brief_overlap(twin.candidate_brief, s.candidate_brief) > MAX_TWIN_OVERLAP:
                raise ValueError(f"{s.id} and {twin.id} briefs overlap too much to be parallel forms")
        return self

    def scenario(self, scenario_id: str) -> Scenario:
        for s in self.scenarios:
            if s.id == scenario_id:
                return s
        raise InvalidInput(
            f"scenario {scenario_id} is not in scenario bank {self.version}",
            details={"scenario_id": scenario_id, "scenario_version": self.version},
        )


def brief_overlap(a: str, b: str) -> Fraction:
    """Jaccard similarity of lower-cased three-word runs."""
    runs_a, runs_b = _word_runs(a), _word_runs(b)
    union = runs_a | runs_b
    return Fraction(len(runs_a & runs_b), len(union)) if union else Fraction(1)


def _word_runs(text: str) -> set[tuple[str, ...]]:
    words = re.findall(r"\w+", text.lower())
    return {tuple(words[i : i + 3]) for i in range(len(words) - 2)}


class CriticalEvent(Frozen):
    event_type: Slug
    when: str


class AssessorKey(Frozen):
    scenario_id: ScenarioId
    checks: tuple[str, ...] = Field(min_length=1)
    critical_events: tuple[CriticalEvent, ...]


# I keep keys out of ScenarioBank on purpose: the catalog serialises the bank, and a key that is not
# in the model cannot leak through it. The version matches the scenario bank it grades.
class AssessorKeys(Frozen):
    version: SemVer
    keys: tuple[AssessorKey, ...] = Field(min_length=1)

    @model_validator(mode="after")
    def _integrity(self) -> AssessorKeys:
        require_unique([k.scenario_id for k in self.keys], "assessor key")
        return self

    def key(self, scenario_id: str) -> AssessorKey:
        # The registry guarantees one key per scenario, so a scenario that resolved always has one.
        return next(k for k in self.keys if k.scenario_id == scenario_id)


class Course(Frozen):
    id: CourseId
    title: str
    covers: tuple[CompetencyCode, ...] = Field(min_length=1)
    prerequisites: tuple[CourseId, ...]


class Curriculum(Frozen):
    version: SemVer
    courses: tuple[Course, ...] = Field(min_length=1)

    _by_id: dict[str, Course] = PrivateAttr()
    _rank: dict[str, int] = PrivateAttr()

    @model_validator(mode="after")
    def _integrity(self) -> Curriculum:
        ids = [c.id for c in self.courses]
        require_unique(ids, "course")
        for course in self.courses:
            if missing := set(course.prerequisites) - set(ids):
                raise ValueError(f"{course.id} requires unknown courses {sorted(missing)}")
        self._by_id = {c.id: c for c in self.courses}
        self._rank = {course_id: i for i, course_id in enumerate(_prerequisite_order(self.courses))}
        return self

    def course(self, course_id: str) -> Course:
        return self._by_id[course_id]

    def rank(self, course_id: str) -> int:
        """Position in a fixed prerequisite-first order. Sorting any subset by it keeps prerequisites first."""
        return self._rank[course_id]


def _prerequisite_order(courses: tuple[Course, ...]) -> list[str]:
    # Kahn's algorithm, breaking ties by position in the file so authors control the order of
    # independent modules and the result never depends on dict or set iteration.
    position = {c.id: i for i, c in enumerate(courses)}
    waiting_on = {c.id: set(c.prerequisites) for c in courses}
    unlocks: dict[str, list[str]] = {c.id: [] for c in courses}
    for c in courses:
        for prerequisite in c.prerequisites:
            unlocks[prerequisite].append(c.id)

    ready = [(position[cid], cid) for cid, deps in waiting_on.items() if not deps]
    heapq.heapify(ready)
    order: list[str] = []
    while ready:
        _, cid = heapq.heappop(ready)
        order.append(cid)
        for nxt in unlocks[cid]:
            waiting_on[nxt].discard(cid)
            if not waiting_on[nxt]:
                heapq.heappush(ready, (position[nxt], nxt))

    if len(order) != len(courses):
        stuck = sorted(cid for cid, deps in waiting_on.items() if deps)
        raise ValueError(f"prerequisite cycle among {', '.join(stuck)}")
    return order


class AllocationPolicy(Frozen):
    """How reviewed gaps become modules. Cross-checked against its curriculum by the registry."""

    version: SemVer
    role_standard_version: SemVer
    curriculum_version: SemVer
    remediation_threshold: int = Field(ge=1, le=4)
    mandatory_modules: tuple[CourseId, ...]
    safety_competency: CompetencyCode
    safety_module: CourseId
    challenge_module: CourseId
    remediation: dict[CompetencyCode, tuple[CourseId, ...]]


class CurrentPins(Frozen):
    """Versions our Go backend should use for new attempts. Old attempts keep the versions they were created with."""

    role_standard: SemVer
    rubric: SemVer
    scenarios: SemVer
    curriculum: SemVer
    allocation_policy: SemVer
