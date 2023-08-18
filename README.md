# Prometheus-CT-e2e-test

Sample repository to e2e test collection of Counter/Histogram/Summary created timestamps

### Running test

First build the instrumented app

```console
make docker-build
```

Now you can run the test with

```
make tests
```