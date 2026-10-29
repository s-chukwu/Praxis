from __future__ import annotations

from praxis_assessment.content.models import (
    AllocationPolicy,
    CurrentPins,
    Curriculum,
    RoleStandard,
    Rubric,
    ScenarioBank,
)
from praxis_assessment.schemas.common import Frozen


class CatalogResponse(Frozen):
    """Every loaded version, so our Go backend can render historical attempts against the content they used.

    Scenario banks carry candidate-visible briefs only; assessor checklists are not part of the
    content model, so they cannot leak through here.
    """

    current: CurrentPins
    role_standards: tuple[RoleStandard, ...]
    rubrics: tuple[Rubric, ...]
    scenario_banks: tuple[ScenarioBank, ...]
    curricula: tuple[Curriculum, ...]
    allocation_policies: tuple[AllocationPolicy, ...]
    sha256: dict[str, dict[str, str]]
