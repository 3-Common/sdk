"""Run with: python examples/contacts/retrieve.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        contact = client.contacts.retrieve("cnt_replace_with_real_id")
        print(f"{contact.full_name} <{contact.email}>")
        print(f"  status:     {contact.status}")
        print(f"  orders:     {contact.order_sum}")
        print(f"  gross:      {contact.gross_sum}")
        print(f"  vendor_id:  {contact.vendor_id}")


if __name__ == "__main__":
    main()
