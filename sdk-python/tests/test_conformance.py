"""Conformance harness — runs the shared YAML scenarios against the Python SDK.

Every other SDK in this monorepo runs the same scenarios; identical pass/fail
across languages is the contract.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import pytest
import yaml
from pytest_httpx import HTTPXMock

from _conformance import (
    dispatch_contacts,
    dispatch_entitlements,
    dispatch_events,
    dispatch_features,
    dispatch_invoices,
    dispatch_prices,
    dispatch_properties,
    dispatch_subscriptions,
)
from threecommon import (
    APIError,
    AsyncThreeCommon,
    AuthError,
    ConflictError,
    NotFoundError,
    PermissionError,
    RateLimitError,
    ServerError,
    ThreeCommon,
    ValidationError,
)

SCENARIOS_DIR = Path(__file__).resolve().parents[2] / "conformance" / "scenarios"


@dataclass
class _Scenario:
    path: str
    name: str
    raw: dict[str, Any] = field(default_factory=dict)


def _load_scenarios() -> list[_Scenario]:
    if not SCENARIOS_DIR.exists():
        msg = f"conformance scenarios dir not found: {SCENARIOS_DIR}"
        raise FileNotFoundError(msg)
    out: list[_Scenario] = []
    for p in sorted(SCENARIOS_DIR.rglob("*.yaml")):
        # Relative path (e.g. "events/list-happy.yaml") makes it obvious which
        # resource the scenario targets when scrolling test output.
        rel = p.relative_to(SCENARIOS_DIR).as_posix()
        raw = yaml.safe_load(p.read_text())
        out.append(_Scenario(path=rel, name=raw.get("name", rel), raw=raw))
    return out


SCENARIOS = _load_scenarios()


_TYPED_ERRORS: dict[str, type[APIError]] = {
    "ThreeCommonAuthError": AuthError,
    "ThreeCommonPermissionError": PermissionError,
    "ThreeCommonNotFoundError": NotFoundError,
    "ThreeCommonValidationError": ValidationError,
    "ThreeCommonConflictError": ConflictError,
    "ThreeCommonRateLimitError": RateLimitError,
    "ThreeCommonServerError": ServerError,
}


def _stage_exchanges(httpx_mock: HTTPXMock, scenario: dict[str, Any]) -> list[dict[str, Any]]:
    """Queue responses on pytest-httpx and return the expected-request shapes
    so the caller can assert against them after the call.

    """
    exchanges = scenario.get("exchanges")
    if not exchanges:
        ex_req = scenario.get("expectedRequest", {})
        mock = scenario.get("mockResponse")
        if mock is None:
            return []
        exchanges = [{"request": ex_req, "response": mock}]

    expected_requests: list[dict[str, Any]] = []
    for ex in exchanges:
        req = ex.get("request", {})
        resp = ex.get("response", {})
        kwargs: dict[str, Any] = {
            "method": req.get("method", "GET"),
            "status_code": resp.get("status", 200),
        }
        if resp.get("headers"):
            kwargs["headers"] = resp["headers"]
        body = resp.get("body")
        if body is not None:
            kwargs["json"] = body
        httpx_mock.add_response(**kwargs)
        expected_requests.append(req)
    return expected_requests


def _assert_request(want: dict[str, Any], actual_requests: list[Any], idx: int) -> None:
    """Verify the idx-th captured request matches the expected shape."""
    assert idx < len(actual_requests), f"missing request #{idx}"
    actual = actual_requests[idx]
    if "method" in want:
        assert actual.method == want["method"], (
            f"request[{idx}].method: want {want['method']}, got {actual.method}"
        )
    if "path" in want:
        assert actual.url.path == want["path"], (
            f"request[{idx}].path: want {want['path']}, got {actual.url.path}"
        )
    if "query" in want:
        params = dict(actual.url.params)
        for k, v in want["query"].items():
            assert params.get(k) == str(v), (
                f"request[{idx}].query[{k}]: want {v}, got {params.get(k)}"
            )
    if "headers" in want:
        for k, v in want["headers"].items():
            assert actual.headers.get(k) == v, (
                f"request[{idx}].headers[{k}]: want {v}, got {actual.headers.get(k)}"
            )
    for absent in want.get("headersAbsent", []):
        assert absent.lower() not in {h.lower() for h in actual.headers}, (
            f"request[{idx}].headers[{absent}] should be absent"
        )


def _dispatch_sync(client: ThreeCommon, call: dict[str, Any]) -> Any:  # noqa: ANN401, PLR0911
    """Route a scenario call to the per-resource sync dispatcher.

    Each resource lives in its own module under ``tests/_conformance/``; add a
    sibling branch here when introducing a new resource.
    """
    resource = call.get("resource", "events")
    method = call["method"]
    args = call.get("args", {}) or {}
    if resource == "events":
        return dispatch_events.dispatch_sync(client, method, args)
    if resource == "invoices":
        return dispatch_invoices.dispatch_sync(client, method, args)
    if resource == "subscriptions":
        return dispatch_subscriptions.dispatch_sync(client, method, args)
    if resource == "contacts":
        return dispatch_contacts.dispatch_sync(client, method, args)
    if resource == "entitlements":
        return dispatch_entitlements.dispatch_sync(client, method, args)
    if resource == "prices":
        return dispatch_prices.dispatch_sync(client, method, args)
    if resource == "features":
        return dispatch_features.dispatch_sync(client, method, args)
    if resource == "properties":
        return dispatch_properties.dispatch_sync(client, method, args)
    pytest.fail(f"unsupported scenario resource: {resource!r}")


def _assert_subset(want: Any, got: Any, prefix: str) -> None:  # noqa: ANN401
    """Recursive subset match — every key in want must appear in got with same value."""
    if isinstance(want, dict):
        if not isinstance(got, dict):
            pytest.fail(f"{prefix}: expected dict, got {type(got).__name__}")
        for k, v in want.items():
            assert k in got, f"{prefix}.{k}: missing"
            _assert_subset(v, got[k], f"{prefix}.{k}")
    elif isinstance(want, list):
        if not isinstance(got, list):
            pytest.fail(f"{prefix}: expected list, got {type(got).__name__}")
        assert len(want) == len(got), f"{prefix}: length mismatch ({len(want)} vs {len(got)})"
        for i, (w, g) in enumerate(zip(want, got, strict=False)):
            _assert_subset(w, g, f"{prefix}[{i}]")
    else:
        assert want == got, f"{prefix}: {want!r} != {got!r}"


def _assert_result(want: Any, got: Any) -> None:  # noqa: ANN401
    """Validate scenario.expectedResult — convert pydantic models to dicts."""
    if hasattr(got, "model_dump"):
        got = got.model_dump(by_alias=True, exclude_none=False)
    elif isinstance(got, list):
        got = [
            g.model_dump(by_alias=True, exclude_none=False) if hasattr(g, "model_dump") else g
            for g in got
        ]
    _assert_subset(want, got, "result")


def _assert_error(want: dict[str, Any], err: BaseException) -> None:
    cls = _TYPED_ERRORS.get(want["type"])
    assert cls is not None, f"unsupported expectedError.type {want['type']!r}"
    assert isinstance(err, cls), f"expected {cls.__name__}, got {type(err).__name__}"
    api = err if isinstance(err, APIError) else None
    assert api is not None
    if "code" in want:
        assert api.code == want["code"]
    if "httpStatus" in want:
        assert api.http_status == want["httpStatus"]
    if "requestId" in want:
        assert api.request_id == want["requestId"]
    if "retryAfterSeconds" in want and isinstance(api, RateLimitError):
        assert api.retry_after_seconds == pytest.approx(want["retryAfterSeconds"])
    if "details" in want and api.details is not None:
        _assert_subset(want["details"], api.details, "details")


@pytest.mark.parametrize("scenario", SCENARIOS, ids=lambda s: s.path)
def test_conformance_sync(scenario: _Scenario, httpx_mock: HTTPXMock) -> None:
    """Run every conformance scenario against the sync client."""
    body = scenario.raw
    expected_requests = _stage_exchanges(httpx_mock, body)

    client_overrides = body.get("client", {}) or {}
    client_kwargs: dict[str, Any] = {
        "api_key": "3co_test",
        "base_url": "http://test.local",
    }
    if "maxRetries" in client_overrides:
        client_kwargs["max_retries"] = client_overrides["maxRetries"]
    if "apiVersion" in client_overrides:
        client_kwargs["api_version"] = client_overrides["apiVersion"]
    if "telemetry" in client_overrides:
        client_kwargs["telemetry"] = client_overrides["telemetry"]

    client = ThreeCommon(**client_kwargs)
    try:
        if "expectedError" in body:
            with pytest.raises(APIError) as exc_info:
                _dispatch_sync(client, body["call"])
            _assert_error(body["expectedError"], exc_info.value)
        elif "expectedAutoPaginated" in body:
            result = _dispatch_sync(client, body["call"])
            _assert_result(body["expectedAutoPaginated"], result)
        elif "expectedResult" in body:
            result = _dispatch_sync(client, body["call"])
            _assert_result(body["expectedResult"], result)
        else:
            _dispatch_sync(client, body["call"])  # smoke
    finally:
        client.close()

    actual = httpx_mock.get_requests()
    if "expectedCallCount" in body:
        assert len(actual) == body["expectedCallCount"], (
            f"expected {body['expectedCallCount']} calls, got {len(actual)}"
        )
    for i, want in enumerate(expected_requests):
        if i < len(actual):
            _assert_request(want, actual, i)


# ────────────────────────────────────────────────────────────────────────────
# Async parity — re-run a subset against AsyncThreeCommon
# ────────────────────────────────────────────────────────────────────────────


_ASYNC_SCENARIOS = SCENARIOS


async def _dispatch_async(client: AsyncThreeCommon, call: dict[str, Any]) -> Any:  # noqa: ANN401, PLR0911
    """Route a scenario call to the per-resource async dispatcher."""
    resource = call.get("resource", "events")
    method = call["method"]
    args = call.get("args", {}) or {}
    if resource == "events":
        return await dispatch_events.dispatch_async(client, method, args)
    if resource == "invoices":
        return await dispatch_invoices.dispatch_async(client, method, args)
    if resource == "subscriptions":
        return await dispatch_subscriptions.dispatch_async(client, method, args)
    if resource == "contacts":
        return await dispatch_contacts.dispatch_async(client, method, args)
    if resource == "entitlements":
        return await dispatch_entitlements.dispatch_async(client, method, args)
    if resource == "prices":
        return await dispatch_prices.dispatch_async(client, method, args)
    if resource == "features":
        return await dispatch_features.dispatch_async(client, method, args)
    if resource == "properties":
        return await dispatch_properties.dispatch_async(client, method, args)
    pytest.fail(f"unsupported scenario resource: {resource!r}")


@pytest.mark.parametrize("scenario", _ASYNC_SCENARIOS, ids=lambda s: f"async-{s.path}")
@pytest.mark.asyncio
async def test_conformance_async(scenario: _Scenario, httpx_mock: HTTPXMock) -> None:
    """Re-run every conformance scenario against the async client to verify parity."""
    body = scenario.raw
    expected_requests = _stage_exchanges(httpx_mock, body)

    client_overrides = body.get("client", {}) or {}
    client_kwargs: dict[str, Any] = {
        "api_key": "3co_test",
        "base_url": "http://test.local",
    }
    if "maxRetries" in client_overrides:
        client_kwargs["max_retries"] = client_overrides["maxRetries"]
    if "apiVersion" in client_overrides:
        client_kwargs["api_version"] = client_overrides["apiVersion"]
    if "telemetry" in client_overrides:
        client_kwargs["telemetry"] = client_overrides["telemetry"]

    client = AsyncThreeCommon(**client_kwargs)
    try:
        if "expectedError" in body:
            with pytest.raises(APIError) as exc_info:
                await _dispatch_async(client, body["call"])
            _assert_error(body["expectedError"], exc_info.value)
        elif "expectedAutoPaginated" in body:
            result = await _dispatch_async(client, body["call"])
            _assert_result(body["expectedAutoPaginated"], result)
        elif "expectedResult" in body:
            result = await _dispatch_async(client, body["call"])
            _assert_result(body["expectedResult"], result)
        else:
            await _dispatch_async(client, body["call"])  # smoke
    finally:
        await client.aclose()

    actual = httpx_mock.get_requests()
    if "expectedCallCount" in body:
        assert len(actual) == body["expectedCallCount"], (
            f"expected {body['expectedCallCount']} calls, got {len(actual)}"
        )
    for i, want in enumerate(expected_requests):
        if i < len(actual):
            _assert_request(want, actual, i)
