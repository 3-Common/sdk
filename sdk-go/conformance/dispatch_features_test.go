package conformance

// Features-resource dispatcher for the conformance harness. Kept in its own
// file so adding new resources doesn't bloat the shared runner.

import (
	"context"
	"testing"

	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/features"
)

func dispatchFeatures(t *testing.T, api *client.API, ctx context.Context, sc scenario) (any, error) {
	t.Helper()

	switch sc.Call.Method {
	case "list":
		return api.Features.List(ctx, buildFeatureListParams(sc.Call.Args))
	case "resolve":
		return api.Features.Resolve(ctx, buildFeatureResolveParams(sc.Call.Args))
	case "retrieve":
		id, _ := sc.Call.Args["id"].(string)
		var rp *features.RetrieveParams
		if raw, ok := sc.Call.Args["params"].(map[string]any); ok {
			rp = &features.RetrieveParams{}
			if f, ok := raw["fields"].(string); ok {
				rp.Fields = f
			}
		}
		return api.Features.Retrieve(ctx, id, rp)
	case "create":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Features.Create(ctx, buildFeatureCreateParams(body))
	case "update":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Features.Update(ctx, id, buildFeatureUpdateParams(body))
	case "archive":
		id, _ := sc.Call.Args["id"].(string)
		return api.Features.Archive(ctx, id)
	case "unarchive":
		id, _ := sc.Call.Args["id"].(string)
		return api.Features.Unarchive(ctx, id)
	case "listAutoPaginate":
		iter := api.Features.ListAutoPaginate(ctx, buildFeatureListParams(sc.Call.Args))
		var collected []features.Feature
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	}
	t.Fatalf("unsupported feature scenario method %q", sc.Call.Method)
	return nil, nil
}

func buildFeatureListParams(args map[string]any) *features.ListParams {
	if len(args) == 0 {
		return nil
	}
	p := &features.ListParams{}
	for k, v := range args {
		switch k {
		case "page":
			p.Page = anyToIntPtr(v)
		case "pageSize":
			p.PageSize = anyToIntPtr(v)
		case "type":
			if s, ok := v.(string); ok {
				p.Type = features.Type(s)
			}
		case "active":
			p.Active = anyToBoolPtr(v)
		case "fields":
			if s, ok := v.(string); ok {
				p.Fields = s
			}
		}
	}
	return p
}

func buildFeatureResolveParams(args map[string]any) *features.ResolveParams {
	p := &features.ResolveParams{}
	if s, ok := args["contactId"].(string); ok {
		p.ContactID = s
	}
	if s, ok := args["featureKey"].(string); ok {
		p.FeatureKey = s
	}
	return p
}

func featureStringSlice(raw any) []string {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(list))
	for _, v := range list {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func featureMetadata(raw any) map[string]string {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func buildFeatureCreateParams(body map[string]any) *features.CreateParams {
	if body == nil {
		return nil
	}
	p := &features.CreateParams{}
	if s, ok := body["key"].(string); ok {
		p.Key = s
	}
	if s, ok := body["name"].(string); ok {
		p.Name = s
	}
	if s, ok := body["type"].(string); ok {
		p.Type = features.Type(s)
	}
	if s, ok := body["description"].(string); ok {
		p.Description = s
	}
	if raw, ok := body["enumValues"]; ok {
		p.EnumValues = featureStringSlice(raw)
	}
	if raw, ok := body["metadata"]; ok {
		p.Metadata = featureMetadata(raw)
	}
	return p
}

func buildFeatureUpdateParams(body map[string]any) *features.UpdateParams {
	if body == nil {
		return &features.UpdateParams{}
	}
	p := &features.UpdateParams{}
	if s, ok := body["name"].(string); ok {
		p.Name = &s
	}
	if s, ok := body["description"].(string); ok {
		p.Description = &s
	}
	if raw, ok := body["enumValues"]; ok {
		p.EnumValues = featureStringSlice(raw)
	}
	if raw, ok := body["metadata"]; ok {
		p.Metadata = featureMetadata(raw)
	}
	return p
}
