GO ?= /Volumes/KyoMac/go/bin/go
GOCACHE ?= /private/tmp/oasgo-go-build
BINARY_NAME := oasgo
CMD_PACKAGE := ./cmd/oasgo
DIST_DIR := dist
VERSION ?= dev

.PHONY: build release clean test

build:
	GOCACHE=$(GOCACHE) $(GO) build -trimpath -ldflags="-s -w" -o ./bin/$(BINARY_NAME) $(CMD_PACKAGE)

release: clean
	GOCACHE=$(GOCACHE) GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PACKAGE)
	GOCACHE=$(GOCACHE) GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PACKAGE)
	GOCACHE=$(GOCACHE) GOOS=linux GOARCH=arm64 $(GO) build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_PACKAGE)
	GOCACHE=$(GOCACHE) GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PACKAGE)
	GOCACHE=$(GOCACHE) GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PACKAGE)

test:
	GOCACHE=$(GOCACHE) $(GO) test ./...

clean:
	rm -rf ./bin ./$(DIST_DIR)
