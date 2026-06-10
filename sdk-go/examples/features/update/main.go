// Run with: go run ./examples/features/update
//
// Updates a feature's display fields. Key and Type are immutable — archive and
// create a new feature to change them.
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

	feature, err := api.Features.Update(context.Background(), "feat_replace_with_real_id", &features.UpdateParams{
		Name:        threecommon.String("API requests"),
		Description: threecommon.String("Monthly API request quota"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("updated %s — %s\n", feature.ID, feature.Name)
}
