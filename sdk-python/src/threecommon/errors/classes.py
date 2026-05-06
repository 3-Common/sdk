"""HTTP-status-specific exception subtypes.

All inherit from [APIError][threecommon.APIError]. Catch the subtype user care
about; the order of `except` clauses can go from specific to general.
"""

from __future__ import annotations

from typing import Any

from threecommon.errors.base import APIError


class AuthError(APIError):
    """401 Unauthorized — invalid, missing, or expired API key."""


class PermissionError(APIError):
    """403 Forbidden — the API key lacks the scope required by the endpoint."""


class NotFoundError(APIError):
    """404 Not Found."""


class ValidationError(APIError):
    """400 Bad Request and 422 Unprocessable Entity — request validation failed."""


class ConflictError(APIError):
    """409 Conflict — the request conflicts with current resource state."""


class RateLimitError(APIError):
    """429 Too Many Requests.

    [retry_after_seconds][threecommon.RateLimitError.retry_after_seconds]
    carries the parsed ``Retry-After`` header so user can implement their
    own backoff; it is ``None`` when the server did not provide one.
    """

    retry_after_seconds: float | None

    def __init__(
        self,
        *,
        code: str,
        message: str,
        http_status: int | None = None,
        request_id: str | None = None,
        details: dict[str, Any] | None = None,
        raw_response: str | None = None,
        cause: BaseException | None = None,
        retry_after_seconds: float | None = None,
    ) -> None:
        super().__init__(
            code=code,
            message=message,
            http_status=http_status,
            request_id=request_id,
            details=details,
            raw_response=raw_response,
            cause=cause,
        )
        self.retry_after_seconds = retry_after_seconds


class ServerError(APIError):
    """5xx — the API returned an unexpected server-side failure."""


class ConnectionError(APIError):
    """The request never completed: DNS failure, TCP reset, TLS error,
    timeout, etc. The original cause is available via ``__cause__``.
    """
