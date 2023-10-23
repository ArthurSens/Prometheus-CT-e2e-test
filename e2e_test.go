package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/efficientgo/core/backoff"
	"github.com/efficientgo/core/testutil"
	"github.com/efficientgo/e2e"
	e2edb "github.com/efficientgo/e2e/db"
	e2emon "github.com/efficientgo/e2e/monitoring"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func TestExampleApp(t *testing.T) {
	e, err := e2e.New()
	t.Cleanup(e.Close)
	testutil.Ok(t, err)

	app := e2emon.AsInstrumented(e.Runnable("example_app").
		WithPorts(map[string]int{"http": 8080}).
		Init(e2e.StartOptions{
			Image: "arthursens/test-ct:v0.0.1",
		}), "http")

	testutil.Ok(t, e2e.StartAndWaitReady(app))
	config := fmt.Sprintf(`
global:
  external_labels:
    prometheus: prometheus-example-app
scrape_configs:
- job_name: 'example-app'
  scrape_interval: 3s
  scrape_timeout: 3s
  static_configs:
  - targets: [%s]
  relabel_configs:
  - source_labels: ['__address__']
    regex: '^.+:80$'
    action: drop
`, app.InternalEndpoint("http"))

	p1 := e2edb.NewPrometheus(e, "prometheus-1", e2edb.WithFlagOverride(map[string]string{
		"--enable-feature": "native-histograms",
	}))
	testutil.Ok(t, p1.SetConfigEncoded([]byte(config)))
	testutil.Ok(t, e2e.StartAndWaitReady(p1))

	fmt.Println("=== Ensure that Prometheus already scraped something")
	// Ensure that Prometheus already scraped something.
	testutil.Ok(t, p1.WaitSumMetrics(e2emon.Greater(5), "prometheus_tsdb_head_samples_appended_total"))

	// Open example in browser.
	exampleAppURL := fmt.Sprintf("http://%s", app.Endpoint("http"))
	fmt.Printf("=== Example application URL: %s\n", exampleAppURL)
	// testutil.Ok(t, e2einteractive.OpenInBrowser(exampleAppURL))

	fmt.Println("=== I need at least 5 requests!")
	testutil.Ok(t, app.WaitSumMetricsWithOptions(
		e2emon.GreaterOrEqual(5),
		[]string{"http_requests_total"},
		e2emon.WithWaitBackoff(
			&backoff.Config{
				Min:        1 * time.Second,
				Max:        10 * time.Second,
				MaxRetries: 100,
			}),
		e2emon.WaitMissingMetrics()),
	)

	// Now opening Prometheus in browser as well.
	prometheusURL := fmt.Sprintf("http://%s", p1.Endpoint("http"))
	fmt.Printf("=== Prometheus URL: %s\n", prometheusURL)
	// testutil.Ok(t, e2einteractive.OpenInBrowser(prometheusURL))

	// We're all done!
	fmt.Println("=== Setup finished!")
	// Wait some time make sure some scrapes happened
	time.Sleep(time.Minute * 1)

	apiClient, err := api.NewClient(api.Config{
		Address: prometheusURL,
	})
	testutil.Ok(t, err)
	promAPI := promv1.NewAPI(apiClient)

	results, _, err := promAPI.Query(context.Background(), `http_requests_total[1d]`, time.Now())
	testutil.Ok(t, err)

	for _, sampleStream := range results.(model.Matrix) {
		for i := 1; i < len(sampleStream.Values); i++ {
			if sampleStream.Values[i].Value < sampleStream.Values[i-1].Value &&
				sampleStream.Values[i].Value != 0 {
				t.Errorf("reset detected in sample %d, expected value 0 but got %0.2f", i, sampleStream.Values[i].Value)
			}
		}
	}

	// testutil.Ok(t, e2einteractive.RunUntilEndpointHit())
}
