// Run with: go run ./examples/properties/update
//
// Only the fields set on UpdateParams are modified. Type and objectType
// cannot be changed on an existing property. Set ClearDescription to remove
// the description entirely.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/properties"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	updated, err := api.Properties.Update(context.Background(), "prop_replace_with_real_id", &properties.UpdateParams{
		Name:             "Allergies",
		ClearDescription: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("updated %s - %s (%s)\n", updated.ID, updated.Name, updated.Status)
}
