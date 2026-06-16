package properties_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/internal/core"
	"github.com/3-Common/sdk/sdk-go/resources/properties"
)

// newTestClient returns a properties.Client whose backend points at the
// supplied httptest server.
func newTestClient(t *testing.T, srv *httptest.Server) *properties.Client {
	t.Helper()
	cl, err := properties.New(threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)
	return cl
}

func TestNew_RequiresAPIKey(t *testing.T) {
	t.Setenv("THREECOMMON_API_KEY", "")
	_, err := properties.New(threecommon.Config{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_api_key", v.Code)
}

func TestFromBackend(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	backend, err := core.NewFromConfig(threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: threecommon.Int(0),
	})
	require.NoError(t, err)

	cl := properties.FromBackend(backend)
	require.NotNil(t, cl)
	_, err = cl.List(context.Background(), nil)
	require.NoError(t, err)
}

func TestList_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/properties", r.URL.Path)
		assert.Equal(t, "contact", r.URL.Query().Get("objectType"))
		assert.Equal(t, "active", r.URL.Query().Get("status"))
		_, _ = w.Write([]byte(`{
			"data":[{"type":"Text","id":"prop_1","name":"Dietary notes","status":"active","objectType":"contact"}],
			"hasMore":false
		}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.List(context.Background(), &properties.ListParams{
		ObjectType: properties.ObjectTypeContact,
		Status:     properties.StatusActive,
	})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, "prop_1", got.Data[0].ID)
	assert.Equal(t, properties.TypeText, got.Data[0].Type)
	assert.False(t, got.HasMore)
}

func TestList_AllParamsEncoded(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "100", q.Get("pageSize"))
		assert.Equal(t, "contact", q.Get("objectType"))
		assert.Equal(t, "Select One", q.Get("propertyType"))
		assert.Equal(t, "archived", q.Get("status"))
		assert.Equal(t, "name", q.Get("sort"))
		assert.Equal(t, "desc", q.Get("order"))
		assert.Equal(t, "shirt", q.Get("search"))
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 2
	size := 100
	_, err := cl.List(context.Background(), &properties.ListParams{
		Page:         &page,
		PageSize:     &size,
		ObjectType:   properties.ObjectTypeContact,
		PropertyType: properties.TypeSelectOne,
		Status:       properties.StatusArchived,
		Sort:         "name",
		Order:        "desc",
		Search:       "shirt",
	})
	require.NoError(t, err)
}

func TestList_NilParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), nil)
	require.NoError(t, err)
}

func TestList_EmptyParamsReturnNil(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery, "empty ListParams must produce no query string")
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), &properties.ListParams{})
	require.NoError(t, err)
}

func TestList_500SurfacesAsServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), nil)
	var server *threecommon.ServerError
	require.True(t, errors.As(err, &server))
}

func TestRetrieve_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/properties/prop_123", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"type":"Select Multiple","id":"prop_123","name":"Allergies","description":"Known allergies","status":"active","objectType":"contact","options":[{"value":"peanuts","label":"Peanuts"},{"value":"shellfish","label":"Shellfish"}]}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "prop_123")
	require.NoError(t, err)
	assert.Equal(t, "prop_123", got.ID)
	assert.Equal(t, properties.TypeSelectMultiple, got.Type)
	assert.Equal(t, "Known allergies", got.Description)
	require.Len(t, got.Options, 2)
	assert.Equal(t, "peanuts", got.Options[0].Value)
}

func TestRetrieve_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := properties.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Retrieve(context.Background(), "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestRetrieve_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Retrieve(context.Background(), "prop_missing")
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestCreate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/properties", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Select One", got["type"])
		assert.Equal(t, "T-shirt size", got["name"])
		opts, _ := got["options"].([]any)
		require.Len(t, opts, 2)

		_, _ = w.Write([]byte(`{"data":{"type":"Select One","id":"prop_new","name":"T-shirt size","status":"active","objectType":"contact","options":[{"value":"s","label":"Small"},{"value":"m","label":"Medium"}]}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Create(context.Background(), &properties.CreateParams{
		Type:       properties.TypeSelectOne,
		Name:       "T-shirt size",
		Status:     properties.StatusActive,
		ObjectType: properties.ObjectTypeContact,
		Options: []properties.Option{
			{Value: "s", Label: "Small"},
			{Value: "m", Label: "Medium"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "prop_new", got.ID)
	require.Len(t, got.Options, 2)
}

func TestCreate_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := properties.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Create(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestCreate_422SurfacesAsValidation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, `{"error":{"code":"validation_error","message":"options required","details":{"field":"options"}}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Create(context.Background(), &properties.CreateParams{
		Type:       properties.TypeSelectOne,
		Name:       "T-shirt size",
		Status:     properties.StatusActive,
		ObjectType: properties.ObjectTypeContact,
	})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "validation_error", v.Code)
}

func TestUpdate_SendsBodyAndReturnsProperty(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/properties/prop_123", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Allergies", got["name"])
		// description explicitly null -> key present with nil value.
		v, ok := got["description"]
		assert.True(t, ok, "description key must be present")
		assert.Nil(t, v)

		_, _ = w.Write([]byte(`{"data":{"type":"Text","id":"prop_123","name":"Allergies","status":"active","objectType":"contact"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Update(context.Background(), "prop_123", &properties.UpdateParams{
		Name:             "Allergies",
		ClearDescription: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "prop_123", got.ID)
	assert.Equal(t, "Allergies", got.Name)
}

func TestUpdate_SetsDescriptionAndStatusAndOptions(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "new notes", got["description"])
		assert.Equal(t, "archived", got["status"])
		opts, _ := got["options"].([]any)
		require.Len(t, opts, 1)

		_, _ = w.Write([]byte(`{"data":{"type":"Select One","id":"prop_9","name":"Size","status":"archived","objectType":"contact","options":[{"value":"s","label":"Small"}]}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	desc := "new notes"
	got, err := cl.Update(context.Background(), "prop_9", &properties.UpdateParams{
		Status:      properties.StatusArchived,
		Description: &desc,
		Options:     []properties.Option{{Value: "s", Label: "Small"}},
	})
	require.NoError(t, err)
	assert.Equal(t, properties.StatusArchived, got.Status)
}

func TestUpdate_ValidatesID(t *testing.T) {
	t.Parallel()
	cl, _ := properties.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "", &properties.UpdateParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUpdate_ValidatesParams(t *testing.T) {
	t.Parallel()
	cl, _ := properties.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "prop_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestUpdate_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"gone"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Update(context.Background(), "prop_missing", &properties.UpdateParams{Name: "x"})
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestUpdateParams_MarshalJSON(t *testing.T) {
	t.Parallel()

	desc := "hello"
	tests := []struct {
		name   string
		params properties.UpdateParams
		assert func(t *testing.T, m map[string]any)
	}{
		{
			name:   "empty omits all fields",
			params: properties.UpdateParams{},
			assert: func(t *testing.T, m map[string]any) {
				assert.Empty(t, m)
			},
		},
		{
			name:   "name only",
			params: properties.UpdateParams{Name: "n"},
			assert: func(t *testing.T, m map[string]any) {
				assert.Equal(t, "n", m["name"])
				_, ok := m["description"]
				assert.False(t, ok)
			},
		},
		{
			name:   "clear description wins over value",
			params: properties.UpdateParams{Description: &desc, ClearDescription: true},
			assert: func(t *testing.T, m map[string]any) {
				v, ok := m["description"]
				assert.True(t, ok)
				assert.Nil(t, v)
			},
		},
		{
			name:   "description value set",
			params: properties.UpdateParams{Description: &desc},
			assert: func(t *testing.T, m map[string]any) {
				assert.Equal(t, "hello", m["description"])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := json.Marshal(tc.params)
			require.NoError(t, err)
			var m map[string]any
			require.NoError(t, json.Unmarshal(b, &m))
			tc.assert(t, m)
		})
	}
}

func TestListAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"type":"Text","id":"prop_1","name":"A","status":"active","objectType":"contact"},{"type":"Email","id":"prop_2","name":"B","status":"active","objectType":"contact"}],"hasMore":true}`,
		`{"data":[{"type":"Phone","id":"prop_3","name":"C","status":"active","objectType":"contact"}],"hasMore":false}`,
	}

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx, _ := strconv.Atoi(r.URL.Query().Get("page"))
		calls.Add(1)
		require.Less(t, idx, len(pages))
		_, _ = io.WriteString(w, pages[idx])
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	iter := cl.ListAutoPaginate(context.Background(), nil)

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"prop_1", "prop_2", "prop_3"}, ids)
	assert.Equal(t, int32(2), calls.Load())
}

func TestListAutoPaginate_HonorsExplicitStartPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5", r.URL.Query().Get("page"))
		_, _ = io.WriteString(w, `{"data":[{"type":"Text","id":"prop_p5","name":"P","status":"active","objectType":"contact"}],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	startPage := 5
	iter := cl.ListAutoPaginate(context.Background(), &properties.ListParams{Page: &startPage})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"prop_p5"}, ids)
}

func TestListAutoPaginate_SurfacesPageError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"code":"internal_error","message":"boom"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	iter := cl.ListAutoPaginate(context.Background(), nil)
	for iter.Next() { /* no values yielded before error */
	}
	require.Error(t, iter.Err())
	var server *threecommon.ServerError
	require.True(t, errors.As(iter.Err(), &server))
}

func TestListAutoPaginate_ContextCancellationStopsIteration(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[],"hasMore":true}`)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cl := newTestClient(t, srv)
	iter := cl.ListAutoPaginate(ctx, nil)
	for iter.Next() { /* no values yielded */
	}
	require.Error(t, iter.Err())
}
