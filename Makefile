export GOBIN := $(CURDIR)/bin
export PATH := $(GOBIN):$(PATH)

GIT_TAG := $(shell git describe --tags --always --abbrev=0)

BUILD_ARGS ?= -ldflags \
	"-X github.com/agalitsyn/activity/version.Tag=$(GIT_TAG)"

ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: build
build: $(GOBIN)
	go build -v $(BUILD_ARGS) -o $(GOBIN) ./cmd/...

.PHONY: clean
clean:
	rm -rf "$(GOBIN)"

$(GOBIN):
	mkdir -p $(GOBIN)


include bin-deps.mk

.PHONY: run-server
run-server:
	go run $(CURDIR)/cmd/server

.PHONY: run-agent
run-agent:
	go run $(CURDIR)/cmd/agent

.PHONY: test-short
test-short:
	go test -v -race -short ./...

.PHONY: test
test:
	go test -v -race ./...

.PHONY: fmt
fmt: $(GOLANGCI_BIN)
	$(GOLANGCI_BIN) run --fix ./...

.PHONY: lint
lint: $(GOLANGCI_BIN)
	@go version
	@$(GOLANGCI_BIN) version
	$(GOLANGCI_BIN) run ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: vendor-server-static
vendor-server-static:
	mkdir -pv cmd/server/static/vendor/htmx.org@2.0.4 && \
		wget -O cmd/server/static/vendor/htmx.org@2.0.4/htmx.min.js https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js && \
	mkdir -pv cmd/server/static/vendor/htmx-ext-ws@2.0.1 && \
		wget -O cmd/server/static/vendor/htmx-ext-ws@2.0.1/ws.js https://unpkg.com/htmx-ext-ws@2.0.1/ws.js
	mkdir -pv cmd/server/static/vendor/bootstrap@5.3.3 && \
		wget -O cmd/server/static/vendor/bootstrap@5.3.3/bootstrap.min.css https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css && \
		wget -O cmd/server/static/vendor/bootstrap@5.3.3/bootstrap.bundle.min.js https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js
	mkdir -pv cmd/server/static/vendor/popperjs@@2.11.8 && \
		wget -O cmd/server/static/vendor/popperjs@@2.11.8/popper.min.js https://cdn.jsdelivr.net/npm/@popperjs/core@2.11.8/dist/umd/popper.min.js
	mkdir -pv cmd/server/static/vendor/bootstrap-icons@1.11.3 && \
		wget -O cmd/server/static/vendor/bootstrap-icons@1.11.3/bootstrap-icons.min.css https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.min.css

.PHONY: db-init
db-init:
	docker exec -it influxdb influx setup --org activity --bucket activity --username admin --password adminadmin --token secret --force
