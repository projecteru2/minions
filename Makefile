.PHONY: deps build test binary

REPO_PATH := github.com/projecteru2/minions
REVISION := $(shell git rev-parse HEAD || unknown)
BUILTAT := $(shell date +%Y-%m-%dT%H:%M:%S)
VERSION := $(shell git describe --tags $(shell git rev-list --tags --max-count=1))
GO_LDFLAGS ?= -s -X $(REPO_PATH)/versioninfo.REVISION=$(REVISION) \
			  -X $(REPO_PATH)/versioninfo.BUILTAT=$(BUILTAT) \
			  -X $(REPO_PATH)/versioninfo.VERSION=$(VERSION)

clean:
	rm -rf eru-minions

deps:
	go mod download

binary:
	go build -ldflags "$(GO_LDFLAGS)" -a -tags netgo -installsuffix netgo -o eru-minions

debug-binary:
	go build -ldflags "$(GO_LDFLAGS)" -gcflags "-N -l" -a -tags netgo -installsuffix netgo -o eru-minions

build: deps binary

debug-build: deps debug-binary

docker_build:
	docker build .

test: deps
	# fix mock docker client bug, see https://github.com/moby/moby/pull/34383 [docker 17.05.0-ce]
	# sed -i.bak "143s/\*http.Transport/http.RoundTripper/" ./vendor/github.com/docker/docker/client/client.go
	go vet `go list ./... | grep -v '/vendor/' | grep -v '/tools'`
	go test -timeout 120s -count=1 -cover ./...

lint:
	golangci-lint run
