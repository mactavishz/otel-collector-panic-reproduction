package slowotlpexporter

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

// SlowOTLPExporter wraps the standard OTLP exporter and delays its
// initialization to simulate slow startup conditions in production.
// This creates a race window where ConsumeLogs can be called before
// the wrapped exporter's gRPC client is initialized.
type SlowOTLPExporter struct {
	wrapped      exporter.Logs
	startupDelay time.Duration
	logger       *zap.Logger
}

// Start returns immediately but delays the actual OTLP exporter initialization.
// This simulates slow startup due to DNS/TLS/network issues in production.
//
// The key insight: by returning from Start() immediately but starting the
// wrapped exporter in a goroutine after a delay, we create a window where
// ConsumeLogs can be called on an uninitialized exporter, causing a panic.
func (e *SlowOTLPExporter) Start(ctx context.Context, host component.Host) error {
	e.logger.Info("SlowOTLPExporter: Start() called - will delay wrapped exporter startup",
		zap.Duration("startup_delay", e.startupDelay))

	// IMPORTANT: Start the wrapped exporter in a goroutine AFTER a delay.
	// This means ConsumeLogs calls before the delay completes will see
	// a nil gRPC client in the wrapped OTLP exporter.
	go func() {
		e.logger.Info("SlowOTLPExporter: Waiting before starting wrapped exporter...",
			zap.Duration("delay", e.startupDelay))
		time.Sleep(e.startupDelay)
		e.logger.Info("SlowOTLPExporter: Now starting wrapped OTLP exporter")
		if err := e.wrapped.Start(ctx, host); err != nil {
			e.logger.Error("SlowOTLPExporter: Failed to start wrapped exporter", zap.Error(err))
		} else {
			e.logger.Info("SlowOTLPExporter: Wrapped exporter started successfully")
		}
	}()

	// Return immediately - the receiver will start and try to send data
	// while the wrapped exporter's gRPC client is still nil
	e.logger.Info("SlowOTLPExporter: Start() returning immediately (race window is now OPEN)",
		zap.String("warning", "ConsumeLogs calls during the next "+e.startupDelay.String()+" will panic"))
	return nil
}

// ConsumeLogs forwards to the wrapped exporter.
// If called before wrapped.Start() completes, this will panic due to nil gRPC client.
func (e *SlowOTLPExporter) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	e.logger.Debug("SlowOTLPExporter: ConsumeLogs called - forwarding to wrapped exporter",
		zap.Int("log_records", logs.LogRecordCount()))
	return e.wrapped.ConsumeLogs(ctx, logs)
}

// Shutdown shuts down the wrapped exporter
func (e *SlowOTLPExporter) Shutdown(ctx context.Context) error {
	e.logger.Info("SlowOTLPExporter: Shutdown called")
	return e.wrapped.Shutdown(ctx)
}

// Capabilities returns the capabilities of the wrapped exporter
func (e *SlowOTLPExporter) Capabilities() consumer.Capabilities {
	return e.wrapped.Capabilities()
}
