// Package threecommon is the official Go client for the 3Common Public API.
//
// # Layout
//
// The package layout mirrors the Node SDK so behavior changes can be paired
// across languages:
//
//   - threecommon (this package) — Config, typed errors, helpers, version constants
//   - [github.com/3-Common/sdk/sdk-go/client]            — recommended entry point: [client.New] returns a *client.API
//   - [github.com/3-Common/sdk/sdk-go/resources/events]  — events resource (client + types)
//   - [github.com/3-Common/sdk/sdk-go/filters]           — typed builder for the API's filters query parameter
//   - [github.com/3-Common/sdk/sdk-go/pagination]        — generic Iter[T] used by every list endpoint
//   - [github.com/3-Common/sdk/sdk-go/internal/core]     — HTTP transport machinery (not user-importable)
//   - [github.com/3-Common/sdk/sdk-go/generated]         — oapi-codegen output from ../openapi/spec.yaml
//
// # File map of this package
//
//   - config.go        — Config, Logger, RetryDelay, DefaultRetryDelay
//   - errors_base.go   — *APIError + Error() + Unwrap()
//   - errors_types.go  — AuthError, PermissionError, NotFoundError, ValidationError,
//     ConflictError, RateLimitError, ServerError, ConnectionError
//   - helpers.go       — String / Int / Int64 / Bool / Float64 pointer helpers
//   - version.go       — Version (the SDK package version)
//   - api_version.go   — APIVersion + APIPath
//
// # Quick start
//
// Construct a client and call any resource method:
//
//	import (
//		"context"
//		"log"
//		"os"
//
//		threecommon "github.com/3-Common/sdk/sdk-go"
//		"github.com/3-Common/sdk/sdk-go/client"
//		"github.com/3-Common/sdk/sdk-go/resources/events"
//	)
//
//	api, err := client.New(threecommon.Config{APIKey: os.Getenv("THREECOMMON_API_KEY")})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	result, err := api.Events.List(context.Background(), &events.ListParams{
//		Status:   events.StatusOpen,
//		PageSize: threecommon.Int(50),
//	})
//
// # Errors
//
// Every error returned by the SDK is or wraps a [*APIError]. Branch on the
// typed subtypes via [errors.As]:
//
//	var notFound *threecommon.NotFoundError
//	if errors.As(err, &notFound) {
//		// 404 — notFound.RequestID, notFound.Code, etc.
//	}
//
// # Pagination
//
// Each list endpoint returns a single page plus an iterator for
// auto-pagination. The iterator type lives in [github.com/3-Common/sdk/sdk-go/pagination]:
//
//	iter := api.Events.ListAutoPaginate(ctx, nil)
//	for iter.Next() {
//		ev := iter.Current()
//		_ = ev
//	}
//	if err := iter.Err(); err != nil {
//		log.Fatal(err)
//	}
//
// On Go 1.23+ the iterator also exposes a range-over-func variant:
//
//	for ev, err := range api.Events.ListAutoPaginate(ctx, nil).All() {
//		if err != nil {
//			log.Fatal(err)
//		}
//		_ = ev
//	}
//
// # Versioning
//
// The SDK follows SemVer. The pinned API version (sent as the
// Threecommon-Version header) is independent — the API can evolve without
// breaking already-deployed SDKs. Override it via [Config.APIVersion].
package threecommon
