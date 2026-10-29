from __future__ import annotations

import json
import logging
import re
import time
import uuid

from starlette.datastructures import MutableHeaders
from starlette.types import ASGIApp, Message, Receive, Scope, Send

from praxis_assessment.api.errors import error_json

log = logging.getLogger("praxis_assessment.access")

REQUEST_ID_HEADER = "X-Request-ID"
_VALID_REQUEST_ID = re.compile(r"^[A-Za-z0-9-]{8,64}$")


# Raw ASGI rather than BaseHTTPMiddleware, so the body cap applies before FastAPI reads anything.
class RequestContextMiddleware:
    def __init__(self, app: ASGIApp, *, max_body_bytes: int) -> None:
        self.app = app
        self.max_body_bytes = max_body_bytes

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        request_id = _request_id(scope)
        scope.setdefault("state", {})["request_id"] = request_id
        started = time.perf_counter()
        status = 500

        async def send_with_id(message: Message) -> None:
            nonlocal status
            if message["type"] == "http.response.start":
                status = message["status"]
                headers = MutableHeaders(scope=message)
                headers.append(REQUEST_ID_HEADER, request_id)
                # Every response here carries candidate evidence or assessor keys; no proxy or browser may keep it.
                headers["Cache-Control"] = "no-store"
            await send(message)

        try:
            body = await self._read_bounded(receive)
            if body is None:
                await error_json(request_id, 413, "payload_too_large", f"body exceeds {self.max_body_bytes} bytes")(
                    scope, receive, send_with_id
                )
                return
            await self.app(scope, _replay(body, receive), send_with_id)
        finally:
            log.info(
                json.dumps(
                    {
                        "event": "request",
                        "request_id": request_id,
                        "method": scope["method"],
                        "path": scope["path"],
                        "status": status,
                        "duration_ms": round((time.perf_counter() - started) * 1000, 2),
                    }
                )
            )

    async def _read_bounded(self, receive: Receive) -> bytes | None:
        # Returns None when the body is over the limit. Content-Length can lie or be absent (chunked),
        # so I count what actually arrives.
        chunks: list[bytes] = []
        size = 0
        while True:
            message = await receive()
            if message["type"] != "http.request":
                break
            chunk = message.get("body", b"")
            size += len(chunk)
            if size > self.max_body_bytes:
                return None
            chunks.append(chunk)
            if not message.get("more_body", False):
                break
        return b"".join(chunks)


def _replay(body: bytes, receive: Receive) -> Receive:
    delivered = False

    async def replay() -> Message:
        nonlocal delivered
        if not delivered:
            delivered = True
            return {"type": "http.request", "body": body, "more_body": False}
        return await receive()

    return replay


def _request_id(scope: Scope) -> str:
    for name, value in scope.get("headers", ()):
        if name == b"x-request-id":
            candidate = value.decode("latin-1")
            if _VALID_REQUEST_ID.match(candidate):
                return candidate
    return str(uuid.uuid4())
