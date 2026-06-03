// Run with: go run ./examples/contacts/update
//
// Returns the richer order-details projection ([contacts.WithOrderDetails]),
// not the compact [contacts.Contact] returned by Retrieve.
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

	updated, err := api.Contacts.Update(context.Background(), "cnt_replace_with_real_id", &contacts.UpdateParams{
		Contact: contacts.ContactUpdate{
			FirstName: "Alex",
			LastName:  "Garcia",
			Email:     "a.garcia@example.com",
			Status:    contacts.StatusOptedIn,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("updated %s → %s (%s)\n", updated.ID, updated.Email, updated.Status)
}
