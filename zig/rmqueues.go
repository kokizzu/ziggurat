package zig

import (
	"fmt"
	"github.com/streadway/amqp"
)

const QueueTypeDelay = "delay"
const QueueTypeInstant = "instant"
const QueueTypeDL = "dead_letter"

func declareExchanges(c *amqpsafe.Connector, topicEntities []string, serviceName string) {
	exchangeTypes := []string{QueueTypeInstant, QueueTypeDelay, QueueTypeDL}
	for _, te := range topicEntities {
		for _, et := range exchangeTypes {
			exchName := constructExchangeName(serviceName, te, et)
			logInfo("rmq queues: creating exchange", map[string]interface{}{"exchange-name": exchName})
			c.ExchangeDeclare(exchName, amqpsafe.ExchangeFanout, true, false, false, false, nil)
		}
	}

}

func createAndBindQueue(c *amqpsafe.Connector, queueName string, exchangeName string, args amqp.Table) error {
	_, queueErr := c.QueueDeclare(queueName, true, false, false, false, args)
	if queueErr != nil {
		return queueErr
	}
	logInfo("rmq queues: binding queue to exchange", map[string]interface{}{
		"queue-name":    queueName,
		"exchange-name": exchangeName,
	})
	bindErr := c.QueueBind(queueName, "", exchangeName, false, nil)
	return bindErr
}

func constructQueueName(serviceName string, topicEntity string, queueType string) string {
	return fmt.Sprintf("%s_%s_%s_queue", topicEntity, serviceName, queueType)
}

func constructExchangeName(serviceName string, topicEntity string, exchangeType string) string {
	return fmt.Sprintf("%s_%s_%s_exchange", topicEntity, serviceName, exchangeType)
}

func createInstantQueues(c *amqpsafe.Connector, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, QueueTypeInstant)
		exchangeName := constructExchangeName(serviceName, te, QueueTypeInstant)
		bindErr := createAndBindQueue(c, queueName, exchangeName, nil)
		logError(bindErr, "rmq queues: error binding queue", nil)

	}
}

func createDelayQueues(c *amqpsafe.Connector, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, QueueTypeDelay)
		exchangeName := constructExchangeName(serviceName, te, QueueTypeDelay)
		deadLetterExchangeName := constructExchangeName(serviceName, te, QueueTypeInstant)
		args := amqp.Table{
			"x-dead-letter-exchange": deadLetterExchangeName,
		}
		bindErr := createAndBindQueue(c, queueName, exchangeName, args)
		logError(bindErr, "rmq queues: error binding queue", nil)
	}
}

func createDeadLetterQueues(c *amqpsafe.Connector, topicEntities []string, serviceName string) {
	for _, te := range topicEntities {
		queueName := constructQueueName(serviceName, te, QueueTypeDL)
		exchangeName := constructExchangeName(serviceName, te, QueueTypeDL)
		bindErr := createAndBindQueue(c, queueName, exchangeName, nil)
		logError(bindErr, "rmq queues: error binding queue", nil)
	}
}
