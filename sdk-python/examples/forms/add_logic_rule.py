"""Run with: python examples/forms/add_logic_rule.py

Conditional logic reveals a target element based on a source element's answer.
Selection questions match on which options are chosen
(``SelectionLogicCondition``); Yes/No questions match on the answer value
(``YesNoLogicCondition``).
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import AddLogicRuleBody, SelectionLogicCondition, YesNoLogicCondition


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        # Selection question: reveal the target when the first option is chosen.
        element = client.forms.add_logic_rule(
            "frm_replace_with_real_id",
            "elm_select_id",
            AddLogicRuleBody(
                revealed_element_id="elm_revealed_id",
                condition=SelectionLogicCondition(option_indices=[0], operator="any_of"),
            ),
        )
        print(f"added selection rule on element {element.id} ({element.type})")

        # Yes/No question: reveal the target when the respondent answers "yes".
        element = client.forms.add_logic_rule(
            "frm_replace_with_real_id",
            "elm_yes_no_id",
            AddLogicRuleBody(
                revealed_element_id="elm_other_revealed_id",
                condition=YesNoLogicCondition(selection_type="is", value=True),
            ),
        )
        print(f"added Yes/No rule on element {element.id} ({element.type})")


if __name__ == "__main__":
    main()
