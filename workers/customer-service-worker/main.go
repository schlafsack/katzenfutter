package main

import (
	"context"
	"customer-service-worker/configuration"
	"github.com/zeebe-io/zeebe/clients/go/pkg/entities"
	"github.com/zeebe-io/zeebe/clients/go/pkg/worker"
	"github.com/zeebe-io/zeebe/clients/go/pkg/zbc"
	"log"
)

func main() {

	log.Println("Starting order-service-worker...")

	c := configuration.New()
	broker := c.GetBrokerEndpoint()
	log.Println("Using broker:", broker)

	zbClient, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         broker,
		UsePlaintextConnection: true})
	if err != nil {
		panic(err)
	}

	failedPickWorker := zbClient.NewJobWorker().JobType("notify_pick_fail_task").Handler(handlePickFail).Open()
	defer failedPickWorker.Close()

	pickSuccessWorker := zbClient.NewJobWorker().JobType("notify_pick_success_task").Handler(handlePickSuccess).Open()
	defer pickSuccessWorker.Close()

	failedPickWorker.AwaitClose()
	pickSuccessWorker.AwaitClose()
}

func handlePickSuccess(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	ctx := context.Background()
	_, _ = request.Send(ctx)

	log.Println("Thanks for ordering with us, you cat will not starve today.  Pick success for consignment",
		variables["consignment"], " in order ", variables["order_id"])
}

func handlePickFail(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	ctx := context.Background()
	_, _ = request.Send(ctx)

	log.Println("I'm sorry, your cat is going to get thin. Pick fail for consignment", variables["fail"],
		" in order ", variables["order_id"])
}

func failJob(client worker.JobClient, job entities.Job) {
	ctx := context.Background()
	_, _ = client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send(ctx)
}
