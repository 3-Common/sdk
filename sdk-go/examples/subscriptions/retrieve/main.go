// Run with: go run ./examples/subscriptions/retrieve
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

	sub, err := api.Subscriptions.Retrieve(context.Background(), "sub_replace_with_real_id", nil)
	if err != nil {
		log.Fatal(err)
	}

	quantity := int64(0)
	if sub.Quantity != nil {
		quantity = *sub.Quantity
	}
	cancelAtPeriodEnd := false
	if sub.CancelAtPeriodEnd != nil {
		cancelAtPeriodEnd = *sub.CancelAtPeriodEnd
	}
	autoCharge := false
	if sub.AutoCharge != nil {
		autoCharge = *sub.AutoCharge
	}

	fmt.Printf("subscription %s [%s]\n", sub.ID, sub.Status)
	fmt.Printf("  price          %s x %d\n", sub.PriceID, quantity)
	fmt.Printf("  current period %s -> %s\n", sub.CurrentPeriodStart, sub.CurrentPeriodEnd)
	fmt.Printf("  cancelAtPeriodEnd: %v\n", cancelAtPeriodEnd)
	fmt.Printf("  autoCharge: %v\n", autoCharge)
}
