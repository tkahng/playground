package apis_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/go-chi/chi/v5"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/tools/store"
)

type TestApi struct {
	TestApi humatest.TestAPI
	Api     apis.Api
	App     *core.BaseApp
	Cfg     *conf.EnvConfig
	Router  http.Handler
}

func SetupApi(t testing.TB, ctx context.Context, db database.Dbx) *TestApi {
	t.Helper()
	cfg := conf.ZeroEnvConfig()
	app := core.NewTestBaseApp(cfg, db)
	router, api := NewHumaApi(t, app)
	appApi := apis.NewAppApi(app, router, api)
	appApi.RegisterRoutes()
	testApi := &TestApi{
		TestApi: api,
		Api:     *appApi,
		App:     app,
		Cfg:     cfg,
		Router:  router,
	}
	return testApi
}
func NewHumaApi(tb testing.TB, app core.App, configs ...huma.Config) (chi.Router, humatest.TestAPI) {
	tb.Helper()
	for _, config := range configs {
		if config.OpenAPI == nil {
			panic("custom huma.Config structs must specify a value for OpenAPI")
		}
	}
	if len(configs) == 0 {
		configs = append(configs, huma.Config{
			OpenAPI: &huma.OpenAPI{
				Info: &huma.Info{
					Title:   "Test API",
					Version: "1.0.0",
				},
			},
			Formats: map[string]huma.Format{
				"application/json": huma.DefaultJSONFormat,
				"json":             huma.DefaultJSONFormat,
			},
			DefaultFormat: "application/json",
		})
	}
	r := chi.NewMux()
	return r, humatest.Wrap(tb, humachi.New(r, configs[0]))
}

// ApiScenario defines a single api request test case/scenario.
type ApiScenario struct {
	// Name is the test name.
	Name string

	// Method is the HTTP method of the test request to use.
	Method string

	// URL is the url/path of the endpoint you want to test.
	URL string

	// Body specifies the body to send with the request.
	//
	// For example:
	//
	//	strings.NewReader(`{"title":"abc"}`)
	Body io.Reader

	// ResponseBody specifies the expected response body.
	//
	// For example:
	//
	//	strings.NewReader(`{"title":"abc"}`)
	ResponseBody io.Reader

	// Headers specifies the headers to send with the request (e.g. "Authorization": "abc")
	Headers []string

	// Delay adds a delay before checking the expectations usually
	// to ensure that all fired non-awaited go routines have finished
	Delay time.Duration

	// Timeout specifies how long to wait before cancelling the request context.
	//
	// A zero or negative value means that there will be no timeout.
	Timeout time.Duration

	// expectations
	// ---------------------------------------------------------------

	// ExpectedStatus specifies the expected response HTTP status code.
	ExpectedStatus int

	// List of keywords that MUST exist in the response body.
	//
	// Either ExpectedContent or NotExpectedContent must be set if the response body is non-empty.
	// Leave both fields empty if you want to ensure that the response didn't have any body (e.g. 204).
	ExpectedContent []string

	// List of keywords that MUST NOT exist in the response body.
	//
	// Either ExpectedContent or NotExpectedContent must be set if the response body is non-empty.
	// Leave both fields empty if you want to ensure that the response didn't have any body (e.g. 204).
	NotExpectedContent []string

	// List of hook events to check whether they were fired or not.
	//
	// You can use the wildcard "*" event key if you want to ensure
	// that no other hook events except those listed have been fired.
	//
	// For example:
	//
	//	map[string]int{ "*": 0 } // no hook events were fired
	//	map[string]int{ "*": 0, "EventA": 2 } // no hook events, except EventA were fired
	//	map[string]int{ EventA": 2, "EventB": 0 } // ensures that EventA was fired exactly 2 times and EventB exactly 0 times.
	// ExpectedEvents map[string]int

	// test hooks
	// ---------------------------------------------------------------

	TestAppFactory func(t testing.TB) *TestApi
	BeforeTestFunc func(t testing.TB, app *core.BaseApp, scenario *ApiScenario)
	AfterTestFunc  func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder)
	Store          *store.Store[string, any]
}

// Test executes the test scenario.
//
// Example:
//
//	func TestListExample(t testing.TB) {
//	    scenario := tests.ApiScenario{
//	        Name:           "list example collection",
//	        Method:         http.MethodGet,
//	        URL:            "/api/collections/example/records",
//	        ExpectedStatus: 200,
//	        ExpectedContent: []string{
//	            `"totalItems":3`,
//	            `"id":"0yxhwia2amd8gec"`,
//	            `"id":"achvryl401bhse3"`,
//	            `"id":"llvuca81nly1qls"`,
//	        },
//	        ExpectedEvents: map[string]int{
//	            "OnRecordsListRequest": 1,
//	            "OnRecordEnrich":       3,
//	        },
//	    }
//
//	    scenario.Test(t)
//	}
func (scenario *ApiScenario) Test(t *testing.T) {
	t.Helper()
	t.Run(scenario.normalizedName(), func(t *testing.T) {
		scenario.test(t)
	})
}

// Benchmark benchmarks the test scenario.
//
// Example:
//
//	func BenchmarkListExample(b *testing.B) {
//	    scenario := tests.ApiScenario{
//	        Name:           "list example collection",
//	        Method:         http.MethodGet,
//	        URL:            "/api/collections/example/records",
//	        ExpectedStatus: 200,
//	        ExpectedContent: []string{
//	            `"totalItems":3`,
//	            `"id":"0yxhwia2amd8gec"`,
//	            `"id":"achvryl401bhse3"`,
//	            `"id":"llvuca81nly1qls"`,
//	        },
//	        ExpectedEvents: map[string]int{
//	            "OnRecordsListRequest": 1,
//	            "OnRecordEnrich":       3,
//	        },
//	    }
//
//	    scenario.Benchmark(b)
//	}
func (scenario *ApiScenario) Benchmark(b *testing.B) {
	b.Run(scenario.normalizedName(), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			scenario.test(b)
		}
	})
}

func (scenario *ApiScenario) normalizedName() string {
	var name = scenario.Name

	if name == "" {
		name = fmt.Sprintf("%s:%s", scenario.Method, scenario.URL)
	}

	return name
}

func (scenario *ApiScenario) test(t testing.TB) {
	t.Helper()
	testApi := scenario.TestAppFactory(t)
	if scenario.BeforeTestFunc != nil {
		scenario.BeforeTestFunc(t, testApi.App, scenario)
	}
	var args []any
	if len(scenario.Headers) != 0 {
		for _, header := range scenario.Headers {
			args = append(args, header)
		}
	}
	if scenario.Body != nil {
		args = append(args, scenario.Body)
	}
	recorder := testApi.TestApi.Do(scenario.Method, scenario.URL, args...)

	if recorder.Code != scenario.ExpectedStatus {
		t.Fatalf("Expected status code %d, got %d", scenario.ExpectedStatus, recorder.Code)
	}
	if len(scenario.ExpectedContent) == 0 && len(scenario.NotExpectedContent) == 0 && scenario.AfterTestFunc == nil {
		if len(recorder.Body.String()) != 0 {
			t.Fatalf("Expected empty body, got \n%v", recorder.Body.String())
		}
	} else {
		// normalize json response format
		scenario.ResponseBody = recorder.Body
		buffer := new(bytes.Buffer)
		err := json.Compact(buffer, recorder.Body.Bytes())
		var normalizedBody string
		if err != nil {
			// not a json...
			normalizedBody = recorder.Body.String()
		} else {
			normalizedBody = buffer.String()
		}

		for _, item := range scenario.ExpectedContent {
			if !strings.Contains(normalizedBody, item) {
				t.Fatalf("Cannot find %v in response body \n%v", item, normalizedBody)
			}
		}

		for _, item := range scenario.NotExpectedContent {
			if strings.Contains(normalizedBody, item) {
				t.Fatalf("Didn't expect %v in response body \n%v", item, normalizedBody)
			}
		}
	}
	if scenario.AfterTestFunc != nil {
		scenario.AfterTestFunc(t, testApi.App, scenario, recorder)
	}
}

func JsonToReader(t testing.TB, input any) *strings.Reader {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Errorf("Error marshalling input: %v", err)
	}
	return strings.NewReader(string(data))
}
