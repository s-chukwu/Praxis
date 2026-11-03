from __future__ import annotations

import json
from collections.abc import Mapping
from dataclasses import dataclass
from pathlib import Path
from typing import Protocol

from praxis_assessment.errors import ContentError


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


@dataclass(frozen=True, slots=True)
class MappedKeySource:
    documents: Mapping[str, bytes]

    def read(self, version: str) -> bytes | None:
        return self.documents.get(version)


def vault_key_source(vault_url: str, secret_name: str, client_id: str) -> MappedKeySource:
    """Read the assessor-keys secret once, at startup, with the app's user-assigned managed identity."""
    # Imported here so local runs and the test suite never load the Azure SDK.
    from azure.core.exceptions import AzureError
    from azure.identity import ManagedIdentityCredential
    from azure.keyvault.secrets import SecretClient

    try:
        with (
            ManagedIdentityCredential(client_id=client_id) as credential,
            SecretClient(vault_url, credential) as client,
        ):
            value = client.get_secret(secret_name).value
    except AzureError as exc:
        # The type and status are enough to diagnose; I don't forward Azure's message text into our logs.
        status = getattr(exc, "status_code", None)
        raise ContentError(
            f"could not read secret {secret_name} from {vault_url}: {type(exc).__name__}"
            + (f" (HTTP {status})" if status else "")
        ) from None
    if not value:
        raise ContentError(f"secret {secret_name} in {vault_url} is empty")
    return parse_key_map(secret_name, value)


def parse_key_map(secret_name: str, value: str) -> MappedKeySource:
    """The secret maps scenario bank version to that version's assessor_keys document."""
    try:
        data = json.loads(value)
    except ValueError:
        raise ContentError(f"secret {secret_name} is not valid JSON") from None
    if not isinstance(data, dict):
        raise ContentError(f"secret {secret_name} must be a JSON object keyed by scenario bank version")

    documents: dict[str, bytes] = {}
    for version, document in data.items():
        # DevOps may store each version as the object itself or as the file's text; both are unambiguous.
        if isinstance(document, dict):
            documents[version] = json.dumps(document).encode()
        elif isinstance(document, str):
            documents[version] = document.encode()
        else:
            raise ContentError(f"secret {secret_name}: version {version} must be a JSON object or a JSON string")
    return MappedKeySource(documents)
