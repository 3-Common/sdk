"""Run with: python examples/forms/add_element.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import AddElementBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        element = client.forms.add_element(
            "frm_replace_with_real_id",
            AddElementBody(type="Text", prompt="What is your name?", required=True),
        )
        print(f"added element {element.id} ({element.type}): {element.prompt}")


if __name__ == "__main__":
    main()
