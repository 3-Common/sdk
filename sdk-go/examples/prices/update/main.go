// Run with: go run ./examples/prices/update
//
// Updates a price's amount and nickname. To change type, currency, or product,
// archive the price and create a new one instead.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/prices"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	price, err := api.Prices.Update(context.Background(), "price_replace_with_real_id", &prices.UpdateParams{
		UnitAmount: threecommon.Int64(1200),
		Nickname:   threecommon.String("Pro monthly (promo)"),
	})
	if err != nil {
		log.Fatal(err)
	}

	amount := int64(0)
	if price.UnitAmount != nil {
		amount = *price.UnitAmount
	}
	fmt.Printf("updated %s — now %d %s\n", price.ID, amount, price.Currency)
}
