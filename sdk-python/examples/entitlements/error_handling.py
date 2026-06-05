"""Run with: python examples/entitlements/error_handling.py"""

from __future__ import annotations

from threecommon import (
    AuthError,
    ConflictError,
    NotFoundError,
    RateLimitError,
    ThreeCommon,
)
from threecommon.entitlements import ConsumeBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        try:
            client.entitlements.consume(
                ConsumeBody(
                    contact_id="cnt_replace_with_real_id",
                    feature_key="api_calls",
                    amount=1_000_000,
                )
            )
        except ConflictError:
            print("insufficient balance — top up before consuming")
        except NotFoundError as e:
            print(f"no entitlement record for this contact + feature — request_id={e.request_id}")
        except AuthError as e:
            print(f"auth failed: check your API key — code={e.code}")
        except RateLimitError as e:
            wait = e.retry_after_seconds or 30
            print(f"rate limited; waiting {wait}s before retry")


if __name__ == "__main__":
    main()
