package ziggurat

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
	"sync"
)

type topicEntity struct {
	handlerFunc      HandlerFunc
	consumers        []*kafka.Consumer
	bootstrapServers string
	originTopics     []string
}

type TopicEntityHandlerMap = map[string]*topicEntity

type StreamRouter struct {
	handlerFunctionMap TopicEntityHandlerMap
}

func newConsumerConfig() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers":        "localhost:9092",
		"group.id":                 "myGroup",
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  2000,
		"debug":                    "consumer,broker",
		"enable.auto.offset.store": false,
	}
}

func NewStreamRouter() *StreamRouter {
	return &StreamRouter{
		handlerFunctionMap: make(map[string]*topicEntity),
	}
}

func (sr *StreamRouter) GetHandlerFunctionMap() map[string]*topicEntity {
	return sr.handlerFunctionMap
}

func (sr *StreamRouter) GetTopicEntities() []*topicEntity {
	var topicEntities []*topicEntity
	for _, te := range sr.handlerFunctionMap {
		topicEntities = append(topicEntities, te)
	}
	return topicEntities
}

func (sr *StreamRouter) HandlerFunc(topicEntityName string, handlerFn HandlerFunc) {
	sr.handlerFunctionMap[topicEntityName] = &topicEntity{handlerFunc: handlerFn}
}

func makeKV(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func notifyRouterStop(stopChannel chan<- int, wg *sync.WaitGroup) {
	wg.Wait()
	close(stopChannel)
}

func (sr *StreamRouter) Start(ctx context.Context, app App) chan int {
	stopNotifierCh := make(chan int)
	config := app.Config
	var wg sync.WaitGroup
	srConfig := config.StreamRouter
	hfMap := sr.handlerFunctionMap
	if len(hfMap) == 0 {
		RouterLogger.Fatal().Err(ErrNoHandlersRegistered).Msg("")
	}

	for topicEntityName, te := range hfMap {
		streamRouterCfg := srConfig[topicEntityName]
		if topicEntityName != streamRouterCfg.TopicEntity {
			RouterLogger.Fatal().Err(ErrTopicEntityMismatch).Msg("")
		}
		consumerConfig := newConsumerConfig()
		bootstrapServers := makeKV("bootstrap.servers", streamRouterCfg.BootstrapServers)
		groupID := makeKV("group.id", streamRouterCfg.GroupID)
		if setErr := consumerConfig.Set(bootstrapServers); setErr != nil {
			RouterLogger.Error().Err(setErr)
		}
		if setErr := consumerConfig.Set(groupID); setErr != nil {
			RouterLogger.Error().Err(setErr)
		}
		topics := strings.Split(streamRouterCfg.OriginTopics, ",")
		consumers := StartConsumers(ctx, app, consumerConfig, topicEntityName, topics, streamRouterCfg.InstanceCount, te.handlerFunc, &wg)
		te.consumers = consumers
	}

	if config.Retry.Enabled {
		RouterLogger.Info().Msg("starting retrier...")
		if retrierStartErr := app.Retrier.Start(ctx, app); retrierStartErr != nil {
			RouterLogger.Fatal().Err(retrierStartErr).Msg("unable to start retrier")
		}

		app.Retrier.Consume(ctx, app)
		RouterLogger.Info().Msg("starting retrier consumer")
	}

	app.HttpServer.Start(ctx, app)
	RouterLogger.Info().Msg("http server started...")

	metricStartErr := app.MetricPublisher.Start(ctx, app)
	if metricStartErr != nil {
		RouterLogger.Error().Err(metricStartErr)
	}
	RouterLogger.Info().Msg("starting metrics...")

	go notifyRouterStop(stopNotifierCh, &wg)

	return stopNotifierCh
}
