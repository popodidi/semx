ifeq ($(origin GITROOT), undefined)
GITROOT := $(shell git rev-parse --show-toplevel)
endif

GO ?= go
CURL ?= curl
MKDIR_P ?= mkdir -p

GO_IMPORT_PATH := github.com/popodidi/semx
GO_BIN_DIR := $(abspath $(dir $(GO)))
BINDIR ?= $(GITROOT)/bin
VERSION ?= dev
GIT_COMMIT ?= $(shell git rev-parse --short=10 HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u +%Y%m%d%H%M%S)

GOLANGCI_LINT_VERSION ?= v2.12.2
GOLANGCI_LINT_VERSION_NUMBER := $(patsubst v%,%,$(GOLANGCI_LINT_VERSION))
GOLANGCI_LINT ?= $(BINDIR)/golangci-lint

GO_BUILD_LDFLAGS =
GO_BUILD_LDFLAGS += -X $(GO_IMPORT_PATH)/internal/version.version=$(VERSION)
GO_BUILD_LDFLAGS += -X $(GO_IMPORT_PATH)/internal/version.commit=$(GIT_COMMIT)
GO_BUILD_LDFLAGS += -X $(GO_IMPORT_PATH)/internal/version.buildTime=$(BUILD_TIME)

.PHONY: build clean fmt-check install-golangci-lint lint pre-build test tidy vet

build: pre-build
	$(GO) build -ldflags "$(GO_BUILD_LDFLAGS)" -o $(BINDIR)/semx ./cmd/semx

pre-build:
	$(MKDIR_P) $(BINDIR)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "These files need gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

install-golangci-lint:
	@if [ ! -x "$(GOLANGCI_LINT)" ] || ! "$(GOLANGCI_LINT)" version | grep -q "version $(GOLANGCI_LINT_VERSION_NUMBER)"; then \
		$(MKDIR_P) "$(BINDIR)"; \
		$(CURL) -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b "$(BINDIR)" "$(GOLANGCI_LINT_VERSION)"; \
	else \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) is installed"; \
	fi

lint: install-golangci-lint
	PATH="$(GO_BIN_DIR):$(PATH)" $(GOLANGCI_LINT) run ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BINDIR)
