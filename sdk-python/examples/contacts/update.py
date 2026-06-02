"""Run with: python examples/contacts/update.py

Returns the richer order-details projection (``ContactWithOrderDetails``),
not the compact ``Contact`` returned by ``retrieve``.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.contacts import ContactUpdate, UpdateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        updated = client.contacts.update(
            "cnt_replace_with_real_id",
            UpdateBody(
                contact=ContactUpdate(
                    first_name="Alex",
                    last_name="Garcia",
                    email="a.garcia@example.com",
                    status="opted-in",
                )
            ),
        )
        print(f"updated {updated.id_} → {updated.email} ({updated.status})")


if __name__ == "__main__":
    main()
