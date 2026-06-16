"""Run with: python examples/forms/error_handling.py"""

from __future__ import annotations

from threecommon import (
    AuthError,
    NotFoundError,
    RateLimitError,
    ThreeCommon,
    ValidationError,
)
from threecommon.forms import CreateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        try:
            client.forms.create(CreateBody(name="Registration", type="standalone"))
        except ValidationError as e:
            print(f"validation: {e.message}")
        except NotFoundError:
            print("form or element not found")
        except AuthError as e:
            print(f"auth failed: check your API key (code={e.code})")
        except RateLimitError as e:
            wait = e.retry_after_seconds or 30
            print(f"rate limited; waiting {wait}s before retry")


if __name__ == "__main__":
    main()
