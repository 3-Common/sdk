package forms_test

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
	"github.com/3-Common/sdk/sdk-go/resources/forms"
)

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
		_, _ = w.Write([]byte(`{
			"data":[{"id":"frm_a","name":"Newsletter Signup","numElements":3,"type":"standalone","status":"active"}],
			"hasMore":false
		}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.List(context.Background(), &forms.ListParams{Type: forms.TypeStandalone})
	require.NoError(t, err)
	require.Len(t, got.Data, 1)
	assert.Equal(t, "frm_a", got.Data[0].ID)
	assert.Equal(t, forms.StatusActive, got.Data[0].Status)
	assert.Equal(t, 3, got.Data[0].NumElements)
	assert.False(t, got.HasMore)
}

func TestList_AllParamsEncoded(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "2", q.Get("page"))
		assert.Equal(t, "25", q.Get("pageSize"))
		assert.Equal(t, "order", q.Get("type"))
		_, _ = w.Write([]byte(`{"data":[],"hasMore":false}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	page := 2
	size := 25
	_, err := cl.List(context.Background(), &forms.ListParams{
		Page:     &page,
		PageSize: &size,
		Type:     forms.TypeOrder,
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
	_, err := cl.List(context.Background(), &forms.ListParams{})
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
		assert.Equal(t, "/v1/forms/frm_123", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"frm_123","name":"Registration","ownerId":"hst_1","type":"standalone","status":"active","elements":[{"id":"elm_1","type":"Text","prompt":"Name"}]}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Retrieve(context.Background(), "frm_123")
	require.NoError(t, err)
	assert.Equal(t, "frm_123", got.ID)
	assert.Equal(t, "hst_1", got.OwnerID)
	require.Len(t, got.Elements, 1)
	assert.Equal(t, forms.ElementText, got.Elements[0].Type)
}

func TestRetrieve_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
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
	_, err := cl.Retrieve(context.Background(), "frm_missing")
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestCreate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/forms", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Registration", got["name"])
		assert.Equal(t, "standalone", got["type"])

		_, _ = w.Write([]byte(`{"data":{"id":"frm_new","name":"Registration","ownerId":"hst_1","type":"standalone","status":"draft"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Create(context.Background(), &forms.CreateParams{
		Name: "Registration",
		Type: forms.TypeStandalone,
	})
	require.NoError(t, err)
	assert.Equal(t, "frm_new", got.ID)
	assert.Equal(t, forms.StatusDraft, got.Status)
}

func TestCreate_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Create(context.Background(), nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestCreate_400SurfacesAsValidation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"code":"validation_failed","message":"bad type"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.Create(context.Background(), &forms.CreateParams{Name: "X", Type: "invalid"})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "validation_failed", v.Code)
}

func TestUpdate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/forms/frm_123", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Updated Registration", got["name"])
		assert.Equal(t, "Sign up", got["submitButtonText"])

		_, _ = w.Write([]byte(`{"data":{"id":"frm_123","name":"Updated Registration","ownerId":"hst_1","type":"standalone","status":"active","submitButtonText":"Sign up"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Update(context.Background(), "frm_123", &forms.UpdateParams{
		Name:             "Updated Registration",
		Status:           forms.StatusActive,
		SubmitButtonText: "Sign up",
	})
	require.NoError(t, err)
	assert.Equal(t, "frm_123", got.ID)
	assert.Equal(t, "Sign up", got.SubmitButtonText)
}

func TestUpdate_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "", &forms.UpdateParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUpdate_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Update(context.Background(), "frm_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestDuplicate_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/duplicate", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Registration (Copy)", got["name"])

		_, _ = w.Write([]byte(`{"data":{"id":"frm_copy","name":"Registration (Copy)","ownerId":"hst_1","type":"standalone","status":"draft"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Duplicate(context.Background(), "frm_123", &forms.DuplicateParams{
		Name:   "Registration (Copy)",
		Status: forms.StatusDraft,
	})
	require.NoError(t, err)
	assert.Equal(t, "frm_copy", got.ID)
}

func TestDuplicate_NilParamsSendsNoBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Empty(t, body, "nil DuplicateParams must send no body")
		_, _ = w.Write([]byte(`{"data":{"id":"frm_copy","name":"Registration","ownerId":"hst_1","type":"standalone","status":"draft"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.Duplicate(context.Background(), "frm_123", nil)
	require.NoError(t, err)
	assert.Equal(t, "frm_copy", got.ID)
}

func TestDuplicate_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.Duplicate(context.Background(), "", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
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

		_, _ = w.Write([]byte(`{"data":{"id":"elm_1","prompt":"What is your name?","type":"Text","required":true}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	required := true
	got, err := cl.AddElement(context.Background(), "frm_123", &forms.AddElementParams{
		Type:     forms.ElementText,
		Prompt:   "What is your name?",
		Required: &required,
	})
	require.NoError(t, err)
	assert.Equal(t, "elm_1", got.ID)
	require.NotNil(t, got.Required)
	assert.True(t, *got.Required)
}

func TestAddElement_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.AddElement(context.Background(), "", &forms.AddElementParams{Type: forms.ElementText})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestAddElement_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.AddElement(context.Background(), "frm_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestUpdateElement_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_1", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "What is your full name?", got["prompt"])
		assert.Equal(t, false, got["required"])

		_, _ = w.Write([]byte(`{"data":{"id":"elm_1","prompt":"What is your full name?","type":"Text","required":false}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	notRequired := false
	got, err := cl.UpdateElement(context.Background(), "frm_123", "elm_1", &forms.UpdateElementParams{
		Prompt:   "What is your full name?",
		Required: &notRequired,
	})
	require.NoError(t, err)
	assert.Equal(t, "elm_1", got.ID)
	require.NotNil(t, got.Required)
	assert.False(t, *got.Required)
}

func TestUpdateElement_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.UpdateElement(context.Background(), "", "elm_1", &forms.UpdateElementParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUpdateElement_RequiresElementID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.UpdateElement(context.Background(), "frm_1", "", &forms.UpdateElementParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestUpdateElement_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.UpdateElement(context.Background(), "frm_1", "elm_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestUpdateElement_404Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"not_found","message":"element missing"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.UpdateElement(context.Background(), "frm_123", "elm_missing", &forms.UpdateElementParams{Prompt: "x"})
	var nf *threecommon.NotFoundError
	require.True(t, errors.As(err, &nf))
}

func TestDeleteElement_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_1", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"deletedElementId":"elm_1"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.DeleteElement(context.Background(), "frm_123", "elm_1")
	require.NoError(t, err)
	assert.Equal(t, "elm_1", got.DeletedElementID)
}

func TestDeleteElement_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.DeleteElement(context.Background(), "", "elm_1")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestDeleteElement_RequiresElementID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.DeleteElement(context.Background(), "frm_1", "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestMoveElement_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_1/position", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, float64(2), got["position"])

		_, _ = w.Write([]byte(`{"data":{"id":"frm_123","name":"Registration","ownerId":"hst_1","type":"standalone","status":"active"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.MoveElement(context.Background(), "frm_123", "elm_1", &forms.MoveElementParams{Position: 2})
	require.NoError(t, err)
	assert.Equal(t, "frm_123", got.ID)
}

func TestMoveElement_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.MoveElement(context.Background(), "", "elm_1", &forms.MoveElementParams{Position: 1})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestMoveElement_RequiresElementID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.MoveElement(context.Background(), "frm_1", "", &forms.MoveElementParams{Position: 1})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestMoveElement_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.MoveElement(context.Background(), "frm_1", "elm_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestAddLogicRule_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_1/logic-rules", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "elm_2", got["revealedElementId"])
		cond, _ := got["condition"].(map[string]any)
		assert.Equal(t, "any_of", cond["operator"])

		_, _ = w.Write([]byte(`{"data":{"id":"elm_1","type":"Select One","prompt":"How did you hear about us?"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.AddLogicRule(context.Background(), "frm_123", "elm_1", &forms.AddLogicRuleParams{
		RevealedElementID: "elm_2",
		Condition: forms.LogicCondition{
			OptionIndices: []int{0},
			Operator:      forms.LogicOperatorAnyOf,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, forms.ElementSelectOne, got.Type)
}

func TestAddLogicRule_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.AddLogicRule(context.Background(), "", "elm_1", &forms.AddLogicRuleParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestAddLogicRule_RequiresElementID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.AddLogicRule(context.Background(), "frm_1", "", &forms.AddLogicRuleParams{})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestAddLogicRule_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.AddLogicRule(context.Background(), "frm_1", "elm_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestAddLogicRule_400Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"code":"invalid_logic_source","message":"unsupported"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.AddLogicRule(context.Background(), "frm_123", "elm_text", &forms.AddLogicRuleParams{
		RevealedElementID: "elm_2",
		Condition:         forms.LogicCondition{OptionIndices: []int{0}, Operator: forms.LogicOperatorAnyOf},
	})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "invalid_logic_source", v.Code)
}

func TestRemoveLogicRule_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_1/logic-rules/elm_2", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"elm_1","type":"Select One"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.RemoveLogicRule(context.Background(), "frm_123", "elm_1", "elm_2")
	require.NoError(t, err)
	assert.Equal(t, forms.ElementSelectOne, got.Type)
}

func TestRemoveLogicRule_RequiresIDs(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})

	_, err := cl.RemoveLogicRule(context.Background(), "", "elm_1", "elm_2")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)

	_, err = cl.RemoveLogicRule(context.Background(), "frm_1", "", "elm_2")
	require.True(t, errors.As(err, &v))

	_, err = cl.RemoveLogicRule(context.Background(), "frm_1", "elm_1", "")
	require.True(t, errors.As(err, &v))
}

func TestEnableOtherOption_SendsBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_1/other-option", r.URL.Path)

		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "Other (please specify)", got["otherPrompt"])

		_, _ = w.Write([]byte(`{"data":{"id":"elm_1","type":"Select One or \"Other\"","otherPrompt":"Other (please specify)"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.EnableOtherOption(context.Background(), "frm_123", "elm_1", &forms.EnableOtherOptionParams{
		OtherPrompt: "Other (please specify)",
	})
	require.NoError(t, err)
	assert.Equal(t, forms.ElementSelectOneOther, got.Type)
	assert.Equal(t, "Other (please specify)", got.OtherPrompt)
}

func TestEnableOtherOption_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.EnableOtherOption(context.Background(), "", "elm_1", &forms.EnableOtherOptionParams{OtherPrompt: "x"})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestEnableOtherOption_RequiresElementID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.EnableOtherOption(context.Background(), "frm_1", "", &forms.EnableOtherOptionParams{OtherPrompt: "x"})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestEnableOtherOption_RequiresParams(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.EnableOtherOption(context.Background(), "frm_1", "elm_1", nil)
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_body", v.Code)
}

func TestEnableOtherOption_400Surfaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"code":"element_not_selection","message":"not selection"}}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	_, err := cl.EnableOtherOption(context.Background(), "frm_123", "elm_text", &forms.EnableOtherOptionParams{OtherPrompt: "x"})
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "element_not_selection", v.Code)
}

func TestDisableOtherOption_HappyPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v1/forms/frm_123/elements/elm_1/other-option", r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"id":"elm_1","type":"Select One"}}`))
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	got, err := cl.DisableOtherOption(context.Background(), "frm_123", "elm_1")
	require.NoError(t, err)
	assert.Equal(t, forms.ElementSelectOne, got.Type)
}

func TestDisableOtherOption_RequiresID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.DisableOtherOption(context.Background(), "", "elm_1")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestDisableOtherOption_RequiresElementID(t *testing.T) {
	t.Parallel()
	cl, _ := forms.New(threecommon.Config{APIKey: "k"})
	_, err := cl.DisableOtherOption(context.Background(), "frm_1", "")
	var v *threecommon.ValidationError
	require.True(t, errors.As(err, &v))
	assert.Equal(t, "missing_id", v.Code)
}

func TestListAutoPaginate_WalksEveryPage(t *testing.T) {
	t.Parallel()

	pages := []string{
		`{"data":[{"id":"frm_1","name":"First"},{"id":"frm_2","name":"Second"}],"hasMore":true}`,
		`{"data":[{"id":"frm_3","name":"Third"}],"hasMore":false}`,
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
	iter := cl.ListAutoPaginate(context.Background(), &forms.ListParams{Type: forms.TypeStandalone})

	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"frm_1", "frm_2", "frm_3"}, ids)
	assert.Equal(t, int32(2), calls.Load())
}

func TestListAutoPaginate_NilParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "0", r.URL.Query().Get("page"))
		_, _ = io.WriteString(w, `{"data":[{"id":"frm_1","name":"Only"}],"hasMore":false}`)
	}))
	defer srv.Close()

	cl := newTestClient(t, srv)
	iter := cl.ListAutoPaginate(context.Background(), nil)
	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	require.NoError(t, iter.Err())
	assert.Equal(t, []string{"frm_1"}, ids)
}

func TestListAutoPaginate_HonorsExplicitStartPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "5", r.URL.Query().Get("page"))
		_, _ = io.WriteString(w, `{"data":[{"id":"frm_p5","name":"Page5"}],"hasMore":false}`)
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
