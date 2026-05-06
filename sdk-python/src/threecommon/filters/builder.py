"""Typed filter builder. Produces JSON payloads for the API's ``filters`` arg."""

from __future__ import annotations

import json
from collections.abc import Iterable
from dataclasses import asdict
from typing import Any

from threecommon.filters.types import (
    FilterCondition,
    FilterGroup,
    FilterRange,
    FilterValue,
)


class FieldRef:
    """Reference to a single field. Operator methods produce a [FilterCondition]."""

    __slots__ = ("_name",)

    def __init__(self, name: str) -> None:
        if not name:
            msg = "filters.field: name must be a non-empty string"
            raise ValueError(msg)
        self._name = name

    # Existence
    def is_empty(self) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_empty")

    def is_not_empty(self) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_not_empty")

    # Equality
    def is_equal_to(self, value: str | int | float | bool) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_equal_to", value=value)

    def is_not_equal_to(self, value: str | int | float | bool) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_not_equal_to", value=value)

    # Set membership
    def is_equal_to_any_of(self, values: Iterable[str | int | float]) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_equal_to_any_of", value=list(values))

    def is_not_equal_to_any_of(self, values: Iterable[str | int | float]) -> FilterCondition:
        return FilterCondition(
            field=self._name, operator="is_not_equal_to_any_of", value=list(values)
        )

    def is_any_of(self, values: Iterable[str | int | float]) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_any_of", value=list(values))

    def is_none_of(self, values: Iterable[str | int | float]) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_none_of", value=list(values))

    # Substring
    def contains(self, value: str) -> FilterCondition:
        return FilterCondition(field=self._name, operator="contains", value=value)

    def contains_exactly(self, value: str) -> FilterCondition:
        return FilterCondition(field=self._name, operator="contains_exactly", value=value)

    # Date
    def is_before(self, value: str) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_before", value=value)

    def is_after(self, value: str) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_after", value=value)

    # Numeric
    def is_greater_than(self, value: int | float) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_greater_than", value=value)

    def is_greater_than_or_equal_to(self, value: int | float) -> FilterCondition:
        return FilterCondition(
            field=self._name, operator="is_greater_than_or_equal_to", value=value
        )

    def is_less_than(self, value: int | float) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_less_than", value=value)

    def is_less_than_or_equal_to(self, value: int | float) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_less_than_or_equal_to", value=value)

    # Range
    def is_between(self, value: FilterRange) -> FilterCondition:
        return FilterCondition(field=self._name, operator="is_between", value=value)


def field(name: str) -> FieldRef:
    """Reference a field by name. See [FieldRef] for available operators."""
    return FieldRef(name)


class SerializableFilter:
    """A [FilterGroup] augmented with JSON-string convenience methods.

    Returned by :func:`and_` / :func:`or_`. Can be nested inside another
    group because it exposes the underlying [group][SerializableFilter.group].
    """

    __slots__ = ("group",)

    def __init__(self, group: FilterGroup) -> None:
        self.group = group

    def to_filters(self) -> list[FilterGroup]:
        """Return ``[group]`` — the wire-format expects an array of groups."""
        return [self.group]

    def serialize(self) -> str:
        """JSON-string form, ready to assign to the ``filters`` query param."""
        return json.dumps([_dump(self.group)], separators=(",", ":"))


class CombinedFilters:
    """Multiple top-level groups joined implicitly with AND.

    Most callers want :func:`and_` / :func:`or_` instead — use :func:`combine`
    only when you specifically need an array of independent groups.
    """

    __slots__ = ("groups",)

    def __init__(self, groups: list[FilterGroup]) -> None:
        if not groups:
            msg = "filters.combine: at least one group is required"
            raise ValueError(msg)
        self.groups = groups

    def to_filters(self) -> list[FilterGroup]:
        return list(self.groups)

    def serialize(self) -> str:
        return json.dumps([_dump(g) for g in self.groups], separators=(",", ":"))


def and_(*items: FilterCondition | FilterGroup | SerializableFilter) -> SerializableFilter:
    """Combine items with AND. At least one is required."""
    return _make("and", items)


def or_(*items: FilterCondition | FilterGroup | SerializableFilter) -> SerializableFilter:
    """Combine items with OR. At least one is required."""
    return _make("or", items)


def combine(*groups: FilterGroup | SerializableFilter) -> CombinedFilters:
    """Combine multiple top-level groups."""
    flattened = [g.group if isinstance(g, SerializableFilter) else g for g in groups]
    return CombinedFilters(flattened)


def _make(
    logic: str,
    items: tuple[FilterCondition | FilterGroup | SerializableFilter, ...],
) -> SerializableFilter:
    if not items:
        msg = f"filters.{logic}_: at least one condition or group is required"
        raise ValueError(msg)
    flattened: list[FilterCondition | FilterGroup] = []
    for it in items:
        if isinstance(it, SerializableFilter):
            flattened.append(it.group)
        else:
            flattened.append(it)
    return SerializableFilter(FilterGroup(logic=logic, conditions=tuple(flattened)))  # type: ignore[arg-type]


def _dump(node: FilterCondition | FilterGroup) -> dict[str, Any]:
    """Recursively convert a condition/group into the JSON-shape dict."""
    if isinstance(node, FilterGroup):
        return {
            "logic": node.logic,
            "conditions": [_dump(c) for c in node.conditions],
        }
    payload: dict[str, Any] = {"field": node.field, "operator": node.operator}
    if node.value is not None:
        payload["value"] = _dump_value(node.value)
    return payload


def _dump_value(value: FilterValue) -> Any:
    if isinstance(value, FilterRange):
        return asdict(value)
    return value
