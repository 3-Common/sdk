"""Public types for the properties resource.

Hand-curated Pydantic models that mirror the wire shape (camelCase aliases
preserved). All response models use ``extra="ignore"`` so newer server-side
fields don't break older SDK versions.
"""

from __future__ import annotations

from typing import Literal

from pydantic import BaseModel, ConfigDict, Field

#: The data type of a property. Set at creation time and immutable thereafter.
#:
#: ``Select One`` and ``Select Multiple`` properties additionally carry an
#: ``options`` array; every other type shares the same base shape.
PropertyType = Literal[
    "Text",
    "Multi-line Text",
    "Select One",
    "Yes/No",
    "Select Multiple",
    "Date",
    "File",
    "Email",
    "Phone",
]

#: The kind of object a property is attached to. Set at creation time and
#: immutable thereafter.
#:
#: * ``event`` - properties on events
#: * ``order`` - properties on orders (buyer-level)
#: * ``ticket`` - properties on individual products within an order (tickets, add-ons, etc.)
#: * ``contact`` - properties on customer contact records
PropertyObjectType = Literal["event", "order", "ticket", "contact"]

#: Lifecycle status of a property. ``archived`` properties are soft-deleted: any
#: existing reference remains valid, but only ``active`` properties should be
#: used in new workflows, forms, etc.
PropertyStatus = Literal["active", "archived"]

#: Field a ``list`` query can be sorted by. Defaults to ``name``.
PropertySortField = Literal["name", "description", "type", "objectType", "status"]

#: Sort direction for a ``list`` query. Defaults to ``asc``.
PropertySortOrder = Literal["asc", "desc"]


class _BaseModel(BaseModel):
    """Shared config: accept snake_case or camelCase, ignore unknown fields."""

    model_config = ConfigDict(
        populate_by_name=True,
        extra="ignore",
        str_strip_whitespace=False,
    )


class PropertyOption(_BaseModel):
    """A single selectable option on a ``Select One`` or ``Select Multiple``
    property.

    The ``value`` is the identity persisted on every instance that selected it;
    ``label`` is the display text.
    """

    value: str
    label: str


class Property(_BaseModel):
    """One property as returned by the API.

    ``options`` is populated only for ``Select One`` and ``Select Multiple``
    properties; every other type leaves it ``None``.
    """

    type: PropertyType
    id: str
    name: str
    status: PropertyStatus
    object_type: PropertyObjectType = Field(
        serialization_alias="objectType", validation_alias="objectType"
    )
    description: str | None = None
    options: list[PropertyOption] | None = None


class ListPropertiesResponse(_BaseModel):
    """Successful response shape from ``GET /v1/properties``."""

    data: list[Property]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")


class ListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/properties``."""

    page: int | None = None
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    object_type: PropertyObjectType | None = Field(
        default=None, serialization_alias="objectType", validation_alias="objectType"
    )
    property_type: PropertyType | None = Field(
        default=None, serialization_alias="propertyType", validation_alias="propertyType"
    )
    status: PropertyStatus | None = None
    sort: PropertySortField | None = None
    order: PropertySortOrder | None = None
    search: str | None = None


class CreateBody(_BaseModel):
    """Body accepted by ``POST /v1/properties``.

    ``type`` and ``objectType`` can only be set here and cannot be modified
    afterwards. For ``Select One`` and ``Select Multiple`` types, ``options`` is
    required and must have at least one entry.
    """

    type: PropertyType
    name: str
    status: PropertyStatus
    object_type: PropertyObjectType = Field(
        serialization_alias="objectType", validation_alias="objectType"
    )
    description: str | None = None
    options: list[PropertyOption] | None = None


class UpdateBody(_BaseModel):
    """Body accepted by ``PATCH /v1/properties/{id}``.

    Only fields you set are sent; ``type`` and ``objectType`` are immutable and
    cannot be changed. ``description`` accepts an explicit ``None`` to clear it
    server-side. To retire a property, set ``status`` to ``archived``
    (properties cannot be fully deleted).
    """

    name: str | None = None
    status: PropertyStatus | None = None
    options: list[PropertyOption] | None = None
    description: str | None = None
