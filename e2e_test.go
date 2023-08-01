package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/efficientgo/core/backoff"
	"github.com/efficientgo/core/testutil"
	"github.com/efficientgo/e2e"
	e2edb "github.com/efficientgo/e2e/db"
	e2einteractive "github.com/efficientgo/e2e/interactive"
	e2emon "github.com/efficientgo/e2e/monitoring"
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
  scrape_interval: 1s
  scrape_timeout: 1s
  static_configs:
  - targets: [%s]
  relabel_configs:
  - source_labels: ['__address__']
    regex: '^.+:80$'
    action: drop
`, app.InternalEndpoint("http"))

	p1 := e2edb.NewPrometheus(e, "prometheus-1", e2edb.WithFlagOverride(map[string]string{
		"--enable-feature": "native-histograms",
	},))
	testutil.Ok(t, p1.SetConfigEncoded([]byte(config)))
	testutil.Ok(t, e2e.StartAndWaitReady(p1))

	fmt.Println("=== Ensure that Prometheus already scraped something")
	// Ensure that Prometheus already scraped something.
	testutil.Ok(t, p1.WaitSumMetrics(e2emon.Greater(50), "prometheus_tsdb_head_samples_appended_total"))

	// Open example in browser.
	exampleAppURL := fmt.Sprintf("http://%s", app.Endpoint("http"))
	fmt.Printf("=== Example application URL: %s\n", exampleAppURL)
	testutil.Ok(t, e2einteractive.OpenInBrowser(exampleAppURL))

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
	testutil.Ok(t, e2einteractive.OpenInBrowser(prometheusURL))

	// We're all done!
	fmt.Println("=== Setup finished!")

	testutil.Ok(t, e2einteractive.RunUntilEndpointHit())
}
