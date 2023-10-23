package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total amount of HTTP requests",
	}, []string{"foo"})
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "http_requests_summary_duration_seconds",
		Help: "Duration of HTTP requests",
	}, []string{"foo"})
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_requests_duration_seconds",
		Help: "Duration of HTTP requests",
	}, []string{"foo"})

	reg := prometheus.NewRegistry()
	reg.MustRegister(requestsTotal, summary, histogram)

	go func() {
		for {
			for i := 0; i < 10; i++ {
				requestsTotal.WithLabelValues("bar").Inc()
				summary.WithLabelValues("bar").Observe(float64(i))
				histogram.WithLabelValues("bar").Observe(float64(i))
				time.Sleep(1 * time.Second)
			}
			requestsTotal.DeleteLabelValues("bar")
			summary.DeleteLabelValues("bar")
			histogram.DeleteLabelValues("bar")
			time.Sleep(3 * time.Second)
		}
	}()

	http.Handle(
		"/metrics", promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{}),
	)

	// To test: curl -H 'Accept: application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.8' localhost:8080/metrics
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
