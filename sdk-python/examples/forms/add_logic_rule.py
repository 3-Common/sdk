"""Run with: python examples/forms/add_logic_rule.py

Reveal a follow-up element when the first option of a selection question is
chosen.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import AddLogicRuleBody, LogicCondition


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        element = client.forms.add_logic_rule(
            "frm_replace_with_real_id",
            "elm_replace_with_real_id",
            AddLogicRuleBody(
                revealed_element_id="elm_followup",
                condition=LogicCondition(option_indices=[0], operator="any_of"),
            ),
        )
        print(f"added logic rule to {element.id}; groups: {element.logic_groups}")


if __name__ == "__main__":
    main()
