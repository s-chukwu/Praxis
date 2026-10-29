from __future__ import annotations

import json
from pathlib import Path

from fastapi.testclient import TestClient

from praxis_assessment.api import create_app
from praxis_assessment.config import Settings
from support import OPERATOR_STRONG, SECRET, edit_json, plan_payload, score_payload, signed_headers

SCORE = "/internal/v1/diagnostics/score"
CATALOG = "/internal/v1/catalog"
PLANS = "/internal/v1/learning-plans"


def test_health_needs_no_auth(client: TestClient) -> None:
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"
    assert response.headers["X-Request-ID"]


def test_static_catalog_route_is_gone(client: TestClient) -> None:
    response = client.get("/catalog")
    assert response.status_code == 404
    assert response.json()["error"]["code"] == "not_found"


def test_catalog_lists_versions_hashes_and_current_pins(call) -> None:
    response = call("GET", CATALOG, actor="system")
    assert response.status_code == 200
    body = response.json()
    assert body["current"] == {
        "role_standard": "1.0.0",
        "rubric": "0.2.0",
        "scenarios": "1.1.0",
        "curriculum": "1.0.0",
        "allocation_policy": "1.0.0",
    }
    assert set(body["sha256"]) == {"role_standard", "rubric", "scenarios", "curriculum", "allocation_policy"}
    assert len(body["sha256"]["rubric"]["0.1.0"]) == 64


def test_catalog_exposes_no_assessor_material(call) -> None:
    scenario_keys = {
        k
        for bank in call("GET", CATALOG, actor="system").json()["scenario_banks"]
        for s in bank["scenarios"]
        for k in s
    }
    assert scenario_keys == {
        "id", "stage", "pair", "parallel_form_id", "rubric_version", "competencies",
        "timebox_minutes", "expected_artifacts", "candidate_brief",
    }  # fmt: skip


def test_score_round_trip(call) -> None:
    payload = score_payload()
    response = call("POST", SCORE, payload)
    assert response.status_code == 200
    body = response.json()
    assert body["request_id"] == payload["request_id"]
    assert body["attempt_id"] == payload["attempt_id"]
    assert body["stage"] == "baseline"
    assert body["rubric"]["version"] == "0.1.0"
    assert len(body["rubric"]["sha256"]) == 64
    assert body["result"]["weighted_score"] == 82.5
    assert body["result"]["proposed_outcome"] == "evidence_verified"


def test_unknown_rubric_version_is_409(call) -> None:
    response = call("POST", SCORE, score_payload(rubric_version="9.9.9"))
    assert response.status_code == 409
    assert response.json()["error"]["code"] == "unsupported_version"


def test_unknown_scenario_is_422(call) -> None:
    response = call("POST", SCORE, score_payload(scenario_id="CS-99-A"))
    assert response.status_code == 422
    assert response.json()["error"]["code"] == "invalid_input"


def test_domain_rejection_uses_error_envelope(call) -> None:
    payload = score_payload()
    del payload["dimensions"]["safety"]
    error = call("POST", SCORE, payload).json()["error"]
    assert error["code"] == "invalid_input"
    assert error["details"] == {"missing": ["safety"], "unexpected": []}
    assert error["request_id"]


def test_request_id_is_propagated_from_caller(client: TestClient) -> None:
    response = client.get("/health", headers={"X-Request-ID": "go-trace-0001"})
    assert response.headers["X-Request-ID"] == "go-trace-0001"


def test_malformed_request_id_is_replaced(client: TestClient) -> None:
    response = client.get("/health", headers={"X-Request-ID": "<script>"})
    assert response.headers["X-Request-ID"] != "<script>"


def test_historical_attempt_keeps_its_rubric(content_dir: Path) -> None:
    """A new rubric version must not change the score of an attempt taken under the old one."""
    reweighted = content_dir / "rubric" / "0.3.0.json"
    reweighted.write_text((content_dir / "rubric" / "0.1.0.json").read_text())

    def bump(data: dict) -> None:
        data["version"] = "0.3.0"
        data["dimensions"][0]["weight_bp"], data["dimensions"][2]["weight_bp"] = 2000, 4000

    edit_json(reweighted, bump)
    bank = content_dir / "scenarios" / "1.2.0.json"
    bank.write_text((content_dir / "scenarios" / "1.0.0.json").read_text())

    def rebind(data: dict) -> None:
        data["version"] = "1.2.0"
        for s in data["scenarios"]:
            s["rubric_version"] = "0.3.0"

    edit_json(bank, rebind)
    keys = content_dir / "assessor_keys" / "1.2.0.json"
    keys.write_text((content_dir / "assessor_keys" / "1.0.0.json").read_text())
    edit_json(keys, lambda d: d.update(version="1.2.0"))

    client = TestClient(create_app(Settings(env="local", shared_secret=SECRET, content_dir=content_dir)))

    def score(scenario_version: str, rubric_version: str):
        body = json.dumps(score_payload(scenario_version=scenario_version, rubric_version=rubric_version)).encode()
        return client.post(SCORE, content=body, headers=signed_headers("POST", SCORE, body))

    old, new = score("1.0.0", "0.1.0"), score("1.2.0", "0.3.0")
    assert old.json()["result"]["weighted_score"] == 82.5
    assert new.json()["result"]["weighted_score"] == 85.0
    mixed = score("1.2.0", "0.1.0")
    assert mixed.status_code == 409
    assert mixed.json()["error"]["details"] == {"requested": "0.1.0", "bound": "0.3.0"}


def test_learning_plan_round_trip(call) -> None:
    payload = plan_payload({"C1": 3, "C2": 4, "C3": 3, "C4": 2, "C5": 1, "C6": 3})
    response = call("POST", PLANS, payload, actor="system")
    assert response.status_code == 200
    body = response.json()
    assert body["review_id"] == payload["review_id"]
    assert body["policy"]["version"] == "1.0.0"
    assert body["curriculum"]["version"] == "1.0.0"
    assert len(body["role_standard"]["sha256"]) == 64
    result = body["result"]
    assert result["status"] == "plan_ready"
    assert [m["course_id"] for m in result["modules"]] == ["M-ROLE", "M-SAFE", "M-VERIFY", "M-CHALLENGE"]
    assert result["reassessment_required"] is True


def test_incomplete_diagnostic_is_still_200(call) -> None:
    scores = {k: v for k, v in OPERATOR_STRONG.items() if k != "C2"}
    result = call("POST", PLANS, plan_payload(scores), actor="system").json()["result"]
    assert result["status"] == "diagnostic_incomplete"
    assert result["missing_competencies"] == ["C2"]
    assert result["modules"] == []


def test_unknown_policy_version_is_409(call) -> None:  # A07
    response = call("POST", PLANS, plan_payload(policy_version="9.9.9"), actor="system")
    assert response.status_code == 409
    assert response.json()["error"]["details"]["kind"] == "allocation_policy"


def test_unsupported_track_is_422(call) -> None:  # A07
    response = call("POST", PLANS, plan_payload(target_track="data_entry"), actor="system")
    assert response.status_code == 422


def test_policy_bound_to_another_role_standard_is_409(content_dir: Path) -> None:
    standard = content_dir / "role_standard" / "1.2.0.json"
    standard.write_text((content_dir / "role_standard" / "1.0.0.json").read_text())
    edit_json(standard, lambda d: d.update(version="1.2.0"))
    client = TestClient(create_app(Settings(env="local", shared_secret=SECRET, content_dir=content_dir)))

    body = json.dumps(plan_payload(role_standard_version="1.2.0")).encode()
    response = client.post(PLANS, content=body, headers=signed_headers("POST", PLANS, body, actor="system"))
    assert response.status_code == 409
    assert response.json()["error"]["details"] == {"requested": "1.2.0", "bound": "1.0.0"}
