from __future__ import annotations

import hashlib
import hmac
import time
from collections.abc import Awaitable, Callable
from dataclasses import dataclass
from typing import Literal, get_args
from uuid import UUID

from fastapi import Request

from praxis_assessment.config import Settings
from praxis_assessment.errors import AuthNotConfigured, Forbidden, Unauthenticated

ActorRole = Literal["reviewer", "candidate", "admin", "system"]
ALL_ROLES: tuple[ActorRole, ...] = get_args(ActorRole)

TIMESTAMP_HEADER = "X-Praxis-Timestamp"
ACTOR_HEADER = "X-Praxis-Actor"
SIGNATURE_HEADER = "X-Praxis-Signature"
SIGNATURE_VERSION = "v1"


@dataclass(frozen=True, slots=True)
class Caller:
    role: ActorRole
    actor_id: UUID | None


# Go must build the exact same canonical string; the format is spelled out in docs/assessment_api.md.
def sign_request(secret: bytes, *, timestamp: str, method: str, path: str, actor: str, body: bytes) -> str:
    canonical = "\n".join((SIGNATURE_VERSION, timestamp, method.upper(), path, actor, hashlib.sha256(body).hexdigest()))
    return hmac.new(secret, canonical.encode(), hashlib.sha256).hexdigest()


def require_caller(*roles: ActorRole) -> Callable[[Request], Awaitable[Caller]]:
    allowed = frozenset(roles)

    async def dependency(request: Request) -> Caller:
        caller = await _authenticate(request)
        if caller.role not in allowed:
            raise Forbidden(
                f"actor role {caller.role} may not call this endpoint", details={"allowed": sorted(allowed)}
            )
        return caller

    return dependency


async def _authenticate(request: Request) -> Caller:
    settings: Settings = request.app.state.settings
    if settings.shared_secret is None:
        raise AuthNotConfigured("AI_SERVICE_SHARED_SECRET is not set; internal endpoints are disabled")

    timestamp = request.headers.get(TIMESTAMP_HEADER)
    actor = request.headers.get(ACTOR_HEADER)
    signature = request.headers.get(SIGNATURE_HEADER)
    if not (timestamp and actor and signature):
        raise Unauthenticated("missing signature headers")
    if not timestamp.isdigit() or abs(time.time() - int(timestamp)) > settings.max_clock_skew_s:
        raise Unauthenticated("timestamp missing or outside the allowed window")

    path = request.url.path + (f"?{request.url.query}" if request.url.query else "")
    body = await request.body()
    presented = signature.encode("latin-1")
    # Two secrets only exist mid-rotation. I check both rather than stopping at the first match;
    # which one matched is not worth leaking through timing.
    matches = [
        hmac.compare_digest(
            sign_request(
                secret, timestamp=timestamp, method=request.method, path=path, actor=actor, body=body
            ).encode(),
            presented,
        )
        for secret in settings.verification_secrets
    ]
    if not any(matches):
        raise Unauthenticated("signature mismatch")
    return _parse_actor(actor)


def _parse_actor(raw: str) -> Caller:
    if raw == "system":
        return Caller(role="system", actor_id=None)
    role, _, actor_id = raw.partition(":")
    if role not in ALL_ROLES or role == "system":
        raise Unauthenticated("malformed actor header")
    try:
        return Caller(role=role, actor_id=UUID(actor_id))
    except ValueError:
        raise Unauthenticated("malformed actor header") from None
