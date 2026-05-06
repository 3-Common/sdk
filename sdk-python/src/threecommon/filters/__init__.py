"""Typed builder + wire types for the API's ``filters`` query parameter.

Every endpoint that accepts ``filters`` consumes this same shape, so the
builder lives in a shared package rather than per-resource. Build filters
with :func:`field`, combine groups with :func:`and_` / :func:`or_` /
:func:`combine`, and serialize before passing as the ``filters`` query
argument:

    from threecommon import filters

    f = filters.and_(
        filters.field("status").is_any_of(["open"]),
        filters.field("ticket_sum").is_greater_than(10),
    )
    client.events.list(filters=f.serialize())

"""

from threecommon.filters.builder import (
    CombinedFilters,
    FieldRef,
    SerializableFilter,
    and_,
    combine,
    field,
    or_,
)
from threecommon.filters.types import (
    FilterCondition,
    FilterGroup,
    FilterLogic,
    FilterOperator,
    FilterRange,
    FilterValue,
)

__all__ = (
    "CombinedFilters",
    "FieldRef",
    "FilterCondition",
    "FilterGroup",
    "FilterLogic",
    "FilterOperator",
    "FilterRange",
    "FilterValue",
    "SerializableFilter",
    "and_",
    "combine",
    "field",
    "or_",
)
