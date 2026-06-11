"""Run with: python examples/forms/duplicate.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import DuplicateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        copy = client.forms.duplicate(
            "frm_replace_with_real_id",
            DuplicateBody(name="Customer survey (copy)"),
        )
        print(f"duplicated into {copy.id} ({copy.name}) status={copy.status}")


if __name__ == "__main__":
    main()
