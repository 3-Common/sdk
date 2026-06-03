"""Run with: python examples/contacts/count.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.contacts.count()
        print(f"host has {result.count} contacts")


if __name__ == "__main__":
    main()
