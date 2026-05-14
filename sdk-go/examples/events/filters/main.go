// Run with: go run ./examples/events/filters
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/filters"
	"github.com/3-Common/sdk/sdk-go/resources/events"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Build a typed filter.
	f := filters.And(
		filters.Field("status").IsAnyOf("open"),
		filters.Field("ticketSum").IsGreaterThan(10),
		filters.Or(
			filters.Field("type").IsEqualTo("event"),
			filters.Field("type").IsEqualTo("class"),
		).Group,
	)

	params := (&events.ListParams{}).FilterWith(f)

	result, err := api.Events.List(context.Background(), params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("matched %d events\n", len(result.Data))
	for _, ev := range result.Data {
		fmt.Printf("  %s — %s [tickets sold: %d]\n", ev.ID, ev.Name, deref(ev.ItemsSold))
	}
}

func deref(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
