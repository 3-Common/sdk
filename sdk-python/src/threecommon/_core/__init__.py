"""Internal HTTP transport. Not part of the public API.

Decomposed into one concern per file: URL building, header construction,
retry policy, response parsing, telemetry header. The sync/async clients in
[http_client][threecommon._core.http_client] orchestrate them.
"""
