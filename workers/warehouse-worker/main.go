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
		//log.Println("New order...")
		_, err = zbClient.NewCreateInstanceCommand().BPMNProcessId("test_process").LatestVersion().Send(ctx)
		if err != nil {
			panic(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
