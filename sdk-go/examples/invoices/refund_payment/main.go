// Run with: go run ./examples/invoices/refund_payment
//
// Refunds all or part of a recorded payment on a paid invoice. The
// IdempotencyKey makes the request safe to replay — refunding twice with the
// same key returns the existing refund instead of issuing a second one.
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

	refunded, err := api.Invoices.RefundPayment(context.Background(), "inv_replace_with_real_id", "pay_replace_with_real_id", &invoices.RefundParams{
		Amount:         25_000, // $250.00 in cents; capped at the refundable balance
		Reason:         "requested_by_customer",
		IdempotencyKey: "rfnd-" + time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("invoice %s now %s\n", refunded.ID, refunded.Status)
}
