# syntax=docker/dockerfile:1

FROM golang:1.20 as build

WORKDIR /go/src/github.com/ArthurSens/Prometheus-CT-e2e-test

RUN apt-get update

COPY . .

RUN make build

FROM gcr.io/distroless/static:latest-arm64

WORKDIR /test-ct

COPY --from=build /go/src/github.com/ArthurSens/Prometheus-CT-e2e-test/bin/* /bin/

USER nobody

ENTRYPOINT [ "/bin/test-ct" ]