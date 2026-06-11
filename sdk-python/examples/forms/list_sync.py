"""Run with: python examples/forms/list_sync.py"""

from __future__ import annotations

from threecommon import ThreeCommon
from threecommon.forms import ListParams


def main() -> None:
    with ThreeCommon(api_key="3co_your_api_key_here") as client:
        result = client.forms.list(ListParams(type="standalone", page_size=50))
        print(f"got {len(result.data)} forms (has_more={result.has_more})\n")
        for f in result.data:
            print(f"  {f.id} - {f.name} ({f.status}, {f.num_elements} elements)")


if __name__ == "__main__":
    main()
