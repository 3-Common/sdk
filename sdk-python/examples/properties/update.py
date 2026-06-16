"""Rename a property and clear its description. ``type`` and ``objectType``
cannot be modified; archive a property by setting ``status`` to ``archived``.

Run with: python examples/properties/update.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.properties import UpdateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        prop = client.properties.update(
            "prop_replace_with_real_id",
            UpdateBody(name="Allergies", description=None),
        )
        print(f"updated {prop.id} - now named {prop.name}")


if __name__ == "__main__":
    main()
