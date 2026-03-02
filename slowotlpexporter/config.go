package slowotlpexporter

import (
	"time"

	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines configuration for the slow OTLP exporter wrapper.
// This exporter delays its initialization to simulate slow startup
// conditions that occur in production (DNS resolution, TLS loading, etc.)
type Config struct {
	// StartupDelay delays the actual OTLP exporter initialization.
	// This simulates slow DNS/TLS/network in production environments.
	// During this delay, ConsumeLogs calls will hit a nil gRPC client.
	StartupDelay time.Duration `mapstructure:"startup_delay"`

	// OTLP-compatible configuration fields (copied from otlpexporter.Config)
	// We don't embed otlpexporter.Config because its Validate() checks component ID
	TimeoutConfig exporterhelper.TimeoutConfig                             `mapstructure:",squash"`
	QueueConfig   configoptional.Optional[exporterhelper.QueueBatchConfig] `mapstructure:"sending_queue"`
	RetryConfig   configretry.BackOffConfig                                `mapstructure:"retry_on_failure"`
	ClientConfig  configgrpc.ClientConfig                                  `mapstructure:",squash"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	return c.ClientConfig.Validate()
}
