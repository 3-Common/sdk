package conformance

// Contacts-resource dispatcher for the conformance harness. Kept in its own
// file so adding new resources doesn't bloat the shared runner.

import (
	"context"
	"testing"

	"github.com/3-Common/sdk/sdk-go/client"
	"github.com/3-Common/sdk/sdk-go/resources/contacts"
)

func dispatchContacts(t *testing.T, api *client.API, ctx context.Context, sc scenario) (any, error) {
	t.Helper()

	switch sc.Call.Method {
	case "list":
		return api.Contacts.List(ctx, buildContactListParams(sc.Call.Args))
	case "count":
		count, err := api.Contacts.Count(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]any{"count": count}, nil
	case "retrieve":
		id, _ := sc.Call.Args["id"].(string)
		return api.Contacts.Retrieve(ctx, id)
	case "create":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Contacts.Create(ctx, buildContactCreateParams(body))
	case "update":
		id, _ := sc.Call.Args["id"].(string)
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Contacts.Update(ctx, id, buildContactUpdateParams(body))
	case "delete":
		id, _ := sc.Call.Args["id"].(string)
		return api.Contacts.Delete(ctx, id)
	case "bulkUpsert":
		body, _ := sc.Call.Args["body"].(map[string]any)
		return api.Contacts.BulkUpsert(ctx, buildContactBulkUpsertParams(body))
	case "listActivity":
		id, _ := sc.Call.Args["id"].(string)
		var params *contacts.ActivityListParams
		if raw, ok := sc.Call.Args["params"].(map[string]any); ok {
			params = buildContactActivityParams(raw)
		}
		return api.Contacts.ListActivity(ctx, id, params)
	case "listAutoPaginate":
		iter := api.Contacts.ListAutoPaginate(ctx, buildContactListParams(sc.Call.Args))
		var collected []contacts.Contact
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	case "listActivityAutoPaginate":
		id, _ := sc.Call.Args["id"].(string)
		var params *contacts.ActivityListParams
		if raw, ok := sc.Call.Args["params"].(map[string]any); ok {
			params = buildContactActivityParams(raw)
		}
		iter := api.Contacts.ListActivityAutoPaginate(ctx, id, params)
		var collected []contacts.Activity
		for iter.Next() {
			collected = append(collected, iter.Current())
		}
		return collected, iter.Err()
	}
	t.Fatalf("unsupported contacts scenario method %q", sc.Call.Method)
	return nil, nil
}

func buildContactListParams(args map[string]any) *contacts.ListParams {
	if len(args) == 0 {
		return nil
	}
	p := &contacts.ListParams{}
	for k, v := range args {
		switch k {
		case "pageNumber":
			p.PageNumber = anyToIntPtr(v)
		case "pageSize":
			p.PageSize = anyToIntPtr(v)
		case "sortField":
			if s, ok := v.(string); ok {
				p.SortField = s
			}
		case "sortDirection":
			if s, ok := v.(string); ok {
				p.SortDirection = s
			}
		case "filter":
			if s, ok := v.(string); ok {
				p.Filter = contacts.QuickFilter(s)
			}
		case "filters":
			if s, ok := v.(string); ok {
				p.Filters = s
			}
		case "search":
			if s, ok := v.(string); ok {
				p.Search = s
			}
		}
	}
	return p
}

func buildContactActivityParams(args map[string]any) *contacts.ActivityListParams {
	if len(args) == 0 {
		return nil
	}
	p := &contacts.ActivityListParams{}
	for k, v := range args {
		switch k {
		case "pageNumber":
			p.PageNumber = anyToIntPtr(v)
		case "pageSize":
			p.PageSize = anyToIntPtr(v)
		case "filter":
			if s, ok := v.(string); ok {
				p.Filter = contacts.ActivityType(s)
			}
		case "sort":
			if s, ok := v.(string); ok {
				p.Sort = s
			}
		}
	}
	return p
}

func buildContactCreateParams(body map[string]any) *contacts.CreateParams {
	if body == nil {
		return nil
	}
	p := &contacts.CreateParams{}
	if s, ok := body["email"].(string); ok {
		p.Email = s
	}
	if s, ok := body["firstName"].(string); ok {
		p.FirstName = s
	}
	if s, ok := body["lastName"].(string); ok {
		p.LastName = s
	}
	if s, ok := body["phone"].(string); ok {
		p.Phone = s
	}
	return p
}

func buildContactUpdateParams(body map[string]any) *contacts.UpdateParams {
	if body == nil {
		return &contacts.UpdateParams{}
	}
	p := &contacts.UpdateParams{}
	if c, ok := body["contact"].(map[string]any); ok {
		if s, ok := c["firstName"].(string); ok {
			p.Contact.FirstName = s
		}
		if s, ok := c["lastName"].(string); ok {
			p.Contact.LastName = s
		}
		if s, ok := c["email"].(string); ok {
			p.Contact.Email = s
		}
		if s, ok := c["phone"].(string); ok {
			p.Contact.Phone = &s
		}
		if s, ok := c["status"].(string); ok {
			p.Contact.Status = contacts.Status(s)
		}
	}
	if s, ok := body["mergeWith"].(string); ok {
		p.MergeWith = s
	}
	if s, ok := body["resolution"].(string); ok {
		p.Resolution = contacts.MergeResolution(s)
	}
	return p
}

func buildContactBulkUpsertParams(body map[string]any) *contacts.BulkUpsertParams {
	if body == nil {
		return nil
	}
	p := &contacts.BulkUpsertParams{}
	if raw, ok := body["contacts"].([]any); ok {
		for _, entry := range raw {
			m, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			item := contacts.BulkUpsertItem{}
			if s, ok := m["email"].(string); ok {
				item.Email = s
			}
			if s, ok := m["firstName"].(string); ok {
				item.FirstName = s
			}
			if s, ok := m["lastName"].(string); ok {
				item.LastName = s
			}
			if s, ok := m["phone"].(string); ok {
				item.Phone = &s
			}
			if s, ok := m["status"].(string); ok {
				item.Status = contacts.Status(s)
			}
			p.Contacts = append(p.Contacts, item)
		}
	}
	return p
}
