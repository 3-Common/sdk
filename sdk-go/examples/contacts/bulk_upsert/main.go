// Run with: go run ./examples/contacts/bulk_upsert
//
// Bulk-upsert contacts (e.g. from a CSV import). Deduplicated server-side
// by email; existing rows are updated rather than rejected.
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

	result, err := api.Contacts.BulkUpsert(context.Background(), &contacts.BulkUpsertParams{
		Contacts: []contacts.BulkUpsertItem{
			{Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace"},
			{Email: "beatrix@example.com", FirstName: "Beatrix", LastName: "Potter"},
			{Email: "charles@example.com", FirstName: "Charles", LastName: "Babbage"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("upserted %d contacts\n", result.Affected)
}
