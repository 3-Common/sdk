// Run with: go run ./examples/invoices/record_payment
//
// Records a manual payment against an open invoice. The IdempotencyKey makes
// the request safe to replay — recording the same payment twice with the same
// key is a no-op.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

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

	updated, err := api.Invoices.RecordPayment(context.Background(), "inv_replace_with_real_id", &invoices.PaymentParams{
		Payment:        50_000, // $500.00 in cents
		IdempotencyKey: "pmt-" + time.Now().UTC().Format(time.RFC3339Nano),
		Note:           "Wire transfer, ref ABCD-1234",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("invoice %s now %s\n", updated.ID, updated.Status)
	paid, due := int64(0), int64(0)
	if updated.AmountPaid != nil {
		paid = *updated.AmountPaid
	}
	if updated.AmountDue != nil {
		due = *updated.AmountDue
	}
	fmt.Printf("  paid: %d, due: %d\n", paid, due)
}
