// Run with: go run ./examples/contacts/retrieve_payment_method
//
// Retrieves the saved card on file for a contact, or nil when none is saved.
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

	method, err := api.Contacts.RetrievePaymentMethod(context.Background(), "cnt_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	if method == nil {
		fmt.Println("no card on file")
		return
	}

	fmt.Printf("%s ****%s\n", method.Card.Brand, method.Card.Last4)
	fmt.Printf("  expires: %d/%d\n", method.Card.ExpMonth, method.Card.ExpYear)
	fmt.Printf("  status:  %s\n", method.Status)
}
