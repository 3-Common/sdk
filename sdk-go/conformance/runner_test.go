// Package conformance runs the shared YAML scenarios at
// ../../conformance/scenarios/**/*.yaml against the Go SDK. Every other SDK in
// this monorepo runs the same scenarios; identical pass/fail across languages
// is the contract. Each resource has its own dispatcher file
// (dispatch_events_test.go, dispatch_invoices_test.go); add a sibling when
// introducing a new resource.
package conformance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"

	threecommon "github.com/3-Common/sdk/sdk-go"
	"github.com/3-Common/sdk/sdk-go/client"
)

type scenario struct {
	Name string `yaml:"name"`
	Call struct {
		Resource string         `yaml:"resource"`
		Method   string         `yaml:"method"`
		Args     map[string]any `yaml:"args"`
	} `yaml:"call"`
	Client struct {
		MaxRetries *int   `yaml:"maxRetries"`
		APIVersion string `yaml:"apiVersion"`
		Telemetry  *bool  `yaml:"telemetry"`
	} `yaml:"client"`
	ExpectedRequest    *expectedRequest `yaml:"expectedRequest"`
	MockResponse       *mockResponse    `yaml:"mockResponse"`
	Exchanges          []exchange       `yaml:"exchanges"`
	ExpectedResult     map[string]any   `yaml:"expectedResult"`
	ExpectedResultNull bool             `yaml:"expectedResultNull"`
	ExpectedAutoPaged  []map[string]any `yaml:"expectedAutoPaginated"`
	ExpectedError      *expectedError   `yaml:"expectedError"`
	ExpectedCallCount  *int             `yaml:"expectedCallCount"`
}

type expectedRequest struct {
	Method        string            `yaml:"method"`
	Path          string            `yaml:"path"`
	Query         map[string]string `yaml:"query"`
	Headers       map[string]string `yaml:"headers"`
	HeadersAbsent []string          `yaml:"headersAbsent"`
	// Body, when present, is deep-equal compared against the JSON request body
	// (mirroring the Node harness). BodyAbsent asserts no request body was sent.
	Body       any  `yaml:"body"`
	BodyAbsent bool `yaml:"bodyAbsent"`
}

type mockResponse struct {
	Status  int               `yaml:"status"`
	Headers map[string]string `yaml:"headers"`
	Body    any               `yaml:"body"`
}

type exchange struct {
	Request  expectedRequest `yaml:"request"`
	Response mockResponse    `yaml:"response"`
}

type expectedError struct {
	Type              string         `yaml:"type"`
	Code              string         `yaml:"code"`
	HTTPStatus        int            `yaml:"httpStatus"`
	RequestID         string         `yaml:"requestId"`
	RetryAfterSeconds *float64       `yaml:"retryAfterSeconds"`
	Details           map[string]any `yaml:"details"`
}

func TestConformance(t *testing.T) {
	t.Parallel()

	scenarios, err := loadScenarios("../../conformance/scenarios")
	require.NoError(t, err)
	require.NotEmpty(t, scenarios, "no conformance scenarios found")

	for _, sc := range scenarios {
		// sc.path is "events/list-happy.yaml" — keeps the resource visible
		// when reading test output.
		t.Run(sc.path, func(t *testing.T) {
			t.Parallel()
			runScenario(t, sc.scn)
		})
	}
}

type loadedScenario struct {
	path string
	scn  scenario
}

func loadScenarios(dir string) ([]loadedScenario, error) {
	var out []loadedScenario
	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		var sc scenario
		if err := yaml.Unmarshal(bytes, &sc); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		// Relative path (e.g. "events/list-happy.yaml") keeps the resource
		// visible in the t.Run subtest name.
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			rel = path
		}
		out = append(out, loadedScenario{path: rel, scn: sc})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return out, nil
}

func runScenario(t *testing.T, sc scenario) {
	t.Helper()

	exchanges := sc.Exchanges
	if len(exchanges) == 0 && sc.MockResponse != nil {
		req := expectedRequest{}
		if sc.ExpectedRequest != nil {
			req = *sc.ExpectedRequest
		}
		exchanges = []exchange{{Request: req, Response: *sc.MockResponse}}
	}

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := int(calls.Add(1)) - 1
		require.Less(t, idx, len(exchanges), "%s: more requests than expected", sc.Name)

		ex := exchanges[idx]
		assertRequestMatches(t, sc.Name, ex.Request, r)
		writeMockResponse(w, ex.Response)
	}))
	defer srv.Close()

	cfg := threecommon.Config{
		APIKey:     "3co_test",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	}
	if sc.Client.MaxRetries != nil {
		cfg.MaxRetries = sc.Client.MaxRetries
	}
	if sc.Client.APIVersion != "" {
		cfg.APIVersion = sc.Client.APIVersion
	}
	if sc.Client.Telemetry != nil {
		cfg.Telemetry = sc.Client.Telemetry
	}

	api, err := client.New(cfg)
	require.NoError(t, err)

	result, callErr := dispatch(t, api, sc)

	// Validate result / error.
	switch {
	case sc.ExpectedError != nil:
		require.Error(t, callErr, "%s: expected error", sc.Name)
		assertExpectedError(t, sc.Name, *sc.ExpectedError, callErr)
	case sc.ExpectedAutoPaged != nil:
		require.NoError(t, callErr, "%s: %v", sc.Name, callErr)
		assertAutoPaginated(t, sc.Name, sc.ExpectedAutoPaged, result)
	case sc.ExpectedResult != nil:
		require.NoError(t, callErr, "%s: %v", sc.Name, callErr)
		assertJSONShape(t, sc.Name, sc.ExpectedResult, result)
	case sc.ExpectedResultNull:
		require.NoError(t, callErr, "%s: %v", sc.Name, callErr)
		assert.Nil(t, result, "%s: result should be nil", sc.Name)
	}

	if sc.ExpectedCallCount != nil {
		assert.Equal(t, int32(*sc.ExpectedCallCount), calls.Load(), "%s: call count", sc.Name)
	}
}

// dispatch routes a scenario call to the appropriate resource dispatcher.
// Each resource lives in its own file (e.g. dispatch_events_test.go,
// dispatch_invoices_test.go); add a sibling case here when introducing a new
// resource.
func dispatch(t *testing.T, api *client.API, sc scenario) (any, error) {
	t.Helper()
	ctx := context.Background()

	resource := sc.Call.Resource
	if resource == "" {
		resource = "events"
	}

	switch resource {
	case "events":
		return dispatchEvents(t, api, ctx, sc)
	case "invoices":
		return dispatchInvoices(t, api, ctx, sc)
	case "subscriptions":
		return dispatchSubscriptions(t, api, ctx, sc)
	case "contacts":
		return dispatchContacts(t, api, ctx, sc)
	case "entitlements":
		return dispatchEntitlements(t, api, ctx, sc)
	case "prices":
		return dispatchPrices(t, api, ctx, sc)
	case "features":
		return dispatchFeatures(t, api, ctx, sc)
	case "forms":
		return dispatchForms(t, api, ctx, sc)
	case "properties":
		return dispatchProperties(t, api, ctx, sc)
	}
	t.Fatalf("unsupported scenario resource %q", resource)
	return nil, nil
}

func anyToIntPtr(v any) *int {
	switch n := v.(type) {
	case int:
		return &n
	case int64:
		i := int(n)
		return &i
	case float64:
		i := int(n)
		return &i
	}
	return nil
}

func assertRequestMatches(t *testing.T, scenarioName string, want expectedRequest, r *http.Request) {
	t.Helper()
	if want.Method != "" {
		assert.Equal(t, want.Method, r.Method, "%s: method", scenarioName)
	}
	if want.Path != "" {
		assert.Equal(t, want.Path, r.URL.Path, "%s: path", scenarioName)
	}
	for k, v := range want.Query {
		assert.Equal(t, v, r.URL.Query().Get(k), "%s: query[%s]", scenarioName, k)
	}
	for k, v := range want.Headers {
		assert.Equal(t, v, r.Header.Get(k), "%s: header[%s]", scenarioName, k)
	}
	for _, k := range want.HeadersAbsent {
		assert.Empty(t, r.Header.Get(k), "%s: header %s should be absent", scenarioName, k)
	}

	// Body assertion mirrors the Node harness: a non-nil want.Body is
	// deep-equal compared against the parsed JSON request body, so an SDK that
	// drops a field, sends an extra one, or omits an explicit null fails here.
	raw, _ := io.ReadAll(r.Body)
	if want.Body != nil {
		var got any
		if len(raw) > 0 {
			require.NoError(t, json.Unmarshal(raw, &got), "%s: parse request body", scenarioName)
		}
		assert.Equal(t, normalize(want.Body), normalize(got), "%s: body", scenarioName)
	}
	if want.BodyAbsent {
		assert.Empty(t, raw, "%s: request body should be absent", scenarioName)
	}
}

func writeMockResponse(w http.ResponseWriter, m mockResponse) {
	for k, v := range m.Headers {
		w.Header().Set(k, v)
	}
	status := m.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	if m.Body != nil {
		_ = json.NewEncoder(w).Encode(m.Body)
	}
}

func assertExpectedError(t *testing.T, scenarioName string, want expectedError, got error) {
	t.Helper()

	switch want.Type {
	case "ThreeCommonAuthError":
		var e *threecommon.AuthError
		require.True(t, errors.As(got, &e), "%s: want AuthError, got %T", scenarioName, got)
		assertAPIErrorFields(t, scenarioName, want, e.APIError)
	case "ThreeCommonPermissionError":
		var e *threecommon.PermissionError
		require.True(t, errors.As(got, &e), "%s: want PermissionError, got %T", scenarioName, got)
		assertAPIErrorFields(t, scenarioName, want, e.APIError)
	case "ThreeCommonNotFoundError":
		var e *threecommon.NotFoundError
		require.True(t, errors.As(got, &e), "%s: want NotFoundError, got %T", scenarioName, got)
		assertAPIErrorFields(t, scenarioName, want, e.APIError)
	case "ThreeCommonValidationError":
		var e *threecommon.ValidationError
		require.True(t, errors.As(got, &e), "%s: want ValidationError, got %T", scenarioName, got)
		assertAPIErrorFields(t, scenarioName, want, e.APIError)
	case "ThreeCommonConflictError":
		var e *threecommon.ConflictError
		require.True(t, errors.As(got, &e), "%s: want ConflictError, got %T", scenarioName, got)
		assertAPIErrorFields(t, scenarioName, want, e.APIError)
	case "ThreeCommonRateLimitError":
		var e *threecommon.RateLimitError
		require.True(t, errors.As(got, &e), "%s: want RateLimitError, got %T", scenarioName, got)
		assertAPIErrorFields(t, scenarioName, want, e.APIError)
		if want.RetryAfterSeconds != nil {
			assert.InDelta(t, *want.RetryAfterSeconds, e.RetryAfter.Seconds(), 0.001, "%s: retryAfter", scenarioName)
		}
	case "ThreeCommonServerError":
		var e *threecommon.ServerError
		require.True(t, errors.As(got, &e), "%s: want ServerError, got %T", scenarioName, got)
		assertAPIErrorFields(t, scenarioName, want, e.APIError)
	default:
		t.Fatalf("%s: unsupported expectedError.type %q", scenarioName, want.Type)
	}
}

func assertAPIErrorFields(t *testing.T, scenarioName string, want expectedError, got *threecommon.APIError) {
	t.Helper()
	if want.Code != "" {
		assert.Equal(t, want.Code, got.Code, "%s: error.code", scenarioName)
	}
	if want.HTTPStatus != 0 {
		assert.Equal(t, want.HTTPStatus, got.HTTPStatus, "%s: error.httpStatus", scenarioName)
	}
	if want.RequestID != "" {
		assert.Equal(t, want.RequestID, got.RequestID, "%s: error.requestId", scenarioName)
	}
	for k, v := range want.Details {
		assert.Equal(t, normalize(v), normalize(got.Details[k]), "%s: error.details[%s]", scenarioName, k)
	}
}

func assertJSONShape(t *testing.T, scenarioName string, want map[string]any, got any) {
	t.Helper()
	gotJSON, err := json.Marshal(got)
	require.NoError(t, err)

	var gotMap map[string]any
	require.NoError(t, json.Unmarshal(gotJSON, &gotMap))

	for k, wantVal := range want {
		assertSubset(t, fmt.Sprintf("%s: result[%s]", scenarioName, k), wantVal, gotMap[k])
	}
}

func assertAutoPaginated(t *testing.T, scenarioName string, want []map[string]any, got any) {
	t.Helper()
	gotJSON, err := json.Marshal(got)
	require.NoError(t, err)

	var gotList []map[string]any
	require.NoError(t, json.Unmarshal(gotJSON, &gotList))

	require.Equal(t, len(want), len(gotList), "%s: auto-paginated length", scenarioName)
	for i := range want {
		for k, wantVal := range want[i] {
			assertSubset(t, fmt.Sprintf("%s: paged[%d][%s]", scenarioName, i, k), wantVal, gotList[i][k])
		}
	}
}

// assertSubset checks whether want is contained in got.
//
// For maps, every key in want must also exist in got with the same value.
// For slices, each element in want is compared with the element at the same
// index in got.
// Scalar values are compared directly after numeric normalization.
//
// got can contain extra fields or elements that aren’t present in want.
func assertSubset(t *testing.T, prefix string, want, got any) {
	t.Helper()

	switch w := normalize(want).(type) {
	case map[string]any:
		gm, ok := normalize(got).(map[string]any)
		if !ok {
			t.Errorf("%s: want object, got %T", prefix, got)
			return
		}
		for k, v := range w {
			assertSubset(t, prefix+"."+k, v, gm[k])
		}
	case []any:
		gs, ok := normalize(got).([]any)
		if !ok {
			t.Errorf("%s: want array, got %T", prefix, got)
			return
		}
		require.Equal(t, len(w), len(gs), "%s: array length", prefix)
		for i, v := range w {
			assertSubset(t, fmt.Sprintf("%s[%d]", prefix, i), v, gs[i])
		}
	default:
		assert.Equal(t, w, normalize(got), prefix)
	}
}

// normalize converts scalars and nested values into a form we can compare.
//
// YAML decoding can produce int, float64, string, []any, and map[string]any.
// JSON decoding produces the same types, except numbers come through as
// float64.
//
// To keep comparisons consistent, ints are converted to float64.
func normalize(v any) any {
	switch x := v.(type) {
	case int:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case []any:
		out := make([]any, len(x))
		for i, item := range x {
			out[i] = normalize(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, item := range x {
			out[k] = normalize(item)
		}
		return out
	default:
		return v
	}
}
