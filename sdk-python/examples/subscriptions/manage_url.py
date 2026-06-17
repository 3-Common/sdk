"""Run with: python examples/subscriptions/manage_url.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        portal = client.subscriptions.retrieve_manage_url("sub_replace_with_real_id")
        print(f"manage URL: {portal.url}")


if __name__ == "__main__":
    main()
