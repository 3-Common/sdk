// Run with: go run ./examples/events/retrieve
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

	ev, err := api.Events.Retrieve(context.Background(), "evt_replace_with_real_id", &events.RetrieveParams{
		Fields: "id,name,start,status",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("event %s — %q [%s]\n", ev.ID, ev.Name, ev.Status)
	if ev.Start != "" {
		fmt.Printf("  starts at %s\n", ev.Start)
	}
}
