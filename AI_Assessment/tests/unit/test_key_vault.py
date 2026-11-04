from __future__ import annotations

import json
from types import SimpleNamespace
from typing import Any

import azure.identity
import azure.keyvault.secrets
import pytest
from azure.core.exceptions import ClientAuthenticationError, ResourceNotFoundError

from praxis_assessment.config import DEFAULT_CONTENT_DIR
from praxis_assessment.content import load_registry
from praxis_assessment.content.keys import parse_key_map, vault_key_source
from praxis_assessment.errors import ContentError
from support import FIXTURES, HELD_OUT_FIXTURES, fixture_registry

VAULT = "https://praxis-kv-26c0.vault.azure.net/"
CLIENT_ID = "33333333-3333-4333-8333-333333333333"
SHIPPED = {p.stem: json.loads(p.read_text()) for p in (FIXTURES / "assessor_keys").glob("*.json")}


class FakeVault:
    """Replaces the credential and SecretClient, so no test touches the network or needs an Azure login."""

    def __init__(self, monkeypatch: pytest.MonkeyPatch, value: str | None = None, error: Exception | None = None):
        self.value, self.error, self.calls = value, error, {}
        fake = self

        class Credential:
            def __init__(self, *, client_id: str) -> None:
                fake.calls["client_id"] = client_id

            def __enter__(self) -> Credential:
                return self

            def __exit__(self, *_: Any) -> None:
                fake.calls["credential_closed"] = True

        class Client:
            def __init__(self, vault_url: str, credential: Credential) -> None:
                fake.calls["vault_url"] = vault_url

            def __enter__(self) -> Client:
                return self

            def __exit__(self, *_: Any) -> None:
                pass

            def get_secret(self, name: str) -> SimpleNamespace:
                fake.calls["secret"] = name
                if fake.error:
                    raise fake.error
                return SimpleNamespace(value=fake.value)

        monkeypatch.setattr(azure.identity, "ManagedIdentityCredential", Credential)
        monkeypatch.setattr(azure.keyvault.secrets, "SecretClient", Client)


def test_keys_load_from_the_vault_with_the_user_assigned_identity(monkeypatch: pytest.MonkeyPatch) -> None:
    vault = FakeVault(monkeypatch, json.dumps(SHIPPED))
    keys = vault_key_source(VAULT, "assessor-keys", CLIENT_ID)
    registry = load_registry(DEFAULT_CONTENT_DIR, keys, HELD_OUT_FIXTURES)
    assert vault.calls == {
        "client_id": CLIENT_ID, "vault_url": VAULT, "secret": "assessor-keys", "credential_closed": True,
    }  # fmt: skip
    assert registry.assessor_key_set("1.1.0").content == fixture_registry().assessor_key_set(
        "1.1.0"
    ).content


def test_versions_stored_as_file_text_load_too() -> None:
    source = parse_key_map("assessor-keys", json.dumps({v: json.dumps(doc) for v, doc in SHIPPED.items()}))
    registry = load_registry(DEFAULT_CONTENT_DIR, source, HELD_OUT_FIXTURES)
    assert registry.assessor_key_set("1.0.0").content.version == "1.0.0"


def test_a_bank_version_missing_from_the_secret_refuses_to_start() -> None:
    source = parse_key_map("assessor-keys", json.dumps({"1.0.0": SHIPPED["1.0.0"]}))
    with pytest.raises(ContentError, match="no assessor keys for scenario bank 1.1.0"):
        load_registry(DEFAULT_CONTENT_DIR, source, HELD_OUT_FIXTURES)


@pytest.mark.parametrize(
    ("value", "message"),
    [
        ("not json", "not valid JSON"),
        ('["1.0.0"]', "must be a JSON object keyed by scenario bank version"),
        ('{"1.0.0": 7}', "version 1.0.0 must be a JSON object or a JSON string"),
    ],
)
def test_malformed_secret_refuses_to_start(value: str, message: str) -> None:
    with pytest.raises(ContentError, match=message):
        parse_key_map("assessor-keys", value)


@pytest.mark.parametrize(
    ("error", "message"),
    [
        (ResourceNotFoundError("SecretNotFound"), "ResourceNotFoundError"),
        (ClientAuthenticationError("token request failed for client 3333"), "ClientAuthenticationError"),
    ],
)
def test_vault_failures_refuse_to_start_without_echoing_azure(
    monkeypatch: pytest.MonkeyPatch, error: Exception, message: str
) -> None:
    FakeVault(monkeypatch, error=error)
    with pytest.raises(ContentError, match=f"could not read secret assessor-keys from {VAULT}: {message}") as exc:
        vault_key_source(VAULT, "assessor-keys", CLIENT_ID)
    assert str(error) not in str(exc.value)


def test_empty_secret_refuses_to_start(monkeypatch: pytest.MonkeyPatch) -> None:
    FakeVault(monkeypatch, "")
    with pytest.raises(ContentError, match="is empty"):
        vault_key_source(VAULT, "assessor-keys", CLIENT_ID)
