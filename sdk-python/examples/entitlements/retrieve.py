"""Run with: python examples/entitlements/retrieve.py"""

from __future__ import annotations

from threecommon import ThreeCommon


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        ent = client.entitlements.retrieve("ent_replace_with_real_id")
        print(f"entitlement {ent.id} [{ent.feature_key}]")
        print(f"  contact        {ent.contact_id}")
        print(f"  balance        {ent.balance}")
        print(f"  total_granted  {ent.total_granted}")
        print(f"  total_consumed {ent.total_consumed}")
        for grant in ent.grants or []:
            print(f"  grant {grant.id} [{grant.source}] {grant.remaining}/{grant.amount} remaining")


if __name__ == "__main__":
    main()
