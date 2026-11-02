"""tests for scripts/dev_signing_proxy.py."""

import importlib.util
import socket
import threading
import urllib.error
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

import pytest

from praxis_assessment.api.auth import ACTOR_HEADER, SIGNATURE_HEADER, TIMESTAMP_HEADER, sign_request

SECRET = b"p" * 40
ACTOR = "admin:11111111-1111-4111-8111-111111111111"

# Load the script
script = Path(__file__).resolve().parents[2] / "scripts" / "dev_signing_proxy.py"
spec = importlib.util.spec_from_file_location("dev_signing_proxy", script)
proxy_script = importlib.util.module_from_spec(spec)
spec.loader.exec_module(proxy_script)


class FakeAIService(BaseHTTPRequestHandler):
    """Pretend AI service: remembers the last request and sends back a set reply."""

    last = None
    reply_status = 200
    reply_body = b'{"ok": true}'

    def do_GET(self):
        self.handle_request()

    def do_POST(self):
        self.handle_request()

    def handle_request(self):
        size = int(self.headers.get("Content-Length") or 0)
        FakeAIService.last = {
            "path": self.path,
            "headers": {k.lower(): v for k, v in self.headers.items()},
            "body": self.rfile.read(size),
        }
        self.send_response(FakeAIService.reply_status)
        self.send_header("Content-Length", str(len(FakeAIService.reply_body)))
        self.send_header("Set-Cookie", "secret=1")  # proxy must not pass this on
        self.end_headers()
        self.wfile.write(FakeAIService.reply_body)

    def log_message(self, *args):
        pass

def start_server(handler):
    server = ThreadingHTTPServer(("127.0.0.1", 0), handler)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    return server, f"http://127.0.0.1:{server.server_port}"


@pytest.fixture(autouse=True)
def reset_fake_service(monkeypatch):
    FakeAIService.reply_status = 200
    FakeAIService.reply_body = b'{"ok": true}'
    monkeypatch.setenv("NO_PROXY", "127.0.0.1")  # ignore system proxy


@pytest.fixture
def proxy_url():
    ai_server, ai_url = start_server(FakeAIService)
    proxy_server, url = start_server(proxy_script.make_handler(ai_url, SECRET, ACTOR))
    yield url
    proxy_server.shutdown()
    proxy_server.server_close()
    ai_server.shutdown()
    ai_server.server_close()


def send(url, path, method="GET", body=None, headers=None):
    request = urllib.request.Request(url + path, data=body, headers=headers or {}, method=method)  # noqa: S310
    try:
        with urllib.request.urlopen(request, timeout=10) as response:  # noqa: S310
            return response.status, dict(response.headers), response.read()
    except urllib.error.HTTPError as error:
        return error.code, dict(error.headers), error.read()


def test_internal_request_is_signed_correctly(proxy_url):
    body = b'{"score": 3}'
    send(proxy_url, "/internal/v1/score", "POST", body)
    seen = FakeAIService.last["headers"]
    expected = sign_request(
        SECRET,
        timestamp=seen[TIMESTAMP_HEADER.lower()],
        method="POST",
        path="/internal/v1/score",
        actor=ACTOR,
        body=body,
    )
    assert seen[SIGNATURE_HEADER.lower()] == expected
    assert seen[ACTOR_HEADER.lower()] == ACTOR


def test_query_string_is_signed(proxy_url):
    path = "/internal/v1/catalog?role=1.0.0"
    send(proxy_url, path)
    seen = FakeAIService.last["headers"]
    expected = sign_request(
        SECRET, timestamp=seen[TIMESTAMP_HEADER.lower()], method="GET", path=path, actor=ACTOR, body=b""
    )
    assert seen[SIGNATURE_HEADER.lower()] == expected


@pytest.mark.parametrize("path", ["/health", "/docs"])
def test_public_paths_are_not_signed(proxy_url, path):
    send(proxy_url, path)
    seen = FakeAIService.last["headers"]
    assert SIGNATURE_HEADER.lower() not in seen
    assert ACTOR_HEADER.lower() not in seen


def test_caller_cannot_fake_actor_or_signature(proxy_url):
    fake = {ACTOR_HEADER: "admin:22222222-2222-4222-8222-222222222222", SIGNATURE_HEADER: "f" * 64}
    send(proxy_url, "/internal/v1/catalog", headers=fake)
    seen = FakeAIService.last["headers"]
    assert seen[ACTOR_HEADER.lower()] == ACTOR
    assert seen[SIGNATURE_HEADER.lower()] != "f" * 64


def test_body_is_forwarded_unchanged(proxy_url):
    body = '{"note": "café"}'.encode()
    send(proxy_url, "/internal/v1/score", "POST", body, {"Content-Type": "application/json"})
    assert FakeAIService.last["body"] == body


def test_response_is_passed_back(proxy_url):
    FakeAIService.reply_body = b'{"overall": 82.5}'
    status, headers, body = send(proxy_url, "/internal/v1/score", "POST", b"{}")
    assert status == 200
    assert body == b'{"overall": 82.5}'
    #drop unsfe heasers
    assert "Set-Cookie" not in headers


@pytest.mark.parametrize("status", [401, 409, 422, 500])
def test_error_status_is_passed_back(proxy_url, status):
    FakeAIService.reply_status = status
    got, _, _ = send(proxy_url, "/internal/v1/score", "POST", b"{}")
    assert got == status


def test_unreachable_service_gives_502():
    with socket.socket() as sock:  # find a port nobody is using
        sock.bind(("127.0.0.1", 0))
        dead_url = f"http://127.0.0.1:{sock.getsockname()[1]}"
    server, url = start_server(proxy_script.make_handler(dead_url, SECRET, ACTOR))
    status, _, body = send(url, "/health")
    server.shutdown()
    server.server_close()
    assert status == 502
    assert b"unreachable" in body
    