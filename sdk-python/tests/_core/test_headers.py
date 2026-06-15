from __future__ import annotations

from threecommon._core.headers import build_headers, user_agent_suffix


def test_required_fields_populated() -> None:
    h = build_headers(
        api_key="3co_test",
        api_version="2026-04-29",
        sdk_version="0.1.0",
    )
    assert h["Authorization"] == "Bearer 3co_test"
    assert h["Threecommon-Version"] == "2026-04-29"
    assert h["Accept"] == "application/json"
    assert h["Content-Type"] == "application/json"
    assert h["User-Agent"].startswith("ThreeCommonPython/0.1.0")


def test_content_type_omitted_when_bodyless() -> None:
    h = build_headers(api_key="k", api_version="v", sdk_version="1", has_body=False)
    assert "Content-Type" not in h


def test_content_type_set_when_body_present() -> None:
    h = build_headers(api_key="k", api_version="v", sdk_version="1", has_body=True)
    assert h["Content-Type"] == "application/json"


def test_optional_fields_omitted_when_empty() -> None:
    h = build_headers(api_key="k", api_version="v", sdk_version="1")
    assert "Threecommon-Client-Telemetry" not in h
    assert "Idempotency-Key" not in h


def test_optional_fields_set_when_provided() -> None:
    h = build_headers(
        api_key="k",
        api_version="v",
        sdk_version="1",
        telemetry_header='{"lang":"python"}',
        idempotency_key="key-1",
    )
    assert h["Threecommon-Client-Telemetry"] == '{"lang":"python"}'
    assert h["Idempotency-Key"] == "key-1"


def test_user_agent_suffix_contains_python_version() -> None:
    suffix = user_agent_suffix()
    assert "Python/" in suffix


def test_user_agent_suffix_appends_extra() -> None:
    suffix = user_agent_suffix("MyApp/1.0")
    assert suffix.endswith("MyApp/1.0")
