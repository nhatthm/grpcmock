VENDOR_DIR = vendor

GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: $(VENDOR_DIR) lint test test-unit

$(VENDOR_DIR):
	@mkdir -p $(VENDOR_DIR)
	@$(GO) mod vendor
	@$(GO) mod tidy

lint:
	@$(GOLANGCI_LINT) run

test: test-unit

## Run unit tests
test-unit:
	@echo ">> unit test"
	@$(GO) test -gcflags=-l -coverprofile=unit.coverprofile -covermode=atomic -race ./... -tags testcoverage

#test-integration:
#	@echo ">> integration test"
#	@$(GO) test ./features/... -gcflags=-l -coverprofile=features.coverprofile -coverpkg ./... -race --godog

gen:
	@rm -rf test/grpctest
	@protoc --go_out=. --go-grpc_out=. resources/protobuf/service.proto
