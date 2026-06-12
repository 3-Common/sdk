"""Run with: python examples/forms/add_logic_rule.py

Reveal one element when a specific option is chosen on a selection element.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import AddLogicRuleBody, LogicCondition


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        element = client.forms.add_logic_rule(
            "frm_replace_with_real_id",
            "elm_source_id",
            AddLogicRuleBody(
                revealed_element_id="elm_revealed_id",
                condition=LogicCondition(option_indices=[0], operator="any_of"),
            ),
        )
        print(f"added logic rule on element {element.id} ({element.type})")


if __name__ == "__main__":
    main()
