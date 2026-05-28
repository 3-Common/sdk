// Run with: go run ./examples/subscriptions/preview_upcoming_invoice
//
// Previews the invoice the next renewal will generate (Stripe-style
// invoice.upcoming). Returns nil when the subscription is set to cancel at
// period end.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	preview, err := api.Subscriptions.PreviewUpcomingInvoice(context.Background(), "sub_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	if preview == nil {
		fmt.Println("subscription is set to cancel at period end — no upcoming invoice")
		return
	}

	fmt.Printf("next invoice — %d %s\n", preview.Total, preview.Currency)
	fmt.Printf("  period %s -> %s\n", preview.PeriodStart, preview.PeriodEnd)
	for _, line := range preview.LineItems {
		fmt.Printf("  - %s — %d x %d\n", line.Description, line.Quantity, line.UnitAmount)
	}
}
