// Run with: go run ./examples/contacts/retrieve
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

	contact, err := api.Contacts.Retrieve(context.Background(), "cnt_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s <%s>\n", contact.FullName, contact.Email)
	fmt.Printf("  status:    %s\n", contact.Status)
	fmt.Printf("  orders:    %d\n", contact.OrderSum)
	fmt.Printf("  gross:     %d\n", contact.GrossSum)
	fmt.Printf("  vendorId:  %s\n", contact.VendorID)
}
