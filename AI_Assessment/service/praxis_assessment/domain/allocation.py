from __future__ import annotations

from collections.abc import Iterable, Mapping

from praxis_assessment.content.models import AllocationPolicy, Course, Curriculum, RoleStandard
from praxis_assessment.errors import InvalidInput
from praxis_assessment.schemas.learning_plan import (
    AssignedModule,
    CompetencyScore,
    LearningPlan,
    LearningPlanRequest,
    Reason,
    SafetyGate,
    SafetyTrigger,
    WaivedModule,
)


def allocate(
    standard: RoleStandard, policy: AllocationPolicy, curriculum: Curriculum, request: LearningPlanRequest
) -> LearningPlan:
    required = standard.required(request.target_track)
    scores = request.reviewed_competencies
    completed = frozenset(request.verified_completed_modules)
    _require_track_codes(request.target_track, required, scores)
    _require_known_courses(curriculum, completed)

    threshold = policy.remediation_threshold
    missing = tuple(code for code in required if code not in scores)
    reviewed = [CompetencyScore(code=c, reviewed_score=scores[c]) for c in required if c in scores]
    demonstrated = tuple(s for s in reviewed if s.reviewed_score >= threshold)
    gaps = tuple(s for s in reviewed if s.reviewed_score < threshold)
    gate = _safety_gate(policy, scores, request.critical_safety_finding)

    modules: tuple[AssignedModule, ...] = ()
    waived: tuple[WaivedModule, ...] = ()
    # Missing evidence is not a low score. I return what is known and refuse to plan until it's complete.
    if not missing:
        modules, waived = _plan(policy, curriculum, scores, gaps, gate, completed)

    return LearningPlan(
        status="diagnostic_incomplete" if missing else "plan_ready",
        threshold=threshold,
        missing_competencies=missing,
        demonstrated=demonstrated,
        gaps=gaps,
        safety_gate=gate,
        modules=modules,
        waived_prerequisites=waived,
        reassessment_required=True,
    )


def _plan(
    policy: AllocationPolicy,
    curriculum: Curriculum,
    scores: Mapping[str, int],
    gaps: tuple[CompetencyScore, ...],
    gate: SafetyGate,
    completed: frozenset[str],
) -> tuple[tuple[AssignedModule, ...], tuple[WaivedModule, ...]]:
    picked: dict[str, list[Reason]] = {}

    def pick(course_id: str, reason: Reason) -> None:
        picked.setdefault(course_id, []).append(reason)

    for course_id in policy.mandatory_modules:
        if course_id not in completed:
            pick(course_id, Reason(code="role_orientation"))

    # A gap re-assigns its module even if it was completed before: finishing a course is not
    # evidence of competence, the latest review is.
    for gap in gaps:
        for course_id in policy.remediation[gap.code]:
            pick(course_id, Reason(code="competency_gap", competency=gap.code, reviewed_score=gap.reviewed_score))

    if gate.required:
        pick(policy.safety_module, Reason(code="safety_gate"))

    # The challenge follows any operator-skill remediation, and is the whole plan for someone with
    # no gaps at all. A builder who only lacks bot skills doesn't redo operator content.
    challenge = curriculum.course(policy.challenge_module)
    if not gaps or {g.code for g in gaps} & set(challenge.covers):
        pick(challenge.id, Reason(code="integrated_challenge"))

    waived: dict[str, WaivedModule] = {}
    frontier = sorted(picked, key=curriculum.rank)
    while frontier:
        course_id = frontier.pop()
        for prerequisite in curriculum.course(course_id).prerequisites:
            if prerequisite in picked or prerequisite in waived:
                continue
            if prerequisite in completed:
                waived[prerequisite] = WaivedModule(
                    course_id=prerequisite, required_by=course_id, reason="verified_completed"
                )
            elif _demonstrated(curriculum.course(prerequisite), scores, policy.remediation_threshold):
                waived[prerequisite] = WaivedModule(
                    course_id=prerequisite, required_by=course_id, reason="competencies_demonstrated"
                )
            else:
                pick(prerequisite, Reason(code="prerequisite", required_by=course_id))
                frontier.append(prerequisite)

    # Sorting by the curriculum's fixed prerequisite order keeps every prerequisite ahead of its
    # dependents. The challenge has no static prerequisites, so I place it last explicitly.
    ordered = sorted((c for c in picked if c != challenge.id), key=curriculum.rank)
    if challenge.id in picked:
        ordered.append(challenge.id)

    modules = tuple(
        AssignedModule(
            order=position,
            course_id=course.id,
            title=course.title,
            covers=course.covers,
            reasons=tuple(picked[course.id]),
            previously_completed=course.id in completed,
        )
        for position, course in enumerate((curriculum.course(c) for c in ordered), start=1)
    )
    return modules, tuple(sorted(waived.values(), key=lambda w: curriculum.rank(w.course_id)))


def _demonstrated(course: Course, scores: Mapping[str, int], threshold: int) -> bool:
    return all(scores.get(code, -1) >= threshold for code in course.covers)


def _safety_gate(policy: AllocationPolicy, scores: Mapping[str, int], critical_finding: bool) -> SafetyGate:
    triggers: list[SafetyTrigger] = []
    safety_score = scores.get(policy.safety_competency)
    if safety_score is not None and safety_score < policy.remediation_threshold:
        triggers.append("safety_competency_below_threshold")
    if critical_finding:
        triggers.append("critical_safety_finding")
    return SafetyGate(required=bool(triggers), module=policy.safety_module, triggers=tuple(triggers))


def _require_track_codes(track: str, required: Iterable[str], scores: Mapping[str, int]) -> None:
    unexpected = sorted(scores.keys() - set(required))
    if unexpected:
        raise InvalidInput(
            f"competencies outside the {track} track", details={"unexpected": unexpected, "track": track}
        )


def _require_known_courses(curriculum: Curriculum, course_ids: Iterable[str]) -> None:
    known = {c.id for c in curriculum.courses}
    if unknown := sorted(set(course_ids) - known):
        raise InvalidInput(f"unknown courses for curriculum {curriculum.version}", details={"unknown": unknown})
