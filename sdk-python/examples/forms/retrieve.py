"""Run with: python examples/forms/retrieve.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        form = client.forms.retrieve("frm_replace_with_real_id")
        print(f"{form.name} ({form.type})")
        print(f"  status:   {form.status}")
        print(f"  owner_id: {form.owner_id}")
        print(f"  elements: {len(form.elements)}")
        for element in form.elements:
            print(f"    - {element.type}: {element.prompt}")


if __name__ == "__main__":
    main()
