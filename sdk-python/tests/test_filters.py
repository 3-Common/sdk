"""Filter builder tests."""

from __future__ import annotations

import json

import pytest

from threecommon import filters


def test_field_panics_on_empty_name() -> None:
    with pytest.raises(ValueError, match="non-empty string"):
        filters.field("")


def test_and_serializes_to_wire_format() -> None:
    f = filters.and_(
        filters.field("status").is_any_of(["open"]),
        filters.field("ticket_sum").is_greater_than(10),
    )
    payload = json.loads(f.serialize())
    assert payload[0]["logic"] == "and"
    assert len(payload[0]["conditions"]) == 2


def test_or_accepts_nested_groups() -> None:
    inner = filters.and_(filters.field("ticket_sum").is_greater_than(0))
    outer = filters.or_(filters.field("status").is_equal_to("open"), inner)
    text = outer.serialize()
    assert '"logic":"or"' in text
    assert '"logic":"and"' in text


def test_and_panics_on_empty() -> None:
    with pytest.raises(ValueError, match="at least one"):
        filters.and_()
    with pytest.raises(ValueError, match="at least one"):
        filters.or_()


def test_combine_joins_top_level_groups() -> None:
    a = filters.and_(filters.field("status").is_equal_to("open"))
    b = filters.or_(filters.field("type").is_any_of(["event"]))
    combined = filters.combine(a, b)
    payload = json.loads(combined.serialize())
    assert len(payload) == 2


def test_combine_panics_on_empty() -> None:
    with pytest.raises(ValueError, match="at least one"):
        filters.combine()


def test_field_operators() -> None:
    """Smoke-cover all 17 operators."""
    f = filters.field("x")
    cases = [
        f.is_empty(),
        f.is_not_empty(),
        f.is_equal_to(1),
        f.is_not_equal_to(1),
        f.is_equal_to_any_of([1, 2]),
        f.is_not_equal_to_any_of([1]),
        f.is_any_of(["a"]),
        f.is_none_of(["a"]),
        f.contains("a"),
        f.contains_exactly("a"),
        f.is_before("2026-01-01"),
        f.is_after("2026-01-01"),
        f.is_greater_than(5),
        f.is_greater_than_or_equal_to(5),
        f.is_less_than(5),
        f.is_less_than_or_equal_to(5),
        f.is_between(filters.FilterRange(start=1, end=10)),
    ]
    assert len(cases) == 17
    for c in cases:
        assert c.field == "x"
        assert c.operator
