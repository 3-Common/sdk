// Run with: go run ./examples/subscriptions/update
//
// Applies a mid-cycle upgrade. The SDK returns the updated subscription, a
// proration summary, and (when the rate difference is positive) a slim
// reference to the proration invoice.
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
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	result, err := api.Subscriptions.Update(context.Background(), "sub_replace_with_real_id", &subscriptions.UpdateParams{
		PriceID:  "price_upgrade_replace_with_real_id",
		Quantity: threecommon.Int64(2),
	})
	if err != nil {
		log.Fatal(err)
	}

	sub := result.Subscription
	quantity := int64(0)
	if sub.Quantity != nil {
		quantity = *sub.Quantity
	}
	fmt.Printf("updated %s -> %s x %d\n", sub.ID, sub.PriceID, quantity)
	fmt.Printf("proration: %d minor units (%d/%d days)\n",
		result.Proration.NetAmountMinor, result.Proration.DaysRemaining, result.Proration.DaysInCycle)

	if result.Invoice != nil {
		inv := result.Invoice
		fmt.Printf("proration invoice %s [%s] — total %d %s\n", inv.ID, inv.Status, inv.Total, inv.Currency)
	} else {
		fmt.Println("downgrade or no-op — no proration invoice issued")
	}
}
