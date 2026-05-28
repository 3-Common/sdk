// Run with: go run ./examples/subscriptions/create
//
// Creates a new subscription with a 14-day trial. The subscription starts in
// trialing and transitions to active once the first payment succeeds.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/subscriptions"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	sub, err := api.Subscriptions.Create(ctx, &subscriptions.CreateParams{
		ContactID:  "cnt_replace_with_real_id",
		PriceID:    "price_replace_with_real_id",
		Quantity:   threecommon.Int64(1),
		TrialDays:  threecommon.Int(14),
		AutoCharge: threecommon.Bool(true),
		Notes:      "Pro plan — annual billing",
		Metadata:   map[string]string{"source": "website-checkout"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("created %s [%s]\n", sub.ID, sub.Status)
	fmt.Printf("  trial ends   %s\n", sub.TrialEnd)
	fmt.Printf("  first bill   %s\n", sub.CurrentPeriodEnd)
}
