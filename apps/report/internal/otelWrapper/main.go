package otelWrapper

import (
	"context"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"log"
)

func InitTracer() *sdktrace.TracerProvider {
	client := otlptracehttp.NewClient()
	exporter, err := otlptrace.New(context.Background(), client)

	if err != nil {
		log.Fatal(err)
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceNameKey.String("producer"),
			semconv.ServiceNamespaceKey.String("example.meetup"),
			semconv.ServiceVersionKey.String("0.1.0"),
		)),
	)
	otel.SetTracerProvider(tp)
	// add propagation for jaeger, W3C and W3C baggage
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(jaeger.Jaeger{}, propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}
