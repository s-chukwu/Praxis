from __future__ import annotations

import json
import shutil
from collections.abc import Callable
from pathlib import Path
from typing import Any

import pytest
from fastapi.testclient import TestClient

from praxis_assessment.api import create_app
from praxis_assessment.config import DEFAULT_CONTENT_DIR, Settings
from support import REVIEWER, SECRET, signed_headers

SignedCall = Callable[..., Any]


@pytest.fixture
def settings() -> Settings:
    return Settings(env="local", shared_secret=SECRET)


@pytest.fixture
def client(settings: Settings) -> TestClient:
    return TestClient(create_app(settings))


@pytest.fixture
def call(client: TestClient) -> SignedCall:
    """Send a correctly signed request, the way our Go backend will."""

    def _call(method: str, path: str, payload: Any = None, *, actor: str = REVIEWER, **kwargs: Any) -> Any:
        body = b"" if payload is None else json.dumps(payload).encode()
        headers = signed_headers(method, path, body, actor=actor, **kwargs)
        return client.request(method, path, content=body, headers=headers)

    return _call


@pytest.fixture
def content_dir(tmp_path: Path) -> Path:
    """A writable copy of the shipped content, for integrity and version tests."""
    target = tmp_path / "content"
    shutil.copytree(DEFAULT_CONTENT_DIR, target)
    return target
