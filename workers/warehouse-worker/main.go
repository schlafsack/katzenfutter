package main

import (
	"context"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
	"log"
	"time"
	"warehouse-worker/configuration"
)

func main() {

	log.Println("Starting...")

	c := configuration.New()
	broker := c.GetBrokerEndpoint()
	log.Println("Using broker:", broker)

	zbClient, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         broker,
		UsePlaintextConnection: true})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	for {

		time.Sleep(10 * time.Millisecond)
	}
}
