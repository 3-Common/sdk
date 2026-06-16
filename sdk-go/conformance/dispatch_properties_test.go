package conformance

// Properties-resource dispatcher for the conformance harness. Kept in its own
// file so adding new resources doesn't bloat the shared runner.

import (
	"context"
	"testing"

	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/properties"
)

func dispatchProperties(t *testing.T, api *client.API, ctx context.Context, sc scenario) (any, error) {
	t.Helper()

	switch sc.Call.Method {
	case "list":
		return api.Properties.List(ctx, buildPropertyListParams(sc.Call.Args))
	case "retrieve":
		id, _ := sc.Call.Args["id"].(string)
		return api.Properties.Retrieve(ctx, id)
	case "create":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Properties.Create(ctx, buildPropertyCreateParams(body))
	case "update":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Properties.Update(ctx, id, buildPropertyUpdateParams(body))
	case "listAutoPaginate":
		iter := api.Properties.ListAutoPaginate(ctx, buildPropertyListParams(sc.Call.Args))
		var collected []properties.Property
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	}
	t.Fatalf("unsupported properties scenario method %q", sc.Call.Method)
	return nil, nil
}

func buildPropertyListParams(args map[string]any) *properties.ListParams {
	if len(args) == 0 {
		return nil
	}
	p := &properties.ListParams{}
	for k, v := range args {
		switch k {
		case "page":
			p.Page = anyToIntPtr(v)
		case "pageSize":
			p.PageSize = anyToIntPtr(v)
		case "objectType":
			if s, ok := v.(string); ok {
				p.ObjectType = properties.ObjectType(s)
			}
		case "propertyType":
			if s, ok := v.(string); ok {
				p.PropertyType = properties.Type(s)
			}
		case "status":
			if s, ok := v.(string); ok {
				p.Status = properties.Status(s)
			}
		case "sort":
			if s, ok := v.(string); ok {
				p.Sort = s
			}
		case "order":
			if s, ok := v.(string); ok {
				p.Order = s
			}
		case "search":
			if s, ok := v.(string); ok {
				p.Search = s
			}
		}
	}
	return p
}

func buildPropertyCreateParams(body map[string]any) *properties.CreateParams {
	if body == nil {
		return nil
	}
	p := &properties.CreateParams{}
	if s, ok := body["type"].(string); ok {
		p.Type = properties.Type(s)
	}
	if s, ok := body["name"].(string); ok {
		p.Name = s
	}
	if s, ok := body["status"].(string); ok {
		p.Status = properties.Status(s)
	}
	if s, ok := body["objectType"].(string); ok {
		p.ObjectType = properties.ObjectType(s)
	}
	if s, ok := body["description"].(string); ok {
		p.Description = s
	}
	p.Options = buildPropertyOptions(body["options"])
	return p
}

func buildPropertyUpdateParams(body map[string]any) *properties.UpdateParams {
	if body == nil {
		return &properties.UpdateParams{}
	}
	p := &properties.UpdateParams{}
	if s, ok := body["name"].(string); ok {
		p.Name = s
	}
	if s, ok := body["status"].(string); ok {
		p.Status = properties.Status(s)
	}
	p.Options = buildPropertyOptions(body["options"])
	if v, ok := body["description"]; ok {
		if v == nil {
			p.ClearDescription = true
		} else if s, ok := v.(string); ok {
			p.Description = &s
		}
	}
	return p
}

func buildPropertyOptions(raw any) []properties.Option {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	var out []properties.Option
	for _, entry := range list {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		opt := properties.Option{}
		if s, ok := m["value"].(string); ok {
			opt.Value = s
		}
		if s, ok := m["label"].(string); ok {
			opt.Label = s
		}
		out = append(out, opt)
	}
	return out
}
