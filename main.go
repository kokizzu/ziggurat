package main

import (
	"fmt"
	"github.com/gojek/ziggurat"
)

type TestEntity struct {
	TestKey   string `json:"key"`
	TestValue string `json:"value"`
}

func main() {

	sr := ziggurat.NewStreamRouter()

	sr.HandlerFunc("test-entity", func(message ziggurat.MessageEvent) ziggurat.ProcessStatus {
		fmt.Printf("[handlerFunc]: Received message for test-entity1 %v\n", message)
		return ziggurat.ProcessingSuccess
	})

	sr.HandlerFunc("test-entity2", func(message ziggurat.MessageEvent) ziggurat.ProcessStatus {
		fmt.Printf("[handlerFunc]: Received message for test-entity2 %v\n", message)
		return ziggurat.RetryMessage
	})

	ziggurat.Start(sr, ziggurat.StartupOptions{
		StartFunction: func(config ziggurat.Config) {
			fmt.Printf("Starting app...\n")
		},
		StopFunction: func() {
			fmt.Printf("Stopping app...\n")
		},
	})

}
