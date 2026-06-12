"""Run with: python examples/forms/update.py

Only the fields you set on ``UpdateBody`` are changed; everything else is
left untouched.
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import UpdateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        updated = client.forms.update(
            "frm_replace_with_real_id",
            UpdateBody(
                name="Updated Registration",
                status="active",
                submit_button_text="Sign up",
            ),
        )
        print(f"updated {updated.id} -> {updated.name} ({updated.status})")


if __name__ == "__main__":
    main()
