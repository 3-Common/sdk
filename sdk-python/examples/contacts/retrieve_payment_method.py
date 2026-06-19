"""Run with: python examples/contacts/retrieve_payment_method.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        method = client.contacts.retrieve_payment_method("cnt_replace_with_real_id")
        if method is None:
            print("no card on file")
            return
        print(f"{method.card.brand} ****{method.card.last4} ({method.status})")
        print(f"  expires:  {method.card.exp_month:02d}/{method.card.exp_year}")
        print(f"  id:       {method.id}")


if __name__ == "__main__":
    main()
