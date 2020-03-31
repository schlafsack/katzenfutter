package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
	"log"
	"order-source-worker/configuration"
	"time"
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
		orderId := uuid.New().String()
		log.Println("New order ", orderId)
		variables := make(map[string]interface{})
		variables["order_id"] = orderId
		request, err := zbClient.NewCreateInstanceCommand().BPMNProcessId("order_process").
			LatestVersion().VariablesFromMap(variables)
		if err != nil {
			panic(err)
		}
		_, err = request.Send(ctx)
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Duration(c.GetFrequency()) * time.Second)
	}
}
