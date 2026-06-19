"""Run with: python examples/contacts/create_payment_method_setup_intent.py

Begins saving a card for a contact. Confirm the returned ``client_secret``
client-side with Stripe Elements, then call ``attach_payment_method`` with the
returned ``setup_intent_id`` to persist the card.
"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        intent = client.contacts.create_payment_method_setup_intent("cnt_replace_with_real_id")
        print(f"setup intent:  {intent.setup_intent_id}")
        print(f"customer:      {intent.customer_id}")
        print(f"client secret: {intent.client_secret}")


if __name__ == "__main__":
    main()
