// Run with: go run ./examples/contacts/remove_payment_method
//
// Detaches a saved card from Stripe and removes it from the contact.
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

	result, err := api.Contacts.RemovePaymentMethod(
		context.Background(),
		"cnt_replace_with_real_id",
		"pm_replace_with_real_id",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("removed: %t\n", result.Removed)
}
