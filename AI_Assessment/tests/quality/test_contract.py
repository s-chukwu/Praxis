from __future__ import annotations

import hashlib
import json
from typing import Any

from fastapi.testclient import TestClient

# Fingerprint of every path, field, type and status code our Go backend depends on. Wording is left out
# (descriptions, titles, examples), so rewording a docstring never trips this; changing the contract does.
WORDING = {"examples", "example", "description", "summary", "title"}
CONTRACT_SHA256 = "9f3abeda99d6349588fffc1ec97072e17c05928d7b2e6cfcb00ce73194c9fb42"


def _shape(node: Any) -> Any:
    if isinstance(node, dict):
        return {k: _shape(v) for k, v in node.items() if k not in WORDING}
    if isinstance(node, list):
        return [_shape(v) for v in node]
    return node


def test_api_contract_only_changes_on_purpose(client: TestClient) -> None:
    schema = client.get("/openapi.json").json()
    digest = hashlib.sha256(json.dumps(_shape(schema), sort_keys=True).encode()).hexdigest()
    assert digest == CONTRACT_SHA256, (
        "The API contract changed. If that is intended, tell the Go team what changed, then set CONTRACT_SHA256 "
        f"to {digest}"
    )
