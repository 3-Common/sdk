"""Per-resource dispatchers for the conformance harness.

Each module here exposes a sync and async dispatcher; the runner
([tests/test_conformance.py]) routes scenarios to the right one based on
``call.resource``. Add a sibling module when introducing a new resource.
"""
