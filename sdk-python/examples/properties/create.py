"""Create a Select One property with two options. ``type`` and ``objectType``
are fixed at creation time and cannot be changed afterwards.

Run with: python examples/properties/create.py
"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.properties import CreateBody, PropertyOption


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        prop = client.properties.create(
            CreateBody(
                type="Select One",
                name="T-shirt size",
                status="active",
                object_type="contact",
                options=[
                    PropertyOption(value="s", label="Small"),
                    PropertyOption(value="m", label="Medium"),
                ],
            )
        )
        print(f"created {prop.id} - {prop.type} - {prop.name}")


if __name__ == "__main__":
    main()
