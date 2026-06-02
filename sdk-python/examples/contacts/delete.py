"""Run with: python examples/contacts/delete.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.contacts.delete("cnt_replace_with_real_id")
        print(f"deleted {result.id}")


if __name__ == "__main__":
    main()
