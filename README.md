# OpenTelemetry Collector OTLP Exporter Panic Reproduction

This repository reproduces a **nil pointer dereference panic** in the OpenTelemetry Collector's OTLP gRPC exporter (`otlpexporter`) when `ConsumeLogs()` is called before the exporter's `Start()` has completed.

## The Bug

The OTLP exporter's internal gRPC client is only initialized during `Start()`. If data arrives before initialization completes, the nil client triggers a panic:

```
panic: runtime error: invalid memory address or nil pointer dereference
    go.opentelemetry.io/collector/exporter/otlpexporter.(*baseExporter).pushLogs at otlp.go:126
```

This can happen in practice when a receiver emits data quickly during startup and the exporter is slow to initialize (e.g. due to DNS resolution, TLS handshakes, or network latency).

## How It Works

Two custom components create the conditions for the panic:

- **`immediatereceiver`** — A receiver that sends log data synchronously inside its `Start()` method, before the collector has finished starting all components.
- **`slowotlpexporter`** — A wrapper around the real OTLP exporter that delays the inner exporter's `Start()` by 500ms, creating a window where the gRPC client is nil.

The sending queue and retry must both be disabled so data flows directly to the uninitialized exporter.

## Prerequisites

- Go 1.25.0+
- make
- curl

## Running the Reproduction

```bash
# Build the custom collector (downloads OCB, generates code, compiles)
make build

# Run it — the collector will panic during startup
make run
```

No external OTLP backend is needed; the panic occurs before any network connection is attempted.
