from __future__ import annotations

import argparse
import contextlib
import os
import sys
import time
import urllib.error
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import urlsplit

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "service"))

from praxis_assessment.api.auth import (
    ACTOR_HEADER,
    SIGNATURE_HEADER,
    TIMESTAMP_HEADER,
    sign_request,
)
from praxis_assessment.api.examples import DEMO_REVIEWER_ID

LOCAL_HOSTS = {"localhost", "127.0.0.1", "::1"}
FORWARDED_RESPONSE_HEADERS = ("Content-Type", "X-Request-ID")


def make_handler(target: str, secret: bytes, actor: str) -> type[BaseHTTPRequestHandler]:
    class SigningProxy(BaseHTTPRequestHandler):
        def do_GET(self) -> None:
            self._forward()

        def do_POST(self) -> None:
            self._forward()

        def _forward(self) -> None:
            body = self.rfile.read(int(self.headers.get("Content-Length") or 0))
            headers = {"Content-Type": self.headers.get("Content-Type", "application/json")}
            # /health and /docs stay unsigned, exactly as our Go backend would call them.
            if self.path.startswith("/internal/"):
                timestamp = str(int(time.time()))
                headers |= {
                    TIMESTAMP_HEADER: timestamp,
                    ACTOR_HEADER: actor,
                    SIGNATURE_HEADER: sign_request(
                        secret, timestamp=timestamp, method=self.command, path=self.path, actor=actor, body=body
                    ),
                }

            # S310: main() only accepts an http(s) URL on a loopback host, so no file: or remote targets.
            request = urllib.request.Request(  # noqa: S310
                target + self.path, data=body or None, headers=headers, method=self.command
            )
            try:
                with urllib.request.urlopen(request, timeout=10) as upstream:  # noqa: S310
                    self._relay(upstream.status, upstream.headers, upstream.read())
            except urllib.error.HTTPError as error:
                self._relay(error.code, error.headers, error.read())
            except urllib.error.URLError as error:
                self._relay(
                    502, {"Content-Type": "text/plain"}, f"AI service unreachable at {target}: {error.reason}".encode()
                )

        def _relay(self, status: int, headers, payload: bytes) -> None:
            self.send_response(status)
            for name in FORWARDED_RESPONSE_HEADERS:
                if value := headers.get(name):
                    self.send_header(name, value)
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)

        def log_message(self, format: str, *args: object) -> None:
            print(f"{self.command} {self.path} -> {args[1] if len(args) > 1 else '?'}", flush=True)

    return SigningProxy


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Sign requests locally so the internal endpoints can be tried from /docs. Local use only."
    )
    parser.add_argument("--target", default="http://localhost:8000", help="AI service URL (local only)")
    parser.add_argument("--port", type=int, default=8001)
    parser.add_argument(
        "--actor",
        default=f"reviewer:{DEMO_REVIEWER_ID}",
        help="signed actor: reviewer:<uuid>, candidate:<uuid>, admin:<uuid> or system (default: the demo reviewer)",
    )
    args = parser.parse_args()

    target = urlsplit(args.target)
    if target.scheme not in ("http", "https") or target.hostname not in LOCAL_HOSTS:
        sys.exit(f"refusing to sign requests for non-local target {args.target}")
    secret = os.environ.get("AI_SERVICE_SHARED_SECRET")
    if not secret:
        sys.exit("set AI_SERVICE_SHARED_SECRET to the value the ai container was started with")

    server = ThreadingHTTPServer(
        ("127.0.0.1", args.port), make_handler(args.target.rstrip("/"), secret.encode(), args.actor)
    )
    print(f"Signing as {args.actor!r} -> {args.target}")
    print(f"Open http://127.0.0.1:{args.port}/docs  (Ctrl+C to stop)")
    with contextlib.suppress(KeyboardInterrupt):
        server.serve_forever()


if __name__ == "__main__":
    main()
