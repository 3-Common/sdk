// Run with: go run ./examples/invoices/retrieve
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

	inv, err := api.Invoices.Retrieve(context.Background(), "inv_replace_with_real_id", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("invoice %s [%s]\n", inv.ID, inv.Status)
	fmt.Printf("  total:       %s %s\n", formatInt64(inv.Total), inv.Currency)
	fmt.Printf("  amountPaid:  %s\n", formatInt64(inv.AmountPaid))
	fmt.Printf("  amountDue:   %s\n", formatInt64(inv.AmountDue))
	fmt.Printf("  line items:  %d\n", len(inv.LineItems))
	fmt.Printf("  payments:    %d\n", len(inv.Payments))
}

func formatInt64(p *int64) string {
	if p == nil {
		return "<unset>"
	}
	return fmt.Sprintf("%d", *p)
}
