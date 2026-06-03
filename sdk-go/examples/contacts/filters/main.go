// Run with: go run ./examples/contacts/filters
//
// Build a typed filter for the contacts list. The `filters` package is
// shared across resources — every endpoint that accepts `filters` consumes
// the same builder.
//
// The simple `Filter` enum (`opted-in`, `unknown`, ...) and the rich
// `filters` builder can be combined; the server ANDs them.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/filters"
	"github.com/3-Common/sdk/sdk-go/resources/contacts"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	// High-value opted-in contacts whose most recent order is in 2026.
	f := filters.And(
		filters.Field("status").IsAnyOf("opted-in"),
		filters.Field("grossSum").IsGreaterThan(100_000),
		filters.Or(
			filters.Field("orderSum").IsGreaterThanOrEqualTo(5),
			filters.Field("lastOrder").IsAfter("2026-01-01T00:00:00.000Z"),
		).Group,
	)

	pageSize := 25
	params := (&contacts.ListParams{
		SortField:     "grossSum",
		SortDirection: "desc",
		PageSize:      &pageSize,
	}).FilterWith(f)

	result, err := api.Contacts.List(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("matched %d contacts (hasMore=%v)\n", len(result.Data), result.HasMore)
	for _, c := range result.Data {
		fmt.Printf("  %s <%s> — gross %d\n", c.FullName, c.Email, c.GrossSum)
	}
}
