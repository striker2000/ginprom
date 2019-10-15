package ginprom

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestPrometheus(t *testing.T) {
	r := initGin()

	w := sendRequest(r, "/ping")
	assert.Equal(t, 200, w.Code)

	w = sendRequest(r, "/panic")
	assert.Equal(t, 500, w.Code)

	w = sendRequest(r, "/panic_connreset")
	assert.Equal(t, 200, w.Code)

	w = sendRequest(r, "/metrics")
	assert.Equal(t, 200, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, "\nginprom_request_duration_seconds_count 3\n")
	assert.Contains(t, body, "\nginprom_requests_in_flight 1\n")
	assert.Contains(t, body, "\nginprom_requests_total{code=\"200\"} 2\n")
	assert.Contains(t, body, "\nginprom_requests_total{code=\"500\"} 1\n")
	assert.Contains(t, body, "\npings_received 1\n")
}

func initGin() *gin.Engine {
	prom := New()

	pings := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pings_received",
	})

	prom.MustRegister(pings)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(gin.RecoveryWithWriter(nil), prom.Middleware)

	router.GET("/metrics", prom.Handler)

	router.GET("/ping", func(c *gin.Context) {
		pings.Inc()
		c.String(200, "pong")
	})

	router.GET("/panic", func(c *gin.Context) {
		panic("Panic!")
	})

	router.GET("/panic_connreset", func(c *gin.Context) {
		c.Status(200)
		panic(&net.OpError{Err: &os.SyscallError{Err: syscall.ECONNRESET}})
	})

	return router
}

func sendRequest(r http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}
