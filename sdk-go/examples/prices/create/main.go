// Run with: go run ./examples/prices/create
//
// Creates a recurring price with a metered feature grant. The quantity grant
// refills the customer's entitlement balance on each renewal.
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

	price, err := api.Prices.Create(context.Background(), &prices.CreateParams{
		ProductID:  "prod_replace_with_real_id",
		Type:       prices.TypeRecurring,
		Currency:   prices.CurrencyUSD,
		UnitAmount: 1500,
		Recurring:  &prices.Recurring{Interval: prices.IntervalMonth, IntervalCount: 1},
		Features: []prices.Feature{
			{
				FeatureKey:      "api_calls",
				Type:            prices.FeatureTypeQuantity,
				Quantity:        threecommon.Int64(1000),
				RolloverEnabled: threecommon.Bool(false),
			},
		},
		Nickname: "Pro monthly",
		Metadata: map[string]string{"tier": "pro"},
	})
	if err != nil {
		log.Fatal(err)
	}

	amount := int64(0)
	if price.UnitAmount != nil {
		amount = *price.UnitAmount
	}
	fmt.Printf("created %s — %d %s\n", price.ID, amount, price.Currency)
}
