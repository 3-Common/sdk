"""Errors module — re-exports the base [APIError][threecommon.APIError] and
every typed subtype.

Customers usually catch via the root re-exports::

    from threecommon import NotFoundError

but importing the submodule directly also works::

    from threecommon.errors import NotFoundError
"""

from threecommon.errors.base import APIError
from threecommon.errors.classes import (
    AuthError,
    ConflictError,
    ConnectionError,
    NotFoundError,
    PermissionError,
    RateLimitError,
    ServerError,
    ValidationError,
)

__all__ = (
    "APIError",
    "AuthError",
    "ConflictError",
    "ConnectionError",
    "NotFoundError",
    "PermissionError",
    "RateLimitError",
    "ServerError",
    "ValidationError",
)
