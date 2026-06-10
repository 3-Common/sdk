// Run with: go run ./examples/prices/archive
//
// Soft-archives a price. Existing subscriptions keep billing; new subscriptions
// can no longer select it until unarchived. Idempotent.
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

	price, err := api.Prices.Archive(context.Background(), "price_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	active := false
	if price.Active != nil {
		active = *price.Active
	}
	fmt.Printf("archived %s — active=%v\n", price.ID, active)
}
