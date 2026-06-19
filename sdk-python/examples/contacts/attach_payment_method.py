"""Run with: python examples/contacts/attach_payment_method.py

Persists the card from a SetupIntent that has already been confirmed
client-side with Stripe Elements. Fails with ``ValidationError`` if the
SetupIntent is unconfirmed or out of scope.
"""

from __future__ import annotations

from threecommon import ThreeCommon, ValidationError
from threecommon.contacts import AttachPaymentMethodBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        try:
            result = client.contacts.attach_payment_method(
                "cnt_replace_with_real_id",
                AttachPaymentMethodBody(setup_intent_id="seti_replace_with_real_id"),
            )
            card = result.data.card
            print(f"saved {card.brand} ****{card.last4}")
            print(f"  replaced existing: {result.replaced_existing}")
        except ValidationError as err:
            print(f"could not attach payment method: {err.code}")


if __name__ == "__main__":
    main()
