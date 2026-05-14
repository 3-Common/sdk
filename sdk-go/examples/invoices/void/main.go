// Run with: go run ./examples/invoices/void
//
// Voids an invoice. Permitted from draft or open. Paid invoices cannot be
// voided — issue a credit note or refund the payment instead.
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

	voided, err := api.Invoices.Void(context.Background(), "inv_replace_with_real_id", &invoices.VoidParams{
		Reason: "Sent to the wrong customer",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("invoice %s status: %s\n", voided.ID, voided.Status)
}
