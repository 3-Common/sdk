// Run with: go run ./examples/features/create
//
// Creates a quantity feature in the catalog. The Key is the stable identifier
// that prices and entitlements reference; Type decides how the feature resolves.
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/features"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	feature, err := api.Features.Create(context.Background(), &features.CreateParams{
		Key:         "api_calls",
		Name:        "API calls",
		Type:        features.TypeQuantity,
		Description: "Monthly API call quota",
		Metadata:    map[string]string{"category": "usage"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("created %s — %s [%s]\n", feature.ID, feature.Key, feature.Type)
}
