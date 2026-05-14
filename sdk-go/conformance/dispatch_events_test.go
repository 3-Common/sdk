package conformance

// Events-resource dispatcher for the conformance harness. Kept in its own
// file so adding new resources doesn't bloat the shared runner.

import (
	"context"
	"testing"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/events"
)

func dispatchEvents(t *testing.T, api *client.API, ctx context.Context, sc scenario) (any, error) {
	t.Helper()

	switch sc.Call.Method {
	case "list":
		return api.Events.List(ctx, buildEventListParams(sc.Call.Args))
	case "retrieve":
		id, _ := sc.Call.Args["id"].(string)
		var rp *events.RetrieveParams
		if raw, ok := sc.Call.Args["params"].(map[string]any); ok {
			rp = &events.RetrieveParams{}
			if f, ok := raw["fields"].(string); ok {
				rp.Fields = f
			}
		}
		return api.Events.Retrieve(ctx, id, rp)
	case "update":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Events.Update(ctx, id, buildEventUpdateParams(body))
	case "listAutoPaginate":
		iter := api.Events.ListAutoPaginate(ctx, buildEventListParams(sc.Call.Args))
		var collected []events.Event
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	}
	t.Fatalf("unsupported event scenario method %q", sc.Call.Method)
	return nil, nil
}

func buildEventListParams(args map[string]any) *events.ListParams {
	if len(args) == 0 {
		return nil
	}
	p := &events.ListParams{}
	for k, v := range args {
		switch k {
		case "status":
			if s, ok := v.(string); ok {
				p.Status = events.Status(s)
			}
		case "pageSize":
			p.PageSize = anyToIntPtr(v)
		case "page":
			p.Page = anyToIntPtr(v)
		case "search":
			if s, ok := v.(string); ok {
				p.Search = s
			}
		case "fields":
			if s, ok := v.(string); ok {
				p.Fields = s
			}
		case "filters":
			if s, ok := v.(string); ok {
				p.Filters = s
			}
		}
	}
	return p
}

func buildEventUpdateParams(body map[string]any) *events.UpdateParams {
	p := &events.UpdateParams{}
	if name, ok := body["name"].(string); ok {
		p.Name = threecommon.String(name)
	}
	return p
}
