// Run with: go run ./examples/contacts/create_payment_method_setup_intent
//
// Starts saving a card for a contact. Returns a Stripe SetupIntent clientSecret
// to confirm client-side with Stripe Elements, after which you call
// AttachPaymentMethod with the returned setupIntentId.
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

	intent, err := api.Contacts.CreatePaymentMethodSetupIntent(context.Background(), "cnt_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("setupIntentId: %s\n", intent.SetupIntentID)
	fmt.Printf("clientSecret:  %s\n", intent.ClientSecret)
	fmt.Printf("customerId:    %s\n", intent.CustomerID)
}
