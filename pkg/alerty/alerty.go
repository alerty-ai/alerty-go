package alerty

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	DefaultAlertyIngestURL = "https://ingest.alerty.ai"

	AlertyOrganizationIDKey = attribute.Key("alerty.organizationId")
)

type AlertyServiceConfig struct {
	// required configuration
	OrganizationID string
	Name           string
	Version        string
	Environment    string

	// optional configuration
	IngestURL string
	Debug     bool // turn on debug mode to see the collected data in stderr
}

var alertyTraceProvider *sdktrace.TracerProvider
var alertyServiceConfig *AlertyServiceConfig

// Start initializes and starts the Alerty service.
func Start(cfg AlertyServiceConfig) error {
	alertyServiceConfig = &cfg

	ctx := context.Background()

	ingestURL := alertyServiceConfig.IngestURL
	if ingestURL == "" {
		ingestURL = DefaultAlertyIngestURL
	}

	endpoint, err := url.Parse(ingestURL)
	if err != nil {
		return errors.WithStack(err)
	}

	otlpExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint.String()),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(alertyServiceConfig.Name),
			semconv.ServiceVersionKey.String(alertyServiceConfig.Version),
			semconv.DeploymentEnvironmentKey.String(alertyServiceConfig.Environment),
			AlertyOrganizationIDKey.String(alertyServiceConfig.OrganizationID),
		),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(otlpExporter),
	}

	if alertyServiceConfig.Debug {
		stdoutExporter, err := stdouttrace.New(
			stdouttrace.WithWriter(os.Stderr),
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return errors.WithStack(err)
		}
		opts = append(opts, sdktrace.WithBatcher(stdoutExporter))
	}

	traceProvider := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(traceProvider)
	alertyTraceProvider = traceProvider

	return nil
}

// Stop shuts down the Alerty service and flushes any remaining events.
func Stop() {
	if alertyTraceProvider == nil {
		return
	}

	if r := recover(); r != nil {
		CapturePanic(r)
	}

	// Shutdown the tracer provider, ensuring all spans are flushed.
	if err := alertyTraceProvider.Shutdown(context.Background()); err != nil {
		log.Printf("Error shutting down tracer provider: %v\n", err)
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// CaptureError captures and reports an error using OpenTelemetry.
func CaptureError(err error) {
	if err == nil {
		return
	}

	tracer := otel.Tracer("alerty-go")
	_, span := tracer.Start(context.Background(), "error")
	defer span.End()

	opts := []trace.EventOption{}
	if stack, ok := err.(stackTracer); ok {
		// this captures the stack of where the stackTracer compatible error was created
		opts = append(opts, trace.WithAttributes(semconv.ExceptionStacktrace(fmt.Sprintf("%+v", stack.StackTrace()))))
	} else {
		// this only capture the stack of where the error is captured not where it was created
		opts = append(opts, trace.WithStackTrace(true))
	}

	span.RecordError(err, opts...)
	span.SetStatus(codes.Error, err.Error())
}

// CapturePanic captures and reports a panic using OpenTelemetry.
func CapturePanic(r interface{}) {
	tracer := otel.Tracer("alerty-go")
	_, span := tracer.Start(context.Background(), "panic")
	defer span.End()

	var err error
	switch x := r.(type) {
	case string:
		err = errors.New(x)
	case error:
		err = x
	default:
		err = errors.New("unknown panic")
	}

	opts := []trace.EventOption{}
	if stack, ok := err.(stackTracer); ok {
		// this captures the stack of where the stackTracer compatible error was created
		opts = append(opts, trace.WithAttributes(semconv.ExceptionStacktrace(fmt.Sprintf("%+v", stack.StackTrace()))))
	} else {
		// this only capture the stack of where the error is captured not where it was created
		opts = append(opts, trace.WithStackTrace(true))
	}

	span.RecordError(err, opts...)
	span.SetStatus(codes.Error, err.Error())
}

func Recover(handlers ...func(r interface{})) {
	if r := recover(); r != nil {
		CapturePanic(r)
		if len(handlers) > 0 {
			for _, handler := range handlers {
				handler(r)
			}
		}
	}
}
