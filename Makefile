.PHONY: build clean run test fmt lint lint-install vet tidy hooks-install helm-install helm-upgrade helm-uninstall helm-package helm-lint

# Binary name
BINARY_NAME=hsm
HELM_RELEASE_NAME=hsm
HELM_NAMESPACE=default
HELM_CHART_DIR=./charts/hsm
#HELM_VALUES_FILE?=
HELM_VALUES_FILE?=./values.yaml

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
all: build

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Build for all platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 .

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*

# Run the application
run: build
	./$(BINARY_NAME) serve

# Run tests
test:
	$(GOTEST) -v ./...

# Format code
fmt:
	$(GOFMT) -s -w .

# Run go vet
vet:
	$(GOVET) ./...

# Install golangci-lint
lint-install:
	curl -sSfL https://golangci-lint.run/install.sh | sh -s v2.8.0

# Lint (requires golangci-lint)
lint:
	./bin/golangci-lint run ./...

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Download dependencies
deps:
	$(GOMOD) download

# Install the binary
install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

# Install git hooks
hooks-install:
	@command -v lefthook >/dev/null 2>&1 || go install github.com/evilmartians/lefthook@latest
	lefthook install

# Helm targets

# Install Helm chart
helm-install:
	helm install $(HELM_RELEASE_NAME) $(HELM_CHART_DIR) --namespace $(HELM_NAMESPACE) --create-namespace $(if $(HELM_VALUES_FILE),-f $(HELM_VALUES_FILE),)

# Upgrade Helm chart
helm-upgrade:
	helm upgrade --install $(HELM_RELEASE_NAME) $(HELM_CHART_DIR) --namespace $(HELM_NAMESPACE) $(if $(HELM_VALUES_FILE),-f $(HELM_VALUES_FILE),)

# Uninstall Helm chart
helm-uninstall:
	helm uninstall $(HELM_RELEASE_NAME) --namespace $(HELM_NAMESPACE)

# Package Helm chart
helm-package:
	helm package $(HELM_CHART_DIR) -d .

# Lint Helm chart
helm-lint:
	helm lint $(HELM_CHART_DIR)

# Template Helm chart (dry-run)
helm-template:
	helm template $(HELM_RELEASE_NAME) $(HELM_CHART_DIR) --namespace $(HELM_NAMESPACE) $(if $(HELM_VALUES_FILE),-f $(HELM_VALUES_FILE),)

# Show Helm chart status
helm-status:
	helm status $(HELM_RELEASE_NAME) --namespace $(HELM_NAMESPACE)

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for all platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  run          - Build and run the application"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golangci-lint"
	@echo "  lint-install - Install golangci-lint"
	@echo "  tidy         - Tidy dependencies"
	@echo "  deps         - Download dependencies"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo ""
	@echo "Helm targets:"
	@echo "  helm-install   - Install Helm chart"
	@echo "  helm-upgrade   - Upgrade or install Helm chart"
	@echo "  helm-uninstall - Uninstall Helm chart"
	@echo "  helm-package   - Package Helm chart"
	@echo "  helm-lint      - Lint Helm chart"
	@echo "  helm-template  - Template Helm chart (dry-run)"
	@echo "  helm-status    - Show Helm chart status"
	@echo ""
	@echo "  help         - Show this help"
