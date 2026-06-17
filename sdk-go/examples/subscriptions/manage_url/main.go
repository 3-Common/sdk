// Run with: go run ./examples/subscriptions/manage_url
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
)

func main() {
	api, err := client.New(threecommon.Config{APIKey: "3co_your_api_key_here"})
	if err != nil {
		log.Fatal(err)
	}

	// Fetch the signed self-service portal URL for a subscription. Share the
	// returned link with the subscriber so they can view, cancel, or resume it.
	manage, err := api.Subscriptions.RetrieveManageURL(context.Background(), "sub_replace_with_real_id")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("manage URL: %s\n", manage.URL)
}
