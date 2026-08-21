MODULE_NAME = grpcmock

VENDOR_DIR = vendor

GOLANGCI_LINT_VERSION ?= v2.13.0
MOCKERY_VERSION ?= v2.53.3

GO ?= go
GOLANGCI_LINT ?= $(shell go env GOPATH)/bin/golangci-lint-$(GOLANGCI_LINT_VERSION)
MOCKERY ?= $(shell $(GO) env GOPATH)/bin/mockery-$(MOCKERY_VERSION)

GITHUB_OUTPUT ?= /dev/null

# Other config
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

ifeq ($(V),1)
  Q = @set -x;
else
  Q = @
endif

.PHONY: $(VENDOR_DIR)
$(VENDOR_DIR):
	$(Q)mkdir -p $(VENDOR_DIR)
	$(Q)$(GO) mod vendor
	$(Q)$(GO) mod tidy

.PHONY: bump-deps
bump-deps:
	$(Q)$(GO) get -u ./...

.PHONY: tidy
tidy:
	$(Q)$(GO) mod tidy

.PHONY: lint
lint: $(GOLANGCI_LINT)
	@printf -- "$(OK_COLOR)==> lint$(NO_COLOR)\n"
	$(Q)$(GOLANGCI_LINT) run -c .golangci.yaml --allow-parallel-runners

.PHONY: test
test: test-unit

## Run unit tests
.PHONY: test-unit
test-unit:
	@printf -- "$(OK_COLOR)==> unit test$(NO_COLOR)\n"
	$(Q)$(GO) test -gcflags=-l -coverprofile=unit.coverprofile -covermode=atomic -race ./... -tags testcoverage

#.PHONY: test-integration
#test-integration:
#	@printf -- "$(OK_COLOR)==> integration test$(NO_COLOR)\n"
#	$(Q)$(GO) test ./features/... -gcflags=-l -coverprofile=features.coverprofile -coverpkg ./... -race --godog

.PHONY: gen
gen: gen-proto-fixtures gen-mocks

.PHONY: gen-mocks
gen-mocks: $(MOCKERY)
	@printf -- "$(OK_COLOR)==> generate mocks$(NO_COLOR)\n"
	$(Q)$(MOCKERY) --config .mockery.yaml

.PHONY: gen-proto-fixtures
gen-proto-fixtures:
	@printf -- "$(OK_COLOR)==> generate fixtures$(NO_COLOR)\n"
	$(Q)rm -rf test/grpctest
	$(Q)protoc --go_out=. --go-grpc_out=. resources/protobuf/service.proto

.PHONY: $(GITHUB_OUTPUT)
$(GITHUB_OUTPUT):
	@echo "MODULE_NAME=$(MODULE_NAME)" >> "$@"
	@echo "GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION)" >> "$@"

$(GOLANGCI_LINT):
	@printf -- "$(OK_COLOR)==> Installing golangci-lint $(GOLANGCI_LINT_VERSION)$(NO_COLOR)\n"
	$(Q)curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin "$(GOLANGCI_LINT_VERSION)"
	$(Q)mv ./bin/golangci-lint $(GOLANGCI_LINT)

$(MOCKERY):
	@printf -- "$(OK_COLOR)==> Installing mockery $(MOCKERY_VERSION)$(NO_COLOR)\n"
	$(Q)GOBIN=/tmp $(GO) install github.com/vektra/mockery/$(shell echo "$(MOCKERY_VERSION)" | cut -d '.' -f 1)$(Q)$(MOCKERY_VERSION)
	$(Q)mv /tmp/mockery $(MOCKERY)
