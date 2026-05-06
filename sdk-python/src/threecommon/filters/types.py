"""Wire types for the API's ``filters`` query parameter."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Literal, Union

#: Boolean connector for a [FilterGroup].
FilterLogic = Literal["and", "or"]

#: Full set of operators supported by the API.
FilterOperator = Literal[
    "is_empty",
    "is_not_empty",
    "is_equal_to",
    "is_not_equal_to",
    "is_equal_to_any_of",
    "is_not_equal_to_any_of",
    "is_any_of",
    "is_none_of",
    "contains",
    "contains_exactly",
    "is_before",
    "is_after",
    "is_greater_than",
    "is_greater_than_or_equal_to",
    "is_less_than",
    "is_less_than_or_equal_to",
    "is_between",
]


@dataclass(frozen=True, slots=True)
class FilterRange:
    """Range envelope used by ``is_between``.

    ``start`` and ``end`` must agree on type (both numeric or both ISO date strings).
    """

    start: int | float | str
    end: int | float | str


#: Every value the wire format accepts.
FilterValue = Union[
    str,
    int,
    float,
    bool,
    "list[str | int | float]",
    FilterRange,
]


@dataclass(frozen=True, slots=True)
class FilterCondition:
    """Single ``field operator value?`` triple."""

    field: str
    operator: FilterOperator
    value: FilterValue | None = None


@dataclass(frozen=True, slots=True)
class FilterGroup:
    """Logical group of conditions or nested groups."""

    logic: FilterLogic
    conditions: tuple[FilterCondition | FilterGroup, ...]
