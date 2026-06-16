"""Run with: python examples/forms/enable_other_option.py

Adds a free-text "Other" choice to a selection element.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import EnableOtherOptionBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        element = client.forms.enable_other_option(
            "frm_replace_with_real_id",
            "elm_replace_with_real_id",
            EnableOtherOptionBody(other_prompt="Other (please specify)"),
        )
        print(f"enabled 'Other' on element {element.id}: {element.other_prompt}")


if __name__ == "__main__":
    main()
