from __future__ import annotations

import os
from collections.abc import Mapping
from dataclasses import dataclass, field
from pathlib import Path
from typing import Literal

DEFAULT_CONTENT_DIR = Path(__file__).parent / "content" / "data"
MIN_SECRET_BYTES = 32


@dataclass(frozen=True, slots=True)
class Settings:
    env: Literal["local", "production"]
    # repr=False keeps secrets out of tracebacks and any accidental log of the settings object.
    shared_secret: bytes | None = field(repr=False)
    # Accepted alongside shared_secret only while a rotation is in progress. Go always signs with the current one.
    previous_shared_secret: bytes | None = field(default=None, repr=False)
    content_dir: Path = DEFAULT_CONTENT_DIR
    max_body_bytes: int = 256 * 1024
    max_clock_skew_s: int = 300

    def __post_init__(self) -> None:
        for name, secret in (
            ("AI_SERVICE_SHARED_SECRET", self.shared_secret),
            ("AI_SERVICE_SHARED_SECRET_PREVIOUS", self.previous_shared_secret),
        ):
            if secret is not None and len(secret) < MIN_SECRET_BYTES:
                raise ValueError(f"{name} must be at least {MIN_SECRET_BYTES} bytes")
        if self.previous_shared_secret is not None:
            if self.shared_secret is None:
                raise ValueError("AI_SERVICE_SHARED_SECRET_PREVIOUS is set without AI_SERVICE_SHARED_SECRET")
            if self.previous_shared_secret == self.shared_secret:
                raise ValueError(
                    "AI_SERVICE_SHARED_SECRET_PREVIOUS equals the current secret; the rotation did nothing"
                )
        # Locally I let the service boot without a secret so /health works; internal routes still
        # refuse every call. In production a missing secret is a deployment bug, so I refuse to start.
        if self.env == "production" and self.shared_secret is None:
            raise ValueError("AI_SERVICE_SHARED_SECRET is required when PRAXIS_ENV=production")

    @property
    def verification_secrets(self) -> tuple[bytes, ...]:
        return tuple(s for s in (self.shared_secret, self.previous_shared_secret) if s is not None)

    @property
    def docs_enabled(self) -> bool:
        return self.env == "local"

    @classmethod
    def from_env(cls, environ: Mapping[str, str] = os.environ) -> Settings:
        env = environ.get("PRAXIS_ENV", "local")
        if env not in ("local", "production"):
            raise ValueError(f"PRAXIS_ENV must be 'local' or 'production', got {env!r}")
        secret = environ.get("AI_SERVICE_SHARED_SECRET") or None
        previous = environ.get("AI_SERVICE_SHARED_SECRET_PREVIOUS") or None
        return cls(
            env=env,
            shared_secret=secret.encode() if secret else None,
            previous_shared_secret=previous.encode() if previous else None,
            content_dir=Path(environ.get("PRAXIS_CONTENT_DIR", DEFAULT_CONTENT_DIR)),
        )
