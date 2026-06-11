package forms_test

import (
	"context"
	"encoding/json"
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
	"github.com/3-Common/sdk/sdk-go/resources/forms"
)

const formJSON = `{"id":"frm_123","name":"Customer survey","ownerId":"hst_1","status":"active","type":"standalone","submitButtonText":"Submit","submitButtonWidth":"auto","rows":[],"elements":[]}`

// newTestClient returns a forms.Client whose backend points at the supplied
// httptest server.
func newTestClient(t *testing.T, srv *httptest.Server) *forms.Client {
	t.Helper()
	cl, err := forms.New(threecommon.Config{
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
	_, err := forms.New(threecommon.Config{})
	var v *threecommon.ValidationError
	require.ErrorAs(t, err, &v)
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

	cl := forms.FromBackend(backend)
	require.NotNil(t, cl)
	_, err = cl.List(context.Background(), nil)
	require.NoError(t, err)
}

func TestList_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/forms", r.URL.Path)
		assert.Equal(t, "standalone", r.URL.Query().Get("type"))
		_, _ = w.Write([]byte(`{"data":[{"id":"frm_a","name":"Customer survey","numElements":4,"type":"standalone","status":"active"}],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.List(context.Background(), &forms.ListParams{Type: forms.FormTypeStandalone})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, "frm_a", got.Data[0].ID)
	assert.Equal(t, 4, got.Data[0].NumElements)
	assert.Equal(t, forms.FormStatusActive, got.Data[0].Status)
	assert.False(t, got.HasMore)
}

func TestList_AllParamsEncoded(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "10", q.Get("pageSize"))
		assert.Equal(t, "order", q.Get("type"))
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 2
	size := 10
	_, err := cl.List(context.Background(), &forms.ListParams{
		Page:     &page,
		PageSize: &size,
		Type:     forms.FormTypeOrder,
	})
	require.NoError(t, err)
}

func TestList_NilAndEmptyParamsSendNoQuery(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.List(context.Background(), nil)
	require.NoError(t, err)
	_, err = cl.List(context.Background(), &forms.ListParams{})
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
	require.ErrorAs(t, err, &server)
}

func TestCreate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/forms", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Customer survey", got["name"])
		assert.Equal(t, "standalone", got["type"])

		_, _ = w.Write([]byte(`{"data":` + formJSON + `}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Create(context.Background(), &forms.CreateParams{
		Name: "Customer survey",
		Type: forms.FormTypeStandalone,
	})
	require.NoError(t, err)
	assert.Equal(t, "frm_123", got.ID)
	assert.Equal(t, forms.FormTypeStandalone, got.Type)
}

func TestCreate_400SurfacesAsValidation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"code":"validation_error","message":"name is required"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Create(context.Background(), &forms.CreateParams{Type: forms.FormTypeStandalone})
	var v *threecommon.ValidationError
	require.ErrorAs(t, err, &v)
}

func TestRetrieve_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/forms/frm_123", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":` + formJSON + `}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "frm_123")
	require.NoError(t, err)
	assert.Equal(t, "frm_123", got.ID)
	assert.Equal(t, "Customer survey", got.Name)
}

func TestRetrieve_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Retrieve(context.Background(), "frm_missing")
	var nf *threecommon.NotFoundError
	require.ErrorAs(t, err, &nf)
}

func TestUpdate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/forms/frm_123", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Renamed survey", got["name"])
		assert.Equal(t, "active", got["status"])

		_, _ = w.Write([]byte(`{"data":` + formJSON + `}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Update(context.Background(), "frm_123", &forms.UpdateParams{
		Name:   threecommon.String("Renamed survey"),
		Status: forms.FormStatusActive,
	})
	require.NoError(t, err)
	assert.Equal(t, "frm_123", got.ID)
}

func TestDuplicate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/duplicate", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Customer survey (copy)", got["name"])

		_, _ = w.Write([]byte(`{"data":{"id":"frm_copy","name":"Customer survey (copy)","ownerId":"hst_1","status":"draft","type":"standalone","submitButtonText":"Submit","submitButtonWidth":"auto","rows":[],"elements":[]}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Duplicate(context.Background(), "frm_123", &forms.DuplicateParams{Name: "Customer survey (copy)"})
	require.NoError(t, err)
	assert.Equal(t, "frm_copy", got.ID)
}

func TestAddElement_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "What is your name?", got["prompt"])
		assert.Equal(t, "Text", got["type"])
		assert.Equal(t, true, got["required"])

		_, _ = w.Write([]byte(`{"data":{"id":"elm_new","prompt":"What is your name?","type":"Text","required":true}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.AddElement(context.Background(), "frm_123", &forms.AddElementParams{
		Prompt:   "What is your name?",
		Type:     forms.ElementTypeText,
		Required: threecommon.Bool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "elm_new", got.ID)
	assert.Equal(t, forms.ElementTypeText, got.Type)
}

func TestAddElement_400SurfacesAsValidation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"code":"validation_error","message":"prompt is required"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.AddElement(context.Background(), "frm_123", &forms.AddElementParams{Type: forms.ElementTypeText})
	var v *threecommon.ValidationError
	require.ErrorAs(t, err, &v)
}

func TestUpdateElement_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_123", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "What is your full name?", got["prompt"])

		_, _ = w.Write([]byte(`{"data":{"id":"elm_123","prompt":"What is your full name?","type":"Text","required":true}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.UpdateElement(context.Background(), "frm_123", "elm_123", &forms.UpdateElementParams{
		Prompt: threecommon.String("What is your full name?"),
	})
	require.NoError(t, err)
	assert.Equal(t, "elm_123", got.ID)
	assert.Equal(t, "What is your full name?", got.Prompt)
}

func TestDeleteElement_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_123", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"deletedElementId":"elm_123"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.DeleteElement(context.Background(), "frm_123", "elm_123")
	require.NoError(t, err)
	assert.Equal(t, "elm_123", got.DeletedElementID)
}

func TestDeleteElement_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"gone"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.DeleteElement(context.Background(), "frm_123", "elm_missing")
	var nf *threecommon.NotFoundError
	require.ErrorAs(t, err, &nf)
}

func TestMoveElement_SendsBodyAndReturnsForm(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_123/position", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.InDelta(t, float64(2), got["position"], 0.001)

		_, _ = w.Write([]byte(`{"data":` + formJSON + `}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.MoveElement(context.Background(), "frm_123", "elm_123", &forms.MoveElementParams{Position: 2})
	require.NoError(t, err)
	assert.Equal(t, "frm_123", got.ID)
}

func TestEnableOtherOption_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_select/other-option", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Other (please specify)", got["otherPrompt"])

		_, _ = w.Write([]byte(`{"data":{"id":"elm_select","type":"Select One or \"Other\"","otherPrompt":"Other (please specify)"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.EnableOtherOption(context.Background(), "frm_123", "elm_select", &forms.EnableOtherOptionParams{
		OtherPrompt: "Other (please specify)",
	})
	require.NoError(t, err)
	assert.Equal(t, forms.ElementTypeSelectOneOrOther, got.Type)
	assert.Equal(t, "Other (please specify)", got.OtherPrompt)
}

func TestDisableOtherOption_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_select/other-option", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"elm_select","type":"Select One"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.DisableOtherOption(context.Background(), "frm_123", "elm_select")
	require.NoError(t, err)
	assert.Equal(t, forms.ElementTypeSelectOne, got.Type)
}

func TestAddLogicRule_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_select/logic-rules", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "elm_followup", got["revealedElementId"])

		_, _ = w.Write([]byte(`{"data":{"id":"elm_select","type":"Select One","logicGroups":[{"revealedElementIndex":3,"optionIndices":[0],"operator":"any_of"}]}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.AddLogicRule(context.Background(), "frm_123", "elm_select", &forms.AddLogicRuleParams{
		RevealedElementID: "elm_followup",
		Condition: forms.LogicCondition{
			OptionIndices: []int{0},
			Operator:      forms.LogicOperatorAnyOf,
		},
	})
	require.NoError(t, err)
	require.Len(t, got.LogicGroups, 1)
	assert.Equal(t, forms.LogicOperatorAnyOf, got.LogicGroups[0].Operator)
	assert.Equal(t, 3, got.LogicGroups[0].RevealedElementIndex)
}

func TestRemoveLogicRule_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_select/logic-rules/elm_followup", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"elm_select","type":"Select One"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.RemoveLogicRule(context.Background(), "frm_123", "elm_select", "elm_followup")
	require.NoError(t, err)
	assert.Equal(t, "elm_select", got.ID)
}

func TestListAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"id":"frm_1","name":"First","type":"standalone","status":"active"},{"id":"frm_2","name":"Second","type":"standalone","status":"active"}],"hasMore":true}`,
		`{"data":[{"id":"frm_3","name":"Third","type":"standalone","status":"draft"}],"hasMore":false}`,
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
	assert.Equal(t, []string{"frm_1", "frm_2", "frm_3"}, ids)
	assert.Equal(t, int32(2), calls.Load())
}

func TestListAutoPaginate_HonorsExplicitStartPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5", r.URL.Query().Get("page"))
		_, _ = io.WriteString(w, `{"data":[{"id":"frm_p5","name":"P5","type":"standalone","status":"active"}],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	startPage := 5
	iter := cl.ListAutoPaginate(context.Background(), &forms.ListParams{Page: &startPage})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"frm_p5"}, ids)
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
	require.ErrorAs(t, iter.Err(), &server)
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

// TestValidation covers the client-side argument guards across every method:
// empty path ids surface as "missing_id" and nil bodies as "missing_body",
// both as *threecommon.ValidationError, before any HTTP request is made.
func TestValidation(t *testing.T) {
	t.Parallel()

	cl, err := forms.New(threecommon.Config{APIKey: "k"})
	require.NoError(t, err)
	ctx := context.Background()

	assertCode := func(code string, e error) {
		t.Helper()
		var v *threecommon.ValidationError
		require.ErrorAs(t, e, &v)
		assert.Equal(t, code, v.Code)
	}

	_, err = cl.Retrieve(ctx, "")
	assertCode("missing_id", err)

	_, err = cl.Create(ctx, nil)
	assertCode("missing_body", err)

	_, err = cl.Update(ctx, "", &forms.UpdateParams{})
	assertCode("missing_id", err)
	_, err = cl.Update(ctx, "frm_1", nil)
	assertCode("missing_body", err)

	_, err = cl.Duplicate(ctx, "", &forms.DuplicateParams{})
	assertCode("missing_id", err)
	_, err = cl.Duplicate(ctx, "frm_1", nil)
	assertCode("missing_body", err)

	_, err = cl.AddElement(ctx, "", &forms.AddElementParams{})
	assertCode("missing_id", err)
	_, err = cl.AddElement(ctx, "frm_1", nil)
	assertCode("missing_body", err)

	_, err = cl.UpdateElement(ctx, "", "elm_1", &forms.UpdateElementParams{})
	assertCode("missing_id", err)
	_, err = cl.UpdateElement(ctx, "frm_1", "", &forms.UpdateElementParams{})
	assertCode("missing_id", err)
	_, err = cl.UpdateElement(ctx, "frm_1", "elm_1", nil)
	assertCode("missing_body", err)

	_, err = cl.DeleteElement(ctx, "", "elm_1")
	assertCode("missing_id", err)
	_, err = cl.DeleteElement(ctx, "frm_1", "")
	assertCode("missing_id", err)

	_, err = cl.MoveElement(ctx, "", "elm_1", &forms.MoveElementParams{})
	assertCode("missing_id", err)
	_, err = cl.MoveElement(ctx, "frm_1", "", &forms.MoveElementParams{})
	assertCode("missing_id", err)
	_, err = cl.MoveElement(ctx, "frm_1", "elm_1", nil)
	assertCode("missing_body", err)

	_, err = cl.EnableOtherOption(ctx, "", "elm_1", &forms.EnableOtherOptionParams{})
	assertCode("missing_id", err)
	_, err = cl.EnableOtherOption(ctx, "frm_1", "", &forms.EnableOtherOptionParams{})
	assertCode("missing_id", err)
	_, err = cl.EnableOtherOption(ctx, "frm_1", "elm_1", nil)
	assertCode("missing_body", err)

	_, err = cl.DisableOtherOption(ctx, "", "elm_1")
	assertCode("missing_id", err)
	_, err = cl.DisableOtherOption(ctx, "frm_1", "")
	assertCode("missing_id", err)

	_, err = cl.AddLogicRule(ctx, "", "elm_1", &forms.AddLogicRuleParams{})
	assertCode("missing_id", err)
	_, err = cl.AddLogicRule(ctx, "frm_1", "", &forms.AddLogicRuleParams{})
	assertCode("missing_id", err)
	_, err = cl.AddLogicRule(ctx, "frm_1", "elm_1", nil)
	assertCode("missing_body", err)

	_, err = cl.RemoveLogicRule(ctx, "", "elm_1", "elm_2")
	assertCode("missing_id", err)
	_, err = cl.RemoveLogicRule(ctx, "frm_1", "", "elm_2")
	assertCode("missing_id", err)
	_, err = cl.RemoveLogicRule(ctx, "frm_1", "elm_1", "")
	assertCode("missing_id", err)
}
