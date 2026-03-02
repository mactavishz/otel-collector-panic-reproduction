package immediatereceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type ImmediateReceiver struct {
	logger      *zap.Logger
	logConsumer consumer.Logs
	startTime   time.Time
}

func NewImmediateReceiver(logger *zap.Logger, logConsumer consumer.Logs) *ImmediateReceiver {
	return &ImmediateReceiver{
		logger:      logger,
		logConsumer: logConsumer,
	}
}

func (r *ImmediateReceiver) Start(ctx context.Context, host component.Host) error {
	r.startTime = time.Now()
	r.logger.Info("ImmediateReceiver: Starting - will send data SYNCHRONOUSLY",
		zap.Time("start_time", r.startTime))

	// SYNCHRONOUS call - NOT in a goroutine
	// This BLOCKS the Start() call and guarantees the race condition
	// because the OTLP exporter may not have finished initializing its gRPC client
	r.logger.Info("ImmediateReceiver: About to send logs to consumer SYNCHRONOUSLY",
		zap.Duration("time_since_start", time.Since(r.startTime)),
		zap.String("warning", "PANIC EXPECTED - EXPORTER NOT READY"))

	// Create a simple log record
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr("service.name", "repro-demo")
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().SetName("immediatereceiver")
	lr := sl.LogRecords().AppendEmpty()
	lr.Body().SetStr("Synchronous log sent during Start() - testing race condition")
	lr.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

	// This call WILL trigger the race condition if exporter isn't ready
	if err := r.logConsumer.ConsumeLogs(ctx, logs); err != nil {
		r.logger.Error("ImmediateReceiver: Failed to consume logs", zap.Error(err))
		return err // Return error to fail startup
	}

	r.logger.Info("ImmediateReceiver: Successfully sent logs",
		zap.Duration("time_since_start", time.Since(r.startTime)))
	return nil
}

func (r *ImmediateReceiver) Shutdown(ctx context.Context) error {
	r.logger.Info("ImmediateReceiver: Shutting down")
	return nil
}
