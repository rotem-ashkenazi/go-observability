package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	logexp "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	otellog "go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	// OTLP/collector address, e.g. "otel-collector:4317". If empty, uses OTEL_EXPORTER_OTLP_ENDPOINT or "localhost:4317".
	Endpoint string
	// If true, uses plaintext (inside cluster / local). If false, use TLS (you'll need creds).
	Insecure bool

	ServiceName string // required
	ServiceVer  string // optional
	Environment string // "prod" | "staging" | "dev" | etc.

	// Optional tuning:
	DialTimeout    time.Duration // default 10s
	ExportInterval time.Duration // default 2s
	MaxQueueSize   int           // default 4096
}

// InitLogs initializes a global OTel LoggerProvider. Call once at startup.
// Returns a shutdown func you should call on exit for a clean flush.
func InitLogs(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if cfg.ServiceName == "" {
		return nil, errors.New("telemetry: ServiceName is required")
	}

	if cfg.Endpoint == "" {
		// Also respects the standard env var if you forgot to pass Endpoint.
		cfg.Endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		if cfg.Endpoint == "" {
			cfg.Endpoint = "localhost:4317"
		}
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 10 * time.Second
	}
	if cfg.ExportInterval == 0 {
		cfg.ExportInterval = 2 * time.Second
	}
	if cfg.MaxQueueSize == 0 {
		cfg.MaxQueueSize = 4096
	}

	// Build a shared resource. Merges OTEL_RESOURCE_ATTRIBUTES automatically.
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVer),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	bo := backoff.Config{
		BaseDelay:  500 * time.Millisecond,
		Multiplier: 1.6,
		MaxDelay:   5 * time.Second,
	}
	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           bo,
			MinConnectTimeout: cfg.DialTimeout,
		}),
	}
	if cfg.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	exp, err := logexp.New(ctx,
		logexp.WithEndpoint(cfg.Endpoint),
		logexp.WithDialOption(dialOpts...),
	)
	if err != nil {
		return nil, fmt.Errorf("otlp log exporter: %w", err)
	}

	bp := sdklog.NewBatchProcessor(exp,
		sdklog.WithExportInterval(cfg.ExportInterval),
		sdklog.WithMaxQueueSize(cfg.MaxQueueSize),
	)

	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(bp),
	)
	logglobal.SetLoggerProvider(lp)

	return lp.Shutdown, nil
}

// Logger returns a named component logger (e.g., "http", "db", "worker").
func Logger(name string) otellog.Logger {
	return logglobal.GetLoggerProvider().Logger(name)
}

// Helper line-level emitters (optional sugar).
func Info(ctx context.Context, l otellog.Logger, msg string, attrs ...otellog.KeyValue) {
	var r otellog.Record
	r.SetTimestamp(time.Now())
	r.SetSeverity(otellog.SeverityInfo)
	r.SetBody(otellog.StringValue(msg))
	for _, a := range attrs {
		r.AddAttributes(a)
	}
	l.Emit(ctx, r)
}

func Error(ctx context.Context, l otellog.Logger, msg string, attrs ...otellog.KeyValue) {
	var r otellog.Record
	r.SetTimestamp(time.Now())
	r.SetSeverity(otellog.SeverityError)
	r.SetBody(otellog.StringValue(msg))
	for _, a := range attrs {
		r.AddAttributes(a)
	}
	l.Emit(ctx, r)
}

func Warn(ctx context.Context, l otellog.Logger, msg string, attrs ...otellog.KeyValue) {
	var r otellog.Record
	r.SetTimestamp(time.Now())
	r.SetSeverity(otellog.SeverityWarn)
	r.SetBody(otellog.StringValue(msg))
	for _, a := range attrs {
		r.AddAttributes(a)
	}
	l.Emit(ctx, r)
}
