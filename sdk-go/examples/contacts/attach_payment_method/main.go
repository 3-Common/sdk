// Run with: go run ./examples/contacts/attach_payment_method
//
// Attaches a card to a contact from a confirmed Stripe SetupIntent. Replaces
// any existing card on file.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/contacts"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := api.Contacts.AttachPaymentMethod(
		context.Background(),
		"cnt_replace_with_real_id",
		&contacts.AttachPaymentMethodParams{SetupIntentID: "seti_replace_with_real_id"},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("saved %s ****%s (%s)\n", result.Data.Card.Brand, result.Data.Card.Last4, result.Data.ID)
	fmt.Printf("  replaced existing: %t\n", result.ReplacedExisting)
}
