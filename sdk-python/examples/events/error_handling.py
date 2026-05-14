"""Run with: python examples/events/error_handling.py"""

from __future__ import annotations

from threecommon import (
    AuthError,
    ConnectionError,
    NotFoundError,
    RateLimitError,
    ThreeCommon,
)


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        try:
            client.events.retrieve("000000000000000000000000")
        except NotFoundError as e:
            print(f"event not found — request_id={e.request_id}")
        except AuthError as e:
            print(f"auth failed: check your API key — code={e.code}")
        except RateLimitError as e:
            wait = e.retry_after_seconds or 30
            print(f"rate limited; waiting {wait}s before retry")
        except ConnectionError as e:
            print(f"network error: {e.__cause__}")


if __name__ == "__main__":
    main()
