// Run with: go run ./examples/events/list
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/events"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	pageSize := 50
	result, err := api.Events.List(context.Background(), &events.ListParams{
		Status:   events.StatusOpen,
		PageSize: &pageSize,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("got %d events (hasMore=%v)\n\n", len(result.Data), result.HasMore)
	for i, ev := range result.Data {
		printEvent(i+1, ev)
	}
}

func printEvent(n int, ev events.Event) {
	fmt.Printf("[%d] %s\n", n, ev.ID)
	fmt.Printf("name:          %s\n", ev.Name)
	fmt.Printf("type:          %s\n", ev.Type)
	fmt.Printf("schedule:      %s\n", ev.Schedule)
	fmt.Printf("start:         %s\n", ev.Start)
	fmt.Printf("status:        %s\n", ev.Status)
	fmt.Printf("currency:      %s\n", ev.Currency)
	fmt.Printf("itemsSold:     %s\n", formatInt64(ev.ItemsSold))
	fmt.Printf("revenueCents:  %s\n", formatInt64(ev.RevenueCents))
	fmt.Printf("minPriceCents: %s\n", formatInt64(ev.MinPriceCents))
	fmt.Printf("maxPriceCents: %s\n", formatInt64(ev.MaxPriceCents))
	fmt.Printf("isPublic:      %s\n", formatBool(ev.IsPublic))
	fmt.Printf("isVirtual:     %s\n\n", formatBool(ev.IsVirtual))
}

// formatInt64 prints "<unset>" when the field wasn't returned by the server,
// or the integer value otherwise. The API marks pointer-typed fields as
// optional in list responses.
func formatInt64(p *int64) string {
	if p == nil {
		return "<unset>"
	}
	return fmt.Sprintf("%d", *p)
}

func formatBool(p *bool) string {
	if p == nil {
		return "<unset>"
	}
	return fmt.Sprintf("%t", *p)
}
