// Package metrics holds the gateway's Prometheus instrumentation: HTTP request metrics (as a
// Gin middleware) and metrics about the calls the gateway makes to the Python solver.
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "hopfield_gateway"

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace, Name: "http_requests_total",
		Help: "HTTP requests handled, by method, route and status code.",
	}, []string{"method", "path", "status"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace, Name: "http_request_duration_seconds",
		Help:    "HTTP request latency, by method and route.",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
	}, []string{"method", "path"})

	httpInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace, Name: "http_requests_in_flight",
		Help: "HTTP requests being served right now.",
	})

	solverRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace, Name: "solver_requests_total",
		Help: "Calls from the gateway to the Python solver, by outcome (success or error).",
	}, []string{"outcome"})

	solverDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace, Name: "solver_duration_seconds",
		Help:    "Time the solver took to answer a call (network included).",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
	})

	solverMatrixSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace, Name: "solver_matrix_size",
		Help:    "Rows (= columns) of the cost matrices sent to the solver.",
		Buckets: []float64{2, 3, 5, 8, 10, 15, 20, 30, 40, 50},
	})

	// The solver reports `iterations` but no converged flag. It stops at its iteration cap
	// (1000) when it does not converge, so the share of solves below the 999 bucket is the
	// convergence rate: see the README for the query.
	solverIterations = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace, Name: "solver_iterations",
		Help:    "Iterations the solver ran for a successful solve.",
		Buckets: []float64{10, 25, 50, 100, 250, 500, 750, 999, 1000},
	})
)

// Middleware records request count, latency and in-flight requests. The route template
// (c.FullPath) is the label, not the raw URL, so label cardinality stays bounded; requests
// that match no route are grouped under "unmatched".
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		httpInFlight.Inc()
		defer httpInFlight.Dec()

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		httpRequests.WithLabelValues(c.Request.Method, path, strconv.Itoa(c.Writer.Status())).Inc()
		httpDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}

// ObserveSolverCall records one call to the solver. iterations is only used on success.
func ObserveSolverCall(matrixSize int, duration time.Duration, iterations int, err error) {
	solverMatrixSize.Observe(float64(matrixSize))
	solverDuration.Observe(duration.Seconds())
	if err != nil {
		solverRequests.WithLabelValues("error").Inc()
		return
	}
	solverRequests.WithLabelValues("success").Inc()
	solverIterations.Observe(float64(iterations))
}
