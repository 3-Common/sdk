// Run with: go run ./examples/invoices/auto_charge
//
// Off-session auto-charges an open invoice against the customer's saved card.
// A decline is not an error — the call returns a result with Outcome "failed"
// and a FailureCode, leaving the invoice in payment_failed. Only network /
// processor (5xx) errors return an error.
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

	result, err := api.Invoices.AutoCharge(context.Background(), "inv_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	if result.Outcome == invoices.AutoChargeOutcomePaid {
		fmt.Printf("invoice %s charged, now %s\n", result.Invoice.ID, result.Invoice.Status)
	} else {
		code := result.FailureCode
		if code == "" {
			code = "unknown"
		}
		fmt.Printf("charge failed (%s); invoice is %s\n", code, result.Invoice.Status)
	}
}
