RELEASE_VERSION ?=$(shell cat VERSION)
RELEASE=1
REVISION ?= $(shell git rev-parse HEAD)
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
BINARY_FOLDER=bin
BINARY_NAME=test-ct
ARTIFACT_NAME=arthursens/$(BINARY_NAME)
GOCMD=go
GOMAIN=main.go
GOBUILD=$(GOCMD) build
GOOS?=$(shell go env GOOS)
ENVVARS=GOOS=$(GOOS) CGO_ENABLED=0

LDFLAGS=-w -extldflags "-static"

docker-build:
	@DOCKER_BUILDKIT=1 docker build -t ${ARTIFACT_NAME}:${RELEASE_VERSION} -f Dockerfile --progress=plain .

build:
	$(ENVVARS) $(GOCMD) build -ldflags '$(LDFLAGS)' -o $(BINARY_FOLDER)/$(BINARY_NAME) -v $(GOMAIN)

deps:
	$(ENVVARS) $(GOCMD) mod download

fmt:
	$(ENVVARS) $(GOCMD) fmt -x ./...

vet:
	$(ENVVARS) $(GOCMD) vet ./...

tests:
	$(ENVVARS) $(GOCMD) test ./...

all: fmt vet tests deps build

.PHONY: build