"""Base exception type carried by every error the SDK raises.

The HTTP-status-specific subtypes ([AuthError][threecommon.AuthError],
[NotFoundError][threecommon.NotFoundError], ...) all inherit from
[APIError][threecommon.APIError]. Branch on the subtype with a normal
`except` clause:

    try:
        client.events.retrieve("evt_missing")
    except threecommon.NotFoundError as e:
        log.warning("missing event %s", e.request_id)
"""

from __future__ import annotations

from typing import Any


class APIError(Exception):
    """Base class for every error raised by the SDK.

    Every field is best-effort: ``http_status`` is ``None`` for connection
    errors, ``request_id`` is ``None`` when the server didn't return one,
    ``raw_response`` is empty for non-text responses.
    """

    code: str
    """Stable string matching the API's error.code field, e.g. ``not_found``."""

    message: str
    """Human-readable description. Safe to surface to end users."""

    http_status: int | None
    """Response status, or ``None`` if the error originated before any response."""

    request_id: str | None
    """Value of the ``X-Request-ID`` response header, when present."""

    details: dict[str, Any] | None
    """Parsed API ``error.details`` payload, when present."""

    raw_response: str | None
    """Raw response body, retained for debugging."""

    __cause__: BaseException | None

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
    ) -> None:
        super().__init__(self._format(code, message, request_id))
        self.code = code
        self.message = message
        self.http_status = http_status
        self.request_id = request_id
        self.details = details
        self.raw_response = raw_response
        if cause is not None:
            self.__cause__ = cause

    @staticmethod
    def _format(code: str, message: str, request_id: str | None) -> str:
        if request_id:
            return f"[{code}] {message} (request_id={request_id})"
        return f"[{code}] {message}"

    def __repr__(self) -> str:
        return (
            f"{self.__class__.__name__}("
            f"code={self.code!r}, "
            f"message={self.message!r}, "
            f"http_status={self.http_status!r}, "
            f"request_id={self.request_id!r})"
        )
