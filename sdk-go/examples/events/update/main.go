// Run with: go run ./examples/events/update
package main

import (
	"context"
	"fmt"
	"log"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/events"
)

func main() {
	api, err := client.New(threecommon.Config{
		APIKey: "3co_your_api_key_here",
	})
	if err != nil {
		log.Fatal(err)
	}

	updated, err := api.Events.Update(context.Background(), "evt_replace_with_real_id", &events.UpdateParams{
		Name: threecommon.String("Renamed via SDK"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("updated %s — name is now %q\n", updated.ID, updated.Name)
}
