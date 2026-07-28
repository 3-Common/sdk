// Run with: go run ./examples/invoices/send
//
// Emails a customer their invoice. An open (or payment_failed) invoice gets a
// payment-link email; a paid invoice gets its receipt. Send is re-callable, so
// it doubles as a resend — draft and void invoices are rejected with a 409.
//
// You can also email as part of finalizing by passing FinalizeParams with
// SendEmail set, shown here as the one-step alternative.
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

	// One step: finalize and email the payment link in a single call.
	sendEmail := true
	issued, err := api.Invoices.Finalize(ctx, "inv_replace_with_real_id", &invoices.FinalizeParams{
		SendEmail: &sendEmail,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("finalized %s [%s]\n", issued.ID, issued.Status)

	// Or send (or re-send) the email for an already-finalized invoice.
	sent, err := api.Invoices.Send(ctx, "inv_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sent invoice %s [%s]\n", sent.ID, sent.Status)
}
