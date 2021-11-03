// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"kafka-example/internal/config"
	"kafka-example/internal/logging"
	"kafka-example/internal/otelWrapper"
	"strings"
	"time"

	"github.com/Shopify/sarama"

	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
)

func main() {
	tp := otelWrapper.InitTracer("consumer")
	logger := logging.GetRootLogger()
	defer logger.Sync()

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Errorw("Error shutting down tracer provider", "error", err)
		}
	}()

	appConfig := config.GetConfig()
	logger.Infof("Kafka brokers: %s", strings.Join(appConfig.Brokers, ", "))

	startConsumerGroup(appConfig, logger)

	select {}
}

func startConsumerGroup(appConfig *config.Config, logger *zap.SugaredLogger) {
	consumerGroupHandler := Consumer{
		Logger: logger,
	}
	// Wrap instrumentation
	handler := otelsarama.WrapConsumerGroupHandler(&consumerGroupHandler)

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = config.KafkaVersion
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(appConfig.Brokers, config.KafkaConsumerGroupID, saramaConfig)
	if err != nil {
		logger.Fatalw("Failed to start sarama consumer group", "error", err)
	}

	err = consumerGroup.Consume(context.Background(), []string{config.KafkaTopic}, handler)
	if err != nil {
		logger.Fatalw("Failed to consume via handler", "error", err)
	}
}

func printMessage(msg *sarama.ConsumerMessage, logger *zap.SugaredLogger) {
	// Extract tracing info from message
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), otelsarama.NewConsumerMessageCarrier(msg))

	tr := otel.Tracer("consumer")
	_, span := tr.Start(ctx, "consume message", trace.WithAttributes(
		semconv.MessagingOperationProcess,
	))
	defer span.End()

	spanLogger := logging.WithSpanContext(logger, span)

	// Emulate Work loads
	time.Sleep(500 * time.Microsecond)

	message := string(msg.Value)
	if message == "consumer_error" { // throw custom error if a custom message is received
		err := fmt.Errorf("custom customer error")
		span.RecordError(err)
		span.SetStatus(codes.Error, "custom_customer_error")
		spanLogger.Errorw("custom customer error", "error", err)
	} else {
		spanLogger.Infof("Successful to read message: %s", message)
	}
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	Logger *zap.SugaredLogger
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		printMessage(message, consumer.Logger)
		session.MarkMessage(message, "")
	}

	return nil
}
