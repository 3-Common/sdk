// Run with: go run ./examples/subscriptions/list
//
// Lists active subscriptions for a contact.
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

	pageSize := 25
	result, err := api.Subscriptions.List(context.Background(), &subscriptions.ListParams{
		Status:    subscriptions.StatusActive,
		ContactID: "cnt_replace_with_real_id",
		PageSize:  &pageSize,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("got %d subscriptions (hasMore=%v)\n", len(result.Data), result.HasMore)
	for _, sub := range result.Data {
		fmt.Printf("%s — %s — renews %s\n", sub.ID, sub.Status, sub.CurrentPeriodEnd)
	}
}
