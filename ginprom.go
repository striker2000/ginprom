// Package ginprom collects metrics in the Gin web framework and exports it to the Prometheus.
//
// Usage:
//
//     r := gin.Default()
//     p := ginprom.New()
//
//     // Attach the middleware to collect common metrics
//     r.Use(p.Middleware)
//
//     // Set the handler to export collected metrics
//     r.GET("/metrics", p.Handler)
//
//     // Register a custom metric
//     pings := prometheus.NewCounter(prometheus.CounterOpts{
//         Name: "pings_received",
//     })
//
//     p.MustRegister(pings)
//
//     r.GET("/ping", func(c *gin.Context) {
//         pings.Inc()
//         c.JSON(200, gin.H{"message": "pong"})
//     })
//

package ginprom

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "ginprom"

// Prometheus provides methods to collect common and custom metrics.
type Prometheus struct {
	reg         *prometheus.Registry
	handler     http.Handler
	reqInFlight prometheus.Gauge
	reqTotal    *prometheus.CounterVec
	reqDuration prometheus.Histogram
}

// New creates a new Prometheus instance.
func New() *Prometheus {
	p := Prometheus{
		reg: prometheus.NewRegistry(),
		reqDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "Histogram of request latencies",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10, 60},
		}),
		reqInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "requests_in_flight",
			Help:      "Current number of requests in flight",
		}),
		reqTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "requests_total",
			Help:      "Total number of processed requests by status code",
		}, []string{"code"}),
	}

	p.handler = promhttp.HandlerFor(p.reg, promhttp.HandlerOpts{})

	p.MustRegister(p.reqDuration)
	p.MustRegister(p.reqInFlight)
	p.MustRegister(p.reqTotal)

	return &p
}

// MustRegister registers the custom collector and panics if any error occurs.
func (p *Prometheus) MustRegister(c prometheus.Collector) {
	p.reg.MustRegister(c)
}

// Middleware used to collect common HTTP metrics. Should be attached to the Gin router through Use().
func (p *Prometheus) Middleware(c *gin.Context) {
	p.reqInFlight.Inc()
	start := time.Now()

	c.Next()

	p.reqDuration.Observe(time.Since(start).Seconds())
	p.reqInFlight.Dec()

	p.reqTotal.WithLabelValues(
		strconv.Itoa(c.Writer.Status()),
	).Inc()
}

// Handler exports collected metrics to the caller.
func (p *Prometheus) Handler(c *gin.Context) {
	p.handler.ServeHTTP(c.Writer, c.Request)
}
