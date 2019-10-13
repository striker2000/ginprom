# ginprom

[![GoDoc](https://godoc.org/github.com/striker2000/ginprom?status.svg)](https://godoc.org/github.com/striker2000/ginprom)
[![Build Status](https://travis-ci.org/striker2000/ginprom.svg?branch=master)](https://travis-ci.org/striker2000/ginprom)
[![Go Report Card](https://goreportcard.com/badge/github.com/striker2000/ginprom)](https://goreportcard.com/report/github.com/striker2000/ginprom)

Ginprom collects metrics in the Gin web framework and exports it to the Prometheus.

By default ginprom collects following metrics:
* `ginprom_request_duration_seconds` - histogram of request latencies
* `ginprom_requests_in_flight` - current number of requests in flight
* `ginprom_requests_total` - total number of processed requests by status code

Ginprom also allows to register and export custom user metrics.

## Usage

Set the middleware to collect common metrics:

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/striker2000/ginprom"
)

func main() {
	r := gin.Default()
	p := ginprom.New()

	// Attach the middleware to the router
	r.Use(p.Middleware)

	// Set the handler to export collected metrics
	r.GET("/metrics", p.Handler)

	r.Run()
}
```

Register your custom metrics:

```go
import "github.com/prometheus/client_golang/prometheus"

pings := prometheus.NewCounter(prometheus.CounterOpts{
	Name: "pings_received",
})

// Registered metric will be exported automatically via p.Handler
p.MustRegister(pings)

r.GET("/ping", func(c *gin.Context) {
	// Use the metric in your code
	pings.Inc()

	c.JSON(200, gin.H{"message": "pong"})
})
```
