// Run with: go run ./examples/invoices/create
//
// Creates a draft invoice and finalizes it. Finalizing assigns a sequential
// number, stamps issuedAt, and transitions the invoice to open.
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

	ctx := context.Background()
	draft, err := api.Invoices.Create(ctx, &invoices.CreateParams{
		CustomerID: "cnt_replace_with_real_id",
		Currency:   invoices.CurrencyUSD,
		LineItems: []invoices.LineItem{
			{Description: "Consulting — May 2026", Quantity: 8, UnitAmount: 12_500},
			{Description: "Onboarding fee", Quantity: 1, UnitAmount: 50_000},
		},
		Notes: "Net 30. Wire transfer preferred.",
	})
	if err != nil {
		log.Fatal(err)
	}

	total := int64(0)
	if draft.Total != nil {
		total = *draft.Total
	}
	fmt.Printf("drafted %s — total %d USD\n", draft.ID, total)

	issued, err := api.Invoices.Finalize(ctx, draft.ID)
	if err != nil {
		log.Fatal(err)
	}

	number := ""
	if issued.Number != nil {
		number = *issued.Number
	}
	fmt.Printf("finalized %s as %s [%s]\n", issued.ID, number, issued.Status)
}
