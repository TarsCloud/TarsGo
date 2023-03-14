module OpentelemetryServer

go 1.14

require (
	github.com/TarsCloud/TarsGo v1.3.10
	go.opentelemetry.io/otel v1.14.0
	go.opentelemetry.io/otel/exporters/jaeger v1.11.2
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.14.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.14.0
	go.opentelemetry.io/otel/exporters/zipkin v1.11.2
	go.opentelemetry.io/otel/sdk v1.14.0
	go.opentelemetry.io/otel/trace v1.14.0
)

replace github.com/TarsCloud/TarsGo => ../../
