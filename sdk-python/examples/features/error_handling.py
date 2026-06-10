"""Run with: python examples/features/error_handling.py"""

from __future__ import annotations

from threecommon import (
    AuthError,
    ConflictError,
    NotFoundError,
    RateLimitError,
    ThreeCommon,
    ValidationError,
)
from threecommon.features import CreateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        try:
            # A feature `key` is unique per host — recreating an existing key conflicts.
            client.features.create(CreateBody(key="api_calls", name="API calls", type="quantity"))
        except ConflictError:
            print("a feature with this key already exists")
        except ValidationError as e:
            print(f"validation: {e.message}")
        except NotFoundError:
            print("feature not found")
        except AuthError as e:
            print(f"auth failed: check your API key — code={e.code}")
        except RateLimitError as e:
            wait = e.retry_after_seconds or 30
            print(f"rate limited; waiting {wait}s before retry")


if __name__ == "__main__":
    main()
