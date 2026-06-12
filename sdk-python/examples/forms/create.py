"""Run with: python examples/forms/create.py

``type`` must be ``"standalone"`` or ``"order"`` and is checked client-side at
model construction. The server can still reject the request for other reasons
(e.g. an invalid name), which surfaces as ``ValidationError``.
"""

from __future__ import annotations

from threecommon import ThreeCommon, ValidationError
from threecommon.forms import CreateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        try:
            created = client.forms.create(CreateBody(name="Registration", type="standalone"))
            print(f"created {created.id} - {created.name} ({created.status})")
        except ValidationError as err:
            print(f"could not create form: {err.code}")


if __name__ == "__main__":
    main()
