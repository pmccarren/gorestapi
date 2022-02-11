package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	MetricRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "HTTP Request Count.",
		},
		[]string{"status", "path"},
	)
	MetricRequestDurationSecondsBucket = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds_bucket",
			Help:    "Histogram of latencies for HTTP requests.",
			Buckets: []float64{0.1, 0.2, 0.4, 1, 3, 8, 20, 60, 120},
		},
		[]string{"status", "path"},
	)
)

func init() {
	prometheus.MustRegister(MetricRequestsTotal)
	prometheus.MustRegister(MetricRequestDurationSecondsBucket)
}

func metricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			MetricRequestsTotal.With(
				prometheus.Labels{
					"path":   path,
					"status": strconv.Itoa(ww.Status()),
				}).Inc()
			MetricRequestDurationSecondsBucket.With(
				prometheus.Labels{
					"path":   path,
					"status": strconv.Itoa(ww.Status()),
				}).Observe(time.Since(start).Seconds())
		})
	}
}
