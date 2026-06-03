// Run with: go run ./examples/contacts/list_activity
//
// Fetch the activity feed (checkouts, refunds, scans, emails, invoice
// payments) for a single contact.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/contacts"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	pageSize := 20
	result, err := api.Contacts.ListActivity(context.Background(), "cnt_replace_with_real_id", &contacts.ActivityListParams{
		PageSize: &pageSize,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("got %d activity records\n", len(result.Data))
	for _, event := range result.Data {
		fmt.Printf("  %s — %s\n", event.CreatedAt, event.Type)
	}
}
