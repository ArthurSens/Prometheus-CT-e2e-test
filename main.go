package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	requestsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total amount of HTTP requests",
	})

	reg := prometheus.NewRegistry()
	reg.MustRegister(requestsTotal)

	go func() {
		for {
			requestsTotal.Inc()
			time.Sleep(2 * time.Second)
		}
	}()

	http.Handle(
		"/metrics", promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				EnableCreatedTimestamps: true,
			}),
	)

	// To test: curl -H 'Accept: application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.8' localhost:8080/metrics
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
