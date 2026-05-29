// Run with: go run ./examples/invoices/update
//
// Revises a draft invoice. Only legal while in draft — once finalized, void it
// and create a new one instead so the audit trail stays intact. Replacing
// LineItems recomputes the totals server-side; only the fields you set are
// sent. The method is Update; "revise" is the domain term for editing a draft.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/invoices"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	revised, err := api.Invoices.Update(context.Background(), "inv_replace_with_real_id", &invoices.UpdateParams{
		Notes: "Net 30. Updated per customer request.",
		DueAt: "2026-07-01T00:00:00.000Z",
		LineItems: []invoices.LineItem{
			{Description: "Consulting (revised)", Quantity: 10, UnitAmount: 12_500},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("revised %s [%s]\n", revised.ID, revised.Status)
}
