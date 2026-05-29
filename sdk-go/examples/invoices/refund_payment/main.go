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
		Amount: 25_000, // $250.00 in cents; capped at the refundable balance
		Reason: "requested_by_customer",
		// Derive the idempotency key from a stable business event id (e.g. the
		// refund-request id in your own system), never the wall clock — a retry
		// must reuse the same key or it refunds a second time.
		IdempotencyKey: "rfnd-8842",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("invoice %s now %s\n", refunded.ID, refunded.Status)
}
