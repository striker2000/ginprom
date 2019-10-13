package ginprom

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestPrometheus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	p := New()

	r.Use(p.Middleware)
	r.GET("/metrics", p.Handler)

	pings := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pings_received",
	})

	p.MustRegister(pings)

	r.GET("/ping", func(c *gin.Context) {
		pings.Inc()
		c.String(200, "pong")
	})

	sendRequest(t, r, "/ping")

	w := sendRequest(t, r, "/metrics")
	body := w.Body.String()

	assert.Contains(t, body, "\nginprom_request_duration_seconds_count 1\n")
	assert.Contains(t, body, "\nginprom_requests_in_flight 1\n")
	assert.Contains(t, body, "\nginprom_requests_total{code=\"200\"} 1\n")
	assert.Contains(t, body, "\npings_received 1\n")
}

func sendRequest(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", path, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	return w
}
