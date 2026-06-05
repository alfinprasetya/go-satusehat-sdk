.PHONY: help build fmt vet test test-cover test-integration test-integration-cover test-all check coverage-clean

COVERAGE_DIR := coverage

# Optional: load Kemkes credentials from .env for integration tests (file is gitignored).
load-env = set -a && [ -f .env ] && . ./.env; set +a

help:
	@echo "Targets:"
	@echo "  make test                    Unit tests (httptest, no credentials)"
	@echo "  make test-cover              Unit tests + HTML coverage in $(COVERAGE_DIR)/"
	@echo "  make test-integration        Staging smoke tests (-tags=integration, uses .env if present)"
	@echo "  make test-integration-cover  Integration tests + HTML coverage in $(COVERAGE_DIR)/"
	@echo "  make test-all                Unit + integration tests"
	@echo "  make coverage-clean          Remove $(COVERAGE_DIR)/"
	@echo "  make vet                     go vet ./..."
	@echo "  make build                   go build ./..."
	@echo "  make fmt                     gofmt -w ."
	@echo "  make check                   vet + unit tests"

build:
	go build ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./...

test-cover:
	@mkdir -p $(COVERAGE_DIR)
	go test -coverprofile=$(COVERAGE_DIR)/unit.out ./...
	go tool cover -func=$(COVERAGE_DIR)/unit.out | tee $(COVERAGE_DIR)/unit.txt
	go tool cover -html=$(COVERAGE_DIR)/unit.out -o $(COVERAGE_DIR)/unit.html
	@echo "Open $(COVERAGE_DIR)/unit.html in a browser for line-by-line coverage"

test-integration:
	@$(load-env) && go test -tags=integration ./...

test-integration-cover:
	@mkdir -p $(COVERAGE_DIR)
	@$(load-env) && go test -tags=integration -coverprofile=$(COVERAGE_DIR)/integration.out ./...
	go tool cover -func=$(COVERAGE_DIR)/integration.out | tee $(COVERAGE_DIR)/integration.txt
	go tool cover -html=$(COVERAGE_DIR)/integration.out -o $(COVERAGE_DIR)/integration.html
	@echo "Open $(COVERAGE_DIR)/integration.html in a browser for line-by-line coverage"

test-all: test test-integration

coverage-clean:
	rm -rf $(COVERAGE_DIR)

check: vet test
