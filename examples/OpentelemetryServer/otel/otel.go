package otel

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	gp "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var resource *sdkresource.Resource
var initResourcesOnce sync.Once

func initResource(serviceName string) *sdkresource.Resource {
	initResourcesOnce.Do(func() {
		extraResources, _ := sdkresource.New(
			context.Background(),
			sdkresource.WithOS(),
			sdkresource.WithProcess(),
			sdkresource.WithContainer(),
			sdkresource.WithHost(),
			sdkresource.WithFromEnv(),
			sdkresource.WithAttributes(semconv.ServiceName(serviceName)),
		)
		resource, _ = sdkresource.Merge(
			sdkresource.Default(),
			extraResources,
		)
	})
	return resource
}

func newOtlpExporter() (sdktrace.SpanExporter, error) {
	ctx := context.Background()
	return otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
}

func newStdoutExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

func newZipkinExporter(serviceNameKey string) (sdktrace.SpanExporter, error) {
	url := "http://localhost:19411/api/v2/spans"
	return zipkin.New(url, zipkin.WithLogger(log.New(os.Stderr, serviceNameKey, log.Ldate|log.Ltime|log.Llongfile)))
}

func newJaegerExporter() (sdktrace.SpanExporter, error) {
	url := "http://localhost:14268/api/traces"
	return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
}

func NewTracerProvider(serviceName, exporterTyp string) *sdktrace.TracerProvider {
	var (
		exporter sdktrace.SpanExporter
		err      error
	)
	switch exporterTyp {
	case "stdout":
		exporter, err = newStdoutExporter()
	case "zipkin":
		exporter, err = newZipkinExporter(serviceName)
	case "jaeger":
		exporter, err = newJaegerExporter()
	case "oltphttp":
		exporter, err = otlptracehttp.New(context.Background(), otlptracehttp.WithInsecure())
	default: // otlp
		exporter, err = newOtlpExporter()
	}
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(
		//sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.5))), // 控制采样
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(initResource(serviceName)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

func NewMeterProvider(serviceName, exporterTyp string) *metric.MeterProvider {
	var (
		exporter metric.Exporter
		err      error

		pexporter *prometheus.Exporter
	)
	switch exporterTyp {
	case "stdout":
		exporter, err = stdoutmetric.New()
	case "prometheus":
		registry := gp.NewRegistry()
		pexporter, err = prometheus.New(prometheus.WithRegisterer(registry))
	case "oltphttp":
		exporter, err = otlpmetrichttp.New(context.Background(), otlpmetrichttp.WithInsecure())
	default: // otlp
		exporter, err = otlpmetricgrpc.New(context.Background(), otlpmetricgrpc.WithInsecure())
	}
	if err != nil {
		log.Fatal(err)
	}

	var mp *metric.MeterProvider
	switch exporterTyp {
	case "prometheus":
		mp = metric.NewMeterProvider(metric.WithResource(initResource(serviceName)), metric.WithReader(pexporter))
	default:
		// Register the exporter with an SDK via a periodic reader.
		read := metric.NewPeriodicReader(exporter, metric.WithInterval(1*time.Second))
		mp = metric.NewMeterProvider(metric.WithResource(initResource(serviceName)), metric.WithReader(read))
	}
	otel.SetMeterProvider(mp)
	return mp
}
