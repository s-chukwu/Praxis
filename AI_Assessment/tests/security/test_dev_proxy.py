from __future__ import annotations

import subprocess
import sys
from pathlib import Path

import pytest

SCRIPT = Path(__file__).resolve().parents[2] / "scripts" / "dev_signing_proxy.py"


def run_proxy(*args: str, secret: str | None = "s" * 40) -> subprocess.CompletedProcess[str]:
    env = {"PATH": "/usr/bin:/bin"} | ({"AI_SERVICE_SHARED_SECRET": secret} if secret else {})
    return subprocess.run([sys.executable, str(SCRIPT), *args], capture_output=True, text=True, env=env, timeout=30)


@pytest.mark.parametrize(
    "target",
    ["https://praxis-ai.azurewebsites.net", "http://10.0.0.5:8000", "http://ai:8000", "file://localhost/etc/passwd"],
)
def test_refuses_non_local_target(target: str) -> None:
    result = run_proxy("--target", target)
    assert result.returncode != 0
    assert "refusing to sign" in result.stderr


def test_refuses_to_start_without_secret() -> None:
    result = run_proxy(secret=None)
    assert result.returncode != 0
    assert "AI_SERVICE_SHARED_SECRET" in result.stderr
