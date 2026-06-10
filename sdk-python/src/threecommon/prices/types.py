"""Public types for the prices resource.

Hand-curated Pydantic models that mirror the wire shape (camelCase aliases
preserved). All response models use ``extra="ignore"`` so newer server-side
fields don't break older SDK versions.
"""

from __future__ import annotations

from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field

#: Price cadence.
#:
#: * ``recurring`` — billed on a fixed cadence (subscription-backed).
#: * ``one_time`` — single charge, typically an add-on / top-up pack.
PriceType = Literal["recurring", "one_time"]

#: Settlement currency of a price.
PriceCurrency = Literal["USD", "CAD"]

#: Cadence unit of a recurring price.
PriceInterval = Literal["day", "week", "month", "year"]


class _BaseModel(BaseModel):
    """Shared config: accept snake_case or camelCase, ignore unknown fields."""

    model_config = ConfigDict(
        populate_by_name=True,
        extra="ignore",
        str_strip_whitespace=False,
    )


class PriceRecurring(_BaseModel):
    """Recurring cadence descriptor, present when ``type`` is ``recurring``."""

    interval: PriceInterval
    interval_count: int = Field(
        serialization_alias="intervalCount", validation_alias="intervalCount"
    )


class PriceFeatureBoolean(_BaseModel):
    """A boolean feature grant — the feature is on or off."""

    feature_key: str = Field(serialization_alias="featureKey", validation_alias="featureKey")
    type: Literal["boolean"]
    enabled: bool


class PriceFeatureQuantity(_BaseModel):
    """A metered feature grant. ``quantity`` ``None`` means unlimited."""

    feature_key: str = Field(serialization_alias="featureKey", validation_alias="featureKey")
    type: Literal["quantity"]
    quantity: int | None
    rollover_enabled: bool = Field(
        serialization_alias="rolloverEnabled", validation_alias="rolloverEnabled"
    )
    rollover_cap: int | None = Field(
        default=None, serialization_alias="rolloverCap", validation_alias="rolloverCap"
    )
    expire_on_cancel: bool | None = Field(
        default=None, serialization_alias="expireOnCancel", validation_alias="expireOnCancel"
    )


class PriceFeatureEnum(_BaseModel):
    """An enum feature grant — selects one named value."""

    feature_key: str = Field(serialization_alias="featureKey", validation_alias="featureKey")
    type: Literal["enum"]
    enum_value: str = Field(serialization_alias="enumValue", validation_alias="enumValue")


class PriceFeatureDuration(_BaseModel):
    """A duration feature grant. ``duration_days`` ``None`` means unlimited."""

    feature_key: str = Field(serialization_alias="featureKey", validation_alias="featureKey")
    type: Literal["duration"]
    duration_days: int | None = Field(
        serialization_alias="durationDays", validation_alias="durationDays"
    )


#: One typed feature grant on a price. Discriminated on ``type``.
PriceFeature = Annotated[
    PriceFeatureBoolean | PriceFeatureQuantity | PriceFeatureEnum | PriceFeatureDuration,
    Field(discriminator="type"),
]


class Price(_BaseModel):
    """One price as returned by the API.

    Optional fields are populated only when the server returned them — list
    responses with a ``fields`` filter omit unrequested values.
    """

    id: str
    host_id: str | None = Field(
        default=None, serialization_alias="hostId", validation_alias="hostId"
    )
    product_id: str | None = Field(
        default=None, serialization_alias="productId", validation_alias="productId"
    )
    type: PriceType | None = None
    currency: PriceCurrency | None = None
    unit_amount: int | None = Field(
        default=None, serialization_alias="unitAmount", validation_alias="unitAmount"
    )
    recurring: PriceRecurring | None = None
    features: list[PriceFeature] | None = None
    nickname: str | None = None
    active: bool | None = None
    metadata: dict[str, str] | None = None
    created_at: str | None = Field(
        default=None, serialization_alias="createdAt", validation_alias="createdAt"
    )
    updated_at: str | None = Field(
        default=None, serialization_alias="updatedAt", validation_alias="updatedAt"
    )


class ListPricesResponse(_BaseModel):
    """Successful response shape from ``GET /v1/prices``."""

    data: list[Price]
    has_more: bool = Field(serialization_alias="hasMore", validation_alias="hasMore")


class ListParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/prices``."""

    page: int | None = None
    page_size: int | None = Field(
        default=None, serialization_alias="pageSize", validation_alias="pageSize"
    )
    product_id: str | None = Field(
        default=None, serialization_alias="productId", validation_alias="productId"
    )
    type: PriceType | None = None
    active: bool | None = None
    fields: str | None = None


class RetrieveParams(_BaseModel):
    """Query parameters accepted by ``GET /v1/prices/{id}``."""

    fields: str | None = None


class CreateBody(_BaseModel):
    """Body accepted by ``POST /v1/prices``.

    ``recurring`` is required when ``type`` is ``recurring`` and forbidden when
    ``type`` is ``one_time``.
    """

    product_id: str = Field(serialization_alias="productId", validation_alias="productId")
    type: PriceType
    currency: PriceCurrency
    unit_amount: int = Field(serialization_alias="unitAmount", validation_alias="unitAmount")
    recurring: PriceRecurring | None = None
    features: list[PriceFeature] | None = None
    nickname: str | None = None
    metadata: dict[str, str] | None = None


class UpdateBody(_BaseModel):
    """Body accepted by ``PATCH /v1/prices/{id}``.

    Only fields you set are sent. ``features``, ``nickname``, and ``metadata``
    accept an explicit ``None`` to clear the value server-side.
    """

    unit_amount: int | None = Field(
        default=None, serialization_alias="unitAmount", validation_alias="unitAmount"
    )
    recurring: PriceRecurring | None = None
    features: list[PriceFeature] | None = None
    nickname: str | None = None
    metadata: dict[str, str] | None = None
