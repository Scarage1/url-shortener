package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	RedirectTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redirects_total",
			Help: "Total number of URL redirections.",
		},
		[]string{"code", "country"},
	)

	DBErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total number of database errors.",
		},
	)

	RedisErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_errors_total",
			Help: "Total number of redis errors.",
		},
	)
)

func init() {
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(RedirectTotal)
	prometheus.MustRegister(DBErrorsTotal)
	prometheus.MustRegister(RedisErrorsTotal)
}

// Metrics returns a middleware that measures request counter and latency.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.Request.URL.Path

		HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
		HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
