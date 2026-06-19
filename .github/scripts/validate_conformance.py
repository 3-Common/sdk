"""Validate the structural shape of every conformance scenario.

Run from the repo root. Exits non-zero on the first malformed scenario so
the failing path is clear in the GitHub Actions log.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path
from typing import Any

import yaml

SCENARIOS_DIR = Path("conformance/scenarios")

# Per-language dispatcher locations. Each glob's basename (minus prefix/suffix)
# is the resource name.
_NODE_DISPATCHERS = Path("sdk-node/test/conformance")
_PYTHON_DISPATCHERS = Path("sdk-python/tests/_conformance")
_GO_DISPATCHERS = Path("sdk-go/conformance")
_NODE_ERRORS = Path("sdk-node/src/errors/classes.ts")

def _resource_from_stem(stem: str) -> str:
    name = stem.removeprefix("dispatch-").removeprefix("dispatch_")
    return name.removesuffix("_test")


def _derive_resources() -> set[str]:
    sources = {
        "node": list(_NODE_DISPATCHERS.glob("dispatch-*.ts")),
        "python": list(_PYTHON_DISPATCHERS.glob("dispatch_*.py")),
        "go": list(_GO_DISPATCHERS.glob("dispatch_*_test.go")),
    }
    per_lang = {lang: {_resource_from_stem(p.stem) for p in paths} for lang, paths in sources.items()}

    union: set[str] = set().union(*per_lang.values())
    for lang, names in per_lang.items():
        missing = union - names
        if missing:
            print(
                f"warning: {lang} is missing dispatchers for {sorted(missing)}",
                file=sys.stderr,
            )
    return union


_CASE_LABEL = re.compile(r"case '([a-zA-Z][a-zA-Z0-9]*)':")


def _derive_methods() -> set[str]:
    methods: set[str] = set()
    for path in _NODE_DISPATCHERS.glob("dispatch-*.ts"):
        methods.update(_CASE_LABEL.findall(path.read_text()))
    return methods


_ERROR_CLASS = re.compile(r"export class (ThreeCommon\w*Error)\b")


def _derive_error_types() -> set[str]:
    return set(_ERROR_CLASS.findall(_NODE_ERRORS.read_text()))


KNOWN_RESOURCES = _derive_resources()
KNOWN_METHODS = _derive_methods()
KNOWN_ERROR_TYPES = _derive_error_types()

errors: list[str] = []


def err(path: Path, msg: str) -> None:
    errors.append(f"{path}: {msg}")


def validate_call(path: Path, call: Any) -> None:
    if not isinstance(call, dict):
        err(path, "call must be a mapping")
        return

    resource = call.get("resource")
    if resource is not None and resource not in KNOWN_RESOURCES:
        err(
            path,
            f"call.resource = {resource!r} is not in known resources "
            f"({sorted(KNOWN_RESOURCES)}). If this is a new resource, add it "
            "to KNOWN_RESOURCES in validate_conformance.py.",
        )

    method = call.get("method")
    if not method:
        err(path, "call.method is required")
        return
    if method not in KNOWN_METHODS:
        err(
            path,
            f"call.method = {method!r} is not in known methods. If this is a "
            "new method, add it to KNOWN_METHODS.",
        )

    if "args" in call and not isinstance(call["args"], dict):
        err(path, "call.args must be a mapping when present")


def validate_expected_request(path: Path, req: Any, *, prefix: str = "expectedRequest") -> None:
    if not isinstance(req, dict):
        err(path, f"{prefix} must be a mapping")
        return
    method = req.get("method")
    if not method:
        err(path, f"{prefix}.method is required")
    elif method not in {"GET", "POST", "PATCH", "PUT", "DELETE"}:
        err(path, f"{prefix}.method = {method!r} must be a standard HTTP verb")
    if not req.get("path"):
        err(path, f"{prefix}.path is required")
    for field in ("query", "headers"):
        if field in req and not isinstance(req[field], dict):
            err(path, f"{prefix}.{field} must be a mapping when present")
    if "headersAbsent" in req and not isinstance(req["headersAbsent"], list):
        err(path, f"{prefix}.headersAbsent must be a list when present")


def validate_mock_response(path: Path, resp: Any, *, prefix: str = "mockResponse") -> None:
    if not isinstance(resp, dict):
        err(path, f"{prefix} must be a mapping")
        return
    status = resp.get("status")
    if not isinstance(status, int):
        err(path, f"{prefix}.status must be an integer (got {type(status).__name__})")
    elif not (100 <= status < 600):
        err(path, f"{prefix}.status = {status} is not in 100–599")


def validate_expected_error(path: Path, exp: Any) -> None:
    if not isinstance(exp, dict):
        err(path, "expectedError must be a mapping")
        return
    err_type = exp.get("type")
    if not err_type:
        err(path, "expectedError.type is required")
    elif err_type not in KNOWN_ERROR_TYPES:
        err(
            path,
            f"expectedError.type = {err_type!r} is not in known types "
            f"({sorted(KNOWN_ERROR_TYPES)}).",
        )
    if not exp.get("code"):
        err(path, "expectedError.code is required (a string like 'not_found')")


def validate_scenario(path: Path) -> None:
    try:
        doc = yaml.safe_load(path.read_text())
    except yaml.YAMLError as exc:
        err(path, f"YAML parse error: {exc}")
        return

    if not isinstance(doc, dict):
        err(path, "scenario root must be a mapping")
        return

    if not doc.get("name"):
        err(path, "name is required")

    if "call" not in doc:
        err(path, "call is required")
        return
    validate_call(path, doc["call"])

    has_exchanges = "exchanges" in doc
    has_single = "expectedRequest" in doc and "mockResponse" in doc

    if not has_exchanges and not has_single:
        err(
            path,
            "scenario must define EITHER (expectedRequest + mockResponse) for "
            "a single call OR exchanges[] for a multi-call sequence",
        )

    if has_exchanges:
        exchanges = doc["exchanges"]
        if not isinstance(exchanges, list) or not exchanges:
            err(path, "exchanges must be a non-empty list")
        else:
            for i, ex in enumerate(exchanges):
                if not isinstance(ex, dict):
                    err(path, f"exchanges[{i}] must be a mapping")
                    continue
                if "request" in ex:
                    validate_expected_request(
                        path, ex["request"], prefix=f"exchanges[{i}].request"
                    )
                if "response" not in ex:
                    err(path, f"exchanges[{i}].response is required")
                else:
                    validate_mock_response(
                        path, ex["response"], prefix=f"exchanges[{i}].response"
                    )

    if "expectedRequest" in doc:
        validate_expected_request(path, doc["expectedRequest"])
    if "mockResponse" in doc:
        validate_mock_response(path, doc["mockResponse"])
    if "expectedError" in doc:
        validate_expected_error(path, doc["expectedError"])
    if "expectedResultNull" in doc and doc["expectedResultNull"] is not True:
        err(path, "expectedResultNull must be `true` when present")

    if not any(
        k in doc
        for k in ("expectedResult", "expectedResultNull", "expectedAutoPaginated", "expectedError")
    ):
        err(
            path,
            "scenario must define one of expectedResult, expectedResultNull, "
            "expectedAutoPaginated, or expectedError",
        )


def main() -> int:
    if not SCENARIOS_DIR.is_dir():
        print(f"scenarios directory not found: {SCENARIOS_DIR}", file=sys.stderr)
        return 1

    files = sorted(SCENARIOS_DIR.rglob("*.yaml"))
    if not files:
        print(f"no scenarios found under {SCENARIOS_DIR}/", file=sys.stderr)
        return 1

    for path in files:
        validate_scenario(path)

    if errors:
        print(f"\n{len(errors)} conformance scenario problem(s):\n", file=sys.stderr)
        for line in errors:
            print(f"  ✗ {line}", file=sys.stderr)
        return 1

    print(f"✓ {len(files)} scenarios valid")
    return 0


if __name__ == "__main__":
    sys.exit(main())
