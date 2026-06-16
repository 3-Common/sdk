// Run with: go run ./examples/properties/create
//
// For "Select One" and "Select Multiple" properties, Options is required and
// must have at least one entry.
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

	created, err := api.Properties.Create(context.Background(), &properties.CreateParams{
		Type:       properties.TypeSelectOne,
		Name:       "T-shirt size",
		Status:     properties.StatusActive,
		ObjectType: properties.ObjectTypeContact,
		Options: []properties.Option{
			{Value: "s", Label: "Small"},
			{Value: "m", Label: "Medium"},
			{Value: "l", Label: "Large"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("created %s - %s (%s)\n", created.ID, created.Name, created.Type)
}
