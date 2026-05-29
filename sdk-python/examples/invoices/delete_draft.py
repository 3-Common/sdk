"""Permanently delete a draft invoice.

Only drafts can be deleted — once an invoice is finalized (it has a number),
void it instead so the audit trail stays intact.

Run with: python examples/invoices/delete_draft.py
"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.invoices.delete_draft("inv_replace_with_real_id")
        print(f"deleted draft invoice {result.id}")


if __name__ == "__main__":
    main()
