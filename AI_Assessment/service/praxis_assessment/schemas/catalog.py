from __future__ import annotations

from praxis_assessment.content.capstone import Capstone
from praxis_assessment.content.models import (
    AllocationPolicy,
    AssessorKey,
    AttemptWorkflow,
    CurrentPins,
    Curriculum,
    RoleStandard,
    Rubric,
    Scenario,
    ScenarioBank,
)
from praxis_assessment.schemas.common import ContentRef, Frozen


class CatalogResponse(Frozen):
    """Every loaded version, so our Go backend can render historical attempts against the content they used.

    Assessor keys and held-out capstone cases are separate models that this response has no field for,
    so they cannot leak through here.
    """

    current: CurrentPins
    role_standards: tuple[RoleStandard, ...]
    rubrics: tuple[Rubric, ...]
    scenario_banks: tuple[ScenarioBank, ...]
    curricula: tuple[Curriculum, ...]
    allocation_policies: tuple[AllocationPolicy, ...]
    attempt_workflows: tuple[AttemptWorkflow, ...]
    capstones: tuple[Capstone, ...]
    sha256: dict[str, dict[str, str]]


class AssessorPacketResponse(Frozen):
    """Reviewer-only: the scenario with its grading key. Never route this to a candidate."""

    scenarios: ContentRef
    assessor_keys: ContentRef
    scenario: Scenario
    key: AssessorKey
