from __future__ import annotations

import re
import unicodedata
from collections.abc import Iterable, Mapping
from fractions import Fraction

from praxis_assessment.content.models import Rubric, Scenario
from praxis_assessment.errors import InvalidInput
from praxis_assessment.schemas.scoring import (
    Artifact,
    CriticalFinding,
    DimensionRating,
    DimensionScore,
    EvidenceIssue,
    OutcomeReason,
    ProposedOutcome,
    ScoreRequest,
    ScoreResult,
)

_WHITESPACE = re.compile(r"\s+")


def score_attempt(rubric: Rubric, scenario: Scenario, request: ScoreRequest) -> ScoreResult:
    _require_rubric_dimensions(rubric, request.dimensions)
    _require_known_events(rubric, request.critical_findings)
    _require_expected_artifacts(scenario, request.artifacts)

    # 100 * sum(w * s / scale_max) with w in basis points reduces to sum(w_bp * s) / (scale_max * 100).
    # I keep it as a Fraction so the threshold comparison is exact.
    denominator = rubric.scale_max * 100
    dimensions = tuple(
        DimensionScore(
            id=d.id,
            score=request.dimensions[d.id].score,
            weight_bp=d.weight_bp,
            weighted_points=float(Fraction(d.weight_bp * request.dimensions[d.id].score, denominator)),
        )
        for d in rubric.dimensions
    )
    weighted = Fraction(sum(d.weight_bp * d.score for d in dimensions), denominator)
    threshold = Fraction(str(rubric.pilot_fixture_min_weighted_score))

    issues = _unverified_quotes(rubric, request)
    confirmed = _events(request.critical_findings, "confirmed")
    pending = _events(request.critical_findings, "pending")
    meets_threshold = weighted >= threshold

    reasons: list[OutcomeReason] = []
    if pending:
        reasons.append("critical_finding_pending")
    if issues:
        reasons.append("evidence_not_found")
    if confirmed:
        reasons.append("critical_finding_confirmed")
    reasons.append("meets_pilot_threshold" if meets_threshold else "below_pilot_threshold")

    return ScoreResult(
        dimensions=dimensions,
        weighted_score=float(weighted),
        pilot_threshold=float(threshold),
        pilot_threshold_is_validated=False,
        critical_veto=bool(confirmed),
        confirmed_critical_events=confirmed,
        pending_critical_events=pending,
        evidence_issues=issues,
        proposed_outcome=_outcome(
            unresolved=bool(pending or issues), vetoed=bool(confirmed), meets_threshold=meets_threshold
        ),
        outcome_reasons=tuple(reasons),
        score_source="human_review",
        requires_signoff=True,
    )


def _outcome(*, unresolved: bool, vetoed: bool, meets_threshold: bool) -> ProposedOutcome:
    # Order matters. Unresolved evidence beats everything because the score itself is not trustworthy yet;
    # a confirmed critical event beats the aggregate because a high average must never hide a safety failure.
    if unresolved:
        return "review_pending"
    if vetoed:
        return "needs_practice"
    return "evidence_verified" if meets_threshold else "needs_practice"


def _events(findings: Iterable[CriticalFinding], decision: str) -> tuple[str, ...]:
    return tuple(sorted({f.event_type for f in findings if f.decision == decision}))


def _normalize(text: str) -> str:
    return _WHITESPACE.sub(" ", unicodedata.normalize("NFC", text)).strip()


def _unverified_quotes(rubric: Rubric, request: ScoreRequest) -> tuple[EvidenceIssue, ...]:
    # A citation must appear in what the candidate actually submitted. I only forgive whitespace and
    # Unicode composition differences, never wording: a paraphrased quote is not evidence.
    corpus = "\n".join(_normalize(a.text) for a in request.artifacts)
    issues = [
        EvidenceIssue(source="dimension", ref=d.id, quote=quote, reason="not_found_in_artifacts")
        for d in rubric.dimensions
        for quote in request.dimensions[d.id].evidence
        if _normalize(quote) not in corpus
    ]
    issues.extend(
        EvidenceIssue(
            source="critical_finding", ref=f.event_type, quote=f.supporting_quote, reason="not_found_in_artifacts"
        )
        for f in request.critical_findings
        if _normalize(f.supporting_quote) not in corpus
    )
    return tuple(issues)


def _require_rubric_dimensions(rubric: Rubric, ratings: Mapping[str, DimensionRating]) -> None:
    expected = {d.id for d in rubric.dimensions}
    missing, unexpected = sorted(expected - ratings.keys()), sorted(ratings.keys() - expected)
    if missing or unexpected:
        raise InvalidInput(
            f"ratings must cover exactly the dimensions of rubric {rubric.version}",
            details={"missing": missing, "unexpected": unexpected},
        )


def _require_known_events(rubric: Rubric, findings: Iterable[CriticalFinding]) -> None:
    unknown = sorted({f.event_type for f in findings} - set(rubric.critical_event_types))
    if unknown:
        raise InvalidInput(
            f"unknown critical event types for rubric {rubric.version}",
            details={"unknown": unknown, "allowed": list(rubric.critical_event_types)},
        )


def _require_expected_artifacts(scenario: Scenario, artifacts: Iterable[Artifact]) -> None:
    unexpected = sorted({a.type for a in artifacts} - set(scenario.expected_artifacts))
    if unexpected:
        raise InvalidInput(
            f"artifact types not expected by scenario {scenario.id}",
            details={"unexpected": unexpected, "allowed": list(scenario.expected_artifacts)},
        )
