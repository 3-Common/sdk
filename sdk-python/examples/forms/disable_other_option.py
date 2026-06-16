"""Run with: python examples/forms/disable_other_option.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        element = client.forms.disable_other_option(
            "frm_replace_with_real_id", "elm_replace_with_real_id"
        )
        print(f"disabled 'Other' on element {element.id} ({element.type})")


if __name__ == "__main__":
    main()
