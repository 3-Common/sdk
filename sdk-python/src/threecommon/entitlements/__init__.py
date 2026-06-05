"""Entitlements resource — sync and async clients plus public types.

Most callers reach this module through
[ThreeCommon.entitlements][threecommon.ThreeCommon] /
[AsyncThreeCommon.entitlements][threecommon.AsyncThreeCommon]; importing the
service classes directly is supported for advanced wiring.
"""

from threecommon.entitlements.service import (
    AsyncEntitlementsService,
    EntitlementsService,
)
from threecommon.entitlements.types import (
    ConsumeBody,
    Entitlement,
    EntitlementGrant,
    EntitlementGrantSource,
    GrantBody,
    ListEntitlementsResponse,
    ListParams,
    LookupParams,
    RetrieveParams,
)

__all__ = (
    "AsyncEntitlementsService",
    "ConsumeBody",
    "Entitlement",
    "EntitlementGrant",
    "EntitlementGrantSource",
    "EntitlementsService",
    "GrantBody",
    "ListEntitlementsResponse",
    "ListParams",
    "LookupParams",
    "RetrieveParams",
)
