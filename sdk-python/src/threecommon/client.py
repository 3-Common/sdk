"""SDK clients: [ThreeCommon][] (sync) and [AsyncThreeCommon][] (async).

Construct once per process; both classes are safe to share across threads and
tasks. Each instance owns one underlying ``httpx`` client unless you supply
your own via the ``http_client`` / ``async_http_client`` kwarg.
"""

from __future__ import annotations

from types import TracebackType
from typing import TYPE_CHECKING

from threecommon._core.http_client import (
    AsyncHTTPClient,
    HTTPClient,
    HTTPClientOptions,
)
from threecommon._core.retry import RetryPolicy
from threecommon._core.telemetry import Telemetry
from threecommon.config import RetryDelay, resolve_config
from threecommon.contacts.service import AsyncContactsService, ContactsService
from threecommon.entitlements.service import AsyncEntitlementsService, EntitlementsService
from threecommon.events.service import AsyncEventsService, EventsService
from threecommon.invoices.service import AsyncInvoicesService, InvoicesService
from threecommon.subscriptions.service import AsyncSubscriptionsService, SubscriptionsService

if TYPE_CHECKING:  # pragma: no cover
    import logging

    import httpx


class ThreeCommon:
    """Synchronous entry point.

    Construct with at minimum an API key, then call any resource method.
    Closing is optional but recommended for short-lived scripts:

        client = ThreeCommon(api_key="3co_...")
        try:
            result = client.events.list()
        finally:
            client.close()

    Or use as a context manager:

        with ThreeCommon(api_key="3co_...") as client:
            result = client.events.list()
    """

    events: EventsService
    """Events resource — ``GET /v1/events``, ``GET /v1/events/{id}``, ``PATCH /v1/events/{id}``."""

    invoices: InvoicesService
    """Invoices resource — list, retrieve, create, update, finalize, void, record_payment."""

    subscriptions: SubscriptionsService
    """Subscriptions resource — list, retrieve, create, update, activate, cancel, bill, renew."""

    contacts: ContactsService
    """Contacts resource — ``list``, ``count``, ``retrieve``, ``create``,
    ``update``, ``delete``, ``bulk_upsert``, ``list_activity``, plus
    auto-paginators."""

    entitlements: EntitlementsService
    """Entitlements resource — ``list``, ``retrieve``, ``lookup``, ``grant``,
    ``consume``, plus ``list_auto_paginate``."""

    _http: HTTPClient
    _telemetry: Telemetry

    def __init__(
        self,
        *,
        api_key: str | None = None,
        base_url: str | None = None,
        api_version: str | None = None,
        timeout_seconds: float | None = None,
        max_retries: int | None = None,
        retry_delay: RetryDelay | None = None,
        http_client: httpx.Client | None = None,
        logger: logging.Logger | None = None,
        telemetry: bool | None = None,
    ) -> None:
        cfg = resolve_config(
            api_key=api_key,
            base_url=base_url,
            api_version=api_version,
            timeout_seconds=timeout_seconds,
            max_retries=max_retries,
            retry_delay=retry_delay,
            http_client=http_client,
            logger=logger,
            telemetry=telemetry,
        )
        self._telemetry = Telemetry(enabled=cfg.telemetry)
        self._http = HTTPClient(
            HTTPClientOptions(
                api_key=cfg.api_key,
                base_url=cfg.base_url,
                api_version=cfg.api_version,
                timeout_seconds=cfg.timeout_seconds,
                retry=RetryPolicy.from_delay(cfg.max_retries, cfg.retry_delay),
                telemetry=self._telemetry,
                logger=cfg.logger,
                httpx_client=cfg.http_client,
            )
        )
        self.events = EventsService(self._http)
        self.invoices = InvoicesService(self._http)
        self.subscriptions = SubscriptionsService(self._http)
        self.contacts = ContactsService(self._http)
        self.entitlements = EntitlementsService(self._http)

    def close(self) -> None:
        """Close the underlying httpx client (no-op if you supplied your own)."""
        self._http.close()

    def disable_telemetry(self) -> None:
        """Stop sending the ``Threecommon-Client-Telemetry`` header at runtime."""
        self._telemetry.disable()

    def __enter__(self) -> ThreeCommon:
        return self

    def __exit__(
        self,
        exc_type: type[BaseException] | None,
        exc: BaseException | None,
        tb: TracebackType | None,
    ) -> None:
        del exc_type, exc, tb
        self.close()


class AsyncThreeCommon:
    """Asynchronous entry point. Same surface as [ThreeCommon] with `await`-able methods."""

    events: AsyncEventsService
    invoices: AsyncInvoicesService
    subscriptions: AsyncSubscriptionsService
    contacts: AsyncContactsService
    entitlements: AsyncEntitlementsService

    _http: AsyncHTTPClient
    _telemetry: Telemetry

    def __init__(
        self,
        *,
        api_key: str | None = None,
        base_url: str | None = None,
        api_version: str | None = None,
        timeout_seconds: float | None = None,
        max_retries: int | None = None,
        retry_delay: RetryDelay | None = None,
        async_http_client: httpx.AsyncClient | None = None,
        logger: logging.Logger | None = None,
        telemetry: bool | None = None,
    ) -> None:
        cfg = resolve_config(
            api_key=api_key,
            base_url=base_url,
            api_version=api_version,
            timeout_seconds=timeout_seconds,
            max_retries=max_retries,
            retry_delay=retry_delay,
            async_http_client=async_http_client,
            logger=logger,
            telemetry=telemetry,
        )
        self._telemetry = Telemetry(enabled=cfg.telemetry)
        self._http = AsyncHTTPClient(
            HTTPClientOptions(
                api_key=cfg.api_key,
                base_url=cfg.base_url,
                api_version=cfg.api_version,
                timeout_seconds=cfg.timeout_seconds,
                retry=RetryPolicy.from_delay(cfg.max_retries, cfg.retry_delay),
                telemetry=self._telemetry,
                logger=cfg.logger,
                async_httpx_client=cfg.async_http_client,
            )
        )
        self.events = AsyncEventsService(self._http)
        self.invoices = AsyncInvoicesService(self._http)
        self.subscriptions = AsyncSubscriptionsService(self._http)
        self.contacts = AsyncContactsService(self._http)
        self.entitlements = AsyncEntitlementsService(self._http)

    async def aclose(self) -> None:
        """Close the underlying async httpx client."""
        await self._http.aclose()

    def disable_telemetry(self) -> None:
        self._telemetry.disable()

    async def __aenter__(self) -> AsyncThreeCommon:
        return self

    async def __aexit__(
        self,
        exc_type: type[BaseException] | None,
        exc: BaseException | None,
        tb: TracebackType | None,
    ) -> None:
        del exc_type, exc, tb
        await self.aclose()
