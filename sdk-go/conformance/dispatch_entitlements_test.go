package conformance

// Entitlements-resource dispatcher for the conformance harness. Kept in its
// own file so adding new resources doesn't bloat the shared runner.

import (
	"context"
	"testing"

	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/entitlements"
)

func dispatchEntitlements(t *testing.T, api *client.API, ctx context.Context, sc scenario) (any, error) {
	t.Helper()

	switch sc.Call.Method {
	case "list":
		return api.Entitlements.List(ctx, buildEntitlementListParams(sc.Call.Args))
	case "retrieve":
		id, _ := sc.Call.Args["id"].(string)
		var rp *entitlements.RetrieveParams
		if raw, ok := sc.Call.Args["params"].(map[string]any); ok {
			rp = &entitlements.RetrieveParams{}
			if f, ok := raw["fields"].(string); ok {
				rp.Fields = f
			}
		}
		return api.Entitlements.Retrieve(ctx, id, rp)
	case "lookup":
		return api.Entitlements.Lookup(ctx, buildEntitlementLookupParams(sc.Call.Args))
	case "grant":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Entitlements.Grant(ctx, buildEntitlementGrantParams(body))
	case "consume":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Entitlements.Consume(ctx, buildEntitlementConsumeParams(body))
	case "listAutoPaginate":
		iter := api.Entitlements.ListAutoPaginate(ctx, buildEntitlementListParams(sc.Call.Args))
		var collected []entitlements.Entitlement
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	}
	t.Fatalf("unsupported entitlement scenario method %q", sc.Call.Method)
	return nil, nil
}

func buildEntitlementListParams(args map[string]any) *entitlements.ListParams {
	if len(args) == 0 {
		return nil
	}
	p := &entitlements.ListParams{}
	for k, v := range args {
		switch k {
		case "page":
			p.Page = anyToIntPtr(v)
		case "pageSize":
			p.PageSize = anyToIntPtr(v)
		case "contactId":
			if s, ok := v.(string); ok {
				p.ContactID = s
			}
		case "featureKey":
			if s, ok := v.(string); ok {
				p.FeatureKey = s
			}
		case "minBalance":
			p.MinBalance = anyToInt64Ptr(v)
		case "fields":
			if s, ok := v.(string); ok {
				p.Fields = s
			}
		}
	}
	return p
}

func buildEntitlementLookupParams(args map[string]any) *entitlements.LookupParams {
	p := &entitlements.LookupParams{}
	if s, ok := args["contactId"].(string); ok {
		p.ContactID = s
	}
	if s, ok := args["featureKey"].(string); ok {
		p.FeatureKey = s
	}
	if s, ok := args["fields"].(string); ok {
		p.Fields = s
	}
	return p
}

func buildEntitlementGrantParams(body map[string]any) *entitlements.GrantParams {
	if body == nil {
		return nil
	}
	p := &entitlements.GrantParams{}
	if s, ok := body["contactId"].(string); ok {
		p.ContactID = s
	}
	if s, ok := body["featureKey"].(string); ok {
		p.FeatureKey = s
	}
	if n := anyToInt64Ptr(body["amount"]); n != nil {
		p.Amount = *n
	}
	if s, ok := body["grantId"].(string); ok {
		p.GrantID = s
	}
	if md, ok := body["metadata"].(map[string]any); ok {
		out := map[string]string{}
		for k, v := range md {
			if s, ok := v.(string); ok {
				out[k] = s
			}
		}
		p.Metadata = out
	}
	return p
}

func buildEntitlementConsumeParams(body map[string]any) *entitlements.ConsumeParams {
	if body == nil {
		return nil
	}
	p := &entitlements.ConsumeParams{}
	if s, ok := body["contactId"].(string); ok {
		p.ContactID = s
	}
	if s, ok := body["featureKey"].(string); ok {
		p.FeatureKey = s
	}
	if n := anyToInt64Ptr(body["amount"]); n != nil {
		p.Amount = *n
	}
	if s, ok := body["reason"].(string); ok {
		p.Reason = s
	}
	return p
}
