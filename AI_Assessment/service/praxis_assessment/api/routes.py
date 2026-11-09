from __future__ import annotations

from typing import Annotated

from fastapi import APIRouter, Body, Depends, Query, Request

from praxis_assessment import SERVICE_NAME, SERVICE_VERSION
from praxis_assessment.api.auth import ALL_ROLES, Caller, require_caller
from praxis_assessment.api.errors import ERROR_RESPONSES
from praxis_assessment.api.examples import (
    CAPSTONE_EXAMPLES,
    COMPARE_EXAMPLES,
    PLAN_EXAMPLES,
    SCORE_EXAMPLES,
    SKILLS_EXAMPLES,
)
from praxis_assessment.content import ContentRegistry
from praxis_assessment.content.models import Scenario, ScenarioBank
from praxis_assessment.content.registry import content_hash
from praxis_assessment.domain.allocation import allocate
from praxis_assessment.domain.capstone import evaluate
from praxis_assessment.domain.comparison import compare_attempts
from praxis_assessment.domain.scoring import score_attempt
from praxis_assessment.domain.skills import instructions_for, mask_contacts, vet_suggestions
from praxis_assessment.errors import Forbidden, ModelUnavailable, UnsupportedVersion
from praxis_assessment.llm import SkillModel
from praxis_assessment.schemas.capstone import CapstoneRequest, CapstoneResponse
from praxis_assessment.schemas.catalog import AssessorPacketResponse, CatalogResponse
from praxis_assessment.schemas.common import ContentRef, ScenarioId, SemVer
from praxis_assessment.schemas.comparison import CompareRequest, CompareResponse
from praxis_assessment.schemas.learning_plan import LearningPlanRequest, LearningPlanResponse
from praxis_assessment.schemas.scoring import ScoreRequest, ScoreResponse
from praxis_assessment.schemas.skills import SkillsRequest, SkillsResponse, SkillsResult

health_router = APIRouter()
internal_router = APIRouter(prefix="/internal/v1", responses=ERROR_RESPONSES)


def get_registry(request: Request) -> ContentRegistry:
    return request.app.state.registry


Registry = Annotated[ContentRegistry, Depends(get_registry)]


def get_skill_model(request: Request) -> SkillModel | None:
    return request.app.state.skill_model


@health_router.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok", "service": SERVICE_NAME, "version": SERVICE_VERSION}


@internal_router.get("/catalog")
async def catalog(
    _caller: Annotated[Caller, Depends(require_caller(*ALL_ROLES))], registry: Registry
) -> CatalogResponse:
    kinds = {
        "role_standard": registry.role_standards,
        "rubric": registry.rubrics,
        "scenarios": registry.scenario_banks,
        "curriculum": registry.curricula,
        "allocation_policy": registry.allocation_policies,
        "attempt_workflow": registry.attempt_workflows,
        "capstone": registry.capstones,
    }
    return CatalogResponse(
        current=registry.current,
        role_standards=tuple(v.content for v in registry.role_standards.values()),
        rubrics=tuple(v.content for v in registry.rubrics.values()),
        scenario_banks=tuple(v.content for v in registry.scenario_banks.values()),
        curricula=tuple(v.content for v in registry.curricula.values()),
        allocation_policies=tuple(v.content for v in registry.allocation_policies.values()),
        attempt_workflows=tuple(v.content for v in registry.attempt_workflows.values()),
        capstones=tuple(v.content for v in registry.capstones.values()),
        sha256={kind: {version: v.sha256 for version, v in items.items()} for kind, items in kinds.items()},
    )


@internal_router.get("/scenarios/{scenario_id}/assessor-packet")
async def assessor_packet(
    scenario_id: ScenarioId,
    version: Annotated[SemVer, Query()],
    _caller: Annotated[Caller, Depends(require_caller("reviewer", "admin"))],
    registry: Registry,
) -> AssessorPacketResponse:
    bank = registry.scenario_bank(version)
    keys = registry.assessor_key_set(version)
    scenario = bank.content.scenario(scenario_id)
    return AssessorPacketResponse(
        scenarios=ContentRef(version=version, sha256=bank.sha256),
        assessor_keys=ContentRef(version=version, sha256=keys.sha256),
        scenario=scenario,
        key=keys.content.key(scenario.id),
    )


@internal_router.post("/diagnostics/score")
async def score(
    body: Annotated[ScoreRequest, Body(openapi_examples=SCORE_EXAMPLES)],
    caller: Annotated[Caller, Depends(require_caller("reviewer"))],
    registry: Registry,
) -> ScoreResponse:
    # The signature proves our Go backend vouches for this reviewer; the body must not claim a different one.
    if caller.actor_id != body.reviewer_id:
        raise Forbidden("reviewer_id does not match the signed actor")
    # Assignment rules belong to our Go backend; this only stops the one case I can see from here.
    if body.reviewer_id == body.candidate_id:
        raise Forbidden("a reviewer cannot score their own attempt")

    rubric = registry.rubric(body.rubric_version)
    bank = registry.scenario_bank(body.scenario_version)
    scenario = _bound_scenario(bank.content, body.scenario_id, body.rubric_version)

    return ScoreResponse(
        request_id=body.request_id,
        attempt_id=body.attempt_id,
        scenario_id=scenario.id,
        stage=scenario.stage,
        rubric=ContentRef(version=rubric.content.version, sha256=rubric.sha256),
        scenarios=ContentRef(version=bank.content.version, sha256=bank.sha256),
        result=score_attempt(rubric.content, scenario, body),
    )


@internal_router.post("/diagnostics/compare")
async def compare(
    body: Annotated[CompareRequest, Body(openapi_examples=COMPARE_EXAMPLES)],
    _caller: Annotated[Caller, Depends(require_caller("reviewer", "system"))],
    registry: Registry,
) -> CompareResponse:
    # I look content up by the baseline's versions; compare_attempts refuses a final scored on others.
    rubric = registry.rubric(body.baseline.rubric_version)
    bank = registry.scenario_bank(body.baseline.scenario_version)
    baseline = _bound_scenario(bank.content, body.baseline.scenario_id, rubric.content.version)
    final = _bound_scenario(bank.content, body.final.scenario_id, rubric.content.version)
    return CompareResponse(
        request_id=body.request_id,
        candidate_id=body.candidate_id,
        baseline_attempt_id=body.baseline.attempt_id,
        final_attempt_id=body.final.attempt_id,
        rubric=ContentRef(version=rubric.content.version, sha256=rubric.sha256),
        scenarios=ContentRef(version=bank.content.version, sha256=bank.sha256),
        result=compare_attempts(rubric.content, baseline, final, body),
    )


@internal_router.post("/capstones/evaluate")
async def evaluate_capstone(
    body: Annotated[CapstoneRequest, Body(openapi_examples=CAPSTONE_EXAMPLES)],
    caller: Annotated[Caller, Depends(require_caller("candidate", "reviewer", "system"))],
    registry: Registry,
) -> CapstoneResponse:
    # Candidates may rerun the practice cases freely. The held-out run is the assessment itself, so only
    # our Go backend (as system or a reviewer) starts it, and only once per locked attempt.
    if caller.role == "candidate":
        if body.suite == "held_out":
            raise Forbidden("candidates may run the practice suite only")
        if caller.actor_id != body.candidate_id:
            raise Forbidden("candidate_id does not match the signed actor")

    version = body.config.capstone_version
    capstone = registry.capstone(version)
    held_out = registry.held_out_cases(version) if body.suite == "held_out" else None
    cases = held_out.content.cases if held_out else capstone.content.practice
    return CapstoneResponse(
        request_id=body.request_id,
        attempt_id=body.attempt_id,
        candidate_id=body.candidate_id,
        capstone=ContentRef(version=version, sha256=capstone.sha256),
        held_out=ContentRef(version=version, sha256=held_out.sha256) if held_out else None,
        config_sha256=content_hash(body.config),
        result=evaluate(capstone.content, body.config, cases, body.suite),
    )


def _bound_scenario(bank: ScenarioBank, scenario_id: str, rubric_version: str) -> Scenario:
    scenario = bank.scenario(scenario_id)
    if scenario.rubric_version != rubric_version:
        raise UnsupportedVersion(
            f"scenario {scenario.id} is bound to rubric {scenario.rubric_version}",
            details={"requested": rubric_version, "bound": scenario.rubric_version},
        )
    return scenario


# A plain def on purpose: FastAPI runs it in a worker thread, so a slow LLM call never blocks other requests.
@internal_router.post("/skills/suggest")
def suggest_skills(
    body: Annotated[SkillsRequest, Body(openapi_examples=SKILLS_EXAMPLES)],
    _caller: Annotated[Caller, Depends(require_caller("system"))],
    registry: Registry,
    model: Annotated[SkillModel | None, Depends(get_skill_model)],
) -> SkillsResponse:
    standard = registry.role_standard(body.role_standard_version)
    if model is None:
        raise ModelUnavailable(
            "skill suggestions are not configured; review the CV manually", details={"reason": "not_configured"}
        )
    sent = mask_contacts(body.text)
    items = model.suggest(instructions_for(standard.content), sent)
    suggestions, dropped = vet_suggestions(standard.content, sent, items)
    return SkillsResponse(
        request_id=body.request_id,
        candidate_id=body.candidate_id,
        consent_ref=body.consent_ref,
        role_standard=ContentRef(version=standard.content.version, sha256=standard.sha256),
        result=SkillsResult(suggestions=suggestions, dropped=dropped, model=model.name),
    )


@internal_router.post("/learning-plans")
async def learning_plan(
    body: Annotated[LearningPlanRequest, Body(openapi_examples=PLAN_EXAMPLES)],
    _caller: Annotated[Caller, Depends(require_caller("reviewer", "system"))],
    registry: Registry,
) -> LearningPlanResponse:
    standard = registry.role_standard(body.role_standard_version)
    policy = registry.allocation_policy(body.policy_version)
    bound = policy.content.role_standard_version
    if bound != body.role_standard_version:
        raise UnsupportedVersion(
            f"allocation policy {policy.content.version} is bound to role standard {bound}",
            details={"requested": body.role_standard_version, "bound": bound},
        )
    curriculum = registry.curriculum(policy.content.curriculum_version)

    return LearningPlanResponse(
        request_id=body.request_id,
        candidate_id=body.candidate_id,
        review_id=body.review_id,
        target_track=body.target_track,
        role_standard=ContentRef(version=standard.content.version, sha256=standard.sha256),
        policy=ContentRef(version=policy.content.version, sha256=policy.sha256),
        curriculum=ContentRef(version=curriculum.content.version, sha256=curriculum.sha256),
        result=allocate(standard.content, policy.content, curriculum.content, body),
    )
