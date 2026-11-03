from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Protocol


class KeySource(Protocol):
    """Where the assessor keys for one scenario bank version come from. Returns the raw JSON document."""

    def read(self, version: str) -> bytes | None: ...


@dataclass(frozen=True, slots=True)
class LocalKeySource:
    """The JSON files under content/data/assessor_keys. Development and tests only."""

    root: Path

    def read(self, version: str) -> bytes | None:
        try:
            return (self.root / "assessor_keys" / f"{version}.json").read_bytes()
        except FileNotFoundError:
            return None
