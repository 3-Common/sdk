"""Run with: python examples/forms/remove_logic_rule.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        element = client.forms.remove_logic_rule(
            "frm_replace_with_real_id",
            "elm_replace_with_real_id",
            "elm_followup",
        )
        print(f"removed logic rule from {element.id}; groups: {element.logic_groups}")


if __name__ == "__main__":
    main()
