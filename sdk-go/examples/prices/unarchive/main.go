// Run with: go run ./examples/prices/unarchive
//
// Reactivates a previously archived price. Idempotent.
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

	price, err := api.Prices.Unarchive(context.Background(), "price_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	active := false
	if price.Active != nil {
		active = *price.Active
	}
	fmt.Printf("unarchived %s — active=%v\n", price.ID, active)
}
