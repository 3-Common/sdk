// Run with: go run ./examples/invoices/list
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

	pageSize := 25
	result, err := api.Invoices.List(context.Background(), &invoices.ListParams{
		Status:     invoices.StatusOpen,
		CustomerID: "cnt_replace_with_real_id",
		PageSize:   &pageSize,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("got %d invoices (hasMore=%v)\n\n", len(result.Data), result.HasMore)
	for i, inv := range result.Data {
		fmt.Printf("[%d] %s\n", i+1, inv.ID)
		fmt.Printf("    status:     %s\n", inv.Status)
		fmt.Printf("    currency:   %s\n", inv.Currency)
		fmt.Printf("    total:      %s\n", formatInt64(inv.Total))
		fmt.Printf("    amountPaid: %s\n", formatInt64(inv.AmountPaid))
		fmt.Printf("    amountDue:  %s\n\n", formatInt64(inv.AmountDue))
	}
}

func formatInt64(p *int64) string {
	if p == nil {
		return "<unset>"
	}
	return fmt.Sprintf("%d", *p)
}
