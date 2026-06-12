"""Run with: python examples/forms/move_element.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import MoveElementBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        form = client.forms.move_element(
            "frm_replace_with_real_id",
            "elm_replace_with_real_id",
            MoveElementBody(position=2),
        )
        print(f"moved element on form {form.id} ({form.status})")


if __name__ == "__main__":
    main()
