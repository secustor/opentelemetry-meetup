package main

import (
	"context"
	"github.com/Shopify/sarama"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"io"
	"kafka-example/internal/config"
	"kafka-example/internal/logging"
	"kafka-example/internal/otelWrapper"
	"net/http"
	"strings"
)

func main() {
	tp := otelWrapper.InitTracer("producer")
	logger := logging.GetRootLogger()
	defer logger.Sync()
	appConfig := config.GetConfig()

	logger.Infof("Kafka brokers: %s", strings.Join(appConfig.Brokers, ", "))
	producer, err := newAccessLogProducer(appConfig, logger)
	if err != nil {
		logger.Fatalw("Failed to start Sarama producer", "error", err)
	}

	defer func() {
		err := producer.Close()
		if err != nil {
			logger.Errorw("Failed to close producer", "error", err)
		}

		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Errorw("Shutting down tracer provider", err)
		}
	}()

	kafkaHandler := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		span := trace.SpanFromContext(ctx)
		defer span.End()
		spanLogger := logging.WithSpanContext(logger, span)
		//bag := baggage.FromContext(ctx)
		//span.AddEvent("handling this...", trace.WithAttributes(uk.String(bag.Member("username").Value())))

		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			spanLogger.Errorw("Failed to request body", "error", err)
		}

		body := string(bodyBytes)

		if strings.Contains(body, "producer_error") { // throw custom error if a custom message is received
			span.SetStatus(codes.Error, "custom producer error")
			spanLogger.Error("Custom producer error")
		} else {
			// Inject tracing info into message
			msg := sarama.ProducerMessage{
				Topic: config.KafkaTopic,
				Key:   sarama.StringEncoder(config.KafkaKey),
				Value: sarama.StringEncoder(body),
			}

			otel.GetTextMapPropagator().Inject(ctx, otelsarama.NewProducerMessageCarrier(&msg))

			producer.Input() <- &msg
			successMsg := <-producer.Successes()
			spanLogger.Infow("Successful to write message", "offset", successMsg.Offset)
		}

		_, err = io.WriteString(w, "{\"msg\": \"Hello, world!\"}")
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			spanLogger.Errorw("Failed to write answer", "error", err)
		}
	}

	otelHandler := otelhttp.NewHandler(http.HandlerFunc(kafkaHandler), "http_server")
	http.Handle("/kafka/receiver", otelHandler)

	statusHandler := func(w http.ResponseWriter, req *http.Request) {}
	http.Handle("/healthz", http.HandlerFunc(statusHandler))
	err = http.ListenAndServe(":7777", nil)
	if err != nil {
		panic(err)
	}
}

func newAccessLogProducer(appConfig *config.Config, logger *zap.SugaredLogger) (sarama.AsyncProducer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = config.KafkaVersion
	// So we can know the partition and offset of messages.
	saramaConfig.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer(appConfig.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	// Wrap instrumentation
	producer = otelsarama.WrapAsyncProducer(saramaConfig, producer)

	// We will log to STDOUT if we're not able to produce messages.
	go func() {
		for err := range producer.Errors() {
			logger.Errorw("Failed to write message", "error", err)
		}
	}()

	return producer, nil
}
