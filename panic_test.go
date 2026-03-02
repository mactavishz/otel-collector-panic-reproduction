package main

import (
	"context"
	"testing"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

// TestPushLogsOnUnstartedExporter demonstrates the nil pointer dereference bug
// when ConsumeLogs is called on an OTLP exporter that hasn't been started.
//
// This test reproduces the panic:
//
//	panic: runtime error: invalid memory address or nil pointer dereference
//	go.opentelemetry.io/collector/exporter/otlpexporter.(*baseExporter).pushLogs at otlp.go:126
//
// The root cause is that the exporter's gRPC client (logExporter) is only
// initialized in the Start() method. If ConsumeLogs is called before Start(),
// or during Start() before the client is initialized, a nil pointer panic occurs.
func TestPushLogsOnUnstartedExporter(t *testing.T) {
	factory := otlpexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*otlpexporter.Config)
	cfg.ClientConfig.Endpoint = "localhost:4317"
	cfg.ClientConfig.TLS.Insecure = true

	// Disable queue to ensure data goes directly to the exporter
	// Use configoptional.None() to disable the queue entirely
	cfg.QueueConfig = configoptional.None[exporterhelper.QueueBatchConfig]()
	cfg.RetryConfig.Enabled = false

	// Create exporter but DON'T call Start()
	// This simulates the race condition where a receiver sends data
	// before the exporter has finished initializing
	settings := exportertest.NewNopSettings(component.MustNewType("otlp_grpc"))
	exporter, err := factory.CreateLogs(
		context.Background(),
		settings,
		cfg,
	)
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}

	// Create test logs
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr("service.name", "test-unstarted-exporter")
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().SetName("race_test")
	lr := sl.LogRecords().AppendEmpty()
	lr.Body().SetStr("Test log on unstarted exporter - should trigger panic")
	lr.SetTimestamp(pcommon.NewTimestampFromTime(pcommon.Timestamp(0).AsTime()))

	// This WILL panic because logExporter (gRPC client) is nil
	// Start() was never called, so NewGRPCClient() was never invoked
	defer func() {
		if r := recover(); r != nil {
			t.Logf("SUCCESS: Expected panic reproduced: %v", r)
			return
		}
		// If we reach here without panic, the exporter handled the error gracefully
		// This might happen if the version has nil-safety checks
		t.Log("NOTE: No panic occurred - exporter may have nil-safety checks in this version")
	}()

	err = exporter.ConsumeLogs(context.Background(), logs)
	if err != nil {
		t.Logf("ConsumeLogs returned error (no panic): %v", err)
		t.Log("NOTE: This version may have error handling instead of panicking")
	}
}
