SHELL := /bin/bash
.SHELLFLAGS = -ec

ROOTDIR := $(shell pwd)

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

# Install OCB (OpenTelemetry Collector Builder)
ocb:
	curl --proto '=https' --tlsv1.2 -fL -o ocb \
	https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/cmd%2Fbuilder%2Fv0.146.1/ocb_0.146.1_darwin_arm64
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
	rm -rf ./ocb
