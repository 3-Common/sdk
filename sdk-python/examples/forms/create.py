"""Run with: python examples/forms/create.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import CreateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        form = client.forms.create(
            CreateBody(name="Customer survey", type="standalone", status="draft")
        )
        print(f"created {form.id} ({form.name}) status={form.status}")


if __name__ == "__main__":
    main()
