from __future__ import annotations

import json

import httpx
import pytest

from praxis_assessment.errors import ModelUnavailable
from praxis_assessment.llm.gemini import GeminiSkillModel

CV = "Handled 60 billing tickets a day. My password is hunter2"
ITEMS = [{"competency": "C3", "quote": "Handled 60 billing tickets", "confidence": "high"}]


def reply(payload: object, status: int = 200) -> httpx.Response:
    return httpx.Response(status, json={"candidates": [{"content": {"parts": [{"text": json.dumps(payload)}]}}]}
                          if status == 200 else {"error": {"message": f"upstream said {CV}"}})  # fmt: skip


def model(*responses: httpx.Response | Exception, seen: list[httpx.Request] | None = None) -> GeminiSkillModel:
    queue = list(responses)

    def handle(request: httpx.Request) -> httpx.Response:
        if seen is not None:
            seen.append(request)
        nxt = queue.pop(0)
        if isinstance(nxt, Exception):
            raise nxt
        return nxt

    return GeminiSkillModel("test-key", "gemini-test", transport=httpx.MockTransport(handle), sleep=lambda _: None)


def test_sends_the_cv_as_data_with_a_strict_schema() -> None:
    seen: list[httpx.Request] = []
    assert model(reply(ITEMS), seen=seen).suggest("RULES", CV) == ITEMS
    request = seen[0]
    body = json.loads(request.content)
    assert body["system_instruction"]["parts"][0]["text"] == "RULES"
    assert body["contents"][0]["parts"][0]["text"] == CV
    assert body["generationConfig"]["responseSchema"]["type"] == "ARRAY"
    assert body["generationConfig"]["temperature"] == 0
    assert request.headers["x-goog-api-key"] == "test-key" and "test-key" not in str(request.url)


@pytest.mark.parametrize("first", [httpx.Response(429), httpx.Response(503), httpx.ReadTimeout("slow")])
def test_retries_once_when_a_retry_can_help(first: httpx.Response | Exception) -> None:
    assert model(first, reply(ITEMS)).suggest("RULES", CV) == ITEMS


@pytest.mark.parametrize(
    ("responses", "reason"),
    [
        ((httpx.Response(429), httpx.Response(429)), "http_429"),  # still rate limited after the retry
        ((httpx.ReadTimeout("slow"), httpx.ReadTimeout("slow")), "timeout"),  # A10
        ((httpx.ConnectError("down"), httpx.ConnectError("down")), "network"),
        ((reply(None, 403),), "http_403"),  # a bad key is not retried
        ((reply({"not": "a list"}),), "unusable_answer"),
        ((httpx.Response(200, json={"promptFeedback": {"blockReason": "SAFETY"}}),), "unusable_answer"),
        ((httpx.Response(200, text="not json"),), "unusable_answer"),
    ],
)
def test_every_failure_is_model_unavailable(responses: tuple, reason: str) -> None:
    with pytest.raises(ModelUnavailable) as exc:
        model(*responses).suggest("RULES", CV)
    assert exc.value.details == {"reason": reason}
    assert "hunter2" not in exc.value.message  # never echo the CV or the provider's text


def test_a_single_retry_means_at_most_two_calls() -> None:
    seen: list[httpx.Request] = []
    with pytest.raises(ModelUnavailable):
        model(httpx.Response(500), httpx.Response(500), httpx.Response(200), seen=seen).suggest("RULES", CV)
    assert len(seen) == 2


