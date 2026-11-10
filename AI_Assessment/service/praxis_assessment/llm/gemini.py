from __future__ import annotations

import json
import time
from collections.abc import Callable
from typing import Any

import httpx

from praxis_assessment.errors import ModelUnavailable

API = "https://generativelanguage.googleapis.com/v1beta"
RETRY_DELAY_S = 0.5

# Gemini enforces this shape on its own output, so a well-behaved answer is already a list of these.
# I still re-validate every item: the schema is a request to the model, not a guarantee.
_SCHEMA = {
    "type": "ARRAY",
    "items": {
        "type": "OBJECT",
        "properties": {
            "competency": {"type": "STRING"},
            "quote": {"type": "STRING"},
            "confidence": {"type": "STRING", "enum": ["low", "medium", "high"]},
        },
        "required": ["competency", "quote", "confidence"],
    },
}


class GeminiSkillModel:
    def __init__(
        self,
        api_key: str,
        model: str,
        *,
        timeout_s: float = 12.0,
        transport: httpx.BaseTransport | None = None,
        sleep: Callable[[float], None] = time.sleep,
    ) -> None:
        self.name = f"gemini/{model}"
        self._url = f"{API}/models/{model}:generateContent"
        # The key goes in a header, never the URL, so it can't end up in an access log.
        self._client = httpx.Client(timeout=timeout_s, transport=transport, headers={"x-goog-api-key": api_key})
        self._sleep = sleep

    def close(self) -> None:
        self._client.close()

    def suggest(self, instructions: str, cv_text: str) -> list[Any]:
        body = {
            "system_instruction": {"parts": [{"text": instructions}]},
            "contents": [{"role": "user", "parts": [{"text": cv_text}]}],
            "generationConfig": {"responseMimeType": "application/json", "responseSchema": _SCHEMA, "temperature": 0},
        }
        # One retry, only for failures that a second try can fix. Two attempts at 12 s stay inside the
        # 30 s our Go backend allows; anything slower goes to a person instead.
        reason = "unknown"
        for attempt in range(2):
            try:
                response = self._client.post(self._url, json=body)
            except httpx.TimeoutException:
                reason = "timeout"
            except httpx.TransportError:
                reason = "network"
            else:
                if response.status_code == 200:
                    return _items(response)
                reason = f"http_{response.status_code}"
                if response.status_code != 429 and response.status_code < 500:
                    break  # a bad key or bad request won't fix itself
            if attempt == 0:
                self._sleep(RETRY_DELAY_S)
        raise _unavailable(reason)


def _items(response: httpx.Response) -> list[Any]:
    # Any surprise here is an unusable answer. I never echo the model's text: it may quote the CV.
    try:
        text = response.json()["candidates"][0]["content"]["parts"][0]["text"]
        items = json.loads(text)
    except (ValueError, KeyError, IndexError, TypeError):
        raise _unavailable("unusable_answer") from None
    if not isinstance(items, list):
        raise _unavailable("unusable_answer")
    return items


def _unavailable(reason: str) -> ModelUnavailable:
    return ModelUnavailable("skill suggestions are unavailable; review the CV manually", details={"reason": reason})
