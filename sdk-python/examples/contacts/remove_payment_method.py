"""Run with: python examples/contacts/remove_payment_method.py"""

from __future__ import annotations

from threecommon import NotFoundError, ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        try:
            result = client.contacts.remove_payment_method(
                "cnt_replace_with_real_id",
                "pm_replace_with_real_id",
            )
            print(f"removed: {result.removed}")
        except NotFoundError:
            print("payment method not found")


if __name__ == "__main__":
    main()
