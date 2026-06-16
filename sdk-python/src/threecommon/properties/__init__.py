"""Properties resource - sync and async clients plus public types.

Most callers reach this module through
[ThreeCommon.properties][threecommon.ThreeCommon] /
[AsyncThreeCommon.properties][threecommon.AsyncThreeCommon]; importing the
service classes directly is supported for advanced wiring.
"""

from threecommon.properties.service import AsyncPropertiesService, PropertiesService
from threecommon.properties.types import (
    CreateBody,
    ListParams,
    ListPropertiesResponse,
    Property,
    PropertyObjectType,
    PropertyOption,
    PropertySortField,
    PropertySortOrder,
    PropertyStatus,
    PropertyType,
    UpdateBody,
)

__all__ = (
    "AsyncPropertiesService",
    "CreateBody",
    "ListParams",
    "ListPropertiesResponse",
    "PropertiesService",
    "Property",
    "PropertyObjectType",
    "PropertyOption",
    "PropertySortField",
    "PropertySortOrder",
    "PropertyStatus",
    "PropertyType",
    "UpdateBody",
)
