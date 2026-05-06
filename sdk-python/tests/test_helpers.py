from __future__ import annotations

import pytest

from threecommon.helpers import not_none


def test_not_none_returns_value() -> None:
    assert not_none("x", "field") == "x"
    assert not_none(0, "n") == 0


def test_not_none_raises_on_none() -> None:
    with pytest.raises(ValueError, match="field must not be None"):
        not_none(None, "field")
