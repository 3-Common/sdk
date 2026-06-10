"""Run with: python examples/prices/error_handling.py"""

from __future__ import annotations

from threecommon import (
    AuthError,
    NotFoundError,
    RateLimitError,
    ThreeCommon,
    ValidationError,
)
from threecommon.prices import CreateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        try:
            # `recurring` is required when type is `recurring`; omitting it
            # triggers a 400 validation error.
            client.prices.create(
                CreateBody(
                    product_id="prod_replace_with_real_id",
                    type="recurring",
                    currency="USD",
                    unit_amount=1500,
                )
            )
        except ValidationError as e:
            print(f"validation: {e.message}")
        except NotFoundError:
            print("product not found")
        except AuthError as e:
            print(f"auth failed: check your API key — code={e.code}")
        except RateLimitError as e:
            wait = e.retry_after_seconds or 30
            print(f"rate limited; waiting {wait}s before retry")


if __name__ == "__main__":
    main()
