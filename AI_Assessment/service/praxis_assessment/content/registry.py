from __future__ import annotations

import hashlib
import json
from collections.abc import Mapping
from dataclasses import dataclass
from pathlib import Path
from types import MappingProxyType

from pydantic import BaseModel, ValidationError

from praxis_assessment.content.models import (
    AllocationPolicy,
    AssessorKeys,
    CurrentPins,
    Curriculum,
    RoleStandard,
    Rubric,
    ScenarioBank,
)
from praxis_assessment.errors import ContentError, UnsupportedVersion


@dataclass(frozen=True, slots=True)
class Versioned[T: BaseModel]:
    content: T
    sha256: str


def _content_hash(model: BaseModel) -> str:
    # I hash the canonical JSON, not the file bytes, so reformatting a file does not change its identity.
    canonical = json.dumps(model.model_dump(mode="json"), sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonical.encode()).hexdigest()


def _load_kind[T: BaseModel](root: Path, kind: str, model: type[T]) -> Mapping[str, Versioned[T]]:
    loaded: dict[str, Versioned[T]] = {}
    for path in sorted((root / kind).glob("*.json")):
        try:
            item = model.model_validate_json(path.read_bytes())
        except ValidationError as exc:
            raise ContentError(f"{kind}/{path.name}: {exc}") from exc
        version = item.version  # type: ignore[attr-defined]
        if version != path.stem:
            raise ContentError(f"{kind}/{path.name} declares version {version}")
        loaded[version] = Versioned(item, _content_hash(item))
    if not loaded:
        raise ContentError(f"no {kind} content in {root / kind}")
    return MappingProxyType(loaded)


# There is no "latest" fallback. An attempt names its versions, and one that isn't loaded is a 409.
def _lookup[T: BaseModel](items: Mapping[str, Versioned[T]], kind: str, version: str) -> Versioned[T]:
    try:
        return items[version]
    except KeyError:
        raise UnsupportedVersion(
            f"{kind} version {version} is not available",
            details={"kind": kind, "requested": version, "available": sorted(items)},
        ) from None


@dataclass(frozen=True, slots=True)
class ContentRegistry:
    role_standards: Mapping[str, Versioned[RoleStandard]]
    rubrics: Mapping[str, Versioned[Rubric]]
    scenario_banks: Mapping[str, Versioned[ScenarioBank]]
    curricula: Mapping[str, Versioned[Curriculum]]
    allocation_policies: Mapping[str, Versioned[AllocationPolicy]]
    assessor_keys: Mapping[str, Versioned[AssessorKeys]]
    current: CurrentPins

    def role_standard(self, version: str) -> Versioned[RoleStandard]:
        return _lookup(self.role_standards, "role_standard", version)

    def rubric(self, version: str) -> Versioned[Rubric]:
        return _lookup(self.rubrics, "rubric", version)

    def scenario_bank(self, version: str) -> Versioned[ScenarioBank]:
        return _lookup(self.scenario_banks, "scenarios", version)

    def assessor_key_set(self, version: str) -> Versioned[AssessorKeys]:
        return _lookup(self.assessor_keys, "assessor_keys", version)

    def curriculum(self, version: str) -> Versioned[Curriculum]:
        return _lookup(self.curricula, "curriculum", version)

    def allocation_policy(self, version: str) -> Versioned[AllocationPolicy]:
        return _lookup(self.allocation_policies, "allocation_policy", version)


def load_registry(root: Path) -> ContentRegistry:
    try:
        current = CurrentPins.model_validate_json((root / "current.json").read_bytes())
    except (OSError, ValidationError) as exc:
        raise ContentError(f"current.json: {exc}") from exc

    registry = ContentRegistry(
        role_standards=_load_kind(root, "role_standard", RoleStandard),
        rubrics=_load_kind(root, "rubric", Rubric),
        scenario_banks=_load_kind(root, "scenarios", ScenarioBank),
        curricula=_load_kind(root, "curriculum", Curriculum),
        allocation_policies=_load_kind(root, "allocation_policy", AllocationPolicy),
        assessor_keys=_load_kind(root, "assessor_keys", AssessorKeys),
        current=current,
    )
    _check_references(registry)
    return registry


def _check_references(reg: ContentRegistry) -> None:
    for version, rubric in reg.rubrics.items():
        standard = reg.role_standards.get(rubric.content.role_standard_version)
        if standard is None:
            raise ContentError(f"rubric {version} needs missing role standard {rubric.content.role_standard_version}")
        for dim in rubric.content.dimensions:
            if unknown := set(dim.competencies) - standard.content.codes:
                raise ContentError(f"rubric {version} dimension {dim.id} names unknown competencies {sorted(unknown)}")

    for version, bank in reg.scenario_banks.items():
        for scenario in bank.content.scenarios:
            rubric = reg.rubrics.get(scenario.rubric_version)
            if rubric is None:
                raise ContentError(f"scenarios {version}/{scenario.id} needs missing rubric {scenario.rubric_version}")
            codes = reg.role_standards[rubric.content.role_standard_version].content.codes
            if unknown := set(scenario.competencies) - codes:
                raise ContentError(f"scenarios {version}/{scenario.id} names unknown competencies {sorted(unknown)}")

    if unpaired := sorted(reg.scenario_banks.keys() ^ reg.assessor_keys.keys()):
        raise ContentError(f"scenario banks and assessor keys must come in pairs, unpaired versions {unpaired}")
    for version, keys in reg.assessor_keys.items():
        _check_keys(reg, version, keys.content)

    known_codes = frozenset().union(*(s.content.codes for s in reg.role_standards.values()))
    for version, curriculum in reg.curricula.items():
        for course in curriculum.content.courses:
            if unknown := set(course.covers) - known_codes:
                raise ContentError(f"curriculum {version}/{course.id} covers unknown competencies {sorted(unknown)}")

    for version, policy in reg.allocation_policies.items():
        _check_policy(reg, version, policy.content)

    pins = reg.current
    try:
        reg.role_standard(pins.role_standard)
        reg.rubric(pins.rubric)
        reg.scenario_bank(pins.scenarios)
        reg.curriculum(pins.curriculum)
        reg.allocation_policy(pins.allocation_policy)
    except UnsupportedVersion as exc:
        raise ContentError(f"current.json pins a missing version: {exc.message}") from exc


def _check_policy(reg: ContentRegistry, version: str, policy: AllocationPolicy) -> None:
    where = f"allocation_policy {version}"
    standard = reg.role_standards.get(policy.role_standard_version)
    curriculum = reg.curricula.get(policy.curriculum_version)
    if standard is None or curriculum is None:
        raise ContentError(
            f"{where} needs role standard {policy.role_standard_version} and curriculum {policy.curriculum_version}"
        )
    courses = {c.id: c for c in curriculum.content.courses}

    # Every competency needs a remediation route, otherwise a gap would silently produce no training.
    if policy.remediation.keys() != standard.content.codes:
        missing = sorted(standard.content.codes - policy.remediation.keys())
        extra = sorted(policy.remediation.keys() - standard.content.codes)
        raise ContentError(
            f"{where} remediation must cover exactly the role standard: missing {missing}, extra {extra}"
        )
    for code, modules in policy.remediation.items():
        for module in modules:
            if module not in courses:
                raise ContentError(f"{where} remediates {code} with unknown course {module}")
            if code not in courses[module].covers:
                raise ContentError(f"{where} remediates {code} with {module}, which does not cover it")

    named = (*policy.mandatory_modules, policy.safety_module, policy.challenge_module)
    if unknown := sorted(set(named) - courses.keys()):
        raise ContentError(f"{where} names unknown courses {unknown}")
    if policy.safety_competency not in courses[policy.safety_module].covers:
        raise ContentError(f"{where} safety module {policy.safety_module} does not cover {policy.safety_competency}")
    # I always place the challenge last, so it must not carry prerequisites of its own.
    if courses[policy.challenge_module].prerequisites:
        raise ContentError(f"{where} challenge module {policy.challenge_module} must have no prerequisites")


def _check_keys(reg: ContentRegistry, version: str, keys: AssessorKeys) -> None:
    scenarios = {s.id: s for s in reg.scenario_banks[version].content.scenarios}
    keyed = {k.scenario_id for k in keys.keys}
    if keyed != scenarios.keys():
        raise ContentError(
            f"assessor_keys {version} must key every scenario exactly: "
            f"unkeyed {sorted(scenarios.keys() - keyed)}, unknown {sorted(keyed - scenarios.keys())}"
        )
    for key in keys.keys:
        allowed = reg.rubrics[scenarios[key.scenario_id].rubric_version].content.critical_event_types
        for event in key.critical_events:
            if event.event_type not in allowed:
                raise ContentError(f"assessor_keys {version}/{key.scenario_id} uses unknown event {event.event_type}")
