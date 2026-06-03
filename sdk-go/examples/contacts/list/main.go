// Run with: go run ./examples/contacts/list
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

	pageSize := 50
	result, err := api.Contacts.List(context.Background(), &contacts.ListParams{
		Filter:        contacts.QuickFilterOptedIn,
		PageSize:      &pageSize,
		SortField:     "mostRecentOrder",
		SortDirection: "desc",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("got %d contacts (hasMore=%v, page=%d)\n\n",
		len(result.Data), result.HasMore, result.PageNumber)
	for _, c := range result.Data {
		fmt.Printf("  %s — %s (%s)\n", c.ID, c.Email, c.Status)
	}
}
