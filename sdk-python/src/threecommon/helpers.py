"""Small utility helpers exposed at the package root."""

from __future__ import annotations

from typing import TypeVar

T = TypeVar("T")


def not_none(value: T | None, name: str) -> T:
    """Return ``value`` if non-None, else raise ``ValueError`` naming the field.

    Used internally by service methods to surface a clean error when a
    required keyword argument is the literal ``None``.
    """
    if value is None:
        msg = f"{name} must not be None"
        raise ValueError(msg)
    return value
