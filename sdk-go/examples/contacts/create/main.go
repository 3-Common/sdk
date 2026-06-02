// Run with: go run ./examples/contacts/create
//
// Returns a 409 ConflictError if a contact with the same email already
// exists for this host.
package main

import (
	"context"
	"errors"
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

	created, err := api.Contacts.Create(context.Background(), &contacts.CreateParams{
		Email:     "guest@example.com",
		FirstName: "Alex",
		LastName:  "Garcia",
	})
	if err != nil {
		var conflict *threecommon.ConflictError
		if errors.As(err, &conflict) {
			fmt.Println("contact with that email already exists for this host")
			return
		}
		log.Fatal(err)
	}

	fmt.Printf("created %s <%s>\n", created.ID, created.Email)
}
