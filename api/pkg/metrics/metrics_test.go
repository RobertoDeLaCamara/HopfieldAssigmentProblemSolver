package metrics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/items/:id", func(c *gin.Context) { c.String(http.StatusOK, c.Param("id")) })
	r.POST("/fail", func(c *gin.Context) { c.String(http.StatusInternalServerError, "x") })
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(20 * time.Millisecond)
		c.String(http.StatusOK, "ok")
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return r
}

func do(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}

func TestMiddleware_CountsByMethodRouteAndStatus(t *testing.T) {
	r := newRouter()
	okBefore := testutil.ToFloat64(httpRequests.WithLabelValues("GET", "/ok", "200"))
	failBefore := testutil.ToFloat64(httpRequests.WithLabelValues("POST", "/fail", "500"))

	do(r, "GET", "/ok")
	do(r, "GET", "/ok")
	do(r, "POST", "/fail")

	assert.Equal(t, okBefore+2, testutil.ToFloat64(httpRequests.WithLabelValues("GET", "/ok", "200")))
	assert.Equal(t, failBefore+1, testutil.ToFloat64(httpRequests.WithLabelValues("POST", "/fail", "500")))
}

func TestMiddleware_LabelsTheRouteTemplateNotTheURL(t *testing.T) {
	r := newRouter()
	before := testutil.ToFloat64(httpRequests.WithLabelValues("GET", "/items/:id", "200"))

	for _, id := range []string{"1", "2", "3", "abc"} {
		do(r, "GET", "/items/"+id)
	}

	// four different URLs, one time series: label cardinality stays bounded
	assert.Equal(t, before+4, testutil.ToFloat64(httpRequests.WithLabelValues("GET", "/items/:id", "200")))
}

func TestMiddleware_UnknownRoutesAreGrouped(t *testing.T) {
	r := newRouter()
	before := testutil.ToFloat64(httpRequests.WithLabelValues("GET", "unmatched", "404"))

	do(r, "GET", "/no/such/route/1")
	do(r, "GET", "/no/such/route/2")

	assert.Equal(t, before+2, testutil.ToFloat64(httpRequests.WithLabelValues("GET", "unmatched", "404")))
}

func TestMiddleware_RecordsLatencyAndReleasesInFlight(t *testing.T) {
	r := newRouter()
	do(r, "GET", "/slow")

	assert.Equal(t, 0.0, testutil.ToFloat64(httpInFlight), "no request is left counted as in flight")
	families, err := prometheusGather()
	require.NoError(t, err)
	hist := families["hopfield_gateway_http_request_duration_seconds"]
	require.NotNil(t, hist)
	var seen bool
	for _, m := range hist.Metric {
		for _, l := range m.Label {
			if l.GetName() == "path" && l.GetValue() == "/slow" {
				seen = true
				assert.GreaterOrEqual(t, m.Histogram.GetSampleSum(), 0.02)
			}
		}
	}
	assert.True(t, seen, "latency histogram has a series for /slow")
}

func TestObserveSolverCall(t *testing.T) {
	okBefore := testutil.ToFloat64(solverRequests.WithLabelValues("success"))
	errBefore := testutil.ToFloat64(solverRequests.WithLabelValues("error"))

	ObserveSolverCall(5, 30*time.Millisecond, 120, nil)
	ObserveSolverCall(5, 10*time.Millisecond, 0, errors.New("boom"))

	assert.Equal(t, okBefore+1, testutil.ToFloat64(solverRequests.WithLabelValues("success")))
	assert.Equal(t, errBefore+1, testutil.ToFloat64(solverRequests.WithLabelValues("error")))
}

func TestObserveSolverCall_IterationsOnlyForSuccessfulSolves(t *testing.T) {
	families, err := prometheusGather()
	require.NoError(t, err)
	before := families["hopfield_gateway_solver_iterations"].Metric[0].Histogram.GetSampleCount()

	ObserveSolverCall(3, time.Millisecond, 500, errors.New("failed"))
	ObserveSolverCall(3, time.Millisecond, 500, nil)

	families, err = prometheusGather()
	require.NoError(t, err)
	assert.Equal(t, before+1, families["hopfield_gateway_solver_iterations"].Metric[0].Histogram.GetSampleCount())
}

func TestMetricsEndpoint_ServesValidPrometheusText(t *testing.T) {
	r := newRouter()
	do(r, "GET", "/ok")
	ObserveSolverCall(4, 5*time.Millisecond, 1000, nil)

	w := do(r, "GET", "/metrics")
	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain"))

	parser := expfmt.NewTextParser(model.UTF8Validation)
	families, err := parser.TextToMetricFamilies(strings.NewReader(w.Body.String()))
	require.NoError(t, err, "the body must be valid Prometheus exposition text")
	for _, name := range []string{
		"hopfield_gateway_http_requests_total",
		"hopfield_gateway_http_request_duration_seconds",
		"hopfield_gateway_http_requests_in_flight",
		"hopfield_gateway_solver_requests_total",
		"hopfield_gateway_solver_duration_seconds",
		"hopfield_gateway_solver_matrix_size",
		"hopfield_gateway_solver_iterations",
	} {
		assert.Contains(t, families, name)
	}
}
