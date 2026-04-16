// Pacote logger fornece utilitários de logging estruturado para a aplicação.
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/gustavoz65/go-backend-boilerplate/backend/internal/config"
)

type LoggerService struct {
	nrApp *newrelic.Application
}

func NewLoggerService(cfg *config.ObservabilityConfig) *LoggerService {
	service := &LoggerService{}

	if cfg.NewRelic.LicenseKey == "" {
		fmt.Println("New Relic license key is not provided. Skipping New Relic initialization.")
		return service
	}

	var configOptions []newrelic.ConfigOption
	configOptions = append(configOptions,
		newrelic.ConfigAppName(cfg.ServiceName),
		newrelic.ConfigLicense(cfg.NewRelic.LicenseKey),
		newrelic.ConfigAppLogForwardingEnabled(cfg.NewRelic.AppLogForwardingEnabled),
		newrelic.ConfigDistributedTracerEnabled(cfg.NewRelic.DistributedTracingEnabled),
	)

	if cfg.NewRelic.DebugLogging {
		configOptions = append(configOptions, newrelic.ConfigDebugLogger(os.Stdout))
	}

	app, err := newrelic.NewApplication(configOptions...)
	if err != nil {
		fmt.Printf("Failed to initialize New Relic application: %v\n", err)
		return service
	}

	service.nrApp = app
	fmt.Printf("New Relic application initialized successfully: %s\n", cfg.ServiceName)
	return service
}

func (ls *LoggerService) Shutdown() {
	if ls.nrApp != nil {
		ls.nrApp.Shutdown(10 * time.Second)
	}
}

func (ls *LoggerService) GetApplication() *newrelic.Application {
	return ls.nrApp
}

func Newlogger(level string, isProd bool) zerolog.Logger {
	return NewLoggerWithService(&config.ObservabilityConfig{
		Logging: config.LoggingConfig{
			Level: level,
		},
		Environment: func() string {
			if isProd {
				return "production"
			}
			return "development"
		}(),
	}, nil)
}

func NewLoggerWithConfig(cfg *config.ObservabilityConfig) zerolog.Logger {
	return NewLoggerWithService(cfg, nil)
}

func NewLoggerWithService(cfg *config.ObservabilityConfig, loggerService *LoggerService) zerolog.Logger {
	_ = loggerService

	var logLevel zerolog.Level
	level := cfg.GetLogLevel()

	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var writer io.Writer

	// setup básico de escrita
	var baseWriter io.Writer
	if cfg.IsProduction() && cfg.Logging.Format == "json" {
		baseWriter = os.Stdout

		writer = baseWriter
	} else {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
		writer = consoleWriter
	}

	logger := zerolog.New(writer).
		Level(logLevel).
		With().Timestamp().
		Str("Service", cfg.ServiceName).
		Str("environment", cfg.Environment).
		Logger()

	if !cfg.IsProduction() {
		logger = logger.With().Stack().Logger()
	}

	return logger
}

func WithTraceContext(logger zerolog.Logger, txn *newrelic.Transaction) zerolog.Logger {
	if txn == nil {
		return logger
	}

	metadata := txn.GetTraceMetadata()

	return logger.With().
		Str("trace.id", metadata.TraceID).
		Str("span.id", metadata.SpanID).
		Logger()
}

func NewPgxLogger(level zerolog.Level) zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatFieldValue: func(i any) string {
			switch v := i.(type) {
			case string:
				// limpando strings muito longas para evitar poluição nos logs
				if len(v) > 20 {
					// corta para os primeiros 200 caracteres
					return v[:200] + "..."
				}
				return v
			case []byte:
				var obj interface{}
				if err := json.Unmarshal(v, &obj); err == nil {
					pretty, _ := json.MarshalIndent(obj, "", "  ")
					return "\n" + string(pretty)
				}
				return string(v)
			default:
				return fmt.Sprintf("%v", v)
			}
		},
	}

	return zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Str("component", "pgx").
		Logger()
}

func GetPgxTraceLogLevel(level zerolog.Level) int {
	switch level {
	case zerolog.DebugLevel:
		return 6
	case zerolog.InfoLevel:
		return 4
	case zerolog.WarnLevel:
		return 3
	case zerolog.ErrorLevel:
		return 2
	default:
		return 0
	}
}
