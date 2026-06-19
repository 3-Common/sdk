"""Contacts service tests — sync + async via pytest-httpx."""

from __future__ import annotations

import pytest
from pytest_httpx import HTTPXMock

from threecommon import (
    AsyncThreeCommon,
    ConflictError,
    NotFoundError,
    ThreeCommon,
    ValidationError,
)
from threecommon.contacts import (
    ActivityListParams,
    AttachPaymentMethodBody,
    BulkUpsertBody,
    BulkUpsertItem,
    ContactUpdate,
    CreateBody,
    ListParams,
    UpdateBody,
)


def _make_sync() -> ThreeCommon:
    return ThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


def _make_async() -> AsyncThreeCommon:
    return AsyncThreeCommon(api_key="3co_test", base_url="http://test.local", max_retries=0)


SAMPLE_CONTACT = {
    "id": "cnt_123",
    "firstName": "Alex",
    "lastName": "Garcia",
    "fullName": "Alex Garcia",
    "email": "alex@example.com",
    "phone": "+15555550123",
    "vendorId": "hst_1",
    "orderSum": 3,
    "grossSum": 15000,
    "firstOrder": 1_700_000_000_000,
    "lastOrder": 1_710_000_000_000,
    "createdAt": "2026-01-01T00:00:00.000Z",
    "status": "opted-in",
    "eventsAttended_IDS": ["evt_a", "evt_b"],
    "itemsPurchased_IDS": [],
    "productsPurchased_IDS": [],
}

SAMPLE_ORDER_DETAILS = {
    "_id": "cnt_123",
    "email": "alex@example.com",
    "vendorId": "hst_1",
    "firstName": "Alex",
    "lastName": "Garcia",
    "fullName": "Alex Garcia",
    "status": "opted-in",
    "grossSum": 15000,
    "orderSum": 3,
    "events_attended": [],
    "items_purchased": [],
    "products_purchased": [],
}

SAMPLE_PAYMENT_METHOD = {
    "id": "pm_123",
    "contactId": "cnt_123",
    "card": {
        "brand": "visa",
        "last4": "4242",
        "expMonth": 12,
        "expYear": 2030,
        "country": "US",
        "funding": "credit",
    },
    "billingDetails": {
        "name": "Alex Garcia",
        "email": "alex@example.com",
        "postalCode": "94107",
        "country": "US",
    },
    "status": "active",
    "createdAt": "2026-06-17T12:00:00.000Z",
    "updatedAt": "2026-06-17T12:00:00.000Z",
}


# ---------------
# Sync contacts
# ---------------


def test_list_decodes_response(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts?filter=opted-in&pageSize=10",
        json={
            "data": [SAMPLE_CONTACT],
            "hasMore": False,
            "pageNumber": 0,
            "pageSize": 10,
        },
    )
    with _make_sync() as c:
        result = c.contacts.list(ListParams(filter="opted-in", page_size=10))
    assert len(result.data) == 1
    assert result.data[0].id == "cnt_123"
    assert result.has_more is False
    assert result.page_number == 0
    assert result.page_size == 10


def test_list_nil_params(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts",
        json={"data": [], "hasMore": False, "pageNumber": 0, "pageSize": 20},
    )
    with _make_sync() as c:
        result = c.contacts.list()
    assert result.data == []


def test_count(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/count",
        json={"data": {"count": 4823}},
    )
    with _make_sync() as c:
        result = c.contacts.count()
    assert result.count == 4823


def test_retrieve_uses_envelope(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123",
        json={"data": SAMPLE_CONTACT},
    )
    with _make_sync() as c:
        contact = c.contacts.retrieve("cnt_123")
    assert contact.id == "cnt_123"
    assert contact.email == "alex@example.com"


def test_retrieve_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.retrieve("")
    assert exc.value.code == "missing_id"


def test_retrieve_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_missing",
        status_code=404,
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.contacts.retrieve("cnt_missing")


def test_create_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts",
        method="POST",
        match_json={"email": "alex@example.com", "firstName": "Alex"},
        json={"data": SAMPLE_CONTACT},
    )
    with _make_sync() as c:
        contact = c.contacts.create(CreateBody(email="alex@example.com", first_name="Alex"))
    assert contact.id == "cnt_123"


def test_create_409_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts",
        method="POST",
        status_code=409,
        json={"error": {"code": "conflict", "message": "duplicate"}},
    )
    with _make_sync() as c, pytest.raises(ConflictError):
        c.contacts.create(CreateBody(email="alex@example.com"))


def test_create_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.create(None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_update_sends_nested_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123",
        method="PATCH",
        match_json={
            "contact": {
                "firstName": "Alex",
                "lastName": "Garcia",
                "email": "a.garcia@example.com",
                "status": "opted-in",
            }
        },
        json={"data": {**SAMPLE_ORDER_DETAILS, "email": "a.garcia@example.com"}},
    )
    with _make_sync() as c:
        updated = c.contacts.update(
            "cnt_123",
            UpdateBody(
                contact=ContactUpdate(
                    first_name="Alex",
                    last_name="Garcia",
                    email="a.garcia@example.com",
                    status="opted-in",
                )
            ),
        )
    assert updated.id_ == "cnt_123"
    assert updated.email == "a.garcia@example.com"


def test_update_with_merge_resolution(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123",
        method="PATCH",
        match_json={
            "contact": {
                "firstName": "A",
                "lastName": "G",
                "email": "a@example.com",
                "status": "opted-in",
            },
            "mergeWith": "cnt_456",
            "resolution": "safe-merge",
        },
        json={"data": SAMPLE_ORDER_DETAILS},
    )
    with _make_sync() as c:
        c.contacts.update(
            "cnt_123",
            UpdateBody(
                contact=ContactUpdate(
                    first_name="A", last_name="G", email="a@example.com", status="opted-in"
                ),
                merge_with="cnt_456",
                resolution="safe-merge",
            ),
        )


def test_update_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.update(
            "",
            UpdateBody(
                contact=ContactUpdate(
                    first_name="A", last_name="B", email="x@example.com", status="opted-in"
                )
            ),
        )
    assert exc.value.code == "missing_id"


def test_update_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.update("cnt_123", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_delete_returns_id(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123",
        method="DELETE",
        json={"data": {"id": "cnt_123"}},
    )
    with _make_sync() as c:
        result = c.contacts.delete("cnt_123")
    assert result.id == "cnt_123"


def test_delete_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.contacts.delete("")


def test_delete_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_missing",
        method="DELETE",
        status_code=404,
        json={"error": {"code": "not_found", "message": "gone"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.contacts.delete("cnt_missing")


def test_bulk_upsert(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/bulk",
        method="POST",
        json={"data": {"affected": 2}},
    )
    with _make_sync() as c:
        result = c.contacts.bulk_upsert(
            BulkUpsertBody(
                contacts=[
                    BulkUpsertItem(email="a@example.com"),
                    BulkUpsertItem(email="b@example.com"),
                ]
            )
        )
    assert result.affected == 2


def test_bulk_upsert_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.contacts.bulk_upsert(None)  # type: ignore[arg-type]


def test_list_activity_happy(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/activity?filter=email_sent",
        json={
            "data": [
                {
                    "_id": "act_1",
                    "vendor_id": "hst_1",
                    "email": "alex@example.com",
                    "contact_id": "cnt_123",
                    "type": "email_sent",
                    "data": {"subject": "hi"},
                    "createdAt": "2026-05-01T00:00:00.000Z",
                    "updatedAt": "2026-05-01T00:00:00.000Z",
                }
            ],
            "hasMore": False,
            "pageNumber": 0,
            "pageSize": 20,
        },
    )
    with _make_sync() as c:
        result = c.contacts.list_activity("cnt_123", ActivityListParams(filter="email_sent"))
    assert len(result.data) == 1
    assert result.data[0].type == "email_sent"


def test_list_activity_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.contacts.list_activity("")


def test_list_activity_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_missing/activity",
        status_code=404,
        json={"error": {"code": "not_found", "message": "gone"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.contacts.list_activity("cnt_missing")


def test_list_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts?pageNumber=0&filter=opted-in",
        json={
            "data": [{**SAMPLE_CONTACT, "id": "cnt_1"}, {**SAMPLE_CONTACT, "id": "cnt_2"}],
            "hasMore": True,
            "pageNumber": 0,
            "pageSize": 20,
        },
    )
    httpx_mock.add_response(
        url="http://test.local/v1/contacts?pageNumber=1&filter=opted-in",
        json={
            "data": [{**SAMPLE_CONTACT, "id": "cnt_3"}],
            "hasMore": False,
            "pageNumber": 1,
            "pageSize": 20,
        },
    )
    with _make_sync() as c:
        ids = [
            contact.id for contact in c.contacts.list_auto_paginate(ListParams(filter="opted-in"))
        ]
    assert ids == ["cnt_1", "cnt_2", "cnt_3"]


def test_list_activity_auto_paginate_walks_pages(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/activity?pageNumber=0",
        json={
            "data": [
                {
                    "_id": "act_1",
                    "vendor_id": "hst_1",
                    "email": "alex@example.com",
                    "type": "email_sent",
                    "data": {},
                    "createdAt": "2026-05-01T00:00:00.000Z",
                    "updatedAt": "2026-05-01T00:00:00.000Z",
                }
            ],
            "hasMore": True,
            "pageNumber": 0,
            "pageSize": 20,
        },
    )
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/activity?pageNumber=1",
        json={
            "data": [
                {
                    "_id": "act_2",
                    "vendor_id": "hst_1",
                    "email": "alex@example.com",
                    "type": "ticket_scanned",
                    "data": {},
                    "createdAt": "2026-05-02T00:00:00.000Z",
                    "updatedAt": "2026-05-02T00:00:00.000Z",
                }
            ],
            "hasMore": False,
            "pageNumber": 1,
            "pageSize": 20,
        },
    )
    with _make_sync() as c:
        ids = [act.id_ for act in c.contacts.list_activity_auto_paginate("cnt_123")]
    assert ids == ["act_1", "act_2"]


def test_list_activity_auto_paginate_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError):
        c.contacts.list_activity_auto_paginate("")


# ---------------
# Async contacts
# ---------------


@pytest.mark.asyncio
async def test_async_list_decodes(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts",
        json={"data": [], "hasMore": False, "pageNumber": 0, "pageSize": 20},
    )
    async with _make_async() as c:
        r = await c.contacts.list()
    assert r.data == []


@pytest.mark.asyncio
async def test_async_count(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/count", json={"data": {"count": 100}}
    )
    async with _make_async() as c:
        r = await c.contacts.count()
    assert r.count == 100


@pytest.mark.asyncio
async def test_async_retrieve(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_1", json={"data": {**SAMPLE_CONTACT, "id": "cnt_1"}}
    )
    async with _make_async() as c:
        contact = await c.contacts.retrieve("cnt_1")
    assert contact.id == "cnt_1"


@pytest.mark.asyncio
async def test_async_create(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts",
        method="POST",
        json={"data": SAMPLE_CONTACT},
    )
    async with _make_async() as c:
        contact = await c.contacts.create(CreateBody(email="alex@example.com"))
    assert contact.id == "cnt_123"


@pytest.mark.asyncio
async def test_async_create_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.contacts.create(None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_update(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123",
        method="PATCH",
        json={"data": SAMPLE_ORDER_DETAILS},
    )
    async with _make_async() as c:
        updated = await c.contacts.update(
            "cnt_123",
            UpdateBody(
                contact=ContactUpdate(
                    first_name="A", last_name="G", email="a@example.com", status="opted-in"
                )
            ),
        )
    assert updated.id_ == "cnt_123"


@pytest.mark.asyncio
async def test_async_update_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.contacts.update("cnt_1", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_delete(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_1",
        method="DELETE",
        json={"data": {"id": "cnt_1"}},
    )
    async with _make_async() as c:
        result = await c.contacts.delete("cnt_1")
    assert result.id == "cnt_1"


@pytest.mark.asyncio
async def test_async_bulk_upsert(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/bulk",
        method="POST",
        json={"data": {"affected": 3}},
    )
    async with _make_async() as c:
        result = await c.contacts.bulk_upsert(
            BulkUpsertBody(contacts=[BulkUpsertItem(email=f"u{i}@example.com") for i in range(3)])
        )
    assert result.affected == 3


@pytest.mark.asyncio
async def test_async_bulk_upsert_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.contacts.bulk_upsert(None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_list_activity(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_1/activity",
        json={"data": [], "hasMore": False, "pageNumber": 0, "pageSize": 20},
    )
    async with _make_async() as c:
        r = await c.contacts.list_activity("cnt_1")
    assert r.data == []


@pytest.mark.asyncio
async def test_async_list_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts?pageNumber=0",
        json={
            "data": [{**SAMPLE_CONTACT, "id": "a"}, {**SAMPLE_CONTACT, "id": "b"}],
            "hasMore": True,
            "pageNumber": 0,
            "pageSize": 20,
        },
    )
    httpx_mock.add_response(
        url="http://test.local/v1/contacts?pageNumber=1",
        json={
            "data": [{**SAMPLE_CONTACT, "id": "c"}],
            "hasMore": False,
            "pageNumber": 1,
            "pageSize": 20,
        },
    )
    async with _make_async() as c:
        ids = [contact.id async for contact in c.contacts.list_auto_paginate()]
    assert ids == ["a", "b", "c"]


@pytest.mark.asyncio
async def test_async_list_activity_auto_paginate(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_1/activity?pageNumber=0",
        json={
            "data": [
                {
                    "_id": "act_a",
                    "vendor_id": "hst_1",
                    "email": "alex@example.com",
                    "type": "email_sent",
                    "data": {},
                    "createdAt": "2026-05-01T00:00:00.000Z",
                    "updatedAt": "2026-05-01T00:00:00.000Z",
                }
            ],
            "hasMore": False,
            "pageNumber": 0,
            "pageSize": 20,
        },
    )
    async with _make_async() as c:
        ids = [act.id_ async for act in c.contacts.list_activity_auto_paginate("cnt_1")]
    assert ids == ["act_a"]


@pytest.mark.asyncio
async def test_async_list_activity_auto_paginate_validates_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            c.contacts.list_activity_auto_paginate("")


# ---------------
# Sync payment methods
# ---------------


def test_retrieve_payment_method_happy(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/payment-methods",
        json={"data": SAMPLE_PAYMENT_METHOD},
    )
    with _make_sync() as c:
        method = c.contacts.retrieve_payment_method("cnt_123")
    assert method is not None
    assert method.id == "pm_123"
    assert method.contact_id == "cnt_123"
    assert method.card.brand == "visa"
    assert method.card.last4 == "4242"
    assert method.card.exp_month == 12
    assert method.status == "active"
    assert method.billing_details is not None
    assert method.billing_details.postal_code == "94107"


def test_retrieve_payment_method_none(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_nocard/payment-methods",
        json={"data": None},
    )
    with _make_sync() as c:
        method = c.contacts.retrieve_payment_method("cnt_nocard")
    assert method is None


def test_retrieve_payment_method_requires_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.retrieve_payment_method("")
    assert exc.value.code == "missing_id"


def test_retrieve_payment_method_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_missing/payment-methods",
        status_code=404,
        json={"error": {"code": "not_found", "message": "missing"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.contacts.retrieve_payment_method("cnt_missing")


def test_attach_payment_method_sends_body(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/payment-methods",
        method="POST",
        match_json={"setupIntentId": "seti_123"},
        json={"data": SAMPLE_PAYMENT_METHOD, "replacedExisting": True},
    )
    with _make_sync() as c:
        result = c.contacts.attach_payment_method(
            "cnt_123", AttachPaymentMethodBody(setup_intent_id="seti_123")
        )
    assert result.replaced_existing is True
    assert result.data.id == "pm_123"
    assert result.data.card.last4 == "4242"


def test_attach_payment_method_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.attach_payment_method("", AttachPaymentMethodBody(setup_intent_id="seti_123"))
    assert exc.value.code == "missing_id"


def test_attach_payment_method_validates_body() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.attach_payment_method("cnt_123", None)  # type: ignore[arg-type]
    assert exc.value.code == "missing_body"


def test_attach_payment_method_400_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/payment-methods",
        method="POST",
        status_code=400,
        json={
            "error": {
                "code": "invalid_setup_intent",
                "message": "setup intent has not been confirmed",
                "details": {"field": "setupIntentId"},
            }
        },
    )
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.attach_payment_method(
            "cnt_123", AttachPaymentMethodBody(setup_intent_id="seti_unconfirmed")
        )
    assert exc.value.code == "invalid_setup_intent"


def test_create_payment_method_setup_intent_happy(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/payment-methods/setup-intent",
        method="POST",
        json={
            "data": {
                "setupIntentId": "seti_123",
                "clientSecret": "seti_123_secret_abc",
                "customerId": "cus_123",
            }
        },
    )
    with _make_sync() as c:
        intent = c.contacts.create_payment_method_setup_intent("cnt_123")
    assert intent.setup_intent_id == "seti_123"
    assert intent.client_secret == "seti_123_secret_abc"  # noqa: S105
    assert intent.customer_id == "cus_123"
    # No request body, so content-type must be absent.
    request = httpx_mock.get_requests()[0]
    assert not request.content
    assert "content-type" not in {k.lower() for k in request.headers}


def test_create_payment_method_setup_intent_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.create_payment_method_setup_intent("")
    assert exc.value.code == "missing_id"


def test_remove_payment_method_happy(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/payment-methods/pm_123",
        method="DELETE",
        json={"data": {"removed": True}},
    )
    with _make_sync() as c:
        result = c.contacts.remove_payment_method("cnt_123", "pm_123")
    assert result.removed is True


def test_remove_payment_method_validates_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.remove_payment_method("", "pm_123")
    assert exc.value.code == "missing_id"


def test_remove_payment_method_validates_method_id() -> None:
    with _make_sync() as c, pytest.raises(ValidationError) as exc:
        c.contacts.remove_payment_method("cnt_123", "")
    assert exc.value.code == "missing_method_id"


def test_remove_payment_method_404_surfaces(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/payment-methods/pm_missing",
        method="DELETE",
        status_code=404,
        json={"error": {"code": "not_found", "message": "gone"}},
    )
    with _make_sync() as c, pytest.raises(NotFoundError):
        c.contacts.remove_payment_method("cnt_123", "pm_missing")


# ---------------
# Async payment methods
# ---------------


@pytest.mark.asyncio
async def test_async_retrieve_payment_method(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/payment-methods",
        json={"data": SAMPLE_PAYMENT_METHOD},
    )
    async with _make_async() as c:
        method = await c.contacts.retrieve_payment_method("cnt_123")
    assert method is not None
    assert method.id == "pm_123"


@pytest.mark.asyncio
async def test_async_retrieve_payment_method_none(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_nocard/payment-methods",
        json={"data": None},
    )
    async with _make_async() as c:
        method = await c.contacts.retrieve_payment_method("cnt_nocard")
    assert method is None


@pytest.mark.asyncio
async def test_async_attach_payment_method(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/payment-methods",
        method="POST",
        json={"data": SAMPLE_PAYMENT_METHOD, "replacedExisting": False},
    )
    async with _make_async() as c:
        result = await c.contacts.attach_payment_method(
            "cnt_123", AttachPaymentMethodBody(setup_intent_id="seti_123")
        )
    assert result.replaced_existing is False
    assert result.data.id == "pm_123"


@pytest.mark.asyncio
async def test_async_attach_payment_method_validates_body() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError):
            await c.contacts.attach_payment_method("cnt_123", None)  # type: ignore[arg-type]


@pytest.mark.asyncio
async def test_async_create_payment_method_setup_intent(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/payment-methods/setup-intent",
        method="POST",
        json={
            "data": {
                "setupIntentId": "seti_123",
                "clientSecret": "seti_123_secret_abc",
                "customerId": "cus_123",
            }
        },
    )
    async with _make_async() as c:
        intent = await c.contacts.create_payment_method_setup_intent("cnt_123")
    assert intent.setup_intent_id == "seti_123"
    request = httpx_mock.get_requests()[0]
    assert not request.content


@pytest.mark.asyncio
async def test_async_remove_payment_method(httpx_mock: HTTPXMock) -> None:
    httpx_mock.add_response(
        url="http://test.local/v1/contacts/cnt_123/payment-methods/pm_123",
        method="DELETE",
        json={"data": {"removed": True}},
    )
    async with _make_async() as c:
        result = await c.contacts.remove_payment_method("cnt_123", "pm_123")
    assert result.removed is True


@pytest.mark.asyncio
async def test_async_remove_payment_method_validates_method_id() -> None:
    async with _make_async() as c:
        with pytest.raises(ValidationError) as exc:
            await c.contacts.remove_payment_method("cnt_123", "")
    assert exc.value.code == "missing_method_id"
