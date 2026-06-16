// Run with: go run ./examples/properties/error_handling
//
// Demonstrate the typed error tree on the properties surface. Each subtype
// wraps a *APIError; branch with errors.As.
package main

import (
	"context"
	"errors"
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

	// A "Select One" property requires at least one option; omitting it
	// returns a 422 ValidationError.
	_, err = api.Properties.Create(context.Background(), &properties.CreateParams{
		Type:       properties.TypeSelectOne,
		Name:       "T-shirt size",
		Status:     properties.StatusActive,
		ObjectType: properties.ObjectTypeContact,
	})
	if err == nil {
		fmt.Println("(no error)")
		return
	}

	var (
		notFound   *threecommon.NotFoundError
		validation *threecommon.ValidationError
		auth       *threecommon.AuthError
		conn       *threecommon.ConnectionError
	)

	switch {
	case errors.As(err, &notFound):
		fmt.Printf("not found - request_id=%s\n", notFound.RequestID)
	case errors.As(err, &validation):
		fmt.Printf("validation: %s - details=%+v\n", validation.Message, validation.Details)
	case errors.As(err, &auth):
		fmt.Printf("auth failed: check your API key - code=%s\n", auth.Code)
	case errors.As(err, &conn):
		fmt.Printf("network error: %v\n", conn.Cause)
	default:
		fmt.Printf("unexpected error: %v\n", err)
	}
}
