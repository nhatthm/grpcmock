MODULE_NAME = grpcmock

VENDOR_DIR = vendor

GOLANGCI_LINT_VERSION ?= v1.64.7
MOCKERY_VERSION ?= v2.53.2

GO ?= go
GOLANGCI_LINT ?= $(shell go env GOPATH)/bin/golangci-lint-$(GOLANGCI_LINT_VERSION)
MOCKERY ?= $(shell $(GO) env GOPATH)/bin/mockery-$(MOCKERY_VERSION)

GITHUB_OUTPUT ?= /dev/null

.PHONY: $(VENDOR_DIR)
$(VENDOR_DIR):
	@mkdir -p $(VENDOR_DIR)
	@$(GO) mod vendor
	@$(GO) mod tidy

.PHONY: update
update:
	@$(GO) get -u ./...

.PHONY: tidy
tidy:
	@$(GO) mod tidy

.PHONY: lint
lint: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run -c .golangci.yaml

.PHONY: test
test: test-unit

## Run unit tests
.PHONY: test-unit
test-unit:
	@echo ">> unit test"
	@$(GO) test -gcflags=-l -coverprofile=unit.coverprofile -covermode=atomic -race ./... -tags testcoverage

#.PHONY: test-integration
#test-integration:
#	@echo ">> integration test"
#	@$(GO) test ./features/... -gcflags=-l -coverprofile=features.coverprofile -coverpkg ./... -race --godog

.PHONY: gen
gen: gen-proto-fixtures gen-mocks

.PHONY: gen-mocks
gen-mocks: $(MOCKERY)
	@$(MOCKERY) --config .mockery.yaml

.PHONY: gen-proto-fixtures
gen-proto-fixtures:
	@rm -rf test/grpctest
	@protoc --go_out=. --go-grpc_out=. resources/protobuf/service.proto

.PHONY: $(GITHUB_OUTPUT)
$(GITHUB_OUTPUT):
	@echo "MODULE_NAME=$(MODULE_NAME)" >> "$@"
	@echo "GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION)" >> "$@"

$(GOLANGCI_LINT):
	@echo "$(OK_COLOR)==> Installing golangci-lint $(GOLANGCI_LINT_VERSION)$(NO_COLOR)"; \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin "$(GOLANGCI_LINT_VERSION)"
	@mv ./bin/golangci-lint $(GOLANGCI_LINT)

$(MOCKERY):
	@echo "$(OK_COLOR)==> Installing mockery $(MOCKERY_VERSION)$(NO_COLOR)"; \
	GOBIN=/tmp $(GO) install github.com/vektra/mockery/$(shell echo "$(MOCKERY_VERSION)" | cut -d '.' -f 1)@$(MOCKERY_VERSION)
	@mv /tmp/mockery $(MOCKERY)
