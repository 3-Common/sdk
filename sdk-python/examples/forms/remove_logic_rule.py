"""Run with: python examples/forms/remove_logic_rule.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        element = client.forms.remove_logic_rule(
            "frm_replace_with_real_id",
            "elm_source_id",
            "elm_revealed_id",
        )
        print(f"removed logic rule from element {element.id} ({element.type})")


if __name__ == "__main__":
    main()
