// Run with: go run ./examples/prices/list
//
// Lists a product's active prices.
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

	pageSize := 25
	result, err := api.Prices.List(context.Background(), &prices.ListParams{
		ProductID: "prod_replace_with_real_id",
		Active:    threecommon.Bool(true),
		PageSize:  &pageSize,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("got %d prices (hasMore=%v)\n", len(result.Data), result.HasMore)
	for _, price := range result.Data {
		amount := int64(0)
		if price.UnitAmount != nil {
			amount = *price.UnitAmount
		}
		fmt.Printf("%s — %s — %d %s\n", price.ID, price.Type, amount, price.Currency)
	}
}
