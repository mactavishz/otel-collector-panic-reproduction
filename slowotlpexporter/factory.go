package slowotlpexporter

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
)

const typeStr = "slowotlp"

// NewFactory creates a factory for the slow OTLP exporter.
// This exporter wraps the standard OTLP exporter but delays its
// initialization to reproduce race conditions during startup.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		exporter.WithLogs(createLogsExporter, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		StartupDelay: 500 * time.Millisecond, // Default delay to trigger race
		ClientConfig: configgrpc.ClientConfig{
			Endpoint: "localhost:4317",
			TLS: configtls.ClientConfig{
				Insecure: true,
			},
		},
	}
}

func createLogsExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {
	sCfg := cfg.(*Config)

	// Build OTLP config from our config fields
	otlpCfg := &otlpexporter.Config{}
	otlpCfg.TimeoutConfig = sCfg.TimeoutConfig
	otlpCfg.QueueConfig = sCfg.QueueConfig
	otlpCfg.RetryConfig = sCfg.RetryConfig
	otlpCfg.ClientConfig = sCfg.ClientConfig

	// Create the underlying OTLP exporter using the OTLP factory
	// We need to modify the settings to use OTLP's component ID because
	// the OTLP factory validates that set.ID.Type() matches its factory type
	otlpFactory := otlpexporter.NewFactory()
	otlpSet := set
	otlpSet.ID = component.NewID(otlpFactory.Type()) // Use "otlp" as the component ID

	otlpExporter, err := otlpFactory.CreateLogs(ctx, otlpSet, otlpCfg)
	if err != nil {
		return nil, err
	}

	return &SlowOTLPExporter{
		wrapped:      otlpExporter,
		startupDelay: sCfg.StartupDelay,
		logger:       set.Logger,
	}, nil
}
