from __future__ import annotations

from collections.abc import Mapping
from typing import Any, ClassVar


class ContentError(Exception):
    """Versioned content failed an integrity check. Raised at startup, never per request."""


class ServiceError(Exception):
    status: ClassVar[int] = 500
    code: ClassVar[str] = "internal_error"

    def __init__(self, message: str, *, details: Mapping[str, Any] | None = None) -> None:
        super().__init__(message)
        self.message = message
        self.details = dict(details or {})


class Unauthenticated(ServiceError):
    status = 401
    code = "unauthenticated"


class Forbidden(ServiceError):
    status = 403
    code = "forbidden"


class UnsupportedVersion(ServiceError):
    status = 409
    code = "unsupported_version"


class NotComparable(ServiceError):
    status = 409
    code = "not_comparable"


class InvalidInput(ServiceError):
    status = 422
    code = "invalid_input"


class AuthNotConfigured(ServiceError):
    status = 503
    code = "service_auth_not_configured"
