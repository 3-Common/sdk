"""Run with: python examples/forms/update_element.py

Only the fields you set on ``UpdateElementBody`` are changed.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import UpdateElementBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        element = client.forms.update_element(
            "frm_replace_with_real_id",
            "elm_replace_with_real_id",
            UpdateElementBody(prompt="What is your full name?", required=False),
        )
        print(f"updated element {element.id} -> {element.prompt} (required={element.required})")


if __name__ == "__main__":
    main()
