// Run with: go run ./examples/invoices/delete_draft
//
// Permanently deletes a draft invoice. Only drafts can be deleted — once an
// invoice is finalized (it has a number), void it instead so the audit trail
// stays intact.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	deleted, err := api.Invoices.DeleteDraft(context.Background(), "inv_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("deleted draft invoice %s\n", deleted.ID)
}
