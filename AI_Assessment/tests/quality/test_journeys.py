from __future__ import annotations

import uuid

from fastapi.testclient import TestClient

from praxis_assessment.api.examples import CAPSTONE_EXAMPLES, SKILLS_EXAMPLES
from support import REVIEWER, plan_payload, with_scores

CURRENT = {"scenario_version": "1.2.0", "rubric_version": "0.2.0"}


def scored(call, scenario_id: str, scores: dict[str, int]) -> dict:
    payload = with_scores(**scores) | CURRENT | {"scenario_id": scenario_id}
    response = call("POST", "/internal/v1/diagnostics/score", payload, actor=REVIEWER)
    assert response.status_code == 200, response.text
    return {d["id"]: d["score"] for d in response.json()["result"]["dimensions"]}


def test_operator_journey_baseline_plan_final_compare(call) -> None:
    # Baseline on the current pins: weak on safety, so the reviewed competencies show a C4 gap.
    baseline = scored(call, "CS-01-A", {"accuracy": 2, "judgment": 2, "safety": 1, "communication": 3})
    plan = call("POST", "/internal/v1/learning-plans",
                plan_payload({"C1": 2, "C2": 3, "C3": 2, "C4": 1, "C5": 3, "C6": 3}), actor="system")  # fmt: skip
    assert "M-SAFE" in [m["course_id"] for m in plan.json()["result"]["modules"]]

    final = scored(call, "CS-01-B", {"accuracy": 3, "judgment": 3, "safety": 3, "communication": 3})
    attempt = {"state": "reviewed", **CURRENT}
    compare = call("POST", "/internal/v1/diagnostics/compare", {
        "request_id": str(uuid.uuid4()), "candidate_id": str(uuid.uuid4()),
        "baseline": attempt | {"attempt_id": str(uuid.uuid4()), "scenario_id": "CS-01-A", "dimensions": baseline},
        "final": attempt | {"attempt_id": str(uuid.uuid4()), "scenario_id": "CS-01-B", "dimensions": final},
    }, actor="system").json()["result"]  # fmt: skip
    assert compare["regressed_dimensions"] == [] and compare["weighted_delta"] > 0


def test_builder_journey_practice_then_held_out(call) -> None:
    practice = CAPSTONE_EXAMPLES["careful_practice"]["value"]
    candidate = f"candidate:{practice['candidate_id']}"
    assert call("POST", "/internal/v1/capstones/evaluate", practice, actor=candidate).json()["result"]["passed"] == 5
    held_out = practice | {"suite": "held_out"}
    response = call("POST", "/internal/v1/capstones/evaluate", held_out, actor="system").json()
    assert response["held_out"]["version"] == practice["config"]["capstone_version"]
    assert response["result"]["requires_human_review"] is True


def test_cv_journey_suggestions_never_shorten_training(client: TestClient, call) -> None:  # A01
    class Model:
        name = "fake/model"

        def suggest(self, instructions: str, cv_text: str) -> list:
            return [{"competency": "C4", "quote": "Escalated suspected account takeovers", "confidence": "high"}]

    client.app.state.skill_model = Model()  # type: ignore[attr-defined]
    suggested = call("POST", "/internal/v1/skills/suggest", SKILLS_EXAMPLES["call_centre_cv"]["value"], actor="system")
    assert [s["competency"] for s in suggested.json()["result"]["suggestions"]] == ["C4"]
    # The plan request has no field for CV claims, so a suggestion cannot waive the safety module.
    claims = call("POST", "/internal/v1/learning-plans", plan_payload(claimed_skills=["C4"]), actor="system")
    assert claims.status_code == 422

