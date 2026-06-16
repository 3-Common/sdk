"""Run with: python examples/forms/delete_element.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.forms.delete_element("frm_replace_with_real_id", "elm_replace_with_real_id")
        print(f"deleted element {result.deleted_element_id}")


if __name__ == "__main__":
    main()
