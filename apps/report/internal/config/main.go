package config

import (
	"flag"
	"github.com/Shopify/sarama"
	"os"
	"strings"
)

var (
	bootStrapServer      = flag.String("bootStrapServer", os.Getenv("KAFKA_BOOTSTRAP"), "The Kafka bootStrapServer to connect to, as a comma separated list")
	KafkaTopic           = "opentelemetry-meetup"
	KafkaKey             = "snapshot_kafka_key"
	KafkaConsumerGroupID = "example"
	KafkaVersion         = sarama.V3_0_0_0

	config *Config
)

type Config struct {
	Brokers []string
}

func GetConfig() *Config {
	if config != nil {
		return config
	}
	flag.Parse()

	if *bootStrapServer == "" {
		flag.PrintDefaults()
	}

	brokerList := strings.Split(*bootStrapServer, ",")
	return &Config{
		Brokers: brokerList,
	}
}
