"""Run with: python examples/forms/update.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import UpdateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        form = client.forms.update(
            "frm_replace_with_real_id",
            UpdateBody(name="Renamed survey", status="active"),
        )
        print(f"updated {form.id} -> {form.name} ({form.status})")


if __name__ == "__main__":
    main()
