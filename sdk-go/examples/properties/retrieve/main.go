// Run with: go run ./examples/properties/retrieve
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

	property, err := api.Properties.Retrieve(context.Background(), "prop_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s - %s\n", property.ID, property.Name)
	fmt.Printf("  type:       %s\n", property.Type)
	fmt.Printf("  objectType: %s\n", property.ObjectType)
	fmt.Printf("  status:     %s\n", property.Status)
	for _, opt := range property.Options {
		fmt.Printf("  option:     %s = %s\n", opt.Label, opt.Value)
	}
}
