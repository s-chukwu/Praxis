from __future__ import annotations

from collections import Counter
from collections.abc import Iterable
from typing import Annotated, Any

from pydantic import BaseModel, ConfigDict, Field, StrictInt, StringConstraints


class Frozen(BaseModel):
    """Immutable and closed.

    I forbid unknown fields on every model: a silently dropped field is how a caller-supplied
    `weighted_score` or a misspelled version key gets past a boundary unnoticed.
    """

    model_config = ConfigDict(frozen=True, extra="forbid")


SemVer = Annotated[str, StringConstraints(pattern=r"^\d+\.\d+\.\d+$")]
CompetencyCode = Annotated[str, StringConstraints(pattern=r"^[CB][1-9]$")]
ScenarioId = Annotated[str, StringConstraints(pattern=r"^CS-\d{2}-[A-Z]$")]
Slug = Annotated[str, StringConstraints(pattern=r"^[a-z][a-z_]{0,63}$")]
# A reviewer rating on the 0-4 anchor scale. Strict, so "3", 2.5 and True are rejected rather than coerced.
Rating = Annotated[StrictInt, Field(ge=0, le=4)]


def require_unique(ids: Iterable[str], what: str) -> None:
    dupes = sorted(k for k, n in Counter(ids).items() if n > 1)
    if dupes:
        raise ValueError(f"duplicate {what}: {', '.join(dupes)}")


class ContentRef(Frozen):
    version: str
    sha256: str


class ErrorBody(Frozen):
    code: str
    message: str
    request_id: str
    details: dict[str, Any] = {}


class ErrorResponse(Frozen):
    error: ErrorBody
