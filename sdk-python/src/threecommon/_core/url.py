"""URL builder. Pure function — no I/O."""

from __future__ import annotations

from urllib.parse import urlencode


def build_url(
    *,
    base_url: str,
    api_path: str,
    path: str,
    query: dict[str, str] | None = None,
) -> str:
    """Concatenate ``base_url + api_path + path`` and append a sorted query string.

    Trailing slashes on ``base_url`` are trimmed; missing leading slashes on
    ``path`` are added. Query keys are stable-sorted for deterministic output.
    """
    base = base_url.rstrip("/")
    p = path if path.startswith("/") else "/" + path
    out = base + api_path + p

    if not query:
        return out

    pairs = [(k, v) for k, v in sorted(query.items()) if v]
    if not pairs:
        return out

    return f"{out}?{urlencode(pairs)}"
