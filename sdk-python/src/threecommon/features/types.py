"""Public types for the features resource.

Hand-curated Pydantic models that mirror the wire shape (camelCase aliases
preserved). All response models use ``extra="ignore"`` so newer server-side
fields don't break older SDK versions.
"""

from __future__ import annotations

from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field

#: Feature value shape.
#:
#: * ``boolean`` — pure on/off.
#: * ``quantity`` — countable; drives entitlement balance.
#: * ``enum`` — one of a fixed ordered set of values.
#: * ``duration`` — number of days (or unlimited).
FeatureType = Literal["boolean", "quantity", "enum", "duration"]


class _BaseModel(BaseModel):
    """Shared config: accept snake_case or camelCase, ignore unknown fields."""

    model_config = ConfigDict(
        populate_by_name=True,
        extra="ignore",
        str_strip_whitespace=False,
    )


class Feature(_BaseModel):
    """One feature in the host's catalog.

    Optional fields are populated only when the server returned them — list
    responses with a ``fields`` filter omit unrequested values.
    """

    id: str
    host_id: str | None = Field(
        default=None, serialization_alias="hostId", validation_alias="hostId"
    )
    key: str | None = None
    name: str | None = None
    description: str | None = None
    type: FeatureType | None = None
    enum_values: list[str] | None = Field(
        default=None, serialization_alias="enumValues", validation_alias="enumValues"
    )
    active: bool | None = None
    metadata: dict[str, str] | None = None
    created_at: str | None = Field(
        default=None, serialization_alias="createdAt", validation_alias="createdAt"
    )
    updated_at: str | None = Field(
        default=None, serialization_alias="updatedAt", validation_alias="updatedAt"
    )


class ResolvedFeatureBoolean(_BaseModel):
    """Resolved value of a boolean feature."""

    type: Literal["boolean"]
    enabled: bool


class ResolvedFeatureQuantity(_BaseModel):
    """Resolved value of a quantity feature. ``quantity`` ``None`` = unlimited."""

    type: Literal["quantity"]
    quantity: int | None
    balance: int | None = None


class ResolvedFeatureEnum(_BaseModel):
    """Resolved value of an enum feature."""

    type: Literal["enum"]
    enum_value: str | None = Field(serialization_alias="enumValue", validation_alias="enumValue")


class ResolvedFeatureDuration(_BaseModel):
    """Resolved value of a duration feature. ``duration_days`` ``None`` = unlimited."""

    type: Literal["duration"]
    duration_days: int | None = Field(
        serialization_alias="durationDays", validation_alias="durationDays"
    )


#: The resolved type-specific value of a feature for a customer. Discriminated
#: on ``type``.
ResolvedFeatureValue = Annotated[
    ResolvedFeatureBoolean
    | ResolvedFeatureQuantity
    | ResolvedFeatureEnum
    | ResolvedFeatureDuration,
    Field(discriminator="type"),
]


class ResolvedFeature(_BaseModel):
    """The resolved state of a feature for a customer, returned by
    ``GET /v1/features/resolve``. Combines the catalog feature, the resolved
    value, and the subscriptions that contributed it.
    """

    feature: Feature
    value: ResolvedFeatureValue
    contributing_subscription_ids: list[str] = Field(
        serialization_alias="contributingSubscriptionIds",
        validation_alias="contributingSubscriptionIds",
    )


class ListFeaturesResponse(_BaseModel):
    """Successful response shape from ``GET /v1/features``."""

    data: list[Feature]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")


class ListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/features``."""

    page: int | None = None
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    type: FeatureType | None = None
    active: bool | None = None
    fields: str | None = None


class RetrieveParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/features/{id}``."""

    fields: str | None = None


class ResolveParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/features/resolve``."""

    contact_id: str = Field(serialization_alias="contactId", validation_alias="contactId")
    feature_key: str = Field(serialization_alias="featureKey", validation_alias="featureKey")


class CreateBody(_BaseModel):
    """Body accepted by ``POST /v1/features``."""

    key: str
    name: str
    type: FeatureType
    description: str | None = None
    enum_values: list[str] | None = Field(
        default=None, serialization_alias="enumValues", validation_alias="enumValues"
    )
    metadata: dict[str, str] | None = None


class UpdateBody(_BaseModel):
    """Body accepted by ``PATCH /v1/features/{id}``.

    Only fields you set are sent. ``description`` and ``metadata`` accept an
    explicit ``None`` to clear the value server-side. ``key`` and ``type`` are
    immutable.
    """

    name: str | None = None
    description: str | None = None
    enum_values: list[str] | None = Field(
        default=None, serialization_alias="enumValues", validation_alias="enumValues"
    )
    metadata: dict[str, str] | None = None
