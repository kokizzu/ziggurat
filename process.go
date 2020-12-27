package ziggurat

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func kafkaProcessor(msg *kafka.Message, route string, c *kafka.Consumer, h Handler, l StructuredLogger, ctx context.Context) {
	event := CreateMessageEvent(msg.Value, map[string]string{HeaderMessageType: "kafka", HeaderMessageRoute: route}, msg.Timestamp, ctx)
	h.HandleEvent(event)
	err := storeOffsets(c, msg.TopicPartition)
	l.Error("error storing offsets: %v", err, nil)
}
