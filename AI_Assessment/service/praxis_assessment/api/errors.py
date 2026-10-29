from __future__ import annotations

import logging
from typing import Any

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from starlette.exceptions import HTTPException as StarletteHTTPException

from praxis_assessment.errors import ServiceError
from praxis_assessment.schemas.common import ErrorBody, ErrorResponse

log = logging.getLogger("praxis_assessment.errors")

ERROR_RESPONSES: dict[int | str, dict[str, Any]] = {
    status: {"model": ErrorResponse} for status in (401, 403, 409, 413, 422, 503)
}

_HTTP_CODES = {404: "not_found", 405: "method_not_allowed"}


def request_id_of(request: Request) -> str:
    return getattr(request.state, "request_id", "unknown")


def error_json(
    request_id: str, status: int, code: str, message: str, details: dict[str, Any] | None = None
) -> JSONResponse:
    body = ErrorResponse(error=ErrorBody(code=code, message=message, request_id=request_id, details=details or {}))
    return JSONResponse(body.model_dump(mode="json"), status_code=status)


def install_error_handlers(app: FastAPI) -> None:
    @app.exception_handler(ServiceError)
    async def _service_error(request: Request, exc: ServiceError) -> JSONResponse:
        return error_json(request_id_of(request), exc.status, exc.code, exc.message, exc.details)

    @app.exception_handler(RequestValidationError)
    async def _validation_error(request: Request, exc: RequestValidationError) -> JSONResponse:
        # FastAPI's default body echoes the offending input back. Inputs here are candidate answers and
        # reviewer notes, so I return only where and why validation failed, never the value.
        errors = [
            {
                "loc": [str(part)[:64] for part in err.get("loc", ())],
                "msg": err.get("msg", ""),
                "type": err.get("type", ""),
            }
            for err in exc.errors()
        ]
        return error_json(
            request_id_of(request), 422, "invalid_request", "request failed schema validation", {"errors": errors}
        )

    @app.exception_handler(StarletteHTTPException)
    async def _http_error(request: Request, exc: StarletteHTTPException) -> JSONResponse:
        code = _HTTP_CODES.get(exc.status_code, "http_error")
        return error_json(request_id_of(request), exc.status_code, code, str(exc.detail))

    @app.exception_handler(Exception)
    async def _unhandled(request: Request, exc: Exception) -> JSONResponse:
        request_id = request_id_of(request)
        log.exception("unhandled error", extra={"request_id": request_id})
        return error_json(request_id, 500, "internal_error", "internal error")
