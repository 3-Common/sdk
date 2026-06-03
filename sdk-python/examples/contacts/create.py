"""Run with: python examples/contacts/create.py

Fails with ``ConflictError`` if a contact with the same email already exists
for this host.
"""

from __future__ import annotations

from threecommon import ConflictError, ThreeCommon
from threecommon.contacts import CreateBody


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        try:
            created = client.contacts.create(
                CreateBody(
                    email="guest@example.com",
                    first_name="Alex",
                    last_name="Garcia",
                )
            )
            print(f"created {created.id} <{created.email}>")
        except ConflictError:
            print("contact with that email already exists for this host")


if __name__ == "__main__":
    main()
