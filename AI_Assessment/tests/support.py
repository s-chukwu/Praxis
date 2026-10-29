from __future__ import annotations

import json
import time
import uuid
from collections.abc import Callable
from pathlib import Path
from typing import Any

from praxis_assessment.api.auth import ACTOR_HEADER, SIGNATURE_HEADER, TIMESTAMP_HEADER, sign_request

SECRET = b"test-secret-that-is-at-least-32-bytes!!"
REVIEWER_ID = uuid.UUID("11111111-1111-4111-8111-111111111111")
REVIEWER = f"reviewer:{REVIEWER_ID}"

ARTIFACTS = [
    {
        "type": "customer_reply",
        "text": "I understand how frustrating this is. I can see one captured payment of RWF 42,000.\n"
        "The second entry is a pending authorization, not a settled charge.",
    },
    {
        "type": "internal_note",
        "text": "Escalated to billing to confirm whether the pending authorization settles. No refund promised.",
    },
]


def score_payload(**overrides: Any) -> dict[str, Any]:
    """A valid CS-01-A review: accuracy 3, judgment 3, safety 4, communication 3 -> 82.5 per rubric.md."""
    payload: dict[str, Any] = {
        "request_id": str(uuid.uuid4()),
        "attempt_id": str(uuid.uuid4()),
        "candidate_id": str(uuid.uuid4()),
        "reviewer_id": str(REVIEWER_ID),
        "scenario_id": "CS-01-A",
        "scenario_version": "1.0.0",
        "rubric_version": "0.1.0",
        "dimensions": {
            "accuracy": {"score": 3, "evidence": ["one captured payment of RWF 42,000"]},
            "judgment": {"score": 3, "evidence": ["Escalated to billing"]},
            "safety": {"score": 4, "evidence": ["No refund promised."]},
            "communication": {"score": 3, "evidence": ["I understand how frustrating this is."]},
        },
        "critical_findings": [],
        "artifacts": ARTIFACTS,
    }
    payload.update(overrides)
    return payload


def with_scores(**scores: int) -> dict[str, Any]:
    payload = score_payload()
    for dim, value in scores.items():
        payload["dimensions"][dim]["score"] = value
    return payload


def signed_headers(
    method: str, path: str, body: bytes, *, actor: str = REVIEWER, timestamp: int | None = None, secret: bytes = SECRET
) -> dict[str, str]:
    ts = str(int(time.time()) if timestamp is None else timestamp)
    signature = sign_request(secret, timestamp=ts, method=method, path=path, actor=actor, body=body)
    return {
        TIMESTAMP_HEADER: ts,
        ACTOR_HEADER: actor,
        SIGNATURE_HEADER: signature,
        "Content-Type": "application/json",
    }


def edit_json(path: Path, mutate: Callable[[dict[str, Any]], None]) -> None:
    data = json.loads(path.read_text())
    mutate(data)
    path.write_text(json.dumps(data))


OPERATOR_STRONG = {"C1": 3, "C2": 3, "C3": 3, "C4": 3, "C5": 3, "C6": 3}


def plan_payload(scores: dict[str, int] | None = None, **overrides: Any) -> dict[str, Any]:
    payload: dict[str, Any] = {
        "request_id": str(uuid.uuid4()),
        "candidate_id": str(uuid.uuid4()),
        "review_id": str(uuid.uuid4()),
        "target_track": "operator",
        "role_standard_version": "1.0.0",
        "policy_version": "1.0.0",
        "reviewed_competencies": OPERATOR_STRONG if scores is None else scores,
        "critical_safety_finding": False,
    }
    payload.update(overrides)
    return payload
