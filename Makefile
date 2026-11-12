.PHONY: all build build-ui test clean install dev

BINARY_NAME=planmark
BUILD_DIR=bin

all: build

build-ui:
	@echo "Building frontend assets..."
	cd web && npm install && npm run build

build: build-ui
	@echo "Compiling $(BINARY_NAME) binary..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/planmark

test:
	@echo "Running unit tests..."
	go test -v -race ./...

dev:
	go run ./cmd/planmark review --dev examples/sample_plan.md

install: build
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	go install ./cmd/planmark

clean:
	@echo "Cleaning artifacts..."
	rm -rf $(BUILD_DIR) web/dist
