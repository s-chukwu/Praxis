from __future__ import annotations

from typing import Any, Protocol


class SkillModel(Protocol):
    """An LLM that reads a CV and proposes skills. Swapping providers means writing one of these, nothing else.

    Implementations raise ModelUnavailable for any timeout, rate limit or unparseable answer, and never retry
    their way past a deadline: our Go backend has its own timeout and a human fallback.
    """

    name: str

    def suggest(self, instructions: str, cv_text: str) -> list[Any]: ...
