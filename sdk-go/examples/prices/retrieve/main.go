// Run with: go run ./examples/prices/retrieve
//
// Retrieves a single price by ID, including its recurring cadence and feature
// grants.
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

	price, err := api.Prices.Retrieve(context.Background(), "price_replace_with_real_id", nil)
	if err != nil {
		log.Fatal(err)
	}

	amount := int64(0)
	if price.UnitAmount != nil {
		amount = *price.UnitAmount
	}
	fmt.Printf("price %s [%s]\n", price.ID, price.Type)
	fmt.Printf("  product  %s\n", price.ProductID)
	fmt.Printf("  amount   %d %s\n", amount, price.Currency)
	if price.Recurring != nil {
		fmt.Printf("  cadence  every %d %s\n", price.Recurring.IntervalCount, price.Recurring.Interval)
	}
	for _, feature := range price.Features {
		fmt.Printf("  feature  %s [%s]\n", feature.FeatureKey, feature.Type)
	}
}
