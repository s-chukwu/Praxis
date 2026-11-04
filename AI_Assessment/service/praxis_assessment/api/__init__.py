from __future__ import annotations

import json
import logging

from fastapi import FastAPI

from praxis_assessment import SERVICE_VERSION
from praxis_assessment.api.errors import install_error_handlers
from praxis_assessment.api.middleware import RequestContextMiddleware
from praxis_assessment.api.routes import health_router, internal_router
from praxis_assessment.config import Settings
from praxis_assessment.content import load_registry
from praxis_assessment.content.keys import vault_key_source

log = logging.getLogger("praxis_assessment.content")


def create_app(settings: Settings | None = None) -> FastAPI:
    settings = settings or Settings.from_env()
    docs = settings.docs_enabled
    app = FastAPI(
        title="Praxis AI / Assessment Service",
        version=SERVICE_VERSION,
        description="Internal service called only by our Go backend. Never exposed to browsers.",
        docs_url="/docs" if docs else None,
        redoc_url=None,
        openapi_url="/openapi.json" if docs else None,
    )
    # Content loads and cross-checks here, so a broken content file stops the deploy instead of a request.
    app.state.settings = settings
    vault = settings.key_vault
    keys = vault_key_source(vault.url, vault.secret_name, vault.client_id) if vault else None
    app.state.registry = load_registry(settings.content_dir, keys)
    source = {"source": "key_vault", "secret": vault.secret_name} if vault else {"source": "local_files"}
    versions = sorted(app.state.registry.assessor_keys)
    log.info(json.dumps({"event": "assessor_keys_loaded", **source, "versions": versions}))

    install_error_handlers(app)
    app.include_router(health_router)
    app.include_router(internal_router)
    app.add_middleware(RequestContextMiddleware, max_body_bytes=settings.max_body_bytes)
    return app
