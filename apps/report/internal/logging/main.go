package logging

import (
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func GetRootLogger() *zap.SugaredLogger {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return logger.Sugar()
}

func WithSpanContext(sugaredLogger *zap.SugaredLogger, span trace.Span) *zap.SugaredLogger {
	return sugaredLogger.With("traceID", span.SpanContext().TraceID(), "spanID", span.SpanContext().SpanID())
}
