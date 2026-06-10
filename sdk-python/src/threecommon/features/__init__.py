"""Features resource — sync and async clients plus public types.

Most callers reach this module through
[ThreeCommon.features][threecommon.ThreeCommon] /
[AsyncThreeCommon.features][threecommon.AsyncThreeCommon]; importing the service
classes directly is supported for advanced wiring.
"""

from threecommon.features.service import AsyncFeaturesService, FeaturesService
from threecommon.features.types import (
    CreateBody,
    Feature,
    FeatureType,
    ListFeaturesResponse,
    ListParams,
    ResolvedFeature,
    ResolvedFeatureBoolean,
    ResolvedFeatureDuration,
    ResolvedFeatureEnum,
    ResolvedFeatureQuantity,
    ResolveParams,
    RetrieveParams,
    UpdateBody,
)

__all__ = (
    "AsyncFeaturesService",
    "CreateBody",
    "Feature",
    "FeatureType",
    "FeaturesService",
    "ListFeaturesResponse",
    "ListParams",
    "ResolveParams",
    "ResolvedFeature",
    "ResolvedFeatureBoolean",
    "ResolvedFeatureDuration",
    "ResolvedFeatureEnum",
    "ResolvedFeatureQuantity",
    "RetrieveParams",
    "UpdateBody",
)
