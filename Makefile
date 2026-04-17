BINARY_NAME := fluxcd-policyctl
DOCKER_REPO := ghcr.io/nikogura/fluxcd-policyctl

.PHONY: build build-ui build-go test lint clean docker-build help

.DEFAULT_GOAL := build

help:
	@echo "Targets:"
	@echo "  build        - Build UI and Go binary"
	@echo "  build-ui     - Build Next.js frontend"
	@echo "  build-go     - Build Go binary"
	@echo "  test         - Run Go tests"
	@echo "  lint         - Run all linters (namedreturns + golangci-lint + eslint)"
	@echo "  clean        - Remove build artifacts"
	@echo "  docker-build - Build Docker image"

build: build-ui build-go

build-ui:
	@echo "Building UI..."
	cd pkg/ui && npm ci && npm run build

build-go:
	@echo "Building Go binary..."
	CGO_ENABLED=0 go build -a -installsuffix cgo -o $(BINARY_NAME) .

test:
	@echo "Running tests..."
	go test -v -race ./test/

lint:
	@echo "Running namedreturns linter..."
	@for pkg in $$(go list ./... | grep -v node_modules); do \
		namedreturns $$pkg || exit 1; \
	done
	@echo "Running golangci-lint..."
	golangci-lint run --timeout=5m
	@echo "Running ESLint..."
	cd pkg/ui && ./node_modules/.bin/eslint 'src/**/*.{ts,tsx}' --max-warnings 0

clean:
	rm -f $(BINARY_NAME)
	rm -rf pkg/ui/dist pkg/ui/.next pkg/ui/node_modules

docker-build:
	docker build -t $(DOCKER_REPO):latest .
