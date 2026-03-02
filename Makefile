SHELL := /bin/bash
.SHELLFLAGS = -ec

ROOTDIR := $(shell pwd)

# ---------- OCB version & platform detection ----------
OCB_VERSION ?= 0.146.1

OS   := $(shell uname -s | sed -e 's/Darwin/darwin/' -e 's/Linux/linux/')
ARCH := $(shell uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')

OCB_URL := https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/cmd%2Fbuilder%2Fv$(OCB_VERSION)/ocb_$(OCB_VERSION)_$(OS)_$(ARCH)

# ---------- Targets ----------
.PHONY: all build run clean help test

all: build

help:
	@echo "Reproduction demo for OTLP exporter nil pointer dereference"
	@echo ""
	@echo "Usage:"
	@echo "  make build    - Build the custom collector"
	@echo "  make run      - Run the collector (will likely panic)"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make test     - Run unit tests"
	@echo ""
	@echo "Detected platform: $(OS)/$(ARCH)"
	@echo "OCB version:       $(OCB_VERSION)"

# Install OCB (OpenTelemetry Collector Builder)
ocb:
	@case "$(OS)" in darwin|linux) ;; \
		*) echo "Error: unsupported OS '$(OS)' (from uname -s). Supported: darwin, linux" >&2; exit 1 ;; esac
	@case "$(ARCH)" in amd64|arm64|ppc64le) ;; \
		*) echo "Error: unsupported architecture '$(ARCH)' (from uname -m). Supported: amd64, arm64, ppc64le" >&2; exit 1 ;; esac
	@echo "Downloading OCB v$(OCB_VERSION) for $(OS)/$(ARCH)..."
	curl --proto '=https' --tlsv1.2 -fL -o ocb \
		"$(OCB_URL)"
	chmod +x ocb

# Build the collector
build: ocb
	@echo "Building custom collector..."
	$(ROOTDIR)/ocb --config=manifest.yaml

# Run the collector - this should reproduce the panic
run: build
	@echo ""
	@echo "========================================="
	@echo "Running collector - expect a PANIC if the"
	@echo "OTLP exporter is not ready when receiver sends data"
	@echo "========================================="
	@echo ""
	$(ROOTDIR)/otelcol-dev/otelcol --config=config.yaml

test:
	go test -v ./...

clean:
	rm -rf otelcol-dev/
	rm -f ./ocb
