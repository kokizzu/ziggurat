package ziggurat

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"time"
)

var consumerLogContext = map[string]interface{}{"component": "consumer"}

var createConsumer = func(consumerConfig *kafka.ConfigMap, topics []string) *kafka.Consumer {
	consumer, err := kafka.NewConsumer(consumerConfig)
	LogError(err, "ziggurat consumer", consumerLogContext)
	subscribeErr := consumer.SubscribeTopics(topics, nil)
	LogError(subscribeErr, "ziggurat consumer", consumerLogContext)
	return consumer
}

var storeOffsets = func(consumer *kafka.Consumer, partition kafka.TopicPartition) error { // at least once delivery
	if partition.Error != nil {
		return ErrOffsetCommit
	}
	offsets := []kafka.TopicPartition{partition}
	offsets[0].Offset++
	if _, err := consumer.StoreOffsets(offsets); err != nil {
		return err
	}
	return nil
}

var readMessage = func(c *kafka.Consumer, pollTimeout time.Duration) (*kafka.Message, error) {
	return c.ReadMessage(pollTimeout)
}

func NewConsumerConfig(bootstrapServers string, groupID string) *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers":        bootstrapServers,
		"group.id":                 groupID,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  2000,
		"debug":                    "consumer,broker",
		"enable.auto.offset.store": false,
		//disable for at-least once delivery
	}
}