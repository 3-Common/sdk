package conformance

// Invoices-resource dispatcher for the conformance harness. Kept in its own
// file so adding new resources doesn't bloat the shared runner.

import (
	"context"
	"testing"

	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/invoices"
)

func dispatchInvoices(t *testing.T, api *client.API, ctx context.Context, sc scenario) (any, error) {
	t.Helper()

	switch sc.Call.Method {
	case "list":
		return api.Invoices.List(ctx, buildInvoiceListParams(sc.Call.Args))
	case "retrieve":
		id, _ := sc.Call.Args["id"].(string)
		var rp *invoices.RetrieveParams
		if raw, ok := sc.Call.Args["params"].(map[string]any); ok {
			rp = &invoices.RetrieveParams{}
			if f, ok := raw["fields"].(string); ok {
				rp.Fields = f
			}
		}
		return api.Invoices.Retrieve(ctx, id, rp)
	case "create":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Invoices.Create(ctx, buildInvoiceCreateParams(body))
	case "update":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Invoices.Update(ctx, id, buildInvoiceUpdateParams(body))
	case "finalize":
		id, _ := sc.Call.Args["id"].(string)
		return api.Invoices.Finalize(ctx, id)
	case "void":
		id, _ := sc.Call.Args["id"].(string)
		var vp *invoices.VoidParams
		if body, ok := sc.Call.Args["body"].(map[string]any); ok {
			vp = &invoices.VoidParams{}
			if r, ok := body["reason"].(string); ok {
				vp.Reason = r
			}
		}
		return api.Invoices.Void(ctx, id, vp)
	case "recordPayment":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Invoices.RecordPayment(ctx, id, buildInvoicePaymentParams(body))
	case "listAutoPaginate":
		iter := api.Invoices.ListAutoPaginate(ctx, buildInvoiceListParams(sc.Call.Args))
		var collected []invoices.Invoice
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	}
	t.Fatalf("unsupported invoice scenario method %q", sc.Call.Method)
	return nil, nil
}

func buildInvoiceListParams(args map[string]any) *invoices.ListParams {
	if len(args) == 0 {
		return nil
	}
	p := &invoices.ListParams{}
	for k, v := range args {
		switch k {
		case "status":
			if s, ok := v.(string); ok {
				p.Status = invoices.Status(s)
			}
		case "page":
			p.Page = anyToIntPtr(v)
		case "pageSize":
			p.PageSize = anyToIntPtr(v)
		case "customerId":
			if s, ok := v.(string); ok {
				p.CustomerID = s
			}
		case "issuedAfter":
			if s, ok := v.(string); ok {
				p.IssuedAfter = s
			}
		case "issuedBefore":
			if s, ok := v.(string); ok {
				p.IssuedBefore = s
			}
		case "fields":
			if s, ok := v.(string); ok {
				p.Fields = s
			}
		}
	}
	return p
}

func buildInvoiceLineItems(raw any) []invoices.LineItem {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]invoices.LineItem, 0, len(list))
	for _, entry := range list {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		li := invoices.LineItem{}
		if s, ok := m["description"].(string); ok {
			li.Description = s
		}
		if q := anyToIntPtr(m["quantity"]); q != nil {
			li.Quantity = int64(*q)
		}
		if u := anyToIntPtr(m["unitAmount"]); u != nil {
			li.UnitAmount = int64(*u)
		}
		if s, ok := m["productId"].(string); ok {
			li.ProductID = s
		}
		if t := anyToIntPtr(m["taxAmount"]); t != nil {
			n := int64(*t)
			li.TaxAmount = &n
		}
		out = append(out, li)
	}
	return out
}

func buildInvoiceCreateParams(body map[string]any) *invoices.CreateParams {
	if body == nil {
		return nil
	}
	p := &invoices.CreateParams{}
	if s, ok := body["customerId"].(string); ok {
		p.CustomerID = s
	}
	if s, ok := body["currency"].(string); ok {
		p.Currency = invoices.Currency(s)
	}
	p.LineItems = buildInvoiceLineItems(body["lineItems"])
	if s, ok := body["notes"].(string); ok {
		p.Notes = s
	}
	if s, ok := body["dueAt"].(string); ok {
		p.DueAt = s
	}
	if s, ok := body["subscriptionId"].(string); ok {
		p.SubscriptionID = s
	}
	if s, ok := body["quoteId"].(string); ok {
		p.QuoteID = s
	}
	return p
}

func buildInvoiceUpdateParams(body map[string]any) *invoices.UpdateParams {
	if body == nil {
		return &invoices.UpdateParams{}
	}
	p := &invoices.UpdateParams{}
	if s, ok := body["customerId"].(string); ok {
		p.CustomerID = s
	}
	if raw, ok := body["lineItems"]; ok {
		p.LineItems = buildInvoiceLineItems(raw)
	}
	if s, ok := body["notes"].(string); ok {
		p.Notes = s
	}
	if s, ok := body["dueAt"].(string); ok {
		p.DueAt = s
	}
	return p
}

func buildInvoicePaymentParams(body map[string]any) *invoices.PaymentParams {
	if body == nil {
		return nil
	}
	p := &invoices.PaymentParams{}
	if v := anyToIntPtr(body["payment"]); v != nil {
		p.Payment = int64(*v)
	}
	if s, ok := body["idempotencyKey"].(string); ok {
		p.IdempotencyKey = s
	}
	if s, ok := body["note"].(string); ok {
		p.Note = s
	}
	return p
}
